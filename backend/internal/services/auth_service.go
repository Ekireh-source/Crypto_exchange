package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"exchange/config"
	"exchange/internal/db/repository"
	"exchange/internal/models"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidReferral  = errors.New("invalid referral code")
)

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepo
	emailSvc EmailService
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepo, emailSvc EmailService) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		emailSvc: emailSvc,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, email, password, referralCode string) (*models.User, error) {
	// Check if user already exists
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrUserExists
	}

	var referredBy *uuid.UUID
	if referralCode != "" {
		referrer, err := s.userRepo.GetByReferralCode(ctx, referralCode)
		if err != nil {
			return nil, fmt.Errorf("checking referral code: %w", err)
		}
		if referrer == nil {
			return nil, ErrInvalidReferral
		}
		referredBy = &referrer.ID
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	code, err := generateReferralCode()
	if err != nil {
		return nil, fmt.Errorf("generating referral code: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		ReferralCode: code,
		ReferredBy:   referredBy,
		KYCStatus:    models.KYCPending,
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Generate Verification Token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)
	
	emailToken := &models.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := s.userRepo.CreateEmailVerificationToken(ctx, emailToken); err != nil {
		return nil, fmt.Errorf("creating verification token: %w", err)
	}

	// Send verification email asynchronously
	if s.emailSvc != nil {
		go func(uEmail, vToken string) {
			templateData := map[string]string{
				"Email":  uEmail,
				"AppURL": s.cfg.FrontendURL,
				"Token":  vToken,
			}
			// Use context.Background() because the request context might be cancelled once the HTTP response is sent
			err := s.emailSvc.SendEmail(context.Background(), uEmail, "Verify Your Email - Crypto Exchange", "verify_email.html", templateData)
			if err != nil {
				fmt.Printf("Failed to send verification email to %s: %v\n", uEmail, err)
			}
		}(user.Email, tokenStr)
	}

	return user, nil
}

// VerifyEmail validates the token and marks the user as verified.
func (s *AuthService) VerifyEmail(ctx context.Context, tokenStr string) error {
	token, err := s.userRepo.GetByVerificationToken(ctx, tokenStr)
	if err != nil {
		return fmt.Errorf("fetching verification token: %w", err)
	}
	if token == nil {
		return ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		_ = s.userRepo.DeleteVerificationToken(context.Background(), tokenStr)
		return errors.New("verification token expired")
	}

	if err := s.userRepo.MarkEmailVerified(ctx, token.UserID); err != nil {
		return fmt.Errorf("marking email verified: %w", err)
	}

	// Clean up the token
	_ = s.userRepo.DeleteVerificationToken(context.Background(), tokenStr)

	return nil
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login authenticates a user and returns the user model and a JWT token pair.
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, *TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching user: %w", err)
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.GenerateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// RefreshToken validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		return nil, ErrInvalidToken
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching user for refresh: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	return s.GenerateTokenPair(user)
}

func (s *AuthService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"sub":            user.ID.String(),
		"role":           user.Role,
		"email_verified": user.IsEmailVerified,
		"type":           "access",
		"exp":            time.Now().Add(s.cfg.JWTAccessExpiry).Unix(),
		"iat":            time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"type": "refresh",
		"exp":  time.Now().Add(s.cfg.JWTRefreshExpiry).Unix(),
		"iat":  time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

func generateReferralCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetProfile fetches the user's profile, referral stats, and referred users.
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *models.ReferralStats, []*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetching user profile: %w", err)
	}
	if user == nil {
		return nil, nil, nil, errors.New("user not found")
	}

	stats, err := s.userRepo.GetReferralStats(ctx, userID, user.ReferralCode)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetching referral stats: %w", err)
	}

	referredUsers, err := s.userRepo.GetReferredUsers(ctx, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetching referred users: %w", err)
	}

	return user, stats, referredUsers, nil
}
