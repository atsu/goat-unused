package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func EncodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func DecodeBase64(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]string, 0, length)
	for i := 0; i < length; i++ {
		idx := rand.Intn(len(charSet))
		result = append(result, string(charSet[idx]))
	}
	return strings.Join(result, "")
}

func Int32InSlice(slice []int32, val int32) bool {
	if slice == nil {
		return false
	}
	for _, item := range slice {
		if val == item {
			return true
		}
	}
	return false
}

func StringInSlice(slice []string, str string) bool {
	if slice == nil {
		return false
	}
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// MarshalWithPretty for use in handlers, check for the 'pretty' query parameter
// if the pretty parameter is true, pretty json will be returned.
func MarshalWithPretty(r *http.Request, obj interface{}) []byte {
	p := r.URL.Query().Get("pretty")
	out, _ := json.Marshal(obj)
	if pretty, _ := strconv.ParseBool(p); pretty {
		out = JsonPrettyPrint(out)
	}
	return out
}

// JsonPrettyPrint attempts to reformat a json byte array to be pretty.
// if any error occurs, just fall back to the original input.
func JsonPrettyPrint(in []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, in, "", "\t")
	if err != nil {
		return in
	}
	return out.Bytes()
}
