package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type authService struct {
	userRepository store.UserRepository
	hashKey        string
	tokenSignKey   string
	tokenIssuer    string
	tokenDuration  time.Duration

	logger *logger.Logger
}

func NewAuthService(userRepository store.UserRepository, cfg config.App, logger *logger.Logger) AuthService {
	return &authService{
		userRepository: userRepository,
		hashKey:        cfg.PasswordHashKey,
		tokenSignKey:   cfg.TokenSignKey,
		tokenIssuer:    cfg.TokenIssuer,
		tokenDuration:  cfg.TokenDuration,
		logger:         logger,
	}
}

func (a *authService) RegisterUser(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	if user.Login == "" || user.MasterPassword == "" {
		log.Error().Any("user", user).Msg("invalid user data provided")
		return models.User{}, ErrInvalidDataProvided
	}

	a.hashPassword(&user)

	registeredUser, err := a.userRepository.CreateUser(ctx, user)
	if err != nil {
		log.Err(err).Any("user", user).Msg("user creation ended with error")
		return models.User{}, fmt.Errorf("user creation ended with error: %w", err)
	}

	return registeredUser, nil
}

func (a *authService) Login(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	if user.Login == "" || user.MasterPassword == "" {
		log.Error().Any("user", user).Msg("invalid user data provided")
		return models.User{}, ErrInvalidDataProvided
	}

	a.hashPassword(&user)
	foundUser, err := a.userRepository.FindUserByLogin(ctx, user)
	if err != nil {
		log.Err(err).Any("user", user).Msg("user search by login failed")
		return models.User{}, fmt.Errorf("user search by login failed: %w", err)
	}

	if foundUser.MasterPassword != user.MasterPassword {
		log.Err(err).
			Int64("id", foundUser.UserID).
			Str("login", foundUser.Login).
			Str("typed password", user.MasterPassword).
			Str("actual password", foundUser.MasterPassword).
			Msg("wrong password")
		return models.User{}, ErrWrongPassword
	}

	return foundUser, nil
}

func (a *authService) CreateToken(ctx context.Context, user models.User) (models.Token, error) {
	token, err := utils.GenerateJWTToken(a.tokenIssuer, user.UserID, a.tokenDuration, a.tokenSignKey)
	if err != nil {
		return models.Token{}, fmt.Errorf("error creating JWT token: %w", err)
	}

	return token, nil
}

func (a *authService) ParseToken(ctx context.Context, tokenString string) (models.Token, error) {
	token, err := utils.ValidateAndParseJWTToken(tokenString, a.tokenSignKey, a.tokenIssuer)
	if err != nil {
		return models.Token{}, fmt.Errorf("error parsing JWT token: %w", err)
	}

	return token, nil
}

func (a *authService) hashPassword(user *models.User) {
	user.MasterPassword = utils.HashString(user.MasterPassword, a.hashKey)
}
