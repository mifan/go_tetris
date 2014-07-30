package utils

import (
	"encoding/base64"
	"fmt"

	"code.google.com/p/go.crypto/scrypt"
)

// encrypt the password using Scrypt
const (
	n      = 1 << 3
	r      = 1 << 2
	p      = 1 << 0
	keyLen = 1 << 5
)

var salt []byte

func SetScryptSalt(val []byte) { salt = val }

func Encrypt(data string) string {
	src, err := scrypt.Key([]byte(data), salt, n, r, p, keyLen)
	if err != nil {
		fmt.Println("encrypt error: ", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(src)
}
