package search

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"sort"
	"strings"

	"github.com/essentialkaos/ek/v12/path"
	"github.com/essentialkaos/ek/v12/strutil"

	"github.com/essentialkaos/rep/v3/repo/data"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	TERM_UNKNOWN uint8 = iota
	TERM_NAME
	TERM_ARCH
	TERM_VERSION
	TERM_RELEASE
	TERM_EPOCH
	TERM_PROVIDES
	TERM_REQUIRES
	TERM_RECOMMENDS
	TERM_CONFLICTS
	TERM_OBSOLETES
	TERM_ENHANCES
	TERM_SUGGESTS
	TERM_SUPPLEMENTS
	TERM_FILE
	TERM_SOURCE
	TERM_LICENSE
	TERM_VENDOR
	TERM_GROUP
	TERM_DATE_ADD
	TERM_DATE_BUILD
	TERM_BUILD_HOST
	TERM_SIZE
	TERM_PAYLOAD
)

const (
	TERM_MOD_NEGATIVE uint8 = 1 << iota
)

const (
	DEP_FLAG_ANY DepFlag = 0
	DEP_FLAG_EQ  DepFlag = 1 // =
	DEP_FLAG_LT  DepFlag = 2 // <
	DEP_FLAG_LE  DepFlag = 3 // <=
	DEP_FLAG_GT  DepFlag = 4 // >
	DEP_FLAG_GE  DepFlag = 5 // >=
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	_SQL_QUERY_TEMPLATE = `SELECT pkgKey FROM %s WHERE %s;`
)

// ////////////////////////////////////////////////////////////////////////////////// //

// DepFlag is dependency comparison flag
type DepFlag uint8

// Term is search term struct (part of search query)
type Term struct {
	Value       interface{}
	Type        uint8
	Modificator uint8

	priority int
}

// Range contains range start and end
type Range struct {
	Start int64
	End   int64
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Query contains search terms
type Query []*Term

func (s Query) Len() int {
	return len(s)
}

func (s Query) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Query) Less(i, j int) bool {
	return s[i].priority < s[j].priority
}

// ////////////////////////////////////////////////////////////////////////////////// //

// termPrettyNameMap contains pretty name for each term type
var termPrettyNameMap = map[uint8]string{
	TERM_NAME:        "name",
	TERM_VERSION:     "version",
	TERM_RELEASE:     "release",
	TERM_EPOCH:       "epoch",
	TERM_PROVIDES:    "provides",
	TERM_REQUIRES:    "requires",
	TERM_RECOMMENDS:  "recommends",
	TERM_CONFLICTS:   "conflicts",
	TERM_OBSOLETES:   "obsoletes",
	TERM_ENHANCES:    "enhances",
	TERM_SUGGESTS:    "suggests",
	TERM_SUPPLEMENTS: "supplements",
	TERM_FILE:        "files",
	TERM_SOURCE:      "source",
	TERM_LICENSE:     "license",
	TERM_GROUP:       "group",
	TERM_VENDOR:      "vendor",
	TERM_DATE_ADD:    "date-add",
	TERM_DATE_BUILD:  "date-build",
	TERM_BUILD_HOST:  "build-host",
	TERM_SIZE:        "size",
	TERM_ARCH:        "arch",

	TERM_UNKNOWN: "unknown",
}

// termPriorityMap contains priority for each type of search term
var termPriorityMap = map[uint8]int{
	TERM_NAME:        1,
	TERM_VERSION:     2,
	TERM_RELEASE:     2,
	TERM_EPOCH:       3,
	TERM_PROVIDES:    4,
	TERM_REQUIRES:    4,
	TERM_RECOMMENDS:  4,
	TERM_CONFLICTS:   4,
	TERM_OBSOLETES:   4,
	TERM_ENHANCES:    4,
	TERM_SUGGESTS:    4,
	TERM_SUPPLEMENTS: 4,
	TERM_FILE:        4,
	TERM_SOURCE:      1,
	TERM_GROUP:       7,
	TERM_LICENSE:     7,
	TERM_VENDOR:      7,
	TERM_DATE_ADD:    2,
	TERM_DATE_BUILD:  2,
	TERM_BUILD_HOST:  7,
	TERM_SIZE:        8,
	TERM_ARCH:        0,
	TERM_PAYLOAD:     9,
}

// termTargetTableMap contains target table for each term
var termTargetTableMap = map[uint8]string{
	TERM_NAME:        "packages",
	TERM_ARCH:        "packages",
	TERM_VERSION:     "packages",
	TERM_RELEASE:     "packages",
	TERM_EPOCH:       "packages",
	TERM_PROVIDES:    "provides",
	TERM_REQUIRES:    "requires",
	TERM_RECOMMENDS:  "recommends",
	TERM_CONFLICTS:   "conflicts",
	TERM_OBSOLETES:   "obsoletes",
	TERM_ENHANCES:    "enhances",
	TERM_SUGGESTS:    "suggests",
	TERM_SUPPLEMENTS: "supplements",
	TERM_FILE:        "files",
	TERM_SOURCE:      "packages",
	TERM_LICENSE:     "packages",
	TERM_VENDOR:      "packages",
	TERM_GROUP:       "packages",
	TERM_DATE_ADD:    "packages",
	TERM_DATE_BUILD:  "packages",
	TERM_BUILD_HOST:  "packages",
	TERM_SIZE:        "packages",
	TERM_PAYLOAD:     "filelist",
}

// termTargetColumnMap contains target table for each term
var termTargetColumnMap = map[uint8]string{
	TERM_NAME:       "name",
	TERM_ARCH:       "arch",
	TERM_VERSION:    "version",
	TERM_RELEASE:    "release",
	TERM_EPOCH:      "epoch",
	TERM_FILE:       "name",
	TERM_SOURCE:     "rpm_sourcerpm",
	TERM_LICENSE:    "rpm_license",
	TERM_VENDOR:     "rpm_vendor",
	TERM_GROUP:      "rpm_group",
	TERM_DATE_ADD:   "time_file",
	TERM_DATE_BUILD: "time_build",
	TERM_BUILD_HOST: "rpm_buildhost",
	TERM_SIZE:       "size_package",
}

// termTargetDBMap contains target DB for each term
var termTargetDBMap = map[uint8]string{
	TERM_NAME:        data.DB_PRIMARY,
	TERM_ARCH:        data.DB_PRIMARY,
	TERM_VERSION:     data.DB_PRIMARY,
	TERM_RELEASE:     data.DB_PRIMARY,
	TERM_EPOCH:       data.DB_PRIMARY,
	TERM_PROVIDES:    data.DB_PRIMARY,
	TERM_REQUIRES:    data.DB_PRIMARY,
	TERM_RECOMMENDS:  data.DB_PRIMARY,
	TERM_CONFLICTS:   data.DB_PRIMARY,
	TERM_OBSOLETES:   data.DB_PRIMARY,
	TERM_ENHANCES:    data.DB_PRIMARY,
	TERM_SUGGESTS:    data.DB_PRIMARY,
	TERM_SUPPLEMENTS: data.DB_PRIMARY,
	TERM_FILE:        data.DB_PRIMARY,
	TERM_SOURCE:      data.DB_PRIMARY,
	TERM_LICENSE:     data.DB_PRIMARY,
	TERM_VENDOR:      data.DB_PRIMARY,
	TERM_GROUP:       data.DB_PRIMARY,
	TERM_DATE_ADD:    data.DB_PRIMARY,
	TERM_DATE_BUILD:  data.DB_PRIMARY,
	TERM_BUILD_HOST:  data.DB_PRIMARY,
	TERM_SIZE:        data.DB_PRIMARY,
	TERM_PAYLOAD:     data.DB_FILELISTS,
}

// ////////////////////////////////////////////////////////////////////////////////// //

// TermName creates name search term with given value and modificators
func TermName(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_NAME, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermArch creates arch search term with given value and modificators
func TermArch(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_ARCH, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermVersion creates version search term with given value and modificators
func TermVersion(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_VERSION, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermRelease creates release search term with given value and modificators
func TermRelease(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_RELEASE, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermEpoch creates epoch search term with given value and modificators
func TermEpoch(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_EPOCH, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermProvides creates provides search term with given value and modificators
func TermProvides(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_PROVIDES, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermRequires creates requires search term with given value and modificators
func TermRequires(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_REQUIRES, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermRecommends creates recommends search term with given value and modificators
func TermRecommends(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_RECOMMENDS, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermConflicts creates conflicts search term with given value and modificators
func TermConflicts(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_CONFLICTS, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermObsoletes creates obsoletes search term with given value and modificators
func TermObsoletes(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_OBSOLETES, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermEnhances creates enhances search term with given value and modificators
func TermEnhances(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_ENHANCES, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermSuggests creates suggests search term with given value and modificators
func TermSuggests(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_SUGGESTS, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermSupplements creates supplements search term with given value and modificators
func TermSupplements(value data.Dependency, mods ...uint8) *Term {
	return &Term{Type: TERM_SUPPLEMENTS, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermFiles creates payload search term with given value and modificators
func TermFile(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_FILE, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermSource creates source search term with given value and modificators
func TermSource(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_SOURCE, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermLicense creates license search term with given value and modificators
func TermLicense(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_LICENSE, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermVendor creates vendor search term with given value and modificators
func TermVendor(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_VENDOR, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermGroup creates group search term with given value and modificators
func TermGroup(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_GROUP, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermDateAdd creates add date search term with given value and modificators
func TermDateAdd(from, to int64, mods ...uint8) *Term {
	return &Term{Type: TERM_DATE_ADD, Value: Range{from, to}, Modificator: getModificatorFromSlice(mods)}
}

// TermDateBuild creates build date search term with given value and modificators
func TermDateBuild(from, to int64, mods ...uint8) *Term {
	return &Term{Type: TERM_DATE_BUILD, Value: Range{from, to}, Modificator: getModificatorFromSlice(mods)}
}

// TermBuildHost creates build host search term with given value and modificators
func TermBuildHost(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_BUILD_HOST, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// TermSize creates size search term with given value and modificators
func TermSize(from, to int64, mods ...uint8) *Term {
	return &Term{Type: TERM_SIZE, Value: Range{from, to}, Modificator: getModificatorFromSlice(mods)}
}

// TermPayload creates payload search term with given value and modificators
func TermPayload(value string, mods ...uint8) *Term {
	return &Term{Type: TERM_PAYLOAD, Value: value, Modificator: getModificatorFromSlice(mods)}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// String returns string representation of search term
func (t *Term) String() string {
	var strMod string

	if t.Modificator&TERM_MOD_NEGATIVE == TERM_MOD_NEGATIVE {
		strMod = "!"
	}

	return fmt.Sprintf("[%s%s:%s]", strMod, termPrettyNameMap[t.Type], t.Value)
}

// IsNegative returns true if is negative search term
func (t *Term) IsNegative() bool {
	return t.Modificator&TERM_MOD_NEGATIVE == TERM_MOD_NEGATIVE
}

// SQL returns target db and term as a slice with SQL queries
func (t *Term) SQL() (string, []string) {
	var result []string

	for _, cond := range termToCond(t) {
		result = append(result, fmt.Sprintf(
			_SQL_QUERY_TEMPLATE,
			termTargetTableMap[t.Type],
			cond,
		))
	}

	return termTargetDBMap[t.Type], result
}

// String returns string representation of range
func (r Range) String() string {
	return fmt.Sprintf("%d→%d", r.Start, r.End)
}

// Validate validate query terms
func (q Query) Validate() []error {
	if len(q) == 0 {
		return nil
	}

	var errs []error

	for index, term := range q {
		switch term.Value.(type) {
		case string, Range, data.Dependency:
			// skip
		default:
			errs = append(errs, fmt.Errorf("Search term %d:%s has unknown/unsupported value type", index, term))
		}

		if termTargetTableMap[term.Type] == "" {
			errs = append(errs, fmt.Errorf("Can't find DB table for term %d:%s", index, term))
		}
	}

	return errs
}

// Terms returns slice with terms sorted by their priority
func (q Query) Terms() []*Term {
	updateTermPriority(q)
	return q
}

// ////////////////////////////////////////////////////////////////////////////////// //

// updateTermPriority adds priority for each term in query
func updateTermPriority(query Query) {
	for _, term := range query {
		term.priority = termPriorityMap[term.Type]
	}

	sort.Sort(query)
}

// termToCond converts term to SQL condition
func termToCond(term *Term) []string {
	switch term.Type {
	case TERM_PAYLOAD:
		return genPayloadTermCond(term)
	case TERM_REQUIRES, TERM_PROVIDES, TERM_RECOMMENDS, TERM_CONFLICTS,
		TERM_OBSOLETES, TERM_ENHANCES, TERM_SUGGESTS, TERM_SUPPLEMENTS:
		return []string{genDepTermCond(term.Value.(data.Dependency), term.IsNegative())}
	default:
		return []string{genBasicTermCond(term)}
	}
}

// genBasicTermCond generates SQL condition for given term
func genBasicTermCond(term *Term) string {
	var cond string

	switch t := term.Value.(type) {
	case string:
		cond = genStrTermCond(t, term.IsNegative())
	case Range:
		cond = genRangeTermCond(t, term.IsNegative())
	}

	column := termTargetColumnMap[term.Type]

	if term.Type == TERM_SOURCE {
		if term.IsNegative() {
			return fmt.Sprintf(
				"(%s %s OR location_href %s OR substr(location_href, 3) %s)",
				column, cond, cond, cond,
			)
		} else {
			return fmt.Sprintf(
				"(%s %s OR location_href %s OR substr(location_href, 3) %s)",
				column, cond, cond, cond,
			)
		}
	}

	return fmt.Sprintf("%s %s", column, cond)
}

// genStrTermCond generates SQL condition for term with text value
func genStrTermCond(value string, isNegative bool) string {
	isGlob := path.IsGlob(value)
	isArray := strings.Contains(value, "|")

	value = sanitizeInput(value)

	switch {
	case isArray:
		return genArraySQL(value, isNegative)
	case isGlob:
		return genGlobSQL(value, isNegative)
	default:
		return genExactSQL(value, isNegative)
	}
}

// genRangeTermCond generates SQL condition for term with range data
func genRangeTermCond(value Range, isNegative bool) string {
	if isNegative {
		return fmt.Sprintf("NOT BETWEEN %d AND %d", value.Start, value.End)
	}

	return fmt.Sprintf("BETWEEN %d AND %d", value.Start, value.End)
}

// genRangeTermCond generates SQL condition for term with dep info
func genDepTermCond(value data.Dependency, isNegative bool) string {
	var result []string

	result = append(result,
		fmt.Sprintf("%s %s", "name", genStrTermCond(value.Name, isNegative)),
	)

	if value.Flag != data.COMP_FLAG_ANY {
		result = append(result,
			fmt.Sprintf("%s %s", "flags", genStrTermCond(value.Flag.String(), isNegative)),
		)

		if value.Epoch != "" {
			result = append(result,
				fmt.Sprintf("%s %s", "epoch", genStrTermCond(value.Epoch, isNegative)),
			)
		}

		if value.Version != "" {
			result = append(result,
				fmt.Sprintf("%s %s", "version", genStrTermCond(value.Version, isNegative)),
			)
		}

		if value.Release != "" {
			result = append(result,
				fmt.Sprintf("%s %s", "release", genStrTermCond(value.Release, isNegative)),
			)
		}
	}

	return strings.Join(result, " AND ")
}

// genPayloadTermCond generates SQL condition for given payload term
func genPayloadTermCond(term *Term) []string {
	var result []string
	var negFlag int

	dirname, filename := path.Split(term.Value.(string))

	isDirGlob := path.IsGlob(dirname)
	isFileGlob := path.IsGlob(filename)

	dirname = sanitizeInput(dirname)
	dirname = strings.TrimRight(dirname, "/")
	filename = sanitizeInput(filename)

	if term.IsNegative() {
		negFlag = 1
	}

	switch {
	case dirname == "" || dirname == ".":
		if isFileGlob {
			result = append(result,
				fmt.Sprintf("filenames %s", genGlobSQL(filename, term.IsNegative())),
			)
		} else {
			result = append(result,
				fmt.Sprintf("filenames %s", genLikeSQL(filename, term.IsNegative())),
			)
		}
	case filename == "*":
		if isDirGlob {
			result = append(result,
				fmt.Sprintf("dirname %s", genGlobSQL(dirname, term.IsNegative())),
			)
		} else {
			result = append(result,
				fmt.Sprintf("dirname %s", genLikeSQL(dirname, term.IsNegative())),
			)
		}
	case !isFileGlob:
		if !isDirGlob {
			result = append(result,
				fmt.Sprintf(
					"dirname %s AND filenames %s",
					genExactSQL(dirname, term.IsNegative()),
					genLikeSQL(filename, term.IsNegative()),
				),
			)
		} else {
			result = append(result,
				fmt.Sprintf(
					"dirname %s AND filenames %s",
					genGlobSQL(dirname, term.IsNegative()),
					genLikeSQL(filename, term.IsNegative()),
				),
			)
		}
	default:
		result = append(result,
			fmt.Sprintf(
				"length(filetypes) = 1 AND (dirname || \"/\" || filenames) %s",
				genGlobSQL(dirname+"/"+filename, term.IsNegative()),
			),
			fmt.Sprintf(
				"length(filetypes) > 1 AND filelist_globber(\"%s\", dirname, filenames, %d)",
				dirname+"/"+filename, negFlag,
			),
		)
	}

	return result
}

// genArraySQL generates part of SQL query for array
func genArraySQL(value string, isNegative bool) string {
	values := strings.Split(value, "|")

	for index := range values {
		values[index] = "\"" + values[index] + "\""
	}

	if isNegative {
		return "NOT IN (" + strings.Join(values, ",") + ")"
	}

	return "IN (" + strings.Join(values, ",") + ")"
}

// genGlobSQL generates part of SQL query for glob
func genGlobSQL(value string, isNegative bool) string {
	if isNegative {
		return "NOT GLOB \"" + value + "\""
	}

	return "GLOB \"" + value + "\""
}

// genLikeSQL generates part of SQL query for pattern value
func genLikeSQL(value string, isNegative bool) string {
	if isNegative {
		return "NOT LIKE \"%" + value + "%\""
	}

	return "LIKE \"%" + value + "%\""
}

// genExactSQL generates part of SQL query for exact value
func genExactSQL(value string, isNegative bool) string {
	if isNegative {
		return "!= \"" + value + "\""
	}

	return "= \"" + value + "\""
}

// getModificatorFromSlice merges modificator flags
func getModificatorFromSlice(m []uint8) uint8 {
	if len(m) == 0 {
		return 0
	}

	var result uint8

	for _, mod := range m {
		result |= mod
	}

	return result
}

// sanitizeInput sanitizes user input
func sanitizeInput(data string) string {
	if data == "" {
		return ""
	}

	result := data

	result = strutil.ReplaceAll(result, "'\"", "")
	result = strutil.ReplaceAll(result, "^$()<>{}#;!=", " ")

	return result
}
