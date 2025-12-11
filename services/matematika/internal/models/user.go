package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email string `json:"email" gorm:"type:varchar(255);not null;unique"`
	Password string `json:"password" gorm:"type:varchar(255);not null"`
	Role string `json:"role" gorm:"not null"`
}
