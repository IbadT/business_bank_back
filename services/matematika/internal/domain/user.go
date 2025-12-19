package domain

import (
	"time"
	"unicode"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Password       string
	Role           string
	AssociatedCard *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewUser(email, passwordHash, role string) (*User, error) {
	createdTime := time.Now()
	isValidRole := models.IsValidRole(role)
	if !isValidRole {
		return nil, helpers.ErrInvalidRole
	}
	return &User{
		ID:             uuid.New(),
		Email:          email,
		Password:       passwordHash,
		Role:           role,
		AssociatedCard: nil,
		CreatedAt:      createdTime,
		UpdatedAt:      createdTime,
	}, nil
}

func NewUserWithID(id uuid.UUID, email, passwordHash, role string, createdAt time.Time) (*User, error) {
	updatedTime := time.Now()
	return &User{
		ID:             id,
		Email:          email,
		Password:       passwordHash,
		Role:           role,
		AssociatedCard: nil,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedTime,
	}, nil
}

func (u *User) SetAssociatedCard(associatedCard string) error {
	// валидация cardNumber
	if associatedCard == "" {
		return helpers.ErrAssociatedCardRequired
	}
	if len(associatedCard) != 16 {
		return helpers.ErrAssociatedCardInvalidLength
	}
	if !isDigitsOnly(associatedCard) {
		return helpers.ErrAssociatedCardInvalidFormat
	}
	u.AssociatedCard = &associatedCard
	return nil
}

func isDigitsOnly(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}
