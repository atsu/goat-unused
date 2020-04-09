package build

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
)

// todo more fields, Exclude, Replace, eventually someone will make this available in the community
type ModInfo struct {
	Module  Module
	Require []Require
}

type Module struct {
	Path string
}

type Require struct {
	Path     string
	Version  string
	Indirect bool
}

func ToModInfoMust(text string) ModInfo {
	mi := ModInfo{}
	if e := json.Unmarshal([]byte(text), &mi); e != nil {
		panic(e)
	}
	return mi
}

func (r ModInfo) String() string {
	b, e := json.Marshal(r)
	if e != nil {
		panic(e)
	}
	return string(b)
}

func (r Module) MajorVersion() string {
	_, lastSegment := filepath.Split(r.Path)
	if !strings.HasPrefix(lastSegment, "v") || len(lastSegment) == 1 {
		return "0"
	}

	numericPart := lastSegment[1:]
	if _, e := strconv.ParseUint(numericPart, 10, 64); e != nil {
		return "0"
	}
	return numericPart
}
