package validator

import (
	"errors"
	"strings"
)

func ValidateEmailAndPassword(email, password string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if !strings.Contains(email, "@") {
		return errors.New("email field must be a valid email address")
	}
	return nil
}