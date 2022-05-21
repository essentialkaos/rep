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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/sortutil"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/version"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/rpm"
	"github.com/essentialkaos/rep/repo/search"
	"github.com/essentialkaos/rep/repo/sign"
	"github.com/essentialkaos/rep/repo/storage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	_SQL_LIST_ALL       = `SELECT name,arch,version,release,epoch,rpm_sourcerpm,location_href FROM packages;`
	_SQL_LIST_LATEST    = `SELECT name,arch,version,release,epoch,rpm_sourcerpm,location_href FROM packages GROUP BY name HAVING MAX(pkgKey);`
	_SQL_LIST_BY_NAME   = `SELECT name,arch,version,release,epoch,rpm_sourcerpm,location_href FROM packages WHERE (name || "-" || version || "-" || release) LIKE ? ORDER BY rpm_sourcerpm;`
	_SQL_FIND_BY_KEYS   = `SELECT name,arch,version,release,epoch,rpm_sourcerpm,location_href FROM packages WHERE pkgKey in (%s);`
	_SQL_EXIST          = `SELECT time_file FROM packages WHERE name = ? AND version = ? AND release = ? AND epoch = ?;`
	_SQL_CONTAINS_ONE   = `SELECT pkgKey FROM filelist WHERE length(filetypes) = 1 AND (dirname || "/" || filenames) GLOB ?;`
	_SQL_CONTAINS_MANY  = `SELECT pkgKey FROM filelist WHERE length(filetypes) > 1 AND filelist_globber(?, dirname, filenames);`
	_SQL_STATS          = `SELECT SUM(size_package),COUNT(*) FROM packages;`
	_SQL_INFO_BASE      = `SELECT pkgId,name,arch,version,release,epoch,rpm_sourcerpm,location_href,summary,description,url,time_file,time_build,rpm_license,rpm_vendor,rpm_group,size_package,size_installed FROM packages WHERE (name || "-" || version || "-" || release) LIKE ? GROUP BY name HAVING MAX(pkgKey) LIMIT 1;`
	_SQL_INFO_FILES     = `SELECT f.dirname,f.filenames,f.filetypes FROM filelist f INNER JOIN packages p ON f.pkgKey = p.pkgKey WHERE p.pkgId = ? ORDER BY f.dirname,f.filenames;`
	_SQL_INFO_REQUIRES  = `SELECT r.name,r.flags,r.epoch,r.version,r.release FROM requires r INNER JOIN packages p ON r.pkgKey = p.pkgKey WHERE p.pkgId = ? ORDER BY r.name;`
	_SQL_INFO_PROVIDES  = `SELECT r.name,r.flags,r.epoch,r.version,r.release FROM provides r INNER JOIN packages p ON r.pkgKey = p.pkgKey WHERE p.pkgId = ? ORDER BY r.name;`
	_SQL_INFO_CHANGELOG = `SELECT c.author,c.changelog FROM changelog c INNER JOIN packages p ON c.pkgKey = p.pkgKey WHERE p.pkgId = ? ORDER BY 1 DESC LIMIT 1;`
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Repository is main repository struct
type Repository struct {
	Name        string
	DefaultArch string
	FileFilter  string
	Replace     bool

	SigningKey *sign.Key

	Testing *SubRepository // Testing sub-repository (with unstable packages)
	Release *SubRepository // Release sub-repository (with stable packages)

	storage storage.Storage
}

// SubRepository is sub-repository struct
type SubRepository struct {
	Name   string      // Sub-repository name
	Parent *Repository // Pointer to parent repository
}

// RepositoryStats contains repository stats data
type RepositoryStats struct {
	Packages      map[string]int
	Sizes         map[string]int64
	TotalPackages int
	TotalSize     int64
	Updated       time.Time
}

// Package contains info about package
type Package struct {
	Name      string        // Name
	Version   string        // Version
	Release   string        // Release
	Epoch     string        // Epoch
	ArchFlags data.ArchFlag // Archs flag
	Src       string        // Source package name
	Files     []PackageFile // RPM files list

	Info *PackageInfo // Additional info
}

// PackageInfo contains additional information about package
type PackageInfo struct {
	Summary       string            // Summary
	Desc          string            // Description
	URL           string            // URL
	Vendor        string            // Vendor
	Packager      string            // Packager
	Group         string            // Group
	License       string            // License
	SizePackage   uint64            // Size of package in bytes
	SizeInstalled uint64            // Size of installed data in bytes
	DateAdded     time.Time         // Add date as unix timestamp
	DateBuild     time.Time         // Build date as unix timestamp
	Changelog     *PackageChangelog // Changelog records
	Requires      []data.Dependency // Requires
	Provides      []data.Dependency // Provides
	Files         PayloadData       // Files and directories
}

// PayloadData is a slice with info about files or directories
type PayloadData []PayloadObject

// PackageChangelog contains changelog data
type PackageChangelog struct {
	Author  string
	Records []string
}

// PackageFile contains info about package file
type PackageFile struct {
	Arch string
	Path string
}

// PayloadObject contains info about file or directory
type PayloadObject struct {
	IsDir bool
	Path  string
}

// PackageBundle is slice of packages built from one source RPM
type PackageBundle []*Package

// PackageStack is slice with package bundles
type PackageStack []PackageBundle

// ////////////////////////////////////////////////////////////////////////////////// //

// packageStackBuilder contains packages info for data grouping
type packageStackBuilder struct {
	Index map[string]int // map [source name] → [bundle index]
	Data  PackageStack   // package bundles
}

// ////////////////////////////////////////////////////////////////////////////////// //

var (
	ErrEmptyPath      = fmt.Errorf("Path to file is empty")
	ErrEmptyRepo      = fmt.Errorf("Repository is empty")
	ErrNilPackage     = fmt.Errorf("Package is nil")
	ErrNilStorage     = fmt.Errorf("Storage is nil")
	ErrNotInitialized = fmt.Errorf("Repository is not initialized")
)

// ////////////////////////////////////////////////////////////////////////////////// //

// repoNameValidationRegex is regex pattern for repository name validation
var repoNameValidationRegex = regexp.MustCompile(`[0-9a-zA-Z_\-]+`)

// ////////////////////////////////////////////////////////////////////////////////// //

// NewRepository creates new struct for repostitory
func NewRepository(name string, repoStorage storage.Storage) (*Repository, error) {
	if !repoNameValidationRegex.MatchString(name) {
		return nil, fmt.Errorf("Name %q is invalid", name)
	}

	if repoStorage == nil {
		return nil, ErrNilStorage
	}

	repo := &Repository{Name: name}

	repo.Release = NewSubRepository(data.REPO_RELEASE)
	repo.Testing = NewSubRepository(data.REPO_TESTING)

	repo.Testing.Parent = repo
	repo.Release.Parent = repo

	repo.DefaultArch = data.ARCH_X64

	repo.storage = repoStorage

	return repo, nil
}

// NewSubRepository creates struct for sub-repository (release/testing)
func NewSubRepository(name string) *SubRepository {
	return &SubRepository{Name: name}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// FullName returns full package name
func (p *Package) FullName() string {
	return p.Name + "-" + p.Version + "-" + p.Release
}

// HasArch returns true if package have file for given arch
func (p *Package) HasArch(arch string) bool {
	archFlag := data.SupportedArchs[arch].Flag

	if archFlag == 0 {
		return false
	}

	return p.ArchFlags.Has(archFlag)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// HasMultiBundles returns true if stack contains bundle with more than 1 package
func (s PackageStack) HasMultiBundles() bool {
	for _, bundle := range s {
		if len(bundle) > 1 {
			return true
		}
	}

	return false
}

// GetArchsFlag returns flag for all packages in all bundles in stack
func (s PackageStack) GetArchsFlag() data.ArchFlag {
	var flag data.ArchFlag

	for _, bundle := range s {
		for _, pkg := range bundle {
			flag |= pkg.ArchFlags
		}
	}

	return flag
}

// GetArchs returns slice with arch names presented in stack
func (s PackageStack) GetArchs() []string {
	var result []string

	flag := s.GetArchsFlag()

	for _, arch := range data.ArchList {
		if flag.Has(data.SupportedArchs[arch].Flag) {
			result = append(result, arch)
		}
	}

	return result
}

// FlattenFiles returns slice with all packages files in stack
func (s PackageStack) FlattenFiles() []PackageFile {
	var result []PackageFile

	for _, bundle := range s {
		for _, pkg := range bundle {
			for _, file := range pkg.Files {
				result = append(result, file)
			}
		}
	}

	return result
}

// IsEmpty returns true if package stack is empty
func (s PackageStack) IsEmpty() bool {
	return len(s) == 0
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Initialize initializes the new repository and creates all required directories
func (r *Repository) Initialize(archList []string) error {
	if r.storage.IsInitialized() {
		return fmt.Errorf("Repository already initialized")
	}

	return r.storage.Initialize(
		[]string{data.REPO_RELEASE, data.REPO_TESTING},
		archList,
	)
}

// Info returns package struct with extended info and release date
func (r *Repository) Info(name, arch string) (*Package, time.Time, error) {
	if !r.storage.IsInitialized() {
		return nil, time.Time{}, ErrNotInitialized
	}

	arch = strutil.Q(arch, r.DefaultArch)

	if !r.HasArch(arch) {
		return nil, time.Time{}, fmt.Errorf("Unknown or unsupported architecture %q", arch)
	}

	if r.Release.IsEmpty(arch) && r.Testing.IsEmpty(arch) {
		return nil, time.Time{}, ErrEmptyRepo
	}

	pkg, err := r.Testing.getPackageInfo(name, arch)

	if err != nil {
		return nil, time.Time{}, err
	}

	if pkg == nil {
		return nil, time.Time{}, fmt.Errorf("Can't find package %q", name)
	}

	_, releaseDate, err := r.IsPackageReleased(pkg)

	if err != nil {
		return nil, time.Time{}, err
	}

	return pkg, releaseDate, nil
}

// CopyPackage copies packages between sub-repositories
func (r *Repository) CopyPackage(source, target *SubRepository, rpmFileRelPath string) error {
	if !r.storage.IsInitialized() {
		return ErrNotInitialized
	}

	switch {
	case source == nil:
		return fmt.Errorf("Source sub-repository is nil")
	case target == nil:
		return fmt.Errorf("Target sub-repository is nil")
	case rpmFileRelPath == "":
		return ErrEmptyPath
	}

	return r.storage.CopyPackage(source.Name, target.Name, rpmFileRelPath)
}

// IsPackageReleased checks if package was released
func (r *Repository) IsPackageReleased(pkg *Package) (bool, time.Time, error) {
	if !r.storage.IsInitialized() {
		return false, time.Time{}, ErrNotInitialized
	}

	var releaseDate time.Time

	if pkg == nil {
		return false, releaseDate, ErrNilPackage
	}

	for _, arch := range data.ArchList {
		if !r.Release.HasArch(arch) || r.Release.IsEmpty(arch) {
			continue
		}

		if !pkg.HasArch(arch) {
			continue
		}

		exist, timeAdd, err := r.Release.hasPackage(pkg, arch)

		if err != nil {
			return false, releaseDate, err
		}

		if !exist {
			return false, releaseDate, nil
		}

		if timeAdd.Unix() > releaseDate.Unix() {
			releaseDate = timeAdd
		}

		return true, releaseDate, nil
	}

	return false, releaseDate, nil
}

// ReadSigningKey securely reads signing key from file
func (r *Repository) ReadSigningKey(file string) error {
	var err error

	r.SigningKey, err = sign.ReadKey(file)

	return err
}

// IsSigningRequired returns true if signing key is set and repository
// requires package signing
func (r *Repository) IsSigningRequired() bool {
	return r.SigningKey != nil
}

// HasArch returns true if release and testing repositories have given arch
func (r *Repository) HasArch(arch string) bool {
	return r.Testing.HasArch(arch) && r.Release.HasArch(arch)
}

// PurgeCache removes all cached data
func (r *Repository) PurgeCache() error {
	err := r.storage.PurgeCache()

	if err != nil {
		return err
	}

	return r.storage.InvalidateCache()
}

// ////////////////////////////////////////////////////////////////////////////////// //

// AddPackage copies given file into sub-repository storage
// Important: This method DO NOT run repository reindex
func (r *SubRepository) AddPackage(rpmFilePath string) error {
	switch {
	case rpmFilePath == "":
		return fmt.Errorf("Can't add package to repository: %w", ErrEmptyPath)
	case !r.Parent.storage.IsInitialized():
		return fmt.Errorf("Can't add package to repository: %w", ErrNotInitialized)
	}

	err := fsutil.ValidatePerms("FRS", rpmFilePath)

	if err != nil {
		return fmt.Errorf("Can't add package to repository: %w", err)
	}

	if !rpm.IsRPM(rpmFilePath) {
		return fmt.Errorf("Can't add file to repository: %s is not an RPM package", rpmFilePath)
	}

	if r.Parent.SigningKey != nil {
		privateKey, err := r.Parent.SigningKey.Get(nil)

		if err != nil {
			return fmt.Errorf("Can't add file to repository: %w", err)
		}

		isSigned, err := sign.IsSigned(rpmFilePath, privateKey)

		if err != nil {
			return fmt.Errorf("Can't add file to repository: %w", err)
		}

		if !isSigned {
			return fmt.Errorf("Can't add file to repository: Repository allows only singed packages")
		}
	}

	return r.Parent.storage.AddPackage(r.Name, rpmFilePath)
}

// RemovePackage removes package with given relative path from sub-repository storage
// Important: This method DO NOT run repository reindex
func (r *SubRepository) RemovePackage(rpmFileRelPath string) error {
	switch {
	case rpmFileRelPath == "":
		return fmt.Errorf("Can't remove package from repository: %w", ErrEmptyPath)
	case !r.Parent.storage.IsInitialized():
		return fmt.Errorf("Can't remove package from repository: %w", ErrNotInitialized)
	}

	return r.Parent.storage.RemovePackage(r.Name, rpmFileRelPath)
}

// HasPackageFile returns true if sub-repository contains file with given name
func (r *SubRepository) HasPackageFile(rpmFileName string) bool {
	switch {
	case rpmFileName == "":
		return false
	case !r.Parent.storage.IsInitialized():
		return false
	}

	return r.Parent.storage.HasPackage(r.Name, rpmFileName)
}

// Stats returns stats for sub-repository
func (r *SubRepository) Stats() (*RepositoryStats, error) {
	if !r.Parent.storage.IsInitialized() {
		return nil, ErrNotInitialized
	}

	stats := &RepositoryStats{
		Packages: make(map[string]int),
		Sizes:    make(map[string]int64),
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" || r.IsEmpty(arch) {
			continue
		}

		count, size, err := r.getRepoStats(arch)

		if err != nil {
			return nil, err
		}

		stats.TotalPackages += count
		stats.TotalSize += size

		stats.Packages[arch] = count
		stats.Sizes[arch] = size

		modTime := r.Parent.storage.GetModTime(r.Name, data.SupportedArchs[arch].Dir)

		if !modTime.IsZero() && modTime.Unix() > stats.Updated.Unix() {
			stats.Updated = modTime
		}
	}

	return stats, nil
}

// List returns list with packages
func (r *SubRepository) List(filter string, all bool) (PackageStack, error) {
	if !r.Parent.storage.IsInitialized() {
		return nil, ErrNotInitialized
	}

	var err error
	var psb *packageStackBuilder

	switch {
	case all && filter == "":
		psb, err = r.listPackages(_SQL_LIST_ALL)
	case !all && filter == "":
		psb, err = r.listPackages(_SQL_LIST_LATEST)
	default:
		psb, err = r.listPackages(_SQL_LIST_BY_NAME, "%"+sanitizeInput(filter)+"%")
	}

	if err != nil {
		return nil, err
	}

	return psb.Data, nil
}

// Find tries to find packages by given search query
func (r *SubRepository) Find(query search.Query) (PackageStack, error) {
	if !r.Parent.storage.IsInitialized() {
		return nil, ErrNotInitialized
	}

	if len(query) == 0 {
		return PackageStack{}, nil
	}

	errs := query.Validate()

	if len(errs) != 0 {
		return nil, errs[0]
	}

	psb, err := r.searchPackages(query)

	if err != nil {
		return nil, err
	}

	if psb == nil {
		return PackageStack{}, nil
	}

	return psb.Data, nil
}

// Reindex generates repository metadata
func (r *SubRepository) Reindex(full bool) error {
	if !r.Parent.storage.IsInitialized() {
		return ErrNotInitialized
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" {
			continue
		}

		err := r.Parent.storage.Reindex(r.Name, arch, full)

		if err != nil {
			return err
		}
	}

	return nil
}

// GetFullPackagePath returns full path to package
func (r *SubRepository) GetFullPackagePath(pkg PackageFile) string {
	return r.Parent.storage.GetPackagePath(r.Name, pkg.Arch, pkg.Path)
}

// HasArch returns true if sub-repository contains packages with given arch
func (r *SubRepository) HasArch(arch string) bool {
	return r.Parent.storage.HasArch(r.Name, arch)
}

// returns true if repository is empty (no packages)
func (r *SubRepository) IsEmpty(arch string) bool {
	return r.Parent.storage.IsEmpty(r.Name, arch)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getRepoStats reads stats info from repository DB
func (r *SubRepository) getRepoStats(arch string) (int, int64, error) {
	var count, size sql.NullInt64

	infoRows, err := r.execQuery(data.DB_PRIMARY, arch, _SQL_STATS)

	if err != nil {
		return 0, 0, fmt.Errorf("Can't collect repository stats (arch: %s): %w", arch, err)
	}

	defer infoRows.Close()

	if !infoRows.Next() {
		return 0, 0, fmt.Errorf("Can't find repository stats (arch: %s)", arch)
	}

	err = infoRows.Scan(&size, &count)

	if err != nil {
		return 0, 0, fmt.Errorf("Error while scaning rows with repository stats (arch: %s): %w", arch, err)
	}

	return int(count.Int64), size.Int64, nil
}

// listPackages returns basic packages info
func (r *SubRepository) listPackages(query string, args ...interface{}) (*packageStackBuilder, error) {
	psb := &packageStackBuilder{
		Index: make(map[string]int),
		Data:  make([]PackageBundle, 0),
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" || r.IsEmpty(arch) {
			continue
		}

		err := r.listArchPackages(psb, arch, query, args...)

		if err != nil {
			return nil, err
		}
	}

	if len(psb.Data) == 0 {
		return psb, nil
	}

	sortPackageStack(psb)

	return psb, nil
}

// listArchPackages appends basic packages info for given arch to stack
func (r *SubRepository) listArchPackages(psb *packageStackBuilder, arch string, query string, args ...interface{}) error {
	rows, err := r.execQuery(data.DB_PRIMARY, arch, query, args...)

	if err != nil {
		return fmt.Errorf("Can't collect arch packages list (%s): %w", arch, err)
	}

	defer rows.Close()

	var sourceRPM string
	var pkgName, pkgArch, pkgVer, pkgRel, pkgEpc, pkgSrc, pkgHREF sql.NullString

ROWSLOOP:
	for rows.Next() {
		err = rows.Scan(&pkgName, &pkgArch, &pkgVer, &pkgRel, &pkgEpc, &pkgSrc, &pkgHREF)

		if err != nil {
			return fmt.Errorf("Error while scaning rows with info about arch packages list (%s): %w", arch, err)
		}

		if pkgArch.String == data.ARCH_SRC {
			sourceRPM = fmt.Sprintf("%s-%s-%s.src.rpm", pkgName.String, pkgVer.String, pkgRel.String)
		} else {
			sourceRPM = pkgSrc.String
		}

		index, hasBundle := psb.Index[sourceRPM]

		if !hasBundle {
			index = len(psb.Data)
			psb.Index[sourceRPM] = index
			psb.Data = append(psb.Data, PackageBundle{})
		} else {
			for _, pkg := range psb.Data[index] {
				if pkg.Name == pkgName.String &&
					pkg.Version == pkgVer.String &&
					pkg.Release == pkgRel.String &&
					pkg.Epoch == pkgEpc.String {
					pkg.ArchFlags |= data.SupportedArchs[pkgArch.String].Flag
					pkg.Files = append(pkg.Files, PackageFile{arch, pkgHREF.String})
					continue ROWSLOOP
				}
			}
		}

		psb.Data[index] = append(
			psb.Data[index],
			&Package{
				Name:      pkgName.String,
				Version:   pkgVer.String,
				Release:   pkgRel.String,
				Epoch:     pkgEpc.String,
				ArchFlags: data.SupportedArchs[pkgArch.String].Flag,
				Src:       sourceRPM,
				Files:     []PackageFile{PackageFile{arch, pkgHREF.String}},
			},
		)
	}

	return nil
}

// searchPackages searches package with given search query
func (r *SubRepository) searchPackages(query search.Query) (*packageStackBuilder, error) {
	index := data.NewPkgKeyIndex()

	for _, term := range query.Terms() {
		for _, arch := range data.ArchList {
			if !r.HasArch(arch) || index.IgnoreArch(arch) ||
				data.SupportedArchs[arch].Dir == "" || r.IsEmpty(arch) {
				continue
			}

			targetDB, sqlQueries := term.SQL()
			keyMap, err := r.searchArchPackages(arch, targetDB, sqlQueries)

			if err != nil {
				return nil, err
			}

			if len(keyMap) == 0 {
				index.Drop(arch)
			} else {
				index.Intersect(arch, keyMap)
			}
		}

		if !index.HasData() {
			return nil, nil
		}
	}

	psb := &packageStackBuilder{
		Index: make(map[string]int),
		Data:  make([]PackageBundle, 0),
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" || !index.HasArch(arch) {
			continue
		}

		err := r.listArchPackages(psb, arch, fmt.Sprintf(_SQL_FIND_BY_KEYS, index.List(arch)))

		if err != nil {
			return nil, err
		}
	}

	sortPackageStack(psb)

	return psb, nil
}

// searchPackages searches package with given search query for some arch
func (r *SubRepository) searchArchPackages(arch, targetDB string, queries []string) (data.PkgKeyMap, error) {
	keyMap := data.NewPkgKeyMap()

	for _, query := range queries {
		rows, err := r.execQuery(targetDB, arch, query)

		if err != nil {
			return nil, fmt.Errorf("Can't collect info about packages architecture (%s): %w", arch, err)
		}

		var pkgKey int

		for rows.Next() {
			err = rows.Scan(&pkgKey)

			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("Error while scaning rows with info about packages architecture (%s): %w", arch, err)
			}

			keyMap.Set(pkgKey)
		}

		rows.Close()
	}

	return keyMap, nil
}

// hasPackage checks if package presented in repository
func (r *SubRepository) hasPackage(pkg *Package, arch string) (bool, time.Time, error) {
	rows, err := r.execQuery(
		data.DB_PRIMARY, arch, _SQL_EXIST,
		pkg.Name, pkg.Version, pkg.Release, pkg.Epoch,
	)

	if err != nil {
		return false, time.Time{}, fmt.Errorf("Can't collect info about package: %w", err)
	}

	defer rows.Close()

	var pTimeFile sql.NullInt64

	if !rows.Next() {
		return false, time.Time{}, nil
	}

	err = rows.Scan(&pTimeFile)

	if err != nil {
		return false, time.Time{}, fmt.Errorf("Error while scaning rows with info about package: %w", err)
	}

	return true, time.Unix(pTimeFile.Int64, 0), nil
}

// getPackageInfo collects detailed info about package with given name
func (r *SubRepository) getPackageInfo(name, arch string) (*Package, error) {
	var err error

	pkg, pkgID, err := r.collectPackageBasicInfo(name, arch)

	if err != nil {
		return nil, err
	}

	if pkgID == "" {
		return nil, nil
	}

	pkg.Info.Changelog, err = r.collectPackageChangelogInfo(pkgID, arch)

	if err != nil {
		return nil, err
	}

	pkg.Info.Files, err = r.collectPackageFilesInfo(pkgID, arch)

	if err != nil {
		return nil, err
	}

	pkg.Info.Requires, err = r.collectPackageDepInfo(pkgID, arch, _SQL_INFO_REQUIRES)

	if err != nil {
		return nil, err
	}

	pkg.Info.Provides, err = r.collectPackageDepInfo(pkgID, arch, _SQL_INFO_PROVIDES)

	if err != nil {
		return nil, err
	}

	return pkg, nil
}

// collectPackageBasicInfo collects basic package info
func (r *SubRepository) collectPackageBasicInfo(name, arch string) (*Package, string, error) {
	name = sanitizeInput(name) + "%"

	rows, err := r.execQuery(data.DB_PRIMARY, arch, _SQL_INFO_BASE, name)

	if err != nil {
		return nil, "", fmt.Errorf("Can't collect basic package info: %w", err)
	}

	defer rows.Close()

	var pkgID, pkgName, pkgArch, pkgVer, pkgRel, pkgEpc, pkgSrc, pkgHREF sql.NullString
	var pkgAddTS, pkgBuildTS, pkgSize, pkgSizeInst sql.NullInt64
	var pkgSum, pkgDesc, pkgURL, pkgLic, pkgVend, pkgGroup sql.NullString

	if !rows.Next() {
		return nil, "", nil
	}

	err = rows.Scan(
		&pkgID, &pkgName, &pkgArch, &pkgVer, &pkgRel, &pkgEpc, &pkgSrc, &pkgHREF,
		&pkgSum, &pkgDesc, &pkgURL, &pkgAddTS, &pkgBuildTS,
		&pkgLic, &pkgVend, &pkgGroup, &pkgSize, &pkgSizeInst,
	)

	if err != nil {
		return nil, "", fmt.Errorf("Error while scaning rows with basic package info: %w", err)
	}

	pkg := &Package{
		Name:      pkgName.String,
		Version:   pkgVer.String,
		Release:   pkgRel.String,
		Epoch:     pkgEpc.String,
		ArchFlags: data.SupportedArchs[pkgArch.String].Flag,
		Src:       pkgSrc.String,
		Files:     []PackageFile{PackageFile{arch, pkgHREF.String}},
		Info: &PackageInfo{
			Summary:       pkgSum.String,
			Desc:          pkgDesc.String,
			URL:           pkgURL.String,
			License:       pkgLic.String,
			Vendor:        pkgVend.String,
			Group:         pkgGroup.String,
			SizePackage:   uint64(pkgSize.Int64),
			SizeInstalled: uint64(pkgSizeInst.Int64),
			DateAdded:     time.Unix(pkgAddTS.Int64, 0),
			DateBuild:     time.Unix(pkgBuildTS.Int64, 0),
		},
	}

	return pkg, pkgID.String, nil
}

// collectPackageChangelogInfo collects changelog records for package
func (r *SubRepository) collectPackageChangelogInfo(pkgID, arch string) (*PackageChangelog, error) {
	rows, err := r.execQuery(data.DB_OTHER, arch, _SQL_INFO_CHANGELOG, pkgID)

	if err != nil {
		return nil, fmt.Errorf("Can't execute query for collecting changelog records: %w", err)
	}

	defer rows.Close()

	var clAuthor, clRecords sql.NullString

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Scan(&clAuthor, &clRecords)

	if err != nil {
		return nil, fmt.Errorf("Error while scaning rows with changelog records: %w", err)
	}

	return &PackageChangelog{
		Author:  clAuthor.String,
		Records: strings.Split(clRecords.String, "\n"),
	}, nil
}

// collectPackageFilesInfo collects info about package payload
func (r *SubRepository) collectPackageFilesInfo(pkgID, arch string) ([]PayloadObject, error) {
	rows, err := r.execQuery(data.DB_FILELISTS, arch, _SQL_INFO_FILES, pkgID)

	if err != nil {
		return nil, fmt.Errorf("Can't execute query for collecting payload info: %w", err)
	}

	defer rows.Close()

	var result []PayloadObject
	var fDir, fObjs, fTypes sql.NullString

	for rows.Next() {
		err = rows.Scan(&fDir, &fObjs, &fTypes)

		if err != nil {
			return nil, fmt.Errorf("Error while scaning rows with payload info: %w", err)
		}

		result = append(result, parsePayloadList(
			fDir.String, fObjs.String, fTypes.String,
		)...)
	}

	return result, nil
}

// collectPackageDepInfo collects requires/provides info
func (r *SubRepository) collectPackageDepInfo(pkgID, arch, query string) ([]data.Dependency, error) {
	rows, err := r.execQuery(data.DB_PRIMARY, arch, query, pkgID)

	if err != nil {
		return nil, fmt.Errorf("Can't execute query for collecting requires/provides data: %w", err)
	}

	defer rows.Close()

	var result []data.Dependency
	var pkgName, pkgFlag, pkgEpc, pkgVer, pkgRel sql.NullString

	for rows.Next() {
		err = rows.Scan(&pkgName, &pkgFlag, &pkgEpc, &pkgVer, &pkgRel)

		if err != nil {
			return nil, fmt.Errorf("Error while scaning rows with requires/provides data: %w", err)
		}

		dep := data.Dependency{
			Name:    pkgName.String,
			Epoch:   pkgEpc.String,
			Version: pkgVer.String,
			Release: pkgRel.String,
			Flag:    data.ParseComp(pkgFlag.String),
		}

		// Replace unversioned dep deplicates
		if len(result) > 0 && result[len(result)-1].Name == pkgName.String {
			prevRec := result[len(result)-1]

			if pkgFlag.String != "" && prevRec.Flag == data.COMP_FLAG_ANY {
				result[len(result)-1] = dep
			}

			continue
		}

		result = append(result, dep)
	}

	return result, nil
}

// execQuery execs SQL query over DB
func (r *SubRepository) execQuery(dbType, arch, query string, args ...interface{}) (*sql.Rows, error) {
	arch = r.guessArch(arch)
	archDir := data.SupportedArchs[arch].Dir

	if archDir == "" {
		return nil, fmt.Errorf("Unknown or unsupported arch %q", arch)
	}

	db, err := r.Parent.storage.GetDB(r.Name, arch, dbType)

	if err != nil {
		return nil, fmt.Errorf("Can't get DB from storage: %w", err)
	}

	return db.Query(query, args...)
}

// guessArch tries to guess real package arch
func (r *SubRepository) guessArch(arch string) string {
	if arch != data.ARCH_NOARCH {
		return arch
	}

	for _, a := range data.ArchList {
		if a == data.ARCH_SRC || a == data.ARCH_NOARCH {
			continue
		}

		if r.HasArch(a) {
			return a
		}
	}

	return arch
}

// ////////////////////////////////////////////////////////////////////////////////// //

// sanitizeInput sanitizes user input
func sanitizeInput(data string) string {
	if data == "" {
		return ""
	}

	result := data

	result = strutil.ReplaceAll(result, "?()", "_")
	result = strutil.ReplaceAll(result, "'\"", "")
	result = strutil.ReplaceAll(result, "^$<>|,+[]{}#;!=", " ")

	return result
}

// sortPackageStack sort packages stack data
func sortPackageStack(psb *packageStackBuilder) {
	if len(psb.Data) <= 1 {
		return
	}

	sort.Sort(psb.Data)
}

// parsePayloadList parses package payload data
func parsePayloadList(dir, objs, types string) []PayloadObject {
	var result []PayloadObject

	for i := 0; i < len(types); i++ {
		obj := strutil.ReadField(objs, i, false, "/")

		switch types[i] {
		case 'd':
			result = append(result, PayloadObject{true, dir + "/" + obj})
		default:
			result = append(result, PayloadObject{false, dir + "/" + obj})
		}
	}

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

func (p PackageStack) Len() int {
	return len(p)
}

func (p PackageStack) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PackageStack) Less(i, j int) bool {
	if p[i][0].Name != p[j][0].Name {
		return sortutil.NaturalLess(p[i][0].Name, p[j][0].Name)
	}

	if p[i][0].Version != p[j][0].Version {
		v1, _ := version.Parse(p[i][0].Version)
		v2, _ := version.Parse(p[j][0].Version)

		return v1.Less(v2)
	}

	return sortutil.NaturalLess(p[i][0].Release, p[j][0].Release)
}

func (p PayloadData) Len() int {
	return len(p)
}

func (p PayloadData) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PayloadData) Less(i, j int) bool {
	iDir := path.Dir(p[i].Path)
	jDir := path.Dir(p[j].Path)

	if iDir == jDir {
		switch {
		case p[i].IsDir && !p[j].IsDir:
			return false
		case !p[i].IsDir && p[j].IsDir:
			return true
		}
	}

	return p[i].Path < p[j].Path
}

// ////////////////////////////////////////////////////////////////////////////////// //
