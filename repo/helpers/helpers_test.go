package helpers

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	"github.com/essentialkaos/rep/v3/repo/data"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type HelpersSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&HelpersSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *HelpersSuite) TestGuessFileArch(c *C) {
	c.Assert(GuessFileArch(""), Equals, "")
	c.Assert(GuessFileArch(".."), Equals, "")
	c.Assert(GuessFileArch(".rpm"), Equals, "")
	c.Assert(GuessFileArch("test.rpm"), Equals, "")

	c.Assert(GuessFileArch("test-package-1.0.0-0.el7.noarch.rpm"), Equals, data.ARCH_NOARCH)
	c.Assert(GuessFileArch("test-package-1.0.0-0.el7.src.rpm"), Equals, data.ARCH_SRC)
	c.Assert(GuessFileArch("test-package-1.0.0-0.el7.x86_64.rpm"), Equals, data.ARCH_X64)
	c.Assert(GuessFileArch("test-package-1.0.0-0.el7.i386.rpm"), Equals, data.ARCH_I386)
	c.Assert(GuessFileArch("test-package-1.0.0-0.el7.i686.rpm"), Equals, data.ARCH_I686)
	c.Assert(GuessFileArch("test-package-1.0.0-0.el7.aarch64.rpm"), Equals, data.ARCH_AARCH64)
}

func (s *HelpersSuite) TestExtractPackageArch(c *C) {
	_, err := ExtractPackageArch("/_unknown_")
	c.Assert(err, ErrorMatches, `open /_unknown_: no such file or directory`)

	_, err = ExtractPackageArch("../../testdata/comps.xml.gz")
	c.Assert(err, ErrorMatches, `file is not an RPM`)

	arch, err := ExtractPackageArch("../../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
	c.Assert(arch, Equals, "x86_64")

	arch, err = ExtractPackageArch("../../testdata/test-package-1.0.0-0.el7.src.rpm")
	c.Assert(err, IsNil)
	c.Assert(arch, Equals, "src")

	arch, err = ExtractPackageArch("../../testdata/git-all-2.27.0-0.el7.noarch.rpm")
	c.Assert(err, IsNil)
	c.Assert(arch, Equals, "noarch")
}
