package myhmac

import (
	"crypto/hmac"
	"encoding/base64"

	"golang.org/x/crypto/sha3"
)

func HmacSha3_512(password []byte, salt []byte) []byte {
	HMAC := hmac.New(sha3.New512, salt)
	_, err := HMAC.Write([]byte(password))
	if err != nil {
		panic(err)
	}
	return HMAC.Sum([]byte{})
}

func HmacSha3_512Base64(password []byte, salt []byte) string {
	return base64.RawStdEncoding.EncodeToString(HmacSha3_512(password, salt))
}
