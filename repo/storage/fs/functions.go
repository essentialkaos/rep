package fs

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"database/sql"
	"fmt"
	"path"
	"strings"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// customFunc contains information about custom function
type customFunc struct {
	Name   string
	Impl   interface{}
	IsPure bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

var customFuncs = map[string][]*customFunc{}
var hasCustomDriver = map[string]bool{}

// ////////////////////////////////////////////////////////////////////////////////// //

// RegisterFunc registers new custom function for given DB type
// Notice that you can not register new functions after creating new storage
func RegisterFunc(dbType, name string, impl interface{}, pure bool) error {
	if len(hasCustomDriver) != 0 {
		return fmt.Errorf("Can't register new custom function after creating storage")
	}

	customFuncs[dbType] = append(customFuncs[dbType], &customFunc{
		Name: name, Impl: impl, IsPure: pure,
	})

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// registerDrivers registers drivers with all custom functions
func registerDrivers() {
	for dbType, funcs := range customFuncs {
		sql.Register("sqlite3_"+dbType,
			&sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					for _, f := range funcs {
						conn.RegisterFunc(f.Name, f.Impl, f.IsPure)
					}

					return nil
				},
			},
		)

		hasCustomDriver[dbType] = true
	}

	customFuncs = nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// filelistGlobberFunc is special function for checking glob
func filelistGlobberFunc(glob, dir, files string, isNegative int) bool {
	for _, file := range strings.Split(files, "/") {
		isMatch, _ := path.Match(glob, dir+"/"+file)

		switch {
		case isNegative == 1 && isMatch:
			return false
		case isNegative == 0 && isMatch:
			return true
		}
	}

	return isNegative == 1
}
