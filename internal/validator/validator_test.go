package validator

import (
	"strings"
	"testing"
)

func TestValidateName_Valid(t *testing.T) {
	names := []string{"alice", "bob_123", "user.name", "Alice-Bob"}
	for _, name := range names {
		if err := ValidateName(name); err != nil {
			t.Errorf("expected %s to be valid, got error: %s", name, err)
		}
	}
}

func TestValidateName_Empty(t *testing.T) {
	if err := ValidateName(""); err == nil {
		t.Errorf("expected empty name to fail")
	}
}

func TestValidateName_TooLong(t *testing.T) {
	long := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := ValidateName(long); err == nil {
		t.Errorf("expected long name to fail")
	}
}

func TestValidateName_InvalidCharacters(t *testing.T) {
	names := []string{"alice!", "bob@email", "user name", "<script>"}
	for _, name := range names {
		if err := ValidateName(name); err == nil {
			t.Errorf("expected %s to fail validation", name)
		}
	}
}

func TestValidateAmount_Valid(t *testing.T) {
	amounts := []int{1, 10, 100, 9999, 10000}
	for _, amount := range amounts {
		if err := ValidateAmount(amount); err != nil {
			t.Errorf("expected %d to be valid, got error: %s", amount, err)
		}
	}
}

func TestValidateAmount_TooLow(t *testing.T) {
	if err := ValidateAmount(0); err == nil {
		t.Errorf("expected 0 to fail")
	}
	if err := ValidateAmount(-1); err == nil {
		t.Errorf("expected -1 to fail")
	}
}

func TestValidateAmount_TooHigh(t *testing.T) {
	if err := ValidateAmount(10001); err == nil {
		t.Errorf("expected 10001 to fail")
	}
}

func TestValidateReason_Valid(t *testing.T) {
	if err := ValidateReason("bob never delivered the goods"); err != nil {
		t.Errorf("expected valid reason to pass, got: %s", err)
	}
}

func TestValidateReason_Empty(t *testing.T) {
	if err := ValidateReason(""); err == nil {
		t.Errorf("expected empty reason to fail")
	}
}

func TestValidateReason_TooLong(t *testing.T) {
	long := strings.Repeat("a", 201)
	if err := ValidateReason(long); err == nil {
		t.Errorf("expected long reason to fail")
	}
}

func TestValidateLink_Valid(t *testing.T) {
	links := []string{
		"https://screenshot.com/proof123",
		"http://imgur.com/abc",
	}
	for _, link := range links {
		if err := ValidateLink(link); err != nil {
			t.Errorf("expected %s to be valid, got error: %s", link, err)
		}
	}
}

func TestValidateLink_Empty(t *testing.T) {
	if err := ValidateLink(""); err == nil {
		t.Errorf("expected empty link to fail")
	}
}

func TestValidateLink_NoProtocol(t *testing.T) {
	if err := ValidateLink("screenshot.com/proof"); err == nil {
		t.Errorf("expected link without protocol to fail")
	}
}

func TestValidateID_Valid(t *testing.T) {
	if err := ValidateID("4190bf8b-edcc-4d99-8401-551cc750fcb1"); err != nil {
		t.Errorf("expected valid UUID to pass, got: %s", err)
	}
}

func TestValidateID_Empty(t *testing.T) {
	if err := ValidateID(""); err == nil {
		t.Errorf("expected empty ID to fail")
	}
}

func TestValidateID_WrongLength(t *testing.T) {
	if err := ValidateID("short-id"); err == nil {
		t.Errorf("expected short ID to fail")
	}
}
