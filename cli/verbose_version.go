package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os/exec"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/system"

	"github.com/essentialkaos/depsy"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// showVerboseAbout prints verbose info about app
func showVerboseAbout(gitrev string, gomod []byte) {
	showApplicationInfo(gitrev)
	showSystemInfo()
	showEnvironmentInfo()
	showDepsInfo(gomod)
	fmtutil.Separator(false)
}

// showApplicationInfo shows verbose information about application
func showApplicationInfo(gitrev string) {
	fmtutil.Separator(false, "APPLICATION INFO")

	fmtc.Printf("  {*}%-12s{!} %s\n", "Name:", APP)
	fmtc.Printf("  {*}%-12s{!} %s\n", "Version:", VER)

	if REL != "" {
		fmtc.Printf("  {*}%-12s{!} %s\n", "Release:", REL)
	}

	if gitrev != "" {
		fmtc.Printf("  {*}%-12s{!} %s\n", "Git SHA:", gitrev)
	}
}

// showSystemInfo shows verbose information about system
func showSystemInfo() {
	systemInfo, err := system.GetSystemInfo()

	if err != nil {
		return
	}

	fmtutil.Separator(false, "SYSTEM INFO")

	fmtInfo := func(s string) string {
		if s == "" {
			return fmtc.Sprintf("{s}unknown{!}")
		}

		return s
	}

	fmtc.Printf("  {*}%-16s{!} %s\n", "OS Name:", fmtInfo(systemInfo.Distribution))
	fmtc.Printf("  {*}%-16s{!} %s\n", "OS Version:", fmtInfo(systemInfo.Version))
	fmtc.Printf("  {*}%-16s{!} %s\n", "Arch:", fmtInfo(systemInfo.Arch))
	fmtc.Printf("  {*}%-16s{!} %s\n", "Kernel:", fmtInfo(systemInfo.Kernel))
}

// showEnvironmentInfo shows info about environment
func showEnvironmentInfo() {
	fmtutil.Separator(false, "ENVIRONMENT")

	cmd := exec.Command("createrepo_c", "--version")
	out, err := cmd.Output()

	if err != nil {
		fmtc.Printf("  {*}%-16s{!} {s}not installed{!}\n", "createrepo_c:")
		return
	}

	createrepoVersion := string(out)
	createrepoVersion = strings.Trim(createrepoVersion, "\n\r")
	createrepoVersion = strutil.Exclude(createrepoVersion, "Version: ")
	createrepoVersion = strings.ReplaceAll(createrepoVersion, " )", ")")

	fmtc.Printf("  {*}%-16s{!} %s\n", "createrepo_c:", createrepoVersion)
}

// showDepsInfo shows information about all dependencies
func showDepsInfo(gomod []byte) {
	deps := depsy.Extract(gomod, false)

	if len(deps) == 0 {
		return
	}

	fmtutil.Separator(false, "DEPENDENCIES")

	for _, dep := range deps {
		if dep.Extra == "" {
			fmtc.Printf(" {s}%8s{!}  %s\n", dep.Version, dep.Path)
		} else {
			fmtc.Printf(" {s}%8s{!}  %s {s-}(%s){!}\n", dep.Version, dep.Path, dep.Extra)
		}
	}
}