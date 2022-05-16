package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdReindex is 'reindex' command handler
func cmdReindex(ctx *context, args options.Arguments) bool {
	reindexAll := !options.GetB(OPT_RELEASE) && !options.GetB(OPT_TESTING)
	full := options.GetB(OPT_FULL)

	if reindexAll || options.GetB(OPT_RELEASE) {
		if !reindexRepository(ctx, ctx.Repo.Release, full) {
			return false
		}

		ctx.Logger.Get(data.REPO_RELEASE).Print("Repository reindexed (full: %t)", full)
	}

	if reindexAll || options.GetB(OPT_TESTING) {
		if !reindexRepository(ctx, ctx.Repo.Testing, full) {
			return false
		}

		ctx.Logger.Get(data.REPO_TESTING).Print("Repository reindexed (full: %t)", full)
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// reindexRepository starts repository reindex
func reindexRepository(ctx *context, r *repo.SubRepository, full bool) bool {
	spinner.Show("Indexing "+colorTagRepository+"%s{!} repository", r.Name)

	err := r.Reindex(full)

	if err == nil {
		spinner.Update("Index for "+colorTagRepository+"%s{!} repository successfully built", r.Name)
	} else {
		spinner.Update("Can't create index for "+colorTagRepository+"%s{!} repository", r.Name)
	}

	spinner.Done(err == nil)

	if err != nil {
		terminal.PrintErrorMessage("   %v", err)
		return false
	}

	return true
}
