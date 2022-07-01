package fs

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
	"regexp"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/system"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/helpers"
	"github.com/essentialkaos/rep/repo/index"
	"github.com/essentialkaos/rep/repo/meta"
	"github.com/essentialkaos/rep/repo/rpm"
	"github.com/essentialkaos/rep/repo/storage/utils"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	PERMS_DIR  os.FileMode = 0755 // Default permissions for directories
	PERMS_FILE os.FileMode = 0644 // Default permissions for files
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Storage is repository storage
type Storage struct {
	dataOptions  *Options       // Data storage options
	indexOptions *index.Options // Index generation options

	depots DepotBundle // Map [repo name] → [depot]
}

// Options is storage options
type Options struct {
	DataDir  string // Path to directory with RPM files
	CacheDir string // Path to directory for cached data

	SplitFiles bool // Split files to separate directories

	User      string      // Repository data directory owner username
	Group     string      // Repository data directory owner group
	DirPerms  os.FileMode // Permissions for directories
	FilePerms os.FileMode // Permissions for files
}

// Depot is storage for specific repository (type + arch)
type Depot struct {
	id           string         // Repository ID (repo + - + arch)
	dataDir      string         // Path to sub-repository directory
	dataOptions  *Options       // Data storage options
	indexOptions *index.Options // Index generation options
	meta         *meta.Index    // Sub-repository metadata index
	dbs          DBBundle       // Map [db type] → [SQL connection]
}

// RepoStorageBundle is map [repo name] → [repo storage]
type DepotBundle map[string]*Depot

// DBBundle is map [db type] → [SQL connection]
type DBBundle map[string]*sql.DB

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	ErrNotInitialized = fmt.Errorf("Repository storage is not initialized")
	ErrEmptyRepoName  = fmt.Errorf("Repository name can't be empty")
	ErrEmptyPath      = fmt.Errorf("Path to file can't be empty")
	ErrEmptyArchName  = fmt.Errorf("Arch name can't be empty")
	ErrUnknownArch    = fmt.Errorf("Unknown RPM package architecture")
	ErrNilDepot       = fmt.Errorf("Can't find depot for given repository or architecture")
)

// DirNameValidatorRegex is directory name validation regexp
var DirNameValidatorRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	chownFunc  = os.Chown
	chmodFunc  = os.Chmod
	removeFunc = os.Remove
	mkdirFunc  = os.Mkdir
)

// ////////////////////////////////////////////////////////////////////////////////// //

// NewStorage creates new FS storage
func NewStorage(dataOptions *Options, indexOptions *index.Options) (*Storage, error) {
	switch {
	case dataOptions == nil:
		return nil, fmt.Errorf("Can't create storage: Data options cannot be nil")
	case indexOptions == nil:
		return nil, fmt.Errorf("Can't create storage: Index options cannot be nil")
	}

	err := dataOptions.Validate()

	if err != nil {
		return nil, fmt.Errorf("Can't create storage: %w", err)
	}

	err = indexOptions.Validate()

	if err != nil {
		return nil, fmt.Errorf("Can't create storage: %w", err)
	}

	// register new driver with custom globber if it not registered yet
	if !hasCustomDriver[data.DB_FILELISTS] {
		RegisterFunc(
			data.DB_FILELISTS, "filelist_globber",
			filelistGlobberFunc, true,
		)

		registerDrivers()
	}

	storage := &Storage{
		dataOptions:  dataOptions,
		indexOptions: indexOptions,
		depots:       make(DepotBundle),
	}

	return storage, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Validate validates storage data options
func (o *Options) Validate() error {
	err := checkDataDir(o.DataDir)

	if err != nil {
		return err
	}

	err = checkCacheDir(o.CacheDir)

	if err != nil {
		return err
	}

	return nil
}

// GetDirPerms returns permissions for directories
func (o *Options) GetDirPerms() os.FileMode {
	if o.DirPerms == 0 {
		return PERMS_DIR
	}

	return o.DirPerms
}

// GetFilePerms returns permissions for files
func (o *Options) GetFilePerms() os.FileMode {
	if o.FilePerms == 0 {
		return PERMS_FILE
	}

	return o.FilePerms
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Initialize initializes the new repository storage and creates all required directories
func (s *Storage) Initialize(repoList, archList []string) error {
	dataDirParent := path.Dir(s.dataOptions.DataDir)

	switch {
	case len(repoList) == 0:
		return fmt.Errorf("Can't initialize the new storage: At least one repository must be defined")
	case len(archList) == 0:
		return fmt.Errorf("Can't initialize the new storage: At least one architecture must be defined")
	case s.dataOptions.DataDir == "":
		return fmt.Errorf("Can't initialize the new storage: Data directory is not set (empty)")
	case !fsutil.CheckPerms("DWX", dataDirParent):
		return fmt.Errorf(
			"Can't initialize the new storage: The current user doesn't have enough permissions for creating new directories in %q",
			dataDirParent,
		)
	case s.IsInitialized():
		return fmt.Errorf("Can't initialize the new storage: Storage already initialized")
	}

	for _, arch := range archList {
		switch {
		case data.SupportedArchs[arch].Flag == data.ARCH_FLAG_UNKNOWN,
			data.SupportedArchs[arch].Flag == data.ARCH_FLAG_NOARCH:
			return fmt.Errorf("Can't initialize the new storage: Unsupported architecture %q", arch)
		}
	}

	dirList := []string{s.dataOptions.DataDir}

	for _, repo := range repoList {
		dirList = append(dirList, joinPath(s.dataOptions.DataDir, repo))
		for _, arch := range archList {
			dirList = append(
				dirList,
				joinPath(s.dataOptions.DataDir, repo, data.SupportedArchs[arch].Dir),
			)
		}
	}

	for _, dir := range dirList {
		err := mkdirFunc(dir, 0700)

		if err != nil {
			return fmt.Errorf("Can't initialize the new storage: %w", err)
		}

		err = updateObjectAttrs(dir, s.dataOptions, true)

		if err != nil {
			return fmt.Errorf("Can't initialize the new storage: %w", err)
		}
	}

	return nil
}

// AddPackage adds package file to the given repository
// Important: This method DO NOT run repository reindex
func (s *Storage) AddPackage(repo, rpmFilePath string) error {
	switch {
	case repo == "":
		return fmt.Errorf("Can't add package to storage: %w", ErrEmptyRepoName)
	case rpmFilePath == "":
		return fmt.Errorf("Can't add package to storage: %w", ErrEmptyPath)
	case !s.HasRepo(repo):
		return fmt.Errorf("Can't add package to storage: Repository %q doesn't exist", repo)
	}

	err := fsutil.ValidatePerms("FRS", rpmFilePath)

	if err != nil {
		return fmt.Errorf("Can't add package to storage: %w", err)
	}

	if !rpm.IsRPM(rpmFilePath) {
		return fmt.Errorf("Can't add file to storage: %s is not an RPM package", rpmFilePath)
	}

	arch, err := helpers.ExtractPackageArch(rpmFilePath)

	if err != nil {
		return fmt.Errorf("Can't extract package architecture tag: %w", err)
	}

	if data.SupportedArchs[arch].Flag == data.ARCH_FLAG_UNKNOWN {
		return fmt.Errorf("Unsupported package architecture %q", arch)
	}

	if arch != data.ARCH_NOARCH {
		return s.GetDepot(repo, arch).AddPackage(rpmFilePath)
	}

	for _, a := range data.BinArchList {
		if !s.HasArch(repo, a) {
			continue
		}

		err := s.GetDepot(repo, a).AddPackage(rpmFilePath)

		if err != nil {
			return err
		}
	}

	return nil
}

// RemovePackage removes package with given relative path from the given repository
// Important: This method DO NOT run repository reindex
func (s *Storage) RemovePackage(repo, rpmFileRelPath string) error {
	arch := helpers.GuessFileArch(rpmFileRelPath)

	switch {
	case repo == "":
		return fmt.Errorf("Can't remove package from storage: %w", ErrEmptyRepoName)
	case rpmFileRelPath == "":
		return fmt.Errorf("Can't remove package from storage: %w", ErrEmptyPath)
	case arch == "":
		return fmt.Errorf("Can't remove package from storage: %w", ErrUnknownArch)
	case !s.HasRepo(repo):
		return fmt.Errorf("Can't remove package from storage: Repository %q doesn't exist", repo)
	}

	if arch != data.ARCH_NOARCH {
		return s.GetDepot(repo, arch).RemovePackage(rpmFileRelPath)
	}

	for _, a := range data.BinArchList {
		if !s.HasArch(repo, a) {
			continue
		}

		err := s.GetDepot(repo, a).RemovePackage(rpmFileRelPath)

		if err != nil {
			return err
		}
	}

	return nil
}

// CopyPackage copies file from one repository to another
// Important: This method DO NOT run repository reindex
func (s *Storage) CopyPackage(fromRepo, toRepo, rpmFileRelPath string) error {
	arch := helpers.GuessFileArch(rpmFileRelPath)

	switch {
	case fromRepo == "":
		return fmt.Errorf("Can't copy package in storage: Source repository name is empty")
	case toRepo == "":
		return fmt.Errorf("Can't copy package in storage: Target repository name is empty")
	case rpmFileRelPath == "":
		return fmt.Errorf("Can't copy package in storage: %w", ErrEmptyPath)
	case arch == "":
		return fmt.Errorf("Can't copy package in storage: %w", ErrUnknownArch)
	case !s.HasRepo(fromRepo):
		return fmt.Errorf("Can't copy package in storage: Source repository %q doesn't exist", fromRepo)
	case !s.HasRepo(toRepo):
		return fmt.Errorf("Can't copy package in storage: Target repository %q doesn't exist", toRepo)
	case arch != data.ARCH_NOARCH && !s.HasArch(fromRepo, arch):
		return fmt.Errorf("Can't copy package in storage: Source repository %q don't support %q architecture", fromRepo, arch)
	case arch != data.ARCH_NOARCH && !s.HasArch(toRepo, arch):
		return fmt.Errorf("Can't copy package in storage: Target repository %q don't support %q architecture", toRepo, arch)
	}

	var depot *Depot

	if arch != data.ARCH_NOARCH {
		depot = s.GetDepot(fromRepo, arch)
	} else {
		depot = s.GetBinDepot(fromRepo)
	}

	return s.AddPackage(toRepo, depot.GetPackagePath(rpmFileRelPath))
}

// Reindex generates index metadata for the given repository and arch
func (s *Storage) Reindex(repo, arch string, full bool) error {
	switch {
	case repo == "":
		return fmt.Errorf("Can't generate index: %w", ErrEmptyRepoName)
	case arch == "":
		return fmt.Errorf("Can't generate index: %w", ErrEmptyArchName)
	case arch == data.ARCH_NOARCH:
		return fmt.Errorf("Can't generate index: Unsupported architecture %q", arch)
	case !s.HasRepo(repo):
		return fmt.Errorf("Can't generate index: Repository %q doesn't exist", repo)
	case !s.HasArch(repo, arch):
		return fmt.Errorf("Can't generate index: Repository %q doesn't contain %q architecture", repo, arch)
	}

	return s.GetDepot(repo, arch).Reindex(full)
}

// IsInitialized returns true if repository already initialized and ready for work
func (s *Storage) IsInitialized() bool {
	return len(fsutil.List(s.dataOptions.DataDir, true, fsutil.ListingFilter{Perms: "DR"})) != 0
}

// IsEmpty returns true if repository is empty (no packages)
func (s *Storage) IsEmpty(repo, arch string) bool {
	switch {
	case repo == "", arch == "":
		return true
	case arch == data.ARCH_NOARCH:
		return true
	case !s.HasRepo(repo):
		return true
	}

	return s.GetDepot(repo, arch).IsEmpty()
}

// HasRepo returns true if given repositry exists
func (s *Storage) HasRepo(repo string) bool {
	if repo == "" {
		return false
	}

	return fsutil.IsExist(joinPath(s.dataOptions.DataDir, repo))
}

// HasArch returns true if repository storage contains directory for specific arch
func (s *Storage) HasArch(repo, arch string) bool {
	switch {
	case repo == "", arch == "":
		return false
	case !s.HasRepo(repo):
		return false
	}

	repoDir := joinPath(s.dataOptions.DataDir, repo)

	if arch == data.ARCH_NOARCH {
		for _, archInfo := range data.SupportedArchs {
			switch archInfo.Flag {
			case data.ARCH_FLAG_SRC, data.ARCH_FLAG_NOARCH:
				continue
			}

			if fsutil.IsExist(joinPath(repoDir, archInfo.Dir)) {
				return true
			}
		}

		return false
	}

	archInfo, hasArch := data.SupportedArchs[arch]

	if !hasArch {
		return false
	}

	return fsutil.IsExist(joinPath(repoDir, archInfo.Dir))
}

// HasPackage checks if repository contains file with given name
func (s *Storage) HasPackage(repo, rpmFileName string) bool {
	arch := helpers.GuessFileArch(rpmFileName)

	switch {
	case repo == "", rpmFileName == "", arch == "":
		return false
	case !s.IsInitialized():
		return false
	case !s.HasRepo(repo):
		return false
	case arch != data.ARCH_NOARCH && !s.HasArch(repo, arch):
		return false
	}

	var depot *Depot

	if arch != data.ARCH_NOARCH {
		depot = s.GetDepot(repo, arch)
	} else {
		depot = s.GetBinDepot(repo)
	}

	return depot.HasPackage(rpmFileName)
}

// GetPackagePath returns full path to package RPM file
func (s *Storage) GetPackagePath(repo, arch, rpmFileRelPath string) string {
	switch {
	case repo == "", arch == "", rpmFileRelPath == "":
		return ""
	case !strings.HasSuffix(rpmFileRelPath, ".rpm"):
		return ""
	case !s.HasRepo(repo):
		return ""
	case arch != data.ARCH_NOARCH && !s.HasArch(repo, arch):
		return ""
	}

	var depot *Depot

	if arch != data.ARCH_NOARCH {
		depot = s.GetDepot(repo, arch)
	} else {
		depot = s.GetBinDepot(repo)
	}

	return depot.GetPackagePath(rpmFileRelPath)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// GetDB returns connection to SQLite DB
func (s *Storage) GetDB(repo, arch, dbType string) (*sql.DB, error) {
	switch {
	case repo == "":
		return nil, fmt.Errorf("Can't find DB connection: %w", ErrEmptyRepoName)
	case arch == "":
		return nil, fmt.Errorf("Can't find DB connection: %w", ErrEmptyArchName)
	case dbType == "":
		return nil, fmt.Errorf("Can't find DB connection: DB type can't be empty")
	case !s.IsInitialized():
		return nil, fmt.Errorf("Can't find DB connection: %w", ErrNotInitialized)
	}

	return s.GetDepot(repo, arch).GetDB(dbType)
}

// GetDepot creates new depot or returns one from the cache
func (s *Storage) GetDepot(repo, arch string) *Depot {
	if repo == "" || arch == "" {
		return nil
	}

	id := repo + "-" + arch
	depot := s.depots[id]

	if depot != nil {
		return depot
	}

	depot = &Depot{
		id:           id,
		dataOptions:  s.dataOptions,
		indexOptions: s.indexOptions,
		dataDir:      joinPath(s.dataOptions.DataDir, repo, data.SupportedArchs[arch].Dir),
		dbs:          make(map[string]*sql.DB),
	}

	s.depots[id] = depot

	return depot
}

// GetBinDepot returns any depot with binary packages (useful for working with noarch packages)
func (s *Storage) GetBinDepot(repo string) *Depot {
	for _, a := range data.BinArchList {
		if !s.HasArch(repo, a) {
			continue
		}

		return s.GetDepot(repo, a)
	}

	return nil
}

// GetModTime returns date of repository index modification
func (s *Storage) GetModTime(repo, arch string) time.Time {
	switch {
	case repo == "":
		return time.Time{}
	case arch == "":
		return time.Time{}
	case !s.IsInitialized():
		return time.Time{}
	}

	indexFile := joinPath(s.dataOptions.DataDir, repo, arch, "/repodata/repomd.xml")
	mTime, _ := fsutil.GetMTime(indexFile)

	return mTime
}

// InvalidateCache invalidates cache and removes SQLite files from cache directory
func (s *Storage) InvalidateCache() error {
	if !s.IsInitialized() {
		return fmt.Errorf("Can't invalidate cache: %w", ErrNotInitialized)
	}

	for _, depot := range s.depots {
		err := depot.InvalidateCache()

		if err != nil {
			return err
		}
	}

	return nil
}

// IsCacheValid returns true if cache is valid
func (s *Storage) IsCacheValid(repo, arch string) bool {
	return s.GetDepot(repo, arch).IsCacheValid()
}

// PurgeCache deletes all SQLite files from cache directory
func (s *Storage) PurgeCache() error {
	if !s.IsInitialized() {
		return fmt.Errorf("Can't purge cache: %w", ErrNotInitialized)
	}

	files := fsutil.List(s.dataOptions.CacheDir, true, fsutil.ListingFilter{
		MatchPatterns: []string{"*.sqlite"},
	})

	fsutil.ListToAbsolute(s.dataOptions.CacheDir, files)

	for _, sqlFile := range files {
		err := removeFunc(sqlFile)

		if err != nil {
			return err
		}
	}

	return nil
}

// WarmupCache warmups cache
func (s *Storage) WarmupCache(repo, arch string) error {
	switch {
	case repo == "":
		return fmt.Errorf("Can't warmup cache: %w", ErrEmptyRepoName)
	case arch == "":
		return fmt.Errorf("Can't warmup cache: %w", ErrEmptyArchName)
	case !s.IsInitialized():
		return fmt.Errorf("Can't warmup cache: %w", ErrNotInitialized)
	}

	for _, dbType := range data.DBList {
		_, err := s.GetDB(repo, arch, dbType)

		if err != nil {
			return err
		}
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Reindex generates index metadata for the given repository and arch
func (d *Depot) Reindex(full bool) error {
	if d == nil {
		return ErrNilDepot
	}

	return index.Generate(d.dataDir, d.indexOptions, full)
}

// AddPackage adds package to depot
func (d *Depot) AddPackage(rpmFile string) error {
	if rpmFile == "" {
		return fmt.Errorf("Can't add package to storage depot: %w", ErrEmptyPath)
	}

	if d == nil {
		return fmt.Errorf("Can't add package to storage depot: %w", ErrNilDepot)
	}

	err := fsutil.ValidatePerms("FRS", rpmFile)

	if err != nil {
		return fmt.Errorf("Can't add package to storage depot: %w", err)
	}

	if !rpm.IsRPM(rpmFile) {
		return fmt.Errorf("Can't add file to storage depot: %s is not an RPM package", rpmFile)
	}

	packageDir := d.dataDir

	if d.dataOptions.SplitFiles {
		packageDir, err = d.makePackageDir(rpmFile)

		if err != nil {
			return fmt.Errorf("Can't add package to storage depot: %w", err)
		}
	}

	err = d.copyFile(rpmFile, packageDir)

	if err != nil {
		return fmt.Errorf("Can't copy package to storage depot: %w", err)
	}

	return nil
}

// RemovePackage removes package from depot
func (d *Depot) RemovePackage(rpmFile string) error {
	if rpmFile == "" {
		return fmt.Errorf("Can't remove package from storage depot: %w", ErrEmptyPath)
	}

	if d == nil {
		return fmt.Errorf("Can't remove package from storage depot: %w", ErrNilDepot)
	}

	filePath := joinPath(d.dataDir, rpmFile)
	err := fsutil.ValidatePerms("FW", filePath)

	if err != nil {
		return fmt.Errorf("Can't remove package from storage depot: %w", err)
	}

	err = removeFunc(filePath)

	if err != nil {
		return fmt.Errorf("Can't remove package from storage depot: %w", err)
	}

	if d.dataOptions.SplitFiles {
		err = d.removePackageDir(rpmFile)

		if err != nil {
			return fmt.Errorf("Can't remove package from storage depot: %w", err)
		}
	}

	return nil
}

// GetPackagePath returns full path to package RPM file
func (d *Depot) GetPackagePath(rpmFileRelPath string) string {
	if d == nil {
		return ""
	}

	rpmFileName := path.Base(rpmFileRelPath)
	return joinPath(d.getPackageDir(rpmFileName), rpmFileName)
}

// HasPackage checks if depot contains file with given name
func (d *Depot) HasPackage(rpmFileName string) bool {
	if d == nil {
		return false
	}

	filePath := joinPath(d.getPackageDir(rpmFileName), rpmFileName)
	return fsutil.IsExist(filePath)
}

// IsEmpty returns true if repository is empty (no packages)
func (d *Depot) IsEmpty() bool {
	if d == nil {
		return true
	}

	return fsutil.IsEmptyDir(d.dataDir)
}

// IsCacheValid checks if database files was updated and depot contains outdated
// metadata info and keeps connection to outdated SQL databases
func (d *Depot) IsCacheValid() bool {
	if d == nil {
		return false
	}

	return d.CheckCache() == nil
}

// CheckCache checks if cache is valid and healthy
func (d *Depot) CheckCache() error {
	if d == nil || d.meta == nil {
		return ErrNilDepot
	}

	metaFile := d.GetMetaIndexPath()
	mTime, err := fsutil.GetMTime(metaFile)

	if err != nil {
		return fmt.Errorf("Can't check meta modification date")
	}

	if mTime.Unix() > d.meta.Revision {
		return fmt.Errorf(
			"Meta modification date is newer than generation date (%d > %d)",
			mTime.Unix(), d.meta.Revision,
		)
	}

	for dbType := range d.dbs {
		dbInfo := d.meta.Get(dbType + "_db")

		if dbInfo == nil {
			continue
		}

		dbFile := joinPath(d.dataDir, dbInfo.Location.HREF)
		dbMTime, err := fsutil.GetMTime(path.Clean(dbFile))

		if err != nil {
			return fmt.Errorf("Can't check database file modification date")
		}

		if dbMTime.Unix() != dbInfo.Timestamp {
			return fmt.Errorf(
				"Database file is newer than cached (%d > %d)",
				dbMTime.Unix(), dbInfo.Timestamp,
			)
		}
	}

	return nil
}

// InvalidateCache invalidates repository cache
func (d *Depot) InvalidateCache() error {
	if d == nil {
		return ErrNilDepot
	}

	d.meta = nil

	for dbName, db := range d.dbs {
		if db != nil && db.Ping() == nil {
			err := db.Close()

			if err != nil {
				return err
			}
		}

		delete(d.dbs, dbName)
	}

	return nil
}

// IsDBCached returns true if SQLite DB is cached
func (d *Depot) IsDBCached(dbType string) bool {
	if d == nil {
		return false
	}

	dbFile := d.GetDBFilePath(dbType)

	if !fsutil.IsExist(dbFile) {
		return false
	}

	dbInfo := d.meta.Get(dbType + "_db")

	if dbInfo == nil {
		return false
	}

	dbMTime, err := fsutil.GetMTime(dbFile)

	if err != nil {
		return false
	}

	if dbInfo.Timestamp > dbMTime.Unix() {
		return false
	}

	return true
}

// CacheDB caches (saves unpacked DB file) SQLite DB
func (d *Depot) CacheDB(dbType string) error {
	if dbType == "" {
		return fmt.Errorf("Can't cache DB: DB type can't be empty")
	}

	if d == nil {
		return fmt.Errorf("Can't cache DB: %w", ErrNilDepot)
	}

	dbInfo := d.meta.Get(dbType + "_db")

	if dbInfo == nil {
		return fmt.Errorf("Can't cache DB: Can't find info about DB %q", dbType)
	}

	dbFile := joinPath(d.dataDir, dbInfo.Location.HREF)

	if !fsutil.IsExist(dbFile) {
		return fmt.Errorf("Can't cache DB: Can't find file with SQLite database %q", dbType)
	}

	cachedDB := d.GetDBFilePath(dbType)
	err := utils.UnpackDB(dbFile, cachedDB)

	if err != nil {
		return fmt.Errorf("Can't cache DB: %w", err)
	}

	return nil
}

// OpenDB opens SQLite DB
func (d *Depot) OpenDB(dbType string) error {
	if d == nil {
		return ErrNilDepot
	}

	dbFile := d.GetDBFilePath(dbType)

	if !fsutil.IsExist(dbFile) {
		return fmt.Errorf("Can't find file %s", dbFile)
	}

	if d.dbs[dbType] != nil {
		d.dbs[dbType].Close()
	}

	var db *sql.DB

	// Use custom driver if required
	if hasCustomDriver[dbType] {
		db, _ = sql.Open("sqlite3_"+dbType, dbFile)
	} else {
		db, _ = sql.Open("sqlite3", dbFile)
	}

	d.dbs[dbType] = db

	return nil
}

// GetDB returns connection to SQLite DB
func (d *Depot) GetDB(dbType string) (*sql.DB, error) {
	if d == nil {
		return nil, ErrNilDepot
	}

	var err error

	if !d.IsCacheValid() {
		err := d.InvalidateCache()

		if err != nil {
			return nil, fmt.Errorf("Can't invalidate cache: %w", err)
		}
	}

	if d.meta == nil {
		d.meta, err = d.GetMetaIndex()

		if err != nil {
			return nil, fmt.Errorf("Can't read meta index: %w", err)
		}
	}

	if !d.IsDBCached(dbType) {
		err = d.CacheDB(dbType)

		if err != nil {
			return nil, fmt.Errorf("Can't cache DB: %w", err)
		}
	}

	if d.dbs[dbType] == nil {
		err = d.OpenDB(dbType)

		if err != nil {
			return nil, fmt.Errorf("Can't open DB: %w", err)
		}
	}

	return d.dbs[dbType], nil
}

// GetMetaIndex reads repository metadata
func (d *Depot) GetMetaIndex() (*meta.Index, error) {
	if d == nil {
		return nil, ErrNilDepot
	}

	metaFile := d.GetMetaIndexPath()

	if !fsutil.CheckPerms("FRS", metaFile) {
		return nil, fmt.Errorf("%s must be path to readable, non-empty XML file", metaFile)
	}

	return meta.Read(metaFile)
}

// GetMetaIndexPath returns path to metadata index file (repomd.xml)
func (d *Depot) GetMetaIndexPath() string {
	if d == nil {
		return ""
	}

	return joinPath(d.dataDir, "/repodata/repomd.xml")
}

// GetDBFilePath returns path to SQLite DB file
func (d *Depot) GetDBFilePath(dbType string) string {
	if d == nil {
		return ""
	}

	return joinPath(d.dataOptions.CacheDir, fmt.Sprintf("%s-%s.sqlite", d.id, dbType))
}

// ////////////////////////////////////////////////////////////////////////////////// //

// copyFile copies package into package directory and change permissions for it
func (d *Depot) copyFile(rpmFile, packageDir string) error {
	if d == nil {
		return fmt.Errorf("Can't change package attributes: %w", ErrNilDepot)
	}

	err := fsutil.CopyFile(rpmFile, packageDir, 0600)

	if err != nil {
		return err
	}

	targetFile := joinPath(packageDir, path.Base(rpmFile))
	err = updateObjectAttrs(targetFile, d.dataOptions, false)

	if err != nil {
		return fmt.Errorf("Can't change package attributes: %w", err)
	}

	return nil
}

// makePackageDir creates directory if required and returns path to directory for packages
// if split-files option is enabled
func (d *Depot) makePackageDir(rpmFile string) (string, error) {
	if d == nil {
		return "", fmt.Errorf("Can't create directory for package: %w", ErrNilDepot)
	}

	rpmFileName := path.Base(rpmFile)
	dirName := strutil.Head(rpmFileName, 1)

	if !DirNameValidatorRegex.MatchString(dirName) {
		return "", fmt.Errorf("Can't create directory for package: Can't use name %q for directory", dirName)
	}

	packageDir := joinPath(d.dataDir, dirName)

	if fsutil.IsExist(packageDir) {
		return packageDir, nil
	}

	err := mkdirFunc(packageDir, 0700)

	if err != nil {
		return "", err
	}

	err = updateObjectAttrs(packageDir, d.dataOptions, true)

	if err != nil {
		return "", fmt.Errorf("Can't change package directory attributes: %w", err)
	}

	return packageDir, err
}

// removePackageDir removes package
func (d *Depot) removePackageDir(rpmFile string) error {
	if d == nil {
		return ErrNilDepot
	}

	rpmFileDir := path.Dir(rpmFile)

	if rpmFileDir == "." || rpmFileDir == "/" {
		return nil
	}

	rpmFileDirFull := joinPath(d.dataDir, rpmFileDir)

	if !fsutil.IsEmptyDir(rpmFileDirFull) {
		return nil
	}

	return removeFunc(rpmFileDirFull)
}

// getPackageDir returns full path to directory for given rpm file
func (d *Depot) getPackageDir(rpmFileName string) string {
	if d == nil {
		return ""
	}

	if d.dataOptions.SplitFiles {
		dirName := strutil.Head(rpmFileName, 1)
		return joinPath(d.dataDir, dirName)
	}

	return d.dataDir
}

// ////////////////////////////////////////////////////////////////////////////////// //

// updateObjectAttrs update object (directory or file) attributes
func updateObjectAttrs(path string, options *Options, isDir bool) error {
	var perms os.FileMode

	uid, gid := -1, -1

	if options.User != "" {
		newUser, err := system.LookupUser(options.User)

		if err != nil {
			return fmt.Errorf("Can't get UID for user %q", options.User)
		} else {
			uid = newUser.UID
		}
	}

	if options.Group != "" {
		newGroup, err := system.LookupGroup(options.Group)

		if err != nil {
			return fmt.Errorf("Can't get GID for group %q", options.Group)
		} else {
			gid = newGroup.GID
		}
	}

	if uid != -1 || gid != -1 {
		err := chownFunc(path, uid, gid)

		if err != nil {
			return err
		}
	}

	if isDir {
		perms = options.GetDirPerms()
	} else {
		perms = options.GetFilePerms()
	}

	return chmodFunc(path, perms)
}

// checkDataDir checks repository directory permissions
func checkDataDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("Path to repository directory can't be empty")
	}

	return nil
}

// checkCacheDir checks cache directory permissions
func checkCacheDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("Path to cache directory can't be empty")
	}

	return fsutil.ValidatePerms("DRWX", dir)
}

// joinPath joins path elements into one string
func joinPath(objs ...string) string {
	return path.Clean(path.Join(objs...))
}
