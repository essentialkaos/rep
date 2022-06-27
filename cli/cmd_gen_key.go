package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/secstr"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo/sign/keygen"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// outputPrivKeyFile is default name for generated private key file
var outputPrivKeyFile = "key.private"

// repoKeyNameValidator is regex pattern for repository key name validation
var repoKeyNameValidator = regexp.MustCompile(`^[A-Za-z0-9_\-\ ]+$`)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdGenKey is 'gen-key' command handler
func cmdGenKey(ctx *context, args options.Arguments) bool {
	if fsutil.IsExist(outputPrivKeyFile) {
		terminal.PrintErrorMessage("Private key file (%s) already exists", outputPrivKeyFile)
		return false
	}

	var err error
	var name, outputPubKeyFile string

	for {
		name, err = terminal.ReadUI("Key name", true)

		if err != nil {
			return false
		}

		if !repoKeyNameValidator.MatchString(name) {
			terminal.PrintErrorMessage("\nGiven name is invalid\n")
			continue
		}

		outputPubKeyFile = "RPM-GPG-KEY-" + strings.ReplaceAll(name, " ", "-")

		if fsutil.IsExist(outputPubKeyFile) {
			terminal.PrintErrorMessage("\nPublic key file for given name (%s) already exists\n", outputPubKeyFile)
			continue
		}

		break
	}

	fmtc.NewLine()

	email, err := terminal.ReadUI("Email address", true)

	if err != nil {
		return false
	}

	fmtc.NewLine()

	password, err := terminal.ReadPasswordSecure("Passphrase", true)

	if err != nil {
		return false
	}

	fmtc.NewLine()

	return generateKeys(name, email, password, outputPubKeyFile)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// generatePrivateKey generates and saves private key file
func generateKeys(name, email string, password *secstr.String, outputPubKeyFile string) bool {
	spinner.Show("Generating keys")

	privKeyData, pubKeyData, err := keygen.Generate(name, email, password)

	if err != nil {
		spinner.Update(err.Error())
		spinner.Done(false)
		return false
	}

	err = ioutil.WriteFile(outputPrivKeyFile, privKeyData, 0600)

	if err != nil {
		spinner.Update("Can't save private key: %v", err)
		spinner.Done(false)
		return false
	}

	err = ioutil.WriteFile(outputPubKeyFile, pubKeyData, 0644)

	if err != nil {
		spinner.Update("Can't save public key: %v", err)
		spinner.Done(false)
		return false
	}

	os.Chmod(outputPrivKeyFile, 0600)
	os.Chmod(outputPubKeyFile, 0644)

	spinner.Update("Keys successfully generated")
	spinner.Done(true)

	fmtc.NewLine()

	fmtc.Printf("{g}Private key saved as {*}%s{!}\n", outputPrivKeyFile)
	fmtc.Printf("{g}Public key saved as {*}%s{!}\n", outputPubKeyFile)

	return true
}
