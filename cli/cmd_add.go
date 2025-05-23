package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/path"
	"github.com/essentialkaos/ek/v13/spinner"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/input"

	"github.com/essentialkaos/rep/v3/repo/data"
	"github.com/essentialkaos/rep/v3/repo/rpm"
	"github.com/essentialkaos/rep/v3/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdAdd is 'add' command handler
func cmdAdd(ctx *context, args options.Arguments) bool {
	files := args.Filter("*.rpm").Strings()
	files = filterRPMPackages(ctx, files)

	if len(files) == 0 {
		terminal.Warn("There are no RPM packages to add")
		return false
	}

	if !checkRPMFiles(files) {
		return false
	}

	if !options.GetB(OPT_FORCE) {
		printFilesList(files)

		ok, err := input.ReadAnswer("Do you want to add these packages?", "n")

		if err != nil || !ok {
			return false
		}
	}

	if !isSignRequired(ctx.Repo.Testing, files) {
		return addRPMFiles(ctx, files, nil)
	}

	signingKey, ok := getRepoSigningKey(ctx.Repo)

	if !ok {
		return false
	}

	return addRPMFiles(ctx, files, signingKey)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printFilesList prints list with new packages
func printFilesList(files []string) {
	for _, file := range files {
		filename := path.Base(file)
		fmtc.Printfn("{s-}•{!} {?package}%s{!}", filename)
	}

	fmtc.NewLine()
}

// addRPMFiles adds given RPM files to testing repository
func addRPMFiles(ctx *context, files []string, signingKey *sign.Key) bool {
	tmpDir, err := ctx.Temp.MkDir("rep")

	if err != nil {
		terminal.Error("Can't create temporary directory: %v", err)
		return false
	}

	isCancelProtected = true

	var hasErrors, hasAdded bool

	for _, file := range files {
		ok := addRPMFile(ctx, file, tmpDir, signingKey)

		if isCanceled {
			return false
		}

		if !ok {
			hasErrors = true
			continue
		}

		hasAdded = true
	}

	if hasAdded && !options.GetB(OPT_POSTPONE_INDEX) {
		fmtc.NewLine()
		reindexRepository(ctx, ctx.Repo.Testing, false)
	}

	isCancelProtected = false

	return !hasErrors
}

// addRPMFile adds given RPM file to testing repository
func addRPMFile(ctx *context, file, tmpDir string, signingKey *sign.Key) bool {
	var err error

	fileName := path.Base(file)

	if options.GetB(OPT_MOVE) {
		spinner.Show("Moving {?package}%s{!}", fileName)
	} else {
		spinner.Show("Copying {?package}%s{!}", fileName)
	}

	if !options.GetB(OPT_IGNORE_FILTER) {
		matchFilePattern, err := path.Match(ctx.Repo.FileFilter, fileName)

		if err != nil {
			printSpinnerAddError(fileName, fmt.Sprintf("Can't parse file filter pattern: %v", err))
			return false
		}

		if !matchFilePattern {
			printSpinnerAddError(fileName, fmt.Sprintf("File doesn't match repository filter (%s)", ctx.Repo.FileFilter))
			return false
		}
	}

	if options.GetB(OPT_NO_SOURCE) {
		matchFilePattern, _ := path.Match("*.src.rpm", fileName)

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

	if signingKey != nil {
		isSignValid, err := sign.IsPackageSignatureValid(file, signingKey)

		if err != nil {
			printSpinnerAddError(fileName, fmt.Sprintf("Can't check package signature: %v", err))
			return false
		}

		if !isSignValid {
			spinner.Update("Signing {?package}%s{!}", fileName)

			pkgFile = path.Join(tmpDir, fileName)
			err = sign.SignPackage(file, pkgFile, signingKey)

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

		spinner.Update("Package {?package}%s{!} moved to {*}{?repo}%s{!}", fileName, data.REPO_TESTING)
		spinner.Done(true)
	} else {
		spinner.Update("Package {?package}%s{!} added to {*}{?repo}%s{!}", fileName, data.REPO_TESTING)
		spinner.Done(true)
	}

	ctx.Logger.Get(data.REPO_TESTING).Print("Added package %s", fileName)

	return true
}

// printSpinnerAddError stops spinner and shows given error
func printSpinnerAddError(fileName string, err string) {
	spinner.Update("Can't add {?package}%s{!}", fileName)
	spinner.Done(false)
	terminal.Error("   %v", err)
}

// filterRPMPackages filters packages using repository file filter pattern
func filterRPMPackages(ctx *context, files []string) []string {
	if options.GetB(OPT_IGNORE_FILTER) {
		return files
	}

	var result []string

	for _, file := range files {
		fileName := path.Base(file)
		matchFilePattern, err := path.Match(ctx.Repo.FileFilter, fileName)

		if err == nil && !matchFilePattern {
			continue
		}

		result = append(result, file)
	}

	return result
}
