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

// authService is the concrete implementation of AuthService.
// It handles user registration, credential verification, and JWT token
// lifecycle using a UserRepository for persistence and HMAC-SHA256 for
// password hashing.
type authService struct {
	// userRepository is the data-access layer used to create and look up users.
	userRepository store.UserRepository

	// hashKey is the HMAC secret used when hashing user passwords before
	// storage or comparison. Must match the value used at registration time.
	hashKey string

	// tokenSignKey is the HMAC secret used to sign and verify JWT tokens.
	tokenSignKey string

	// tokenIssuer is the "iss" claim embedded in every issued JWT.
	// Tokens whose issuer does not match this value are rejected during parsing.
	tokenIssuer string

	// tokenDuration controls how long a newly issued JWT remains valid.
	tokenDuration time.Duration

	// logger is the structured logger used for diagnostic and error output.
	logger *logger.Logger
}

// NewAuthService constructs a new AuthService wired to the given UserRepository
// and populated with security parameters from cfg.
//
// The returned service is safe for concurrent use; all state is read-only after
// construction.
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

// RegisterUser creates a new user account.
//
// It validates that both Login and MasterPassword are non-empty, hashes the
// password with the configured HMAC key, and delegates persistence to the
// UserRepository.
//
// Returns the persisted user (with a server-assigned UserID) or:
//   - ErrInvalidDataProvided if Login or MasterPassword is empty.
//   - A wrapped storage error if the repository call fails (e.g. login already
//     taken — see store.ErrLoginAlreadyExists).
func (a *authService) RegisterUser(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	if user.Login == "" || user.AuthHash == "" {
		log.Error().Any("user", user).Msg("invalid user data provided")
		return models.User{}, ErrInvalidDataProvided
	}

	registeredUser, err := a.userRepository.CreateUser(ctx, user)
	if err != nil {
		log.Err(err).Any("user", user).Msg("user creation ended with error")
		return models.User{}, fmt.Errorf("user creation ended with error: %w", err)
	}

	return registeredUser, nil
}

// Login authenticates an existing user.
//
// It validates that both Login and MasterPassword are non-empty, hashes the
// supplied password, looks up the account by login, and compares the hashed
// passwords.
//
// Returns the authenticated user record or:
//   - ErrInvalidDataProvided if Login or MasterPassword is empty.
//   - A wrapped storage error if the repository lookup fails (e.g. user not
//     found — see store.ErrNoUserWasFound).
//   - ErrWrongPassword if the hashed passwords do not match.
func (a *authService) Login(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	if user.Login == "" || user.AuthHash == "" {
		log.Error().Any("user", user).Msg("invalid user data provided")
		return models.User{}, ErrInvalidDataProvided
	}

	foundUser, err := a.userRepository.FindUserByLogin(ctx, user)
	if err != nil {
		log.Err(err).Any("user", user).Msg("user search by login failed")
		return models.User{}, fmt.Errorf("user search by login failed: %w", err)
	}

	if foundUser.AuthHash != user.AuthHash {
		log.Err(err).
			Int64("id", foundUser.UserID).
			Str("login", foundUser.Login).
			Str("foundUser.AuthHash", foundUser.AuthHash).
			Str("user.AuthHash", user.AuthHash).
			Msg("wrong password")
		return models.User{}, ErrWrongPassword
	}

	return foundUser, nil
}

func (a *authService) Params(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	if user.Login == "" {
		log.Error().Any("user", user).Msg("invalid user data provided")
		return models.User{}, ErrInvalidDataProvided
	}

	foundUser, err := a.userRepository.FindUserByLogin(ctx, user)
	if err != nil {
		log.Err(err).Any("user", user).Msg("user search by login failed")
		return models.User{}, fmt.Errorf("user search by login failed: %w", err)
	}

	return foundUser, nil
}

// CreateToken issues a signed JWT for the given user.
//
// The token is signed with the configured tokenSignKey, carries the configured
// tokenIssuer as the "iss" claim, and expires after tokenDuration.
//
// Returns the token model on success or a wrapped error if JWT generation fails.
func (a *authService) CreateToken(ctx context.Context, user models.User) (models.Token, error) {
	token, err := utils.GenerateJWTToken(a.tokenIssuer, user.UserID, a.tokenDuration, a.tokenSignKey)
	if err != nil {
		return models.Token{}, fmt.Errorf("%w: %w", ErrTokenCreationFailed, err)
	}

	return token, nil
}

// ParseToken validates and parses a raw JWT string.
//
// It delegates to utils.ValidateAndParseJWTToken, verifying the signature and
// the issuer claim. Any validation failure (expired, wrong issuer, malformed)
// is normalised to ErrTokenIsExpiredOrInvalid so that callers do not need to
// inspect low-level JWT errors.
//
// Returns the decoded token model on success or ErrTokenIsExpiredOrInvalid on
// any validation failure.
func (a *authService) ParseToken(ctx context.Context, tokenString string) (models.Token, error) {
	token, err := utils.ValidateAndParseJWTToken(tokenString, a.tokenSignKey, a.tokenIssuer)
	if err != nil {
		return models.Token{}, ErrTokenIsExpiredOrInvalid
	}

	return token, nil
}

// hashPassword replaces the plain-text MasterPassword in user with its
// HMAC-SHA256 hash computed using the service's hashKey.
// The mutation is applied in-place via a pointer receiver.
func (a *authService) hashPassword(user *models.User) {
	user.AuthHash = utils.HashString(user.AuthHash, a.hashKey)
}
