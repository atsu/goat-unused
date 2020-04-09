package crypto_test

import (
	"testing"

	"github.com/atsu/goat/crypto"
	"github.com/stretchr/testify/assert"
)

func TestEncryptGCM(t *testing.T) {
	password := []byte("password123")
	scryptPkg, e := crypto.NewScryptPkg(password)
	assert.NoError(t, e)

	encodedScryptPkg := scryptPkg.Encode()
	decodedPkg, e := encodedScryptPkg.Decode()
	assert.NoError(t, e)
	assert.Equal(t, decodedPkg, crypto.FromEncodedMust(string(encodedScryptPkg)))

	plainText := []byte("a quick brown goat jumps over a squirrel")
	data, e := decodedPkg.EncryptGCM(plainText)
	assert.NoError(t, e)

	decrypted, e := decodedPkg.DecryptGCM(data)
	assert.NoError(t, e)
	assert.Equal(t, plainText, decrypted)

	plainText = []byte("goats, muts, squirrels, a menagerie!")
	base64Data, e := decodedPkg.EncryptGCMBase64(plainText)
	assert.NoError(t, e)

	decrypted, e = decodedPkg.DecryptGCMBase64(base64Data)
	assert.NoError(t, e)
	assert.Equal(t, plainText, decrypted)
}

func TestEncryptAndDecryptUsingPlaintextPassword(t *testing.T) {
	password := []byte("912734klweo712340-`lmwer	puqu😀")
	plainText := []byte("a quick brown goat jumps over a squirrel and a rat")

	data, e := crypto.EncryptGCM(plainText, password)
	assert.NoError(t, e)

	decrypted, e := crypto.DecryptGCM(data, password)
	assert.NoError(t, e)
	assert.Equal(t, plainText, decrypted)
}
