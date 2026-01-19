package helpers

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"os"
	"strings"

	"github.com/sassoftware/go-rpmutils"

	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// GuessFileArch extracts arch name from RPM file name
func GuessFileArch(fileName string) string {
	index := strings.LastIndex(fileName, ".")

	if index == -1 {
		return ""
	}

	fileName = fileName[:index]

	index = strings.LastIndex(fileName, ".")

	if index == -1 || index == len(fileName) {
		return ""
	}

	switch fileName[index+1:] {
	case data.ARCH_SRC, data.ARCH_NOARCH, data.ARCH_I386, data.ARCH_I586,
		data.ARCH_I686, data.ARCH_X64, data.ARCH_AARCH64, data.ARCH_PPC64,
		data.ARCH_PPC64LE, data.ARCH_ARM, data.ARCH_ARMV7HL:
		return fileName[index+1:]
	}

	return ""
}

// ExtractPackageArch reads package arch tag from header
func ExtractPackageArch(rpmFile string) (string, error) {
	fd, err := os.OpenFile(rpmFile, os.O_RDONLY, 0)

	if err != nil {
		return "", err
	}

	defer fd.Close()

	r := bufio.NewReader(fd)

	header, err := rpmutils.ReadHeader(r)

	if err != nil {
		return "", err
	}

	if !header.HasTag(rpmutils.SOURCERPM) {
		return data.ARCH_SRC, nil
	}

	return header.GetString(rpmutils.ARCH)
}
