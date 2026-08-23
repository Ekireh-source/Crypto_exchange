package models

import (
	"time"

	"github.com/google/uuid"
)

// APIApplication represents a developer's registered application.
type APIApplication struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	Name      string    `db:"name"       json:"name"`
	IsLive    bool      `db:"is_live"    json:"is_live"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// APIKeyType distinguishes between secret and publishable keys.
type APIKeyType string

const (
	APIKeyTypeSecret      APIKeyType = "secret"
	APIKeyTypePublishable APIKeyType = "publishable"
)

// APIKey represents a stored API key pair.
type APIKey struct {
	ID               uuid.UUID  `db:"id"                 json:"id"`
	AppID            uuid.UUID  `db:"app_id"             json:"app_id"`
	SecretHash       string     `db:"secret_hash"        json:"-"`
	SecretHint       string     `db:"secret_hint"        json:"secret_hint"`
	PublishableHash  string     `db:"publishable_hash"   json:"-"`
	PublishableHint  string     `db:"publishable_hint"   json:"publishable_hint"`
	IsActive         bool       `db:"is_active"          json:"is_active"`
	LastUsedAt       *time.Time `db:"last_used_at"       json:"last_used_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at"         json:"created_at"`
}

// APIKeyContext is injected into echo.Context by the API key middleware
// after successful authentication.
type APIKeyContext struct {
	KeyID       uuid.UUID
	AppID       uuid.UUID
	OwnerUserID uuid.UUID
	KeyType     APIKeyType // Which of the two keys was actually used for the request
	Scopes      []string
	IsLive      bool
}

// APIRequestLog represents a single request made with an API key.
type APIRequestLog struct {
	ID         int64     `db:"id"          json:"id"`
	KeyID      uuid.UUID `db:"key_id"      json:"key_id"`
	Method     string    `db:"method"      json:"method"`
	Path       string    `db:"path"        json:"path"`
	StatusCode int       `db:"status_code" json:"status_code"`
	LatencyMs  int       `db:"latency_ms"  json:"latency_ms"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

// Webhook represents a registered webhook endpoint for an application.
type Webhook struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	AppID     uuid.UUID `db:"app_id"     json:"app_id"`
	URL       string    `db:"url"        json:"url"`
	Events    []string  `db:"events"     json:"events"`
	Secret    string    `db:"secret"     json:"secret,omitempty"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// GeneratedKeyPair is the one-time response when creating an app,
// containing the raw keys that are never stored.
type GeneratedKeyPair struct {
	ID             uuid.UUID `json:"id"`
	SecretKey      string    `json:"secret_key"`
	PublishableKey string    `json:"publishable_key"`
}
