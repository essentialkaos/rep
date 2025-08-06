package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/hashutil"
	"github.com/essentialkaos/ek/v13/lscolors"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/strutil"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/timeutil"

	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Maximum number of files in info output
const INFO_MAX_OBJECTS = 30

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdInfo is 'info' command handler
func cmdInfo(ctx *context, args options.Arguments) bool {
	pkgName := args.Get(0).String()
	pkgArch := options.GetS(OPT_ARCH)

	pkg, releaseDate, err := ctx.Repo.Info(pkgName, pkgArch)

	if err != nil {
		terminal.Error(err.Error())
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
	printPackagePayloadInfo(pkg.Info.Payload)
	printPackageRequiresInfo(pkg.Info.Requires)
	printPackageProvidesInfo(pkg.Info.Provides)
	printPackageChangelogInfo(pkg.Info.Changelog)

	fmtutil.Separator(true)
}

// printPackageBasicInfo prints basic package info
func printPackageBasicInfo(r *repo.Repository, pkg *repo.Package, releaseDate time.Time) {
	fmtc.Printfn("{*}%-16s{!}%s", "Name", pkg.Name)
	fmtc.Printfn("{*}%-16s{!}%s", "Summary", pkg.Info.Summary)
	fmtc.Printfn("{*}%-14s{!}{s-}%s:{!}%s", "Version", pkg.Epoch, pkg.Version)
	fmtc.Printfn("{*}%-16s{!}%s", "Release", pkg.Release)

	if pkg.Info.Group != "" {
		fmtc.Printfn("{*}%-16s{!}%s", "Group", pkg.Info.Group)
	}

	if pkg.Info.URL != "" {
		fmtc.Printfn("{*}%-16s{!}%s", "URL", pkg.Info.URL)
	}

	if pkg.Info.License != "" {
		fmtc.Printfn("{*}%-16s{!}%s", "License", pkg.Info.License)
	}

	if pkg.Info.Packager != "" {
		fmtc.Printfn("{*}%-16s{!}%s", "Packager", pkg.Info.Packager)
	}

	if pkg.Info.Vendor != "" {
		fmtc.Printfn("{*}%-16s{!}%s", "Vendor", pkg.Info.Vendor)
	}

	fmtc.NewLine()

	if !releaseDate.IsZero() {
		fmtc.Printfn("{*}%-16s{!}testing release", "Repository")
	} else {
		fmtc.Printfn("{*}%-16s{!}testing", "Repository")
	}

	fmtc.NewLine()

	fmtc.Printfn(
		"{*}%-16s{!}%s {s-}(%s){!}", "Built",
		timeutil.Format(pkg.Info.DateBuild, "%d/%m/%Y %H:%M"),
		getDaysSinceDate(pkg.Info.DateBuild),
	)

	fmtc.Printfn(
		"{*}%-16s{!}%s {s-}(%s){!}", "Added",
		timeutil.Format(pkg.Info.DateAdded, "%d/%m/%Y %H:%M"),
		getDaysSinceDate(pkg.Info.DateAdded),
	)

	if !releaseDate.IsZero() {
		fmtc.Printfn(
			"{*}%-16s{!}%s {s-}(%s){!}", "Released",
			timeutil.Format(releaseDate, "%d/%m/%Y %H:%M"),
			getDaysSinceDate(releaseDate),
		)
	}

	fmtc.NewLine()

	if len(pkg.Files) != 0 {
		fmtc.Printfn(
			"{*}%-16s{!}%s", "RPM File",
			getPackageFileInfoWithMark(r, pkg.Files[0], !releaseDate.IsZero()),
		)
		fmtc.Printfn(
			"{*}%-16s{!}%s", "Checksum",
			getPackageFileCRCWithMark(r, pkg.Files[0], !releaseDate.IsZero()),
		)
		fmtc.NewLine()
	}

	if pkg.Src != "" {
		fmtc.Printfn("{*}%-16s{!}%s", "Source File", pkg.Src)
		fmtc.NewLine()
	}

	fmtc.Printfn(
		"{*}%-16s{!}%s", "Package size",
		fmtutil.PrettySize(pkg.Info.SizePackage, " "),
	)

	fmtc.Printfn(
		"{*}%-16s{!}%s", "Payload size",
		fmtutil.PrettySize(pkg.Info.SizeInstalled, " "),
	)

	fmtc.NewLine()
}

// printPackagePayloadInfo prints info about package data
func printPackagePayloadInfo(payload repo.PackagePayload) {
	if len(payload) == 0 {
		return
	}

	for i, obj := range payload {
		if i == 0 {
			fmtc.Printfn("{*}%-16s{!}%s", "Payload", formatPayloadPath(obj.Path))
		} else {
			fmtc.Printfn("{*}%-16s{!}%s", "", formatPayloadPath(obj.Path))
		}

		if i == INFO_MAX_OBJECTS {
			fmtc.Printfn(
				"{*}%-16s{!}{s}+%d more{!}", "",
				len(payload)-INFO_MAX_OBJECTS,
			)

			fmtc.NewLine()

			return
		}
	}

	fmtc.NewLine()
}

// printPackageRequiresInfo prints info about required packages and binaries
func printPackageRequiresInfo(reqs []data.Dependency) {
	if len(reqs) == 0 {
		return
	}

	for i, dep := range reqs {
		if i == 0 {
			fmtc.Printfn("{*}%-16s{!}%s", "Requires", formatDepName(dep, true))
		} else {
			fmtc.Printfn("{*}%-16s{!}%s", "", formatDepName(dep, true))
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
			fmtc.Printfn("{*}%-16s{!}%s", "Provides", formatDepName(dep, true))
		} else {
			fmtc.Printfn("{*}%-16s{!}%s", "", formatDepName(dep, true))
		}
	}

	fmtc.NewLine()
}

// printPackageChangelogInfo prints info from changelog
func printPackageChangelogInfo(changelog *repo.PackageChangelog) {
	if changelog == nil || len(changelog.Records) == 0 {
		return
	}

	fmtc.Printfn(
		"{*}%-16s{!}{*s}%s{!} {s}%s{!}", "Changelog",
		timeutil.Format(changelog.Date, "%a %b %d %Y"),
		changelog.Author,
	)

	for _, rec := range changelog.Records {
		fmtc.Printfn("%-16s%s", "", rec)
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

	return fmt.Sprintf("%s ago", timeutil.Pretty(dur).InDays())
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

// getPackageFileInfoWithMark returns status mark for package file
func getPackageFileInfoWithMark(r *repo.Repository, pkgFile repo.PackageFile, isReleased bool) string {
	testingFile := r.Testing.GetFullPackagePath(pkgFile)

	if !fsutil.IsExist(testingFile) {
		return fmtc.Sprintf("%s {r}✖ {!} {s-}(No file on disk in testing repository){!}", pkgFile.Path)
	}

	if isReleased {
		releaseFile := r.Release.GetFullPackagePath(pkgFile)

		if !fsutil.IsExist(releaseFile) {
			return fmtc.Sprintf("%s {r}✖ {!} {s-}(No file on disk in release repository){!}", pkgFile.Path)
		}
	}

	return fmtc.Sprintf("%s {g}✔ {!}", pkgFile.Path)
}

// getPackageFileCRCWithMark returns status mark for package file
func getPackageFileCRCWithMark(r *repo.Repository, pkgFile repo.PackageFile, isReleased bool) string {
	testingFile := r.Testing.GetFullPackagePath(pkgFile)

	if !fsutil.IsExist(testingFile) {
		return fmtc.Sprintf("%s", pkgFile.Path)
	}

	hasher := sha256.New()

	testingHash := strutil.Head(hashutil.File(testingFile, hasher).String(), 7)

	if testingHash != pkgFile.CRC {
		return fmtc.Sprintf("%s {r}✖ {!} {s-}(CRC mismatch in testing repository){!}", pkgFile.Path)
	}

	if isReleased {
		releaseFile := r.Release.GetFullPackagePath(pkgFile)

		if !fsutil.IsExist(releaseFile) {
			return fmtc.Sprintf("%s", pkgFile.Path)
		}

		releaseHash := strutil.Head(hashutil.File(releaseFile, hasher).String(), 7)

		if releaseHash != pkgFile.CRC {
			return fmtc.Sprintf("%s {r}✖ {!} {s-}(CRC mismatch in release repository){!}", pkgFile.Path)
		}
	}

	return fmtc.Sprintf("%s {g}✔ {!}", pkgFile.CRC)
}
