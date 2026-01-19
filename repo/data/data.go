package data

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"strings"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Repository types
const (
	REPO_RELEASE = "release"
	REPO_TESTING = "testing"
)

// DB types
const (
	DB_PRIMARY   = "primary"
	DB_OTHER     = "other"
	DB_FILELISTS = "filelists"
)

// Arch flags
const (
	ARCH_FLAG_UNKNOWN ArchFlag = 0
	ARCH_FLAG_SRC     ArchFlag = 1 << iota
	ARCH_FLAG_NOARCH
	ARCH_FLAG_I386
	ARCH_FLAG_I586
	ARCH_FLAG_I686
	ARCH_FLAG_X64
	ARCH_FLAG_AARCH64
	ARCH_FLAG_PPC64
	ARCH_FLAG_PPC64LE
	ARCH_FLAG_ARM
	ARCH_FLAG_ARMV7HL
)

// Arch names
const (
	ARCH_SRC     = "src"
	ARCH_NOARCH  = "noarch"
	ARCH_I386    = "i386"
	ARCH_I586    = "i586"
	ARCH_I686    = "i686"
	ARCH_X64     = "x86_64"
	ARCH_AARCH64 = "aarch64"
	ARCH_PPC64   = "ppc64"
	ARCH_PPC64LE = "ppc64le"
	ARCH_ARM     = "arm"
	ARCH_ARMV7HL = "armv7hl"
)

// Comparison flags
const (
	COMP_FLAG_ANY CompFlag = 0
	COMP_FLAG_EQ  CompFlag = 1 // =
	COMP_FLAG_LT  CompFlag = 2 // <
	COMP_FLAG_LE  CompFlag = 3 // <=
	COMP_FLAG_GT  CompFlag = 4 // >
	COMP_FLAG_GE  CompFlag = 5 // >=
)

// ////////////////////////////////////////////////////////////////////////////////// //

// CompFlag is comparison flag
type CompFlag uint8

// Dependency contains info about dependency
type Dependency struct {
	Name    string
	Epoch   string
	Version string
	Release string
	Flag    CompFlag
}

// ArchInfo is arch flag
type ArchFlag uint16

// ArchInfo contains info about specific arch
type ArchInfo struct {
	Dir  string
	Tag  string
	Flag ArchFlag
}

// ////////////////////////////////////////////////////////////////////////////////// //

// SupportedArchs is a slice with info about supported archs
var SupportedArchs = map[string]ArchInfo{
	ARCH_SRC:     {"SRPMS", "src", ARCH_FLAG_SRC},
	ARCH_NOARCH:  {"", "noarch", ARCH_FLAG_NOARCH},
	ARCH_I386:    {"i386", "x32", ARCH_FLAG_I386},
	ARCH_I586:    {"i586", "i586", ARCH_FLAG_I586},
	ARCH_I686:    {"i686", "i686", ARCH_FLAG_I686},
	ARCH_X64:     {"x86_64", "x64", ARCH_FLAG_X64},
	ARCH_AARCH64: {"aarch64", "aa64", ARCH_FLAG_AARCH64},
	ARCH_PPC64:   {"ppc64", "p64", ARCH_FLAG_PPC64},
	ARCH_PPC64LE: {"ppc64le", "p64l", ARCH_FLAG_PPC64LE},
	ARCH_ARM:     {"arm", "arm", ARCH_FLAG_ARM},
	ARCH_ARMV7HL: {"armv7hl", "arm7", ARCH_FLAG_ARMV7HL},
}

// ArchList is a slice with supported archs
var ArchList = []string{
	ARCH_SRC,
	ARCH_NOARCH,
	ARCH_I386,
	ARCH_I586,
	ARCH_I686,
	ARCH_X64,
	ARCH_AARCH64,
	ARCH_PPC64,
	ARCH_PPC64LE,
	ARCH_ARM,
	ARCH_ARMV7HL,
}

// BinArchList is a slice with supported binary archs
var BinArchList = []string{
	ARCH_I386,
	ARCH_I586,
	ARCH_I686,
	ARCH_X64,
	ARCH_AARCH64,
	ARCH_PPC64,
	ARCH_PPC64LE,
	ARCH_ARM,
	ARCH_ARMV7HL,
}

// DBList is a slice with names of databases
var DBList = []string{
	DB_PRIMARY,
	DB_OTHER,
	DB_FILELISTS,
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Has returns true if flag contains given arch flag
func (f ArchFlag) Has(flag ArchFlag) bool {
	if f == 0 {
		return false
	}

	return f&flag == flag
}

// String returns string representation of supported archs
func (f ArchFlag) String() string {
	var result []string

	for _, arch := range ArchList {
		if f.Has(SupportedArchs[arch].Flag) {
			result = append(result, arch)
		}
	}

	return strings.Join(result, "/")
}

// String returns string representation of comparison flag
func (f CompFlag) String() string {
	switch f {
	case COMP_FLAG_EQ:
		return "EQ"
	case COMP_FLAG_LT:
		return "LT"
	case COMP_FLAG_LE:
		return "LE"
	case COMP_FLAG_GT:
		return "GT"
	case COMP_FLAG_GE:
		return "GE"
	default:
		return ""
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// ParseComp parses text value of flag
func ParseComp(v string) CompFlag {
	switch v {
	case "EQ":
		return COMP_FLAG_EQ
	case "LT":
		return COMP_FLAG_LT
	case "LE":
		return COMP_FLAG_LE
	case "GT":
		return COMP_FLAG_GT
	case "GE":
		return COMP_FLAG_GE
	default:
		return COMP_FLAG_ANY
	}
}
