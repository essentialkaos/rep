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

	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo/rpm"
	"github.com/essentialkaos/rep/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdSign is 'sign' command handler
func cmdSign(ctx *context, args options.Arguments) bool {
	files := args.Filter("*.rpm").Strings()

	if len(files) == 0 {
		terminal.PrintWarnMessage("There are no RPM packages to sign")
		return false
	}

	if !checkRPMFiles(files) {
		return false
	}

	privateKey, ok := getRepoPrivateKey(ctx.Repo)

	if !ok {
		return false
	}

	return signRPMFiles(files, ctx, privateKey)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// signRPMFiles signs given RPM files
func signRPMFiles(files []string, ctx *context, privateKey *sign.PrivateKey) bool {
	tmpDir, err := ctx.Temp.MkDir("rep")

	if err != nil {
		terminal.PrintErrorMessage("Can't create temporary directore: %v", err)
		return false
	}

	var hasErrors bool

	for _, file := range files {
		ok := signRPMFile(file, tmpDir, ctx, privateKey)

		if !ok && !hasErrors {
			hasErrors = true
		}
	}

	return hasErrors == false
}

// signRPMFile signs given RPM file
func signRPMFile(file, tmpDir string, ctx *context, privateKey *sign.PrivateKey) bool {
	fileName := path.Base(file)

	spinner.Show("Signing "+colorTagPackage+"%s{!}", file)

	matchFilePattern, _ := path.Match(ctx.Repo.FileFilter, fileName)

	if !matchFilePattern {
		printSpinnerSignError(fileName, fmt.Sprintf("File doesn't match repository filter (%s)", ctx.Repo.FileFilter))
		return false
	}

	if !rpm.IsRPM(file) {
		printSpinnerSignError(fileName, "File is not an RPM package")
		return false
	}

	isSigned, err := sign.IsSigned(file, privateKey)

	if err != nil {
		printSpinnerSignError(fileName, err.Error())
		return false
	}

	if isSigned {
		spinner.Update("Package "+colorTagPackage+"%s{!} already signed with this key", file)
		spinner.Done(true)
		return true
	}

	tmpFile := path.Join(tmpDir, fileName)
	err = sign.Sign(file, tmpFile, privateKey)

	if err != nil {
		printSpinnerSignError(fileName, err.Error())
		return false
	}

	err = replaceSignedRPMFile(file, tmpFile)

	if err != nil {
		printSpinnerSignError(fileName, err.Error())
		return false
	}

	spinner.Update("Package "+colorTagPackage+"%s{!} signed", file)
	spinner.Done(true)

	return true
}

// replaceSignedRPMFile replaces original file with the signed one
func replaceSignedRPMFile(original, signed string) error {
	err := fsutil.CopyAttr(original, signed)

	if err != nil {
		return fmt.Errorf("Can't copy attributes: %v", err)
	}

	err = os.Remove(original)

	if err != nil {
		return fmt.Errorf("Can't replace file: %v", err)
	}

	return os.Rename(signed, original)
}

// printSpinnerSignError stops spinner and shows given error
func printSpinnerSignError(fileName string, err string) {
	spinner.Update("Can't sign "+colorTagPackage+"%s{!}", fileName)
	spinner.Done(false)
	terminal.PrintErrorMessage("   %v", err)
}