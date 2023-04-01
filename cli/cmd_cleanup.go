package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"sort"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// MIN_CLEANUP_VERS is minimal number of versions
const MIN_CLEANUP_VERS = 3

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdCleanup is 'cleanup' command handler
func cmdCleanup(ctx *context, args options.Arguments) bool {
	var testingStack, releaseStack repo.PackageStack

	all := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)
	keepVerNum, err := getCleanupVersionNum(args)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if all || options.GetB(OPT_RELEASE) {
		releaseStack, err = getStackToCleanup(ctx.Repo.Release, keepVerNum)

		if err != nil {
			terminal.PrintErrorMessage(err.Error())
			return false
		}
	}

	if all || options.GetB(OPT_TESTING) {
		testingStack, err = getStackToCleanup(ctx.Repo.Testing, keepVerNum)

		if err != nil {
			terminal.PrintErrorMessage(err.Error())
			return false
		}
	}

	if testingStack.IsEmpty() && releaseStack.IsEmpty() {
		fmtc.Println("{g}No packages to cleanup{!}")
		return true
	}

	return cleanupPackages(ctx, releaseStack, testingStack)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getCleanupVersionNum returns number of versions from arguments
func getCleanupVersionNum(args options.Arguments) (int, error) {
	var err error

	keepVerNum := MIN_CLEANUP_VERS

	if args.Has(0) {
		keepVerNum, err = args.Get(0).Int()

		if err != nil {
			return 0, fmt.Errorf("Can't parse number of versions: %v", err)
		}

		if keepVerNum < MIN_CLEANUP_VERS {
			return 0, fmt.Errorf(
				"Number of versions can't be less than %d",
				MIN_CLEANUP_VERS,
			)
		}
	}

	return keepVerNum, nil
}

// cleanupPackages removes packages from both repositories
func cleanupPackages(ctx *context, releaseStack, testingStack repo.PackageStack) bool {
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

// getStackToCleanup returns stack with packages to remove
func getStackToCleanup(r *repo.SubRepository, keepVerNum int) (repo.PackageStack, error) {
	stack, err := r.List("", true)

	if err != nil {
		return nil, err
	}

	return extractPackagesToCleanup(stack, keepVerNum), nil
}

// extractPackagesToCleanup extracts bundles to remove from stack
// with all packages
func extractPackagesToCleanup(stack repo.PackageStack, keepVerNum int) repo.PackageStack {
	var result repo.PackageStack
	var prevPkgName, prevPkgVer string
	var pkgCount int

	sort.Sort(sort.Reverse(stack))

	for _, bundle := range stack {
		pkg := getMainPackageFromBundle(bundle)

		if pkg == nil {
			continue
		}

		switch {
		case prevPkgName != pkg.Name:
			pkgCount = 1
		case prevPkgName == pkg.Name:
			if prevPkgVer != pkg.Version {
				pkgCount++
			}
		}

		prevPkgName = pkg.Name
		prevPkgVer = pkg.Version

		if pkgCount > keepVerNum {
			result = append(result, bundle)
		}
	}

	if len(result) != 0 {
		sort.Sort(result)
	}

	return result
}

// getMainPackageFromBundle returns main package from bundle
func getMainPackageFromBundle(bundle repo.PackageBundle) *repo.Package {
	if len(bundle) == 1 {
		return bundle[0]
	}

	for _, pkg := range bundle {
		srcName := fmt.Sprintf(
			"%s-%s-%s.src.rpm",
			pkg.Name, pkg.Version, pkg.Release,
		)

		if srcName == pkg.Src {
			return pkg
		}
	}

	return nil
}

// printEmptyFoundPackageList prints empty packages listing
func printEmptyFoundPackageList(r *repo.SubRepository) {
	fmtutil.Separator(true, strings.ToUpper(r.Name))
	fmtc.Println("\n{s-}-- no packages --{!}\n")
}
