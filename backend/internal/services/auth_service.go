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
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
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

	return user, nil
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
		"sub":  user.ID.String(),
		"role": user.Role,
		"type": "access",
		"exp":  time.Now().Add(s.cfg.JWTAccessExpiry).Unix(),
		"iat":  time.Now().Unix(),
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
