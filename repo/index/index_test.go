package index

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/essentialkaos/ek/v12/fsutil"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type IndexSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&IndexSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *IndexSuite) SetUpSuite(c *C) {
	if !IsCreaterepoInstalled() {
		c.Fatal("createrepo_c is required for tests")
	}
}

func (s *IndexSuite) TestClone(c *C) {
	opts := &Options{
		GroupFile:      "/path/to/groups/file.xml",
		Pretty:         true,
		Update:         true,
		Split:          true,
		SkipSymlinks:   true,
		Deltas:         true,
		CheckSum:       CHECKSUM_SHA384,
		ChangelogLimit: 17,
		MDFilenames:    MDF_UNIQUE,
		Distro:         "cpeid,textname",
		Revision:       "c5af8a1",
		NumDeltas:      8,
		Workers:        11,
		CompressType:   COMPRESSION_XZ,
		Zchunk:         true,

		User:  "nobody",
		Group: "nobody",
	}

	optsCopy := opts.Clone()

	c.Assert(opts, DeepEquals, optsCopy)
}

func (s *IndexSuite) TestValidate(c *C) {
	tmpDir := c.MkDir()
	tmpFile := tmpDir + "/comps.xml"

	ioutil.WriteFile(tmpFile, []byte("TEST"), 0644)

	opts := &Options{
		GroupFile:      tmpFile,
		Pretty:         true,
		Update:         true,
		Split:          true,
		SkipSymlinks:   true,
		Deltas:         true,
		CheckSum:       CHECKSUM_SHA384,
		ChangelogLimit: 17,
		MDFilenames:    MDF_UNIQUE,
		Distro:         "cpeid,textname",
		Revision:       "c5af8a1",
		NumDeltas:      8,
		Workers:        11,
		CompressType:   COMPRESSION_XZ,
		Zchunk:         true,

		User:      "nobody",
		Group:     "nobody",
		DirPerms:  0700,
		FilePerms: 0600,
	}

	c.Assert(opts.Validate(), IsNil)

	opts.CompressType = "unknown"
	c.Assert(opts.Validate(), NotNil)

	opts.MDFilenames = "unknown"
	c.Assert(opts.Validate(), NotNil)

	opts.CheckSum = "unknown"
	c.Assert(opts.Validate(), NotNil)

	opts.ChangelogLimit = -1
	c.Assert(opts.Validate(), NotNil)

	opts.Workers = -1
	c.Assert(opts.Validate(), NotNil)

	opts.NumDeltas = -1
	c.Assert(opts.Validate(), NotNil)

	opts.Group = "unknown"
	c.Assert(opts.Validate(), NotNil)

	opts.User = "unknown"
	c.Assert(opts.Validate(), NotNil)

	opts.GroupFile = "/unknown/comps.xml"
	c.Assert(opts.Validate(), NotNil)
}

func (s *IndexSuite) TestToArg(c *C) {
	opts := &Options{
		GroupFile:      "/opt/rep/groups.xml",
		Pretty:         true,
		Update:         true,
		Split:          true,
		SkipSymlinks:   true,
		Deltas:         true,
		CheckSum:       CHECKSUM_SHA384,
		ChangelogLimit: 17,
		MDFilenames:    MDF_UNIQUE,
		Content:        "test",
		Distro:         "cpeid,textname",
		Revision:       "c5af8a1",
		NumDeltas:      8,
		Workers:        11,
		CompressType:   COMPRESSION_XZ,
		Zchunk:         true,
	}

	args := opts.ToArgs()

	c.Assert(args, DeepEquals, []string{
		"--database",
		"--groupfile=/opt/rep/groups.xml",
		"--checksum=sha384",
		"--pretty",
		"--update",
		"--split",
		"--skip-symlinks",
		"--deltas",
		"--changelog-limit=17",
		"--distro=cpeid,textname",
		"--content=test",
		"--revision=c5af8a1",
		"--num-deltas=8",
		"--workers=11",
		"--compress-type=xz",
		"--general-compress-type=xz",
		"--zck",
		"--unique-md-filenames",
	})

	opts.CompressType = ""
	opts.MDFilenames = ""

	args = opts.ToArgs()

	c.Assert(args, DeepEquals, []string{
		"--database",
		"--groupfile=/opt/rep/groups.xml",
		"--checksum=sha384",
		"--pretty",
		"--update",
		"--split",
		"--skip-symlinks",
		"--deltas",
		"--changelog-limit=17",
		"--distro=cpeid,textname",
		"--content=test",
		"--revision=c5af8a1",
		"--num-deltas=8",
		"--workers=11",
		"--compress-type=bz2",
		"--general-compress-type=bz2",
		"--zck",
		"--simple-md-filenames",
	})
}

func (s *IndexSuite) TestCreaterepo(c *C) {
	c.Assert(IsCreaterepoInstalled(), Equals, true)

	repoDir := c.MkDir()

	fsutil.CopyFile(
		"../../testdata/test-package-1.0.0-0.el7.x86_64.rpm",
		repoDir+"/test-package-1.0.0-0.el7.x86_64.rpm",
	)

	err := Generate(repoDir, DefaultOptions, true)

	c.Assert(err, IsNil)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/filelists.sqlite.xz"), Equals, true)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/filelists.xml.xz"), Equals, true)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/other.sqlite.xz"), Equals, true)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/other.xml.xz"), Equals, true)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/primary.sqlite.xz"), Equals, true)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/primary.xml.xz"), Equals, true)
	c.Assert(fsutil.IsExist(repoDir+"/repodata/repomd.xml"), Equals, true)
}

func (s *IndexSuite) TestCreaterepoWithChown(c *C) {
	c.Assert(IsCreaterepoInstalled(), Equals, true)

	repoDir := c.MkDir()

	fsutil.CopyFile(
		"../../testdata/test-package-1.0.0-0.el7.x86_64.rpm",
		repoDir+"/test-package-1.0.0-0.el7.x86_64.rpm",
	)

	Generate(repoDir, &Options{
		Update:       true,
		MDFilenames:  MDF_SIMPLE,
		CompressType: COMPRESSION_BZ2,
		User:         "nobody",
		Group:        "nobody",
	}, false)

	chownFunc = func(name string, uid, gid int) error {
		return nil
	}

	err := updateIndexOwner(repoDir, &Options{User: "unknown", Group: "nobody"})
	c.Assert(err, ErrorMatches, "Can't get UID for user \"unknown\"")

	err = updateIndexOwner(repoDir, &Options{User: "nobody", Group: "unknown"})
	c.Assert(err, ErrorMatches, "Can't get GID for group \"unknown\"")

	err = updateIndexOwner(repoDir, &Options{})
	c.Assert(err, IsNil)

	chownFunc = func(name string, uid, gid int) error {
		return errors.New("Chown error")
	}

	err = updateIndexOwner(repoDir, &Options{})
	c.Assert(err, ErrorMatches, "Chown error")

	chownFunc = os.Chown
}

func (s *IndexSuite) TestCreaterepoWithChmod(c *C) {
	c.Assert(IsCreaterepoInstalled(), Equals, true)

	repoDir := c.MkDir()

	fsutil.CopyFile(
		"../../testdata/test-package-1.0.0-0.el7.x86_64.rpm",
		repoDir+"/test-package-1.0.0-0.el7.x86_64.rpm",
	)

	Generate(repoDir, &Options{
		Update:       true,
		MDFilenames:  MDF_SIMPLE,
		CompressType: COMPRESSION_BZ2,
		DirPerms:     0700,
		FilePerms:    0600,
	}, false)

	err := updateIndexPerms(repoDir, &Options{})
	c.Assert(err, IsNil)

	chmodFunc = func(name string, mode os.FileMode) error {
		if mode == 0700 {
			return errors.New("Chmod dir error")
		}

		return nil
	}

	err = updateIndexPerms(repoDir, &Options{DirPerms: 0700})
	c.Assert(err, ErrorMatches, "Chmod dir error")

	chmodFunc = func(name string, mode os.FileMode) error {
		if mode == 0600 {
			return errors.New("Chmod file error")
		}

		return nil
	}

	err = updateIndexPerms(repoDir, &Options{FilePerms: 0600})
	c.Assert(err, ErrorMatches, "Chmod file error")

	chmodFunc = os.Chmod
}

func (s *IndexSuite) TestCreaterepoErrors(c *C) {
	err := Generate("/unknown", &Options{GroupFile: "/unknown"}, false)
	c.Assert(err, NotNil)

	err = Generate("/unknown", DefaultOptions.Clone(), false)
	c.Assert(err, NotNil)
}

func (s *IndexSuite) TestOptionsHelpers(c *C) {
	o := &Options{}

	c.Assert(o.GetDirPerms(), Equals, os.FileMode(0755))
	c.Assert(o.GetFilePerms(), Equals, os.FileMode(0644))

	o = &Options{DirPerms: 0700, FilePerms: 0600}

	c.Assert(o.GetDirPerms(), Equals, os.FileMode(0700))
	c.Assert(o.GetFilePerms(), Equals, os.FileMode(0600))
}
