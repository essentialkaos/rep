package rpm

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"strings"
	"testing"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type RPMSuite struct {
	TmpDir string
}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&RPMSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *RPMSuite) SetUpSuite(c *C) {
	s.TmpDir = c.MkDir()
}

func (s *RPMSuite) TestRPMCheck(c *C) {
	c.Assert(IsRPM("../../testdata/git-all-2.27.0-0.el7.noarch.rpm"), Equals, true)
}

func (s *RPMSuite) TestRPMLEADParsing(c *C) {
	lead, err := ReadLEAD("../../testdata/git-all-2.27.0-0.el7.noarch.rpm")

	c.Assert(err, IsNil)
	c.Assert(lead.Name, Equals, "git-all-2.27.0-0.el7")
	c.Assert(lead.ArchType, Equals, ARCH_X86_64)
	c.Assert(lead.OSType, Equals, OS_LINUX)
	c.Assert(lead.SigType, Equals, SIGTYPE_HEADERSIG)
	c.Assert(lead.Major, Equals, uint8(3))
	c.Assert(lead.Minor, Equals, uint8(0))
	c.Assert(lead.IsSrc, Equals, false)
}

func (s *RPMSuite) TestErrors(c *C) {
	p1 := s.TmpDir + "/package1.rpm"
	p2 := s.TmpDir + "/package2.rpm"
	p3 := s.TmpDir + "/package3.rpm"

	err := os.WriteFile(p2, []byte(""), 0644)

	if err != nil {
		c.Fatal(err.Error())
	}

	err = os.WriteFile(p3, []byte(strings.Repeat("DATA", 30)), 0644)

	if err != nil {
		c.Fatal(err.Error())
	}

	c.Assert(IsRPM(p1), Equals, false)
	c.Assert(IsRPM(p2), Equals, false)
	c.Assert(IsRPM(p3), Equals, false)

	_, err = ReadLEAD(p1)
	c.Assert(err, NotNil)

	_, err = ReadLEAD(p2)
	c.Assert(err, NotNil)

	_, err = ReadLEAD(p3)
	c.Assert(err, NotNil)
}
