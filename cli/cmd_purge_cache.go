package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/terminal"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdPurgeCache is 'purge-cache' command handler
func cmdPurgeCache(ctx *context, args options.Arguments) bool {
	isCancelProtected.Store(true)

	err := ctx.Repo.PurgeCache()

	isCancelProtected.Store(false)

	if err != nil {
		terminal.Error("Can't clean cached data: %v", err)
		return false
	}

	fmtc.Println("{g}All cached data successfully deleted{!}")

	return true
}
