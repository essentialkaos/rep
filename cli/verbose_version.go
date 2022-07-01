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
func showVerboseAbout(gitRev string, gomod []byte) {
	showApplicationInfo(gitRev)
	showSystemInfo()
	showEnvironmentInfo()
	showDepsInfo(gomod)
	fmtutil.Separator(false)
}

// showApplicationInfo shows verbose information about application
func showApplicationInfo(gitRev string) {
	fmtutil.Separator(false, "APPLICATION INFO")

	fmtc.Printf("  {*}%-12s{!} %s\n", "Name:", APP)
	fmtc.Printf("  {*}%-12s{!} %s\n", "Version:", VER)

	if gitRev != "" {
		if fmtc.IsTrueColorSupported() {
			fmtc.Printf("  {*}%-12s{!} %s {#"+strutil.Head(gitRev, 6)+"}●{!}\n", "Git SHA:", gitRev)
		} else {
			fmtc.Printf("  {*}%-12s{!} %s\n", "Git SHA:", gitRev)
		}
	}
}

// showSystemInfo shows verbose information about system
func showSystemInfo() {
	fmtInfo := func(s string) string {
		if s == "" {
			return fmtc.Sprintf("{s}unknown{!}")
		}

		return s
	}

	osInfo, err := system.GetOSInfo()

	if err == nil {
		fmtutil.Separator(false, "SYSTEM INFO")
		fmtc.Printf("  {*}%-16s{!} %s\n", "OS Name:", fmtInfo(osInfo.Name))
		fmtc.Printf("  {*}%-16s{!} %s\n", "OS Version:", fmtInfo(osInfo.VersionID))
		fmtc.Printf("  {*}%-16s{!} %s\n", "OS ID:", fmtInfo(osInfo.ID))
	}

	systemInfo, err := system.GetSystemInfo()

	if err != nil {
		return
	} else {
		if osInfo == nil {
			fmtutil.Separator(false, "SYSTEM INFO")
			fmtc.Printf("  {*}%-16s{!} %s\n", "OS Name:", fmtInfo(systemInfo.Distribution))
			fmtc.Printf("  {*}%-16s{!} %s\n", "OS Version:", fmtInfo(systemInfo.Version))
		}
	}

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
