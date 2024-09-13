package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"

	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/terminal"

	"github.com/essentialkaos/rep/v3/cli/query"
	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdFind is 'find' command handler
func cmdFind(ctx *context, args options.Arguments) bool {
	searchRequest, err := query.Parse(args.Strings())

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	if options.GetB(OPT_DEBUG) {
		printQueryDebug(searchRequest)
	}

	showAll := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)

	if showAll || options.GetB(OPT_RELEASE) {
		status := findAndShowPackages(ctx.Repo.Release, searchRequest)

		if status != true {
			return false
		}
	}

	if showAll || options.GetB(OPT_TESTING) {
		status := findAndShowPackages(ctx.Repo.Testing, searchRequest)

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

// findPackages tries to find packages with given search request
func findPackages(r *repo.SubRepository, searchRequest *query.Request) (repo.PackageStack, error) {
	if searchRequest == nil {
		return nil, fmt.Errorf("Search query must have at least one search (non-filtering) term")
	}

	stack, err := r.Find(searchRequest.Query)

	if err != nil {
		return nil, err
	}

	if !stack.IsEmpty() && searchRequest.FilterFlag != query.FILTER_FLAG_NONE {
		switch searchRequest.FilterFlag {
		case query.FILTER_FLAG_RELEASED:
			stack = filterPackagesByReleaseStatus(r, stack, true)
		case query.FILTER_FLAG_UNRELEASED:
			stack = filterPackagesByReleaseStatus(r, stack, false)
		}
	}

	return stack, err
}

// findAndShowPackages tries to find packages with given search request and show it
func findAndShowPackages(r *repo.SubRepository, searchRequest *query.Request) bool {
	stack, err := findPackages(r, searchRequest)

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	printPackageList(r, stack, "")

	return true
}

// filterPackagesByReleaseStatus filters given package stack by released status
func filterPackagesByReleaseStatus(r *repo.SubRepository, stack repo.PackageStack, released bool) repo.PackageStack {
	if r.Is(data.REPO_RELEASE) {
		if released {
			return stack
		} else {
			return nil
		}
	}

	for _, bundle := range stack {
		if bundle != nil {
			for index, pkg := range bundle {
				if pkg == nil {
					continue
				}

				isReleased, _, err := r.Parent.IsPackageReleased(pkg)

				if err == nil && isReleased != released {
					bundle[index] = nil
				}
			}
		}
	}

	return stack
}
