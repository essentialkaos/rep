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
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/essentialkaos/ek/v13/sortutil"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type packageSlice []*GroupPackage

func (s packageSlice) Len() int {
	return len(s)
}

func (s packageSlice) Less(i, j int) bool {
	if s[i].Type == s[j].Type {
		return sortutil.NaturalLess(s[i].Name, s[j].Name)
	} else {
		return packgageTypePriority[s[i].Type] > packgageTypePriority[s[j].Type]
	}
}

func (s packageSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type groupSlice []*Group

func (s groupSlice) Len() int {
	return len(s)
}

func (s groupSlice) Less(i, j int) bool {
	return sortutil.NaturalLess(s[i].ID, s[j].ID)
}

func (s groupSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type categorySlice []*Category

func (s categorySlice) Len() int {
	return len(s)
}

func (s categorySlice) Less(i, j int) bool {
	return sortutil.NaturalLess(s[i].ID, s[j].ID)
}

func (s categorySlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type envSlice []*Environment

func (s envSlice) Len() int {
	return len(s)
}

func (s envSlice) Less(i, j int) bool {
	return sortutil.NaturalLess(s[i].ID, s[j].ID)
}

func (s envSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type langpackSlice []*Langpack

func (s langpackSlice) Len() int {
	return len(s)
}

func (s langpackSlice) Less(i, j int) bool {
	return sortutil.NaturalLess(s[i].Name, s[j].Name)
}

func (s langpackSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// ////////////////////////////////////////////////////////////////////////////////// //

// packgageTypePriority contains priority for each package type
var packgageTypePriority = map[uint8]int{
	PKG_TYPE_CONDITIONAL: 0,
	PKG_TYPE_OPTIONAL:    1,
	PKG_TYPE_DEFAULT:     2,
	PKG_TYPE_MANDATORY:   3,
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Write writes information about groups, categories and environment to given file
func Write(comps *Comps, file string) error {
	return writeFile(comps, file, false)
}

// Write writes compressed information about groups, categories and environment to given file
func WriteGz(comps *Comps, file string) error {
	return writeFile(comps, file, true)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// writeFile writes comps file
func writeFile(comps *Comps, file string, compressed bool) error {
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}

	defer fd.Close()

	var ww io.Writer

	w := bufio.NewWriter(fd)

	if compressed {
		ww = gzip.NewWriter(w)
	} else {
		ww = w
	}

	writeCompsXML(ww, comps)

	return w.Flush()
}

// writeCompsXML writes XML data to given writer
func writeCompsXML(w io.Writer, comps *Comps) {
	writeString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	writeString(w, `<!DOCTYPE comps PUBLIC "-//CentOS//DTD Comps info//EN" "comps.dtd">`+"\n")
	writeString(w, "<comps>\n")

	sort.Sort(groupSlice(comps.Groups))

	for _, group := range comps.Groups {
		writeGroupXML(w, group)
	}

	sort.Sort(categorySlice(comps.Categories))

	for _, category := range comps.Categories {
		writeCategoryXML(w, category)
	}

	sort.Sort(envSlice(comps.Environments))

	for _, env := range comps.Environments {
		writeEnvXML(w, env)
	}

	sort.Sort(langpackSlice(comps.Langpacks))

	writeLangpackXML(w, comps.Langpacks)

	writeString(w, "</comps>\n")
}

// writeGroupXML writes XML node with group info to given writer
func writeGroupXML(w io.Writer, group *Group) {
	sort.Sort(packageSlice(group.Packages))

	writeString(w, "  <group>\n")
	writeString(w, fmt.Sprintf("    <id>%s</id>\n", group.ID))

	writeLocStringXML(w, "name", group.Name)
	writeLocStringXML(w, "description", group.Description)

	writeString(w, fmt.Sprintf("    <default>%t</default>\n", group.IsDefault))
	writeString(w, fmt.Sprintf("    <uservisible>%t</uservisible>\n", group.IsVisible))
	writeString(w, "    <packagelist>\n")

	for _, pkg := range group.Packages {
		writeString(w, renderPackageInfo(pkg))
	}

	writeString(w, "    </packagelist>\n")
	writeString(w, "  </group>\n")
}

// writeCategoryXML writes XML node with category info to given writer
func writeCategoryXML(w io.Writer, category *Category) {
	writeString(w, "  <category>\n")
	writeString(w, fmt.Sprintf("    <id>%s</id>\n", category.ID))

	writeLocStringXML(w, "name", category.Name)
	writeLocStringXML(w, "description", category.Description)

	writeString(w, "    <grouplist>\n")

	sort.Strings(category.Groups)

	for _, id := range category.Groups {
		writeString(w, fmt.Sprintf("      <groupid>%s</groupid>\n", id))
	}

	writeString(w, "    </grouplist>\n")
	writeString(w, "  </category>\n")
}

// writeCategoryXML writes XML node with environment info to given writer
func writeEnvXML(w io.Writer, env *Environment) {
	writeString(w, "  <environment>\n")
	writeString(w, fmt.Sprintf("    <id>%s</id>\n", env.ID))

	writeLocStringXML(w, "name", env.Name)
	writeLocStringXML(w, "description", env.Description)

	writeString(w, fmt.Sprintf("    <display_order>%d</display_order>\n", env.DisplayOrder))
	writeString(w, "    <grouplist>\n")

	sort.Strings(env.Groups)

	for _, id := range env.Groups {
		writeString(w, fmt.Sprintf("      <groupid>%s</groupid>\n", id))
	}

	writeString(w, "    </grouplist>\n")
	writeString(w, "    <optionlist>\n")

	sort.Strings(env.Options)

	for _, id := range env.Options {
		writeString(w, fmt.Sprintf("      <groupid>%s</groupid>\n", id))
	}

	writeString(w, "    </optionlist>\n")
	writeString(w, "  </environment>\n")
}

// writeCategoryXML writes XML node with langpack info to given writer
func writeLangpackXML(w io.Writer, packs []*Langpack) {
	if len(packs) != 0 {
		writeString(w, "  <langpacks>\n")

		for _, pack := range packs {
			writeString(w, fmt.Sprintf("    <match install=\"%s\" name=\"%s\"/>\n", pack.Install, pack.Name))
		}

		writeString(w, "  </langpacks>\n")
	}
}

// writeLocString writes XML nodes with localized data
func writeLocStringXML(w io.Writer, name string, data *LocString) {
	writeString(w, fmt.Sprintf("    <%s>%s</%s>\n", name, data, name))

	if data.Num() != 0 {
		for _, l := range getLocStringLangs(data) {
			writeString(w, fmt.Sprintf("    <%s xml:lang=\"%s\">%s</%s>\n", name, l, data.Get(l), name))
		}
	}
}

// renderPackageInfo renders XML node with package info
func renderPackageInfo(pkg *GroupPackage) string {
	var attrs string

	switch pkg.Type {
	case PKG_TYPE_DEFAULT:
		attrs = `type="default"`
	case PKG_TYPE_OPTIONAL:
		attrs = `type="optional"`
	case PKG_TYPE_MANDATORY:
		attrs = `type="mandatory"`
	case PKG_TYPE_CONDITIONAL:
		attrs = `type="conditional"`
	}

	if pkg.Requires != "" {
		attrs += " requires=\"" + pkg.Requires + "\""
	}

	if len(pkg.Arch) != 0 {
		attrs += " arch=\"" + strings.Join(pkg.Arch, ",") + "\""
	}

	if pkg.BaseArchOnly {
		attrs += ` basearchonly="true"`
	}

	return fmt.Sprintf("      <packagereq %s>%s</packagereq>\n", attrs, pkg.Name)
}

// getLocStringLangs returns slice with sorted language names
func getLocStringLangs(l *LocString) []string {
	var result []string

	for lang := range l.Localized {
		result = append(result, lang)
	}

	sort.Strings(result)

	return result
}

// writeString writes string to writer
func writeString(w io.Writer, data string) {
	w.Write([]byte(data))
}
