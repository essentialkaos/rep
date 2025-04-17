package keygen

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"fmt"

	"github.com/essentialkaos/ek/v13/secstr"

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

// Generate generates private and public keys for signing packages
func Generate(name, email string, password *secstr.String) ([]byte, []byte, error) {
	if name == "" {
		return nil, nil, fmt.Errorf("Can't generate keys: name is empty")
	}

	if email == "" {
		return nil, nil, fmt.Errorf("Can't generate keys: email is empty")
	}

	if password == nil || password.IsEmpty() {
		return nil, nil, fmt.Errorf("Can't generate keys: password is empty")
	}

	e, err := generateKey(name, email, password)

	if err != nil {
		return nil, nil, fmt.Errorf("Can't generate keys: %w", err)
	}

	return exportPrivateKey(e), exportPublicKey(e), nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// generateKey generates private and public keys
func generateKey(name, email string, password *secstr.String) (*openpgp.Entity, error) {
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

	return e, nil
}

// exportPublicKey serializes public key data
func exportPublicKey(e *openpgp.Entity) []byte {
	buf := &bytes.Buffer{}
	w, _ := armor.Encode(buf, openpgp.PublicKeyType, headers)

	e.Serialize(w)
	w.Close()

	buf.WriteRune('\n')

	return buf.Bytes()
}

// exportPrivateKey serializes private key data
func exportPrivateKey(e *openpgp.Entity) []byte {
	buf := &bytes.Buffer{}
	w, _ := armor.Encode(buf, openpgp.PrivateKeyType, headers)

	e.SerializePrivateWithoutSigning(w, &packet.Config{RSABits: keySize})
	w.Close()

	buf.WriteRune('\n')

	return buf.Bytes()
}
