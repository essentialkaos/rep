package sign

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"testing"

	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/secstr"

	"github.com/ProtonMail/go-crypto/openpgp/packet"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type SignSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&SignSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *SignSuite) TestPackageSigning(c *C) {
	srcDir := c.MkDir()
	trgDir := c.MkDir()

	srcPkg := srcDir + "/test-package-1.0.0-0.el7.x86_64.rpm"
	trgPkg := trgDir + "/test-package-1.0.0-0.el7.x86_64.rpm"

	fsutil.CopyFile("../../testdata/test-package-1.0.0-0.el7.x86_64.rpm", srcPkg, 0644)

	hasSign, err := IsPackageSigned(srcPkg)

	c.Assert(hasSign, Equals, false)
	c.Assert(err, IsNil)

	armKey, err := ReadKey("../../testdata/reptest.private")

	c.Assert(armKey, NotNil)
	c.Assert(err, IsNil)

	key, err := armKey.Read(nil)

	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	// Private key is encrypted
	c.Assert(SignPackage(srcPkg, trgPkg, key), NotNil)

	password, _ := secstr.NewSecureString("test1234TEST")
	key, err = armKey.Read(password)

	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	isSigned, err := IsPackageSignatureValid(srcPkg, key)

	c.Assert(isSigned, Equals, false)
	c.Assert(err, IsNil)

	c.Assert(SignPackage(srcPkg, trgPkg, key), IsNil)

	isSigned, err = IsPackageSignatureValid(trgPkg, key)

	c.Assert(isSigned, Equals, true)
	c.Assert(err, IsNil)
}

func (s *SignSuite) TestFileSigning(c *C) {
	tmpDir := c.MkDir()
	armKey, err := ReadKey("../../testdata/reptest.private")

	c.Assert(armKey, NotNil)
	c.Assert(err, IsNil)

	password, _ := secstr.NewSecureString("test1234TEST")
	key, err := armKey.Read(password)

	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	os.WriteFile(tmpDir+"/temp.txt", []byte("TEST1234ABCD!@#$"), 0644)

	c.Assert(SignFile(tmpDir+"/temp.txt", key), IsNil)

	c.Assert(SignFile(tmpDir+"/temp.txt", nil), NotNil)
	c.Assert(SignFile("_unknown_", key), NotNil)
	c.Assert(SignFile("/etc/passwd", key), NotNil)
}

func (s *SignSuite) TestReadKey(c *C) {
	armKey, err := ReadKey("../../testdata/reptest.private")
	c.Assert(armKey, NotNil)
	c.Assert(err, IsNil)

	key, err := armKey.Read(nil)
	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	password, _ := secstr.NewSecureString("test1234TEST")
	key, err = armKey.Read(password)
	c.Assert(key, NotNil)
	c.Assert(err, IsNil)

	password, _ = secstr.NewSecureString("123")
	key, err = armKey.Read(password)
	c.Assert(key, IsNil)
	c.Assert(err, ErrorMatches, "openpgp: invalid data: private key checksum failure")

	_, err = ReadKey("/_unknown_")
	c.Assert(err, ErrorMatches, "open /_unknown_: no such file or directory")

	tmpFile := c.MkDir() + "/key.private"
	os.WriteFile(tmpFile, []byte("TEST"), 0640)
	_, err = ReadKey(tmpFile)
	c.Assert(err, ErrorMatches, "openpgp: invalid argument: no armored data found")

	armKey = &ArmoredKey{false, []byte{}}
	_, err = armKey.Read(password)
	c.Assert(err, ErrorMatches, ErrKeyIsEmpty.Error())

	armKey = &ArmoredKey{false, []byte("TEST")}
	_, err = armKey.Read(password)
	c.Assert(err, ErrorMatches, "openpgp: invalid argument: no armored data found")
}

func (s *SignSuite) TestErrors(c *C) {
	_, err := ReadKey("../../testdata/empty.private")

	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrKeyringIsEmpty)

	armKey, err := ReadKey("../../testdata/reptest.private")

	c.Assert(err, IsNil)
	c.Assert(armKey, NotNil)
	c.Assert(armKey.IsEncrypted, Equals, true)

	password, _ := secstr.NewSecureString("test1234TEST")

	_, err = IsPackageSigned("_unknown_")
	c.Assert(err, NotNil)

	_, err = IsPackageSignatureValid("_unknown_", nil)
	c.Assert(err, NotNil)

	err = SignPackage("_unknown_", "_unknown_", nil)
	c.Assert(err, NotNil)

	key, _ := armKey.Read(password)

	_, err = IsPackageSignatureValid("_unknown_", key)
	c.Assert(err, NotNil)

	err = SignPackage("_unknown_", "_unknown_", key)
	c.Assert(err, NotNil)

	hdr, _ := readHeader("../../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	_, err = checkSignature(hdr, key)
	c.Assert(err, NotNil)

	_, err = checkSignaturePacket([]byte("ABCD"), nil)
	c.Assert(err, NotNil)

	_, err = isSignatureBaseOneKey(&packet.SymmetricKeyEncrypted{}, key)
	c.Assert(err, NotNil)
}

// ////////////////////////////////////////////////////////////////////////////////// //
