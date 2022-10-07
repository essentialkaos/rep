package meta

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type MetaSuite struct {
	TmpDir string
}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&MetaSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

var repoDir = "../../testdata/testrepo/release/x86_64"
var metaFile = "../../testdata/testrepo/release/x86_64/repodata/repomd.xml"

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *MetaSuite) SetUpSuite(c *C) {
	s.TmpDir = c.MkDir()

	metaDate := time.Unix(1644506277, 0)
	err := os.Chtimes(metaFile, metaDate, metaDate)

	if err != nil {
		c.Fatalf("Can't set metadata mtime: %v", err)
	}
}

func (s *MetaSuite) TestReadingErrors(c *C) {
	index, err := Read("unknown.xml")

	c.Assert(err, NotNil)
	c.Assert(index, IsNil)

	err = ioutil.WriteFile(s.TmpDir+"/test.xml", []byte("TEST:TEST"), 0600)

	c.Assert(err, IsNil)

	index, err = Read(s.TmpDir + "/test.xml")

	c.Assert(err, NotNil)
	c.Assert(index, IsNil)
}

func (s *MetaSuite) TestReading(c *C) {
	index, err := Read(metaFile)

	c.Assert(err, IsNil)
	c.Assert(index, NotNil)

	c.Assert(index.Revision, Equals, int64(1644506277))

	info := index.Get(TYPE_PRIMARY_DB)

	c.Assert(info, NotNil)
	c.Assert(info.Type, Equals, TYPE_PRIMARY_DB)
	c.Assert(info.Checksum.Type, Equals, "sha256")
	c.Assert(info.Checksum.Hash, Equals, "2e59d48129f6fe0d2657182417610fc2e4fc5d436e70fe1e5dba8cc581f87d80")
	c.Assert(info.OpenChecksum.Type, Equals, "sha256")
	c.Assert(info.OpenChecksum.Hash, Equals, "2399625e0d40ae63ab1b5958f60a8c6d6304c44241fc1502155fdbfdd0f76cb3")
	c.Assert(info.Location.HREF, Equals, "repodata/primary.sqlite.bz2")
	c.Assert(info.DatabaseVersion, Equals, 10)
	c.Assert(info.Timestamp, Equals, int64(1644506277))
	c.Assert(info.Size, Equals, int64(1742))
	c.Assert(info.OpenSize, Equals, int64(106496))

	info = index.Get(TYPE_PRIMARY_ZCK)

	c.Assert(info.HeaderSize, Equals, int64(134))
}

func (s *MetaSuite) TestValidation(c *C) {
	index, err := Read(metaFile)

	c.Assert(err, IsNil)
	c.Assert(index, NotNil)

	c.Assert(index.Validate(""), NotNil)
	c.Assert(index.Get("primary_db").Validate(""), NotNil)

	c.Assert(index.Validate(repoDir), IsNil)

	index.Get("primary_db").Checksum.Hash = "0000"

	c.Assert(index.Validate(repoDir), NotNil)
}

func (s *MetaSuite) TestHelpers(c *C) {
	index, err := Read(metaFile)

	c.Assert(err, IsNil)
	c.Assert(index, NotNil)

	c.Assert(index.Get("primary_db"), NotNil)
	c.Assert(index.Get("primary_db1"), IsNil)

	c.Assert(index.Get("primary_db").Location.String(), Not(Equals), "")
}

func (s *MetaSuite) TestChecksumValidation(c *C) {
	c.Assert(ValidateChecksum("unknown.xml", "sha256", "0"), NotNil)

	c.Assert(ValidateChecksum(metaFile, "md5", "0"), NotNil)
	c.Assert(ValidateChecksum(metaFile, "sha1", "0"), NotNil)
	c.Assert(ValidateChecksum(metaFile, "sha224", "0"), NotNil)
	c.Assert(ValidateChecksum(metaFile, "sha256", "0"), NotNil)
	c.Assert(ValidateChecksum(metaFile, "sha384", "0"), NotNil)
	c.Assert(ValidateChecksum(metaFile, "sha512", "0"), NotNil)
	c.Assert(ValidateChecksum(metaFile, "shaXXX", "0"), NotNil)

	copyFunc = func(dst io.Writer, src io.Reader) (written int64, err error) {
		return 0, errors.New("Error")
	}

	c.Assert(ValidateChecksum(metaFile, "md5", "0"), NotNil)

	copyFunc = io.Copy
}
