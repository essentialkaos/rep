package sign

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"
	"testing"

	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/secstr"

	"golang.org/x/crypto/openpgp/packet"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type SignSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&SignSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *SignSuite) TestSigning(c *C) {
	srcDir := c.MkDir()
	trgDir := c.MkDir()

	srcPkg := srcDir + "/test-package-1.0.0-0.el7.x86_64.rpm"
	trgPkg := trgDir + "/test-package-1.0.0-0.el7.x86_64.rpm"

	fsutil.CopyFile("../../testdata/test-package-1.0.0-0.el7.x86_64.rpm", srcPkg, 0644)

	hasSign, err := HasSignature(srcPkg)

	c.Assert(hasSign, Equals, false)
	c.Assert(err, IsNil)

	key, err := ReadKey("../../testdata/reptest.private")

	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	privKey, err := key.Get(nil)

	c.Assert(privKey, NotNil)
	c.Assert(err, IsNil)

	// Private key is encrypted
	c.Assert(Sign(srcPkg, trgPkg, privKey), NotNil)

	_, err = IsSigned(srcPkg, privKey)
	c.Assert(err, NotNil)

	password, _ := secstr.NewSecureString("test1234TEST")
	privKey, err = key.Get(password)

	c.Assert(privKey, NotNil)
	c.Assert(err, IsNil)

	isSigned, err := IsSigned(srcPkg, privKey)

	c.Assert(isSigned, Equals, false)
	c.Assert(err, IsNil)

	c.Assert(Sign(srcPkg, trgPkg, privKey), IsNil)

	isSigned, err = IsSigned(trgPkg, privKey)

	c.Assert(isSigned, Equals, true)
	c.Assert(err, IsNil)
}

func (s *SignSuite) TestReadKey(c *C) {
	key, err := ReadKey("../../testdata/reptest.private")
	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	privKey, err := key.Get(nil)
	c.Assert(privKey, NotNil)
	c.Assert(err, IsNil)

	password, _ := secstr.NewSecureString("test1234TEST")
	privKey, err = key.Get(password)
	c.Assert(privKey, NotNil)
	c.Assert(err, IsNil)

	password, _ = secstr.NewSecureString("123")
	privKey, err = key.Get(password)
	c.Assert(err, ErrorMatches, "openpgp: invalid data: private key checksum failure")

	_, err = ReadKey("/_unknown_")
	c.Assert(err, ErrorMatches, "open /_unknown_: no such file or directory")

	tmpFile := c.MkDir() + "/key.private"
	ioutil.WriteFile(tmpFile, []byte("TEST"), 0640)
	_, err = ReadKey(tmpFile)
	c.Assert(err, ErrorMatches, "openpgp: invalid argument: no armored data found")

	key = &Key{false, []byte{}}
	_, err = key.Get(password)
	c.Assert(err, ErrorMatches, "Private key is empty")

	key = &Key{false, []byte("TEST")}
	_, err = key.Get(password)
	c.Assert(err, ErrorMatches, "openpgp: invalid argument: no armored data found")
}

func (s *SignSuite) TestErrors(c *C) {
	key, err := ReadKey("../../testdata/reptest.private")

	c.Assert(err, IsNil)
	c.Assert(key, NotNil)
	c.Assert(key.IsEncrypted, Equals, true)

	password, _ := secstr.NewSecureString("test1234TEST")

	_, err = HasSignature("_unknown_")
	c.Assert(err, NotNil)

	_, err = IsSigned("_unknown_", nil)
	c.Assert(err, NotNil)

	err = Sign("_unknown_", "_unknown_", nil)
	c.Assert(err, NotNil)

	privKey, _ := key.Get(password)

	_, err = IsSigned("_unknown_", privKey)
	c.Assert(err, NotNil)

	err = Sign("_unknown_", "_unknown_", privKey)
	c.Assert(err, NotNil)

	hdr, _ := readHeader("../../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	_, err = checkSignature(hdr, privKey)
	c.Assert(err, NotNil)

	_, err = checkSignaturePacket([]byte("ABCD"), nil)
	c.Assert(err, NotNil)

	sigV3 := &packet.SignatureV3{IssuerKeyId: 287}
	c.Assert(getSigKeyID(sigV3), Equals, uint64(287))

	c.Assert(getSigV4KeyID(nil), Equals, uint64(0))
	c.Assert(getSigV3KeyID(nil), Equals, uint64(0))

	sig := &packet.SymmetricKeyEncrypted{}
	c.Assert(getSigKeyID(sig), Equals, uint64(0))
}

// ////////////////////////////////////////////////////////////////////////////////// //
