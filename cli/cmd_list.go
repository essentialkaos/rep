package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdList is 'list' command handler
func cmdList(ctx *context, args options.Arguments) bool {
	filter := args.Get(0).String()
	showAll := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)

	if showAll || options.GetB(OPT_RELEASE) {
		status := listPackages(ctx.Repo.Release, filter)

		if status != true {
			return false
		}
	}

	if showAll || options.GetB(OPT_TESTING) {
		status := listPackages(ctx.Repo.Testing, filter)

		if status != true {
			return false
		}
	}

	if !rawOutput {
		fmtutil.Separator(true)
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// listPackages prints package listing for given sub-repository
func listPackages(r *repo.SubRepository, filter string) bool {
	stack, err := r.List(filter, options.GetB(OPT_SHOW_ALL))

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	printPackageList(r, stack, filter)

	return true
}

// printPackageList prints package listing for given sub-repository
func printPackageList(r *repo.SubRepository, stack repo.PackageStack, filter string) {
	if !rawOutput {
		fmtutil.Separator(true, strings.ToUpper(r.Name))
		fmtc.NewLine()
		printPackageStack(r, stack, filter)
		fmtc.NewLine()
	} else {
		printRawPackageStack(r, stack)
	}
}

// printPackageStack prints info about packages in stack
func printPackageStack(r *repo.SubRepository, stack repo.PackageStack, filter string) {
	if stack.IsEmpty() {
		fmtc.Println("{s-}-- empty --{!}")
		return
	}

	archList := stack.GetArchs()

	for _, bundle := range stack {
		if bundle != nil {
			printPackageBundle(r, bundle, archList, stack.HasMultiBundles(), filter)
		}
	}
}

// printRawPackageStack prints info about packages in stack
func printRawPackageStack(r *repo.SubRepository, stack repo.PackageStack) {
	for _, file := range stack.FlattenFiles() {
		fmt.Println(r.GetFullPackagePath(file))
	}
}

// printPackageBundle prints info about packages in bundle
func printPackageBundle(r *repo.SubRepository, bundle repo.PackageBundle, archList []string, hasMultiBundle bool, filter string) {
	var groupSym string

	for index, pkg := range bundle {
		if pkg == nil {
			continue
		}

		switch {
		case index == 0 && len(bundle) == 1 && hasMultiBundle:
			groupSym = "   "
		case len(bundle) != 1 && index == 0:
			groupSym = " {s}┌{!} "
		case len(bundle) != 1 && index == len(bundle)-1:
			groupSym = " {s}└{!} "
		case len(bundle) != 1:
			groupSym = " {s}│{!} "
		default:
			groupSym = " "
		}

		fmtc.Println(genListArchInfo(pkg, archList) + groupSym + genListPkgName(r, pkg, filter))
	}
}

// genListArchInfo generates arches info for listing
func genListArchInfo(pkg *repo.Package, archList []string) string {
	result := "[ "

	for _, arch := range archList {
		tag := archTags[arch]

		if pkg.HasArch(arch) {
			result += archColors[arch] + tag + "{!} "
		} else {
			result += "{s-}" + strings.Repeat("∙", len(tag)) + "{!} "
		}
	}

	return result + "]"
}

// genListPkgName generates package name for listing
func genListPkgName(r *repo.SubRepository, pkg *repo.Package, filter string) string {
	pkgName := pkg.FullName()

	if options.GetB(OPT_STATUS) && r.Name == data.REPO_TESTING {
		isReleased, _, _ := r.Parent.IsPackageReleased(pkg)

		if isReleased {
			pkgName = "{g}" + pkgName + "{!}"
		}
	}

	if options.GetB(OPT_EPOCH) {
		pkgName = "{s-}" + pkg.Epoch + ":{!}" + pkgName
	}

	if filter != "" {
		pkgName = strings.ReplaceAll(pkgName, filter, "{_}"+filter+"{!_}")
	}

	return pkgName
}
