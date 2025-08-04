package utils

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/essentialkaos/ek/v13/hashutil"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type UtilsSuite struct {
	TmpDir string
}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&UtilsSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *UtilsSuite) SetUpSuite(c *C) {
	s.TmpDir = c.MkDir()
}

func (s *UtilsSuite) TestUnpack(c *C) {
	var err error

	hasher := sha256.New()
	dbFile := s.TmpDir + "/db.sqlite"
	dbHash := hashutil.File("../../../testdata/sqlite/db.sqlite", sha256.New())

	err = UnpackDB("../../../testdata/sqlite/db.sqlite", dbFile)

	c.Assert(err, IsNil)
	c.Assert(hashutil.File(dbFile, hasher).Equal(dbHash), Equals, true)

	err = UnpackDB("../../../testdata/sqlite/db.sqlite.gz", dbFile)

	c.Assert(err, IsNil)
	c.Assert(hashutil.File(dbFile, hasher).Equal(dbHash), Equals, true)

	err = UnpackDB("../../../testdata/sqlite/db.sqlite.bz2", dbFile)

	c.Assert(err, IsNil)
	c.Assert(hashutil.File(dbFile, hasher).Equal(dbHash), Equals, true)

	err = UnpackDB("../../../testdata/sqlite/db.sqlite.xz", dbFile)

	c.Assert(err, IsNil)
	c.Assert(hashutil.File(dbFile, hasher).Equal(dbHash), Equals, true)

	err = UnpackDB("../../../testdata/sqlite/db.sqlite.zst", dbFile)

	c.Assert(err, IsNil)
	c.Assert(hashutil.File(dbFile, hasher).Equal(dbHash), Equals, true)
}

func (s *UtilsSuite) TestUnpackErrors(c *C) {
	dbFile := s.TmpDir + "/db.sqlite"

	c.Assert(UnpackDB("/unknown.sqlite", dbFile), NotNil)
	c.Assert(UnpackDB("/unknown.jpg", dbFile), NotNil)
	c.Assert(UnpackDB("../../../testdata/sqlite/broken.sqlite.bz2", dbFile), NotNil)
	c.Assert(UnpackDB("../../../testdata/sqlite/db.sqlite", "."), NotNil)

	c.Assert(checkMagicHeader(bytes.NewBufferString("ABCD")), NotNil)
}
