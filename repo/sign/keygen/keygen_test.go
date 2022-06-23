package keygen

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	"github.com/essentialkaos/ek/v12/secstr"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type KeygenSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&KeygenSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *KeygenSuite) TestErrors(c *C) {
	_, err := Generate("", "", nil)
	c.Assert(err, ErrorMatches, "Can't generate signing key: name is empty")

	_, err = Generate("test", "", nil)
	c.Assert(err, ErrorMatches, "Can't generate signing key: email is empty")

	_, err = Generate("test", "test@domain.com", nil)
	c.Assert(err, ErrorMatches, "Can't generate signing key: password is empty")

	p, _ := secstr.NewSecureString("")
	_, err = Generate("test", "test@domain.com", p)
	c.Assert(err, ErrorMatches, "Can't generate signing key: password is empty")

	p, _ = secstr.NewSecureString("test1234TEST")
	_, err = Generate("<<<<", "test@domain.com", p)
	c.Assert(err, ErrorMatches, "Can't generate signing key: openpgp: invalid argument: user id field contained invalid characters")
}

func (s *KeygenSuite) TestGeneration(c *C) {
	keySize = 1024

	p, _ := secstr.NewSecureString("test1234TEST")
	key, err := Generate("test", "test@domain.com", p)

	c.Assert(err, IsNil)
	c.Assert(key, NotNil)
}
