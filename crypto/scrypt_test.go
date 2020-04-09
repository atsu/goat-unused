package crypto_test

import (
	"strings"
	"testing"

	"github.com/atsu/goat/crypto"
	"github.com/stretchr/testify/assert"
)

func TestMakeScryptPkg(t *testing.T) {
	tests := []struct {
		name            string
		password        string
		comparePassword string
		matchExpected   bool
	}{
		{"isMatch", "foobar", "foobar", true},
		{"isNotMatch", "foobar", "moobaa", false},
	}

	dp := crypto.GetDefaultParams()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scryptPkg, err := crypto.NewScryptPkg([]byte(test.password))
			assert.NoError(t, err)
			assert.Equal(t, dp.DKLen, len(scryptPkg.Key))
			assert.Equal(t, dp.SaltLen, len(scryptPkg.Salt))
			assert.Equal(t, dp, scryptPkg.Params)

			encodedScrypt := scryptPkg.Encode()
			assert.Equal(t, 1, strings.Count(string(encodedScrypt), "$"))

			ok, e := encodedScrypt.CompareHashAndPassword([]byte(test.comparePassword))
			assert.NoError(t, e)
			if test.matchExpected {
				assert.True(t, ok)
			} else {
				assert.False(t, ok)
			}

			decodedScrypt, err := encodedScrypt.Decode()
			assert.NoError(t, err)
			assert.Equal(t, scryptPkg, decodedScrypt)
		})
	}
}
