package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"exchange/internal/db"
	"exchange/internal/models"
)

// APIKeyRepo provides database operations for API applications, keys, and logs.
type APIKeyRepo struct {
	pool *db.Pool
}

// NewAPIKeyRepo creates a new APIKeyRepo.
func NewAPIKeyRepo(pool *db.Pool) *APIKeyRepo {
	return &APIKeyRepo{pool: pool}
}

// ── Applications ──────────────────────────────────────────────────────────────

// CreateApp inserts a new developer application.
func (r *APIKeyRepo) CreateApp(ctx context.Context, app *models.APIApplication) error {
	query := `
		INSERT INTO api_applications (id, user_id, name, is_live, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, app.ID, app.UserID, app.Name, app.IsLive, app.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting api application: %w", err)
	}
	return nil
}

// GetAppsByUser returns all applications owned by a user.
func (r *APIKeyRepo) GetAppsByUser(ctx context.Context, userID uuid.UUID) ([]models.APIApplication, error) {
	query := `
		SELECT id, user_id, name, is_live, created_at
		FROM api_applications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying apps: %w", err)
	}
	defer rows.Close()

	var apps []models.APIApplication
	for rows.Next() {
		var a models.APIApplication
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.IsLive, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning app: %w", err)
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

// GetAppByID returns a single application, verifying ownership.
func (r *APIKeyRepo) GetAppByID(ctx context.Context, appID, userID uuid.UUID) (*models.APIApplication, error) {
	query := `
		SELECT id, user_id, name, is_live, created_at
		FROM api_applications
		WHERE id = $1 AND user_id = $2
	`
	a := &models.APIApplication{}
	err := r.pool.QueryRow(ctx, query, appID, userID).Scan(&a.ID, &a.UserID, &a.Name, &a.IsLive, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting app by id: %w", err)
	}
	return a, nil
}

// DeleteApp removes an application and all its keys (CASCADE).
func (r *APIKeyRepo) DeleteApp(ctx context.Context, appID, userID uuid.UUID) error {
	query := `DELETE FROM api_applications WHERE id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, query, appID, userID)
	if err != nil {
		return fmt.Errorf("deleting app: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("app not found or not owned by user")
	}
	return nil
}

// UpdateAppLiveStatus updates the is_live status of an application.
func (r *APIKeyRepo) UpdateAppLiveStatus(ctx context.Context, appID uuid.UUID, userID uuid.UUID, isLive bool) error {
	query := `
		UPDATE api_applications 
		SET is_live = $3 
		WHERE id = $1 AND user_id = $2
	`
	cmdTag, err := r.pool.Exec(ctx, query, appID, userID, isLive)
	if err != nil {
		return fmt.Errorf("updating app live status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("application not found or you don't have permission")
	}
	return nil
}

// ── API Keys ──────────────────────────────────────────────────────────────────

// CreateKey inserts a new API key record.
func (r *APIKeyRepo) CreateKey(ctx context.Context, key *models.APIKey) error {
	query := `
		INSERT INTO api_keys (id, app_id, secret_hash, secret_hint, publishable_hash, publishable_hint, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		key.ID, key.AppID, key.SecretHash, key.SecretHint, key.PublishableHash, key.PublishableHint, key.IsActive, key.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting api key: %w", err)
	}
	return nil
}

// GetKeyByHash looks up a key by comparing the hash against BOTH columns.
// It returns the key details, the matched KeyType, and the owner's userID.
func (r *APIKeyRepo) GetKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, models.APIKeyType, uuid.UUID, error) {
	query := `
		SELECT k.id, k.app_id, k.secret_hash, k.secret_hint, k.publishable_hash, k.publishable_hint, k.is_active, k.last_used_at, k.created_at,
		       a.user_id, a.is_live
		FROM api_keys k
		JOIN api_applications a ON a.id = k.app_id
		WHERE k.secret_hash = $1 OR k.publishable_hash = $1
	`
	k := &models.APIKey{}
	var ownerID uuid.UUID
	var isLive bool

	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&k.ID, &k.AppID, &k.SecretHash, &k.SecretHint, &k.PublishableHash, &k.PublishableHint, &k.IsActive, &k.LastUsedAt, &k.CreatedAt,
		&ownerID, &isLive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", uuid.Nil, nil
		}
		return nil, "", uuid.Nil, fmt.Errorf("looking up key by hash: %w", err)
	}

	// Determine which key was actually used
	matchedType := models.APIKeyTypePublishable
	if k.SecretHash == keyHash {
		matchedType = models.APIKeyTypeSecret
	}

	return k, matchedType, ownerID, nil
}

// ListKeysByApp returns all keys for an application.
func (r *APIKeyRepo) ListKeysByApp(ctx context.Context, appID uuid.UUID) ([]models.APIKey, error) {
	query := `
		SELECT id, app_id, secret_hash, secret_hint, publishable_hash, publishable_hint, is_active, last_used_at, created_at
		FROM api_keys
		WHERE app_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, fmt.Errorf("querying keys: %w", err)
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.AppID, &k.SecretHash, &k.SecretHint, &k.PublishableHash, &k.PublishableHint, &k.IsActive, &k.LastUsedAt, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// RevokeKey deactivates an API key pair.
func (r *APIKeyRepo) RevokeKey(ctx context.Context, keyID, appID uuid.UUID) error {
	query := `UPDATE api_keys SET is_active = FALSE WHERE id = $1 AND app_id = $2`
	tag, err := r.pool.Exec(ctx, query, keyID, appID)
	if err != nil {
		return fmt.Errorf("revoking key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("key not found")
	}
	return nil
}

// UpdateKeyLastUsed updates the last_used_at timestamp for a key.
func (r *APIKeyRepo) UpdateKeyLastUsed(ctx context.Context, keyID uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, keyID)
	return err
}

// ── Request Logs ──────────────────────────────────────────────────────────────

// LogRequest inserts a request log entry.
func (r *APIKeyRepo) LogRequest(ctx context.Context, log *models.APIRequestLog) error {
	query := `
		INSERT INTO api_request_logs (key_id, method, path, status_code, latency_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err := r.pool.Exec(ctx, query, log.KeyID, log.Method, log.Path, log.StatusCode, log.LatencyMs)
	return err
}

// GetRequestLogs returns paginated request logs for a specific key.
func (r *APIKeyRepo) GetRequestLogs(ctx context.Context, keyID uuid.UUID, limit, offset int) ([]models.APIRequestLog, int, error) {
	countQuery := `SELECT count(*) FROM api_request_logs WHERE key_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, keyID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting logs: %w", err)
	}

	query := `
		SELECT id, key_id, method, path, status_code, latency_ms, created_at
		FROM api_request_logs
		WHERE key_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, keyID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying logs: %w", err)
	}
	defer rows.Close()

	var logs []models.APIRequestLog
	for rows.Next() {
		var l models.APIRequestLog
		if err := rows.Scan(&l.ID, &l.KeyID, &l.Method, &l.Path, &l.StatusCode, &l.LatencyMs, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}

// GetLogsByApp returns paginated request logs for all keys in an application.
func (r *APIKeyRepo) GetLogsByApp(ctx context.Context, appID uuid.UUID, limit, offset int) ([]models.APIRequestLog, int, error) {
	countQuery := `
		SELECT count(*)
		FROM api_request_logs rl
		JOIN api_keys k ON k.id = rl.key_id
		WHERE k.app_id = $1
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, appID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting app logs: %w", err)
	}

	query := `
		SELECT rl.id, rl.key_id, rl.method, rl.path, rl.status_code, rl.latency_ms, rl.created_at
		FROM api_request_logs rl
		JOIN api_keys k ON k.id = rl.key_id
		WHERE k.app_id = $1
		ORDER BY rl.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, appID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying app logs: %w", err)
	}
	defer rows.Close()

	var logs []models.APIRequestLog
	for rows.Next() {
		var l models.APIRequestLog
		if err := rows.Scan(&l.ID, &l.KeyID, &l.Method, &l.Path, &l.StatusCode, &l.LatencyMs, &l.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning app log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}

// ── Webhooks ──────────────────────────────────────────────────────────────────

// CreateWebhook inserts a new webhook.
func (r *APIKeyRepo) CreateWebhook(ctx context.Context, wh *models.Webhook) error {
	query := `
		INSERT INTO webhooks (id, app_id, url, events, secret, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, wh.ID, wh.AppID, wh.URL, wh.Events, wh.Secret, wh.IsActive, wh.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting webhook: %w", err)
	}
	return nil
}

// GetWebhooksByApp returns all webhooks for an application.
func (r *APIKeyRepo) GetWebhooksByApp(ctx context.Context, appID uuid.UUID) ([]models.Webhook, error) {
	query := `
		SELECT id, app_id, url, events, secret, is_active, created_at
		FROM webhooks
		WHERE app_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, fmt.Errorf("querying webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		if err := rows.Scan(&w.ID, &w.AppID, &w.URL, &w.Events, &w.Secret, &w.IsActive, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, rows.Err()
}

// DeleteWebhook removes a webhook.
func (r *APIKeyRepo) DeleteWebhook(ctx context.Context, webhookID, appID uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1 AND app_id = $2`
	tag, err := r.pool.Exec(ctx, query, webhookID, appID)
	if err != nil {
		return fmt.Errorf("deleting webhook: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// GetWebhooksForEvent returns active webhooks for an app that are subscribed to a specific event.
func (r *APIKeyRepo) GetWebhooksForEvent(ctx context.Context, appID uuid.UUID, event string) ([]models.Webhook, error) {
	query := `
		SELECT id, app_id, url, events, secret, is_active, created_at
		FROM webhooks
		WHERE app_id = $1 AND is_active = TRUE AND $2 = ANY(events)
	`
	rows, err := r.pool.Query(ctx, query, appID, event)
	if err != nil {
		return nil, fmt.Errorf("querying webhooks for event: %w", err)
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		if err := rows.Scan(&w.ID, &w.AppID, &w.URL, &w.Events, &w.Secret, &w.IsActive, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, rows.Err()
}
