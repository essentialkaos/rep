package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/hash"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/system"

	"github.com/essentialkaos/depsy"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// showVerboseAbout prints verbose info about app
func showVerboseAbout(gitRev string, gomod []byte) {
	showApplicationInfo(gitRev)
	showOSInfo()
	showEnvironmentInfo()
	showDepsInfo(gomod)
	fmtutil.Separator(false)
}

// showApplicationInfo shows verbose information about application
func showApplicationInfo(gitRev string) {
	fmtutil.Separator(false, "APPLICATION INFO")

	fmtc.Printf("  {*}%-12s{!} %s\n", "Name:", APP)
	fmtc.Printf("  {*}%-12s{!} %s\n", "Version:", VER)

	fmtc.Printf(
		"  {*}%-12s{!} %s {s}(%s/%s){!}\n", "Go:",
		strings.TrimLeft(runtime.Version(), "go"),
		runtime.GOOS, runtime.GOARCH,
	)

	if gitRev != "" {
		if !fmtc.DisableColors && fmtc.IsTrueColorSupported() {
			fmtc.Printf("  {*}%-12s{!} %s {#"+strutil.Head(gitRev, 6)+"}●{!}\n", "Git SHA:", gitRev)
		} else {
			fmtc.Printf("  {*}%-12s{!} %s\n", "Git SHA:", gitRev)
		}
	}

	bin, _ := os.Executable()
	binSHA := hash.FileHash(bin)

	if binSHA != "" {
		binSHA = strutil.Head(binSHA, 7)
		if !fmtc.DisableColors && fmtc.IsTrueColorSupported() {
			fmtc.Printf("  {*}%-12s{!} %s {#"+strutil.Head(binSHA, 6)+"}●{!}\n", "Bin SHA:", binSHA)
		} else {
			fmtc.Printf("  {*}%-12s{!} %s\n", "Bin SHA:", binSHA)
		}
	}
}

// showOSInfo shows verbose information about system
func showOSInfo() {
	fmtInfo := func(s string) string {
		if s == "" {
			return fmtc.Sprintf("{s}unknown{!}")
		}

		return s
	}

	osInfo, err := system.GetOSInfo()

	if err == nil {
		fmtutil.Separator(false, "OS INFO")
		fmtc.Printf("  {*}%-16s{!} %s\n", "Name:", fmtInfo(osInfo.Name))
		fmtc.Printf("  {*}%-16s{!} %s\n", "Pretty Name:", fmtInfo(osInfo.PrettyName))
		fmtc.Printf("  {*}%-16s{!} %s\n", "Version:", fmtInfo(osInfo.VersionID))
		fmtc.Printf("  {*}%-16s{!} %s\n", "ID:", fmtInfo(osInfo.ID))
		fmtc.Printf("  {*}%-16s{!} %s\n", "ID Like:", fmtInfo(osInfo.IDLike))
		fmtc.Printf("  {*}%-16s{!} %s\n", "Version ID:", fmtInfo(osInfo.VersionID))
		fmtc.Printf("  {*}%-16s{!} %s\n", "Version Code:", fmtInfo(osInfo.VersionCodename))
		fmtc.Printf("  {*}%-16s{!} %s\n", "CPE:", fmtInfo(osInfo.CPEName))
	}

	systemInfo, err := system.GetSystemInfo()

	if err != nil {
		return
	} else {
		if osInfo == nil {
			fmtutil.Separator(false, "SYSTEM INFO")
			fmtc.Printf("  {*}%-16s{!} %s\n", "Name:", fmtInfo(systemInfo.OS))
			fmtc.Printf("  {*}%-16s{!} %s\n", "Version:", fmtInfo(systemInfo.Version))
		}
	}

	fmtc.Printf("  {*}%-16s{!} %s\n", "Arch:", fmtInfo(systemInfo.Arch))
	fmtc.Printf("  {*}%-16s{!} %s\n", "Kernel:", fmtInfo(systemInfo.Kernel))

	containerEngine := "No"

	switch {
	case fsutil.IsExist("/.dockerenv"):
		containerEngine = "Yes (Docker)"
	case fsutil.IsExist("/run/.containerenv"):
		containerEngine = "Yes (Podman)"
	}

	fmtc.NewLine()
	fmtc.Printf("  {*}%-16s{!} %s\n", "Container:", containerEngine)
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
