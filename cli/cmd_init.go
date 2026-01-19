package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"slices"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/sliceutil"
	"github.com/essentialkaos/ek/v13/terminal"

	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdAdd is 'init' command handler
func cmdInit(ctx *context, args options.Arguments) bool {
	supportedArchs := sliceutil.Exclude(data.ArchList, data.ARCH_NOARCH)
	archList := args.Strings()

	for _, arch := range archList {
		if !slices.Contains(supportedArchs, arch) {
			terminal.Error("Architecture %q is not supported (typo?)", arch)
			return false
		}
	}

	err := ctx.Repo.Initialize(archList)

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	fmtc.Println("{g}Repository successfully initialized!{!}")

	return false
}
