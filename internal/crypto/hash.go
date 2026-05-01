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
	result := hex.EncodeToString(bytes)
	ZeroBytes(bytes)
	return result
}

func CheckSecret(hash string, guess string) bool {
	return HashSecret(guess) == hash
}

func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func ZeroString(s *string) {
	b := []byte(*s)
	for i := range b {
		b[i] = 0
	}
	*s = ""
}
