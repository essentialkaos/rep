package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/essentialkaos/ek/v13/fmtc"
	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/options"
	"github.com/essentialkaos/ek/v13/passwd"
	"github.com/essentialkaos/ek/v13/secstr"
	"github.com/essentialkaos/ek/v13/spinner"
	"github.com/essentialkaos/ek/v13/terminal"
	"github.com/essentialkaos/ek/v13/terminal/input"

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

	name, err = input.Read("Key name", input.NotEmpty, inputValidatorKeyName{})

	if err != nil {
		return false
	}

	outputPubKeyFile = "RPM-GPG-KEY-" + strings.ReplaceAll(name, " ", "-")

	email, err = input.Read("Email address", input.NotEmpty, inputValidatorEmail{})

	if err != nil {
		return false
	}

	password, err = input.ReadPasswordSecure(
		"Passphrase", input.NotEmpty, inputValidatorPassword{},
	)

	if err != nil {
		return false
	}

	return generateKeys(name, email, password, outputPubKeyFile)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// generatePrivateKey generates and saves private key file
func generateKeys(name, email string, password *secstr.String, outputPubKeyFile string) bool {
	spinner.Show("Generating keys")

	privKeyData, pubKeyData, err := keygen.Generate(name, email, password)

	if err != nil {
		spinner.Update("Can't generate key: %v", err.Error())
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

	fmtc.Printfn("{g}Private key saved as {*}%s{!}", outputPrivKeyFile)
	fmtc.Printfn("{g}Public key saved as {*}%s{!}", outputPubKeyFile)

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

type inputValidatorKeyName struct{}
type inputValidatorEmail struct{}
type inputValidatorPassword struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

// Validate validates key name input
func (v inputValidatorKeyName) Validate(input string) (string, error) {
	if !repoKeyNameValidator.MatchString(input) {
		return "", fmt.Errorf("Given name is invalid")
	}

	outputPubKeyFile := "RPM-GPG-KEY-" + strings.ReplaceAll(input, " ", "-")

	if fsutil.IsExist(outputPubKeyFile) {
		return "", fmt.Errorf("Public key file for given name (%s) already exists", outputPubKeyFile)
	}

	return input, nil
}

// Validate validates email input
func (v inputValidatorEmail) Validate(input string) (string, error) {
	if !emailValidator.MatchString(input) {
		return "", fmt.Errorf("Given email address is invalid")
	}

	return input, nil
}

// Validate validates key password input
func (v inputValidatorPassword) Validate(input string) (string, error) {
	if passwd.GetPasswordStrength(input) < passwd.STRENGTH_MEDIUM {
		return "", fmt.Errorf("Given passphrase is not strong enough")
	}

	return input, nil
}
