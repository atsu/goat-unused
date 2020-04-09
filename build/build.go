package build

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const unknown = "unknown"

// For usage see build/doc.go
var (
	version    = unknown
	commitHash = unknown
	date       = unknown
	agent      = unknown
)

type Info struct {
	Component  string `json:"component"`
	Version    string `json:"version"`
	CommitHash string `json:"commit"`
	Date       string `json:"builddate"`
	Agent      string `json:"agent"`
	SemVer     string    `json:"semver,omitempty"`
	ParsedDate time.Time `json:"parseddate,omitempty"`
}

type info struct {
	Info

	mux sync.Mutex
}

var buildInfo info

func init() {
	initBuildInfo()
}

// Use a discrete init function here so unit tests can access it.
func initBuildInfo() {
	i := Info{Version: version, CommitHash: commitHash, Date: date, Agent: agent}
	buildInfo = info{Info: i}
}

func GetInfo(component string) Info { // return by value to prevent inadvertent field changes
	buildInfo.mux.Lock()
	defer buildInfo.mux.Unlock()

	info := &buildInfo.Info

	if info.Component != "" {
		return buildInfo.Info
	}

	info.Component = component

	// If Version matches a semantic-version format then return Version as is,
	// otherwise prepend a "v0.0.0-" to Version
	if !strings.HasPrefix(info.Version, "v") || strings.Count(info.Version, ".") != 2 {
		info.SemVer = fmt.Sprintf("v0.0.0-%s", info.Version)
	} else {
		info.SemVer = info.Version
	}
	// Attempt to parse the Date
	epochSecs, _ := strconv.ParseInt(info.Date, 10, 64)
	info.ParsedDate = time.Unix(epochSecs, 0)

	return buildInfo.Info
}

func LDFlags() string {
	return fmt.Sprintf("%s-%s %v %s", version, commitHash, date, agent)
}

func (r Info) String() string {
	b, e := json.Marshal(r)
	if e != nil {
		log.Panic(e)
	}
	return string(b)
}

func ToInfoMust(text string) Info {
	i := Info{}
	e := json.Unmarshal([]byte(text), &i)
	if e != nil {
		log.Panic(e)
	}

	return i
}

// Banner if you want it!
func (r Info) Banner() string {
	return r.Bannerf(fireBannerf)
}

// Custom banner if you want it!
func (r Info) Bannerf(f string) string {
	var dateText string

	if r.ParsedDate.Unix() == 0 {
		dateText = r.Date
	} else {
		dateText = fmt.Sprintf("%v (%s)", r.Date, fmt.Sprintf(r.ParsedDate.In(time.Local).Format("Jan 02 2006 15:04:05 MST")))
	}

	v := r.Version
	if v != r.SemVer {
		v = fmt.Sprintf("%s (%s)", v, r.SemVer)
	}

	return fmt.Sprintf(f, r.Component, v, r.CommitHash, dateText, r.Agent)
}

var fireBannerf = `
🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥
    _  _/   _       _     _ 
   (/  /  _)  (/   (  () /  /)
                           /
🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥
   🔥 component: %s
   🔥   version: %s
   🔥     build: %s
   🔥      date: %s
   🔥     agent: %s
🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥🔥`

// For Seth
func KeyboardBannerFormat() string {
	return keyboardBannerf
}

var keyboardBannerf = `
 ____ ____ ____ ____ ____ __________ ____ ____ ____ ____ ____ 
||🔥|||a |||t |||s |||u |||    •   |||c |||o |||r |||p |||🔥||
||__|||__|||__|||__|||__|||________|||__|||__|||__|||__|||__||
|/__\|/__\|/__\|/__\|/__\|/________\|/__\|/__\|/__\|/__\|/__\|
  ䷝ component: %s
  ䷝   version: %s 
  ䷝     build: %s
  ䷝      date: %s
  ䷝     agent: %s`
