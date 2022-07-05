package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/rpm"
	"github.com/essentialkaos/rep/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdAdd is 'add' command handler
func cmdAdd(ctx *context, args options.Arguments) bool {
	files := args.Filter("*.rpm").Strings()

	if len(files) == 0 {
		terminal.PrintWarnMessage("There are no RPM packages to add")
		return false
	}

	if !checkRPMFiles(files) {
		return false
	}

	if !options.GetB(OPT_FORCE) {
		printFilesList(files)

		ok, err := terminal.ReadAnswer("Do you want to add these packages?", "n")

		if err != nil || !ok {
			return false
		}

		fmtc.NewLine()
	}

	if !isSignRequired(ctx.Repo.Testing, files) {
		return addRPMFiles(ctx, files, nil)
	}

	privateKey, ok := getRepoPrivateKey(ctx.Repo)

	if !ok {
		return false
	}

	return addRPMFiles(ctx, files, privateKey)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printFilesList prints list with new packages
func printFilesList(files []string) {
	for _, file := range files {
		filename := path.Base(file)
		fmtc.Printf("{s-}•{!} "+colorTagPackage+"%s{!}\n", filename)
	}

	fmtc.NewLine()
}

// addRPMFiles adds given RPM files to testing repository
func addRPMFiles(ctx *context, files []string, privateKey *sign.PrivateKey) bool {
	tmpDir, err := ctx.Temp.MkDir("rep")

	if err != nil {
		terminal.PrintErrorMessage("Can't create temporary directory: %v", err)
		return false
	}

	isCancelProtected = true

	var hasErrors, hasAdded bool

	for _, file := range files {
		ok := addRPMFile(ctx, file, tmpDir, privateKey)

		if !ok && !hasErrors {
			hasErrors = true
			continue
		}

		if isCanceled {
			return false
		}

		hasAdded = true
	}

	if hasAdded {
		fmtc.NewLine()
		reindexRepository(ctx, ctx.Repo.Testing, false)
	}

	isCancelProtected = false

	return hasErrors == false
}

// addRPMFile adds given RPM file to testing repository
func addRPMFile(ctx *context, file, tmpDir string, privateKey *sign.PrivateKey) bool {
	fileName := path.Base(file)

	spinner.Show("Adding "+colorTagPackage+"%s{!}", fileName)

	matchFilePattern, err := path.Match(ctx.Repo.FileFilter, fileName)

	if err != nil {
		printSpinnerAddError(fileName, fmt.Sprintf("Can't parse file filter pattern: %v", err))
		return false
	}

	if !matchFilePattern {
		printSpinnerAddError(fileName, fmt.Sprintf("File doesn't match repository filter (%s)", ctx.Repo.FileFilter))
		return false
	}

	if options.GetB(OPT_NO_SOURCE) {
		matchFilePattern, _ = path.Match("*.src.rpm", fileName)

		if matchFilePattern {
			skipOption, _ := options.ParseOptionName(OPT_NO_SOURCE)
			spinner.Update("{s}Skip %s (due to --%s option){!}", fileName, skipOption)
			spinner.Skip()
			return true
		}
	}

	if !rpm.IsRPM(file) {
		printSpinnerAddError(fileName, "File is not an RPM package")
		return false
	}

	if ctx.Repo.Testing.HasPackageFile(fileName) && !ctx.Repo.Replace {
		printSpinnerAddError(fileName, "Package already present in repository and replacement is forbidden in the configuration file")
		return false
	}

	pkgFile := file

	if privateKey != nil {
		isSigned, err := sign.IsSigned(file, privateKey)

		if err != nil {
			printSpinnerAddError(fileName, fmt.Sprintf("Can't check package signature: %v", err))
			return false
		}

		if !isSigned {
			spinner.Update("Signing "+colorTagPackage+"%s{!}", fileName)

			pkgFile = path.Join(tmpDir, fileName)
			err = sign.Sign(file, pkgFile, privateKey)

			if err != nil {
				printSpinnerAddError(fileName, fmt.Sprintf("Can't sign package: %v", err))
				return false
			}

			defer os.Remove(pkgFile)
		}
	}

	err = ctx.Repo.Testing.AddPackage(pkgFile)

	if err != nil {
		printSpinnerAddError(fileName, err.Error())
		return false
	}

	if options.GetB(OPT_MOVE) {
		err = os.Remove(file)

		if err != nil {
			printSpinnerAddError(fileName, fmt.Sprintf("Can't remove file: %v", err))
			return false
		}

		spinner.Update("Package "+colorTagPackage+"%s{!} moved to "+colorTagRepository+"%s{!}", fileName, data.REPO_TESTING)
		spinner.Done(true)
	} else {
		spinner.Update("Package "+colorTagPackage+"%s{!} added to "+colorTagRepository+"%s{!}", fileName, data.REPO_TESTING)
		spinner.Done(true)
	}

	ctx.Logger.Get(data.REPO_TESTING).Print("Added package %s", fileName)

	return true
}

// printSpinnerAddError stops spinner and shows given error
func printSpinnerAddError(fileName string, err string) {
	spinner.Update("Can't add "+colorTagPackage+"%s{!}", fileName)
	spinner.Done(false)
	terminal.PrintErrorMessage("   %v", err)
}
