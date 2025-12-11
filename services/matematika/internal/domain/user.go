package domain

import (
	"errors"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/google/uuid"
)



type User struct {
	ID uuid.UUID
	Email string
	Password string
	Role string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email, passwordHash, role string) (*User, error) {
	createdTime := time.Now()
	isValidRole := models.IsValidRole(role)
	if !isValidRole {
		return nil, errors.New("invalid role")
	}
	return &User{
		ID: uuid.New(),
		Email: email,
		Password: passwordHash,
		Role: role,
		CreatedAt: createdTime,
		UpdatedAt: createdTime,
	}, nil
}
