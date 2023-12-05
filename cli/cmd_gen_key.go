package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"regexp"
	"strings"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/passwd"
	"github.com/essentialkaos/ek/v12/secstr"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/v3/repo/sign/keygen"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// outputPrivKeyFile is default name for generated private key file
var outputPrivKeyFile = "key.private"

// repoKeyNameValidator is regex pattern for repository key name validation
var repoKeyNameValidator = regexp.MustCompile(`^[A-Za-z0-9_\-\ ]+$`)

// emailValidator is regex for email validation
var emailValidator = regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+(;[^@]+@[^@]+\.[^@]+)*$`)

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdGenKey is 'gen-key' command handler
func cmdGenKey(ctx *context, args options.Arguments) bool {
	if fsutil.IsExist(outputPrivKeyFile) {
		terminal.Error("Private key file (%s) already exists", outputPrivKeyFile)
		return false
	}

	var err error
	var password *secstr.String
	var name, email, outputPubKeyFile string

	for {
		name, err = terminal.Read("Key name", true)

		if err != nil {
			return false
		}

		if !repoKeyNameValidator.MatchString(name) {
			terminal.Error("\nGiven name is invalid\n")
			continue
		}

		outputPubKeyFile = "RPM-GPG-KEY-" + strings.ReplaceAll(name, " ", "-")

		if fsutil.IsExist(outputPubKeyFile) {
			terminal.Error("\nPublic key file for given name (%s) already exists\n", outputPubKeyFile)
			continue
		}

		break
	}

	fmtc.NewLine()

	for {
		email, err = terminal.Read("Email address", true)

		if err != nil {
			return false
		}

		if !emailValidator.MatchString(email) {
			terminal.Error("\nGiven email address is invalid\n")
			continue
		}

		break
	}

	fmtc.NewLine()

	for {
		password, err = terminal.ReadPasswordSecure("Passphrase", true)

		if err != nil {
			return false
		}

		fmtc.NewLine()

		if passwd.GetPasswordBytesStrength(password.Data) < passwd.STRENGTH_MEDIUM {
			terminal.Warn("Given passphrase is not strong enough.\n")

			ok, _ := terminal.ReadAnswer("Use this passphrase anyway?", "n")

			fmtc.NewLine()

			if ok {
				break
			}
		} else {
			break
		}
	}

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

	err = os.WriteFile(outputPrivKeyFile, privKeyData, 0600)

	if err != nil {
		spinner.Update("Can't save private key: %v", err)
		spinner.Done(false)
		return false
	}

	err = os.WriteFile(outputPubKeyFile, pubKeyData, 0644)

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
