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

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/sliceutil"
	"github.com/essentialkaos/ek/v12/terminal"
	"github.com/essentialkaos/ek/v12/usage"

	"github.com/essentialkaos/rep/cli/query"
	"github.com/essentialkaos/rep/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type commandHelp struct {
	command  string
	shortcut string
	info     *usage.Info
	examples []commandExample
	isGlobal bool
}

type commandExample struct {
	command string
	desc    string
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Usage shows basic usage info
func (c *commandHelp) Usage() {
	if len(configs) > 1 && !c.isGlobal {
		fmtc.Print("{*}Usage:{!} rep {c}{repo-id}{!}")
	} else {
		fmtc.Print("{*}Usage:{!} rep")
	}

	fmtc.Printf(" {y}%s{!}", c.command)

	cmd := c.info.GetCommand(c.command)

	if cmd != nil && len(cmd.Args) != 0 {
		fmtc.Print(" " + c.renderArgs(cmd.Args) + "\n")
	} else {
		fmtc.Print("\n")
	}

	fmtc.NewLine()
}

// Shortcut shows info about shortcut version of command
func (c *commandHelp) Shortcut() {
	fmtc.Println("{*}Shortcut:{!}\n")
	fmtc.Printf("  {y}%s{!} → {y}%s{!}\n\n", c.command, c.shortcut)
}

// Examples shows usage examples
func (c *commandHelp) Examples() {
	if len(c.examples) == 0 {
		return
	}

	fmtc.Println("{*}Examples:{!}\n")

	for index, example := range c.examples {
		if len(configs) > 1 && !c.isGlobal {
			fmtc.Printf("  rep {c}{repo-id}{!} {y}%s{!} {s}%s{!}\n", c.command, example.command)
		} else {
			fmtc.Printf("  rep {y}%s{!} {s}%s{!}\n", c.command, example.command)
		}

		fmtc.Printf("{s-}%s{!}\n", fmtutil.Wrap(example.desc, "  ", 88))

		if index+1 < len(c.examples) {
			fmtc.NewLine()
		}
	}
}

// Options shows list of command related options
func (c *commandHelp) Options() {
	cmd := c.info.GetCommand(c.command)

	if cmd == nil || len(cmd.BoundOptions) == 0 {
		return
	}

	if len(c.examples) != 0 {
		fmtc.NewLine()
	}

	fmtc.Println("{*}Options:{!}\n")

	for _, option := range c.info.Options {
		if option == nil {
			continue
		}

		if !sliceutil.Contains(cmd.BoundOptions, option.Long) {
			continue
		}

		option.Render()
	}
}

// Paragraph renders paragraph text
func (c *commandHelp) Paragraph(text string) {
	fmt.Println(fmtutil.Wrap(fmtc.Sprint(text), "  ", 88))
	fmtc.NewLine()
}

// Query renders query info
func (c *commandHelp) Query(short, long, desc, typ string) {
	fmtc.Printf("   {m}%2s{!} {s}or{!} {m}%-12s{!} %s {s-}(%s){!}\n", short, long, desc, typ)
}

// renderArgs renders command arguments with colors
func (c *commandHelp) renderArgs(args []string) string {
	var result string

	for _, a := range args {
		if strings.HasPrefix(a, "?") {
			result += "{s-}" + a[1:] + "{!} "
		} else {
			result += "{s}" + a + "{!} "
		}
	}

	return fmtc.Sprintf(strings.TrimRight(result, " "))
}

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdHelp is 'help' command handler
func cmdHelp(ctx *context, args options.Arguments) bool {
	cmdName := args.Get(0)

	switch cmdName {
	case "":
		helpAll()

	case COMMAND_INIT:
		helpInit()

	case COMMAND_GEN_KEY:
		helpGenKey()

	case COMMAND_LIST, COMMAND_SHORT_LIST:
		helpList()

	case COMMAND_WHICH_SOURCE, COMMAND_SHORT_WHICH_SOURCE:
		helpWhichSource()

	case COMMAND_FIND, COMMAND_SHORT_FIND:
		helpFind()

	case COMMAND_INFO, COMMAND_SHORT_INFO:
		helpInfo()

	case COMMAND_PAYLOAD, COMMAND_SHORT_PAYLOAD:
		helpPayload()

	case COMMAND_SIGN, COMMAND_SHORT_SIGN:
		helpSign()

	case COMMAND_RESIGN, COMMAND_SHORT_RESIGN:
		helpResign()

	case COMMAND_ADD, COMMAND_SHORT_ADD:
		helpAdd()

	case COMMAND_REMOVE, COMMAND_SHORT_REMOVE:
		helpRemove()

	case COMMAND_RELEASE, COMMAND_SHORT_RELEASE:
		helpRelease()

	case COMMAND_UNRELEASE, COMMAND_SHORT_UNRELEASE:
		helpUnrelease()

	case COMMAND_REINDEX, COMMAND_SHORT_REINDEX:
		helpReindex()

	case COMMAND_PURGE_CACHE, COMMAND_SHORT_PURGE_CACHE:
		helpPurgeCache()

	case COMMAND_STATS, COMMAND_SHORT_STATS:
		helpStats()

	case COMMAND_HELP, COMMAND_SHORT_HELP:
		helpHelp()

	default:
		terminal.PrintErrorMessage("Unknown command \"%s\"", cmdName)
		return false
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// helpAll shows info about all supported commands
func helpAll() {
	usageInfo := genUsage()

	for _, c := range usageInfo.Commands {
		if c.Name == COMMAND_HELP {
			continue
		}

		c.Render()
	}

	fmtc.NewLine()
	fmtc.Println("  {s}For detailed information about command use{!} {y}help {command}{!}")
}

// helpInit shows help content about "init" command
func helpInit() {
	help := &commandHelp{
		command: COMMAND_INIT,
		info:    genUsage(),
		examples: []commandExample{
			{"src i386 x86_64", "Initialize the new repository with specific architectures"},
		},
	}

	help.Usage()
	help.Paragraph("The command creates all required directories for new repository.")
	help.Paragraph("You must define at least one architecture for repository. List of supported architectures:")

	for _, arch := range sliceutil.Exclude(data.ArchList, data.ARCH_NOARCH) {
		if fmtc.Is256ColorsSupported() {
			fmtc.Printf("    {s-}•{!} "+archColorsExt[arch]+"%s{!}\n", arch)
		} else {
			fmtc.Printf("    {s-}•{!} "+archColors[arch]+"%s{!}\n", arch)
		}
	}

	fmtc.NewLine()
	help.Examples()
}

// helpGenKey shows help content about "gen-key" command
func helpGenKey() {
	help := &commandHelp{
		command: COMMAND_GEN_KEY,
		info:    genUsage(),
		examples: []commandExample{
			{"", "Generate new private and public keys"},
		},
	}

	help.Usage()
	help.Paragraph("The command generates new 4096 bits long RSA private and public keys for signing packages.")
	help.Examples()
}

// helpList shows help content about "list" command
func helpList() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_LIST,
		shortcut: COMMAND_SHORT_LIST,
		info:     info,
		examples: []commandExample{
			{"", "Show a list of all the latest versions of packages in all (release and testing) repositories"},
			{"my-package", "Show a list of all versions of the package with the given name"},
			{info.GetOption(OPT_TESTING).String() + " my-package", "Show a list of all package versions with the given name only in the testing repository"},
			{"| grep my-package | grep -v '.src.'", "Show a list of packages files and filter it with grep"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("The command shows a list of all packages in the repository. By default, the command shows only the latest versions of packages within all repositories.")
	help.Paragraph("You can filter the listing providing part of the package name. In this case, the command will show all versions of packages with the given name part.")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpWhichSource shows help content about "which-source" command
func helpWhichSource() {
	help := &commandHelp{
		command:  COMMAND_WHICH_SOURCE,
		shortcut: COMMAND_SHORT_WHICH_SOURCE,
		info:     genUsage(),
		examples: []commandExample{
			{"my-package-1.0", "Simple package search"},
			{"n:my-package v:1.0* d:3w", "Find packages with search query syntax"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("This command shows the source package used for package building or source package created while package building. This command is very useful for package searching. You may find the source package and use it in the search query ({s}s:{!} or {s}source:{!} query prefix with {y}" + COMMAND_REMOVE + "{!}, {y}" + COMMAND_RELEASE + "{!}, and {y}" + COMMAND_UNRELEASE + "{!} commands).")
	help.Paragraph("You can use search query syntax for package selection. For more information about query syntax, see \"rep {y}" + COMMAND_HELP + "{!} {s}" + COMMAND_FIND + "{!}\".")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpFind shows help content about "find" command
func helpFind() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_FIND,
		shortcut: COMMAND_SHORT_FIND,
		info:     info,
		examples: []commandExample{
			{"nginx", "Search packages which name starts with \"nginx\""},
			{"n:nginx", "Search packages with name \"nginx\""},
			{info.GetOption(OPT_TESTING).String() + " n:nginx", "Search packages with name \"nginx\" only in the testing repository"},
			{"n:'*utils*'", "Search packages with substring \"utils\" in name"},
			{"n:nginx v:1.21.3", "Search packages with given name and version"},
			{"n:nginx v:1.21.3 r:1.el7", "Search packages with given name, version and release"},
			{"n:nginx v:1.21.3 r::1.*", "Search packages with given name, version and release which NOT equals 1"},
			{"n:nginx v:'1.19.6|1.21.3|1.21.0'", "Search packages with given name and versions"},
			{"my-package a:x86_64", "Search packages with given name and architecture"},
			{"s:redis-6.0.4-0.el7.src", "Search packages built from given source package"},
			{"R:'mylib>=1.16'", "Search packages which require mylib 1.16 or greater"},
			{"R:'/usr/sbin/useradd'", "Search packages which require useradd utility"},
			{"P:'postgresql-server=11.*'", "Search packages which provide \"postgresql-server\" package"},
			{"n:nginx d:7", "Search packages with name \"nginx\" added to the repository in last 7 days"},
			{"D:1w3d12h15m30s", "Search packages built in last 1 week, 3 days, 12 hours, 15 minutes, and 30 seconds"},
			{"S:10mb", "Search packages with a size around 10 megabytes (size +/- 2%)"},
			{"S:100mb+", "Search packages bigger than 100 megabytes"},
			{"S:20mb-", "Search packages smaller than 20 kilobytes"},
			{"S:50mb-100mb", "Search packages with a size between 50 and 100 megabytes"},
			{"f:'/etc/redis.conf'", "Search packages with configuration file \"/etc/redis.conf\""},
			{"@:'/usr/include/curl/*.h'", "Search packages with header files for cURL"},
			{"n:nginx ^:no", "All nginx packages which not yet released"},
			{"n:nginx ^:true", "All released nginx packages"},
			{
				"postgres v:'10.*' | grep -E '(devel|docs)' | awk -F'/' '{print $NF}' | sort -u",
				"Search packages and process list with found rpm files with grep, awk, and sort",
			},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Search packages within the repository. By default, command search packages within all {s}(release and testing){!} repositories.")

	fmtc.Println("{*}Query syntax:{!}\n")
	help.Paragraph("For search you can use rich query syntax. You may define different filters:")

	help.Query(query.TERM_SHORT_NAME, query.TERM_NAME, "Package name", "String")
	help.Query(query.TERM_SHORT_VERSION, query.TERM_VERSION, "Package version", "SemVer")
	help.Query(query.TERM_SHORT_RELEASE, query.TERM_RELEASE, "Package release", "String")
	help.Query(query.TERM_SHORT_EPOCH, query.TERM_EPOCH, "Package epoch", "Number")
	help.Query(query.TERM_SHORT_ARCH, query.TERM_ARCH, "Package architecture", "Architecture")
	help.Query(query.TERM_SHORT_SOURCE, query.TERM_SOURCE, "Name of source package used for build or created while building", "String")
	help.Query(query.TERM_SHORT_LICENSE, query.TERM_LICENSE, "Package license", "String")
	help.Query(query.TERM_SHORT_GROUP, query.TERM_GROUP, "Package group", "String")
	help.Query(query.TERM_SHORT_VENDOR, query.TERM_VENDOR, "Package vendor", "String")
	help.Query(query.TERM_SHORT_PROVIDES, query.TERM_PROVIDES, "Package name or binary name provided by the package", "Dependency")
	help.Query(query.TERM_SHORT_REQUIRES, query.TERM_REQUIRES, "Package name or binary name required by the package", "Dependency")
	help.Query(query.TERM_SHORT_CONFLICTS, query.TERM_CONFLICTS, "Name of conflicting package", "Dependency")
	help.Query(query.TERM_SHORT_OBSOLETES, query.TERM_OBSOLETES, "Name of obsolete package", "Dependency")
	help.Query(query.TERM_SHORT_RECOMMENDS, query.TERM_RECOMMENDS, "Name of package defined as recomended", "Dependency")
	help.Query(query.TERM_SHORT_ENHANCES, query.TERM_ENHANCES, "Name of package defined as the enhancement", "Dependency")
	help.Query(query.TERM_SHORT_SUGGESTS, query.TERM_SUGGESTS, "Name of package defined as the suggestion", "Dependency")
	help.Query(query.TERM_SHORT_SUPPLEMENTS, query.TERM_SUPPLEMENTS, "Name of package defined as the supplement", "Dependency")
	help.Query(query.TERM_SHORT_DATE_ADD, query.TERM_DATE_ADD, "Duration since package was added to repository", "Duration")
	help.Query(query.TERM_SHORT_DATE_BUILD, query.TERM_DATE_BUILD, "Duration since package was built", "Duration")
	help.Query(query.TERM_SHORT_SIZE, query.TERM_SIZE, "Package size", "Size")
	help.Query(query.TERM_SHORT_FILE, query.TERM_FILE, "Path of config, binary or executable file provided by package", "String")
	help.Query(query.TERM_SHORT_PAYLOAD, query.TERM_PAYLOAD, "Path of file or directory in package", "String")
	help.Query(query.TERM_SHORT_RELEASED, query.TERM_RELEASED, "Release status", "Boolean")

	fmtc.NewLine()

	help.Paragraph("More info about supported query data types:")

	fmtc.Println("    {s-}•{!} String        Any string value")
	fmtc.Println("    {s-}•{!} Number        Integer greater or equal to zero")
	fmtc.Println("    {s-}•{!} Boolean       Boolean value {s}(yes/true/1 or no/false/0){!}")
	fmtc.Println("    {s-}•{!} SemVer        Version in semantic versioning format")
	fmtc.Println("    {s-}•{!} Dependency    Package name with or without version and release condition")
	fmtc.Println("    {s-}•{!} Architecture  Package architecture {s}(" + strings.Join(data.ArchList, ", ") + "){!}")
	fmtc.Println("    {s-}•{!} Size          Size {s}(b/kb/mb/gb){!} with modificators {s-}(see examples){!}")
	fmtc.Println("    {s-}•{!} Duration      Duration in days or custom duration {s-}(see examples){!}")

	fmtc.NewLine()

	help.Paragraph("You can define a few filters at once, in this case, data that match the previous filter will be filtered by the next filter in the query. For negative search use additional colon ({s}:{!}) symbol.")

	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpInfo shows help content about "info" command
func helpInfo() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_INFO,
		shortcut: COMMAND_SHORT_INFO,
		info:     info,
		examples: []commandExample{
			{"redis", "Show info about the latest version and release of the package"},
			{"redis-6.0.2", "Show info about the latest release of the specific version of the package"},
			{"redis-6.0.1-2", "Show info about specific version and release of the package"},
			{info.GetOption(OPT_ARCH).String() + " src redis", "Show info about the latest version and release of the source package"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Show detailed information about a package. If the package version wasn't provided command will show information about the latest version.")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpPayload shows help content about "payload" command
func helpPayload() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_PAYLOAD,
		shortcut: COMMAND_SHORT_PAYLOAD,
		info:     info,
		examples: []commandExample{
			{"redis", "Show info about the latest version and release of the package"},
			{"redis-6.0.2", "Show info about the latest release of the specific version of the package"},
			{"redis-6.0.1-2", "Show info about specific version and release of the package"},
			{info.GetOption(OPT_ARCH).String() + " src redis", "Show info about the latest version and release of the source package"},
			{"redis | grep '\\.conf'", "Show list of files and directories in the package and filter it with grep"},
			{"redis requires", "Show a list of required dependencies"},
			{"redis provides", "Show a list of provided dependencies"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Show information about package payload.")
	fmtc.Println("{*}Payload type:{!}\n")
	fmtc.Printf(
		"  {m}%-8s{!} {s}or{!} {m}%-6s{!} %s {s}(used by default){!}\n", "files", "f",
		"Files and directories",
	)
	fmtc.Printf(
		"  {m}%-8s{!} {s}or{!} {m}%-6s{!} %s\n", "requires", "reqs",
		"Reqiured dependencies",
	)
	fmtc.Printf(
		"  {m}%-8s{!} {s}or{!} {m}%-6s{!} %s\n", "provides", "provs",
		"Provided dependencies",
	)
	fmtc.NewLine()
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpSign shows help content about "sign" command
func helpSign() {
	help := &commandHelp{
		command:  COMMAND_SIGN,
		shortcut: COMMAND_SHORT_SIGN,
		info:     genUsage(),
		examples: []commandExample{
			{"*.rpm", "Sign all RPM packages in the current directory"},
		},
	}

	help.Usage()
	help.Paragraph("Add GPG signature to RPM file or files.")
	help.Shortcut()
	help.Examples()
}

// helpSign shows help content about "resign" command
func helpResign() {
	help := &commandHelp{
		command:  COMMAND_RESIGN,
		shortcut: COMMAND_SHORT_RESIGN,
		info:     genUsage(),
		examples: []commandExample{
			{"", "Re-sign all packages"},
		},
	}

	help.Usage()
	help.Paragraph("Re-sign all packages in testing and release repositories.")
	help.Shortcut()
	help.Examples()
}

// helpAdd shows help content about "add" command
func helpAdd() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_ADD,
		shortcut: COMMAND_SHORT_ADD,
		info:     info,
		examples: []commandExample{
			{"*.rpm", "Add all RPM packages in the current directory"},
			{info.GetOption(OPT_MOVE).String() + " *.rpm", "Add all RPM packages in the current directory and remove them after success"},
			{info.GetOption(OPT_NO_SOURCE).String() + " *.rpm", "Add all RPM packages in the current directory except source packages"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Add RPM file or files to the testing repository.")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpRemove shows help content about "remove" command
func helpRemove() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_REMOVE,
		shortcut: COMMAND_SHORT_REMOVE,
		info:     info,
		examples: []commandExample{
			{"n:nginx v:1.21.3", "Remove all packages from the testing repository with a specific name and version"},
			{"s:redis-6.0.4-0.el7.src", "Remove all packages from the testing repository built from the given source package"},
			{info.GetOption(OPT_ALL).String() + " n:nginx v:1.21.3", "Remove all packages from testing and release repositories with specific name and version"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Remove package or packages from the testing repository. By default, the command removes packages from the testing repository. You can use option {g}" + info.GetOption(OPT_ALL).String() + "{!} for removing packages from the testing and release repository.")
	help.Paragraph("The command uses search query syntax for package selection. For more information about query syntax, see \"rep {y}" + COMMAND_HELP + "{!} {s}" + COMMAND_FIND + "{!}\".")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpRelease shows help content about "release" command
func helpRelease() {
	help := &commandHelp{
		command:  COMMAND_RELEASE,
		shortcut: COMMAND_SHORT_RELEASE,
		info:     genUsage(),
		examples: []commandExample{
			{"d:3d", "Release all packages added in the last 3 days"},
			{"s:redis-6.0.4-0.el7.src", "Release all packages built from the given source package"},
		},
	}

	help.Usage()
	help.Paragraph("Copy package or packages from the testing repository to the release repository.")
	help.Paragraph("The command uses search query syntax for package selection. For more information about query syntax, see \"rep {y}" + COMMAND_HELP + "{!} {s}" + COMMAND_FIND + "{!}\".")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpUnrelease shows help content about "unrelease" command
func helpUnrelease() {
	help := &commandHelp{
		command:  COMMAND_UNRELEASE,
		shortcut: COMMAND_SHORT_UNRELEASE,
		info:     genUsage(),
		examples: []commandExample{
			{"d:3d", "Unrelease all packages released in the last 3 days"},
			{"s:redis-6.0.4-0.el7.src", "Unrelease all packages built from the given source package"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Remove package or packages from the release repository.")
	help.Paragraph("If package or packages were previously removed from the testing repository, this command will move packages from the release repository to testing.")
	help.Paragraph("The command uses search query syntax for package selection. For more information about query syntax, see \"rep {y}" + COMMAND_HELP + "{!} {s}" + COMMAND_FIND + "{!}\".")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpReindex shows help content about "reindex" command
func helpReindex() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_REINDEX,
		shortcut: COMMAND_SHORT_REINDEX,
		info:     info,
		examples: []commandExample{
			{"", "Regenerate index for testing and release repositories"},
			{info.GetOption(OPT_TESTING).String(), "Regenerate index only for the testing repository"},
			{info.GetOption(OPT_FULL).String(), "Generate index for testing and release repositories from scratch"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Generate repository index with createrepo utility.")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpPurgeCache shows help content about "purge-cache" command
func helpPurgeCache() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_PURGE_CACHE,
		shortcut: COMMAND_SHORT_PURGE_CACHE,
		info:     info,
		examples: []commandExample{
			{"", "Remove cached SQLite databases for testing and release repositories"},
			{info.GetOption(OPT_TESTING).String(), "Remove cached SQLite databases only for the testing repository"},
		},
	}

	help.Usage()
	help.Paragraph("Remove all cached SQLite databases.")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpStats shows help content about "stats" command
func helpStats() {
	info := genUsage()
	help := &commandHelp{
		command:  COMMAND_STATS,
		shortcut: COMMAND_SHORT_STATS,
		info:     info,
		examples: []commandExample{
			{"", "Show statistic information about testing and release repositories"},
			{info.GetOption(OPT_TESTING).String(), "Show statistic information only about the testing repository"},
		},
		isGlobal: false,
	}

	help.Usage()
	help.Paragraph("Show repository statistics.")
	help.Shortcut()
	help.Examples()
	help.Options()
}

// helpHelp shows help content about "help" command
func helpHelp() {
	help := &commandHelp{
		command:  COMMAND_HELP,
		shortcut: COMMAND_SHORT_HELP,
		info:     genUsage(),
		examples: []commandExample{
			{COMMAND_FIND, "Show usage information about command \"" + COMMAND_FIND + "\""},
		},
		isGlobal: true,
	}

	help.Usage()
	help.Paragraph("Show detailed usage information about specific command.")
	help.Shortcut()
	help.Examples()
}
