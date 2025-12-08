package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/IbadT/business_bank_back/services/matematika/internal/database"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(email, password string) (*domain.Token, error)
	Register(email, password string) (*domain.Token, error)
	generateTokens(userID uuid.UUID) (string, string, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService {
		userRepo: userRepo,
	}
}

func (s *userService) Login(email, password string) (*domain.Token, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return domain.NewToken(accessToken, refreshToken), nil
}

func (s *userService) Register(email, password string) (*domain.Token, error) {
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}
	fmt.Println("EXISTING USER: ", existingUser)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleUser,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return domain.NewToken(accessToken, refreshToken), nil
}

func (s *userService) generateTokens(userID uuid.UUID) (string, string, error) {
	// Сохраняем UUID как строку для корректной работы с JWT
	userIDStr := userID.String()
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userIDStr,
		"exp": time.Now().Add(time.Hour * 4).Unix(), // 4 часа
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userIDStr,
		"exp": time.Now().Add(time.Hour * 24 * 2).Unix(), // 2 дня
	})

	// Используем тот же дефолтный ключ, что и в middleware
	secretKey := database.GetEnv("JWT_SECRET", "super-secret-word")
	if secretKey == "" {
		secretKey = "super-secret-word"
	}

	accessTokenStr, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	refreshTokenStr, err := refreshToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenStr, refreshTokenStr, nil
}