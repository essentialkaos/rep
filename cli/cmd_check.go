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
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/pluralize"
	"github.com/essentialkaos/ek/v12/progress"
	"github.com/essentialkaos/ek/v12/sortutil"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/system"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// checkMaxErrNum is minimal number of check errors to print
var checkMaxErrNum int

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdCheck is 'check' command handler
func cmdCheck(ctx *context, args options.Arguments) bool {
	checkMaxErrNum, _ = args.Get(0).Int()
	checkMaxErrNum = mathutil.Between(checkMaxErrNum, 20, 99999)

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

	if !waitForUserToContinue() {
		return false
	}

	if !checkRepositoriesCRCInfo(r, releaseIndex, testingIndex) {
		hasProblems = true
	}

	if !waitForUserToContinue() {
		return false
	}

	if !checkRepositoriesPermissions(r, releaseIndex, testingIndex) {
		hasProblems = true
	}

	if !waitForUserToContinue() {
		return false
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

	for _, pkgName := range getSortedPackageIndexKeys(testingIndex) {
		testingPkg := testingIndex[pkgName]
		releasePkg := releaseIndex[pkgName]

		if releasePkg == nil {
			continue
		}

		if len(testingPkg.Files) != len(releasePkg.Files) {
			errs.Add(fmt.Errorf(
				"Package %s contains different number of files in release (%d) and testing (%d) repositories",
				pkgName, len(releasePkg.Files), len(testingPkg.Files),
			))
			continue
		}

		for fileIndex := range testingPkg.Files {
			if testingPkg.Files[fileIndex].CRC != releasePkg.Files[fileIndex].CRC {
				errs.Add(fmt.Errorf(
					"Package %s contains file %s with different checksum in release (%s) and testing (%s) repositories",
					pkgName, releasePkg.Files[fileIndex].Path, releasePkg.Files[fileIndex].CRC, testingPkg.Files[fileIndex].CRC,
				))
				continue
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

	for _, pkgName := range getSortedPackageIndexKeys(index) {
		for _, file := range index[pkgName].Files {
			filePath := r.GetFullPackagePath(file)
			fileCRC := strutil.Head(hash.FileHash(filePath), 7)

			if fileCRC != file.CRC {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s with checksum mismatch between DB (%s) data and file on disk (%s)",
					pkgName, r.Name, file.Path, file.CRC, fileCRC,
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

	fmtc.Println("\n{*}[3/4]{!} Validating permissions…")

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
	filePerms := repoCfg.GetM(PERMISSIONS_FILE)
	dirPerms := repoCfg.GetM(PERMISSIONS_DIR)

	checkedDirs := map[string]bool{}

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

	for _, pkgName := range getSortedPackageIndexKeys(index) {
		for _, file := range index[pkgName].Files {
			filePath := r.GetFullPackagePath(file)
			fileUID, fileGID, err := fsutil.GetOwner(filePath)

			if err != nil {
				errs.Add(fmt.Errorf(
					"Error while checking package %s permissions in %s repository for file %s: %v",
					pkgName, r.Name, file.Path, err,
				))

				continue
			}

			if userID != -1 && fileUID != userID {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s width wrong owner UID (%d ≠ %d)",
					pkgName, r.Name, file.Path, fileUID, userID,
				))

				continue
			}

			if groupID != -1 && fileGID != groupID {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s width wrong owner GID (%d ≠ %d)",
					pkgName, r.Name, file.Path, fileGID, groupID,
				))

				continue
			}

			pkgFilePerms := fsutil.GetMode(filePath)
			pkgFileDir := path.Dir(filePath)

			if filePerms != 0 && pkgFilePerms != 0 && pkgFilePerms != filePerms {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s width wrong permissions (%s ≠ %s)",
					pkgName, r.Name, file.Path, pkgFilePerms, filePerms,
				))

				continue
			}

			if !checkedDirs[pkgFileDir] {
				pkgDirPerms := fsutil.GetMode(pkgFileDir)

				checkedDirs[pkgFileDir] = true

				if dirPerms != 0 && pkgDirPerms != 0 && pkgDirPerms != dirPerms {
					errs.Add(fmt.Errorf(
						"Repository %s contains directory %s width wrong permissions (%s ≠ %s)",
						r.Name, pkgFileDir, pkgDirPerms, dirPerms,
					))

					continue
				}
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

	key, err := r.SigningKey.Read(nil)

	if err != nil {
		terminal.PrintErrorMessage("Can't read signing key: %v", err)
		return false
	}

	totalPackages := len(releaseIndex) + len(testingIndex)

	pb := progress.New(int64(totalPackages), "")
	pb.Start()

	if len(testingIndex) != 0 {
		errs.Add(checkRepositorySignatures(pb, r.Testing, key, testingIndex))
	}

	if len(releaseIndex) != 0 {
		errs.Add(checkRepositorySignatures(pb, r.Release, key, releaseIndex))
	}

	pb.Finish()

	if !printCheckErrorsInfo(errs) {
		return false
	}

	return true
}

// checkRepositorySignatures checks packages signatures in given repository
func checkRepositorySignatures(pb *progress.Bar, r *repo.SubRepository, key *sign.Key, index map[string]*repo.Package) errutil.Errors {
	var errs errutil.Errors

	for _, pkgName := range getSortedPackageIndexKeys(index) {
		for _, file := range index[pkgName].Files {
			filePath := r.GetFullPackagePath(file)
			hasSign, err := sign.IsPackageSigned(filePath)

			if err != nil {
				errs.Add(fmt.Errorf(
					"Error while checking package %s signature in %s repository for file %s: %v",
					pkgName, r.Name, file.Path, err,
				))

				continue
			}

			if !hasSign {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s without signature",
					pkgName, r.Name, file.Path,
				))

				continue
			}

			isSignValid, err := sign.IsPackageSignatureValid(filePath, key)

			if err != nil {
				errs.Add(fmt.Errorf(
					"Error while checking package %s signature in %s repository for file %s: %v",
					pkgName, r.Name, file.Path, err,
				))

				continue
			}

			if !isSignValid {
				errs.Add(fmt.Errorf(
					"Package %s in %s repository contains file %s signed with different key",
					pkgName, r.Name, file.Path,
				))

				continue
			}
		}

		pb.Add(1)
	}

	return errs
}

// getSortedPackageIndexKeys reads keys from index and returns sorted slice of keys
func getSortedPackageIndexKeys(index map[string]*repo.Package) []string {
	var result []string

	for pkgName := range index {
		result = append(result, pkgName)
	}

	sortutil.StringsNatural(result)

	return result
}

// printCheckErrorsInfo prints info about check errors
func printCheckErrorsInfo(errs errutil.Errors) bool {
	if !errs.HasErrors() {
		fmtc.Println("{g}No problems found{!}")
		return true
	}

	errsList := errs.All()

	terminal.PrintErrorMessage(
		"Found %s %s. First %s %s:\n",
		fmtutil.PrettyNum(errs.Num()),
		pluralize.Pluralize(errs.Num(), "problem", "problems"),
		fmtutil.PrettyNum(checkMaxErrNum),
		pluralize.Pluralize(checkMaxErrNum, "problem", "problems"),
	)

	for i := 0; i < mathutil.Min(errs.Num(), checkMaxErrNum); i++ {
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

// waitForUserToContinue blocks execution waiting user input
func waitForUserToContinue() bool {
	if options.GetB(OPT_FORCE) {
		return true
	}

	fmtc.NewLine()
	ok, _ := terminal.ReadAnswer("Continue?", "Y")

	return ok
}
