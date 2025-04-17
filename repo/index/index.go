package index

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/essentialkaos/ek/v13/fsutil"
	"github.com/essentialkaos/ek/v13/sliceutil"
	"github.com/essentialkaos/ek/v13/system"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	CHECKSUM_MD5    = "md5"
	CHECKSUM_SHA1   = "sha1"
	CHECKSUM_SHA224 = "sha224"
	CHECKSUM_SHA256 = "sha256"
	CHECKSUM_SHA384 = "sha384"
	CHECKSUM_SHA512 = "sha512"
)

const (
	MDF_SIMPLE = "simple"
	MDF_UNIQUE = "unique"
)

const (
	COMPRESSION_GZ   = "gz"
	COMPRESSION_BZ2  = "bz2"
	COMPRESSION_XZ   = "xz"
	COMPRESSION_ZSTD = "zstd"
)

const (
	PERMS_DIR  os.FileMode = 0755 // Default permissions for directories
	PERMS_FILE os.FileMode = 0644 // Default permissions for files
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options contains options used for generating repository index
type Options struct {
	User      string      // Repository data directory owner username
	Group     string      // Repository data directory owner group
	DirPerms  os.FileMode // Permissions for directories
	FilePerms os.FileMode // Permissions for files

	GroupFile      string // Path to groupfile to include in metadata
	CheckSum       string // Checksum used in repomd.xml and for packages in the metadata (default: sha256)
	MDFilenames    string // Include the file's checksum in the filename,helps with proxies (unique/simple)
	CompressType   string // Which compression type to use (default: bz2)
	Distro         string // Distro tag and optional CPE ID
	Content        string // Tags for the content in the repository
	Revision       string // User-specified revision for repository
	NumDeltas      int    // The number of older versions to make deltas against
	ChangelogLimit int    // Only import the last N changelog entries
	Workers        int    // Number of workers to spawn to read rpms
	Pretty         bool   // Make sure all xml generated is formatted
	Update         bool   // Use the existing repodata to speed up creation of new repository
	Split          bool   // Generate split meta
	SkipSymlinks   bool   // Ignore symlinks of packages
	Deltas         bool   // Create delta rpms and metadata
	Zchunk         bool   // Generate zchunk files as well as the standard repodata
}

// ////////////////////////////////////////////////////////////////////////////////// //

// CheckSumMethods contains all supported checksum methods
var CheckSumMethods = []string{
	CHECKSUM_MD5,
	CHECKSUM_SHA1,
	CHECKSUM_SHA224,
	CHECKSUM_SHA256,
	CHECKSUM_SHA384,
	CHECKSUM_SHA512,
}

// MDFilenames contains all supported types of supported metadata names generation
// methods
var MDFilenames = []string{
	MDF_SIMPLE,
	MDF_UNIQUE,
}

// Contains all supported compression methods
var CompressionMethods = []string{
	COMPRESSION_GZ,
	COMPRESSION_BZ2,
	COMPRESSION_XZ,
	COMPRESSION_ZSTD,
}

// DefaultOptions is default options
var DefaultOptions = &Options{
	Update:       true,
	MDFilenames:  MDF_SIMPLE,
	CompressType: COMPRESSION_BZ2,
}

// ////////////////////////////////////////////////////////////////////////////////// //

var chownFunc = os.Chown
var chmodFunc = os.Chmod

// ////////////////////////////////////////////////////////////////////////////////// //

// IsCreaterepoInstalled returns true if createrepo_c utility is installed on the
// system
func IsCreaterepoInstalled() bool {
	_, err := exec.LookPath("createrepo_c")
	return err == nil
}

// Generate creates repository index using createrepo_c utility
func Generate(path string, options *Options, full bool) error {
	if !IsCreaterepoInstalled() {
		return fmt.Errorf("Can't generate index: createrepo_c not installed")
	}

	err := options.Validate()

	if err != nil {
		return fmt.Errorf("Error while options validation: %w", err)
	}

	if full && options.Update {
		options = options.Clone()
		options.Update = false
	}

	var stdErrBuf bytes.Buffer

	cmd := exec.Command("createrepo_c", options.ToArgs()...)
	cmd.Args = append(cmd.Args, path)
	cmd.Stderr = &stdErrBuf

	if cmd.Run() != nil {
		errorMessage := strings.TrimRight(stdErrBuf.String(), "\r\n")
		return fmt.Errorf("Error while executing createrepo_c: %s", errorMessage)
	}

	if options.User != "" || options.Group != "" {
		err = updateIndexOwner(path, options)

		if err != nil {
			return err
		}
	}

	if options.DirPerms != 0 || options.FilePerms != 0 {
		return updateIndexPerms(path, options)
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Clone creates options copy
func (o *Options) Clone() *Options {
	return &Options{
		User:      o.User,
		Group:     o.Group,
		DirPerms:  o.DirPerms,
		FilePerms: o.FilePerms,

		GroupFile:      o.GroupFile,
		CheckSum:       o.CheckSum,
		MDFilenames:    o.MDFilenames,
		CompressType:   o.CompressType,
		Distro:         o.Distro,
		Content:        o.Content,
		Revision:       o.Revision,
		NumDeltas:      o.NumDeltas,
		ChangelogLimit: o.ChangelogLimit,
		Workers:        o.Workers,
		Pretty:         o.Pretty,
		Update:         o.Update,
		Split:          o.Split,
		SkipSymlinks:   o.SkipSymlinks,
		Deltas:         o.Deltas,
		Zchunk:         o.Zchunk,
	}
}

// Validate validates options
func (o *Options) Validate() error {
	var err error

	if o.GroupFile != "" {
		err = fsutil.ValidatePerms("FRS", o.GroupFile)

		if err != nil {
			return fmt.Errorf("Can't use given group file: %w", err)
		}
	}

	if o.User != "" && !system.IsUserExist(o.User) {
		return fmt.Errorf("User \"%s\" is not present on the system", o.User)
	}

	if o.Group != "" && !system.IsUserExist(o.Group) {
		return fmt.Errorf("Group \"%s\" is not present on the system", o.Group)
	}

	if o.NumDeltas < 0 {
		return fmt.Errorf("NumDeltas can't be less than 0")
	}

	if o.Workers < 0 {
		return fmt.Errorf("Workers can't be less than 0")
	}

	if o.ChangelogLimit < 0 {
		return fmt.Errorf("ChangelogLimit can't be less than 0")
	}

	if o.CheckSum != "" && !sliceutil.Contains(CheckSumMethods, o.CheckSum) {
		return fmt.Errorf("Unsupported CheckSum method \"%s\"", o.CheckSum)
	}

	if o.MDFilenames != "" && !sliceutil.Contains(MDFilenames, o.MDFilenames) {
		return fmt.Errorf("Unsupported MDFilenames method \"%s\"", o.MDFilenames)
	}

	if o.CompressType != "" && !sliceutil.Contains(CompressionMethods, o.CompressType) {
		return fmt.Errorf("Unsupported compression method \"%s\"", o.CompressType)
	}

	return nil
}

// ToArgs converts options to slice with arguments for createrepo_c utility
func (o *Options) ToArgs() []string {
	var args []string

	args = append(args, "--database")

	if o.GroupFile != "" {
		args = append(args, "--groupfile="+o.GroupFile)
	}

	if o.CheckSum != "" {
		args = append(args, "--checksum="+o.CheckSum)
	}

	if o.Pretty {
		args = append(args, "--pretty")
	}

	if o.Update {
		args = append(args, "--update")
	}

	if o.Split {
		args = append(args, "--split")
	}

	if o.SkipSymlinks {
		args = append(args, "--skip-symlinks")
	}

	if o.Deltas {
		args = append(args, "--deltas")
	}

	if o.ChangelogLimit > 0 {
		args = append(args, "--changelog-limit="+strconv.Itoa(o.ChangelogLimit))
	}

	if o.Distro != "" {
		args = append(args, "--distro="+o.Distro)
	}

	if o.Content != "" {
		args = append(args, "--content="+o.Content)
	}

	if o.Revision != "" {
		args = append(args, "--revision="+o.Revision)
	}

	if o.NumDeltas > 0 {
		args = append(args, "--num-deltas="+strconv.Itoa(o.NumDeltas))
	}

	if o.Workers > 1 {
		args = append(args, "--workers="+strconv.Itoa(o.Workers))
	}

	if o.CompressType != "" {
		args = append(args,
			"--compress-type="+o.CompressType,
			"--general-compress-type="+o.CompressType,
		)
	} else {
		args = append(args,
			"--compress-type="+COMPRESSION_BZ2,
			"--general-compress-type="+COMPRESSION_BZ2,
		)
	}

	if o.Zchunk {
		args = append(args, "--zck")
	}

	if o.MDFilenames != "" {
		args = append(args, "--"+o.MDFilenames+"-md-filenames")
	} else {
		args = append(args, "--"+MDF_SIMPLE+"-md-filenames")
	}

	return args
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

// updateIndexOwner updates owner for repodata directory and files in it
func updateIndexOwner(path string, options *Options) error {
	repodataPath := path + "/repodata"
	uid, gid := -1, -1

	if options.User != "" {
		newUser, err := system.LookupUser(options.User)

		if err != nil {
			return fmt.Errorf("Can't get UID for user \"%s\"", options.User)
		} else {
			uid = newUser.UID
		}
	}

	if options.Group != "" {
		newGroup, err := system.LookupGroup(options.Group)

		if err != nil {
			return fmt.Errorf("Can't get GID for group \"%s\"", options.Group)
		} else {
			gid = newGroup.GID
		}
	}

	objects := fsutil.List(repodataPath, true)
	fsutil.ListToAbsolute(repodataPath, objects)
	objects = append(objects, repodataPath)

	for _, obj := range objects {
		err := chownFunc(obj, uid, gid)

		if err != nil {
			return err
		}
	}

	return nil
}

// updateIndexPerms updates permissions for repodata directory and files in it
func updateIndexPerms(path string, options *Options) error {
	repodataPath := path + "/repodata"

	if options.DirPerms != 0 {
		err := chmodFunc(repodataPath, options.DirPerms)

		if err != nil {
			return err
		}
	}

	if options.FilePerms != 0 {
		files := fsutil.List(repodataPath, true)
		fsutil.ListToAbsolute(repodataPath, files)

		for _, file := range files {
			err := chmodFunc(file, options.FilePerms)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
