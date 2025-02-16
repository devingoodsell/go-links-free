package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/devingoodsell/go-links-free/internal/models"
)

type AuthService struct {
	userRepo    *models.UserRepository
	jwtManager  *JWTManager
	enableOkta  bool
	oktaService *OktaService
}

func NewAuthService(userRepo *models.UserRepository, jwtManager *JWTManager, enableOkta bool) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		enableOkta: enableOkta,
	}
}

func (s *AuthService) SetOktaService(oktaService *OktaService) {
	s.oktaService = oktaService
}

// GetUserByEmail retrieves a user by their email
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !s.userRepo.VerifyPassword(user, password) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) GetOktaAuthURL(state string) (string, error) {
	if !s.enableOkta {
		return "", errors.New("okta authentication is not enabled")
	}
	return s.oktaService.GetAuthURL(state), nil
}

func (s *AuthService) HandleOktaCallback(ctx context.Context, code string) (string, error) {
	if !s.enableOkta {
		return "", errors.New("okta authentication is not enabled")
	}

	token, err := s.oktaService.ExchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %v", err)
	}

	return s.LoginWithOkta(ctx, token.AccessToken)
}

func (s *AuthService) LoginWithOkta(ctx context.Context, oktaToken string) (string, error) {
	if !s.enableOkta {
		return "", errors.New("okta authentication is not enabled")
	}

	userInfo, err := s.oktaService.ValidateToken(ctx, oktaToken)
	if err != nil {
		return "", fmt.Errorf("okta token validation failed: %v", err)
	}

	// Get or create user
	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		// Create new user if not found
		user = &models.User{
			Email:   userInfo.Email,
			IsAdmin: false, // Set default role
		}
		// Create user without password for Okta users
		if err := s.userRepo.Create(ctx, user, ""); err != nil {
			return "", fmt.Errorf("failed to create user: %v", err)
		}
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return "", errors.New("email already registered")
	}

	// Create new user
	user := &models.User{
		Email:   email,
		IsAdmin: false, // Default to non-admin
	}

	// Create user with password
	if err := s.userRepo.Create(ctx, user, password); err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}
