package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"

	"github.com/essentialkaos/ek/v13/errors"
	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fmtutil"
	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/knf"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/pager"
	"github.com/essentialkaos/ek/v13/progress"
	"github.com/essentialkaos/ek/v13/signal"
	"github.com/essentialkaos/ek/v13/sortutil"
	"github.com/essentialkaos/ek/v13/spinner"
	"github.com/essentialkaos/ek/v13/support"
	"github.com/essentialkaos/ek/v13/support/apps"
	"github.com/essentialkaos/ek/v13/support/deps"
	"github.com/essentialkaos/ek/v13/system"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/input"
	"github.com/essentialkaos/ek/v13/terminal/tty"
	"github.com/essentialkaos/ek/v13/usage"
	"github.com/essentialkaos/ek/v13/usage/completion/bash"
	"github.com/essentialkaos/ek/v13/usage/completion/fish"
	"github.com/essentialkaos/ek/v13/usage/completion/zsh"
	"github.com/essentialkaos/ek/v13/usage/man"
	"github.com/essentialkaos/ek/v13/usage/update"

	knfv "github.com/essentialkaos/ek/v13/knf/validators"
	knff "github.com/essentialkaos/ek/v13/knf/validators/fs"
	knfr "github.com/essentialkaos/ek/v13/knf/validators/regexp"
	knfs "github.com/essentialkaos/ek/v13/knf/validators/system"

	"github.com/essentialkaos/rep/v3/repo/index"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// App info
const (
	APP  = "rep"
	VER  = "3.5.9"
	DESC = "Package repository management tool"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Commands
const (
	COMMAND_INIT         = "init"
	COMMAND_GEN_KEY      = "gen-key"
	COMMAND_LIST         = "list"
	COMMAND_WHICH_SOURCE = "which-source"
	COMMAND_FIND         = "find"
	COMMAND_INFO         = "info"
	COMMAND_PAYLOAD      = "payload"
	COMMAND_CLEANUP      = "cleanup"
	COMMAND_CHECK        = "check"
	COMMAND_SIGN         = "sign"
	COMMAND_RESIGN       = "resign"
	COMMAND_ADD          = "add"
	COMMAND_REMOVE       = "remove"
	COMMAND_RELEASE      = "release"
	COMMAND_UNRELEASE    = "unrelease"
	COMMAND_REINDEX      = "reindex"
	COMMAND_PURGE_CACHE  = "purge-cache"
	COMMAND_STATS        = "stats"
	COMMAND_HELP         = "help"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Commands shortcuts
const (
	COMMAND_SHORT_LIST         = "l"
	COMMAND_SHORT_WHICH_SOURCE = "ws"
	COMMAND_SHORT_FIND         = "f"
	COMMAND_SHORT_INFO         = "i"
	COMMAND_SHORT_PAYLOAD      = "p"
	COMMAND_SHORT_CLEANUP      = "c"
	COMMAND_SHORT_CHECK        = "ch"
	COMMAND_SHORT_SIGN         = "s"
	COMMAND_SHORT_RESIGN       = "rs"
	COMMAND_SHORT_ADD          = "a"
	COMMAND_SHORT_REMOVE       = "rm"
	COMMAND_SHORT_RELEASE      = "r"
	COMMAND_SHORT_UNRELEASE    = "u"
	COMMAND_SHORT_REINDEX      = "ri"
	COMMAND_SHORT_PURGE_CACHE  = "pc"
	COMMAND_SHORT_STATS        = "st"
	COMMAND_SHORT_HELP         = "h"
)

// Global preferences
const (
	STORAGE_TYPE        = "storage:type"
	STORAGE_DATA        = "storage:data"
	STORAGE_CACHE       = "storage:cache"
	STORAGE_SPLIT_FILES = "storage:split-files"

	INDEX_CHECKSUM         = "index:checksum"
	INDEX_PRETTY           = "index:pretty"
	INDEX_UPDATE           = "index:update"
	INDEX_SPLIT            = "index:split"
	INDEX_SKIP_SYMLINKS    = "index:skip-symlinks"
	INDEX_CHANGELOG_LIMIT  = "index:changelog-limit"
	INDEX_MD_FILENAMES     = "index:md-filenames"
	INDEX_DISTRO           = "index:distro"
	INDEX_CONTENT          = "index:content"
	INDEX_REVISION         = "index:revision"
	INDEX_DELTAS           = "index:deltas"
	INDEX_NUM_DELTAS       = "index:num-deltas"
	INDEX_WORKERS          = "index:workers"
	INDEX_COMPRESSION_TYPE = "index:compression-type"

	LOG_DIR_PERMS  = "log:dir-perms"
	LOG_FILE_PERMS = "log:file-perms"
	LOG_DIR        = "log:dir"

	TEMP_DIR = "temp:dir"
)

// Repository preferences
const (
	REPOSITORY_NAME        = "repository:name"
	REPOSITORY_FILE_FILTER = "repository:file-filter"
	REPOSITORY_REPLACE     = "repository:replace"

	PERMISSIONS_USER  = "permissions:user"
	PERMISSIONS_GROUP = "permissions:group"
	PERMISSIONS_FILE  = "permissions:file"
	PERMISSIONS_DIR   = "permissions:dir"

	SIGN_REQUIRED = "sign:required"
	SIGN_KEY      = "sign:key"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_TESTING        = "t:testing"
	OPT_RELEASE        = "r:release"
	OPT_ALL            = "a:all"
	OPT_ARCH           = "aa:arch"
	OPT_MOVE           = "m:move"
	OPT_NO_SOURCE      = "ns:no-source"
	OPT_IGNORE_FILTER  = "if:ignore-filter"
	OPT_POSTPONE_INDEX = "pi:postpone-index"
	OPT_FORCE          = "f:force"
	OPT_FULL           = "F:full"
	OPT_SHOW_ALL       = "A:show-all"
	OPT_EPOCH          = "E:epoch"
	OPT_STATUS         = "S:status"
	OPT_PAGER          = "P:pager"
	OPT_NO_COLOR       = "nc:no-color"
	OPT_HELP           = "h:help"
	OPT_VER            = "v:version"

	OPT_DEBUG    = "D:debug"
	OPT_VERB_VER = "vv:verbose-version"

	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Path to global configuration file
const CONFIG_FILE = "/etc/rep.knf"

// Path to directory with repositories configuration files
const CONFIG_DIR = "/etc/rep.d"

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap is map with supported options
var optMap = options.Map{
	OPT_ARCH:           {},
	OPT_TESTING:        {Type: options.BOOL},
	OPT_RELEASE:        {Type: options.BOOL},
	OPT_ALL:            {Type: options.BOOL},
	OPT_MOVE:           {Type: options.BOOL},
	OPT_NO_SOURCE:      {Type: options.BOOL},
	OPT_IGNORE_FILTER:  {Type: options.BOOL},
	OPT_POSTPONE_INDEX: {Type: options.BOOL},
	OPT_FORCE:          {Type: options.BOOL},
	OPT_FULL:           {Type: options.BOOL},
	OPT_SHOW_ALL:       {Type: options.BOOL},
	OPT_EPOCH:          {Type: options.BOOL},
	OPT_STATUS:         {Type: options.BOOL},
	OPT_PAGER:          {Type: options.BOOL},
	OPT_NO_COLOR:       {Type: options.BOOL},
	OPT_HELP:           {Type: options.BOOL},
	OPT_VER:            {Type: options.MIXED},

	OPT_DEBUG:    {Type: options.BOOL},
	OPT_VERB_VER: {Type: options.BOOL},

	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

// repoNameRegex is regexp for repository name validation
var repoNamePattern = `[0-9a-zA-Z_\-]+`

// ////////////////////////////////////////////////////////////////////////////////// //

// configs contains repositories configs
var configs map[string]*knf.Config

// isCanceled is a flag for marking that user want to cancel app execution
var isCanceled = false

// isCancelProtected is a flag for marking current execution from canceling
var isCancelProtected = false

// rawOutput is raw output flag
var rawOutput = false

// ////////////////////////////////////////////////////////////////////////////////// //

func Init(gitRev string, gomod []byte) {
	args, errs := options.Parse(optMap)

	if !errs.IsEmpty() {
		terminal.Error("Options parsing errors:")
		terminal.Error(errs.Error("- "))
		os.Exit(1)
	}

	configureUI()

	switch {
	case options.Has(OPT_COMPLETION):
		os.Exit(printCompletion())
	case options.Has(OPT_GENERATE_MAN):
		printMan()
		os.Exit(0)
	case options.GetB(OPT_VER):
		genAbout(gitRev).Print(options.GetS(OPT_VER))
		os.Exit(0)
	case options.GetB(OPT_VERB_VER):
		support.Collect(APP, VER).
			WithRevision(gitRev).
			WithDeps(deps.Extract(gomod)).
			WithApps(getCreaterepoVersion()).
			Print()
		os.Exit(0)
	case options.GetB(OPT_HELP) || len(args) == 0:
		genUsage().Print()
		os.Exit(0)
	}

	err := errors.Chain(
		checkPermissions,
		loadGlobalConfig,
		validateGlobalConfig,
		loadRepoConfigs,
		validateRepoConfigs,
		configureRepoCache,
		configureSignalHandlers,
	)

	if err != nil {
		terminal.Error(err.Error())
		shutdown(1)
	}

	ok := process(args)

	if !ok {
		shutdown(1)
	}

	shutdown(0)
}

// configureUI configure user interface
func configureUI() {
	fmtc.DisableColors = true
	fmtutil.SizeSeparator = " "
	fmtutil.SeparatorSymbol = "–"
	fmtutil.SeparatorColorTag = "{s}"
	fmtutil.SeparatorTitleColorTag = "{*}"

	input.Prompt = "› "
	input.MaskSymbol = "•"
	input.MaskSymbolColorTag = "{s-}"
	input.TitleColorTag = "{s}"
	input.NewLine = true

	progress.DefaultSettings.NameColorTag = "{*}"
	progress.DefaultSettings.PercentColorTag = "{*}"
	progress.DefaultSettings.ProgressColorTag = "{s}"
	progress.DefaultSettings.SpeedColorTag = "{s}"
	progress.DefaultSettings.RemainingColorTag = ""
	progress.DefaultSettings.BarFgColorTag = "{c}"
	progress.DefaultSettings.IsSize = false
	progress.DefaultSettings.WindowSizeSec = 60.0

	spinner.DurationFormat = spinner.DURATION_MINI

	fmtc.AddColor("package", "{m}")
	fmtc.AddColor("repo", "{c}")

	if fmtc.IsColorsSupported() {
		fmtc.DisableColors = false
	}

	if fmtc.Is256ColorsSupported() {
		fmtc.AddColor("package", "{#108}")
		fmtc.AddColor("repo", "{#33}")
		progress.DefaultSettings.BarFgColorTag = "{#33}"
	}

	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if !tty.IsTTY() {
		fmtc.DisableColors = true
		rawOutput = true
	}
}

// checkPermissions checks that user has enough permissions
func checkPermissions() error {
	curUser, err := system.CurrentUser()

	if err != nil {
		return fmt.Errorf("Can't get info about current user: %w", err)
	}

	if !curUser.IsRoot() {
		return fmt.Errorf("This app requires superuser (root) privileges")
	}

	return nil
}

// loadGlobalConfig loads global configuration file
func loadGlobalConfig() error {
	err := knf.Global(CONFIG_FILE)

	if err != nil {
		return fmt.Errorf("Can't load global configuration: %w", err)
	}

	return nil
}

// validateGlobalConfig validates global configuration file properties
func validateGlobalConfig() error {
	validators := knf.Validators{
		{STORAGE_DATA, knfv.Set, nil},
		{STORAGE_CACHE, knfv.Set, nil},
		{LOG_DIR, knfv.Set, nil},
		{TEMP_DIR, knfv.Set, nil},

		{STORAGE_DATA, knff.Perms, "DRWX"},
		{STORAGE_CACHE, knff.Perms, "DRWX"},

		{LOG_DIR, knff.Perms, "DWX"},
		{TEMP_DIR, knff.Perms, "DRWX"},

		{INDEX_CHECKSUM, knfv.SetToAny, index.CheckSumMethods},
		{INDEX_MD_FILENAMES, knfv.SetToAny, index.MDFilenames},
		{INDEX_COMPRESSION_TYPE, knfv.SetToAny, index.CompressionMethods},
	}

	errs := knf.Validate(validators)

	if !errs.IsEmpty() {
		return fmt.Errorf("Error while global configuration file validation: %w", errs.First())
	}

	return nil
}

// loadRepoConfigs loads repositories configuration files
func loadRepoConfigs() error {
	filter := fsutil.ListingFilter{MatchPatterns: []string{"*.knf"}}
	configFiles := fsutil.List(CONFIG_DIR, false, filter)

	if len(configFiles) == 0 {
		return nil
	}

	fsutil.ListToAbsolute(CONFIG_DIR, configFiles)

	configs = make(map[string]*knf.Config)

	for _, cf := range configFiles {
		cfg, err := knf.Read(cf)

		if err != nil {
			return err
		}

		configs[cfg.GetS(REPOSITORY_NAME)] = cfg
	}

	return nil
}

// validateRepoConfigs validates repositories configuration files
func validateRepoConfigs() error {
	for _, cfg := range configs {
		validators := knf.Validators{
			{PERMISSIONS_USER, knfs.User, nil},
			{PERMISSIONS_GROUP, knfs.Group, nil},
			{REPOSITORY_NAME, knfr.Regexp, repoNamePattern},
		}

		validators = validators.AddIf(cfg.Has(SIGN_KEY),
			knf.Validators{
				{SIGN_KEY, knff.Perms, "FR"},
				{SIGN_KEY, knff.FileMode, os.FileMode(0600)},
			},
		)

		errs := cfg.Validate(validators)

		if errs.IsEmpty() {
			continue
		}

		return fmt.Errorf(
			"Error while repository configuration file validation (%s): %w",
			cfg.File(), errs.First(),
		)
	}

	return nil
}

// configureRepoCache configures cache for repository data
func configureRepoCache() error {
	cacheDir := knf.GetS(STORAGE_CACHE)

	for repo := range configs {
		repoCacheDir := cacheDir + "/" + repo

		if !fsutil.IsExist(repoCacheDir) {
			err := os.Mkdir(repoCacheDir, 0700)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// configureSignalHandlers configures handlers for signals
func configureSignalHandlers() error {
	signal.Handlers{
		signal.QUIT: sigHandler,
		signal.TERM: sigHandler,
		signal.INT:  sigHandler,
	}.TrackAsync()

	return nil
}

// getPrimaryRepoName returns primary repository name
func getPrimaryRepoName() string {
	var repos []string

	for repo := range configs {
		repos = append(repos, repo)
	}

	if len(repos) == 1 {
		return repos[0]
	}

	sortutil.StringsNatural(repos)

	return repos[len(repos)-1]
}

// process starts command processing
func process(args options.Arguments) bool {
	if len(configs) == 0 {
		terminal.Warn("No repository configuration files were found")
		return false
	}

	if len(configs) == 1 && configs[args.Get(0).String()] == nil {
		args = args.Unshift(getPrimaryRepoName())
	}

	repo := args.Get(0).String()

	switch repo {
	case COMMAND_HELP, COMMAND_SHORT_HELP, COMMAND_GEN_KEY:
		return runSimpleCommand(repo, args[1:])
	}

	if configs[repo] == nil {
		terminal.Error(
			"Unknown repository '%s'. Maybe you meant 'rep %s %s'?",
			repo, getPrimaryRepoName(), repo,
		)

		return false
	}

	if options.GetB(OPT_PAGER) && tty.IsTTY() {
		if pager.Setup() == nil {
			defer pager.Complete()
		}
	}

	// List repositories by default
	if args.Get(1).String() == "" {
		return runCommand(configs[repo], COMMAND_LIST, nil)
	}

	return runCommand(configs[repo], args.Get(1).String(), args[2:])
}

// sigHandler is handler for TERM, QUIT and INT signals
func sigHandler() {
	if !isCancelProtected {
		shutdown(1)
	}

	isCanceled = true
}

// shutdown cleans temporary data and exits from CLI
func shutdown(ec int) {
	os.Exit(ec)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getCreaterepoVersion returns version of installed createrepo_c
func getCreaterepoVersion() support.App {
	return apps.ExtractVersion("createrepo_c --version", 0, 1)
}

// printCompletion prints completion for given shell
func printCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Print(bash.Generate(info, APP))
	case "fish":
		fmt.Print(fish.Generate(info, APP))
	case "zsh":
		fmt.Print(zsh.Generate(info, optMap, APP))
	default:
		return 1
	}

	return 0
}

// printMan prints man page
func printMan() {
	fmt.Println(man.Generate(genUsage(), genAbout("")))
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo()

	if fmtc.Is256ColorsSupported() {
		info.AppNameColorTag = "{*}{#33}"
	}

	info.AddSpoiler(
		`  Note that if you have more than one repository, you should specify its name as the
  first argument. You can read detailed information about each command with usage
  examples by using the {y}help{!} command.`)

	info.AddCommand(COMMAND_INIT, "Initialize new repository", "arch…")
	info.AddCommand(COMMAND_GEN_KEY, "Generate keys for signing packages")
	info.AddCommand(COMMAND_LIST, "List latest versions of packages within repository", "?filter")
	info.AddCommand(COMMAND_FIND, "Search packages", "query…")
	info.AddCommand(COMMAND_WHICH_SOURCE, "Show source package name", "query…")
	info.AddCommand(COMMAND_INFO, "Show info about package", "package")
	info.AddCommand(COMMAND_PAYLOAD, "Show package payload", "package", "?type")
	info.AddCommand(COMMAND_CLEANUP, "Remove old versions of packages", "?num", "?filter")
	info.AddCommand(COMMAND_CHECK, "Check repositories consistency", "?errors-num")
	info.AddCommand(COMMAND_SIGN, "Sign one or more packages", "file…")
	info.AddCommand(COMMAND_RESIGN, "Resign all packages in repository")
	info.AddCommand(COMMAND_ADD, "Add one or more packages to testing repository", "file…")
	info.AddCommand(COMMAND_REMOVE, "Remove package or packages from repository", "query…")
	info.AddCommand(COMMAND_RELEASE, "Copy package or packages from testing to release repository", "query…")
	info.AddCommand(COMMAND_UNRELEASE, "Remove package or packages from release repository", "query…")
	info.AddCommand(COMMAND_REINDEX, "Create or update repository index")
	info.AddCommand(COMMAND_PURGE_CACHE, "Clean all cached data")
	info.AddCommand(COMMAND_STATS, "Show some statistics information about repositories")
	info.AddCommand(COMMAND_HELP, "Show detailed information about command", "command")

	info.AddOption(OPT_RELEASE, "Run command only on release {s}(stable){!} repository")
	info.AddOption(OPT_TESTING, "Run command only on testing {s}(unstable){!} repository")
	info.AddOption(OPT_ALL, "Run command on all repositories")
	info.AddOption(OPT_ARCH, `Package architecture`, "arch")
	info.AddOption(OPT_MOVE, `Move {s}(remove after successful action){!} packages`)
	info.AddOption(OPT_NO_SOURCE, `Ignore source packages`)
	info.AddOption(OPT_IGNORE_FILTER, `Ignore repository file filter`)
	info.AddOption(OPT_POSTPONE_INDEX, `Postpone repository reindex`)
	info.AddOption(OPT_FORCE, `Answer "yes" for all questions`)
	info.AddOption(OPT_FULL, `Full reindex`)
	info.AddOption(OPT_SHOW_ALL, `Show all versions of packages`)
	info.AddOption(OPT_STATUS, "Show package status {s-}(released or not){!}")
	info.AddOption(OPT_EPOCH, `Show epoch info`)
	info.AddOption(OPT_PAGER, "Use pager for long output")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.BoundOptions(COMMAND_ADD, OPT_FORCE)
	info.BoundOptions(COMMAND_ADD, OPT_IGNORE_FILTER)
	info.BoundOptions(COMMAND_ADD, OPT_MOVE)
	info.BoundOptions(COMMAND_ADD, OPT_NO_SOURCE)
	info.BoundOptions(COMMAND_CLEANUP, OPT_FORCE)
	info.BoundOptions(COMMAND_CLEANUP, OPT_RELEASE)
	info.BoundOptions(COMMAND_CLEANUP, OPT_TESTING)
	info.BoundOptions(COMMAND_FIND, OPT_RELEASE)
	info.BoundOptions(COMMAND_FIND, OPT_STATUS)
	info.BoundOptions(COMMAND_FIND, OPT_TESTING)
	info.BoundOptions(COMMAND_FIND, OPT_PAGER)
	info.BoundOptions(COMMAND_INFO, OPT_ARCH)
	info.BoundOptions(COMMAND_INFO, OPT_PAGER)
	info.BoundOptions(COMMAND_LIST, OPT_EPOCH)
	info.BoundOptions(COMMAND_LIST, OPT_RELEASE)
	info.BoundOptions(COMMAND_LIST, OPT_SHOW_ALL)
	info.BoundOptions(COMMAND_LIST, OPT_STATUS)
	info.BoundOptions(COMMAND_LIST, OPT_TESTING)
	info.BoundOptions(COMMAND_LIST, OPT_PAGER)
	info.BoundOptions(COMMAND_PAYLOAD, OPT_ARCH)
	info.BoundOptions(COMMAND_PAYLOAD, OPT_PAGER)
	info.BoundOptions(COMMAND_PURGE_CACHE, OPT_RELEASE)
	info.BoundOptions(COMMAND_PURGE_CACHE, OPT_TESTING)
	info.BoundOptions(COMMAND_REINDEX, OPT_FULL)
	info.BoundOptions(COMMAND_REINDEX, OPT_RELEASE)
	info.BoundOptions(COMMAND_REINDEX, OPT_TESTING)
	info.BoundOptions(COMMAND_RELEASE, OPT_FORCE)
	info.BoundOptions(COMMAND_REMOVE, OPT_ALL)
	info.BoundOptions(COMMAND_REMOVE, OPT_FORCE)
	info.BoundOptions(COMMAND_SIGN, OPT_IGNORE_FILTER)
	info.BoundOptions(COMMAND_RESIGN, OPT_FORCE)
	info.BoundOptions(COMMAND_STATS, OPT_RELEASE)
	info.BoundOptions(COMMAND_STATS, OPT_TESTING)
	info.BoundOptions(COMMAND_STATS, OPT_PAGER)
	info.BoundOptions(COMMAND_UNRELEASE, OPT_FORCE)
	info.BoundOptions(COMMAND_WHICH_SOURCE, OPT_EPOCH)
	info.BoundOptions(COMMAND_WHICH_SOURCE, OPT_RELEASE)
	info.BoundOptions(COMMAND_WHICH_SOURCE, OPT_TESTING)
	info.BoundOptions(COMMAND_WHICH_SOURCE, OPT_PAGER)

	return info
}

// genAbout generates info about version
func genAbout(gitRev string) *usage.About {
	about := &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2009,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/rep", update.GitHubChecker},

		DescSeparator: "{s}—{!}",
	}

	if gitRev != "" {
		about.Build = "git:" + gitRev
	}

	if fmtc.Is256ColorsSupported() {
		about.AppNameColorTag = "{*}{#33}"
		about.VersionColorTag = "{#33}"
	}

	crInfo := getCreaterepoVersion()

	if crInfo.Version != "" {
		about.Environment = append(about.Environment,
			usage.EnvironmentInfo{crInfo.Name, crInfo.Version},
		)
	}

	return about
}
