package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"exchange/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type IdempotencyRepo struct {
	pool *db.Pool
}

func NewIdempotencyRepo(pool *db.Pool) *IdempotencyRepo {
	return &IdempotencyRepo{pool: pool}
}

type IdempotencyStatus string

const (
	StatusInProgress IdempotencyStatus = "IN_PROGRESS"
	StatusCompleted  IdempotencyStatus = "COMPLETED"
)

type IdempotencyKey struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	IdempotencyKey string            `json:"idempotency_key"`
	RequestMethod  string            `json:"request_method"`
	RequestPath    string            `json:"request_path"`
	ResponseStatus int               `json:"response_status"`
	ResponseBody   json.RawMessage   `json:"response_body"`
	Status         IdempotencyStatus `json:"status"`
	ExpiresAt      time.Time         `json:"expires_at"`
}

var ErrKeyExists = errors.New("idempotency key already exists")

// AcquireLock attempts to insert a new idempotency key with IN_PROGRESS status.
// If it already exists, it returns the existing record and ErrKeyExists.
func (r *IdempotencyRepo) AcquireLock(ctx context.Context, userID, key, method, path string, expiresAt time.Time) (*IdempotencyKey, error) {
	query := `
		INSERT INTO idempotency_keys (user_id, idempotency_key, request_method, request_path, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, idempotency_key, request_method, request_path, status, expires_at
	`
	var k IdempotencyKey
	err := r.pool.QueryRow(ctx, query, userID, key, method, path, StatusInProgress, expiresAt).Scan(
		&k.ID, &k.UserID, &k.IdempotencyKey, &k.RequestMethod, &k.RequestPath, &k.Status, &k.ExpiresAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			// Key exists, fetch it
			existingKey, fetchErr := r.GetKey(ctx, userID, key)
			if fetchErr != nil {
				return nil, fetchErr
			}
			return existingKey, ErrKeyExists
		}
		return nil, err
	}

	return &k, nil
}

// GetKey retrieves an existing idempotency key.
func (r *IdempotencyRepo) GetKey(ctx context.Context, userID, key string) (*IdempotencyKey, error) {
	query := `
		SELECT id, user_id, idempotency_key, request_method, request_path, response_status, response_body, status, expires_at
		FROM idempotency_keys
		WHERE user_id = $1 AND idempotency_key = $2
	`
	var k IdempotencyKey
	var responseBody []byte
	err := r.pool.QueryRow(ctx, query, userID, key).Scan(
		&k.ID, &k.UserID, &k.IdempotencyKey, &k.RequestMethod, &k.RequestPath, &k.ResponseStatus, &responseBody, &k.Status, &k.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if responseBody != nil {
		k.ResponseBody = responseBody
	}

	return &k, nil
}

// SaveResponse updates an idempotency key to COMPLETED with the response details.
func (r *IdempotencyRepo) SaveResponse(ctx context.Context, userID, key string, statusCode int, responseBody []byte) error {
	query := `
		UPDATE idempotency_keys
		SET status = $1, response_status = $2, response_body = $3, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $4 AND idempotency_key = $5
	`
	_, err := r.pool.Exec(ctx, query, StatusCompleted, statusCode, responseBody, userID, key)
	return err
}
