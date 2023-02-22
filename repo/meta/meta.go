package meta

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"os"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	TYPE_PRIMARY       = "primary"
	TYPE_FILELISTS     = "filelists"
	TYPE_OTHER         = "other"
	TYPE_PRIMARY_DB    = "primary_db"
	TYPE_FILELISTS_DB  = "filelists_db"
	TYPE_OTHER_DB      = "other_db"
	TYPE_PRIMARY_ZCK   = "primary_zck"
	TYPE_FILELISTS_ZCK = "filelists_zck"
	TYPE_OTHER_ZCK     = "other_zck"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Index contains info about all metadata files
type Index struct {
	Revision int64       `xml:"revision"`
	Data     []*Metadata `xml:"data"`
}

// Metadata contains info about metadata
type Metadata struct {
	Type            string   `xml:"type,attr"`
	Checksum        Checksum `xml:"checksum"`
	OpenChecksum    Checksum `xml:"open-checksum"`
	Location        Location `xml:"location"`
	DatabaseVersion int      `xml:"database_version"`
	Timestamp       int64    `xml:"timestamp"`
	Size            int64    `xml:"size"`
	OpenSize        int64    `xml:"open-size"`
	HeaderSize      int64    `xml:"header-size"`
}

// Location contains info about data location
type Location struct {
	HREF string `xml:"href,attr"`
}

// Checksum contains info about checksum
type Checksum struct {
	Type string `xml:"type,attr"`
	Hash string `xml:",chardata"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

var copyFunc = io.Copy

// ////////////////////////////////////////////////////////////////////////////////// //

// Read reads metadata file
func Read(file string) (*Index, error) {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	r := bufio.NewReader(fd)
	d := xml.NewDecoder(r)

	result := &Index{}
	err = d.Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Validate validates all repository metadata
func (m *Index) Validate(dir string) error {
	if dir == "" {
		return fmt.Errorf("Path to repository directory cannot be empty")
	}

	for _, meta := range m.Data {
		err := meta.Validate(dir)

		if err != nil {
			return err
		}
	}

	return nil
}

// Get returns metadata struct with given type
func (m *Index) Get(dbType string) *Metadata {
	for _, m := range m.Data {
		if m.Type == dbType {
			return m
		}
	}

	return nil
}

// Validate validates metadata file
func (m *Metadata) Validate(dir string) error {
	if dir == "" {
		return fmt.Errorf("Path to repository directory can't be empty")
	}

	err := ValidateChecksum(
		dir+"/"+m.Location.HREF,
		m.Checksum.Type,
		m.Checksum.Hash,
	)

	if err != nil {
		return fmt.Errorf("Error while checksum validation for %s: %v", m.Type, err)
	}

	return nil
}

// String returns string representation of location
func (l Location) String() string {
	return l.HREF
}

// ////////////////////////////////////////////////////////////////////////////////// //

// ValidateChecksum validates file checksum
func ValidateChecksum(file, checksumType, checksumHash string) error {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer fd.Close()

	var hasher hash.Hash

	switch checksumType {
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	case "sha224":
		hasher = sha256.New224()
	case "sha256":
		hasher = sha256.New()
	case "sha384":
		hasher = sha512.New384()
	case "sha512":
		hasher = sha512.New()
	default:
		return fmt.Errorf("Unsupported checksum type %s", checksumType)
	}

	_, err = copyFunc(hasher, fd)

	if err != nil {
		return err
	}

	fileChecksum := fmt.Sprintf("%x", hasher.Sum(nil))

	if checksumHash != fileChecksum {
		return fmt.Errorf("Checksum validation failed (%s ≠ %s)", checksumHash, fileChecksum)
	}

	return nil
}
