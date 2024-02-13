package groups

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"io"
	"os"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Read reads and parses comps.xml file
func Read(file string) (*Comps, error) {
	return readFile(file, false)
}

// ReadGz reads, uncopresses and parses comps.xml.gz file
func ReadGz(file string) (*Comps, error) {
	return readFile(file, true)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// readFile reads comps file
func readFile(file string, compressed bool) (*Comps, error) {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	var rr io.Reader

	r := bufio.NewReader(fd)

	if compressed {
		rr, _ = gzip.NewReader(r)
	} else {
		rr = r
	}

	d := xml.NewDecoder(rr)

	result := &Comps{}
	err = d.Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}
