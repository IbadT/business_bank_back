package domain

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	AccessToken string
	RefreshToken string
}

type User struct {
	ID uuid.UUID
	Email string
	PasswordHash string
	Role string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewToken(accessToken, refreshToken string) *Token {
	return &Token{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}
}

// func NewUser(email, password, role string) *User {
// 	return &User{
// 		Email: email,
// 		PasswordHash: password,
// 		Role: role,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}
// }