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
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/cli/query"
	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/search"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdFind is 'find' command handler
func cmdFind(ctx *context, args options.Arguments) bool {
	searchQuery, err := query.Parse(args.Strings())

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if options.GetB(OPT_DEBUG) {
		printQueryDebug(searchQuery)
	}

	showAll := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)

	if showAll || options.GetB(OPT_RELEASE) {
		status := findPackages(ctx.Repo.Release, searchQuery)

		if status != true {
			return false
		}
	}

	if showAll || options.GetB(OPT_TESTING) {
		status := findPackages(ctx.Repo.Testing, searchQuery)

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

// findPackages tries to find packages with given query
func findPackages(r *repo.SubRepository, q search.Query) bool {
	stack, err := r.Find(q)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	printPackageList(r, stack, "")

	return true
}

// printQueryDebug prints debug search query info
func printQueryDebug(q search.Query) {

	for index, term := range q {
		db, qrs := term.SQL()

		for _, qr := range qrs {
			fmtc.Printf("{s-}{%d|%s} %s → %s{!}\n", index, db, term, qr)
		}
	}

	fmtc.NewLine()
}
