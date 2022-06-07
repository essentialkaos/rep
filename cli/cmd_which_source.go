package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/mathutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/cli/query"
	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdWhichSource is 'which-source' command handler
func cmdWhichSource(ctx *context, args options.Arguments) bool {
	showAll := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)
	searchRequest, err := query.Parse(args.Strings())

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if options.GetB(OPT_RELEASE) || showAll {
		status := findSources(ctx.Repo.Release, searchRequest)

		if status != true {
			return false
		}
	}

	if options.GetB(OPT_TESTING) || showAll {
		status := findSources(ctx.Repo.Testing, searchRequest)

		if status != true {
			return false
		}
	}

	fmtutil.Separator(true)

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// findSources tries to find source package name
func findSources(r *repo.SubRepository, searchRequest *query.Request) bool {
	fmtutil.Separator(true, strings.ToUpper(r.Name))
	fmtc.NewLine()

	stack, err := findPackages(r, searchRequest)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	printPackageStackSources(r, stack)
	fmtc.NewLine()

	return true
}

// printPackageStackSources prints list of packages with info about source package
func printPackageStackSources(r *repo.SubRepository, stack repo.PackageStack) {
	if len(stack) == 0 {
		fmtc.Println("{s-}-- empty --{!}")
		return
	}

	hasMultiBundle := stack.HasMultiBundles()
	maxSrcSize := getMaxSourceLengthInStack(stack)

	for _, bundle := range stack {
		for index, pkg := range bundle {
			pkgInfo := strings.Repeat(" ", maxSrcSize-len(pkg.Src))

			switch {
			case index == 0 && len(bundle) == 1 && hasMultiBundle:
				pkgInfo += "   "
			case len(bundle) != 1 && index == 0:
				pkgInfo += " {s}┌{!} "
			case len(bundle) != 1 && index == len(bundle)-1:
				pkgInfo += " {s}└{!} "
			case len(bundle) != 1:
				pkgInfo += " {s}│{!} "
			default:
				pkgInfo += " "
			}

			if options.GetB(OPT_EPOCH) {
				pkgInfo += "{s-}" + pkg.Epoch + ":{!}"
			}

			if options.GetB(OPT_STATUS) && r.Name == data.REPO_TESTING {
				isReleased, _, _ := r.Parent.IsPackageReleased(pkg)

				if isReleased {
					pkgInfo += "{g}" + pkg.FullName() + "{!}"
				} else {
					pkgInfo += pkg.FullName()
				}
			} else {
				pkgInfo += pkg.FullName()
			}

			switch index {
			case 0:
				fmtc.Printf("[ {*}%s{!} ]"+pkgInfo+"\n", pkg.Src)
			default:
				fmtc.Printf("[ {s-}%s{!} ]"+pkgInfo+"\n", pkg.Src)
			}
		}
	}
}

// getMaxSourceLengthInStack returns max size of source rpm in stack
func getMaxSourceLengthInStack(stack repo.PackageStack) int {
	var size int

	for _, bundle := range stack {
		for _, pkg := range bundle {
			size = mathutil.Max(size, len(pkg.Src))
		}
	}

	return size
}
