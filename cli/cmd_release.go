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

	"github.com/essentialkaos/rep/cli/query"
	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdRelease is 'release' command handler
func cmdRelease(ctx *context, args options.Arguments) bool {
	searchRequest, err := query.Parse(args.Strings())

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	stack, err := findPackages(ctx.Repo.Testing, searchRequest)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if stack.IsEmpty() {
		terminal.PrintWarnMessage("No packages found")
		return false
	}

	return releasePackages(ctx, stack)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// releasePackages copies packages from testing to release repository
func releasePackages(ctx *context, stack repo.PackageStack) bool {
	printPackageList(ctx.Repo.Release, stack, "")

	fmtutil.Separator(true)
	fmtc.NewLine()

	if !options.GetB(OPT_FORCE) {
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

	for _, file := range files {
		ok := releasePackageFile(ctx, file.Path)

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

	return hasErrors == false
}

// releasePackageFile copies package file from testing to release repository
func releasePackageFile(ctx *context, file string) bool {
	fileName := path.Base(file)

	spinner.Show("Releasing "+colorTagPackage+"%s{!}", fileName)

	err := ctx.Repo.CopyPackage(ctx.Repo.Testing, ctx.Repo.Release, file)

	if err != nil {
		spinner.Update("Can't release "+colorTagPackage+"%s{!}", fileName)
		spinner.Done(false)
		terminal.PrintErrorMessage("   %v", err)
		return false
	}

	spinner.Update("Package "+colorTagPackage+"%s{!} released", fileName)
	spinner.Done(true)

	ctx.Logger.Get(data.REPO_RELEASE).Print("Released package %s", fileName)

	return true
}
