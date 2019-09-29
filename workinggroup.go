package main

import (
	"errors"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
)

var (
	errNoSuchGroup = errors.New("no such group")
)

type workingGroup interface {
	Group() *gokeepasslib.Group
	Prev() workingGroup
	ChGroup(name string) (workingGroup, error)
	String() string
}

type rootGroup struct {
	db *gokeepasslib.Database
}

func (rg *rootGroup) Group() *gokeepasslib.Group {
	return &rg.db.Content.Root.Groups[0]
}

func (rg *rootGroup) Prev() workingGroup {
	return nil
}

func (rg *rootGroup) ChGroup(name string) (workingGroup, error) {
	var g *gokeepasslib.Group
	for _, subGroup := range rg.Group().Groups {
		if subGroup.Name == name {
			g = &subGroup
			break
		}
	}
	if g == nil {
		return rg, errNoSuchGroup
	}

	return &subGroup{g, "/" + name, rg}, nil
}

func (rg *rootGroup) String() string {
	return "/"
}

func newRootGroup(db *gokeepasslib.Database) workingGroup {
	return &rootGroup{db}
}

type subGroup struct {
	workingGroup *gokeepasslib.Group
	path         string
	prev         workingGroup
}

func (sg *subGroup) Group() *gokeepasslib.Group {
	return sg.workingGroup
}

func (sg *subGroup) Prev() workingGroup {
	return sg.prev
}

func (sg *subGroup) ChGroup(name string) (workingGroup, error) {
	var g *gokeepasslib.Group
	for _, subGroup := range sg.Group().Groups {
		if subGroup.Name == name {
			g = &subGroup
			break
		}
	}
	if g == nil {
		return sg, errNoSuchGroup
	}

	return &subGroup{g, sg.path + "/" + name, sg}, nil
}

func (sg *subGroup) String() string {
	return sg.path
}

func travel(cwd workingGroup, path string) (workingGroup, error) {
	var err error
	parts := strings.Split(path, "/")
	for i := 0; err == nil && cwd != nil && i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}

		part := parts[i]
		if part == ".." {
			if cwd.Prev() != nil {
				cwd = cwd.Prev()
			}
		} else if part != "." {
			cwd, err = cwd.ChGroup(part)
		}
	}

	return cwd, err
}
