package sign

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"fmt"
	"os"

	"github.com/essentialkaos/ek/v12/directio"
	"github.com/essentialkaos/ek/v12/secstr"

	"github.com/sassoftware/go-rpmutils"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Key contains raw key data
type Key struct {
	IsEncrypted bool

	data []byte
}

// PrivateKey represents a possibly encrypted private key
type PrivateKey = packet.PrivateKey

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	ErrKeyIsEncrypted = fmt.Errorf("Private key is encrypted")
	ErrKeyIsNil       = fmt.Errorf("Private key is nil")
	ErrKeyIsEmpty     = fmt.Errorf("Private key is empty")
	ErrKeyringIsEmpty = fmt.Errorf("Keyring is empty (there is no private key)")
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Sign signs package with given private key
// Notice that encrypted private key must be decrypted before signing
func Sign(file, output string, privateKey *PrivateKey) error {
	if privateKey == nil {
		return ErrKeyIsNil
	}

	if privateKey.Encrypted {
		return ErrKeyIsEncrypted
	}

	f, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = rpmutils.SignRpmFile(f, output, privateKey, nil)

	return err
}

// IsSigned checks if package is signed with given key
func IsSigned(file string, privateKey *PrivateKey) (bool, error) {
	if privateKey == nil {
		return false, ErrKeyIsNil
	}

	hdr, err := readHeader(file)

	if err != nil {
		return false, err
	}

	if !hdr.HasTag(rpmutils.SIG_PGP) {
		return false, nil
	}

	return checkSignature(hdr, privateKey)
}

// HasSignature checks if package has PGP/GPG signature
func HasSignature(file string) (bool, error) {
	hdr, err := readHeader(file)

	if err != nil {
		return false, err
	}

	return hdr.HasTag(rpmutils.SIG_PGP), nil
}

// LoadKey loads private key data
func LoadKey(data []byte) (*Key, error) {
	k := &Key{IsEncrypted: false, data: data}
	pk, err := k.getKey()

	if err != nil {
		return nil, err
	}

	k.IsEncrypted = pk.Encrypted

	return k, nil
}

// ReadKey securely reads signing key from file
func ReadKey(file string) (*Key, error) {
	data, err := directio.ReadFile(file)

	if err != nil {
		return nil, err
	}

	return LoadKey(data)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Get decrypts raw signing key and returns private key
//
// You MUST NOT decrypt signing key (provide password) for checking package
// signature. Decrypted signing key ONLY required for signing packages.
func (k *Key) Get(password *secstr.String) (*PrivateKey, error) {
	if len(k.data) == 0 {
		return nil, ErrKeyIsEmpty
	}

	pk, err := k.getKey()

	if err != nil {
		return nil, err
	}

	if pk.Encrypted && password != nil && !password.IsEmpty() {
		err = pk.Decrypt(password.Data)

		if err != nil {
			return nil, err
		}
	}

	return pk, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getKey returns private key from keyring
func (k *Key) getKey() (*PrivateKey, error) {
	r := bytes.NewReader([]byte(k.data))
	kr, err := openpgp.ReadArmoredKeyRing(r)

	if err != nil {
		return nil, err
	}

	if kr[0].PrivateKey == nil {
		return nil, ErrKeyringIsEmpty
	}

	return kr[0].PrivateKey, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// readHeader reads RPM package header
func readHeader(file string) (*rpmutils.RpmHeader, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return rpmutils.ReadHeader(f)
}

// checkSignature checks signature from header and compare it with private key
func checkSignature(hdr *rpmutils.RpmHeader, privateKey *PrivateKey) (bool, error) {
	sigBlob, err := hdr.GetBytes(rpmutils.SIG_PGP)

	if err != nil {
		return false, fmt.Errorf("Can't read signature tag: %w", err)
	}

	return checkSignaturePacket(sigBlob, privateKey)
}

// checkSignaturePacket checks signature packet
func checkSignaturePacket(signature []byte, privateKey *PrivateKey) (bool, error) {
	packetReader := packet.NewReader(bytes.NewReader(signature))
	pkt, err := packetReader.Next()

	if err != nil {
		return false, fmt.Errorf("Can't decode signature: %w", err)
	}

	return getSigKeyID(pkt) == privateKey.KeyId, nil
}

// getSigKeyID returns signature key ID
func getSigKeyID(genPkt packet.Packet) uint64 {
	switch pkt := genPkt.(type) {
	case *packet.Signature:
		return getSigV4KeyID(pkt)
	case *packet.SignatureV3:
		return getSigV3KeyID(pkt)
	}

	return 0
}

// getSigKeyID returns signature V4 key ID
func getSigV4KeyID(pkt *packet.Signature) uint64 {
	if pkt != nil {
		return *pkt.IssuerKeyId
	}

	return 0
}

// getSigKeyID returns signature V3 key ID
func getSigV3KeyID(pkt *packet.SignatureV3) uint64 {
	if pkt != nil {
		return pkt.IssuerKeyId
	}

	return 0
}
