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

// cmdRelease is 'release' command handler
func cmdRelease(ctx *context, args options.Arguments) bool {
	stack, filter, err := smartPackageSearch(ctx.Repo.Testing, args)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if stack.IsEmpty() {
		terminal.PrintWarnMessage("No packages found")
		return false
	}

	return releasePackages(ctx, stack, filter)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// releasePackages copies packages from testing to release repository
func releasePackages(ctx *context, stack repo.PackageStack, filter string) bool {
	if !options.GetB(OPT_FORCE) {
		printPackageList(ctx.Repo.Release, stack, filter)

		fmtutil.Separator(true)
		fmtc.NewLine()

		ok, err := terminal.ReadAnswer("Do you really want to release these packages?", "n")

		if err != nil || !ok {
			return false
		}

		fmtc.NewLine()
	}

	return releasePackagesFiles(ctx, stack.FlattenFiles())
}

// releasePackagesFiles copies packages files from testing to release repository
func releasePackagesFiles(ctx *context, files []repo.PackageFile) bool {
	var hasErrors, released bool

	isCancelProtected = true

	for _, file := range files {
		ok := releasePackageFile(ctx, file)

		if isCanceled {
			return false
		}

		if !ok {
			hasErrors = true
			continue
		}

		released = true
	}

	if released {
		fmtc.NewLine()
		reindexRepository(ctx, ctx.Repo.Release, false)
	}

	isCancelProtected = false

	return hasErrors == false
}

// releasePackageFile copies package file from testing to release repository
func releasePackageFile(ctx *context, file repo.PackageFile) bool {
	fileName := path.Base(file.Path)
	repoArch := file.BaseArchFlag.String()
	archTag := fmtc.If(file.ArchFlag == data.ARCH_FLAG_NOARCH).Sprintf(" {s-}[%s]{!}", repoArch)

	spinner.Show("Releasing {?package}%s{!}%s", fileName, archTag)

	err := ctx.Repo.CopyPackage(ctx.Repo.Testing, ctx.Repo.Release, file)

	if err != nil {
		spinner.Update("Can't release {?package}%s{!}%s", fileName, archTag)

		spinner.Done(false)
		terminal.PrintErrorMessage("   %v", err)

		return false
	}

	spinner.Update("Package {?package}%s{!}%s released", fileName, archTag)

	spinner.Done(true)

	if file.ArchFlag == data.ARCH_FLAG_NOARCH {
		ctx.Logger.Get(data.REPO_RELEASE).Print("Released package %s (%s)", fileName, repoArch)
	} else {
		ctx.Logger.Get(data.REPO_RELEASE).Print("Released package %s", fileName)
	}

	return true
}
