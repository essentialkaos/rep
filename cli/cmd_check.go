package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"

	"github.com/essentialkaos/ek/v12/errutil"
	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/hash"
	"github.com/essentialkaos/ek/v12/mathutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/pluralize"
	"github.com/essentialkaos/ek/v12/progress"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/system"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdCheck is 'check' command handler
func cmdCheck(ctx *context, args options.Arguments) bool {
	releaseStack, err := ctx.Repo.Release.List("", true)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	testingStack, err := ctx.Repo.Testing.List("", true)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if releaseStack.IsEmpty() && testingStack.IsEmpty() {
		terminal.PrintWarnMessage("Release and testing repositories are empty")
		return false
	}

	if !checkRepositoriesData(ctx.Repo, releaseStack, testingStack) {
		return false
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// checkRepositoriesConsistency checks consistency between testing and release
// repositories
func checkRepositoriesData(r *repo.Repository, releaseStack, testingStack repo.PackageStack) bool {
	var hasProblems bool

	releaseIndex := createIndexForStack(releaseStack)
	testingIndex := createIndexForStack(testingStack)

	if !checkRepositoriesConsistency(releaseIndex, testingIndex) {
		hasProblems = true
	}

	if !checkRepositoriesCRCInfo(r, releaseIndex, testingIndex) {
		hasProblems = true
	}

	if !checkRepositoriesPermissions(r, releaseIndex, testingIndex) {
		hasProblems = true
	}

	if !checkRepositoriesSignatures(r, releaseIndex, testingIndex) {
		hasProblems = true
	}

	return hasProblems == false
}

// checkRepositoriesConsistency check consistency between release and testing repositories
func checkRepositoriesConsistency(releaseIndex, testingIndex map[string]*repo.Package) bool {
	var errs errutil.Errors

	fmtc.Println("{*}[1/4]{!} Checking consistency between {?repo}testing{!} and {?repo}release{!} repository…")

	switch {
	case len(releaseIndex) == 0:
		terminal.PrintWarnMessage("Release repository is empty, skipping check…")
	case len(testingIndex) == 0:
		terminal.PrintWarnMessage("Testing repository is empty, skipping check…")
	}

	for pkgName, testingPkg := range testingIndex {
		releasePkg := releaseIndex[pkgName]

		if releasePkg == nil {
			continue
		}

		if len(testingPkg.Files) != len(releasePkg.Files) {
			errs.Add(fmt.Errorf(
				"Package %s contains different number of files in release (%d) and testing (%d) repositories",
				pkgName, len(releasePkg.Files), len(testingPkg.Files),
			))
		}

		for fileIndex := range testingPkg.Files {
			if testingPkg.Files[fileIndex].CRC != releasePkg.Files[fileIndex].CRC {
				errs.Add(fmt.Errorf(
					"Package %s contains file %s with different checksum in release (%s) and testing (%s) repositories",
					pkgName, releasePkg.Files[fileIndex].Path, releasePkg.Files[fileIndex].CRC, testingPkg.Files[fileIndex].CRC,
				))
			}
		}
	}

	if !printCheckErrorsInfo(errs) {
		return false
	}

	return true
}

// checkRepositoriesCRCInfo validates checksum info
func checkRepositoriesCRCInfo(r *repo.Repository, releaseIndex, testingIndex map[string]*repo.Package) bool {
	var errs errutil.Errors

	fmtc.Println("\n{*}[2/4]{!} Validating checksum data…")

	totalPackages := len(releaseIndex) + len(testingIndex)
	pb := progress.New(int64(totalPackages), "")
	pb.Start()

	if len(releaseIndex) != 0 {
		errs.Add(checkRepositoryCRCInfo(pb, r.Release, releaseIndex))
	}

	if len(testingIndex) != 0 {
		errs.Add(checkRepositoryCRCInfo(pb, r.Testing, testingIndex))
	}

	pb.Finish()

	if !printCheckErrorsInfo(errs) {
		return false
	}

	return true
}

// checkRepositoryCRCInfo validates checksum for all repository files
func checkRepositoryCRCInfo(pb *progress.Bar, r *repo.SubRepository, index map[string]*repo.Package) errutil.Errors {
	var errs errutil.Errors

	for name, pkg := range index {
		for _, file := range pkg.Files {
			filePath := r.GetFullPackagePath(file)
			fileCRC := strutil.Head(hash.FileHash(filePath), 7)

			if fileCRC != file.CRC {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s with checksum mismatch between DB (%s) data and file on disk (%s)",
					name, r.Name, file.Path, file.CRC, fileCRC,
				))
			}
		}

		pb.Add(1)
	}

	return errs
}

// checkRepositoriesPermissions checks packages permissions in release and testing repositories
func checkRepositoriesPermissions(r *repo.Repository, releaseIndex, testingIndex map[string]*repo.Package) bool {
	var errs errutil.Errors

	fmtc.Println("\n{*}[3/4]{!} Validating packages permissions…")

	totalPackages := len(releaseIndex) + len(testingIndex)
	pb := progress.New(int64(totalPackages), "")
	pb.Start()

	if len(testingIndex) != 0 {
		errs.Add(checkRepositoryPermissions(pb, r.Testing, testingIndex))
	}

	if len(releaseIndex) != 0 {
		errs.Add(checkRepositoryPermissions(pb, r.Release, releaseIndex))
	}

	pb.Finish()

	if !printCheckErrorsInfo(errs) {
		return false
	}

	return true
}

// checkRepositoryPermissions checks packages permissions in given repository
func checkRepositoryPermissions(pb *progress.Bar, r *repo.SubRepository, index map[string]*repo.Package) errutil.Errors {
	var errs errutil.Errors

	userID, groupID := -1, -1

	repoCfg := configs[r.Parent.Name]
	user := repoCfg.GetS(PERMISSIONS_USER)
	group := repoCfg.GetS(PERMISSIONS_GROUP)
	perms := repoCfg.GetM(PERMISSIONS_FILE)

	if user != "" {
		userInfo, err := system.LookupUser(user)

		if err != nil {
			errs.Add(err)
			return errs
		}

		userID = userInfo.UID
	}

	if group != "" {
		groupInfo, err := system.LookupGroup(group)

		if err != nil {
			errs.Add(err)
			return errs
		}

		groupID = groupInfo.GID
	}

	for name, pkg := range index {
		for _, file := range pkg.Files {
			filePath := r.GetFullPackagePath(file)
			fileUID, fileGID, err := fsutil.GetOwner(filePath)

			if err != nil {
				errs.Add(fmt.Errorf(
					"Error while checking package %s permissions in %s repository for file %s: %v",
					name, r.Name, file.Path, err,
				))

				continue
			}

			if userID != -1 && fileUID != userID {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s width different owner UID (%d ≠ %d)",
					name, r.Name, file.Path, fileUID, userID,
				))

				continue
			}

			if groupID != -1 && fileGID != groupID {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s width different owner GID (%d ≠ %d)",
					name, r.Name, file.Path, fileGID, groupID,
				))

				continue
			}

			filePerms := fsutil.GetMode(filePath)

			if perms != 0 && filePerms != perms {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s width different permissions (%s ≠ %s)",
					name, r.Name, file.Path, filePerms, perms,
				))

				continue
			}
		}

		pb.Add(1)
	}

	return errs
}

// checkRepositoriesSignatures checks packages signatures in release and testing repositories
func checkRepositoriesSignatures(r *repo.Repository, releaseIndex, testingIndex map[string]*repo.Package) bool {
	var errs errutil.Errors

	fmtc.Println("\n{*}[4/4]{!} Validating packages signatures…")

	privateKey, err := r.SigningKey.Get(nil)

	if err != nil {
		terminal.PrintErrorMessage("Can't read signing key: %v", err)
		return false
	}

	totalPackages := len(releaseIndex) + len(testingIndex)

	pb := progress.New(int64(totalPackages), "")
	pb.Start()

	if len(testingIndex) != 0 {
		errs.Add(checkRepositorySignatures(pb, r.Testing, privateKey, testingIndex))
	}

	if len(releaseIndex) != 0 {
		errs.Add(checkRepositorySignatures(pb, r.Release, privateKey, releaseIndex))
	}

	pb.Finish()

	if !printCheckErrorsInfo(errs) {
		return false
	}

	return true
}

// checkRepositorySignatures checks packages signatures in given repository
func checkRepositorySignatures(pb *progress.Bar, r *repo.SubRepository, privateKey *sign.PrivateKey, index map[string]*repo.Package) errutil.Errors {
	var errs errutil.Errors

	for name, pkg := range index {
		for _, file := range pkg.Files {
			filePath := r.GetFullPackagePath(file)
			hasSignature, err := sign.HasSignature(filePath)

			if err != nil {
				errs.Add(fmt.Errorf(
					"Error while checking package %s signature in %s repository for file %s: %v",
					name, r.Name, file.Path, err,
				))

				continue
			}

			if !hasSignature {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s without signature",
					name, r.Name, file.Path,
				))

				continue
			}

			isSigned, err := sign.IsSigned(filePath, privateKey)

			if err != nil {
				errs.Add(fmt.Errorf(
					"Error while checking package %s signature in %s repository for file %s: %v",
					name, r.Name, file.Path, err,
				))

				continue
			}

			if !isSigned {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s signed with different key",
					name, r.Name, file.Path,
				))

				continue
			}
		}

		pb.Add(1)
	}

	return errs
}

// printCheckErrorsInfo prints info about check errors
func printCheckErrorsInfo(errs errutil.Errors) bool {
	if !errs.HasErrors() {
		fmtc.Println("{g}No problems found{!}")
		return true
	}

	errsList := errs.All()

	terminal.PrintErrorMessage(
		"Found %s %s. First 20 problems:\n",
		fmtutil.PrettyNum(errs.Num()),
		pluralize.Pluralize(errs.Num(), "problem", "problems"),
	)

	for i := 0; i < mathutil.Min(errs.Num(), 20); i++ {
		terminal.PrintErrorMessage(" • %v", errsList[i])
	}

	return false
}

// createIndexForStack creates index map fullname→package from stack
func createIndexForStack(stack repo.PackageStack) map[string]*repo.Package {
	result := map[string]*repo.Package{}

	for _, bundle := range stack {
		for _, pkg := range bundle {
			result[pkg.FullName()] = pkg
		}
	}

	return result
}