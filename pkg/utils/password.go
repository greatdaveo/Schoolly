package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, ".")
	if len(parts) != 2 {
		return ErrorHandler(errors.New("❌ invalid Encoded hash format"), "❌ Invalid Encoded hash format")
	}

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(err, "❌ failed to decode the salt")

	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return ErrorHandler(err, "❌ failed to decode the hashed password")
	}

	// For Hashing
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPassword) {
		return ErrorHandler(errors.New("❌ incorrect password"), "❌ incorrect password")
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		return nil
	}

	return ErrorHandler(errors.New("❌ incorrect password"), "❌ incorrect password")

}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrorHandler(errors.New("❌ password is blank"), "❌ Please enter password")
	}

	// To hash the password
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("❌ failed to generate salt"), "❌ Internal Server Error")
	}

	// For Hashing
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	// To encode the salt
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	// To override the password field with the hashed password

	return encodedHash, nil
}
