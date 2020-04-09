package util_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/atsu/goat/util"
	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name    string
		slice   []string
		str     string
		inslice bool
	}{
		{"only_one", []string{"one"}, "one", true},
		{"first", []string{"one", "two"}, "one", true},
		{"last", []string{"a", "b", "c", "one"}, "one", true},
		{"empty slice", []string{}, "one", false},
		{"empty str", []string{"test"}, "", false},
		{"both empty", []string{}, "", false},
		{"nil slice", nil, "", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := util.StringInSlice(test.slice, test.str)
			assert.Equal(t, test.inslice, result)
		})
	}
}

func TestInt32InSlice(t *testing.T) {
	tests := []struct {
		name    string
		slice   []int32
		integer int32
		inslice bool
	}{
		{"only_one", []int32{1}, 1, true},
		{"first", []int32{1, 2}, 1, true},
		{"last", []int32{0, 4, 7, 10}, 10, true},
		{"empty slice", []int32{}, 1, false},
		{"nil slice", nil, 7, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := util.Int32InSlice(test.slice, test.integer)
			assert.Equal(t, test.inslice, result)
		})
	}
}

func TestRandomString(t *testing.T) {
	for i := 0; i < 100; i++ {
		lastValue := ""
		str := util.RandomString(i)
		assert.Len(t, str, i)
		if i > 0 {
			assert.NotEmpty(t, str)
			assert.NotEqual(t, lastValue, str)
		} else {
			assert.Empty(t, str)
		}
		lastValue = str
	}
}

func TestBase64Encoding(t *testing.T) {
	str := "test value... atsu atsu atsu 🔥🔥🔥"
	encoded := util.EncodeBase64(str)
	decoded, err := util.DecodeBase64(encoded)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, str, string(decoded))
}

func TestMarshalWithPretty(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		pretty string
		want   string
	}{
		{"not pretty", struct {
			One string `json:"one"`
		}{
			"two",
		}, "", `{"one":"two"}`},
		{"not pretty-false", struct {
			One string `json:"one"`
		}{
			"two",
		}, "false", `{"one":"two"}`},
		{"not pretty-0", struct {
			One string `json:"one"`
		}{
			"two",
		}, "0", `{"one":"two"}`},
		{"pretty-true", struct {
			One string `json:"one"`
		}{
			"two",
		}, "true", "{\n\t\"one\": \"two\"\n}"},
		{"pretty-true", struct {
			One string `json:"one"`
		}{
			"two",
		}, "true", "{\n\t\"one\": \"two\"\n}"},
		{"pretty-1", struct {
			One string `json:"one"`
		}{
			"two",
		}, "true", "{\n\t\"one\": \"two\"\n}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &http.Request{
				URL: &url.URL{
					RawQuery: fmt.Sprintf("pretty=%s", test.pretty),
				},
			}
			got := util.MarshalWithPretty(r, test.input)
			assert.Equal(t, test.want, string(got))
		})
	}
}

func TestJsonPrettyPrint(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"pretty",
			`{"one":"two","map":{"key":"val"},"ary":["a","b"]}`,
			"{\n\t\"one\": \"two\",\n\t\"map\": {\n\t\t\"key\": \"val\"\n\t},\n\t\"ary\": [\n\t\t\"a\",\n\t\t\"b\"\n\t]\n}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := util.JsonPrettyPrint([]byte(test.input))
			assert.Equal(t, test.want, string(got))

			// For visual inspection
			fmt.Printf("input:\n%v\n", test.input)
			fmt.Printf("output:\n%v\n", string(got))
		})
	}
}
