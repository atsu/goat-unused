package build_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/atsu/goat/build"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"empty", "", "0"},
		{"no-path-separators", "foo", "0"},
		{"one-path-separator", sepAdjust("foo/bar"), "0"},
		{"no-v-in-version-number", "foo/bar/1", "0"},
		{"bad-version-number", "foo/bar/v1x", "0"},
		{"decimal-version-number", "foo/bar/v1.2", "0"},
		{"negative-version-number", "foo/bar/v-1", "0"},
		{"good-version-number", "foo/bar/v1", "1"},
		{"max-version-number", "foo/bar/v" + strconv.FormatUint(^uint64(0), 10), "18446744073709551615"},
		{"local-path", "/foo/v3", "3"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			expected := sepAdjust(test.expected)
			actual := sepAdjust(build.Module{Path: test.path}.MajorVersion())

			assert.Equal(t, expected, actual)
		})
	}
}

// Ensure tests can run on other architectures
func sepAdjust(path string) string {
	if '/' == os.PathSeparator {
		return path
	}
	return strings.Replace(path, "/", string(os.PathSeparator), -1)
}

func TestModInfo_String(t *testing.T) {
	json := `{"Module":{"Path":"github.com/atsu/iomkr"},"Require":[{"Path":"github.com/atsu/goat","Version":"v0.0.0-20181019194343-55bbe80b078a","Indirect":false},{"Path":"github.com/davecgh/go-spew","Version":"v1.1.1","Indirect":true}]}`
	mi := build.ToModInfoMust(json)
	assert.Equal(t, json, mi.String())
}
