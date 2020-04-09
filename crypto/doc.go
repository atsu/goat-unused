/*
	Package crypto provides functionality to scrypt a password and to AES encrypt and decrypt []bytes
	The following sources were used as references:
		https://golang.org/src/crypto/cipher/example_test.go
		https://proandroiddev.com/security-best-practices-symmetric-encryption-with-aes-in-java-7616beaaade9
		https://www.thepolyglotdeveloper.com/2018/02/encrypt-decrypt-data-golang-application-crypto-packages/
		https://github.com/elithrar/simple-scrypt

	To symmetrically encrypt arbitrary data using a plain password, this form may be used:
		crypto.EncryptGCM([]byte("data to encrypt", []byte("password123")) // to encrypt
		crypto.DecryptGCM(encryptedData, []byte("password123")) // to decrypt

	If it is desired to "hide" the password, first encrypt the password:
		scryptPkg, e := crypto.NewScryptPkg([]byte("password123"))

	and use the resulting ScryptPkg package to perform the encryption:
		encryptedDataRawBytes, e := scryptPkg.EncryptGCM([]byte("data to encrypt")) // raw encrypted bytes
	or
		encryptedData, e := scryptPkg.EncryptGCMBase64([]byte("data to encrypt")) // base64 encoded encrypted bytes

	to decrypt:
		decrypted, e := scryptPkg.DecryptGCM(rawEncryptedBytes)
	or
		decrypted, e := scryptPkg.EncryptGCMBase64(base64EncodedEncryptedBytes)

	In either case, note that the crypto package relies on some hard coded defaults
	that determine key strength (see crypto.GetDefaultParams()). At some point,
	this package could be refactored to support settings other than the default,
	or something like https://github.com/elithrar/simple-scrypt could be used as is.
*/
package crypto
