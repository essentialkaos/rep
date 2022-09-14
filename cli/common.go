package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/secstr"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/cli/query"
	"github.com/essentialkaos/rep/repo"
	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/sign"
)

// ////////////////////////////////////////////////////////////////////////////////// //

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

var archTags = map[string]string{
	data.ARCH_SRC:     "src",
	data.ARCH_NOARCH:  "noarch",
	data.ARCH_I386:    "x32",
	data.ARCH_I586:    "i586",
	data.ARCH_I686:    "i686",
	data.ARCH_X64:     "x64",
	data.ARCH_AARCH64: "aa64",
	data.ARCH_PPC64:   "p64",
	data.ARCH_PPC64LE: "p64l",
	data.ARCH_ARM:     "arm",
	data.ARCH_ARMV7HL: "arm7",
}

// ////////////////////////////////////////////////////////////////////////////////// //

// checkRPMFiles checks if we have enough permissions to manipulate with RPM files
func checkRPMFiles(files []string) bool {
	var hasErrors bool

	for _, file := range files {
		err := fsutil.ValidatePerms("FRS", file)

		if err != nil {
			terminal.PrintErrorMessage(err.Error())
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
	privateKey, err := r.Parent.SigningKey.Get(nil)

	if err != nil {
		return true
	}

	for _, file := range files {
		hasSignature, err := sign.HasSignature(file)

		if err != nil || !hasSignature {
			return true
		}

		isSigned, err := sign.IsSigned(file, privateKey)

		if err != nil || !isSigned {
			return true
		}
	}

	return false
}

// getRepoPrivateKey reads password and decrypts repository private key
func getRepoPrivateKey(r *repo.Repository) (*sign.PrivateKey, bool) {
	if r.SigningKey == nil {
		terminal.PrintWarnMessage("No signing key defined in configuration file")
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

	privateKey, err := r.SigningKey.Get(password)

	password.Destroy()

	if err != nil {
		terminal.PrintErrorMessage("Can't decrypt private key (wrong passphrase?)")
		return nil, false
	}

	return privateKey, true
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
