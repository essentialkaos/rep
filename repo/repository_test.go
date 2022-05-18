package repo

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"testing"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/index"
	"github.com/essentialkaos/rep/repo/search"
	"github.com/essentialkaos/rep/repo/sign"
	"github.com/essentialkaos/rep/repo/storage/fs"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type RepoSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&RepoSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *RepoSuite) TestNewRepository(c *C) {
	_, err := NewRepository("!!", nil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Name "!!" is invalid`)

	_, err = NewRepository("test", nil)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Storage is nil`)

	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)
}

func (s *RepoSuite) TestRepositoryInitialize(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, NotNil)

	c.Assert(r.HasArch(data.ARCH_X64), Equals, true)
	c.Assert(r.HasArch(data.ARCH_I686), Equals, false)
}

func (s *RepoSuite) TestRepositoryCopyPackage(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.CopyPackage(r.Testing, r.Release, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.CopyPackage(nil, r.Release, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Source sub-repository is nil`)

	err = r.CopyPackage(r.Testing, nil, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Target sub-repository is nil`)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.CopyPackage(r.Testing, r.Release, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestRepositoryInfo(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	_, _, err = r.Info("test-package-1.0.0-0.el7.x86_64.rpm", data.ARCH_X64)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	_, _, err = r.Info("test-package-1.0.0-0.el7.x86_64.rpm", "unknown")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Unknown or unsupported architecture "unknown"`)

	_, _, err = r.Info("test-package-1.0.0-0.el7.x86_64.rpm", data.ARCH_X64)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrEmptyRepo)

	err = r.CopyPackage(nil, r.Release, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Source sub-repository is nil`)

	err = r.CopyPackage(r.Testing, nil, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Target sub-repository is nil`)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.CopyPackage(r.Testing, r.Release, "test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.Testing.Reindex(false)
	c.Assert(err, IsNil)
	err = r.Release.Reindex(false)
	c.Assert(err, IsNil)

	_, _, err = r.Info("test.rpm", data.ARCH_X64)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't find package "test.rpm"`)

	pkg, mdt, err := r.Info("test-package", data.ARCH_X64)
	c.Assert(err, IsNil)
	c.Assert(pkg, NotNil)
	c.Assert(mdt.IsZero(), Equals, false)
}

func (s *RepoSuite) TestRepositorySigning(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	c.Assert(r.IsSigningRequired(), Equals, false)

	err = r.ReadSigningKey("/_unknown_")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `open /_unknown_: no such file or directory`)

	err = r.ReadSigningKey("../testdata/reptest.private")
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestRepositoryPurgeCache(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.PurgeCache()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't purge cache: Repository storage is not initialized`)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.PurgeCache()
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestSubRepositoryAddPackage(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Testing.AddPackage("")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't add package to repository: Path to file is empty`)

	err = r.Testing.AddPackage("test.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't add package to repository: Repository is not initialized`)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.Testing.AddPackage("test.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't add package to repository: File test.rpm doesn't exist or not accessible`)

	err = r.Testing.AddPackage("../testdata/comps.xml")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't add file to repository: ../testdata/comps.xml is not an RPM package`)

	err = r.ReadSigningKey("../testdata/reptest.private")
	c.Assert(err, IsNil)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't add file to repository: Repository allows only singed packages`)

	r.SigningKey = &sign.Key{}

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't add file to repository: Private key is empty`)

	r.SigningKey = nil

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestSubRepositoryRemovePackage(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Testing.RemovePackage("")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't remove package from repository: Path to file is empty`)

	err = r.Testing.RemovePackage("test.rpm")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't remove package from repository: Repository is not initialized`)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.Testing.RemovePackage("test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestSubRepositoryHasPackageFile(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	c.Assert(r.Testing.HasPackageFile(""), Equals, false)
	c.Assert(r.Testing.HasPackageFile("test.rpm"), Equals, false)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	c.Assert(r.Testing.HasPackageFile("test-package-1.0.0-0.el7.x86_64.rpm"), Equals, false)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	c.Assert(r.Testing.HasPackageFile("test-package-1.0.0-0.el7.x86_64.rpm"), Equals, true)
}

func (s *RepoSuite) TestSubRepositoryStats(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	_, err = r.Testing.Stats()
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_SRC, data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.Testing.Reindex(false)
	c.Assert(err, IsNil)

	stats, err := r.Testing.Stats()
	c.Assert(err, IsNil)
	c.Assert(stats, NotNil)

	c.Assert(stats.TotalPackages, Equals, 1)
	c.Assert(stats.TotalSize, Equals, int64(2288))
}

func (s *RepoSuite) TestSubRepositoryList(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	_, err = r.Testing.List("", false)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/git-all-2.27.0-0.el7.noarch.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.Reindex(false)
	c.Assert(err, IsNil)

	stk, err := r.Testing.List("", false)
	c.Assert(err, IsNil)
	c.Assert(stk, HasLen, 2)

	stk, err = r.Testing.List("", true)
	c.Assert(err, IsNil)
	c.Assert(stk, HasLen, 2)

	stk, err = r.Testing.List("git", false)
	c.Assert(err, IsNil)
	c.Assert(stk, HasLen, 1)
}

func (s *RepoSuite) TestSubRepositoryFind(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	_, err = r.Testing.Find(nil)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/git-all-2.27.0-0.el7.noarch.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.Reindex(false)
	c.Assert(err, IsNil)

	ps, err := r.Testing.Find(nil)
	c.Assert(err, IsNil)
	c.Assert(ps, HasLen, 0)

	ps, err = r.Testing.Find(search.Query{&search.Term{Value: false}})
	c.Assert(err, NotNil)

	ps, err = r.Testing.Find(search.Query{search.TermName("git-all")})
	c.Assert(err, IsNil)
	c.Assert(ps, HasLen, 1)

	ps, err = r.Testing.Find(search.Query{search.TermName("unknown")})
	c.Assert(err, IsNil)
	c.Assert(ps, HasLen, 0)
}

func (s *RepoSuite) TestSubRepositoryReindex(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Testing.Reindex(false)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/git-all-2.27.0-0.el7.noarch.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.Reindex(false)
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestSubRepositoryGetFullPackagePath(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	pkg := PackageFile{Arch: "x86_64", Path: "test-package-1.0.0-0.el7.x86_64.rpm"}
	c.Assert(r.Testing.GetFullPackagePath(pkg), Matches, `.*/data/testing/x86_64/test-package-1.0.0-0.el7.x86_64.rpm`)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func makeFSStorage(c *C) *fs.Storage {
	dir := c.MkDir()

	os.Mkdir(dir+"/data", 0755)
	os.Mkdir(dir+"/cache", 0755)

	fss, err := fs.NewStorage(
		&fs.Options{
			DataDir:    dir + "/data",
			CacheDir:   dir + "/cache",
			SplitFiles: false,
		},
		&index.Options{
			DirPerms:     0755,
			FilePerms:    0644,
			Pretty:       false,
			Update:       true,
			Split:        false,
			SkipSymlinks: false,
			Deltas:       false,
			MDFilenames:  index.MDF_SIMPLE,
			CompressType: index.COMPRESSION_BZ2,
			CheckSum:     index.CHECKSUM_SHA256,
			Workers:      0,
		},
	)

	if err != nil {
		c.Fatal(err.Error())
	}

	return fss
}
