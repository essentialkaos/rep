package utils

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	_FORMAT_RAW  uint8 = 0
	_FORMAT_GZIP uint8 = 1
	_FORMAT_BZIP uint8 = 2
	_FORMAT_XZ   uint8 = 3
	_FORMAT_ZSTD uint8 = 4
)

// ////////////////////////////////////////////////////////////////////////////////// //

// sqliteMagicHeader is SQLite magic header
// https://www.sqlite.org/fileformat.html#magic_header_string
var sqliteMagicHeader = []byte("SQLite format 3\x00")

// ////////////////////////////////////////////////////////////////////////////////// //

// UnpackDB unpacks compressed SQLite DB
func UnpackDB(source, output string) error {
	switch {
	case strings.HasSuffix(source, ".gz"):
		return unpackDBData(source, output, _FORMAT_GZIP)

	case strings.HasSuffix(source, ".bz2"):
		return unpackDBData(source, output, _FORMAT_BZIP)

	case strings.HasSuffix(source, ".xz"):
		return unpackDBData(source, output, _FORMAT_XZ)

	case strings.HasSuffix(source, ".zst"):
		return unpackDBData(source, output, _FORMAT_ZSTD)

	case strings.HasSuffix(source, ".sqlite"):
		return unpackDBData(source, output, _FORMAT_RAW)

	default:
		return fmt.Errorf("Unsupported DB format")
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// unpackDBData uncompress file data and writes it to given file
func unpackDBData(source, output string, format uint8) error {
	sourceFd, err := os.OpenFile(source, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer sourceFd.Close()

	r := getFormatReader(format, sourceFd)

	err = checkMagicHeader(r)

	if err != nil {
		return err
	}

	outputFd, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	defer outputFd.Close()

	outputFd.Write(sqliteMagicHeader)

	_, err = io.Copy(outputFd, r)

	return err
}

// checkMagicHeader checks SQLite magic header
func checkMagicHeader(sourceFd io.Reader) error {
	buf := make([]byte, len(sqliteMagicHeader))
	_, err := io.ReadAtLeast(sourceFd, buf, len(sqliteMagicHeader))

	if err != nil {
		return err
	}

	if !bytes.Equal(buf, sqliteMagicHeader) {
		return fmt.Errorf("Given file doesn't contain SQLite magic header (not an SQLite DB?)")
	}

	return nil
}

// getFormatReader returns reader for given format
func getFormatReader(format uint8, sourceFd io.Reader) io.Reader {
	var r io.Reader

	switch format {
	case _FORMAT_GZIP:
		r, _ = gzip.NewReader(sourceFd)
	case _FORMAT_BZIP:
		r = bzip2.NewReader(sourceFd)
	case _FORMAT_XZ:
		r, _ = xz.NewReader(sourceFd)
	case _FORMAT_ZSTD:
		r, _ = zstd.NewReader(sourceFd)
	default:
		r = sourceFd
	}

	return r
}
