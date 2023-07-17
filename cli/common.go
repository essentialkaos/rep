package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/knf"
	"github.com/essentialkaos/ek/v12/lock"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/secstr"
	"github.com/essentialkaos/ek/v12/terminal"
	"github.com/essentialkaos/ek/v12/tmp"

	"github.com/essentialkaos/rep/v3/cli/logger"
	"github.com/essentialkaos/rep/v3/cli/query"
	"github.com/essentialkaos/rep/v3/repo"
	"github.com/essentialkaos/rep/v3/repo/data"
	"github.com/essentialkaos/rep/v3/repo/index"
	"github.com/essentialkaos/rep/v3/repo/sign"
	"github.com/essentialkaos/rep/v3/repo/storage"
	"github.com/essentialkaos/rep/v3/repo/storage/fs"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	FLAG_NONE          uint8 = 0
	FLAG_REQUIRE_CACHE uint8 = 1 << iota // Require cache warming
	FLAG_REQUIRE_LOCK                    // Create and check lock
)

// ////////////////////////////////////////////////////////////////////////////////// //

// handler is function which handle CLI command
type handler func(ctx *context, args options.Arguments) bool

// command contains basic information about command (handler + min args + options)
type command struct {
	Handler handler
	MinArgs int
	Flags   uint8
}

// context is struct which contains all required data for handling CLI command
type context struct {
	Repo   *repo.Repository
	Temp   *tmp.Temp
	Logger *logger.Logger
}

// ////////////////////////////////////////////////////////////////////////////////// //

// archColorsExt contains archs colors tags for terminals with 16-colors support
var archColors = map[string]string{
	data.ARCH_SRC:     "{*}",
	data.ARCH_NOARCH:  "{c*}",
	data.ARCH_I386:    "{m*}",
	data.ARCH_I586:    "{m*}",
	data.ARCH_I686:    "{m*}",
	data.ARCH_X64:     "{y*}",
	data.ARCH_AARCH64: "{y*}",
	data.ARCH_PPC64:   "{y*}",
	data.ARCH_PPC64LE: "{y*}",
	data.ARCH_ARM:     "{g*}",
	data.ARCH_ARMV7HL: "{g*}",
}

// archColorsExt contains archs colors tags for terminals with 256-colors support
var archColorsExt = map[string]string{
	data.ARCH_SRC:     "{*}",
	data.ARCH_NOARCH:  "{*}{#75}",
	data.ARCH_I386:    "{*}{#105}",
	data.ARCH_I586:    "{*}{#144}",
	data.ARCH_I686:    "{*}{#128}",
	data.ARCH_X64:     "{*}{#214}",
	data.ARCH_AARCH64: "{*}{#166}",
	data.ARCH_PPC64:   "{*}{#99}",
	data.ARCH_PPC64LE: "{*}{#105}",
	data.ARCH_ARM:     "{*}{#76}",
	data.ARCH_ARMV7HL: "{*}{#78}",
}

// commands is map [long command → {handler + min args + options}]
var commands = map[string]command{
	COMMAND_INIT:         {cmdInit, 1, FLAG_REQUIRE_LOCK},
	COMMAND_GEN_KEY:      {cmdGenKey, 0, FLAG_NONE},
	COMMAND_LIST:         {cmdList, 0, FLAG_REQUIRE_CACHE},
	COMMAND_WHICH_SOURCE: {cmdWhichSource, 1, FLAG_REQUIRE_CACHE},
	COMMAND_FIND:         {cmdFind, 1, FLAG_REQUIRE_CACHE},
	COMMAND_INFO:         {cmdInfo, 1, FLAG_REQUIRE_CACHE},
	COMMAND_PAYLOAD:      {cmdPayload, 1, FLAG_REQUIRE_CACHE},
	COMMAND_CLEANUP:      {cmdCleanup, 0, FLAG_REQUIRE_CACHE | FLAG_REQUIRE_LOCK},
	COMMAND_CHECK:        {cmdCheck, 0, FLAG_REQUIRE_CACHE},
	COMMAND_SIGN:         {cmdSign, 1, FLAG_NONE},
	COMMAND_RESIGN:       {cmdResign, 0, FLAG_REQUIRE_CACHE | FLAG_REQUIRE_LOCK},
	COMMAND_ADD:          {cmdAdd, 1, FLAG_REQUIRE_LOCK},
	COMMAND_REMOVE:       {cmdRemove, 1, FLAG_REQUIRE_CACHE | FLAG_REQUIRE_LOCK},
	COMMAND_RELEASE:      {cmdRelease, 1, FLAG_REQUIRE_CACHE | FLAG_REQUIRE_LOCK},
	COMMAND_UNRELEASE:    {cmdUnrelease, 1, FLAG_REQUIRE_CACHE | FLAG_REQUIRE_LOCK},
	COMMAND_REINDEX:      {cmdReindex, 0, FLAG_REQUIRE_LOCK},
	COMMAND_PURGE_CACHE:  {cmdPurgeCache, 0, FLAG_REQUIRE_LOCK},
	COMMAND_STATS:        {cmdStats, 0, FLAG_REQUIRE_CACHE},
	COMMAND_HELP:         {cmdHelp, 0, FLAG_NONE},

	"": {cmdList, 0, FLAG_REQUIRE_CACHE}, // default command
}

// commandsShortcurts is map [shortcut → long command]
var commandsShortcurts = map[string]string{
	COMMAND_SHORT_LIST:         COMMAND_LIST,
	COMMAND_SHORT_WHICH_SOURCE: COMMAND_WHICH_SOURCE,
	COMMAND_SHORT_FIND:         COMMAND_FIND,
	COMMAND_SHORT_INFO:         COMMAND_INFO,
	COMMAND_SHORT_PAYLOAD:      COMMAND_PAYLOAD,
	COMMAND_SHORT_CLEANUP:      COMMAND_CLEANUP,
	COMMAND_SHORT_CHECK:        COMMAND_CHECK,
	COMMAND_SHORT_SIGN:         COMMAND_SIGN,
	COMMAND_SHORT_RESIGN:       COMMAND_RESIGN,
	COMMAND_SHORT_ADD:          COMMAND_ADD,
	COMMAND_SHORT_REMOVE:       COMMAND_REMOVE,
	COMMAND_SHORT_RELEASE:      COMMAND_RELEASE,
	COMMAND_SHORT_UNRELEASE:    COMMAND_UNRELEASE,
	COMMAND_SHORT_REINDEX:      COMMAND_REINDEX,
	COMMAND_SHORT_PURGE_CACHE:  COMMAND_PURGE_CACHE,
	COMMAND_SHORT_STATS:        COMMAND_STATS,
	COMMAND_SHORT_HELP:         COMMAND_HELP,
}

// ////////////////////////////////////////////////////////////////////////////////// //

// runCommand runs command
func runCommand(repoCfg *knf.Config, cmdName string, cmdArgs options.Arguments) bool {
	fmtc.If(!rawOutput).NewLine()

	if commandsShortcurts[cmdName] != "" {
		cmdName = commandsShortcurts[cmdName]
	}

	if !checkCommand(cmdName, cmdArgs) {
		return false
	}

	ctx, err := getRepoContext(repoCfg)

	if err != nil {
		terminal.Error(err.Error() + "\n")
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

	if cmd.RequireLock() {
		if !checkForLock() {
			terminal.Error("Can't run command due to lock\n")
			return false
		}

		lock.Create(APP)
		defer lock.Remove(APP)
	}

	if cmd.RequireCache() {
		warmUpCache(ctx.Repo)
	}

	ok = cmd.Handler(ctx, cmdArgs)

	fmtc.If(!rawOutput).NewLine()

	return ok
}

// runSimpleCommand runs some simple commands like help or gen-key
func runSimpleCommand(cmdName string, cmdArgs options.Arguments) bool {
	fmtc.If(!rawOutput).NewLine()

	if !checkCommand(cmdName, cmdArgs) {
		return false
	}

	cmd := commands[cmdName]
	ok := cmd.Handler(nil, cmdArgs)

	fmtc.If(!rawOutput).NewLine()

	return ok
}

// checkCommand checks command before execution
func checkCommand(cmdName string, args options.Arguments) bool {
	cmd, ok := commands[cmdName]

	if !ok {
		return true // Unknown command
	}

	if len(args) < cmd.MinArgs {
		terminal.Error("Command '%s' requires more arguments (at least %d)\n", cmdName, cmd.MinArgs)
		return false
	}

	return true
}

// warmUpCache warms up repository cache if required
func warmUpCache(r *repo.Repository) {
	var warmupTesting, warmupRelease bool

	warmupTesting = r.Testing.IsCacheValid() == false
	warmupRelease = r.Release.IsCacheValid() == false

	if !warmupRelease && !warmupTesting {
		return
	}

	if !options.GetB(OPT_ALL) && !options.GetB(OPT_RELEASE) && options.GetB(OPT_TESTING) {
		warmupRelease, warmupTesting = false, true
	}

	if !options.GetB(OPT_ALL) && !options.GetB(OPT_TESTING) && options.GetB(OPT_RELEASE) {
		warmupRelease, warmupTesting = true, false
	}

	if warmupTesting {
		fmtc.If(!rawOutput && !options.GetB(OPT_PAGER)).TPrintf("{s-}Warming up testing repository cache (it can take a while)…{!}")
		r.Testing.WarmupCache()
	}

	if warmupRelease {
		fmtc.If(!rawOutput && !options.GetB(OPT_PAGER)).TPrintf("{s-}Warming up release repository cache (it can take a while)…{!}")
		r.Release.WarmupCache()
	}

	fmtc.If(!rawOutput && !options.GetB(OPT_PAGER)).TPrintf("")
}

// checkForLock check for lock file
func checkForLock() bool {
	if !lock.Has(APP) {
		return true
	}

	if lock.IsExpired(APP, 5*time.Minute) {
		lock.Remove(APP) // Remove outdated lock file
		return true
	}

	fmtc.If(!rawOutput && !options.GetB(OPT_PAGER)).TPrintf("{s-}Found lock file, waiting for lock to release…{!}")

	ok := lock.Wait(APP, time.Now().Add(5*time.Minute))

	fmtc.If(!rawOutput && !options.GetB(OPT_PAGER)).TPrintf("")

	return ok
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

// checkRPMFiles checks if we have enough permissions to manipulate with RPM files
func checkRPMFiles(files []string) bool {
	var hasErrors bool

	for _, file := range files {
		err := fsutil.ValidatePerms("FRS", file)

		if err != nil {
			terminal.Error(err.Error())
			hasErrors = true
		}
	}

	return hasErrors == false
}

// isSignRequired returns true if some of given files require signing
func isSignRequired(r *repo.SubRepository, files []string) bool {
	if !r.Parent.IsSigningRequired() {
		return false
	}

	// We don't decrypt key, because we can check signature without decrypting
	key, err := r.Parent.SigningKey.Read(nil)

	if err != nil {
		return true
	}

	for _, file := range files {
		isSigned, err := sign.IsPackageSigned(file)

		if err != nil || !isSigned {
			return true
		}

		isSignValid, err := sign.IsPackageSignatureValid(file, key)

		if err != nil || !isSignValid {
			return true
		}
	}

	return false
}

// getRepoSigningKey reads password and decrypts repository private key
func getRepoSigningKey(r *repo.Repository) (*sign.Key, bool) {
	if r.SigningKey == nil {
		terminal.Warn("No signing key defined in configuration file")
		return nil, false
	}

	var err error
	var password *secstr.String

	if r.SigningKey.IsEncrypted {
		password, err = terminal.ReadPasswordSecure("Enter passphrase to unlock the secret key", true)

		if err != nil {
			return nil, false
		}
	}

	fmtc.NewLine()

	key, err := r.SigningKey.Read(password)

	password.Destroy()

	if err != nil {
		terminal.Error("Can't decrypt the secret key (wrong passphrase?)")
		return nil, false
	}

	return key, true
}

// smartPackageSearch uses queary search or simple search based on given command
// arguments
func smartPackageSearch(r *repo.SubRepository, args options.Arguments) (repo.PackageStack, string, error) {
	var err error
	var searchRequest *query.Request
	var stack repo.PackageStack
	var filter string

	if isExtendedSearchRequest(args) {
		searchRequest, err = query.Parse(args.Strings())

		if err != nil {
			return nil, "", err
		}

		stack, err = findPackages(r, searchRequest)
	} else {
		filter = args.Get(0).String()
		stack, err = r.List(filter, true)
	}

	return stack, filter, err
}

// isExtendedSearchRequest returns true if arguments contains search query
func isExtendedSearchRequest(args options.Arguments) bool {
	if len(args) > 1 {
		return true
	}

	if strings.Contains(args.Get(0).String(), ":") {
		return true
	}

	return false
}

// printQueryDebug prints debug search query info
func printQueryDebug(searchRequest *query.Request) {
	for index, term := range searchRequest.Query {
		db, qrs := term.SQL()

		for _, qr := range qrs {
			fmtc.Printf("{s-}{%d|%s} %s → %s{!}\n", index, db, term, qr)
		}
	}

	fmtc.NewLine()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// RequireCache returns true if command requires warm cache
func (c command) RequireCache() bool {
	return c.Flags&FLAG_REQUIRE_CACHE == FLAG_REQUIRE_CACHE
}

// RequireLock returns true if command requires lock
func (c command) RequireLock() bool {
	return c.Flags&FLAG_REQUIRE_LOCK == FLAG_REQUIRE_LOCK
}
