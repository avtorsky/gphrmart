package auth

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
)

const saltBytes = 16

func GenerateSalt() (string, error) {
	salt := make([]byte, saltBytes)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func GenerateHash(password, salt string) string {
	passwordSum := md5.Sum([]byte(password))
	passwordSaltSum := md5.Sum(append(passwordSum[:], []byte(salt)...))
	return base64.StdEncoding.EncodeToString(passwordSaltSum[:])
}
