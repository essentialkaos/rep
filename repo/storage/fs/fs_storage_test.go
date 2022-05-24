package fs

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/essentialkaos/ek/v12/fsutil"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/index"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type StorageSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&StorageSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

var dataDir = "../../../testdata/testrepo"
var defRepos = []string{data.REPO_RELEASE, data.REPO_TESTING}
var defArchs = []string{data.ARCH_SRC, data.ARCH_X64}

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *StorageSuite) TestNewStorage(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	db, err := fs.GetDB(data.REPO_RELEASE, data.ARCH_X64, data.DB_PRIMARY)

	c.Assert(db, NotNil)
	c.Assert(err, IsNil)

	mtime := fs.GetModTime(data.REPO_RELEASE, data.ARCH_X64)

	c.Assert(mtime.IsZero(), Equals, false)

	c.Assert(fs.HasRepo(data.REPO_RELEASE), Equals, true)
	c.Assert(fs.HasRepo("unknown"), Equals, false)
	c.Assert(fs.HasArch(data.REPO_RELEASE, data.ARCH_NOARCH), Equals, true)
	c.Assert(fs.HasArch(data.REPO_RELEASE, "aarch64"), Equals, false)
	c.Assert(fs.HasArch(data.REPO_RELEASE, "somearch"), Equals, false)
	c.Assert(fs.HasArch("unknown", data.ARCH_NOARCH), Equals, false)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, data.ARCH_X64, "test-package-1.0.0-0.el7.x86_64.rpm"),
		Equals,
		"../../../testdata/testrepo/release/x86_64/test-package-1.0.0-0.el7.x86_64.rpm",
	)
}

func (s *StorageSuite) TestNewStorageErrors(c *C) {
	dopts := genStorageOptions(c, "")

	_, err := NewStorage(&Options{"", dopts.CacheDir, false, "", "", 0, 0}, index.DefaultOptions)
	c.Assert(err, ErrorMatches, `Can't create storage: Path to repository directory can't be empty`)

	_, err = NewStorage(&Options{dopts.DataDir, "", false, "", "", 0, 0}, index.DefaultOptions)
	c.Assert(err, ErrorMatches, `Can't create storage: Path to cache directory can't be empty`)

	_, err = NewStorage(&Options{dopts.DataDir, "/unknown", false, "", "", 0, 0}, index.DefaultOptions)
	c.Assert(err, ErrorMatches, `Can't create storage: Directory /unknown doesn't exist or not accessible`)

	_, err = NewStorage(dopts, nil)
	c.Assert(err, ErrorMatches, `Can't create storage: Index options cannot be nil`)

	_, err = NewStorage(nil, index.DefaultOptions)
	c.Assert(err, ErrorMatches, `Can't create storage: Data options cannot be nil`)

	opts := index.DefaultOptions.Clone()
	opts.CompressType = "unknown"

	_, err = NewStorage(dopts, opts)
	c.Assert(err, ErrorMatches, `Can't create storage: Unsupported compression method "unknown"`)
}

func (s *StorageSuite) TestStorageInitialize(c *C) {
	chownFunc = func(name string, uid, gid int) error { return nil }

	fs, err := NewStorage(genStorageOptions(c, "/_unknown_"), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(
		fs.Initialize(defRepos, []string{data.ARCH_X64}).Error(), Equals,
		`Can't initialize the new storage: The current user doesn't have enough permissions for creating new directories in "/"`,
	)

	fs.dataOptions.DataDir = ""
	c.Assert(
		fs.Initialize(defRepos, []string{data.ARCH_X64}).Error(), Equals,
		`Can't initialize the new storage: Data directory is not set (empty)`,
	)

	fs, err = NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	fs.dataOptions.User = "nobody"
	fs.dataOptions.Group = "nobody"
	fs.dataOptions.DirPerms = 0750
	fs.dataOptions.FilePerms = 0640

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.Initialize(nil, defArchs)
	c.Assert(err, ErrorMatches, `Can't initialize the new storage: At least one repository must be defined`)

	err = fs.Initialize(defRepos, nil)
	c.Assert(err, ErrorMatches, `Can't initialize the new storage: At least one architecture must be defined`)

	err = fs.Initialize(defRepos, []string{"unknown"})
	c.Assert(err, ErrorMatches, `Can't initialize the new storage: Unsupported architecture "unknown"`)

	err = fs.Initialize(defRepos, []string{data.ARCH_NOARCH})
	c.Assert(err, ErrorMatches, `Can't initialize the new storage: Unsupported architecture "noarch"`)

	err = fs.Initialize(defRepos, defArchs)
	c.Assert(err, IsNil)

	c.Assert(fsutil.CheckPerms("DWRX", fs.dataOptions.DataDir+"/testing/x86_64"), Equals, true)
	c.Assert(fsutil.CheckPerms("DWRX", fs.dataOptions.DataDir+"/testing/SRPMS"), Equals, true)
	c.Assert(fsutil.CheckPerms("DWRX", fs.dataOptions.DataDir+"/release/x86_64"), Equals, true)
	c.Assert(fsutil.CheckPerms("DWRX", fs.dataOptions.DataDir+"/release/SRPMS"), Equals, true)
	c.Assert(fsutil.GetMode(fs.dataOptions.DataDir+"/testing/x86_64"), Equals, os.FileMode(0750))
	c.Assert(fsutil.GetMode(fs.dataOptions.DataDir+"/testing/SRPMS"), Equals, os.FileMode(0750))
	c.Assert(fsutil.GetMode(fs.dataOptions.DataDir+"/release/x86_64"), Equals, os.FileMode(0750))
	c.Assert(fsutil.GetMode(fs.dataOptions.DataDir+"/release/SRPMS"), Equals, os.FileMode(0750))
	c.Assert(fs.IsEmpty(data.REPO_TESTING, data.ARCH_X64), Equals, true)
	c.Assert(fs.IsEmpty(data.REPO_TESTING, data.ARCH_NOARCH), Equals, true)

	fs, err = NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	os.MkdirAll(fs.dataOptions.DataDir+"/testing", 0755)
	os.MkdirAll(fs.dataOptions.DataDir+"/release", 0755)

	err = fs.Initialize(defRepos, defArchs)
	c.Assert(err, ErrorMatches, `Can't initialize the new storage: Storage already initialized`)

	fs, err = NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	chmodFunc = func(name string, mode os.FileMode) error { return fmt.Errorf("ERROR") }

	err = fs.Initialize(defRepos, defArchs)
	c.Assert(err, NotNil)

	mkdirFunc = func(name string, mode os.FileMode) error { return fmt.Errorf("ERROR") }

	err = fs.Initialize(defRepos, defArchs)
	c.Assert(err, NotNil)

	chownFunc = os.Chown
	chmodFunc = os.Chmod
	mkdirFunc = os.Mkdir
}

func (s *StorageSuite) TestAddPackage(c *C) {
	chownFunc = func(name string, uid, gid int) error { return nil }

	opts := genStorageOptions(c, "")

	opts.SplitFiles = true
	opts.User = "nobody"
	opts.Group = "nobody"
	opts.DirPerms = 0777
	opts.FilePerms = 0666

	fs, err := NewStorage(opts, index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.Initialize(defRepos, []string{data.ARCH_X64})

	c.Assert(err, IsNil)

	tempDir := c.MkDir()
	fsutil.CopyFile(
		"../../../testdata/comps.xml",
		tempDir+"/test-package-1.0.0-0.el7.x86_64.rpm",
		0644,
	)

	c.Assert(fs.AddPackage("", "/path/to/file"), ErrorMatches, `Can't add package to storage: Repository name can't be empty`)
	c.Assert(fs.AddPackage(data.REPO_TESTING, ""), ErrorMatches, `Can't add package to storage: Path to file can't be empty`)
	c.Assert(fs.AddPackage("unknown", "/pkgs/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't add package to storage: Repository "unknown" doesn't exist`)
	c.Assert(fs.AddPackage(data.REPO_RELEASE, "/pkgs/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't add package to storage: File /pkgs/test-package-1.0.0-0.el7.x86_64.rpm doesn't exist or not accessible`)
	c.Assert(fs.AddPackage(data.REPO_RELEASE, tempDir+"/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't add file to storage: .* is not an RPM package`)

	dp := fs.GetDepot(data.REPO_RELEASE, data.ARCH_X64)

	c.Assert(dp.AddPackage(""), ErrorMatches, `Can't add package to storage depot: Path to file can't be empty`)
	c.Assert(dp.AddPackage("/pkgs/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't add package to storage depot: File .*.rpm doesn't exist or not accessible`)
	c.Assert(dp.AddPackage(tempDir+"/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't add file to storage depot: .*.rpm is not an RPM package`)

	origDataDir := dp.dataDir
	dp.dataDir = "/unknown"
	c.Assert(dp.AddPackage("../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't add package to storage depot: mkdir /unknown/t: no such file or directory`)
	opts.SplitFiles = false
	c.Assert(dp.AddPackage("../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't copy package to storage depot: Can't copy file: Directory "/" is not writable`)
	dp.dataDir = origDataDir

	opts.SplitFiles = true

	// Add package 2 times
	c.Assert(dp.AddPackage("../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(dp.AddPackage("../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fsutil.IsExist(dp.dataDir+"/t/test-package-1.0.0-0.el7.x86_64.rpm"), Equals, true)

	c.Assert(fs.AddPackage(data.REPO_RELEASE, "../../../testdata/git-all-2.27.0-0.el7.noarch.rpm"), IsNil)
	c.Assert(fsutil.IsExist(dp.dataDir+"/g/git-all-2.27.0-0.el7.noarch.rpm"), Equals, true)

	c.Assert(fs.AddPackage(data.REPO_RELEASE, "unknown-package-1.0.0-0.el7.noarch.rpm"), ErrorMatches, `Can't add package to storage: File unknown-package-1.0.0-0.el7.noarch.rpm doesn't exist or not accessible`)

	_, err = dp.makePackageDir("пакет.x86_64.rpm")
	c.Assert(err, ErrorMatches, `Can't create directory for package: Can't use name "п" for directory`)

	chownFunc = func(name string, uid, gid int) error { return fmt.Errorf("ERROR") }
	chmodFunc = func(name string, mode os.FileMode) error { return fmt.Errorf("ERROR") }

	opts.SplitFiles = true
	_, err = dp.makePackageDir("abcd-package.rpm")
	c.Assert(err, ErrorMatches, `.*: ERROR`)
	opts.SplitFiles = false

	err = dp.copyFile("../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm", dp.dataDir)
	c.Assert(err, ErrorMatches, `.*: ERROR`)

	chownFunc = os.Chown
	chmodFunc = os.Chmod
}

func (s *StorageSuite) TestRemovePackage(c *C) {
	opts := genStorageOptions(c, "")
	fs, err := NewStorage(opts, index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.Initialize(defRepos, []string{data.ARCH_X64})

	c.Assert(err, IsNil)

	c.Assert(fs.RemovePackage("", "/path/to/file"), ErrorMatches, `Can't remove package from storage: Repository name can't be empty`)
	c.Assert(fs.RemovePackage(data.REPO_TESTING, ""), ErrorMatches, `Can't remove package from storage: Path to file can't be empty`)
	c.Assert(fs.RemovePackage(data.REPO_TESTING, "/path/to/file"), ErrorMatches, `Can't remove package from storage: Unknown RPM package architecture`)
	c.Assert(fs.RemovePackage("unknown", "test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't remove package from storage: Repository "unknown" doesn't exist`)

	dp := fs.GetDepot(data.REPO_RELEASE, data.ARCH_X64)

	c.Assert(dp.RemovePackage(""), ErrorMatches, `Can't remove package from storage depot: Path to file can't be empty`)
	c.Assert(dp.RemovePackage("test-package-1.0.0-0.el7.x86_64.rpm"), ErrorMatches, `Can't remove package from storage depot: File .*.rpm doesn't exist or not accessible`)

	fsutil.TouchFile(dp.dataDir+"/test-package-1.0.0-0.el7.x86_64.rpm", 0644)

	c.Assert(fs.RemovePackage(data.REPO_RELEASE, "test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fsutil.IsExist(dp.dataDir+"/test-package-1.0.0-0.el7.x86_64.rpm"), Equals, false)

	fsutil.TouchFile(dp.dataDir+"/test-package-1.0.0-0.el7.noarch.rpm", 0644)

	c.Assert(fs.RemovePackage(data.REPO_RELEASE, "test-package-1.0.0-0.el7.noarch.rpm"), IsNil)
	c.Assert(fsutil.IsExist(dp.dataDir+"/test-package-1.0.0-0.el7.noarch.rpm"), Equals, false)

	c.Assert(fs.RemovePackage(data.REPO_RELEASE, "test-package-1.0.1-0.el7.noarch.rpm"), ErrorMatches, `Can't remove package from storage depot: File .*.rpm doesn't exist or not accessible`)

	opts.SplitFiles = true

	os.MkdirAll(dp.dataDir+"/t", 0755)
	fsutil.TouchFile(dp.dataDir+"/t/test-package-1.0.0-0.el7.x86_64.rpm", 0644)
	fsutil.TouchFile(dp.dataDir+"/t/test-package-1.0.1-0.el7.x86_64.rpm", 0644)

	c.Assert(fs.RemovePackage(data.REPO_RELEASE, "t/test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fsutil.IsExist(dp.dataDir+"/t/test-package-1.0.0-0.el7.x86_64.rpm"), Equals, false)
	c.Assert(fsutil.IsExist(dp.dataDir+"/t"), Equals, true)
	c.Assert(fs.RemovePackage(data.REPO_RELEASE, "t/test-package-1.0.1-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fsutil.IsExist(dp.dataDir+"/t/test-package-1.0.1-0.el7.x86_64.rpm"), Equals, false)
	c.Assert(fsutil.IsExist(dp.dataDir+"/t"), Equals, false)

	c.Assert(dp.removePackageDir("test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)

	removeFunc = func(path string) error { return fmt.Errorf("ERROR") }
	opts.SplitFiles = false
	fsutil.TouchFile(dp.dataDir+"/test-package-1.0.0-0.el7.x86_64.rpm", 0644)
	err = fs.RemovePackage(data.REPO_RELEASE, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `.*: ERROR`)
	removeFunc = os.Remove
}

func (s *StorageSuite) TestCopyPackage(c *C) {
	opts := genStorageOptions(c, "")
	fs, err := NewStorage(opts, index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.Initialize(defRepos, []string{data.ARCH_X64})

	c.Assert(err, IsNil)

	c.Assert(fs.CopyPackage("", "", ""), ErrorMatches, `Can't copy package in storage: Source repository name is empty`)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, "", ""), ErrorMatches, `Can't copy package in storage: Target repository name is empty`)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, data.REPO_RELEASE, ""), ErrorMatches, `Can't copy package in storage: Path to file can't be empty`)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, data.REPO_RELEASE, "test-package-1.0.1-0.el7.ABCD.rpm"), ErrorMatches, `Can't copy package in storage: Unknown RPM package architecture`)
	c.Assert(fs.CopyPackage("unknown", data.REPO_RELEASE, "test-package-1.0.1-0.el7.x86_64.rpm"), ErrorMatches, `Can't copy package in storage: Source repository "unknown" doesn't exist`)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, "unknown", "test-package-1.0.1-0.el7.x86_64.rpm"), ErrorMatches, `Can't copy package in storage: Target repository "unknown" doesn't exist`)

	c.Assert(fs.AddPackage(data.REPO_TESTING, "../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fs.AddPackage(data.REPO_TESTING, "../../../testdata/git-all-2.27.0-0.el7.noarch.rpm"), IsNil)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, data.REPO_RELEASE, "test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, data.REPO_RELEASE, "git-all-2.27.0-0.el7.noarch.rpm"), IsNil)
	c.Assert(fsutil.IsExist(fs.dataOptions.DataDir+"/testing/x86_64/test-package-1.0.0-0.el7.x86_64.rpm"), Equals, true)
	c.Assert(fsutil.IsExist(fs.dataOptions.DataDir+"/testing/x86_64/git-all-2.27.0-0.el7.noarch.rpm"), Equals, true)

	c.Assert(fs.CopyPackage(data.REPO_TESTING, data.REPO_RELEASE, "test-package-1.0.1-0.el7.i386.rpm"), ErrorMatches, `Can't copy package in storage: Source repository "testing" don't support "i386" architecture`)
	c.Assert(os.Mkdir(fs.dataOptions.DataDir+"/testing/i386", 0755), IsNil)
	c.Assert(fs.CopyPackage(data.REPO_TESTING, data.REPO_RELEASE, "test-package-1.0.1-0.el7.i386.rpm"), ErrorMatches, `Can't copy package in storage: Target repository "release" don't support "i386" architecture`)
}

func (s *StorageSuite) TestHasPackage(c *C) {
	opts := genStorageOptions(c, "")
	opts.SplitFiles = true

	fs, err := NewStorage(opts, index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.HasPackage(data.REPO_TESTING, "test-package-1.0.0-0.el7.x86_64.rpm"), Equals, false)

	err = fs.Initialize(defRepos, []string{data.ARCH_X64})

	c.Assert(err, IsNil)

	c.Assert(fs.AddPackage(data.REPO_TESTING, "../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm"), IsNil)
	c.Assert(fs.AddPackage(data.REPO_TESTING, "../../../testdata/git-all-2.27.0-0.el7.noarch.rpm"), IsNil)

	c.Assert(fs.HasPackage("", "test-package-1.0.0-0.el7.x86_64.rpm"), Equals, false)
	c.Assert(fs.HasPackage(data.REPO_TESTING, ""), Equals, false)
	c.Assert(fs.HasPackage("unknown", "test-package-1.0.0-0.el7.x86_64.rpm"), Equals, false)

	c.Assert(fs.HasPackage(data.REPO_TESTING, "test-package-1.0.0-0.el7.x86_64.rpm"), Equals, true)
	c.Assert(fs.HasPackage(data.REPO_TESTING, "git-all-2.27.0-0.el7.noarch.rpm"), Equals, true)
	c.Assert(fs.HasPackage(data.REPO_TESTING, "test-package-2.0.0-0.el7.x86_64.rpm"), Equals, false)
	c.Assert(fs.HasPackage(data.REPO_TESTING, "test-package-2.0.0-0.el7.i386.rpm"), Equals, false)
	c.Assert(fs.HasPackage(data.REPO_TESTING, "git-all-2.28.0-0.el7.noarch.rpm"), Equals, false)

	dp := fs.GetDepot(data.REPO_TESTING, data.ARCH_X64)

	c.Assert(dp.getPackageDir("test-package-1.0.0-0.el7.x86_64.rpm"), Matches, `.*/testing/x86_64/t`)
	dp.dataOptions.SplitFiles = false
	c.Assert(dp.getPackageDir("test-package-1.0.0-0.el7.x86_64.rpm"), Matches, `.*/testing/x86_64`)
}

func (s *StorageSuite) TestUpdateObjectOwner(c *C) {
	chownFunc = func(name string, uid, gid int) error { return nil }
	chmodFunc = func(name string, mode os.FileMode) error { return nil }

	options := &Options{
		User:      "nobody",
		Group:     "nobody",
		DirPerms:  0777,
		FilePerms: 0666,
	}

	c.Assert(updateObjectAttrs("/path", options, true), IsNil)

	options.User = "_unknown_"
	err := updateObjectAttrs("/path", options, false)
	c.Assert(err, ErrorMatches, `Can't get UID for user "_unknown_"`)

	options.User = "nobody"
	options.Group = "_unknown_"
	err = updateObjectAttrs("/path", options, true)
	c.Assert(err, ErrorMatches, `Can't get GID for group "_unknown_"`)

	chownFunc = func(name string, uid, gid int) error { return fmt.Errorf("ERROR") }

	options.User = "nobody"
	options.Group = "nobody"
	c.Assert(updateObjectAttrs("/path", options, true), NotNil)

	chownFunc = os.Chown
	chmodFunc = os.Chmod
}

func (s *StorageSuite) TestStorageReindex(c *C) {
	fs, err := NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.Initialize(defRepos, []string{data.ARCH_X64})

	c.Assert(err, IsNil)

	err = fsutil.CopyFile(
		"../../../testdata/test-package-1.0.0-0.el7.x86_64.rpm",
		fs.dataOptions.DataDir+"/testing/x86_64/test-package-1.0.0-0.el7.x86_64.rpm",
		0644,
	)

	c.Assert(err, IsNil)

	c.Assert(fs.Reindex("", data.ARCH_X64, false), ErrorMatches, `Can't generate index: Repository name can't be empty`)
	c.Assert(fs.Reindex(data.REPO_TESTING, "", false), ErrorMatches, `Can't generate index: Arch name can't be empty`)
	c.Assert(fs.Reindex(data.REPO_TESTING, data.ARCH_NOARCH, false), ErrorMatches, `Can't generate index: Unsupported architecture "noarch"`)
	c.Assert(fs.Reindex(data.REPO_TESTING, "src", false), ErrorMatches, `Can't generate index: Repository "testing" doesn't contain "src" architecture`)
	c.Assert(fs.Reindex("unknown", data.ARCH_X64, false), ErrorMatches, `Can't generate index: Repository "unknown" doesn't exist`)

	err = fs.Reindex(data.REPO_TESTING, data.ARCH_X64, false)
	c.Assert(err, IsNil)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/filelists.sqlite.bz2"), Equals, true)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/filelists.xml.gz"), Equals, true)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/other.sqlite.bz2"), Equals, true)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/other.xml.gz"), Equals, true)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/primary.sqlite.bz2"), Equals, true)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/primary.xml.gz"), Equals, true)
	c.Assert(fsutil.CheckPerms("FRS", fs.dataOptions.DataDir+"/testing/x86_64/repodata/repomd.xml"), Equals, true)
}

func (s *StorageSuite) TestStorageIsEmpty(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.IsEmpty("", data.ARCH_X64), Equals, true)
	c.Assert(fs.IsEmpty("unknown", ""), Equals, true)
	c.Assert(fs.IsEmpty(data.REPO_RELEASE, data.ARCH_NOARCH), Equals, true)
	c.Assert(fs.IsEmpty("unknown", data.ARCH_X64), Equals, true)

	c.Assert(fs.IsEmpty(data.REPO_RELEASE, data.ARCH_X64), Equals, false)
	c.Assert(fs.IsEmpty(data.REPO_TESTING, data.ARCH_X64), Equals, true)
}

func (s *StorageSuite) TestStorageHasRepo(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.HasRepo(""), Equals, false)

	c.Assert(fs.HasRepo(data.REPO_RELEASE), Equals, true)
	c.Assert(fs.HasRepo("unknown"), Equals, false)
}

func (s *StorageSuite) TestStorageHasArch(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.HasArch("", data.ARCH_NOARCH), Equals, false)
	c.Assert(fs.HasArch(data.REPO_RELEASE, ""), Equals, false)
	c.Assert(fs.HasArch("unknown", data.ARCH_NOARCH), Equals, false)

	c.Assert(fs.HasArch(data.REPO_RELEASE, data.ARCH_NOARCH), Equals, true)
	c.Assert(fs.HasArch(data.REPO_RELEASE, "aarch64"), Equals, false)
	c.Assert(fs.HasArch(data.REPO_RELEASE, "somearch"), Equals, false)
	c.Assert(fs.IsEmpty(data.REPO_RELEASE, data.ARCH_X64), Equals, false)

	fs, err = NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = os.MkdirAll(fs.dataOptions.DataDir+"/release/SRPMS", 0755)

	c.Assert(err, IsNil)

	c.Assert(fs.HasArch(data.REPO_RELEASE, data.ARCH_NOARCH), Equals, false)
}

func (s *StorageSuite) TestStorageErrorsNotInitialized(c *C) {
	fs, err := NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	_, err = fs.GetDB(data.REPO_RELEASE, data.ARCH_X64, data.DB_PRIMARY)

	c.Assert(err, ErrorMatches, `Can't find DB connection: Repository storage is not initialized`)

	c.Assert(fs.GetModTime(data.REPO_RELEASE, data.ARCH_X64).IsZero(), Equals, true)
	c.Assert(fs.InvalidateCache(), ErrorMatches, `Can't invalidate cache: Repository storage is not initialized`)
	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), ErrorMatches, `Can't warmup cache: Repository storage is not initialized`)
	c.Assert(fs.HasArch(data.REPO_RELEASE, data.ARCH_X64), Equals, false)
	c.Assert(fs.HasRepo(data.REPO_RELEASE), Equals, false)
}

func (s *StorageSuite) TestStorageGetPackagePath(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, data.ARCH_X64, "test-package-1.0.0-0.el7.x86_64.rpm"),
		Equals, "../../../testdata/testrepo/release/x86_64/test-package-1.0.0-0.el7.x86_64.rpm",
	)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, data.ARCH_NOARCH, "test-package-1.0.0-0.el7.noarch.rpm"),
		Equals, "../../../testdata/testrepo/release/x86_64/test-package-1.0.0-0.el7.noarch.rpm",
	)

	c.Assert(
		fs.GetPackagePath("", data.ARCH_X64, "test-package-1.0.0-0.el7.x86_64.rpm"),
		Equals, "",
	)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, "", "test-package-1.0.0-0.el7.x86_64.rpm"),
		Equals, "",
	)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, data.ARCH_X64, ""),
		Equals, "",
	)

	c.Assert(
		fs.GetPackagePath("unknown", data.ARCH_X64, "test-package-1.0.0-0.el7.x86_64.rpm"),
		Equals, "",
	)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, "i686", "test-package-1.0.0-0.el7.x86_64.rpm"),
		Equals, "",
	)

	c.Assert(
		fs.GetPackagePath(data.REPO_RELEASE, data.ARCH_X64, "test-package-1.0.0-0.el7.x86_64.jpg"),
		Equals, "",
	)
}

func (s *StorageSuite) TestStorageGetDepot(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.GetDepot("", data.ARCH_X64), IsNil)
	c.Assert(fs.GetDepot(data.REPO_RELEASE, ""), IsNil)
	c.Assert(fs.GetDepot(data.REPO_RELEASE, data.ARCH_X64), NotNil)
}

func (s *StorageSuite) TestStorageGetBinDepot(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.GetBinDepot(data.REPO_RELEASE), NotNil)

	fs, err = NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.Initialize(defRepos, []string{data.ARCH_SRC})

	c.Assert(err, IsNil)

	c.Assert(fs.GetBinDepot(data.REPO_RELEASE), IsNil)
}

func (s *StorageSuite) TestStorageGetDB(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	dbc, err := fs.GetDB("", data.ARCH_X64, data.DB_PRIMARY)
	c.Assert(dbc, IsNil)
	c.Assert(err, NotNil)

	dbc, err = fs.GetDB(data.REPO_RELEASE, "", data.DB_PRIMARY)
	c.Assert(dbc, IsNil)
	c.Assert(err, NotNil)

	dbc, err = fs.GetDB(data.REPO_RELEASE, data.ARCH_X64, "")
	c.Assert(dbc, IsNil)
	c.Assert(err, NotNil)

	dbc, err = fs.GetDB(data.REPO_RELEASE, data.ARCH_X64, data.DB_PRIMARY)
	c.Assert(dbc, NotNil)
	c.Assert(err, IsNil)
}

func (s *StorageSuite) TestStorageGetModTime(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.GetModTime("", data.ARCH_X64).IsZero(), Equals, true)
	c.Assert(fs.GetModTime(data.REPO_RELEASE, "").IsZero(), Equals, true)

	c.Assert(fs.GetModTime(data.REPO_RELEASE, data.ARCH_X64).IsZero(), Equals, false)
}

func (s *StorageSuite) TestStorageWarmupCache(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache("", data.ARCH_X64), NotNil)
	c.Assert(fs.WarmupCache(data.REPO_RELEASE, ""), NotNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)
}

func (s *StorageSuite) TestStorageDBCaching(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	dp := fs.depots["release-x86_64"]

	c.Assert(dp, NotNil)
	c.Assert(dp.IsCacheValid(), Equals, true)
	c.Assert(dp.OpenDB(data.DB_PRIMARY), IsNil)
	c.Assert(dp.IsDBCached(data.DB_PRIMARY), Equals, true)
	c.Assert(dp.IsDBCached("unknown"), Equals, false)
	c.Assert(dp.CacheDB("unknown"), NotNil)

	c.Assert(fs.InvalidateCache(), IsNil)
	c.Assert(fs.PurgeCache(), IsNil)

	db, err := fs.GetDB(data.REPO_RELEASE, data.ARCH_X64, data.DB_PRIMARY)

	c.Assert(db, NotNil)
	c.Assert(err, IsNil)

	err = fs.PurgeCache()
	c.Assert(err, IsNil)

	err = RegisterFunc("filelists", "filelist_globber", filelistGlobberFunc, true)
	c.Assert(err, ErrorMatches, `Can't register new custom function after creating storage`)
}

func (s *StorageSuite) TestStoragePurgeCache(c *C) {
	fs, err := NewStorage(genStorageOptions(c, ""), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	err = fs.PurgeCache()

	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't purge cache: Repository storage is not initialized`)

	fs, err = NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)
	c.Assert(fs.PurgeCache(), IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	removeFunc = func(path string) error { return fmt.Errorf("ERROR") }
	c.Assert(fs.PurgeCache(), NotNil)
	c.Assert(fs.PurgeCache(), ErrorMatches, `ERROR`)
	removeFunc = os.Remove
}

func (s *StorageSuite) TestDepotIsCacheValid(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	dp := fs.depots["release-x86_64"]

	c.Assert(dp, NotNil)

	dp.dbs["test1"] = nil
	c.Assert(dp.CheckCache(), IsNil)
	c.Assert(dp.IsCacheValid(), Equals, true)
	delete(dp.dbs, "test1")

	origDataDir := dp.dataDir
	dp.dataDir = "/_unknown_"
	c.Assert(dp.CheckCache(), NotNil)
	c.Assert(dp.IsCacheValid(), Equals, false)
	dp.dataDir = origDataDir

	origDataDir = dp.dataDir
	dp.dataDir = c.MkDir()
	os.MkdirAll(dp.dataDir+"/repodata", 0755)
	fsutil.CopyDir(dataDir+"/release/x86_64/repodata", dp.dataDir+"/repodata")
	c.Assert(dp.IsCacheValid(), Equals, false)
	os.Chtimes(dp.dataDir+"/repodata/repomd.xml", time.Time{}, time.Unix(1644506277, 0))
	c.Assert(dp.IsCacheValid(), Equals, false)
	delete(dp.dbs, "filelists")
	delete(dp.dbs, "other")
	os.Remove(dp.dataDir + "/repodata/primary.sqlite.bz2")
	c.Assert(dp.IsCacheValid(), Equals, false)
	dp.dataDir = origDataDir
}

func (s *StorageSuite) TestDepotIsDBCached(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	dp := fs.depots["release-x86_64"]

	c.Assert(dp.IsDBCached(data.DB_PRIMARY), Equals, true)
	c.Assert(dp.IsDBCached("unknown"), Equals, false)

	fsutil.TouchFile(dp.GetDBFilePath("fake"), 0644)

	c.Assert(dp.IsDBCached("fake"), Equals, false)

	dp.meta.Get("primary_db").Timestamp = time.Now().Unix() + 3600

	c.Assert(dp.IsDBCached(data.DB_PRIMARY), Equals, false)
}

func (s *StorageSuite) TestDepotCacheDB(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	dp := fs.depots["release-x86_64"]

	c.Assert(dp.CacheDB(""), ErrorMatches, `Can't cache DB: DB type can't be empty`)
	c.Assert(dp.CacheDB("unknown"), ErrorMatches, `Can't cache DB: Can't find info about DB "unknown"`)

	origDataDir := dp.dataDir
	dp.dataDir = "/unknown"
	c.Assert(dp.CacheDB(data.DB_PRIMARY), ErrorMatches, `Can't cache DB: Can't find file with SQLite database "primary"`)
	dp.dataDir = origDataDir

	origDataDir = dp.dataDir
	dp.dataDir = c.MkDir()
	dbInfo := dp.meta.Get("primary_db")
	dbFile := dp.dataDir + "/" + dbInfo.Location.HREF
	os.MkdirAll(dp.dataDir+"/repodata", 0755)
	fsutil.TouchFile(dbFile, 0644)
	c.Assert(dp.CacheDB(data.DB_PRIMARY), ErrorMatches, `Can't cache DB: unexpected EOF`)
	dp.dataDir = origDataDir
}

func (s *StorageSuite) TestDepotGetMetaIndex(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	dp := fs.depots["release-x86_64"]

	c.Assert(dp, NotNil)

	meta, err := dp.GetMetaIndex()

	c.Assert(meta, NotNil)
	c.Assert(err, IsNil)

	dp.dataDir = "/_unknown_"

	meta, err = dp.GetMetaIndex()

	c.Assert(meta, IsNil)
	c.Assert(err, NotNil)
}

func (s *StorageSuite) TestDepotOpenDB(c *C) {
	fs, err := NewStorage(genStorageOptions(c, dataDir), index.DefaultOptions)

	c.Assert(fs, NotNil)
	c.Assert(err, IsNil)

	c.Assert(fs.WarmupCache(data.REPO_RELEASE, data.ARCH_X64), IsNil)

	dp := fs.depots["release-x86_64"]

	c.Assert(dp, NotNil)
	c.Assert(dp.OpenDB(data.DB_PRIMARY), IsNil)

	dp.dataOptions.CacheDir = "/_unknown_"
	c.Assert(dp.OpenDB(data.DB_PRIMARY), ErrorMatches, `Can't find file /_unknown_/release-x86_64-primary.sqlite`)
}

func (s *StorageSuite) TestFuncs(c *C) {
	c.Assert(filelistGlobberFunc("a/e", "a", "b/c/d", 0), Equals, false)
	c.Assert(filelistGlobberFunc("a/b", "a", "b/c/d", 0), Equals, true)

	c.Assert(filelistGlobberFunc("a/e", "a", "b/c/d", 1), Equals, true)
	c.Assert(filelistGlobberFunc("a/b", "a", "b/c/d", 1), Equals, false)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func genStorageOptions(c *C, dataDir string) *Options {
	if dataDir == "" {
		return &Options{c.MkDir() + "/testrepo", c.MkDir(), false, "", "", 0, 0}
	}

	return &Options{dataDir, c.MkDir(), false, "", "", 0, 0}
}
