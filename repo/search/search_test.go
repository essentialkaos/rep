package search

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	"github.com/essentialkaos/rep/v3/repo/data"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type SearchSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&SearchSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *SearchSuite) TestHelpers(c *C) {
	d := data.Dependency{Name: "test"}

	c.Assert(TermName("test").Type, Equals, TERM_NAME)
	c.Assert(TermArch("test").Type, Equals, TERM_ARCH)
	c.Assert(TermVersion("test").Type, Equals, TERM_VERSION)
	c.Assert(TermRelease("test").Type, Equals, TERM_RELEASE)
	c.Assert(TermEpoch("test").Type, Equals, TERM_EPOCH)
	c.Assert(TermProvides(d).Type, Equals, TERM_PROVIDES)
	c.Assert(TermRequires(d).Type, Equals, TERM_REQUIRES)
	c.Assert(TermRecommends(d).Type, Equals, TERM_RECOMMENDS)
	c.Assert(TermConflicts(d).Type, Equals, TERM_CONFLICTS)
	c.Assert(TermObsoletes(d).Type, Equals, TERM_OBSOLETES)
	c.Assert(TermEnhances(d).Type, Equals, TERM_ENHANCES)
	c.Assert(TermSuggests(d).Type, Equals, TERM_SUGGESTS)
	c.Assert(TermSupplements(d).Type, Equals, TERM_SUPPLEMENTS)
	c.Assert(TermFile("test").Type, Equals, TERM_FILE)
	c.Assert(TermSource("test").Type, Equals, TERM_SOURCE)
	c.Assert(TermLicense("test").Type, Equals, TERM_LICENSE)
	c.Assert(TermVendor("test").Type, Equals, TERM_VENDOR)
	c.Assert(TermGroup("test").Type, Equals, TERM_GROUP)
	c.Assert(TermDateAdd(0, 1).Type, Equals, TERM_DATE_ADD)
	c.Assert(TermDateBuild(0, 1).Type, Equals, TERM_DATE_BUILD)
	c.Assert(TermBuildHost("test").Type, Equals, TERM_BUILD_HOST)
	c.Assert(TermSize(0, 1).Type, Equals, TERM_SIZE)
	c.Assert(TermPayload("file").Type, Equals, TERM_PAYLOAD)
}

func (s *SearchSuite) TestTermsHelpers(c *C) {
	t1 := TermName("test")
	t2 := TermName("test", TERM_MOD_NEGATIVE)
	t3 := TermSize(10, 65)

	c.Assert(t1.String(), Equals, "[name:test]")
	c.Assert(t2.String(), Equals, "[!name:test]")
	c.Assert(t3.String(), Equals, "[size:10→65]")

	c.Assert(t1.IsNegative(), Equals, false)
	c.Assert(t2.IsNegative(), Equals, true)
}

func (s *SearchSuite) TestTermsValidation(c *C) {
	q := Query{}
	c.Assert(q.Validate(), HasLen, 0)

	q = Query{TermName("test")}
	c.Assert(q.Validate(), HasLen, 0)

	q = Query{&Term{Type: 255, Value: "test"}}
	c.Assert(q.Validate(), HasLen, 1)

	q = Query{&Term{Type: TERM_NAME, Value: nil}}
	c.Assert(q.Validate(), HasLen, 1)
}

func (s *SearchSuite) TestQueryToSQL(c *C) {
	q := Query{
		TermLicense("*Apache*"),
		TermVersion("1.*"),
		TermName("test"),
	}

	c.Assert(q, NotNil)

	terms := q.Terms()

	c.Assert(terms, HasLen, 3)

	qd, qc := terms[0].SQL()
	c.Assert(qd, Equals, "primary")
	c.Assert(qc, DeepEquals, []string{"SELECT pkgKey FROM packages WHERE name = \"test\";"})

	qd, qc = terms[1].SQL()
	c.Assert(qd, Equals, "primary")
	c.Assert(qc, DeepEquals, []string{"SELECT pkgKey FROM packages WHERE version GLOB \"1.*\";"})

	qd, qc = terms[2].SQL()
	c.Assert(qd, Equals, "primary")
	c.Assert(qc, DeepEquals, []string{"SELECT pkgKey FROM packages WHERE rpm_license GLOB \"*Apache*\";"})

	q = Query{TermPayload("/[a-z]/file.*")}
	terms = q.Terms()
	c.Assert(terms, HasLen, 1)

	qd, qc = terms[0].SQL()
	c.Assert(qd, Equals, "filelists")
	c.Assert(qc, DeepEquals, []string{
		"SELECT pkgKey FROM filelist WHERE length(filetypes) = 1 AND (dirname || \"/\" || filenames) GLOB \"/[a-z]/file.*\";",
		"SELECT pkgKey FROM filelist WHERE length(filetypes) > 1 AND filelist_globber(\"/[a-z]/file.*\", dirname, filenames, 0);",
	})
}

func (s *SearchSuite) TestTermToCond(c *C) {
	c.Assert(tc(TermName("abcd")), Equals, "name = \"abcd\"")
	c.Assert(tc(TermName("abcd", TERM_MOD_NEGATIVE)), Equals, "name != \"abcd\"")
	c.Assert(tc(TermName("abcd*")), Equals, "name GLOB \"abcd*\"")
	c.Assert(tc(TermName("abcd*", TERM_MOD_NEGATIVE)), Equals, "name NOT GLOB \"abcd*\"")
	c.Assert(tc(TermName("ab|cd")), Equals, "name IN (\"ab\",\"cd\")")
	c.Assert(tc(TermName("ab|cd", TERM_MOD_NEGATIVE)), Equals, "name NOT IN (\"ab\",\"cd\")")
	c.Assert(tc(TermSource("abcd")), Equals, "(rpm_sourcerpm = \"abcd\" OR location_href = \"abcd\" OR substr(location_href, 3) = \"abcd\")")
	c.Assert(tc(TermSource("abcd", TERM_MOD_NEGATIVE)), Equals, "(rpm_sourcerpm != \"abcd\" OR location_href != \"abcd\" OR substr(location_href, 3) != \"abcd\")")
	c.Assert(tc(TermSize(0, 100)), Equals, "size_package BETWEEN 0 AND 100")
	c.Assert(tc(TermSize(0, 100, TERM_MOD_NEGATIVE)), Equals, "size_package NOT BETWEEN 0 AND 100")

	d := data.Dependency{
		Name:    "test",
		Epoch:   "1",
		Version: "2.3",
		Release: "0.el7",
		Flag:    data.COMP_FLAG_GT,
	}

	c.Assert(tc(TermRequires(d)), Equals, "name = \"test\" AND flags = \"GT\" AND epoch = \"1\" AND version = \"2.3\" AND release = \"0.el7\"")
	c.Assert(tc(TermRequires(d, TERM_MOD_NEGATIVE)), Equals, "name != \"test\" AND flags != \"GT\" AND epoch != \"1\" AND version != \"2.3\" AND release != \"0.el7\"")
}

func (s *SearchSuite) TestPayloadTermToCond(c *C) {
	q := genPayloadTermCond(TermPayload("/test/abcd", 0))
	c.Assert(q, DeepEquals, []string{"dirname = \"/test\" AND filenames LIKE \"%abcd%\""})

	q = genPayloadTermCond(TermPayload("/test/abcd", 1))
	c.Assert(q, DeepEquals, []string{"dirname != \"/test\" AND filenames NOT LIKE \"%abcd%\""})

	q = genPayloadTermCond(TermPayload("/test/*", 0))
	c.Assert(q, DeepEquals, []string{"dirname LIKE \"%/test%\""})

	q = genPayloadTermCond(TermPayload("/test/*", 1))
	c.Assert(q, DeepEquals, []string{"dirname NOT LIKE \"%/test%\""})

	q = genPayloadTermCond(TermPayload("/[a-z]/*", 0))
	c.Assert(q, DeepEquals, []string{"dirname GLOB \"/[a-z]\""})

	q = genPayloadTermCond(TermPayload("/[a-z]/file", 0))
	c.Assert(q, DeepEquals, []string{"dirname GLOB \"/[a-z]\" AND filenames LIKE \"%file%\""})

	q = genPayloadTermCond(TermPayload("file.log", 0))
	c.Assert(q, DeepEquals, []string{"filenames LIKE \"%file.log%\""})

	q = genPayloadTermCond(TermPayload("file.*", 0))
	c.Assert(q, DeepEquals, []string{"filenames GLOB \"file.*\""})

	q = genPayloadTermCond(TermPayload("/test/[a-z]/test.*", 0))
	c.Assert(q, DeepEquals, []string{
		"length(filetypes) = 1 AND (dirname || \"/\" || filenames) GLOB \"/test/[a-z]/test.*\"",
		"length(filetypes) > 1 AND filelist_globber(\"/test/[a-z]/test.*\", dirname, filenames, 0)",
	})

	q = genPayloadTermCond(TermPayload("/test/[a-z]/test.*", 1))
	c.Assert(q, DeepEquals, []string{
		"length(filetypes) = 1 AND (dirname || \"/\" || filenames) NOT GLOB \"/test/[a-z]/test.*\"",
		"length(filetypes) > 1 AND filelist_globber(\"/test/[a-z]/test.*\", dirname, filenames, 1)",
	})

	// c.Assert(q, DeepEquals, []string{})
}

func (s *SearchSuite) TestAux(c *C) {
	c.Assert(sanitizeInput(""), Equals, "")
}

// ////////////////////////////////////////////////////////////////////////////////// //

func tc(term *Term) string {
	return termToCond(term)[0]
}
