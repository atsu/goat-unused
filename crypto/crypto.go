package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/scrypt"
)

func EncryptGCM(plainText, password []byte) ([]byte, error) {
	scryptPkg, err := NewScryptPkg(password)
	if err != nil {
		return nil, err
	}
	return encryptGCM(plainText, *scryptPkg, true)
}

func (r ScryptPkg) EncryptGCMBase64(plainText []byte) (string, error) {
	bytes, e := r.EncryptGCM(plainText)
	if e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (r ScryptPkg) EncryptGCM(plainText []byte) ([]byte, error) {
	return encryptGCM(plainText, r, false)
}

func encryptGCM(plainText []byte, scryptPkg ScryptPkg, addSalt bool) ([]byte, error) {
	aesGCM, err := createAESGCM(scryptPkg.Key)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	// Comment source is from: https://golang.org/src/crypto/cipher/example_test.go
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	var prefix []byte
	if addSalt {
		prefix = scryptPkg.Salt
		prefix = append(prefix, nonce...)
	} else {
		prefix = nonce
	}
	return aesGCM.Seal(prefix, nonce, plainText, nil), nil
}

// +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func DecryptGCM(data, password []byte) ([]byte, error) {
	dp := GetDefaultParams()
	salt := data[:dp.SaltLen]
	key, err := scrypt.Key([]byte(password), salt, dp.N, dp.R, dp.P, dp.DKLen)
	if err != nil {
		return nil, err
	}
	scryptPkg := ScryptPkg{Key: key, Salt: salt, Params: dp}
	return decryptGCM(data, scryptPkg, true)
}

func (r ScryptPkg) DecryptGCMBase64(dataBase64 string) ([]byte, error) {
	bytes, e := base64.StdEncoding.DecodeString(dataBase64)
	if e != nil {
		return nil, e
	}
	return r.DecryptGCM(bytes)
}

func (r ScryptPkg) DecryptGCM(data []byte) ([]byte, error) {
	return decryptGCM(data, r, false)
}

func decryptGCM(data []byte, r ScryptPkg, hasSalt bool) ([]byte, error) {
	aesGCM, err := createAESGCM(r.Key)
	if err != nil {
		return nil, err
	}

	var nonce, cipherText []byte
	nonceSize := aesGCM.NonceSize()

	if hasSalt {
		saltSize := len(r.Salt)
		nonce, cipherText = data[saltSize:saltSize+nonceSize], data[saltSize+nonceSize:]

	} else {
		nonce, cipherText = data[:nonceSize], data[nonceSize:]
	}

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}

func createAESGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aesGCM, nil
}
