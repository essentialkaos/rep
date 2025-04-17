package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/spinner"
	"github.com/essentialkaos/ek/v13/terminal"

	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
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

	if isCanceled {
		return false
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
	spinner.Show("Indexing {*}{?repo}%s{!} repository", r.Name)

	isCancelProtected = true

	ch := make(chan string, len(data.SupportedArchs))

	go updateReindexStatus(ch, r.Name)

	err := r.Reindex(full, ch)

	if err == nil {
		spinner.Update("Index for {*}{?repo}%s{!} repository successfully built", r.Name)
	} else {
		spinner.Update("Can't create index for {*}{?repo}%s{!} repository", r.Name)
	}

	spinner.Done(err == nil)

	isCancelProtected = false

	if err != nil {
		terminal.Error("   %v", err)
		return false
	}

	return true
}

// updateReindexStatus updates spinner status
func updateReindexStatus(ch chan string, name string) {
	for arch := range ch {
		spinner.Update("Indexing {*}{?repo}%s{!} {s-}(%s){!} repository", name, arch)
	}
}
