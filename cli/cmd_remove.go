package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/path"
	"github.com/essentialkaos/ek/v13/pluralize"
	"github.com/essentialkaos/ek/v13/spinner"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/input"

	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdRemove is 'remove' command handler
func cmdRemove(ctx *context, args options.Arguments) bool {
	var err error
	var filter string
	var testingStack, releaseStack repo.PackageStack

	testingStack, filter, err = smartPackageSearch(ctx.Repo.Testing, args)

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	if options.GetB(OPT_ALL) {
		releaseStack, _, err = smartPackageSearch(ctx.Repo.Release, args)

		if err != nil {
			terminal.Error(err.Error())
			return false
		}
	}

	if testingStack.IsEmpty() && releaseStack.IsEmpty() {
		terminal.Warn("No packages found")
		return false
	}

	return removePackages(ctx, releaseStack, testingStack, filter)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// removePackages removes packages from testing or all sub-repositories
func removePackages(ctx *context, releaseStack, testingStack repo.PackageStack, filter string) bool {
	var testingFiles, releaseFiles repo.PackageFiles

	if !releaseStack.IsEmpty() {
		releaseFiles = releaseStack.FlattenFiles()
	}

	if !testingStack.IsEmpty() {
		testingFiles = testingStack.FlattenFiles()
	}

	if !options.GetB(OPT_FORCE) {
		if !releaseStack.IsEmpty() {
			printPackageList(ctx.Repo.Release, releaseStack, filter)
		}

		if !testingStack.IsEmpty() {
			printPackageList(ctx.Repo.Testing, testingStack, filter)
		}

		fmtutil.Separator(false)

		if !releaseStack.IsEmpty() {
			fmtc.Printfn(
				" {s}{*}Release:{!*} -%s %s / -%s{!}", fmtutil.PrettyNum(len(testingFiles)),
				pluralize.Pluralize(len(testingFiles), "package", "packages"),
				fmtutil.PrettySize(testingFiles.Size()),
			)
		}

		if !testingStack.IsEmpty() {
			fmtc.Printfn(
				" {s}{*}Testing:{!*} -%s %s / -%s{!}", fmtutil.PrettyNum(len(testingFiles)),
				pluralize.Pluralize(len(testingFiles), "package", "packages"),
				fmtutil.PrettySize(testingFiles.Size()),
			)
		}

		fmtutil.Separator(false)

		ok, err := input.ReadAnswer("Do you really want to remove these packages?", "n")

		if err != nil || !ok {
			return false
		}
	}

	return removePackagesFiles(ctx, releaseFiles, testingFiles)
}

// removePackagesFiles removes packages files from testing or all sub-repositories
func removePackagesFiles(ctx *context, releaseFiles, testingFiles []repo.PackageFile) bool {
	var hasErrors, releaseRemoved, testingRemoved bool
	var file repo.PackageFile

	isCancelProtected = true

	for _, file = range releaseFiles {
		ok := removePackageFile(ctx, ctx.Repo.Release, file)

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
		ok := removePackageFile(ctx, ctx.Repo.Testing, file)

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

	if (releaseRemoved || testingRemoved) && !options.GetB(OPT_POSTPONE_INDEX) {
		fmtc.NewLine()

		if releaseRemoved {
			reindexRepository(ctx, ctx.Repo.Release, false)
		}

		if testingRemoved {
			reindexRepository(ctx, ctx.Repo.Testing, false)
		}
	}

	return !hasErrors
}

// removePackageFile removes package file from repository
func removePackageFile(ctx *context, r *repo.SubRepository, file repo.PackageFile) bool {
	fileName := path.Base(file.Path)
	repoArch := file.BaseArchFlag.String()
	archTag := fmtc.If(file.ArchFlag == data.ARCH_FLAG_NOARCH).Sprintf(" {s}[%s]{!}", repoArch)

	spinner.Show("Removing {?package}%s{!}%s", fileName, archTag)

	err := r.RemovePackage(file)

	if err != nil {
		spinner.Update(
			"Can't remove {?package}%s{!}%s from {*}{?repo}%s{!}",
			fileName, archTag, r.Name,
		)

		spinner.Done(false)
		terminal.Error("   %v", err)
		return false
	}

	spinner.Update(
		"Package {?package}%s{!}%s removed from {*}{?repo}%s{!}",
		fileName, archTag, r.Name,
	)

	spinner.Done(true)

	if file.ArchFlag == data.ARCH_FLAG_NOARCH {
		ctx.Logger.Get(r.Name).Print("Removed package %s (%s)", fileName, repoArch)
	} else {
		ctx.Logger.Get(r.Name).Print("Removed package %s", fileName)
	}

	return true
}
