package service

import (
	"errors"
	"fmt"

	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	jwt_pkg "github.com/IbadT/business_bank_back/services/matematika/pkg/jwt"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(email, password string) (*domain.Token, error)
	Register(email, password string) (*domain.Token, error)
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
	if err := validator.ValidateEmailAndPassword(email, password); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	accessToken, refreshToken, err := jwt_pkg.GenerateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return domain.NewToken(accessToken, refreshToken), nil
}

func (s *userService) Register(email, password string) (*domain.Token, error) {
	if err := validator.ValidateEmailAndPassword(email, password); err != nil {
		return nil, err
	}
	
	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}
	fmt.Println("EXISTING USER: ", existingUser)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(email, string(hashedPassword), models.RoleUser)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := jwt_pkg.GenerateTokens(user.ID)
	if err != nil {
		return nil, err
	}

	return domain.NewToken(accessToken, refreshToken), nil
}

