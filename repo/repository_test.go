package repo

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

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

func (s *RepoSuite) TestPackage(c *C) {
	var p *Package

	c.Assert(p.FullName(), Equals, "")
	c.Assert(p.HasArch(data.ARCH_X64), Equals, false)

	p = &Package{
		Name:      "test-package",
		Version:   "1.0.0",
		Release:   "0.el7",
		ArchFlags: data.ARCH_FLAG_X64 | data.ARCH_FLAG_SRC,
	}

	c.Assert(p.FullName(), Equals, "test-package-1.0.0-0.el7")
	c.Assert(p.HasArch(data.ARCH_X64), Equals, true)
	c.Assert(p.HasArch(data.ARCH_I386), Equals, false)
	c.Assert(p.HasArch("abcd"), Equals, false)
}

func (s *RepoSuite) TestPackageFiles(c *C) {
	pf := PackageFiles{
		PackageFile{"test-package-1.0.0-0.el7.src.rpm", data.ARCH_FLAG_SRC, data.ARCH_FLAG_SRC},
		PackageFile{"test-package-1.0.0-0.el7.x86_64.rpm", data.ARCH_FLAG_X64, data.ARCH_FLAG_X64},
	}

	c.Assert(pf.HasArch(data.ARCH_SRC), Equals, true)
	c.Assert(pf.HasArch(data.ARCH_X64), Equals, true)
	c.Assert(pf.HasArch(""), Equals, false)
	c.Assert(pf.HasArch(data.ARCH_I386), Equals, false)
}

func (s *RepoSuite) TestPackageStack(c *C) {
	var ps PackageStack

	c.Assert(ps.HasMultiBundles(), Equals, false)
	c.Assert(ps.GetArchsFlag(), Equals, data.ARCH_FLAG_UNKNOWN)
	c.Assert(ps.GetArchs(), IsNil)
	c.Assert(ps.FlattenFiles(), IsNil)
	c.Assert(ps.IsEmpty(), Equals, true)

	ps = PackageStack{
		PackageBundle{
			&Package{
				Name:      "test-package",
				Version:   "1.0.0",
				Release:   "0.el7",
				ArchFlags: data.ARCH_FLAG_X64 | data.ARCH_FLAG_SRC,
				Files: PackageFiles{
					PackageFile{"test-package-1.0.0-0.el7.src.rpm", data.ARCH_FLAG_SRC, data.ARCH_FLAG_SRC},
					PackageFile{"test-package-1.0.0-0.el7.x86_64.rpm", data.ARCH_FLAG_X64, data.ARCH_FLAG_X64},
				},
			},
			&Package{
				Name:      "test-package",
				Version:   "1.0.1",
				Release:   "0.el7",
				ArchFlags: data.ARCH_FLAG_X64,
				Files: PackageFiles{
					PackageFile{"test-package-1.0.1-0.el7.x86_64.rpm", data.ARCH_FLAG_X64, data.ARCH_FLAG_X64},
				},
			},
		},
	}

	c.Assert(ps.HasMultiBundles(), Equals, true)
	c.Assert(ps.GetArchsFlag(), Equals, data.ARCH_FLAG_X64|data.ARCH_FLAG_SRC)
	c.Assert(ps.GetArchs(), DeepEquals, []string{"src", "x86_64"})
	c.Assert(ps.FlattenFiles(), DeepEquals, PackageFiles{
		PackageFile{"test-package-1.0.0-0.el7.src.rpm", data.ARCH_FLAG_SRC, data.ARCH_FLAG_SRC},
		PackageFile{"test-package-1.0.0-0.el7.x86_64.rpm", data.ARCH_FLAG_X64, data.ARCH_FLAG_X64},
		PackageFile{"test-package-1.0.1-0.el7.x86_64.rpm", data.ARCH_FLAG_X64, data.ARCH_FLAG_X64},
	})

	ps = PackageStack{
		PackageBundle{
			&Package{},
		},
	}

	c.Assert(ps.HasMultiBundles(), Equals, false)

	ps = PackageStack{
		PackageBundle{&Package{Name: "b", Version: "1.0.0", Release: "0.el7"}},
		PackageBundle{&Package{Name: "b", Version: "1.0.1", Release: "0.el7"}},
		PackageBundle{&Package{Name: "b", Version: "1.0.1", Release: "1.el7"}},
		PackageBundle{&Package{Name: "a", Version: "1.0.0", Release: "0.el7"}},
	}

	sort.Sort(ps)

	c.Assert(ps[0][0].FullName(), Equals, "a-1.0.0-0.el7")
	c.Assert(ps[1][0].FullName(), Equals, "b-1.0.0-0.el7")
	c.Assert(ps[2][0].FullName(), Equals, "b-1.0.1-0.el7")
	c.Assert(ps[3][0].FullName(), Equals, "b-1.0.1-1.el7")
}

func (s *RepoSuite) TestPayloadData(c *C) {
	pd := PayloadData{
		PayloadObject{false, "/d/test1"},
		PayloadObject{true, "/d/test2"},
		PayloadObject{true, "/c/test1"},
		PayloadObject{false, "/c/test2"},
		PayloadObject{false, "/b/test"},
		PayloadObject{false, "/a/test"},
	}

	sort.Sort(pd)
}

func (s *RepoSuite) TestRepositoryInitialize(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	c.Assert(r.HasArch(data.ARCH_X64), Equals, true)
	c.Assert(r.HasArch(data.ARCH_I686), Equals, false)

	c.Assert(r.Testing.Is(data.REPO_TESTING), Equals, true)
	c.Assert(r.Testing.Is(data.REPO_RELEASE), Equals, false)
}

func (s *RepoSuite) TestRepositoryCopyPackage(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	pkgFile := PackageFile{
		"test-package-1.0.0-0.el7.x86_64.rpm",
		data.ARCH_FLAG_X64, data.ARCH_FLAG_X64,
	}

	err = r.CopyPackage(r.Testing, r.Release, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.CopyPackage(nil, r.Release, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Source sub-repository is nil`)

	err = r.CopyPackage(r.Testing, nil, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Target sub-repository is nil`)

	err = r.CopyPackage(r.Testing, r.Release, PackageFile{})
	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrEmptyPath)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.CopyPackage(r.Testing, r.Release, pkgFile)
	c.Assert(err, IsNil)
}

func (s *RepoSuite) TestRepositoryIsPackageReleased(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	_, _, err = r.IsPackageReleased(nil)
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	pkgFile := PackageFile{
		"test-package-1.0.0-0.el7.x86_64.rpm",
		data.ARCH_FLAG_X64, data.ARCH_FLAG_X64,
	}

	err = r.CopyPackage(nil, r.Release, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Source sub-repository is nil`)

	err = r.CopyPackage(r.Testing, nil, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Target sub-repository is nil`)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.CopyPackage(r.Testing, r.Release, pkgFile)
	c.Assert(err, IsNil)

	err = r.Testing.Reindex(false, nil)
	c.Assert(err, IsNil)
	err = r.Release.Reindex(false, nil)
	c.Assert(err, IsNil)

	_, _, err = r.IsPackageReleased(nil)
	c.Assert(err, NotNil)
	c.Assert(err, Equals, ErrNilPackage)

	p := &Package{}
	ok, _, err := r.IsPackageReleased(p)
	c.Assert(err, IsNil)
	c.Assert(ok, Equals, false)

	p = &Package{ArchFlags: data.ARCH_FLAG_X64}
	ok, _, err = r.IsPackageReleased(p)
	c.Assert(err, IsNil)
	c.Assert(ok, Equals, false)

	p = &Package{ArchFlags: data.ARCH_FLAG_X64}
	ok, _, err = r.IsPackageReleased(p)
	c.Assert(err, IsNil)
	c.Assert(ok, Equals, false)

	r.storage = &FailStorage{}
	_, _, err = r.IsPackageReleased(p)
	c.Assert(err, NotNil)
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

	pkgFile := PackageFile{
		"test-package-1.0.0-0.el7.x86_64.rpm",
		data.ARCH_FLAG_X64, data.ARCH_FLAG_X64,
	}

	err = r.CopyPackage(nil, r.Release, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Source sub-repository is nil`)

	err = r.CopyPackage(r.Testing, nil, pkgFile)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Target sub-repository is nil`)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	err = r.CopyPackage(r.Testing, r.Release, pkgFile)
	c.Assert(err, IsNil)

	err = r.Testing.Reindex(false, nil)
	c.Assert(err, IsNil)
	err = r.Release.Reindex(false, nil)
	c.Assert(err, IsNil)

	_, _, err = r.Info("test.rpm", data.ARCH_X64)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't find package "test.rpm"`)

	pkg, mdt, err := r.Info("test-package", data.ARCH_X64)
	c.Assert(err, IsNil)
	c.Assert(pkg, NotNil)
	c.Assert(mdt.IsZero(), Equals, false)

	r.storage = &FailStorage{}
	_, _, err = r.Info("test-package", data.ARCH_X64)
	c.Assert(err, NotNil)

	_, err = r.Testing.collectPackageDepInfo("", "", "")
	c.Assert(err, NotNil)
	_, err = r.Testing.collectPackageFilesInfo("", "")
	c.Assert(err, NotNil)
	_, err = r.Testing.collectPackageChangelogInfo("", "")
	c.Assert(err, NotNil)
	_, _, err = r.Testing.collectPackageBasicInfo("", "")
	c.Assert(err, NotNil)
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

	err = r.Testing.RemovePackage(PackageFile{})
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't remove package from repository: Path to file is empty`)

	err = r.Testing.RemovePackage(PackageFile{Path: "test.rpm"})
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Can't remove package from repository: Repository is not initialized`)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)

	pkgFile := PackageFile{
		"test-package-1.0.0-0.el7.x86_64.rpm",
		data.ARCH_FLAG_X64, data.ARCH_FLAG_X64,
	}

	err = r.Testing.RemovePackage(pkgFile)
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

	err = r.Testing.Reindex(false, nil)
	c.Assert(err, IsNil)

	stats, err := r.Testing.Stats()
	c.Assert(err, IsNil)
	c.Assert(stats, NotNil)

	c.Assert(stats.TotalPackages, Equals, 1)
	c.Assert(stats.TotalSize, Equals, int64(2288))

	r.storage = &FailStorage{}
	_, err = r.Testing.Stats()
	c.Assert(err, NotNil)
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
	err = r.Testing.Reindex(false, nil)
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

	r.storage = &FailStorage{}
	_, err = r.Testing.List("git", false)
	c.Assert(err, NotNil)
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
	err = r.Testing.Reindex(false, nil)
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

	r.storage = &FailStorage{}
	_, err = r.Testing.Find(search.Query{search.TermName("git-all")})
	c.Assert(err, NotNil)
}

func (s *RepoSuite) TestSubRepositoryReindex(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Testing.Reindex(false, make(chan string, 99))
	c.Assert(err, NotNil)
	c.Assert(err, DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/test-package-1.0.0-0.el7.x86_64.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.AddPackage("../testdata/git-all-2.27.0-0.el7.noarch.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.Reindex(false, make(chan string, 99))
	c.Assert(err, IsNil)

	r.storage = &FailStorage{}
	err = r.Testing.Reindex(false, make(chan string, 99))
	c.Assert(err, NotNil)
}

func (s *RepoSuite) TestSubRepositoryCaching(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	c.Assert(r.Testing.IsCacheValid(), Equals, false)
	c.Assert(r.Testing.WarmupCache(), DeepEquals, ErrNotInitialized)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	err = r.Testing.AddPackage("../testdata/git-all-2.27.0-0.el7.noarch.rpm")
	c.Assert(err, IsNil)
	err = r.Testing.Reindex(true, nil)
	c.Assert(err, IsNil)

	c.Assert(r.Testing.IsCacheValid(), Equals, false)
	c.Assert(r.Testing.WarmupCache(), IsNil)
	c.Assert(r.Testing.IsCacheValid(), Equals, true)

	r.storage = &FailStorage{}

	c.Assert(r.Testing.WarmupCache(), NotNil)
}

func (s *RepoSuite) TestSubRepositoryGetFullPackagePath(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	pkg := PackageFile{"test-package-1.0.0-0.el7.x86_64.rpm", data.ARCH_FLAG_X64, data.ARCH_FLAG_X64}
	c.Assert(r.Testing.GetFullPackagePath(pkg), Matches, `.*/testing/x86_64/test-package-1.0.0-0.el7.x86_64.rpm`)
}

func (s *RepoSuite) TestSubRepositoryExecQuery(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	_, err = r.Testing.execQuery("", "abcd", "")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, `Unknown or unsupported arch "abcd"`)
}

func (s *RepoSuite) TestSubRepositoryGuessArch(c *C) {
	r, err := NewRepository("test", makeFSStorage(c))
	c.Assert(err, IsNil)
	c.Assert(r, NotNil)

	err = r.Initialize([]string{data.ARCH_X64})
	c.Assert(err, IsNil)

	c.Assert(r.Testing.guessArch(data.ARCH_X64), Equals, data.ARCH_X64)
	c.Assert(r.Testing.guessArch(data.ARCH_NOARCH), Equals, data.ARCH_X64)
}

func (s *RepoSuite) TestAux(c *C) {
	c.Assert(sanitizeInput(""), Equals, "")
	c.Assert(sanitizeInput("?'$"), Equals, "_ ")
}

// ////////////////////////////////////////////////////////////////////////////////// //

type FailStorage struct{}

func (s *FailStorage) Initialize(repoList, archList []string) error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) AddPackage(repo, rpmFilePath string) error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) RemovePackage(repo, arch, rpmFileRelPath string) error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) CopyPackage(fromRepo, toRepo, arch, rpmFileRelPath string) error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) IsInitialized() bool {
	return true
}

func (s *FailStorage) IsEmpty(repo, arch string) bool {
	return false
}

func (s *FailStorage) HasRepo(repo string) bool {
	return true
}

func (s *FailStorage) HasArch(repo, arch string) bool {
	return true
}

func (s *FailStorage) HasPackage(repo, arch, rpmFileName string) bool {
	return false
}

func (s *FailStorage) GetPackagePath(repo, arch, pkg string) string {
	return ""
}

func (s *FailStorage) Reindex(repo, arch string, full bool) error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) GetDB(repo, arch, dbType string) (*sql.DB, error) {
	return nil, fmt.Errorf("ERROR")
}

func (s *FailStorage) GetModTime(repo, arch string) (time.Time, error) {
	return time.Time{}, nil
}

func (s *FailStorage) InvalidateCache() error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) IsCacheValid(repo, arch string) bool {
	return false
}

func (s *FailStorage) PurgeCache() error {
	return fmt.Errorf("ERROR")
}

func (s *FailStorage) WarmupCache(repo, arch string) error {
	return fmt.Errorf("ERROR")
}

// ////////////////////////////////////////////////////////////////////////////////// //

func makeFSStorage(c *C) *fs.Storage {
	dir := c.MkDir()

	os.Mkdir(dir+"/data", 0755)
	os.Mkdir(dir+"/cache", 0755)

	fss, err := fs.NewStorage(
		&fs.Options{
			DataDir:    dir + "/data/testrepo",
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
