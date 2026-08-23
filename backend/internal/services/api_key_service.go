package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"exchange/internal/db/repository"
	"exchange/internal/models"
)

// APIKeyService handles API key generation, validation, and management.
type APIKeyService struct {
	repo *repository.APIKeyRepo
}

// NewAPIKeyService creates a new APIKeyService.
func NewAPIKeyService(repo *repository.APIKeyRepo) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// ── Application Management ───────────────────────────────────────────────────

// CreateApp creates a new developer application and generates its initial key pair.
func (s *APIKeyService) CreateApp(ctx context.Context, userID uuid.UUID, name string) (*models.APIApplication, *models.GeneratedKeyPair, error) {
	app := &models.APIApplication{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		IsLive:    false,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateApp(ctx, app); err != nil {
		return nil, nil, fmt.Errorf("creating app: %w", err)
	}

	// Generate initial key pair
	keyPair, err := s.GenerateKeyPair(ctx, app.ID, app.IsLive)
	if err != nil {
		return nil, nil, fmt.Errorf("generating initial keys: %w", err)
	}

	return app, keyPair, nil
}

// GetAppsByUser returns all applications owned by a user.
func (s *APIKeyService) GetAppsByUser(ctx context.Context, userID uuid.UUID) ([]models.APIApplication, error) {
	return s.repo.GetAppsByUser(ctx, userID)
}

// GetAppByID returns a single application, verifying ownership.
func (s *APIKeyService) GetAppByID(ctx context.Context, appID, userID uuid.UUID) (*models.APIApplication, error) {
	return s.repo.GetAppByID(ctx, appID, userID)
}

// DeleteApp removes an application and all its keys.
func (s *APIKeyService) DeleteApp(ctx context.Context, appID, userID uuid.UUID) error {
	return s.repo.DeleteApp(ctx, appID, userID)
}

// UpdateAppLiveStatus updates the live status of the application.
func (s *APIKeyService) UpdateAppLiveStatus(ctx context.Context, userID uuid.UUID, appID uuid.UUID, isLive bool) error {
	return s.repo.UpdateAppLiveStatus(ctx, appID, userID, isLive)
}

// ── Key Generation ───────────────────────────────────────────────────────────

// GenerateKeyPair creates both a secret key and a publishable key in a single row.
// The raw keys are returned ONCE and never stored.
func (s *APIKeyService) GenerateKeyPair(ctx context.Context, appID uuid.UUID, isLive bool) (*models.GeneratedKeyPair, error) {
	skPrefix := "sk_test_"
	pkPrefix := "pk_test_"
	if isLive {
		skPrefix = "sk_live_"
		pkPrefix = "pk_live_"
	}

	// Generate secret key
	rawSK, skHash, skHint, err := s.generateKeyComponents(skPrefix)
	if err != nil {
		return nil, fmt.Errorf("generating secret key components: %w", err)
	}

	// Generate publishable key
	rawPK, pkHash, pkHint, err := s.generateKeyComponents(pkPrefix)
	if err != nil {
		return nil, fmt.Errorf("generating publishable key components: %w", err)
	}

	keyPair := &models.APIKey{
		ID:              uuid.New(),
		AppID:           appID,
		SecretHash:      skHash,
		SecretHint:      skHint,
		PublishableHash: pkHash,
		PublishableHint: pkHint,
		IsActive:        true,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.CreateKey(ctx, keyPair); err != nil {
		return nil, fmt.Errorf("storing key pair: %w", err)
	}

	return &models.GeneratedKeyPair{
		ID:             keyPair.ID,
		SecretKey:      rawSK,
		PublishableKey: rawPK,
	}, nil
}

// generateKeyComponents creates a single key string, hashes it, and extracts the hint.
func (s *APIKeyService) generateKeyComponents(prefix string) (rawKey, hashHex, hint string, err error) {
	// Generate 20 random bytes → 40 hex chars
	randomBytes := make([]byte, 20)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", "", fmt.Errorf("generating random bytes: %w", err)
	}
	rawKey = prefix + hex.EncodeToString(randomBytes)

	// SHA-256 hash the raw key
	hash := sha256.Sum256([]byte(rawKey))
	hashHex = hex.EncodeToString(hash[:])

	// Last 6 chars as hint
	hint = rawKey[len(rawKey)-6:]

	return rawKey, hashHex, hint, nil
}

// ── Key Validation ───────────────────────────────────────────────────────────

// ValidateKey validates a raw API key and returns the context needed for authorization.
func (s *APIKeyService) ValidateKey(ctx context.Context, rawKey string) (*models.APIKeyContext, error) {
	// Hash the incoming key
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// Look up by hash
	key, matchedType, ownerID, err := s.repo.GetKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("looking up key: %w", err)
	}
	if key == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if key is active
	if !key.IsActive {
		return nil, fmt.Errorf("API key pair has been revoked")
	}

	return &models.APIKeyContext{
		KeyID:       key.ID,
		AppID:       key.AppID,
		OwnerUserID: ownerID,
		KeyType:     matchedType,
		IsLive:      rawKey[:7] == "sk_live" || rawKey[:7] == "pk_live",
	}, nil
}

// ── Key Management ───────────────────────────────────────────────────────────

// ListKeysByApp returns all keys for an application.
func (s *APIKeyService) ListKeysByApp(ctx context.Context, appID uuid.UUID) ([]models.APIKey, error) {
	return s.repo.ListKeysByApp(ctx, appID)
}

// RevokeKey deactivates an API key pair.
func (s *APIKeyService) RevokeKey(ctx context.Context, keyID, appID uuid.UUID) error {
	return s.repo.RevokeKey(ctx, keyID, appID)
}

// ── Request Logging ──────────────────────────────────────────────────────────

// LogRequest records an API request (fire-and-forget).
func (s *APIKeyService) LogRequest(ctx context.Context, log *models.APIRequestLog) {
	go func() {
		_ = s.repo.LogRequest(context.Background(), log)
	}()
}

// UpdateKeyLastUsed updates the last-used timestamp (fire-and-forget).
func (s *APIKeyService) UpdateKeyLastUsed(keyID uuid.UUID) {
	go func() {
		_ = s.repo.UpdateKeyLastUsed(context.Background(), keyID)
	}()
}

// GetLogsByApp returns paginated request logs for an application.
func (s *APIKeyService) GetLogsByApp(ctx context.Context, appID uuid.UUID, limit, offset int) ([]models.APIRequestLog, int, error) {
	return s.repo.GetLogsByApp(ctx, appID, limit, offset)
}
