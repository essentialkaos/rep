package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/fmtutil/panel"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/path"
	"github.com/essentialkaos/ek/v13/pluralize"
	"github.com/essentialkaos/ek/v13/progress"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/input"

	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
	"github.com/essentialkaos/rep/v3/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdSign is 'resign' command handler
func cmdResign(ctx *context, args options.Arguments) bool {
	if !options.GetB(OPT_FORCE) {
		panel.Warn(
			"Command can take a lot of time",
			`This command will re-sign all packages in the repo. Re-sign process requires 
rewriting {*}every{!} package in repository and can take a lot of time (depending on 
how many packages you have and how big they are).`,
			panel.INDENT_INNER,
		)

		fmtc.NewLine()

		ok, err := input.ReadAnswer("Do you really want to re-sign all packages?", "n")

		if err != nil || !ok {
			return false
		}
	}

	key, ok := getRepoSigningKey(ctx.Repo)

	if !ok {
		return false
	}

	return resignAllPackages(ctx, key)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// resignAllPackages re-singes all packages in testing and release repositories
func resignAllPackages(ctx *context, key *sign.Key) bool {
	isResigned := false

	if !resignRepoPackages(ctx, key, ctx.Repo.Testing) {
		ctx.Logger.Get(data.REPO_TESTING).Print("Packages re-signing finished with error")
		return false
	} else {
		isResigned = true //nolint:ineffassign
		fmtc.NewLine()
	}

	if !resignRepoPackages(ctx, key, ctx.Repo.Release) {
		ctx.Logger.Get(data.REPO_RELEASE).Print("Packages re-signing finished with error")
		return false
	} else {
		isResigned = true
		fmtc.NewLine()
	}

	if isResigned {
		reindexRepository(ctx, ctx.Repo.Testing, false)
		reindexRepository(ctx, ctx.Repo.Release, false)
	}

	return true
}

// resignRepoPackages re-signs all packages in given repository
func resignRepoPackages(ctx *context, key *sign.Key, r *repo.SubRepository) bool {
	stack, err := r.List("", true)

	ctx.Logger.Get(r.Name).Print("Started packages re-signing")

	if err != nil {
		terminal.Error(err.Error())
		return false
	}

	if stack.IsEmpty() {
		fmtc.Printfn("There are no packages in {*}{?repo}%s{!} repository. Nothing to re-sign.", r.Name)
		return true
	}

	tmpDir, err := ctx.Temp.MkDir("rep")

	if err != nil {
		terminal.Error("Can't create temporary directory: %v", err)
		return false
	}

	files := stack.FlattenFiles()

	fmtc.Printf(
		"Re-signing %s %s in {*}{?repo}%s{!} repository…\n",
		fmtutil.PrettyNum(len(files)),
		pluralize.Pluralize(len(files), "package", "packages"),
		r.Name,
	)

	pb := progress.New(int64(len(files)), "Re-signing")
	pb.Start()

	for _, file := range files {
		isCancelProtected.Store(true)

		filePath := r.GetFullPackagePath(file)
		fileName := path.Base(filePath)
		tmpFile := path.Join(tmpDir, fileName)

		err = sign.SignPackage(filePath, tmpFile, key)

		if err != nil {
			pb.Finish()
			terminal.Error("Can't re-sign package: %v", err)
			return false
		}

		err = replaceSignedRPMFile(filePath, tmpFile)

		if err != nil {
			pb.Finish()
			terminal.Error("Can't re-sign package: %v", err)
			return false
		}

		pb.Add(1)

		if isCanceled.Load() {
			pb.Finish()
			return false
		}

		isCancelProtected.Store(false)
	}

	pb.Finish()

	ctx.Logger.Get(r.Name).Print("Packages re-signing finished with success")

	return true
}
