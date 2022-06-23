package keygen

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"fmt"

	"github.com/essentialkaos/ek/v12/secstr"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// keySize is private key size in bits
var keySize = 4096

// headers contains default key headers
var headers = map[string]string{
	"Version": "REP 3 (GPG/PGP Compatible)",
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Generate generates private key for signing packages
func Generate(name, email string, password *secstr.String) ([]byte, error) {
	if name == "" {
		return nil, fmt.Errorf("Can't generate signing key: name is empty")
	}

	if email == "" {
		return nil, fmt.Errorf("Can't generate signing key: email is empty")
	}

	if password == nil || password.IsEmpty() {
		return nil, fmt.Errorf("Can't generate signing key: password is empty")
	}

	data, err := generatePrivateKey(name, email, password)

	if err != nil {
		return nil, fmt.Errorf("Can't generate signing key: %w", err)
	}

	return data, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// generatePrivateKey generates private key
func generatePrivateKey(name, email string, password *secstr.String) ([]byte, error) {
	e, err := openpgp.NewEntity(name, "", email, &packet.Config{RSABits: keySize})

	if err != nil {
		return nil, err
	}

	for _, id := range e.Identities {
		err := id.SelfSignature.SignUserId(id.UserId.Id, e.PrimaryKey, e.PrivateKey, nil)

		if err != nil {
			return nil, err
		}
	}

	e.PrivateKey.Encrypt(password.Data)
	password.Destroy()

	buf := &bytes.Buffer{}
	w, err := armor.Encode(buf, openpgp.PrivateKeyType, headers)

	if err != nil {
		return nil, err
	}

	e.Serialize(w)
	w.Close()

	return buf.Bytes(), nil
}
