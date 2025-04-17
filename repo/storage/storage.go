package storage

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"database/sql"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	TYPE_FS = "fs"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Storage is interface for repository storage
type Storage interface {
	// BASIC METHODS --

	// Initialize initializes the new repository and creates all required directories
	Initialize(repoList, archList []string) error

	// AddPackage adds package file to the given repository
	// Important: This method DO NOT run repository reindex
	AddPackage(repo, rpmFilePath string) error

	// RemovePackage removes package with given relative path from the given repository
	// Important: This method DO NOT run repository reindex
	RemovePackage(repo, arch, rpmFileRelPath string) error

	// CopyPackage copies file from one repository to another
	// Important: This method DO NOT run repository reindex
	CopyPackage(fromRepo, toRepo, arch, rpmFileRelPath string) error

	// IsInitialized returns true if the repository already initialized and ready for work
	IsInitialized() bool

	// IsEmpty returns true if repository is empty (no packages)
	IsEmpty(repo, arch string) bool

	// HasRepo returns true if given repository exists
	HasRepo(repo string) bool

	// HasArch returns true if repository storage contains directory for specific arch
	HasArch(repo, arch string) bool

	// HasPackage checks if repository contains file with given name
	HasPackage(repo, arch, rpmFileName string) bool

	// GetPackagePath returns path to package file
	GetPackagePath(repo, arch, pkg string) string

	// METADATA & DB --

	// Reindex generates index metadata for the given repository and arch
	Reindex(repo, arch string, full bool) error

	// GetDB returns connection to SQLite DB
	GetDB(repo, arch, dbType string) (*sql.DB, error)

	// GetModTime returns date of repository index modification
	GetModTime(repo, arch string) (time.Time, error)

	// InvalidateCache invalidates cache
	InvalidateCache() error

	// IsCacheValid returns true if cache is valid
	IsCacheValid(repo, arch string) bool

	// PurgeCache deletes all SQLite files from cache directory
	PurgeCache() error

	// WarmupCache warmups cache
	WarmupCache(repo, arch string) error
}

// ////////////////////////////////////////////////////////////////////////////////// //
