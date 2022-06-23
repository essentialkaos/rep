package cli

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"

	"github.com/essentialkaos/ek/v12/fmtc"
	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/options"
	"github.com/essentialkaos/ek/v12/secstr"
	"github.com/essentialkaos/ek/v12/spinner"
	"github.com/essentialkaos/ek/v12/terminal"

	"github.com/essentialkaos/rep/repo/sign/keygen"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// outputKeyFile is default name for generated private key file
var outputKeyFile = "key.private"

// ////////////////////////////////////////////////////////////////////////////////// //

// cmdGenKey is 'gen-key' command handler
func cmdGenKey(ctx *context, args options.Arguments) bool {
	if fsutil.IsExist(outputKeyFile) {
		terminal.PrintErrorMessage("Private key file (%s) already exists", outputKeyFile)
		return false
	}

	name, err := terminal.ReadUI("Key name", true)

	if err != nil {
		return false
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

	return generatePrivateKey(name, email, password)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// generatePrivateKey generates and saves private key file
func generatePrivateKey(name, email string, password *secstr.String) bool {
	spinner.Show("Generating private key")

	data, err := keygen.Generate(name, email, password)

	if err != nil {
		spinner.Update(err.Error())
		spinner.Done(false)
		return false
	}

	err = ioutil.WriteFile(outputKeyFile, data, 0600)

	if err != nil {
		spinner.Update("Can't generate signing key: %v", err)
		spinner.Done(false)
		return false
	}

	spinner.Update("Private key saved as {*}%s{!}", outputKeyFile)
	spinner.Done(true)

	return true
}
