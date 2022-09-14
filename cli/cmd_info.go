package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtc/lscolors"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/terminal"
	"github.com/essentialkaos/ek/v12/timeutil"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Maximum number of files in info output
const INFO_MAX_FILES = 30

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdInfo is 'info' command handler
func cmdInfo(ctx *context, args options.Arguments) bool {
	pkgName := args.Get(0).String()
	pkgArch := options.GetS(OPT_ARCH)

	pkg, releaseDate, err := ctx.Repo.Info(pkgName, pkgArch)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	printPackageInfo(ctx.Repo, pkg, releaseDate)

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// printPackageInfo prints all info about package
func printPackageInfo(r *repo.Repository, pkg *repo.Package, releaseDate time.Time) {
	fmtutil.Separator(true, "PACKAGE INFO")
	fmtc.NewLine()

	printPackageBasicInfo(r, pkg, releaseDate)
	printPackagePayloadInfo(pkg.Info.Files)
	printPackageRequiresInfo(pkg.Info.Requires)
	printPackageProvidesInfo(pkg.Info.Provides)
	printPackageChangelogInfo(pkg.Info.Changelog)

	fmtutil.Separator(true)
}

// printPackageBasicInfo prints basic package info
func printPackageBasicInfo(r *repo.Repository, pkg *repo.Package, releaseDate time.Time) {
	fmtc.Printf("{*}%-16s{!}%s\n", "Name", pkg.Name)
	fmtc.Printf("{*}%-16s{!}%s\n", "Summary", pkg.Info.Summary)
	fmtc.Printf("{*}%-14s{!}{s-}%s:{!}%s\n", "Version", pkg.Epoch, pkg.Version)
	fmtc.Printf("{*}%-16s{!}%s\n", "Release", pkg.Release)

	if pkg.Info.Group != "" {
		fmtc.Printf("{*}%-16s{!}%s\n", "Group", pkg.Info.Group)
	}

	if pkg.Info.URL != "" {
		fmtc.Printf("{*}%-16s{!}%s\n", "URL", pkg.Info.URL)
	}

	if pkg.Info.License != "" {
		fmtc.Printf("{*}%-16s{!}%s\n", "License", pkg.Info.License)
	}

	if pkg.Info.Packager != "" {
		fmtc.Printf("{*}%-16s{!}%s\n", "Packager", pkg.Info.Packager)
	}

	if pkg.Info.Vendor != "" {
		fmtc.Printf("{*}%-16s{!}%s\n", "Vendor", pkg.Info.Vendor)
	}

	fmtc.NewLine()

	if !releaseDate.IsZero() {
		fmtc.Printf("{*}%-16s{!}testing release\n", "Repository")
	} else {
		fmtc.Printf("{*}%-16s{!}testing\n", "Repository")
	}

	fmtc.NewLine()

	fmtc.Printf(
		"{*}%-16s{!}%s {s-}(%s){!}\n", "Built",
		timeutil.Format(pkg.Info.DateBuild, "%d/%m/%Y %H:%M"),
		getDaysSinceDate(pkg.Info.DateBuild),
	)

	fmtc.Printf(
		"{*}%-16s{!}%s {s-}(%s){!}\n", "Added",
		timeutil.Format(pkg.Info.DateAdded, "%d/%m/%Y %H:%M"),
		getDaysSinceDate(pkg.Info.DateAdded),
	)

	if !releaseDate.IsZero() {
		fmtc.Printf(
			"{*}%-16s{!}%s {s-}(%s){!}\n", "Released",
			timeutil.Format(releaseDate, "%d/%m/%Y %H:%M"),
			getDaysSinceDate(releaseDate),
		)
	}

	fmtc.NewLine()

	// TODO: Show package checksum
	if len(pkg.Files) != 0 {
		fmtc.Printf("{*}%-16s{!}%s\n", "RPM File", pkg.Files[0].Path)
		fmtc.NewLine()
	}

	if pkg.Src != "" {
		fmtc.Printf("{*}%-16s{!}%s\n", "Source File", pkg.Src)
		fmtc.NewLine()
	}

	fmtc.Printf(
		"{*}%-16s{!}%s\n", "Package size",
		fmtutil.PrettySize(pkg.Info.SizePackage, " "),
	)

	fmtc.Printf(
		"{*}%-16s{!}%s\n", "Payload size",
		fmtutil.PrettySize(pkg.Info.SizeInstalled, " "),
	)

	fmtc.NewLine()
}

// printPackagePayloadInfo prints info about package data
func printPackagePayloadInfo(files repo.PayloadData) {
	if len(files) == 0 {
		return
	}

	for i, obj := range files {
		if i == 0 {
			fmtc.Printf("{*}%-16s{!}%s\n", "Payload", formatPayloadPath(obj.Path))
		} else {
			fmtc.Printf("{*}%-16s{!}%s\n", "", formatPayloadPath(obj.Path))
		}

		if i == INFO_MAX_FILES {
			fmtc.Printf(
				"{*}%-16s{!}{s}+%d more{!}\n", "",
				len(files)-INFO_MAX_FILES,
			)

			fmtc.NewLine()

			return
		}
	}

	fmtc.NewLine()
}

// printPackageRequiresInfo prints info about reqired packages and binaries
func printPackageRequiresInfo(reqs []data.Dependency) {
	if len(reqs) == 0 {
		return
	}

	for i, dep := range reqs {
		if i == 0 {
			fmtc.Printf("{*}%-16s{!}%s\n", "Requires", formatDepName(dep, true))
		} else {
			fmtc.Printf("{*}%-16s{!}%s\n", "", formatDepName(dep, true))
		}
	}

	fmtc.NewLine()
}

// printPackageProvidesInfo prints info about provided packages and binaries
func printPackageProvidesInfo(provs []data.Dependency) {
	if len(provs) == 0 {
		return
	}

	for i, dep := range provs {
		if i == 0 {
			fmtc.Printf("{*}%-16s{!}%s\n", "Provides", formatDepName(dep, true))
		} else {
			fmtc.Printf("{*}%-16s{!}%s\n", "", formatDepName(dep, true))
		}
	}

	fmtc.NewLine()
}

// printPackageChangelogInfo prints info from changelog
func printPackageChangelogInfo(changelog *repo.PackageChangelog) {
	if changelog == nil || len(changelog.Records) == 0 {
		return
	}

	fmtc.Printf("{*}%-16s{!}{s}%s{!}\n", "Changelog", changelog.Author)

	for _, rec := range changelog.Records {
		fmtc.Printf("%-16s%s\n", "", rec)
	}

	fmtc.NewLine()
}

// formatDepName formats provided/reqired package
func formatDepName(dep data.Dependency, pretty bool) string {
	result := dep.Name
	version := dep.Version

	if dep.Epoch != "" && dep.Epoch != "0" {
		version = dep.Epoch + ":" + version
	}

	if dep.Release != "" {
		version += "-" + dep.Release
	}

	if pretty {
		switch dep.Flag {
		case data.COMP_FLAG_EQ:
			result += " = " + version
		case data.COMP_FLAG_LT:
			result += " < " + version
		case data.COMP_FLAG_LE:
			result += " ≤ " + version
		case data.COMP_FLAG_GT:
			result += " > " + version
		case data.COMP_FLAG_GE:
			result += " ≥ " + version
		}
	} else {
		switch dep.Flag {
		case data.COMP_FLAG_EQ:
			result += " = " + version
		case data.COMP_FLAG_LT:
			result += " < " + version
		case data.COMP_FLAG_LE:
			result += " <= " + version
		case data.COMP_FLAG_GT:
			result += " > " + version
		case data.COMP_FLAG_GE:
			result += " >= " + version
		}
	}

	result = strutil.Exclude(result, "()")

	return result
}

// getDaysSinceDate formats duration
func getDaysSinceDate(d time.Time) string {
	dur := time.Since(d)

	switch {
	case dur <= time.Minute:
		return "just now"
	case !d.Before(timeutil.StartOfDay(time.Now())):
		return "today"
	}

	return fmt.Sprintf("%s ago", timeutil.PrettyDurationInDays(dur))
}

// formatPayloadPath formats payload path
func formatPayloadPath(path string) string {
	if strings.HasPrefix(path, "./") {
		path = strings.ReplaceAll(path, "./", "")
	}

	if fmtc.DisableColors {
		return path
	}

	return lscolors.ColorizePath(path)
}
