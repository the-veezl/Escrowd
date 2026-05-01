package crypto

import "testing"

func TestHashSecret_SameInputSameOutput(t *testing.T) {
	hash1 := HashSecret("hello")
	hash2 := HashSecret("hello")

	if hash1 != hash2 {
		t.Errorf("expected same hash for same input, got %s and %s", hash1, hash2)
	}
}

func TestHashSecret_DifferentInputDifferentOutput(t *testing.T) {
	hash1 := HashSecret("hello")
	hash2 := HashSecret("world")

	if hash1 == hash2 {
		t.Errorf("expected different hashes for different inputs")
	}
}

func TestHashSecret_NotEmpty(t *testing.T) {
	hash := HashSecret("anything")

	if hash == "" {
		t.Errorf("expected non-empty hash, got empty string")
	}
}

func TestGenerateSecret_IsRandom(t *testing.T) {
	secret1 := GenerateSecret()
	secret2 := GenerateSecret()
	secret3 := GenerateSecret()

	if secret1 == secret2 || secret2 == secret3 || secret1 == secret3 {
		t.Errorf("expected unique secrets, got duplicates")
	}
}

func TestGenerateSecret_NotEmpty(t *testing.T) {
	secret := GenerateSecret()

	if secret == "" {
		t.Errorf("expected non-empty secret, got empty string")
	}
}

func TestCheckSecret_CorrectSecret(t *testing.T) {
	secret := GenerateSecret()
	hash := HashSecret(secret)

	if !CheckSecret(hash, secret) {
		t.Errorf("expected CheckSecret to return true for correct secret")
	}
}

func TestCheckSecret_WrongSecret(t *testing.T) {
	secret := GenerateSecret()
	hash := HashSecret(secret)

	if CheckSecret(hash, "wrong-secret") {
		t.Errorf("expected CheckSecret to return false for wrong secret")
	}
}

func TestCheckSecret_EmptySecret(t *testing.T) {
	secret := GenerateSecret()
	hash := HashSecret(secret)

	if CheckSecret(hash, "") {
		t.Errorf("expected CheckSecret to return false for empty secret")
	}
}
