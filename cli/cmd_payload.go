package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/lscolors"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/strutil"
	"github.com/essentialkaos/ek/v13/terminal"

	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdPayload is 'payload' command handler
func cmdPayload(ctx *context, args options.Arguments) bool {
	pkgName := args.Get(0).String()
	pkgArch := options.GetS(OPT_ARCH)
	payloadType := "files"

	if args.Has(1) {
		switch args.Get(1).String() {
		case "files", "file", "f", "requires", "req", "reqs", "provides", "prov", "provs":
			payloadType = args.Get(1).String()
		default:
			terminal.Error("Unknown payload type %q", args.Get(1).String())
			return false
		}
	}

	pkg, _, err := ctx.Repo.Info(pkgName, pkgArch)

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	printPackagePayload(pkg, payloadType)

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printPackagePayload prints package payload
func printPackagePayload(pkg *repo.Package, payloadType string) {
	if !rawOutput {
		fmtutil.Separator(true)
		fmtc.NewLine()

		tag := data.SupportedArchs[pkg.ArchFlags.String()].Tag
		color := archColors[pkg.ArchFlags.String()]
		archColoredTag := color + "[" + tag + "]{!}"

		if tag == "" || color == "" {
			archColoredTag = "[unknown]"
		}

		fmtc.Printfn(" ▾ "+archColoredTag+" {*}%s{!} {s-}(%s){!}", pkg.FullName(), pkg.Info.Summary)

		fmtutil.Separator(false)
	}

	switch payloadType {
	case "files", "file", "f":
		if rawOutput {
			printRawPackagePayload(pkg)
		} else {
			printPackageFilesTree(pkg)
		}

	case "requires", "req", "reqs":
		for _, req := range pkg.Info.Requires {
			if rawOutput {
				fmt.Println(formatDepName(req, false))
			} else {
				fmt.Printf(" %s\n", formatDepName(req, true))
			}
		}

	case "provides", "prov", "provs":
		for _, prov := range pkg.Info.Provides {
			if rawOutput {
				fmt.Println(formatDepName(prov, false))
			} else {
				fmt.Printf(" %s\n", formatDepName(prov, true))
			}
		}
	}

	if !rawOutput {
		fmtc.NewLine()
		fmtutil.Separator(true)
	}
}

// printRawPackagePayload prints raw package payload
func printRawPackagePayload(pkg *repo.Package) {
	payload := pkg.Info.Payload

	sort.Sort(payload)

	for _, obj := range pkg.Info.Payload {
		if pkg.ArchFlags == data.ARCH_FLAG_SRC {
			fmt.Println(strings.TrimLeft(obj.Path, "./"))
		} else {
			fmt.Println(obj.Path)
		}
	}
}

// printPackageFilesTree prints files tree
func printPackageFilesTree(pkg *repo.Package) {
	payload := pkg.Info.Payload

	sort.Sort(payload)

	var curDir, nextObjDir string
	var nextObjIsDir bool

	for index, obj := range payload {
		objDir := path.Dir(obj.Path)
		objName := strutil.Exclude(obj.Path, objDir+"/")

		if index+1 < len(payload) {
			nextObjDir = path.Dir(payload[index+1].Path)
			nextObjIsDir = payload[index+1].IsDir
		} else {
			nextObjDir = ""
			nextObjIsDir = false
		}

		if curDir != objDir {
			if nextObjDir != objDir && strings.HasPrefix(nextObjDir, objDir) {
				curDir = objDir
				continue
			}

			if pkg.ArchFlags == data.ARCH_FLAG_SRC && index == 0 {
				fmtc.Printfn(" {*}%s.src.cpio{!}", pkg.FullName())
			} else {
				fmtc.Printfn(" {*}%s{!}", objDir)
			}

			if !obj.IsDir {
				if nextObjDir == objDir {
					if !nextObjIsDir {
						fmtc.Printfn(" {s}├─{!} %s", lscolors.Colorize(objName))
					} else {
						fmtc.Printfn(" {s}└─{!} %s", lscolors.Colorize(objName))
					}
				} else {
					fmtc.Printfn(" {s}└─{!} %s", lscolors.Colorize(objName))
				}
			} else {
				if nextObjDir == objDir {
					fmtc.Printfn(" {s}├─{!} {*}%s{!}", objName)
				} else {
					fmtc.Printfn(" {s}└─{!} {*}%s{!}", objName)
				}
			}

			curDir = objDir
		} else {
			if !obj.IsDir {
				if nextObjDir == objDir {
					fmtc.Printfn(" {s}├─{!} %s", lscolors.Colorize(objName))
				} else {
					fmtc.Printfn(" {s}└─{!} %s", lscolors.Colorize(objName))
				}
			} else {
				if nextObjDir == objDir {
					fmtc.Printfn(" {s}├─{!} {*}%s{!}", objName)
				} else if nextObjDir != objDir+"/"+objName {
					fmtc.Printfn(" {s}└─{!} {*}%s{!}", objName)
				}
			}
		}
	}
}
