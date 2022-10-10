package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/pluralize"
	"github.com/essentialkaos/ek/v12/progress"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdSign is 'resign' command handler
func cmdResign(ctx *context, args options.Arguments) bool {
	if !options.GetB(OPT_FORCE) {
		terminal.PrintWarnMessage("▲ This command will re-sign all packages in the repo. This action")
		terminal.PrintWarnMessage("  can take a lot of time (depending on how many packages you have).\n")

		ok, err := terminal.ReadAnswer("Do you really want to re-sign all packages?", "n")

		if err != nil || !ok {
			return false
		}

		fmtc.NewLine()
	}

	privateKey, ok := getRepoPrivateKey(ctx.Repo)

	if !ok {
		return false
	}

	return resignAllPackages(ctx, privateKey)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// resignAllPackages re-singes all packages in testing and release repositories
func resignAllPackages(ctx *context, privateKey *sign.PrivateKey) bool {
	var isResigned bool

	if !resignRepoPackages(ctx, privateKey, ctx.Repo.Testing) {
		return false
	} else {
		isResigned = true
		fmtc.NewLine()
	}

	if !resignRepoPackages(ctx, privateKey, ctx.Repo.Release) {
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

// resignRepoPackages re-signes all packages in given repository
func resignRepoPackages(ctx *context, privateKey *sign.PrivateKey, r *repo.SubRepository) bool {
	stack, err := r.List("", true)

	if err != nil {
		terminal.PrintErrorMessage(err.Error())
		return false
	}

	if stack.IsEmpty() {
		fmtc.Printf("There are no packages in {*}{?repo}%s{!} repository. Nothing to re-sign.\n", r.Name)
		return true
	}

	tmpDir, err := ctx.Temp.MkDir("rep")

	if err != nil {
		terminal.PrintErrorMessage("Can't create temporary directory: %v", err)
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
		isCancelProtected = true

		filePath := r.GetFullPackagePath(file)
		fileName := path.Base(filePath)
		tmpFile := path.Join(tmpDir, fileName)

		err = sign.Sign(filePath, tmpFile, privateKey)

		if err != nil {
			pb.Finish()
			terminal.PrintErrorMessage("Can't re-sign package: %v", err)
			return false
		}

		pb.Add(1)

		if isCanceled {
			pb.Finish()
			return false
		}

		isCancelProtected = false
	}

	pb.Finish()

	return true
}
