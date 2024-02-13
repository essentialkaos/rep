package sign

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
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

// ArmoredKey contains raw key data
type ArmoredKey struct {
	IsEncrypted bool

	data []byte
}

// Key contains parsed OpenGPG entity
type Key struct {
	entity *openpgp.Entity
}

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	ErrKeyIsEncrypted = fmt.Errorf("Key is encrypted (decrypted key is required)")
	ErrKeyIsNil       = fmt.Errorf("Key is nil")
	ErrKeyIsEmpty     = fmt.Errorf("Key is empty")
	ErrKeyringIsEmpty = fmt.Errorf("Keyring is empty (there is no private key)")
)

// ////////////////////////////////////////////////////////////////////////////////// //

// SignPackage signs package with given private key
// Notice that encrypted private key MUST BE decrypted before signing
func SignPackage(pkgFile, output string, key *Key) error {
	err := checkKey(key)

	if err != nil {
		return err
	}

	fd, err := os.OpenFile(pkgFile, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer fd.Close()

	_, err = rpmutils.SignRpmFile(fd, output, key.entity.PrivateKey, nil)

	return err
}

// SignFile generates asc file with PGP signature
func SignFile(file string, key *Key) error {
	err := checkKey(key)

	if err != nil {
		return err
	}

	srcFd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer srcFd.Close()

	outFd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer outFd.Close()

	return openpgp.ArmoredDetachSign(outFd, key.entity, srcFd, &packet.Config{})
}

// IsPackageSignatureValid checks if package is signed with given key
func IsPackageSignatureValid(pkgFile string, key *Key) (bool, error) {
	if key == nil || key.entity == nil || key.entity.PrivateKey == nil {
		return false, ErrKeyIsNil
	}

	hdr, err := readHeader(pkgFile)

	if err != nil {
		return false, err
	}

	if !hdr.HasTag(rpmutils.SIG_PGP) {
		return false, nil
	}

	return checkSignature(hdr, key.entity.PrivateKey)
}

// IsPackageSigned checks if package has PGP/GPG signature
func IsPackageSigned(pkgFile string) (bool, error) {
	hdr, err := readHeader(pkgFile)

	if err != nil {
		return false, err
	}

	return hdr.HasTag(rpmutils.SIG_PGP), nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// LoadKey loads private key data
func LoadKey(data []byte) (*ArmoredKey, error) {
	rk := &ArmoredKey{IsEncrypted: false, data: data}
	k, err := rk.Read(nil)

	if err != nil {
		return nil, err
	}

	rk.IsEncrypted = k.entity.PrivateKey.Encrypted

	return rk, nil
}

// ReadKey securely reads signing key from file
func ReadKey(file string) (*ArmoredKey, error) {
	data, err := directio.ReadFile(file)

	if err != nil {
		return nil, err
	}

	return LoadKey(data)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Read reads and decrypts (if password is provided) raw OpenPGP key
//
// You MUST NOT decrypt signing key (provide password) for checking package
// signature. Decrypted signing key ONLY required for signing packages/meta.
func (k *ArmoredKey) Read(password *secstr.String) (*Key, error) {
	if len(k.data) == 0 {
		return nil, ErrKeyIsEmpty
	}

	r := bytes.NewReader(k.data)
	kr, err := openpgp.ReadArmoredKeyRing(r)

	if err != nil {
		return nil, err
	}

	if kr[0].PrivateKey == nil {
		return nil, ErrKeyringIsEmpty
	}

	if kr[0].PrivateKey.Encrypted && password != nil && !password.IsEmpty() {
		err = kr[0].PrivateKey.Decrypt(password.Data)

		if err != nil {
			return nil, err
		}
	}

	return &Key{kr[0]}, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// checkKey checks key for problems
func checkKey(key *Key) error {
	if key == nil || key.entity == nil || key.entity.PrivateKey == nil {
		return ErrKeyIsNil
	}

	if key.entity.PrivateKey.Encrypted {
		return ErrKeyIsEncrypted
	}

	return nil
}

// readHeader reads RPM package header
func readHeader(pkgFile string) (*rpmutils.RpmHeader, error) {
	f, err := os.OpenFile(pkgFile, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return rpmutils.ReadHeader(f)
}

// checkSignature checks signature from header and compare it with private key
func checkSignature(hdr *rpmutils.RpmHeader, privateKey *packet.PrivateKey) (bool, error) {
	sigBlob, err := hdr.GetBytes(rpmutils.SIG_PGP)

	if err != nil {
		return false, fmt.Errorf("Can't read signature tag: %w", err)
	}

	return checkSignaturePacket(sigBlob, privateKey)
}

// checkSignaturePacket checks signature packet
func checkSignaturePacket(signature []byte, privateKey *packet.PrivateKey) (bool, error) {
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
