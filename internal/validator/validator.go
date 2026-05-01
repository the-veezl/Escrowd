package validator

import (
	"errors"
	"strings"
	"unicode"
)

const (
	MaxNameLength   = 50
	MaxReasonLength = 200
	MaxLinkLength   = 500
	MinAmount       = 1
	MaxAmount       = 10000
)

func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if len(name) > MaxNameLength {
		return errors.New("name too long — maximum 50 characters")
	}
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_' && c != '-' && c != '.' {
			return errors.New("name contains invalid characters — use letters, numbers, _ - . only")
		}
	}
	return nil
}

func ValidateAmount(amount int) error {
	if amount < MinAmount {
		return errors.New("amount must be at least 1")
	}
	if amount > MaxAmount {
		return errors.New("amount cannot exceed 10000")
	}
	return nil
}

func ValidateReason(reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("reason cannot be empty")
	}
	if len(reason) > MaxReasonLength {
		return errors.New("reason too long — maximum 200 characters")
	}
	return nil
}

func ValidateLink(link string) error {
	link = strings.TrimSpace(link)
	if link == "" {
		return errors.New("link cannot be empty")
	}
	if len(link) > MaxLinkLength {
		return errors.New("link too long — maximum 500 characters")
	}
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return errors.New("link must start with http:// or https://")
	}
	return nil
}

func ValidateID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("ID cannot be empty")
	}
	if len(id) != 36 {
		return errors.New("invalid ID format")
	}
	return nil
}
