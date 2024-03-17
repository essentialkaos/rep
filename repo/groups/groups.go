package groups

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2024 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/xml"
	"strings"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	PKG_TYPE_CONDITIONAL uint8 = 0
	PKG_TYPE_OPTIONAL    uint8 = 1
	PKG_TYPE_DEFAULT     uint8 = 2
	PKG_TYPE_MANDATORY   uint8 = 3
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Comps contains info about groups
type Comps struct {
	Groups       []*Group       `xml:"group"`
	Categories   []*Category    `xml:"category"`
	Environments []*Environment `xml:"environment"`
	Langpacks    []*Langpack    `xml:"langpacks>match"`
}

// Group contains info about package group
type Group struct {
	ID          string          `xml:"id"`
	Name        *LocString      `xml:"name"`
	IsDefault   bool            `xml:"default"`
	IsVisible   bool            `xml:"uservisible"`
	Description *LocString      `xml:"description"`
	Packages    []*GroupPackage `xml:"packagelist>packagereq"`
}

// Category contains info about category
type Category struct {
	ID          string     `xml:"id"`
	Name        *LocString `xml:"name"`
	Description *LocString `xml:"description"`
	Groups      []string   `xml:"grouplist>groupid"`
}

// Environment contains environment info
type Environment struct {
	ID           string     `xml:"id"`
	Name         *LocString `xml:"name"`
	Description  *LocString `xml:"description"`
	DisplayOrder int        `xml:"display_order"`
	Groups       []string   `xml:"grouplist>groupid"`
	Options      []string   `xml:"optionlist>groupid"`
}

// GroupPackage contains info about package in group
type GroupPackage struct {
	Type         uint8
	BaseArchOnly bool
	Arch         []string
	Name         string
	Requires     string
}

// Langpack contains info about language pack
type Langpack struct {
	Install string `xml:"install,attr"`
	Name    string `xml:"name,attr"`
}

// LocString is collection of localized strings
type LocString struct {
	Default   string
	Localized map[string]string
}

// ////////////////////////////////////////////////////////////////////////////////// //

// UnmarshalXML is custom unmarshaller for localized info
func (s *LocString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	v := struct {
		Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
		Value string `xml:",chardata"`
	}{}

	err := d.DecodeElement(&v, &start)

	if err != nil {
		return err
	}

	if v.Lang == "" {
		s.Default = v.Value
	} else {
		if s.Localized == nil {
			s.Localized = make(map[string]string)
		}

		s.Localized[v.Lang] = v.Value
	}

	return nil
}

// UnmarshalXML is custom unmarshaller for group info
func (p *GroupPackage) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	v := struct {
		Arch         string `xml:"arch,attr"`
		BaseArchOnly bool   `xml:"basearchonly,attr"`
		Name         string `xml:",chardata"`
		Requires     string `xml:"requires,attr"`
		Type         string `xml:"type,attr"`
	}{}

	err := d.DecodeElement(&v, &start)

	if err != nil {
		return err
	}

	p.Name = v.Name
	p.Requires = v.Requires
	p.BaseArchOnly = v.BaseArchOnly

	switch strings.ToLower(v.Type) {
	case "conditional":
		p.Type = PKG_TYPE_CONDITIONAL
	case "mandatory":
		p.Type = PKG_TYPE_MANDATORY
	case "default":
		p.Type = PKG_TYPE_DEFAULT
	case "optional":
		p.Type = PKG_TYPE_OPTIONAL
	}

	if v.Arch != "" {
		p.Arch = strings.Split(v.Arch, ",")
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// String returns default (english) value
func (s *LocString) String() string {
	if s == nil {
		return ""
	}

	return s.Default
}

// Get returns value on given language
func (s *LocString) Get(lang string) string {
	if s == nil || s.Localized == nil {
		return ""
	}

	return s.Localized[lang]
}

// Num returns number of localized strings
func (s *LocString) Num() int {
	if s == nil {
		return 0
	}

	return len(s.Localized)
}
