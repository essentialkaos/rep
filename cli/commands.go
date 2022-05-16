package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/knf"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/terminal"
	"github.com/essentialkaos/ek/v12/tmp"

	"github.com/essentialkaos/rep/cli/logger"

	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/index"
	"github.com/essentialkaos/rep/repo/storage"
	"github.com/essentialkaos/rep/repo/storage/fs"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// handler is function which handle CLI command
type handler func(ctx *context, args options.Arguments) bool

// command contains basic information about command (handler + min num of args)
type command struct {
	Handler  handler
	MinArgs  int
	AllowRaw bool
}

// context is struct which contains all required data for handling CLI command
type context struct {
	Repo   *repo.Repository
	Temp   *tmp.Temp
	Logger *logger.Logger
}

// ////////////////////////////////////////////////////////////////////////////////// //

// commands is map [long command → {handler + min args}]
var commands = map[string]command{
	COMMAND_INIT:               {cmdInit, 1, false},
	COMMAND_LIST:               {cmdList, 0, true},
	COMMAND_WHICH_SOURCE:       {cmdWhichSource, 1, false},
	COMMAND_FIND:               {cmdFind, 1, true},
	COMMAND_INFO:               {cmdInfo, 1, false},
	COMMAND_PAYLOAD:            {cmdPayload, 1, true},
	COMMAND_SIGN:               {cmdSign, 1, false},
	COMMAND_ADD:                {cmdAdd, 1, false},
	COMMAND_REMOVE:             {cmdRemove, 1, false},
	COMMAND_RELEASE:            {cmdRelease, 1, false},
	COMMAND_UNRELEASE:          {cmdUnrelease, 1, false},
	COMMAND_REINDEX:            {cmdReindex, 0, false},
	COMMAND_PURGE_CACHE:        {cmdPurgeCache, 0, false},
	COMMAND_STATS:              {cmdStats, 0, false},
	COMMAND_HELP:               {cmdHelp, 1, false},
	COMMAND_SHORT_LIST:         {cmdList, 0, true},
	COMMAND_SHORT_WHICH_SOURCE: {cmdWhichSource, 1, false},
	COMMAND_SHORT_FIND:         {cmdFind, 1, true},
	COMMAND_SHORT_INFO:         {cmdInfo, 1, false},
	COMMAND_SHORT_PAYLOAD:      {cmdPayload, 1, true},
	COMMAND_SHORT_SIGN:         {cmdSign, 0, false},
	COMMAND_SHORT_ADD:          {cmdAdd, 1, false},
	COMMAND_SHORT_REMOVE:       {cmdRemove, 1, false},
	COMMAND_SHORT_RELEASE:      {cmdRelease, 1, false},
	COMMAND_SHORT_UNRELEASE:    {cmdUnrelease, 1, false},
	COMMAND_SHORT_REINDEX:      {cmdReindex, 0, false},
	COMMAND_SHORT_PURGE_CACHE:  {cmdPurgeCache, 0, false},
	COMMAND_SHORT_STATS:        {cmdStats, 0, false},
	COMMAND_SHORT_HELP:         {cmdHelp, 1, false},
	"":                         {cmdList, 0, false}, // default command
}

// ////////////////////////////////////////////////////////////////////////////////// //

// runCommand runs command
func runCommand(repoCfg *knf.Config, cmdName string, cmdArgs options.Arguments) bool {
	if !rawOutput {
		fmtc.NewLine()
	}

	if !checkCommand(cmdName, cmdArgs) {
		return false
	}

	ctx, err := getRepoContext(repoCfg)

	if err != nil {
		terminal.PrintErrorMessage(err.Error() + "\n")
		return false
	}

	defer ctx.Temp.Clean()   // Clean temporary data
	defer ctx.Logger.Flush() // Flush logs

	cmd, ok := commands[cmdName]

	if !ok {
		if len(cmdArgs) != 0 {
			cmd, cmdArgs = commands[COMMAND_FIND], cmdArgs.Unshift(cmdName)
		} else {
			if strings.Contains(cmdName, ":") {
				cmd, cmdArgs = commands[COMMAND_FIND], options.NewArguments(cmdName)
			} else {
				cmd, cmdArgs = commands[COMMAND_LIST], options.NewArguments(cmdName)
			}
		}
	}

	if !cmd.AllowRaw {
		rawOutput = false
	}

	ok = cmd.Handler(ctx, cmdArgs)

	if !rawOutput {
		fmtc.NewLine()
	}

	return ok
}

// checkCommand checks command before execution
func checkCommand(cmdName string, args options.Arguments) bool {
	cmd, ok := commands[cmdName]

	if !ok {
		return true // Unknown command
	}

	if len(args) < cmd.MinArgs {
		terminal.PrintErrorMessage("Command %s requires more arguments (at least %d)\n", cmdName, cmd.MinArgs)
		return false
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getRepoContext generates repository context based on given repository configuration
func getRepoContext(repoCfg *knf.Config) (*context, error) {
	repoStorage, err := getRepoStorage(knf.GetS(STORAGE_TYPE), repoCfg)

	if err != nil {
		return nil, err
	}

	repo, err := repo.NewRepository(repoCfg.GetS(REPOSITORY_NAME), repoStorage)

	if err != nil {
		return nil, err
	}

	repo.FileFilter = repoCfg.GetS(REPOSITORY_FILE_FILTER)
	repo.Replace = repoCfg.GetB(REPOSITORY_REPLACE, true)

	if repoCfg.HasProp(SIGN_KEY) {
		err = repo.ReadSigningKey(repoCfg.GetS(SIGN_KEY))

		if err != nil {
			return nil, err
		}
	}

	temp, err := tmp.NewTemp(knf.GetS(TEMP_DIR))

	if err != nil {
		return nil, err
	}

	logger, err := getCLILogger(repoCfg.GetS(REPOSITORY_NAME))

	if err != nil {
		return nil, err
	}

	return &context{repo, temp, logger}, nil
}

// getRepoStorage configures repository storage
func getRepoStorage(typ string, repoCfg *knf.Config) (storage.Storage, error) {
	switch typ {
	case storage.TYPE_FS:
		return getRepoFSStorage(repoCfg)
	}

	return nil, fmt.Errorf("Unknown storage type %q", typ)
}

// getRepoFSStorage configures new filesystem storage
func getRepoFSStorage(repoCfg *knf.Config) (*fs.Storage, error) {
	return fs.NewStorage(
		&fs.Options{
			DataDir:    path.Join(knf.GetS(STORAGE_DATA), repoCfg.GetS(REPOSITORY_NAME)),
			CacheDir:   path.Join(knf.GetS(STORAGE_CACHE), repoCfg.GetS(REPOSITORY_NAME)),
			SplitFiles: knf.GetB(STORAGE_SPLIT_FILES, false),
			User:       repoCfg.GetS(PERMISSIONS_USER),
			Group:      repoCfg.GetS(PERMISSIONS_GROUP),
			DirPerms:   repoCfg.GetM(PERMISSIONS_DIR),
			FilePerms:  repoCfg.GetM(PERMISSIONS_FILE),
		},
		&index.Options{
			User:           repoCfg.GetS(PERMISSIONS_USER),
			Group:          repoCfg.GetS(PERMISSIONS_GROUP),
			DirPerms:       repoCfg.GetM(PERMISSIONS_DIR),
			FilePerms:      repoCfg.GetM(PERMISSIONS_FILE),
			Pretty:         knf.GetB(INDEX_PRETTY),
			Update:         knf.GetB(INDEX_UPDATE),
			Split:          knf.GetB(INDEX_SPLIT),
			SkipSymlinks:   knf.GetB(INDEX_SKIP_SYMLINKS),
			Deltas:         knf.GetB(INDEX_DELTAS),
			NumDeltas:      knf.GetI(INDEX_NUM_DELTAS),
			MDFilenames:    knf.GetS(INDEX_MD_FILENAMES, index.MDF_SIMPLE),
			CompressType:   knf.GetS(INDEX_REVISION, index.COMPRESSION_BZ2),
			CheckSum:       knf.GetS(INDEX_CHECKSUM, index.CHECKSUM_SHA256),
			ChangelogLimit: knf.GetI(INDEX_CHANGELOG_LIMIT),
			Distro:         knf.GetS(INDEX_DISTRO),
			Content:        knf.GetS(INDEX_CONTENT),
			Revision:       knf.GetS(INDEX_REVISION),
			Workers:        knf.GetI(INDEX_WORKERS, 0),
		},
	)
}

// getCLILogger returns logger for CLI
func getCLILogger(repoName string) (*logger.Logger, error) {
	var err error

	logDir := path.Join(knf.GetS(LOG_DIR), repoName)

	if !fsutil.IsExist(logDir) {
		err = os.Mkdir(logDir, knf.GetM(LOG_DIR_PERMS, 0755))

		if err != nil {
			return nil, fmt.Errorf("Can't create directory for logs: %w", err)
		}
	}

	l := logger.New(logDir, knf.GetM(LOG_FILE_PERMS, 0644))

	err = l.Add(data.REPO_RELEASE)

	if err != nil {
		return nil, fmt.Errorf(
			"Can't create log for repository %s/%s: %w",
			repoName, data.REPO_RELEASE, err,
		)
	}

	err = l.Add(data.REPO_TESTING)

	if err != nil {
		return nil, fmt.Errorf(
			"Can't create log for repository %s/%s: %w",
			repoName, data.REPO_TESTING, err,
		)
	}

	return l, nil
}
