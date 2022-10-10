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
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdUnrelease is 'unrelease' command handler
func cmdUnrelease(ctx *context, args options.Arguments) bool {
	stack, filter, err := smartPackageSearch(ctx.Repo.Release, args)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if stack.IsEmpty() {
		terminal.PrintWarnMessage("No packages found")
		return false
	}

	return unreleasePackages(ctx, stack, filter)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// unreleasePackages removes packages from release sub-repository
func unreleasePackages(ctx *context, stack repo.PackageStack, filter string) bool {
	if !options.GetB(OPT_FORCE) {
		printPackageList(ctx.Repo.Release, stack, filter)

		fmtutil.Separator(true)
		fmtc.NewLine()

		ok, err := terminal.ReadAnswer("Do you really want to unrelease these packages?", "n")

		if err != nil || !ok {
			return false
		}

		fmtc.NewLine()
	}

	files := stack.FlattenFiles()

	return unreleasePackagesFiles(ctx, files)
}

// unreleasePackagesFiles removes packages files from release sub-repository
func unreleasePackagesFiles(ctx *context, files []repo.PackageFile) bool {
	var hasErrors, unreleased, restored bool

	isCancelProtected = true

	for _, file := range files {
		ok, testingRestored := unreleasePackageFile(ctx, file)

		if isCanceled {
			return false
		}

		if !ok {
			hasErrors = true
			continue
		}

		if testingRestored {
			restored = true
		}

		unreleased = true
	}

	if unreleased {
		fmtc.NewLine()

		reindexRepository(ctx, ctx.Repo.Release, false)

		if restored {
			reindexRepository(ctx, ctx.Repo.Testing, false)
		}
	}

	isCancelProtected = false

	return hasErrors == false
}

// unreleasePackageFile removes package file from repository
func unreleasePackageFile(ctx *context, file repo.PackageFile) (bool, bool) {
	var err error
	var restored bool

	fileName := path.Base(file.Path)
	repoArch := file.BaseArchFlag.String()
	archTag := fmtc.If(file.ArchFlag == data.ARCH_FLAG_NOARCH).Sprintf(" {s}[%s]{!}", repoArch)

	spinner.Show("Unreleasing {?package}%s{!}", fileName)

	if !ctx.Repo.Testing.HasPackageFile(fileName) {
		spinner.Show(
			"Moving {?package}%s{!}%s to {*}{?repo}%s{!}",
			fileName, archTag, data.REPO_TESTING,
		)

		err = ctx.Repo.CopyPackage(ctx.Repo.Release, ctx.Repo.Testing, file)

		if err != nil {
			printSpinnerUnreleaseError(file, err.Error())
			return false, false
		}

		restored = true
	}

	err = ctx.Repo.Release.RemovePackage(file)

	if err != nil {
		printSpinnerUnreleaseError(file, err.Error())
		return false, false
	}

	ctx.Logger.Get(data.REPO_RELEASE).Print("Unreleased package %s", fileName)

	if restored {
		spinner.Update(
			"Package {?package}%s{!}%s moved from {*}{?repo}%s{!} to {*}{?repo}%s{!}",
			fileName, archTag, data.REPO_RELEASE, data.REPO_TESTING,
		)
		ctx.Logger.Get(data.REPO_TESTING).Print("Restored package %s", fileName)
	} else {
		spinner.Update(
			"Package {?package}%s{!}%s removed from {*}{?repo}%s{!}",
			fileName, archTag, data.REPO_RELEASE,
		)
	}

	spinner.Done(true)

	if file.ArchFlag == data.ARCH_FLAG_NOARCH {
		ctx.Logger.Get(data.REPO_RELEASE).Print("Unreleased package %s (%s)", fileName, repoArch)
	} else {
		ctx.Logger.Get(data.REPO_RELEASE).Print("Unreleased package %s", fileName)
	}

	return true, restored
}

// printSpinnerUnreleaseError stops spinner and shows given error
func printSpinnerUnreleaseError(file repo.PackageFile, err string) {
	spinner.Update("Can't unrelease {?package}%s{!}", path.Base(file.Path))
	spinner.Done(false)
	terminal.PrintErrorMessage("   %v", err)
}
