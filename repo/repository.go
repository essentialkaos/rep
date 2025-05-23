package repo

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
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

	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/path"
	"github.com/essentialkaos/ek/v13/sortutil"
	"github.com/essentialkaos/ek/v13/strutil"
	"github.com/essentialkaos/ek/v13/version"

	"github.com/essentialkaos/rep/v3/repo/data"
	"github.com/essentialkaos/rep/v3/repo/helpers"
	"github.com/essentialkaos/rep/v3/repo/rpm"
	"github.com/essentialkaos/rep/v3/repo/search"
	"github.com/essentialkaos/rep/v3/repo/sign"
	"github.com/essentialkaos/rep/v3/repo/storage"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	_SQL_LIST_ALL       = `SELECT pkgId,name,arch,version,release,epoch,rpm_sourcerpm,location_href,size_package FROM packages;`
	_SQL_LIST_LATEST    = `SELECT pkgId,name,arch,version,release,epoch,rpm_sourcerpm,location_href,size_package FROM packages GROUP BY name HAVING MAX(pkgKey);`
	_SQL_LIST_BY_NAME   = `SELECT pkgId,name,arch,version,release,epoch,rpm_sourcerpm,location_href,size_package FROM packages WHERE (name || "-" || version || "-" || release) LIKE @filter ORDER BY rpm_sourcerpm;`
	_SQL_FIND_BY_KEYS   = `SELECT pkgId,name,arch,version,release,epoch,rpm_sourcerpm,location_href,size_package FROM packages WHERE pkgKey in (%s);`
	_SQL_EXIST          = `SELECT time_file FROM packages WHERE name = @name AND version = @version AND release = @release AND epoch = @epoch;`
	_SQL_STATS          = `SELECT SUM(size_package),COUNT(*) FROM packages;`
	_SQL_INFO_BASE      = `SELECT pkgId,name,arch,version,release,epoch,rpm_sourcerpm,location_href,summary,description,url,time_file,time_build,rpm_license,rpm_vendor,rpm_group,size_package,size_installed FROM packages WHERE (name || "-" || version || "-" || release) LIKE @name GROUP BY name HAVING MAX(time_build) LIMIT 1;`
	_SQL_INFO_FILES     = `SELECT f.dirname,f.filenames,f.filetypes FROM filelist f INNER JOIN packages p ON f.pkgKey = p.pkgKey WHERE p.pkgId = @id ORDER BY f.dirname,f.filenames;`
	_SQL_INFO_REQUIRES  = `SELECT r.name,r.flags,r.epoch,r.version,r.release FROM requires r INNER JOIN packages p ON r.pkgKey = p.pkgKey WHERE p.pkgId = @id ORDER BY r.name;`
	_SQL_INFO_PROVIDES  = `SELECT r.name,r.flags,r.epoch,r.version,r.release FROM provides r INNER JOIN packages p ON r.pkgKey = p.pkgKey WHERE p.pkgId = @id ORDER BY r.name;`
	_SQL_INFO_CHANGELOG = `SELECT c.author,c.date,c.changelog FROM changelog c INNER JOIN packages p ON c.pkgKey = p.pkgKey WHERE p.pkgId = @id AND c.author LIKE @version ORDER BY c.date DESC LIMIT 1;`
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Repository is main repository struct
type Repository struct {
	Name        string
	DefaultArch string
	FileFilter  string
	Replace     bool

	SigningKey *sign.ArmoredKey

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
	Files     PackageFiles  // RPM files list

	Info *PackageInfo // Additional info
}

// PackageFiles is slice with package files
type PackageFiles []PackageFile

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
	Payload       PackagePayload    // Files and directories
}

// PackagePayload is a slice with info about package files or directories
type PackagePayload []PayloadObject

// PackageChangelog contains changelog data
type PackageChangelog struct {
	Records []string
	Author  string
	Date    time.Time
}

// PackageFile contains info about package file
type PackageFile struct {
	CRC          string        // File checksum (first 7 symbols)
	Path         string        // Path to file
	Size         uint64        // Package size in bytes
	ArchFlag     data.ArchFlag // Package arch flag
	BaseArchFlag data.ArchFlag // Sub-repo (i.e. directory arch) flag
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
	if p == nil {
		return ""
	}

	return p.Name + "-" + p.Version + "-" + p.Release
}

// HasArch returns true if package have file for given arch
func (p *Package) HasArch(arch string) bool {
	if p == nil {
		return false
	}

	archFlag := data.SupportedArchs[arch].Flag

	if archFlag == data.ARCH_FLAG_UNKNOWN {
		return false
	}

	return p.ArchFlags.Has(archFlag)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// HasArch returns true if slice contains given arch
func (p PackageFiles) HasArch(arch string) bool {
	for _, pf := range p {
		if arch != "" && pf.ArchFlag.Has(data.SupportedArchs[arch].Flag) {
			return true
		}
	}

	return false
}

// Size calculates total size of all packages in slice
func (p PackageFiles) Size() uint64 {
	var size uint64

	for _, pkg := range p {
		size += pkg.Size
	}

	return size
}

// ////////////////////////////////////////////////////////////////////////////////// //

// PackageBundle returns size of package bundle
func (b PackageBundle) Size() int {
	size := 0

	for _, pkg := range b {
		if pkg != nil {
			size++
		}
	}

	return size
}

// ////////////////////////////////////////////////////////////////////////////////// //

// HasMultiBundles returns true if stack contains bundle with more than 1 package
func (s PackageStack) HasMultiBundles() bool {
	if s.IsEmpty() {
		return false
	}

	for _, bundle := range s {
		if bundle.Size() > 1 {
			return true
		}
	}

	return false
}

// GetArchsFlag returns flag for all packages in all bundles in stack
func (s PackageStack) GetArchsFlag() data.ArchFlag {
	if s.IsEmpty() {
		return data.ARCH_FLAG_UNKNOWN
	}

	var flag data.ArchFlag

	for _, bundle := range s {
		for _, pkg := range bundle {
			if pkg != nil {
				flag |= pkg.ArchFlags
			}
		}
	}

	return flag
}

// GetArchs returns slice with arch names presented in stack
func (s PackageStack) GetArchs() []string {
	if s.IsEmpty() {
		return nil
	}

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
func (s PackageStack) FlattenFiles() PackageFiles {
	if s.IsEmpty() {
		return nil
	}

	var result PackageFiles

	for _, bundle := range s {
		for _, pkg := range bundle {
			if pkg != nil {
				result = append(result, pkg.Files...)
			}
		}
	}

	return result
}

// IsEmpty returns true if package stack is empty
func (s PackageStack) IsEmpty() bool {
	for _, bundle := range s {
		for _, pkg := range bundle {
			if pkg != nil {
				return false
			}
		}
	}

	return true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Initialize initializes the new repository and creates all required directories
func (r *Repository) Initialize(archList []string) error {
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
func (r *Repository) CopyPackage(source, target *SubRepository, packageFile PackageFile) error {
	if !r.storage.IsInitialized() {
		return ErrNotInitialized
	}

	switch {
	case source == nil:
		return fmt.Errorf("Source sub-repository is nil")
	case target == nil:
		return fmt.Errorf("Target sub-repository is nil")
	case packageFile.Path == "":
		return ErrEmptyPath
	}

	return r.storage.CopyPackage(
		source.Name, target.Name,
		packageFile.BaseArchFlag.String(), packageFile.Path,
	)
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
		switch {
		case arch != data.ARCH_SRC && pkg.HasArch(data.ARCH_NOARCH):
			// Package is noarch package, do the check
		case arch == data.ARCH_NOARCH, // Skip if it pseudo arch
			!r.Release.HasArch(arch), // Skip if the release repo doesn't contain this arch
			r.Release.IsEmpty(arch),  // Skip if there are no packages with this arch
			!pkg.HasArch(arch):       // Skip if the package doesn't contain this arch
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
		key, err := r.Parent.SigningKey.Read(nil)

		if err != nil {
			return fmt.Errorf("Can't add file to repository: %w", err)
		}

		isSignValid, err := sign.IsPackageSignatureValid(rpmFilePath, key)

		if err != nil {
			return fmt.Errorf("Can't add file to repository: %w", err)
		}

		if !isSignValid {
			return fmt.Errorf("Can't add file to repository: Repository allows only signed packages")
		}
	}

	return r.Parent.storage.AddPackage(r.Name, rpmFilePath)
}

// RemovePackage removes package with given relative path from sub-repository storage
// Important: This method DO NOT run repository reindex
func (r *SubRepository) RemovePackage(packageFile PackageFile) error {
	switch {
	case packageFile.Path == "":
		return fmt.Errorf("Can't remove package from repository: %w", ErrEmptyPath)
	case !r.Parent.storage.IsInitialized():
		return fmt.Errorf("Can't remove package from repository: %w", ErrNotInitialized)
	}

	return r.Parent.storage.RemovePackage(
		r.Name, packageFile.BaseArchFlag.String(),
		packageFile.Path,
	)
}

// HasPackageFile returns true if sub-repository contains file with given name
func (r *SubRepository) HasPackageFile(rpmFileName string) bool {
	arch := helpers.GuessFileArch(rpmFileName)

	switch {
	case rpmFileName == "":
		return false
	case arch == "":
		return false
	case !r.Parent.storage.IsInitialized():
		return false
	}

	if arch != data.ARCH_NOARCH {
		return r.Parent.storage.HasPackage(r.Name, arch, rpmFileName)
	}

	for _, arch = range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" || r.IsEmpty(arch) {
			continue
		}

		if r.Parent.storage.HasPackage(r.Name, arch, rpmFileName) {
			return true
		}
	}

	return false
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

		modTime, err := r.Parent.storage.GetModTime(r.Name, arch)

		if err != nil {
			return nil, err
		}

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
		psb, err = r.listPackages(
			_SQL_LIST_BY_NAME,
			sql.Named("filter", "%"+sanitizeInput(filter)+"%"),
		)
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
func (r *SubRepository) Reindex(full bool, ch chan string) error {
	if !r.Parent.storage.IsInitialized() {
		return ErrNotInitialized
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" {
			continue
		}

		if ch != nil {
			ch <- arch
		}

		err := r.Parent.storage.Reindex(r.Name, arch, full)

		if err != nil {
			return err
		}
	}

	if ch != nil {
		close(ch)
	}

	return nil
}

// IsCacheValid returns true if cache for architectures is valid
func (r *SubRepository) IsCacheValid() bool {
	if !r.Parent.storage.IsInitialized() {
		return false
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" {
			continue
		}

		if !r.Parent.storage.IsCacheValid(r.Name, arch) {
			return false
		}
	}

	return true
}

// WarmupCache warmups cache for all architectures
func (r *SubRepository) WarmupCache() error {
	if !r.Parent.storage.IsInitialized() {
		return ErrNotInitialized
	}

	for _, arch := range data.ArchList {
		if !r.HasArch(arch) || data.SupportedArchs[arch].Dir == "" {
			continue
		}

		err := r.Parent.storage.WarmupCache(r.Name, arch)

		if err != nil {
			return fmt.Errorf("Can't warmup %s cache: %w", r.Name, err)
		}
	}

	return nil
}

// GetFullPackagePath returns full path to package
func (r *SubRepository) GetFullPackagePath(pkg PackageFile) string {
	return r.Parent.storage.GetPackagePath(r.Name, pkg.BaseArchFlag.String(), pkg.Path)
}

// HasArch returns true if sub-repository contains packages with given arch
func (r *SubRepository) HasArch(arch string) bool {
	return r.Parent.storage.HasArch(r.Name, arch)
}

// IsEmpty returns true if sub-repository is empty (no packages)
func (r *SubRepository) IsEmpty(arch string) bool {
	return r.Parent.storage.IsEmpty(r.Name, arch)
}

// Is is shortcut for checking sub-repository name
func (r *SubRepository) Is(name string) bool {
	return r.Name == name
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
		return 0, 0, fmt.Errorf("Error while scanning rows with repository stats (arch: %s): %w", arch, err)
	}

	return int(count.Int64), size.Int64, nil
}

// listPackages returns basic packages info
func (r *SubRepository) listPackages(query string, args ...sql.NamedArg) (*packageStackBuilder, error) {
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
func (r *SubRepository) listArchPackages(psb *packageStackBuilder, arch string, query string, args ...sql.NamedArg) error {
	rows, err := r.execQuery(data.DB_PRIMARY, arch, query, args...)

	if err != nil {
		return fmt.Errorf("Can't collect arch packages list (%s): %w", arch, err)
	}

	defer rows.Close()

	var sourceRPM string
	var pkgID, pkgName, pkgArch, pkgVer, pkgRel, pkgEpc, pkgSrc, pkgHREF sql.NullString
	var pkgSize sql.NullInt64

ROWSLOOP:
	for rows.Next() {
		err = rows.Scan(&pkgID, &pkgName, &pkgArch, &pkgVer, &pkgRel, &pkgEpc, &pkgSrc, &pkgHREF, &pkgSize)

		if err != nil {
			return fmt.Errorf("Error while scanning rows with info about arch packages list (%s): %w", arch, err)
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
					pkg.Files = append(pkg.Files, PackageFile{
						CRC:          strutil.Head(pkgID.String, 7),
						Path:         pkgHREF.String,
						Size:         uint64(pkgSize.Int64),
						ArchFlag:     data.SupportedArchs[pkgArch.String].Flag,
						BaseArchFlag: data.SupportedArchs[arch].Flag,
					})
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
				Files: PackageFiles{PackageFile{
					CRC:          strutil.Head(pkgID.String, 7),
					Path:         pkgHREF.String,
					Size:         uint64(pkgSize.Int64),
					ArchFlag:     data.SupportedArchs[pkgArch.String].Flag,
					BaseArchFlag: data.SupportedArchs[arch].Flag,
				}},
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
				return nil, fmt.Errorf("Error while scanning rows with info about packages architecture (%s): %w", arch, err)
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
		sql.Named("name", pkg.Name),
		sql.Named("version", pkg.Version),
		sql.Named("release", pkg.Release),
		sql.Named("epoch", pkg.Epoch),
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
		return false, time.Time{}, fmt.Errorf("Error while scanning rows with info about package: %w", err)
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

	pkg.Info.Payload, err = r.collectPackagePayloadInfo(pkgID, arch)

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

	err = r.appendPackageChangelogInfo(pkg, pkgID, arch)

	if err != nil {
		return nil, err
	}

	return pkg, nil
}

// collectPackageBasicInfo collects basic package info
func (r *SubRepository) collectPackageBasicInfo(name, arch string) (*Package, string, error) {
	name = sanitizeInput(name) + "%"

	rows, err := r.execQuery(
		data.DB_PRIMARY, arch, _SQL_INFO_BASE,
		sql.Named("name", name),
	)

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
		return nil, "", fmt.Errorf("Error while scanning rows with basic package info: %w", err)
	}

	pkg := &Package{
		Name:      pkgName.String,
		Version:   pkgVer.String,
		Release:   pkgRel.String,
		Epoch:     pkgEpc.String,
		ArchFlags: data.SupportedArchs[pkgArch.String].Flag,
		Src:       pkgSrc.String,
		Files: PackageFiles{PackageFile{
			CRC:          strutil.Head(pkgID.String, 7),
			Path:         pkgHREF.String,
			Size:         uint64(pkgSize.Int64),
			ArchFlag:     data.SupportedArchs[pkgArch.String].Flag,
			BaseArchFlag: data.SupportedArchs[arch].Flag,
		}},
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

// appendPackageChangelogInfo appends changelog records for package
func (r *SubRepository) appendPackageChangelogInfo(pkg *Package, pkgID, arch string) error {
	if pkg == nil {
		return fmt.Errorf("Can't append changelog records: package is nil")
	}

	version := fmt.Sprintf("%% - %s-%s", pkg.Version, formatReleaseVersion(pkg.Release))
	rows, err := r.execQuery(
		data.DB_OTHER, arch, _SQL_INFO_CHANGELOG,
		sql.Named("id", pkgID),
		sql.Named("version", version),
	)

	if err != nil {
		return fmt.Errorf("Can't execute query for collecting changelog records: %w", err)
	}

	defer rows.Close()

	var cAuthor, cRecords sql.NullString
	var cDate sql.NullInt64

	if !rows.Next() {
		return nil
	}

	err = rows.Scan(&cAuthor, &cDate, &cRecords)

	if err != nil {
		return fmt.Errorf("Error while scanning rows with changelog records: %w", err)
	}

	pkg.Info.Changelog = &PackageChangelog{
		Records: strings.Split(cRecords.String, "\n"),
		Author:  cAuthor.String,
		Date:    time.Unix(cDate.Int64, 0),
	}

	return nil
}

// collectPackagePayloadInfo collects info about package payload
func (r *SubRepository) collectPackagePayloadInfo(pkgID, arch string) ([]PayloadObject, error) {
	rows, err := r.execQuery(
		data.DB_FILELISTS, arch, _SQL_INFO_FILES,
		sql.Named("id", pkgID),
	)

	if err != nil {
		return nil, fmt.Errorf("Can't execute query for collecting payload info: %w", err)
	}

	defer rows.Close()

	var result []PayloadObject
	var fDir, fObjs, fTypes sql.NullString

	for rows.Next() {
		err = rows.Scan(&fDir, &fObjs, &fTypes)

		if err != nil {
			return nil, fmt.Errorf("Error while scanning rows with payload info: %w", err)
		}

		result = append(result, parsePayloadList(
			fDir.String, fObjs.String, fTypes.String,
		)...)
	}

	return result, nil
}

// collectPackageDepInfo collects requires/provides info
func (r *SubRepository) collectPackageDepInfo(pkgID, arch, query string) ([]data.Dependency, error) {
	rows, err := r.execQuery(
		data.DB_PRIMARY, arch, query,
		sql.Named("id", pkgID),
	)

	if err != nil {
		return nil, fmt.Errorf("Can't execute query for collecting requires/provides data: %w", err)
	}

	defer rows.Close()

	var result []data.Dependency
	var pkgName, pkgFlag, pkgEpc, pkgVer, pkgRel sql.NullString

	for rows.Next() {
		err = rows.Scan(&pkgName, &pkgFlag, &pkgEpc, &pkgVer, &pkgRel)

		if err != nil {
			return nil, fmt.Errorf("Error while scanning rows with requires/provides data: %w", err)
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
func (r *SubRepository) execQuery(dbType, arch, query string, args ...sql.NamedArg) (*sql.Rows, error) {
	arch = r.guessArch(arch)
	archDir := data.SupportedArchs[arch].Dir

	if archDir == "" {
		return nil, fmt.Errorf("Unknown or unsupported arch %q", arch)
	}

	db, err := r.Parent.storage.GetDB(r.Name, arch, dbType)

	if err != nil {
		return nil, fmt.Errorf("Can't get DB from storage: %w", err)
	}

	return db.Query(query, sqlArgToAny(args)...)
}

// guessArch tries to guess real package arch
func (r *SubRepository) guessArch(arch string) string {
	if arch != data.ARCH_NOARCH {
		return arch
	}

	for _, a := range data.BinArchList {
		if r.HasArch(a) {
			arch = a
			break
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
		obj := strutil.ReadField(objs, i, false, '/')

		switch types[i] {
		case 'd':
			result = append(result, PayloadObject{true, dir + "/" + obj})
		default:
			result = append(result, PayloadObject{false, dir + "/" + obj})
		}
	}

	return result
}

// formatReleaseVersion formats release version removing OS info
func formatReleaseVersion(r string) string {
	i := strings.LastIndex(r, ".")

	if i == -1 {
		return r
	}

	return r[:i]
}

// sqlArgToAny converts sql.NamedArg slice into any slice
func sqlArgToAny(s []sql.NamedArg) []any {
	result := make([]any, len(s))

	for i := 0; i < len(s); i++ {
		result[i] = s[i]
	}

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Len is the number of elements in the collection
func (p PackageStack) Len() int {
	return len(p)
}

// Swap swaps the elements with indexes i and j
func (p PackageStack) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less reports whether the element with index i
// must sort before the element with index j
func (p PackageStack) Less(i, j int) bool {
	if p[i][0].Name != p[j][0].Name {
		return sortutil.NaturalLess(p[i][0].Name, p[j][0].Name)
	}

	if p[i][0].Version != p[j][0].Version {
		// Use natural sort if version is not semver
		if strings.Trim(p[i][0].Version, ".0123456789") != p[i][0].Version ||
			strings.Trim(p[j][0].Version, ".0123456789") != p[j][0].Version {
			return sortutil.NaturalLess(p[i][0].Version, p[j][0].Version)
		}

		v1, _ := version.Parse(p[i][0].Version)
		v2, _ := version.Parse(p[j][0].Version)

		return v1.Less(v2)
	}

	return sortutil.NaturalLess(p[i][0].Release, p[j][0].Release)
}

// Len is the number of elements in the collection
func (p PackagePayload) Len() int {
	return len(p)
}

// Swap swaps the elements with indexes i and j
func (p PackagePayload) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less reports whether the element with index i
// must sort before the element with index j
func (p PackagePayload) Less(i, j int) bool {
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
