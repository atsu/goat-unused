package build_test

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/atsu/goat/build"
	"github.com/stretchr/testify/assert"
)

func TestInitDefaults(t *testing.T) {
	componentName := "goat"
	info := build.GetInfo(componentName)

	actualBanner := info.Banner()
	log.Println(actualBanner)

	const unknown = "unknown"
	assert.Equal(t, componentName, info.Component)
	assert.Equal(t, unknown, info.Version)
	assert.Equal(t, unknown, info.CommitHash)
	assert.Equal(t, unknown, info.Date)
	assert.Equal(t, unknown, info.Agent)

	assert.Equal(t, fmt.Sprintf("%s-%s %s %s", unknown, unknown, unknown, unknown), build.LDFlags())

	assert.True(t, strings.Contains(actualBanner, "component: "+componentName))
	assert.True(t, strings.Contains(actualBanner, "version: "+unknown))
	assert.True(t, strings.Contains(actualBanner, "build: "+unknown))
	assert.True(t, strings.Contains(actualBanner, "date: "+unknown))
	assert.True(t, strings.Contains(actualBanner, "agent: "+unknown))
}

func TestInfo_String(t *testing.T) {
	info := build.GetInfo("goat")

	serialized := info.String()

	ds := build.ToInfoMust(serialized)

	// The extra three lines are to get around
	// challenges comparing time.
	// https://golang.org/pkg/time/#Time
	assert.True(t, ds.ParsedDate.Equal(info.ParsedDate))

	ds.ParsedDate = time.Time{}
	info.ParsedDate = time.Time{}

	assert.Equal(t, info, ds)
}
