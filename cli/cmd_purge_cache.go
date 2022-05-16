package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/terminal"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdPurgeCache is 'purge-cache' command handler
func cmdPurgeCache(ctx *context, args options.Arguments) bool {
	err := ctx.Repo.PurgeCache()

	if err != nil {
		terminal.PrintErrorMessage("Can't clean cached data: %v", err)
		return false
	}

	fmtc.Println("{g}All cached data successfully deleted{!}")

	return true
}
