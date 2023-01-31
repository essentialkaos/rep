package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/terminal"
	"github.com/essentialkaos/ek/v12/timeutil"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdStats is 'stats' command handler
func cmdStats(ctx *context, args options.Arguments) bool {
	showAll := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)

	if showAll || options.GetB(OPT_RELEASE) {
		stats, err := ctx.Repo.Release.Stats()

		if err != nil {
			terminal.PrintErrorMessage(err.Error())
			return false
		}

		printRepoStats(ctx.Repo.Release, stats)

		fmtc.NewLine()
	}

	if showAll || options.GetB(OPT_TESTING) {
		stats, err := ctx.Repo.Testing.Stats()

		if err != nil {
			terminal.PrintErrorMessage(err.Error())
			return false
		}

		printRepoStats(ctx.Repo.Testing, stats)

		fmtc.NewLine()
	}

	fmtutil.Separator(true)

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printRepoStats prints repo stats
func printRepoStats(r *repo.SubRepository, stats *repo.RepositoryStats) {
	fmtutil.Separator(true, strings.ToUpper(r.Name))
	fmtc.NewLine()

	if stats.TotalPackages == 0 {
		fmtc.Println("{s-}-- empty --{!}")
		return
	}

	fmtc.Printf(
		"{*}Packages:{!}  %s {s}(%s){!}\n",
		fmtutil.PrettyNum(stats.TotalPackages),
		fmtutil.PrettySize(stats.TotalSize),
	)

	fmtc.NewLine()

	for _, arch := range data.ArchList {
		_, ok := stats.Packages[arch]

		if !ok {
			continue
		}

		color := archColors[arch]

		if fmtc.Is256ColorsSupported() {
			color = archColorsExt[arch]
		}

		fmtc.Printf(
			color+"%-9s{!}  %s {s}(%s){!}\n",
			arch, fmtutil.PrettyNum(stats.Packages[arch]),
			fmtutil.PrettySize(stats.Sizes[arch]),
		)
	}

	fmtc.NewLine()

	fmtc.Printf(
		"{*}Updated:{!}   %s\n",
		timeutil.Format(stats.Updated, "%Y/%m/%d %H:%M"),
	)
}
