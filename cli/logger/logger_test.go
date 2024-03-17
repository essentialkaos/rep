package logger

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"errors"
	"testing"

	"github.com/essentialkaos/ek/v12/fsutil"
	"github.com/essentialkaos/ek/v12/system"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type LoggerSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&LoggerSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *LoggerSuite) TestLogger(c *C) {
	tmpDir := c.MkDir()

	l := New(tmpDir, 0644)
	c.Assert(l, NotNil)

	err := l.Add("testing")
	c.Assert(err, IsNil)
	c.Assert(fsutil.IsExist(tmpDir+"/testing.log"), Equals, true)

	l.Get("testing").Print("TEST1")
	l.Get("testing").Print("TEST2")
	l.Flush()
	l.Get("unknown").Print("TEST3")

	c.Assert(fsutil.IsNonEmpty(tmpDir+"/testing.log"), Equals, true)
}

func (s *LoggerSuite) TestErrors(c *C) {
	l := New("/_unknown_", 0644)
	c.Assert(l, NotNil)

	err := l.Add("testing")
	c.Assert(err, NotNil)

	usernameCache = ""
	getUserFunc = func(avoidCache ...bool) (*system.User, error) {
		return nil, errors.New("ERROR")
	}

	c.Assert(getUserName(), Equals, "unknown")

	getUserFunc = system.CurrentUser
}
