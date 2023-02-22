package rpm

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2023 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// RPM_MAGIC is RPM magic number
// https://github.com/rpm-software-management/rpm/blob/master/lib/rpmlead.h#L15-L18
const RPM_MAGIC = 0xEDABEEDB

// Arch type list
// https://github.com/rpm-software-management/rpm/blob/master/rpmrc.in#L159-L258
const (
	ARCH_UNKNOWN      uint16 = 0
	ARCH_ATHLON       uint16 = 1
	ARCH_GEODE        uint16 = 1
	ARCH_PENTIUM4     uint16 = 1
	ARCH_PENTIUM3     uint16 = 1
	ARCH_I686         uint16 = 1
	ARCH_I586         uint16 = 1
	ARCH_I486         uint16 = 1
	ARCH_I386         uint16 = 1
	ARCH_X86_64       uint16 = 1
	ARCH_AMD64        uint16 = 1
	ARCH_IA32E        uint16 = 1
	ARCH_EM64T        uint16 = 1
	ARCH_ALPHA        uint16 = 2
	ARCH_ALPHAEV5     uint16 = 2
	ARCH_ALPHAEV56    uint16 = 2
	ARCH_ALPHAPCA56   uint16 = 2
	ARCH_ALPHAEV6     uint16 = 2
	ARCH_ALPHAEV67    uint16 = 2
	ARCH_SPARC64      uint16 = 2
	ARCH_SUN4U        uint16 = 2
	ARCH_SPARC64V     uint16 = 2
	ARCH_SPARC        uint16 = 3
	ARCH_SUN4         uint16 = 3
	ARCH_SUN4M        uint16 = 3
	ARCH_SUN4C        uint16 = 3
	ARCH_SUN4D        uint16 = 3
	ARCH_SPARCV8      uint16 = 3
	ARCH_SPARCV9      uint16 = 3
	ARCH_SPARCV9V     uint16 = 3
	ARCH_MIPS         uint16 = 4
	ARCH_MIPSEL       uint16 = 4
	ARCH_PPC          uint16 = 5
	ARCH_PPC8260      uint16 = 5
	ARCH_PPC8560      uint16 = 5
	ARCH_PPC32DY4     uint16 = 5
	ARCH_PPCISERIES   uint16 = 5
	ARCH_PPCPSERIES   uint16 = 5
	ARCH_M68K         uint16 = 6
	ARCH_IP           uint16 = 7
	ARCH_RS6000       uint16 = 8
	ARCH_IA64         uint16 = 9
	ARCH_MIPS64       uint16 = 11
	ARCH_MIPS64EL     uint16 = 11
	ARCH_ARMV3L       uint16 = 12
	ARCH_ARMV4B       uint16 = 12
	ARCH_ARMV4L       uint16 = 12
	ARCH_ARMV5TL      uint16 = 12
	ARCH_ARMV5TEL     uint16 = 12
	ARCH_ARMV5TEJL    uint16 = 12
	ARCH_ARMV6L       uint16 = 12
	ARCH_ARMV6HL      uint16 = 12
	ARCH_ARMV7L       uint16 = 12
	ARCH_ARMV7HL      uint16 = 12
	ARCH_ARMV7HNL     uint16 = 12
	ARCH_ARMV8L       uint16 = 12
	ARCH_ARMV8HL      uint16 = 12
	ARCH_M68KMINT     uint16 = 13
	ARCH_ATARIST      uint16 = 13
	ARCH_ATARISTE     uint16 = 13
	ARCH_ATARITT      uint16 = 13
	ARCH_FALCON       uint16 = 13
	ARCH_ATARICLONE   uint16 = 13
	ARCH_MILAN        uint16 = 13
	ARCH_HADES        uint16 = 13
	ARCH_S390         uint16 = 14
	ARCH_I370         uint16 = 14
	ARCH_S390X        uint16 = 15
	ARCH_PPC64        uint16 = 16
	ARCH_PPC64LE      uint16 = 16
	ARCH_PPC64PSERIES uint16 = 16
	ARCH_PPC64ISERIES uint16 = 16
	ARCH_PPC64P7      uint16 = 16
	ARCH_SH           uint16 = 17
	ARCH_SH3          uint16 = 17
	ARCH_SH4          uint16 = 17
	ARCH_SH4A         uint16 = 17
	ARCH_XTENSA       uint16 = 18
	ARCH_AARCH64      uint16 = 19
	ARCH_MIPSR6       uint16 = 20
	ARCH_MIPSR6EL     uint16 = 20
	ARCH_MIPS64R6     uint16 = 21
	ARCH_MIPS64R6EL   uint16 = 21
	ARCH_RISCV        uint16 = 22
	ARCH_RISCV64      uint16 = 22
)

// OS type list
// https://github.com/rpm-software-management/rpm/blob/master/rpmrc.in#L261-L291
const (
	OS_UNKNOWN     uint16 = 0
	OS_LINUX       uint16 = 1
	OS_IRIX        uint16 = 2
	OS_SOLARIS     uint16 = 3
	OS_SUNOS       uint16 = 4
	OS_AMIGAOS     uint16 = 5
	OS_AIX         uint16 = 5
	OS_HP_UX       uint16 = 6
	OS_OSF1        uint16 = 7
	OS_OSF4_0      uint16 = 7
	OS_OSF3_2      uint16 = 7
	OS_FREEBSD     uint16 = 8
	OS_SCO_SV      uint16 = 9
	OS_IRIX64      uint16 = 10
	OS_NEXTSTEP    uint16 = 11
	OS_BSD_OS      uint16 = 12
	OS_MACHTEN     uint16 = 13
	OS_CYGWIN32_NT uint16 = 14
	OS_CYGWIN32_95 uint16 = 15
	UNIX_SV        uint16 = 16
	OS_MINT        uint16 = 17
	OS_OS_390      uint16 = 18
	OS_VM_ESA      uint16 = 19
	OS_LINUX_390   uint16 = 20
	OS_LINUX_ESA   uint16 = 20
	OS_DARWIN      uint16 = 21
	OS_MACOSX      uint16 = 21
)

// SIGTYPE_HEADERSIG is signature type
// https://github.com/rpm-software-management/rpm/blob/master/lib/rpmlead.c#L21
const SIGTYPE_HEADERSIG uint16 = 5

// ////////////////////////////////////////////////////////////////////////////////// //

// LEAD is RPM LEAD header
// https://github.com/rpm-software-management/rpm/blob/master/lib/rpmlead.c#L33-L43
type LEAD struct {
	Name     string
	ArchType uint16
	OSType   uint16
	SigType  uint16
	Major    uint8
	Minor    uint8
	IsSrc    bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

// IsRPM checks if given file is an RPM file
func IsRPM(file string) bool {
	lead, err := readLead(file, 4)

	if err != nil {
		return false
	}

	magic := binary.BigEndian.Uint32(lead)

	return magic&0xffffffff == RPM_MAGIC
}

// ReadLead reads LEAD part of package
func ReadLEAD(file string) (LEAD, error) {
	lead, err := readLead(file, 80)

	if err != nil {
		return LEAD{}, err
	}

	magic := binary.BigEndian.Uint32(lead[0:4])

	if magic&0xffffffff != RPM_MAGIC {
		return LEAD{}, errors.New("File " + file + " is not an RPM package")
	}

	return LEAD{
		Major:    uint8(lead[4]),
		Minor:    uint8(lead[5]),
		IsSrc:    binary.BigEndian.Uint16(lead[6:8]) == 1,
		ArchType: binary.BigEndian.Uint16(lead[8:10]),
		Name:     strings.ReplaceAll(string(lead[10:76]), "\x00", ""),
		OSType:   binary.BigEndian.Uint16(lead[76:78]),
		SigType:  binary.BigEndian.Uint16(lead[78:80]),
	}, err
}

// ////////////////////////////////////////////////////////////////////////////////// //

// readLead reads first 80 bytes of RPM file
func readLead(file string, size int) ([]byte, error) {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	buf := make([]byte, size)
	_, err = io.ReadFull(fd, buf)

	if err != nil {
		return nil, err
	}

	return buf, nil
}
