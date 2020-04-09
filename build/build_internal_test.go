package build

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInfo(t *testing.T) {
	defer resetBuildInfo()

	version = "v0.18.10090159"
	commitHash = "86241aab92d35f9058fdddf4f3442cc9ba9f1a9d"
	date = "1539211300"
	agent = "travis-unit-test"
	initBuildInfo()

	info := GetInfo("goat")
	log.Println(info.Bannerf(KeyboardBannerFormat()))

	assert.Equal(t, "goat", info.Component)
	assert.Equal(t, "v0.18.10090159", info.Version)
	assert.Equal(t, "86241aab92d35f9058fdddf4f3442cc9ba9f1a9d", info.CommitHash)
	assert.Equal(t, int64(1539211300), info.ParsedDate.Unix())
	assert.Equal(t, "travis-unit-test", info.Agent)
	assert.Equal(t, info.Version, info.SemVer)

	fmt.Println(info.String())
}

func TestInfo_SemVer(t *testing.T) {
	defer resetBuildInfo()

	version = "86241aa-modified"
	initBuildInfo()
	info := GetInfo("goat")
	assert.Equal(t, "v0.0.0-86241aa-modified", info.SemVer)
}

func resetBuildInfo() {
	version = unknown
	commitHash = unknown
	date = unknown
	agent = unknown
	initBuildInfo()
}
