package data

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"strconv"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// PkgKeyIndex is map with packages keys for every supported arch
type PkgKeyIndex map[string]PkgKeyMap

// PkgKeyMap is map with packages keys
type PkgKeyMap map[int]bool

// ////////////////////////////////////////////////////////////////////////////////// //

// NewPkgKeyIndex creates new index
func NewPkgKeyIndex() PkgKeyIndex {
	return make(map[string]PkgKeyMap)
}

// NewPkgKeyMap creates new key map
func NewPkgKeyMap() PkgKeyMap {
	return make(map[int]bool)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Set adds key to map
func (m PkgKeyMap) Set(key int) {
	m[key] = true
}

// ////////////////////////////////////////////////////////////////////////////////// //

// HasData returns true if index contains any data
func (i PkgKeyIndex) HasData() bool {
	if i == nil {
		return false
	}

	for arch := range i {
		if i.HasArch(arch) {
			return true
		}
	}

	return false
}

// HasArch returns true if index contains keys for given arch
func (i PkgKeyIndex) HasArch(arch string) bool {
	return i[arch] != nil && len(i[arch]) != 0
}

// IgnoreArch returns true if given arch was added to index, but then dropped
func (i PkgKeyIndex) IgnoreArch(arch string) bool {
	keyMap, hadArch := i[arch]
	return keyMap == nil && hadArch
}

// Intersect finds intersection between two maps
func (i PkgKeyIndex) Intersect(arch string, src PkgKeyMap) {
	if i[arch] == nil {
		if !i.IgnoreArch(arch) {
			i[arch] = src
		}

		return
	}

	dst := i[arch]

	for k := range dst {
		if !src[k] {
			delete(dst, k)
		}
	}
}

// Drop removes all data for given arch from index
func (i PkgKeyIndex) Drop(arch string) {
	i[arch] = nil
}

// List converts key map to list with keys
func (i PkgKeyIndex) List(arch string) string {
	if i == nil || len(i[arch]) == 0 {
		return ""
	}

	var buf bytes.Buffer

	for k := range i[arch] {
		buf.WriteString(strconv.Itoa(k))
		buf.WriteRune(',')
	}

	return buf.String()[:buf.Len()-1]
}
