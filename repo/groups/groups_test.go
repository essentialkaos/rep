package groups

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2025 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/xml"
	"os"
	"sort"
	"strings"
	"testing"

	. "github.com/essentialkaos/check"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type GroupsSuite struct {
	TmpDir string
}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&GroupsSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *GroupsSuite) SetUpSuite(c *C) {
	s.TmpDir = c.MkDir()
}

func (s *GroupsSuite) TestReadingErrors(c *C) {
	comps, err := Read("unknown.xml")

	c.Assert(err, NotNil)
	c.Assert(comps, IsNil)

	err = os.WriteFile(s.TmpDir+"/test.xml", []byte("TEST:TEST"), 0600)

	c.Assert(err, IsNil)

	comps, err = Read(s.TmpDir + "/test.xml")

	c.Assert(err, NotNil)
	c.Assert(comps, IsNil)

	err = Write(&Comps{}, "/UNKNOWN/comps2.xml")

	c.Assert(err, NotNil)
}

func (s *GroupsSuite) TestReading(c *C) {
	comps, err := Read("../../testdata/comps.xml")

	c.Assert(err, IsNil)
	c.Assert(comps, NotNil)

	c.Assert(comps.Groups, HasLen, 2)
	c.Assert(comps.Categories, HasLen, 2)
	c.Assert(comps.Environments, HasLen, 2)
	c.Assert(comps.Langpacks, HasLen, 3)

	c.Assert(comps.Groups[0].ID, Equals, "additional-devel")
	c.Assert(comps.Groups[0].Name.String(), Equals, "Additional Development")
	c.Assert(comps.Groups[0].Name.Get("es"), Equals, "Desarrollo adicional")
	c.Assert(comps.Groups[0].Description.String(), Equals, "Additional development headers and libraries for building open-source applications.")
	c.Assert(comps.Groups[0].Description.Get("pl"), Equals, "Dodatkowe nagłówki i biblioteki do rozwijania aplikacji open source.")

	c.Assert(comps.Groups[0].IsDefault, Equals, true)
	c.Assert(comps.Groups[0].IsVisible, Equals, true)

	c.Assert(comps.Groups[0].Packages, HasLen, 5)
	c.Assert(comps.Groups[0].Packages[0].Type, Equals, uint8(3))
	c.Assert(comps.Groups[0].Packages[1].Type, Equals, uint8(2))
	c.Assert(comps.Groups[0].Packages[2].Type, Equals, uint8(1))
	c.Assert(comps.Groups[0].Packages[4].Type, Equals, uint8(0))

	c.Assert(comps.Groups[0].Packages[0].Arch, DeepEquals, []string{"i686", "s390"})
	c.Assert(comps.Groups[0].Packages[0].BaseArchOnly, Equals, false)
	c.Assert(comps.Groups[0].Packages[1].BaseArchOnly, Equals, true)
	c.Assert(comps.Groups[0].Packages[4].Requires, Equals, "ruby")
	c.Assert(comps.Groups[0].Packages[4].Name, Equals, "rubygem-abrt")

	c.Assert(comps.Categories[0].ID, Equals, "applications")
	c.Assert(comps.Categories[0].Name.String(), Equals, "Applications")
	c.Assert(comps.Categories[0].Name.Get("ru"), Equals, "Приложения")
	c.Assert(comps.Categories[0].Description.String(), Equals, "End-user applications.")
	c.Assert(comps.Categories[0].Description.Get("ru"), Equals, "Приложения пользователя.")
	c.Assert(comps.Categories[0].Groups, DeepEquals, []string{"emacs", "gnome-apps", "graphics"})

	c.Assert(comps.Environments[1].ID, Equals, "minimal")
	c.Assert(comps.Environments[1].Name.String(), Equals, "Minimal Install")
	c.Assert(comps.Environments[1].Name.Get("it"), Equals, "Installazione minima")
	c.Assert(comps.Environments[1].Description.String(), Equals, "Basic functionality.")
	c.Assert(comps.Environments[1].Description.Get("zh_CN"), Equals, "基本功能。")
	c.Assert(comps.Environments[1].Description.Get("zh_TW"), Equals, "基本功能。")
	c.Assert(comps.Environments[1].DisplayOrder, Equals, 23)
	c.Assert(comps.Environments[1].Groups, HasLen, 4)
	c.Assert(comps.Environments[1].Options, HasLen, 3)
	c.Assert(comps.Environments[1].Groups, DeepEquals, []string{"base", "core", "virtualization-hypervisor", "virtualization-tools"})
	c.Assert(comps.Environments[1].Options, DeepEquals, []string{"debugging", "network-file-system-client", "remote-system-management"})

	c.Assert(comps.Langpacks[1].Install, Equals, "firefox-langpack-%s")
	c.Assert(comps.Langpacks[1].Name, Equals, "firefox")

	comps, err = ReadGz("../../testdata/comps.xml.gz")

	c.Assert(err, IsNil)
	c.Assert(comps, NotNil)
}

func (s *GroupsSuite) TestWriting(c *C) {
	comps, err := Read("../../testdata/comps.xml")

	c.Assert(err, IsNil)
	c.Assert(comps, NotNil)

	err = Write(comps, s.TmpDir+"/comps2.xml")

	c.Assert(err, IsNil)

	err = WriteGz(comps, s.TmpDir+"/comps2.xml.gz")

	c.Assert(err, IsNil)

	comps2, err := Read(s.TmpDir + "/comps2.xml")

	c.Assert(err, IsNil)
	c.Assert(comps2, NotNil)

	c.Assert(comps2, DeepEquals, comps)
}

func (s *GroupsSuite) TestDataSorting(c *C) {
	pkgs := []*GroupPackage{
		{Type: 1, Name: "packageA"},
		{Type: 3, Name: "packageB"},
		{Type: 2, Name: "package10"},
		{Type: 2, Name: "package2"},
		{Type: 0, Name: "packageX"},
	}

	groups := []*Group{
		{ID: "test10"},
		{ID: "test5"},
	}

	categories := []*Category{
		{ID: "test10"},
		{ID: "test5"},
	}

	envs := []*Environment{
		{ID: "test10"},
		{ID: "test5"},
	}

	langpacks := []*Langpack{
		{Name: "test10"},
		{Name: "test5"},
	}

	sort.Sort(packageSlice(pkgs))
	sort.Sort(groupSlice(groups))
	sort.Sort(categorySlice(categories))
	sort.Sort(envSlice(envs))
	sort.Sort(langpackSlice(langpacks))

	c.Assert(pkgs[1].Name, Equals, "package2")
	c.Assert(groups[0].ID, Equals, "test5")
	c.Assert(categories[0].ID, Equals, "test5")
	c.Assert(envs[0].ID, Equals, "test5")
	c.Assert(langpacks[0].Name, Equals, "test5")
}

func (s *GroupsSuite) TestAUX(c *C) {
	var l *LocString

	c.Assert(l.String(), Equals, "")
	c.Assert(l.Get("ru"), Equals, "")
	c.Assert(l.Num(), Equals, 0)
}

func (s *GroupsSuite) TestCustomXMLDecoders(c *C) {
	r := strings.NewReader("")
	d := xml.NewDecoder(r)

	l := &LocString{}
	p := &GroupPackage{}

	c.Assert(l.UnmarshalXML(d, xml.StartElement{}), NotNil)
	c.Assert(p.UnmarshalXML(d, xml.StartElement{}), NotNil)
}
