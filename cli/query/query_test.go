package query

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	"github.com/essentialkaos/rep/repo/data"
	"github.com/essentialkaos/rep/repo/search"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type QueryParserSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&QueryParserSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *QueryParserSuite) TestParser(c *C) {
	q, err := Parse(nil)

	c.Assert(q, HasLen, 0)

	q, err = Parse([]string{"", "", ""})

	c.Assert(q, HasLen, 0)

	q, err = Parse([]string{"k:test"})

	c.Assert(err, NotNil)
	c.Assert(q, IsNil)

	q, err = Parse([]string{"n:test"})

	c.Assert(err, IsNil)
	c.Assert(q, NotNil)
	c.Assert(q, HasLen, 1)
}

func (s *QueryParserSuite) TestTermParser(c *C) {
	t, err := parseTerm("k:test")

	c.Assert(t, IsNil)
	c.Assert(err, NotNil)

	checkTermParser(c, "test", search.TERM_NAME)
	checkTermParser(c, TERM_SHORT_NAME+":test", search.TERM_NAME)
	checkTermParser(c, TERM_SHORT_VERSION+":test", search.TERM_VERSION)
	checkTermParser(c, TERM_SHORT_RELEASE+":test", search.TERM_RELEASE)
	checkTermParser(c, TERM_SHORT_EPOCH+":test", search.TERM_EPOCH)
	checkTermParser(c, TERM_SHORT_ARCH+":test", search.TERM_ARCH)
	checkTermParser(c, TERM_SHORT_SOURCE+":test", search.TERM_SOURCE)
	checkTermParser(c, TERM_SHORT_LICENSE+":test", search.TERM_LICENSE)
	checkTermParser(c, TERM_SHORT_GROUP+":test", search.TERM_GROUP)
	checkTermParser(c, TERM_SHORT_FILE+":test", search.TERM_FILE)
	checkTermParser(c, TERM_SHORT_PROVIDES+":test", search.TERM_PROVIDES)
	checkTermParser(c, TERM_SHORT_REQUIRES+":test", search.TERM_REQUIRES)
	checkTermParser(c, TERM_SHORT_RECOMMENDS+":test", search.TERM_RECOMMENDS)
	checkTermParser(c, TERM_SHORT_CONFLICTS+":test", search.TERM_CONFLICTS)
	checkTermParser(c, TERM_SHORT_OBSOLETES+":test", search.TERM_OBSOLETES)
	checkTermParser(c, TERM_SHORT_ENHANCES+":test", search.TERM_ENHANCES)
	checkTermParser(c, TERM_SHORT_SUGGESTS+":test", search.TERM_SUGGESTS)
	checkTermParser(c, TERM_SHORT_SUPPLEMENTS+":test", search.TERM_SUPPLEMENTS)
	checkTermParser(c, TERM_SHORT_DATE_ADD+":1w", search.TERM_DATE_ADD)
	checkTermParser(c, TERM_SHORT_DATE_BUILD+":1w", search.TERM_DATE_BUILD)
	checkTermParser(c, TERM_SHORT_BUILD_HOST+":test", search.TERM_BUILD_HOST)
	checkTermParser(c, TERM_SHORT_SIZE+":1mb", search.TERM_SIZE)
	checkTermParser(c, TERM_SHORT_VENDOR+":test", search.TERM_VENDOR)
	checkTermParser(c, TERM_SHORT_PAYLOAD+":/test/file.log", search.TERM_PAYLOAD)

	checkTermParser(c, TERM_NAME+":test", search.TERM_NAME)
	checkTermParser(c, TERM_VERSION+":test", search.TERM_VERSION)
	checkTermParser(c, TERM_RELEASE+":test", search.TERM_RELEASE)
	checkTermParser(c, TERM_EPOCH+":test", search.TERM_EPOCH)
	checkTermParser(c, TERM_ARCH+":test", search.TERM_ARCH)
	checkTermParser(c, TERM_PROVIDES+":test", search.TERM_PROVIDES)
	checkTermParser(c, TERM_REQUIRES+":test", search.TERM_REQUIRES)
	checkTermParser(c, TERM_RECOMMENDS+":test", search.TERM_RECOMMENDS)
	checkTermParser(c, TERM_CONFLICTS+":test", search.TERM_CONFLICTS)
	checkTermParser(c, TERM_OBSOLETES+":test", search.TERM_OBSOLETES)
	checkTermParser(c, TERM_ENHANCES+":test", search.TERM_ENHANCES)
	checkTermParser(c, TERM_SUGGESTS+":test", search.TERM_SUGGESTS)
	checkTermParser(c, TERM_SUPPLEMENTS+":test", search.TERM_SUPPLEMENTS)
	checkTermParser(c, TERM_FILE+":test", search.TERM_FILE)
	checkTermParser(c, TERM_LICENSE+":test", search.TERM_LICENSE)
	checkTermParser(c, TERM_GROUP+":test", search.TERM_GROUP)
	checkTermParser(c, TERM_VENDOR+":test", search.TERM_VENDOR)
	checkTermParser(c, TERM_DATE_ADD+":1w", search.TERM_DATE_ADD)
	checkTermParser(c, TERM_DATE_BUILD+":1w", search.TERM_DATE_BUILD)
	checkTermParser(c, TERM_BUILD_HOST+":test", search.TERM_BUILD_HOST)
	checkTermParser(c, TERM_SIZE+":1mb", search.TERM_SIZE)
	checkTermParser(c, TERM_PAYLOAD+":/test/file.log", search.TERM_PAYLOAD)

	checkTermParser(c, TERM_SHORT_NAME+"::test", search.TERM_NAME)
}

func (s *QueryParserSuite) TestDateTermParser(c *C) {
	t, err := parseTerm("d:1w")

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)

	t, err = parseTerm("D:1w")

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)

	t, err = parseTerm("d:1z")

	c.Assert(t, IsNil)
	c.Assert(err, NotNil)
}

func (s *QueryParserSuite) TestSizeTermParser(c *C) {
	t, err := parseTerm("S:1mb")

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)
	c.Assert(t.Value.(search.Range).Start < 1024*1024, Equals, true)
	c.Assert(t.Value.(search.Range).End > 1024*1024, Equals, true)

	t, err = parseTerm("S:1mb+")

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)
	c.Assert(t.Value.(search.Range).Start == 1024*1024, Equals, true)
	c.Assert(t.Value.(search.Range).End == 1024*1024*1024, Equals, true)

	t, err = parseTerm("S:1mb-")

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)
	c.Assert(t.Value.(search.Range).Start == 0, Equals, true)
	c.Assert(t.Value.(search.Range).End == 1024*1024, Equals, true)

	t, err = parseTerm("S:1mb-2mb")

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)
	c.Assert(t.Value.(search.Range).Start == 1024*1024, Equals, true)
	c.Assert(t.Value.(search.Range).End == 2*1024*1024, Equals, true)

	t, err = parseTerm("S:2mb-1mb")

	c.Assert(t, IsNil)
	c.Assert(err, NotNil)
}

func (s *QueryParserSuite) TestDepNameParser(c *C) {
	dep := extractDepInfo("webkaos>=2:3.8.1-4.el7")
	c.Assert(dep.Name, Equals, "webkaos")
	c.Assert(dep.Epoch, Equals, "2")
	c.Assert(dep.Version, Equals, "3.8.1")
	c.Assert(dep.Release, Equals, "4.el7")
	c.Assert(dep.Flag, Equals, data.CompFlag(data.COMP_FLAG_GE))

	dep = extractDepInfo("webkaos")
	c.Assert(dep.Name, Equals, "webkaos")
	c.Assert(dep.Epoch, Equals, "")
	c.Assert(dep.Version, Equals, "")
	c.Assert(dep.Release, Equals, "")
	c.Assert(dep.Flag, Equals, data.CompFlag(data.COMP_FLAG_ANY))

	dep = extractDepInfo("webkaos>1.1.2")
	c.Assert(dep.Name, Equals, "webkaos")
	c.Assert(dep.Epoch, Equals, "")
	c.Assert(dep.Version, Equals, "1.1.2")
	c.Assert(dep.Release, Equals, "")
	c.Assert(dep.Flag, Equals, data.CompFlag(data.COMP_FLAG_GT))

	dep = extractDepInfo("webkaos=1.1.2-9")
	c.Assert(dep.Name, Equals, "webkaos")
	c.Assert(dep.Epoch, Equals, "")
	c.Assert(dep.Version, Equals, "1.1.2")
	c.Assert(dep.Release, Equals, "9")
	c.Assert(dep.Flag, Equals, data.CompFlag(data.COMP_FLAG_EQ))

	c.Assert(condToFlag(">="), Equals, data.CompFlag(data.COMP_FLAG_GE))
	c.Assert(condToFlag("<="), Equals, data.CompFlag(data.COMP_FLAG_LE))
	c.Assert(condToFlag(">"), Equals, data.CompFlag(data.COMP_FLAG_GT))
	c.Assert(condToFlag("<"), Equals, data.CompFlag(data.COMP_FLAG_LT))
	c.Assert(condToFlag("="), Equals, data.CompFlag(data.COMP_FLAG_EQ))
	c.Assert(condToFlag(""), Equals, data.CompFlag(data.COMP_FLAG_ANY))

	// Check parsing errors
	t, err := parseDepTerm(search.TERM_PROVIDES, "webkaos>=", 0)
	c.Assert(t, IsNil)
	c.Assert(err, NotNil)
}

// ////////////////////////////////////////////////////////////////////////////////// //

func checkTermParser(c *C, term string, termType uint8) {
	t, err := parseTerm(term)

	c.Assert(t, NotNil)
	c.Assert(err, IsNil)

	c.Assert(t.Type, Equals, termType)
}