package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdRemove is 'remove' command handler
func cmdRemove(ctx *context, args options.Arguments) bool {
	var err error
	var testingStack, releaseStack repo.PackageStack

	testingStack, err = smartPackageSearch(ctx.Repo.Testing, args)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if options.GetB(OPT_ALL) {
		releaseStack, err = smartPackageSearch(ctx.Repo.Release, args)

		if err != nil {
			terminal.PrintErrorMessage(err.Error())
			return false
		}
	}

	if testingStack.IsEmpty() && releaseStack.IsEmpty() {
		terminal.PrintWarnMessage("No packages found")
		return false
	}

	return removePackages(ctx, releaseStack, testingStack)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// removePackages removes packages from testing or all sub-repositories
func removePackages(ctx *context, releaseStack, testingStack repo.PackageStack) bool {
	if !options.GetB(OPT_FORCE) {
		if !releaseStack.IsEmpty() {
			printPackageList(ctx.Repo.Release, releaseStack, "")
		}

		if !testingStack.IsEmpty() {
			printPackageList(ctx.Repo.Testing, testingStack, "")
		}

		fmtutil.Separator(true)
		fmtc.NewLine()

		ok, err := terminal.ReadAnswer("Do you really want to remove these packages?", "n")

		if err != nil || !ok {
			return false
		}

		fmtc.NewLine()
	}

	testingFiles := testingStack.FlattenFiles()
	releaseFiles := releaseStack.FlattenFiles()

	return removePackagesFiles(ctx, releaseFiles, testingFiles)
}

// removePackagesFiles removes packages files from testing or all sub-repositories
func removePackagesFiles(ctx *context, releaseFiles, testingFiles []repo.PackageFile) bool {
	var hasErrors, releaseRemoved, testingRemoved bool
	var file repo.PackageFile

	isCancelProtected = true

	for _, file = range releaseFiles {
		ok := removePackageFile(ctx, ctx.Repo.Release, file.Path)

		if isCanceled {
			return false
		}

		if !ok {
			hasErrors = true
			continue
		}

		releaseRemoved = true
	}

	for _, file = range testingFiles {
		ok := removePackageFile(ctx, ctx.Repo.Testing, file.Path)

		if !ok {
			hasErrors = true
			continue
		}

		if isCanceled {
			return false
		}

		testingRemoved = true
	}

	isCancelProtected = false

	if releaseRemoved || testingRemoved {
		fmtc.NewLine()

		if releaseRemoved {
			reindexRepository(ctx, ctx.Repo.Release, false)
		}

		if testingRemoved {
			reindexRepository(ctx, ctx.Repo.Testing, false)
		}
	}

	return hasErrors == false
}

// removePackageFile removes package file from repository
func removePackageFile(ctx *context, r *repo.SubRepository, file string) bool {
	fileName := path.Base(file)

	spinner.Show("Removing "+colorTagPackage+"%s{!}", fileName)

	err := r.RemovePackage(file)

	if err != nil {
		spinner.Update(
			"Can't remove "+colorTagPackage+"%s{!} from "+colorTagRepository+"%s{!}",
			fileName, r.Name,
		)
		spinner.Done(false)
		terminal.PrintErrorMessage("   %v", err)
		return false
	}

	spinner.Update(
		"Package "+colorTagPackage+"%s{!} removed from "+colorTagRepository+"%s{!}",
		fileName, r.Name,
	)
	spinner.Done(true)

	ctx.Logger.Get(r.Name).Print("Removed package %s", fileName)

	return true
}
