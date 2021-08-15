package main

import (
    "crypto/rand"
    "crypto/aes"
	"encoding/base64"
    "github.com/rs/xid"
	"crypto/cipher"
	"errors"
	"io"
)


// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.

func GenerateRandomBytes(n int) (*[32]byte) {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}


// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (*[32]byte, string) {
    b := GenerateRandomBytes(s)
    return b, base64.URLEncoding.EncodeToString((b[:]))
}

func Generateid() (string) {
    guid := xid.New()
    return guid.String()
}
func MessageEncrypt(plaintext []byte, key *[32]byte) (ciphertext string) {

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil))
}

func MessageDecrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}