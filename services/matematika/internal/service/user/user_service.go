package userservice

import (
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	"github.com/IbadT/business_bank_back/services/matematika/internal/models"
	"github.com/IbadT/business_bank_back/services/matematika/internal/repository"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	jwt_pkg "github.com/IbadT/business_bank_back/services/matematika/pkg/jwt"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/validator"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(email, password string) (*domain.Token, error)
	Register(email, password string) (*domain.Token, error)
	SaveAssociatedCard(userIDStr string, associatedCard string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Login(email, password string) (*domain.Token, error) {
	op := "service.user.login"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"email": email})
	log.Info("Processing login")

	if err := validator.ValidateEmailAndPassword(email, password); err != nil {
		log.Error(err, "Email and password validation failed")
		return nil, err
	}
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		log.Error(err, "Failed to get user by email")
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Warn("Invalid password for email: %s", email)
		return nil, helpers.ErrInvalidPassword
	}

	accessToken, refreshToken, err := jwt_pkg.GenerateTokens(user.ID)
	if err != nil {
		log.Error(err, "Failed to generate tokens")
		return nil, err
	}

	log.Success("User logged in successfully")
	return domain.NewToken(accessToken, refreshToken), nil
}

func (s *userService) Register(email, password string) (*domain.Token, error) {
	op := "service.user.register"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{"email": email})
	log.Info("Processing registration")

	if err := validator.ValidateEmailAndPassword(email, password); err != nil {
		log.Error(err, "Email and password validation failed")
		return nil, err
	}

	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		log.Warn("User already exists: %s", email)
		return nil, helpers.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error(err, "Failed to hash password")
		return nil, err
	}

	user, err := domain.NewUser(email, string(hashedPassword), models.RoleUser)
	if err != nil {
		log.Error(err, "Failed to create user domain object")
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		log.Error(err, "Failed to create user in repository")
		return nil, err
	}

	accessToken, refreshToken, err := jwt_pkg.GenerateTokens(user.ID)
	if err != nil {
		log.Error(err, "Failed to generate tokens")
		return nil, err
	}

	log.Success("User registered successfully")
	return domain.NewToken(accessToken, refreshToken), nil
}

func (s *userService) SaveAssociatedCard(userIDStr string, associatedCard string) error {
	op := "service.user.saveAssociatedCard"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"user_id": userIDStr,
		"card":    associatedCard,
	})
	log.Info("Saving associated card")

	// валидация uuid
	if userIDStr == "" {
		log.Warn("userID is required")
		return helpers.ErrUserIDRequired
	}

	userID, err := helpers.ParseUserID(userIDStr)
	if err != nil {
		log.Error(err, "Invalid userID format")
		return err
	}
	if userID == uuid.Nil {
		log.Warn("userID is nil")
		return helpers.ErrInvalidUserID
	}

	// проверяем пользователя по userID
	userExisting, err := s.userRepo.GetByID(userID)
	if err != nil {
		log.Error(err, "Failed to get user by ID")
		return err
	}
	if userExisting == nil {
		log.Warn("User not found")
		return helpers.ErrUserNotFound
	}

	user, err := domain.NewUserWithID(userID, userExisting.Email, userExisting.Password, userExisting.Role, userExisting.CreatedAt)
	if err != nil {
		log.Error(err, "Failed to create user domain object")
		return err
	}

	if err := user.SetAssociatedCard(associatedCard); err != nil {
		log.Error(err, "Failed to set associated card")
		return err
	}

	// сохраняем associatedCard
	if err := s.userRepo.UpdateAssociatedCard(user); err != nil {
		log.Error(err, "Failed to update associated card in repository")
		return err
	}
	
	log.Success("Associated card saved successfully")
	return nil
}
