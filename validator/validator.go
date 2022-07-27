package validator

import (
	"net/mail"
	"regexp"

	"github.com/pkg/errors"
)

var (
	isValidUsername = regexp.MustCompile(`^\w+$`).MatchString
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\\s]$`).MatchString
)

func ValidateString(value string, minLen, maxLen int) error {
	n := len(value)
	if n < minLen || n > maxLen {
		return errors.Errorf("string length must be between %d and %d", minLen, maxLen)
	}

	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	if !isValidUsername(value) {
		return errors.Errorf("username must be alphanumeric")
	}

	return nil
}

func ValidatePassword(value string) error {
	return ValidateString(value, 6, 100)
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 6, 200); err != nil {
		return err
	}

	_, err := mail.ParseAddress(value)

	return errors.Wrap(err, "invalid email address")
}

func ValidateFullName(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}

	if !isValidFullName(value) {
		return errors.Errorf("full name must contain only alphabets and spaces")
	}

	return nil
}
