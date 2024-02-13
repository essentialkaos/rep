package data

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type DataSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&DataSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *DataSuite) TestHelpers(c *C) {
	c.Assert(COMP_FLAG_ANY.String(), Equals, "")
	c.Assert(COMP_FLAG_EQ.String(), Equals, "EQ")
	c.Assert(COMP_FLAG_LT.String(), Equals, "LT")
	c.Assert(COMP_FLAG_LE.String(), Equals, "LE")
	c.Assert(COMP_FLAG_GT.String(), Equals, "GT")
	c.Assert(COMP_FLAG_GE.String(), Equals, "GE")

	c.Assert(ParseComp(""), Equals, COMP_FLAG_ANY)
	c.Assert(ParseComp("EQ"), Equals, COMP_FLAG_EQ)
	c.Assert(ParseComp("LT"), Equals, COMP_FLAG_LT)
	c.Assert(ParseComp("LE"), Equals, COMP_FLAG_LE)
	c.Assert(ParseComp("GT"), Equals, COMP_FLAG_GT)
	c.Assert(ParseComp("GE"), Equals, COMP_FLAG_GE)
}

func (s *DataSuite) TestPkgKeyIndex(c *C) {
	var index PkgKeyIndex

	c.Assert(index.HasData(), Equals, false)

	index = NewPkgKeyIndex()

	c.Assert(index, NotNil)

	c.Assert(index.HasData(), Equals, false)
	c.Assert(index.HasArch("test"), Equals, false)

	keyMap := NewPkgKeyMap()

	c.Assert(keyMap, NotNil)

	keyMap.Set(5)

	c.Assert(keyMap[5], Equals, true)
	c.Assert(keyMap[4], Equals, false)

	km1 := NewPkgKeyMap()
	km2 := NewPkgKeyMap()

	km1.Set(1)
	km1.Set(2)
	km1.Set(3)
	km2.Set(2)
	km2.Set(3)
	km2.Set(4)

	index.Intersect("test", km1)

	c.Assert(index["test"][1], Equals, true)
	c.Assert(index["test"][3], Equals, true)
	c.Assert(index["test"][9], Equals, false)

	index.Intersect("test", km2)

	c.Assert(index["test"][1], Equals, false)
	c.Assert(index["test"][3], Equals, true)
	c.Assert(index["test"][9], Equals, false)

	c.Assert(index.HasData(), Equals, true)
	c.Assert(index.HasArch("test"), Equals, true)

	c.Assert(index.List("unknown"), Equals, "")
	c.Assert(index.List("test"), Not(Equals), "")

	index.Drop("test")

	c.Assert(index.IgnoreArch("unknown"), Equals, false)
	c.Assert(index.IgnoreArch("test"), Equals, true)

	c.Assert(index.HasData(), Equals, false)
}

func (s *DataSuite) TestArchFlag(c *C) {
	var f ArchFlag

	c.Assert(f.Has(ARCH_FLAG_X64), Equals, false)

	f = ARCH_FLAG_I386 | ARCH_FLAG_X64 | ARCH_FLAG_NOARCH

	c.Assert(f.Has(ARCH_FLAG_I686), Equals, false)
	c.Assert(f.Has(ARCH_FLAG_X64), Equals, true)

	c.Assert(f.String(), Equals, "noarch/i386/x86_64")
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *DataSuite) BenchmarkList(c *C) {
	index := NewPkgKeyIndex()
	k := NewPkgKeyMap()

	k.Set(1)
	k.Set(2)
	k.Set(3)
	k.Set(4)
	k.Set(5)

	index.Intersect("test", k)

	for i := 0; i < c.N; i++ {
		index.List("test")
	}
}
