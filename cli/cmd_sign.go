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

	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/path"
	"github.com/essentialkaos/ek/v13/spinner"
	"github.com/essentialkaos/ek/v13/terminal"

	"github.com/essentialkaos/rep/v3/repo/rpm"
	"github.com/essentialkaos/rep/v3/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdSign is 'sign' command handler
func cmdSign(ctx *context, args options.Arguments) bool {
	files := args.Filter("*.rpm").Strings()

	if len(files) == 0 {
		terminal.Warn("There are no RPM packages to sign")
		return false
	}

	if !checkRPMFiles(files) {
		return false
	}

	key, ok := getRepoSigningKey(ctx.Repo)

	if !ok {
		return false
	}

	return signRPMFiles(files, ctx, key)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// signRPMFiles signs given RPM files
func signRPMFiles(files []string, ctx *context, key *sign.Key) bool {
	tmpDir, err := ctx.Temp.MkDir("rep")

	if err != nil {
		terminal.Error("Can't create temporary directory: %v", err)
		return false
	}

	isCancelProtected = true

	var hasErrors bool

	for _, file := range files {
		ok := signRPMFile(file, tmpDir, ctx, key)

		if isCanceled {
			return false
		}

		if !ok {
			hasErrors = true
		}
	}

	isCancelProtected = false

	return hasErrors == false
}

// signRPMFile signs given RPM file
func signRPMFile(file, tmpDir string, ctx *context, key *sign.Key) bool {
	var err error

	fileName := path.Base(file)

	spinner.Show("Signing {?package}%s{!}", file)

	if !options.GetB(OPT_IGNORE_FILTER) {
		matchFilePattern, err := path.Match(ctx.Repo.FileFilter, fileName)

		if err != nil {
			printSpinnerSignError(fileName, fmt.Sprintf("Can't parse file filter pattern: %v", err))
			return false
		}

		if !matchFilePattern {
			printSpinnerSignError(fileName, fmt.Sprintf("File doesn't match repository filter (%s)", ctx.Repo.FileFilter))
			return false
		}
	}

	if !rpm.IsRPM(file) {
		printSpinnerSignError(fileName, "File is not an RPM package")
		return false
	}

	isSignValid, err := sign.IsPackageSignatureValid(file, key)

	if err != nil {
		printSpinnerSignError(fileName, err.Error())
		return false
	}

	if isSignValid {
		spinner.Update("Package {?package}%s{!} already signed with this key", file)
		spinner.Done(true)
		return true
	}

	tmpFile := path.Join(tmpDir, fileName)
	err = sign.SignPackage(file, tmpFile, key)

	if err != nil {
		printSpinnerSignError(fileName, err.Error())
		return false
	}

	err = replaceSignedRPMFile(file, tmpFile)

	if err != nil {
		printSpinnerSignError(fileName, err.Error())
		return false
	}

	spinner.Update("Package {?package}%s{!} signed", file)
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
		return fmt.Errorf("Can't remove original (non-signed) file: %v", err)
	}

	err = fsutil.CopyFile(signed, original)

	if err != nil {
		return fmt.Errorf("Can't copy signed file: %v", err)
	}

	err = os.Remove(signed)

	if err != nil {
		return fmt.Errorf("Can't remove temporary file: %v", err)
	}

	return nil
}

// printSpinnerSignError stops spinner and shows given error
func printSpinnerSignError(fileName string, err string) {
	spinner.Update("Can't sign {?package}%s{!}", fileName)
	spinner.Done(false)
	terminal.Error("   %v", err)
}
