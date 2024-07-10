package myrand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func Read(b []byte) {
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
}

func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func RandBytesHex(n int) string {
	return fmt.Sprintf("%x", RandBytes(n))
}

func RandBytesBase64(n int) string {
	return base64.StdEncoding.EncodeToString(RandBytes(n))
}
