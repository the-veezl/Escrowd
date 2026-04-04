package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(secret string) string {

	bytes := sha256.Sum256([]byte(secret))

	return hex.EncodeToString(bytes[:])
}

func GenerateSecret() string {

	bytes := make([]byte, 16)

	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
