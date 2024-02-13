package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
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

	"github.com/essentialkaos/rep/v3/repo"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// MIN_KEEP_CLEANUP_VERS is minimal number of versions to keep
const MIN_KEEP_CLEANUP_VERS = 3

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdCleanup is 'cleanup' command handler
func cmdCleanup(ctx *context, args options.Arguments) bool {
	var testingStack, releaseStack repo.PackageStack

	all := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)
	keepNum, filter, err := getCleanupOptions(args)

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	if all || options.GetB(OPT_RELEASE) {
		releaseStack, err = getStackToCleanup(ctx.Repo.Release, keepNum, filter)

		if err != nil {
			terminal.Error(err.Error())
			return false
		}
	}

	if all || options.GetB(OPT_TESTING) {
		testingStack, err = getStackToCleanup(ctx.Repo.Testing, keepNum, filter)

		if err != nil {
			terminal.Error(err.Error())
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

// getCleanupOptions returns number of versions and filter from arguments
func getCleanupOptions(args options.Arguments) (int, string, error) {
	var err error

	keepNum := MIN_KEEP_CLEANUP_VERS

	if args.Has(0) {
		keepNum, err = args.Get(0).Int()

		if err != nil {
			return 0, "", fmt.Errorf("Can't parse number of versions: %v", err)
		}

		if keepNum < MIN_KEEP_CLEANUP_VERS {
			return 0, "", fmt.Errorf(
				"Number of versions can't be less than %d",
				MIN_KEEP_CLEANUP_VERS,
			)
		}
	}

	filter := args.Get(1).String()

	return keepNum, filter, nil
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
func getStackToCleanup(r *repo.SubRepository, keepNum int, filter string) (repo.PackageStack, error) {
	stack, err := r.List("", true)

	if err != nil {
		return nil, err
	}

	return extractPackagesToCleanup(stack, keepNum, filter), nil
}

// extractPackagesToCleanup extracts bundles to remove from stack
// with all packages
func extractPackagesToCleanup(stack repo.PackageStack, keepNum int, filter string) repo.PackageStack {
	var result repo.PackageStack
	var prevPkgName, prevPkgVer string
	var pkgCount int

	sort.Sort(sort.Reverse(stack))

	for _, bundle := range stack {
		pkg := getMainPackageFromBundle(bundle)

		if pkg == nil {
			continue
		}

		if filter != "" && !strings.HasPrefix(pkg.Src, filter) {
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

		prevPkgName, prevPkgVer = pkg.Name, pkg.Version

		if pkgCount > keepNum {
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
