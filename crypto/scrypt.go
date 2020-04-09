package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

// For package usage, see docs in crypto/doc.go

// Params describes the input parameters to the scrypt
// key derivation function as per Colin Percival's scrypt
// paper: http://www.tarsnap.com/scrypt/scrypt.pdf
// Also see: https://en.wikipedia.org/wiki/Scrypt

type ScryptParams struct {
	N       int // CPU/memory cost parameter (logN)
	R       int // block size parameter (octets)
	P       int // parallelization parameter (positive int)
	SaltLen int // bytes to use as salt (octets)
	DKLen   int // length of the derived key (octets)
}

func GetDefaultParams() ScryptParams {
	return ScryptParams{N: 32768, R: 8, P: 1, SaltLen: 16, DKLen: 32}
}

type ScryptPkg struct {
	Key    []byte
	Salt   []byte
	Params ScryptParams
}

type EncodedScryptPkg string

func NewScryptPkg(password []byte) (*ScryptPkg, error) {
	return NewScryptPkgWithParams(password, GetDefaultParams())
}

func FromEncodedMust(esp string) *ScryptPkg {
	pkg, e := EncodedScryptPkg(esp).Decode()
	if e != nil {
		panic(e)
	}
	return pkg
}

func NewScryptPkgWithParams(password []byte, params ScryptParams) (*ScryptPkg, error) {
	salt := make([]byte, params.SaltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	key, err := scrypt.Key([]byte(password), salt, params.N, params.R, params.P, params.DKLen)
	if err != nil {
		return nil, err
	}
	return &ScryptPkg{Key: key, Salt: salt, Params: params}, nil
}

func (r ScryptPkg) Encode() EncodedScryptPkg {
	encodedSalt := base64.StdEncoding.EncodeToString(r.Salt)
	encodedKey := base64.StdEncoding.EncodeToString(r.Key)
	return EncodedScryptPkg(fmt.Sprintf("%s$%s", encodedSalt, encodedKey))
}

func (r ScryptPkg) CompareHashAndPasswordWithParams(password []byte) (bool, error) {
	other, err := scrypt.Key(password, r.Salt, r.Params.N, r.Params.R, r.Params.P, r.Params.DKLen)
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare(r.Key, other) == 1 {
		return true, nil
	}
	return false, nil
}

var ErrInvalidHash = errors.New("scrypt: the provided hash is not in the correct format")

func (r EncodedScryptPkg) Decode() (*ScryptPkg, error) {
	values := strings.Split(string(r), "$")
	if len(values) != 2 {
		return nil, ErrInvalidHash
	}

	salt, err := base64.StdEncoding.DecodeString(values[0])
	if err != nil {
		return nil, ErrInvalidHash
	}

	key, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return nil, ErrInvalidHash
	}
	return &ScryptPkg{Key: key, Salt: salt, Params: GetDefaultParams()}, nil
}

func (r EncodedScryptPkg) CompareHashAndPassword(password []byte) (bool, error) {
	scryptPkg, e := r.Decode()
	if e != nil {
		return false, e
	}
	return scryptPkg.CompareHashAndPasswordWithParams(password)
}
