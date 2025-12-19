package validator

import (
	"strings"

	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
)

func ValidateEmailAndPassword(email, password string) error {
	if email == "" {
		return helpers.ErrEmailRequired
	}
	if password == "" {
		return helpers.ErrPasswordRequired
	}
	if len(password) < 8 {
		return helpers.ErrPasswordTooShort
	}
	if !strings.Contains(email, "@") {
		return helpers.ErrInvalidEmail
	}
	return nil
}