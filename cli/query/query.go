package query

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/essentialkaos/ek/v12/fmtutil"
	"github.com/essentialkaos/ek/v12/mathutil"
	"github.com/essentialkaos/ek/v12/strutil"
	"github.com/essentialkaos/ek/v12/timeutil"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/search"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	TERM_SHORT_NAME        = "n"
	TERM_SHORT_VERSION     = "v"
	TERM_SHORT_RELEASE     = "r"
	TERM_SHORT_EPOCH       = "e"
	TERM_SHORT_ARCH        = "a"
	TERM_SHORT_SOURCE      = "s"
	TERM_SHORT_LICENSE     = "l"
	TERM_SHORT_GROUP       = "g"
	TERM_SHORT_VENDOR      = "V"
	TERM_SHORT_PROVIDES    = "P"
	TERM_SHORT_REQUIRES    = "R"
	TERM_SHORT_RECOMMENDS  = "RC"
	TERM_SHORT_CONFLICTS   = "C"
	TERM_SHORT_OBSOLETES   = "O"
	TERM_SHORT_ENHANCES    = "E"
	TERM_SHORT_SUGGESTS    = "SG"
	TERM_SHORT_SUPPLEMENTS = "SP"
	TERM_SHORT_FILE        = "f"
	TERM_SHORT_DATE_ADD    = "d"
	TERM_SHORT_DATE_BUILD  = "D"
	TERM_SHORT_BUILD_HOST  = "h"
	TERM_SHORT_SIZE        = "S"
	TERM_SHORT_PAYLOAD     = "@"

	TERM_NAME        = "name"
	TERM_VERSION     = "version"
	TERM_RELEASE     = "release"
	TERM_EPOCH       = "epoch"
	TERM_ARCH        = "arch"
	TERM_SOURCE      = "source"
	TERM_LICENSE     = "license"
	TERM_GROUP       = "group"
	TERM_VENDOR      = "vendor"
	TERM_PROVIDES    = "provides"
	TERM_REQUIRES    = "requires"
	TERM_RECOMMENDS  = "recommends"
	TERM_CONFLICTS   = "conflicts"
	TERM_OBSOLETES   = "obsoletes"
	TERM_ENHANCES    = "enhances"
	TERM_SUGGESTS    = "suggests"
	TERM_SUPPLEMENTS = "supplements"
	TERM_FILE        = "file"
	TERM_DATE_ADD    = "date-add"
	TERM_DATE_BUILD  = "date-build"
	TERM_BUILD_HOST  = "host"
	TERM_SIZE        = "size"
	TERM_PAYLOAD     = "payload"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var terms = map[string]uint8{
	TERM_SHORT_NAME:        search.TERM_NAME,
	TERM_SHORT_VERSION:     search.TERM_VERSION,
	TERM_SHORT_RELEASE:     search.TERM_RELEASE,
	TERM_SHORT_EPOCH:       search.TERM_EPOCH,
	TERM_SHORT_PROVIDES:    search.TERM_PROVIDES,
	TERM_SHORT_REQUIRES:    search.TERM_REQUIRES,
	TERM_SHORT_RECOMMENDS:  search.TERM_RECOMMENDS,
	TERM_SHORT_CONFLICTS:   search.TERM_CONFLICTS,
	TERM_SHORT_OBSOLETES:   search.TERM_OBSOLETES,
	TERM_SHORT_ENHANCES:    search.TERM_ENHANCES,
	TERM_SHORT_SUGGESTS:    search.TERM_SUGGESTS,
	TERM_SHORT_SUPPLEMENTS: search.TERM_SUPPLEMENTS,
	TERM_SHORT_FILE:        search.TERM_FILE,
	TERM_SHORT_SOURCE:      search.TERM_SOURCE,
	TERM_SHORT_LICENSE:     search.TERM_LICENSE,
	TERM_SHORT_GROUP:       search.TERM_GROUP,
	TERM_SHORT_VENDOR:      search.TERM_VENDOR,
	TERM_SHORT_DATE_ADD:    search.TERM_DATE_ADD,
	TERM_SHORT_DATE_BUILD:  search.TERM_DATE_BUILD,
	TERM_SHORT_BUILD_HOST:  search.TERM_BUILD_HOST,
	TERM_SHORT_SIZE:        search.TERM_SIZE,
	TERM_SHORT_ARCH:        search.TERM_ARCH,
	TERM_SHORT_PAYLOAD:     search.TERM_PAYLOAD,

	TERM_NAME:        search.TERM_NAME,
	TERM_VERSION:     search.TERM_VERSION,
	TERM_RELEASE:     search.TERM_RELEASE,
	TERM_EPOCH:       search.TERM_EPOCH,
	TERM_PROVIDES:    search.TERM_PROVIDES,
	TERM_REQUIRES:    search.TERM_REQUIRES,
	TERM_RECOMMENDS:  search.TERM_RECOMMENDS,
	TERM_CONFLICTS:   search.TERM_CONFLICTS,
	TERM_OBSOLETES:   search.TERM_OBSOLETES,
	TERM_ENHANCES:    search.TERM_ENHANCES,
	TERM_SUGGESTS:    search.TERM_SUGGESTS,
	TERM_SUPPLEMENTS: search.TERM_SUPPLEMENTS,
	TERM_FILE:        search.TERM_FILE,
	TERM_SOURCE:      search.TERM_SOURCE,
	TERM_LICENSE:     search.TERM_LICENSE,
	TERM_GROUP:       search.TERM_GROUP,
	TERM_VENDOR:      search.TERM_VENDOR,
	TERM_DATE_ADD:    search.TERM_DATE_ADD,
	TERM_DATE_BUILD:  search.TERM_DATE_BUILD,
	TERM_BUILD_HOST:  search.TERM_BUILD_HOST,
	TERM_SIZE:        search.TERM_SIZE,
	TERM_ARCH:        search.TERM_ARCH,
	TERM_PAYLOAD:     search.TERM_PAYLOAD,
}

var depRegex = regexp.MustCompile(`([a-zA-Z0-9\._\-:\(\)\*]+)(>=|<=|>|<|=)?([0-9]:)?([0-9a-z\.\*]+)?-?(.*)?`)

// ////////////////////////////////////////////////////////////////////////////////// //

// Parse parses string with data and creates search query
func Parse(q []string) (search.Query, error) {
	var query search.Query

	for _, rawTerm := range q {
		if len(rawTerm) == 0 {
			continue
		}

		term, err := parseTerm(rawTerm)

		if err != nil {
			return nil, err
		}

		query = append(query, term)
	}

	return query, nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// parseTerm parses query term
func parseTerm(t string) (*search.Term, error) {
	name, value, isNegative := extractTermInfo(t)
	termType, mod := terms[name], uint8(0)

	if name != "" {
		if termType == search.TERM_UNKNOWN {
			return nil, fmt.Errorf("Unknown query term \"%s\"", name)
		}
	} else {
		termType = 255 // term without name = name prefix search
	}

	if isNegative {
		mod = search.TERM_MOD_NEGATIVE
	}

	switch termType {
	case search.TERM_NAME:
		return search.TermName(value, mod), nil
	case search.TERM_VERSION:
		return search.TermVersion(value, mod), nil
	case search.TERM_RELEASE:
		return search.TermRelease(value, mod), nil
	case search.TERM_EPOCH:
		return search.TermEpoch(value, mod), nil
	case search.TERM_ARCH:
		value = strings.ReplaceAll(value, "x64", "x86_64")
		value = strings.ReplaceAll(value, "x32", "i386")
		return search.TermArch(value, mod), nil
	case search.TERM_REQUIRES:
		return parseDepTerm(search.TERM_REQUIRES, value, mod)
	case search.TERM_PROVIDES:
		return parseDepTerm(search.TERM_PROVIDES, value, mod)
	case search.TERM_RECOMMENDS:
		return parseDepTerm(search.TERM_RECOMMENDS, value, mod)
	case search.TERM_CONFLICTS:
		return parseDepTerm(search.TERM_CONFLICTS, value, mod)
	case search.TERM_OBSOLETES:
		return parseDepTerm(search.TERM_OBSOLETES, value, mod)
	case search.TERM_ENHANCES:
		return parseDepTerm(search.TERM_ENHANCES, value, mod)
	case search.TERM_SUGGESTS:
		return parseDepTerm(search.TERM_SUGGESTS, value, mod)
	case search.TERM_SUPPLEMENTS:
		return parseDepTerm(search.TERM_SUPPLEMENTS, value, mod)
	case search.TERM_FILE:
		return search.TermFile(value, mod), nil
	case search.TERM_SOURCE:
		return search.TermSource(value, mod), nil
	case search.TERM_LICENSE:
		return search.TermLicense(value, mod), nil
	case search.TERM_VENDOR:
		return search.TermVendor(value, mod), nil
	case search.TERM_GROUP:
		return search.TermGroup(value, mod), nil
	case search.TERM_BUILD_HOST:
		return search.TermBuildHost(value, mod), nil
	case search.TERM_DATE_ADD, search.TERM_DATE_BUILD:
		return parseDateTerm(termType, value, mod)
	case search.TERM_SIZE:
		return parseSizeTerm(value, mod)
	case search.TERM_PAYLOAD:
		return search.TermPayload(value, mod), nil
	default:
		return search.TermName(value+"*", mod), nil
	}
}

// parseDateTerm parses date term
func parseDateTerm(termType uint8, value string, mod uint8) (*search.Term, error) {
	dur, err := timeutil.ParseDuration(value, 'd')

	if err != nil {
		return nil, fmt.Errorf("Can't parse \"%s\" as duration: %v", value, err)
	}

	to := time.Now().Unix()
	from := to - dur

	return &search.Term{Type: termType, Value: search.Range{from, to}, Modificator: mod}, nil
}

// parseSizeTerm parses size term
func parseSizeTerm(value string, mod uint8) (*search.Term, error) {
	var from, to uint64

	switch {
	case strings.HasSuffix(value, "-"):
		from = 0
		to = fmtutil.ParseSize(strings.TrimRight(value, "-"))

	case strings.HasSuffix(value, "+"):
		from = fmtutil.ParseSize(strings.TrimRight(value, "+"))
		to = 1024 * 1024 * 1024

	case strings.Contains(value, "-"):
		from = fmtutil.ParseSize(strutil.ReadField(value, 0, false, "-"))
		to = fmtutil.ParseSize(strutil.ReadField(value, 1, false, "-"))

	default:
		size := fmtutil.ParseSize(value)
		diff := uint64(float64(size) * 0.2)
		from = mathutil.BetweenU64(size-diff, 0, 1024*1024*1024)
		to = mathutil.BetweenU64(size+diff, 0, 1024*1024*1024)
	}

	if from > to {
		return nil, fmt.Errorf("Range %d→%d is invalid", from, to)
	}

	return search.TermSize(int64(from), int64(to), mod), nil
}

// parseDepTerm parses term with dependency info (used for requires/provides)
func parseDepTerm(termType uint8, value string, mod uint8) (*search.Term, error) {
	dep := extractDepInfo(value)

	if dep.Flag != data.COMP_FLAG_ANY {
		if dep.Epoch == "" && dep.Version == "" && dep.Release == "" {
			return nil, fmt.Errorf("Can't use \"%v\" - condition without value", value)
		}
	}

	return &search.Term{Type: termType, Value: dep, Modificator: mod}, nil
}

// extractDepInfo parses and extracts dependency info
func extractDepInfo(v string) data.Dependency {
	info := depRegex.FindStringSubmatch(v)

	return data.Dependency{
		Name:    info[1],
		Epoch:   strings.TrimRight(info[3], ":"),
		Version: info[4],
		Release: info[5],
		Flag:    condToFlag(info[2]),
	}
}

// extractTermInfo extracts info from token
func extractTermInfo(t string) (string, string, bool) {
	if !strings.Contains(t, ":") {
		return "", t, false
	}

	sepIndex := strings.Index(t, ":")
	name := t[:sepIndex]
	value := t[sepIndex+1:]
	isNegative := false

	if strings.HasPrefix(value, ":") {
		isNegative = true
		value = value[1:]
	}

	return name, value, isNegative
}

// condToFlag transforms conditional to flag
func condToFlag(c string) data.CompFlag {
	switch c {
	case ">=":
		return data.COMP_FLAG_GE
	case "<=":
		return data.COMP_FLAG_LE
	case ">":
		return data.COMP_FLAG_GT
	case "<":
		return data.COMP_FLAG_LT
	case "=":
		return data.COMP_FLAG_EQ
	default:
		return data.COMP_FLAG_ANY
	}
}