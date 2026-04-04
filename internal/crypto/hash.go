package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func HashSecret(secret string) string {

	bytes := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(bytes[:])
}

func GenerateSecret() string {

	bytes := make([]byte, 16)

	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
func CheckSecret(hash string, guess string) bool {
	return HashSecret(guess) == hash
}
