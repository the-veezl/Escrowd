package stellar

import (
	"os"
	"testing"
)

func TestValidateNetwork_Testnet(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "testnet")
	if err := ValidateNetwork(); err != nil {
		t.Errorf("expected testnet to be valid, got: %s", err)
	}
}

func TestValidateNetwork_Missing(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "")
	if err := ValidateNetwork(); err == nil {
		t.Errorf("expected error for missing network")
	}
	os.Setenv("STELLAR_NETWORK", "testnet")
}

func TestValidateNetwork_Invalid(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "invalidnet")
	if err := ValidateNetwork(); err == nil {
		t.Errorf("expected error for invalid network")
	}
	os.Setenv("STELLAR_NETWORK", "testnet")
}

func TestGenerateEscrowWallet_Testnet(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "testnet")

	wallet, err := GenerateEscrowWallet("test-escrow-001")
	if err != nil {
		t.Errorf("expected wallet to be generated, got: %s", err)
	}
	if wallet.PublicKey == "" {
		t.Errorf("expected non-empty public key")
	}
	if wallet.SecretKey == "" {
		t.Errorf("expected non-empty secret key")
	}
	if wallet.EscrowID != "test-escrow-001" {
		t.Errorf("expected escrow ID to match")
	}
	if wallet.PublicKey[0] != 'G' {
		t.Errorf("expected public key to start with G, got %c", wallet.PublicKey[0])
	}
	if wallet.SecretKey[0] != 'S' {
		t.Errorf("expected secret key to start with S, got %c", wallet.SecretKey[0])
	}
}

func TestGenerateEscrowWallet_Mainnet(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "mainnet")

	_, err := GenerateEscrowWallet("test-escrow-002")
	if err == nil {
		t.Errorf("expected error on mainnet — mainnet not enabled")
	}
	os.Setenv("STELLAR_NETWORK", "testnet")
}

func TestGetMasterPublicKey_Valid(t *testing.T) {
	os.Setenv("STELLAR_MASTER_SECRET", os.Getenv("STELLAR_MASTER_SECRET"))

	pub, err := GetMasterPublicKey()
	if err != nil {
		t.Errorf("expected master public key, got: %s", err)
	}
	if pub == "" {
		t.Errorf("expected non-empty public key")
	}
	if pub[0] != 'G' {
		t.Errorf("expected public key to start with G")
	}
}

func TestGetMasterPublicKey_Missing(t *testing.T) {
	original := os.Getenv("STELLAR_MASTER_SECRET")
	os.Setenv("STELLAR_MASTER_SECRET", "")

	_, err := GetMasterPublicKey()
	if err == nil {
		t.Errorf("expected error for missing master secret")
	}

	os.Setenv("STELLAR_MASTER_SECRET", original)
}

func TestNetworkInfo(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "testnet")
	info := NetworkInfo()
	if info == "" {
		t.Errorf("expected non-empty network info")
	}
}

func TestIsTestnet_True(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "testnet")
	if !isTestnet() {
		t.Errorf("expected isTestnet to return true")
	}
}

func TestIsTestnet_False(t *testing.T) {
	os.Setenv("STELLAR_NETWORK", "mainnet")
	if isTestnet() {
		t.Errorf("expected isTestnet to return false for mainnet")
	}
	os.Setenv("STELLAR_NETWORK", "testnet")
}

