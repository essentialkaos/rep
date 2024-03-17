package keygen

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	"github.com/essentialkaos/ek/v12/secstr"

	"github.com/essentialkaos/rep/v3/repo/sign"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type KeygenSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&KeygenSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *KeygenSuite) TestErrors(c *C) {
	_, _, err := Generate("", "", nil)
	c.Assert(err, ErrorMatches, "Can't generate keys: name is empty")

	_, _, err = Generate("test", "", nil)
	c.Assert(err, ErrorMatches, "Can't generate keys: email is empty")

	_, _, err = Generate("test", "test@domain.com", nil)
	c.Assert(err, ErrorMatches, "Can't generate keys: password is empty")

	p, _ := secstr.NewSecureString("")
	_, _, err = Generate("test", "test@domain.com", p)
	c.Assert(err, ErrorMatches, "Can't generate keys: password is empty")
	p.Destroy()

	p, _ = secstr.NewSecureString("test1234TEST")
	_, _, err = Generate("<<<<", "test@domain.com", p)
	c.Assert(err, ErrorMatches, "Can't generate keys: openpgp: invalid argument: user id field contained invalid characters")
	p.Destroy()
}

func (s *KeygenSuite) TestGeneration(c *C) {
	keySize = 1024

	p, _ := secstr.NewSecureString("test1234TEST")
	privKeyData, pubKeyData, err := Generate("test", "test@domain.com", p)

	c.Assert(err, IsNil)
	c.Assert(privKeyData, NotNil)
	c.Assert(pubKeyData, NotNil)
}

func (s *KeygenSuite) TestFullCycle(c *C) {
	keySize = 1024

	p, _ := secstr.NewSecureString("test1234TEST")
	privKeyData, pubKeyData, err := Generate("test", "test@domain.com", p)

	c.Assert(err, IsNil)
	c.Assert(privKeyData, NotNil)
	c.Assert(pubKeyData, NotNil)

	armKey, err := sign.LoadKey(privKeyData)

	c.Assert(err, IsNil)
	c.Assert(armKey, NotNil)

	key, err := armKey.Read(p)

	c.Assert(err, IsNil)
	c.Assert(key, NotNil)
}
