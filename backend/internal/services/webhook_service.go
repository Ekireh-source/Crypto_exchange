package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"exchange/internal/db/repository"
	"exchange/internal/models"
)

// WebhookService handles webhook registration and event dispatching.
type WebhookService struct {
	repo   *repository.APIKeyRepo
	client *http.Client
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(repo *repository.APIKeyRepo) *WebhookService {
	return &WebhookService{
		repo: repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterWebhook creates a new webhook endpoint for an application.
func (s *WebhookService) RegisterWebhook(ctx context.Context, appID uuid.UUID, url string, events []string) (*models.Webhook, error) {
	// Generate a random HMAC signing secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("generating webhook secret: %w", err)
	}
	secret := "whsec_" + hex.EncodeToString(secretBytes)

	wh := &models.Webhook{
		ID:        uuid.New(),
		AppID:     appID,
		URL:       url,
		Events:    events,
		Secret:    secret,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateWebhook(ctx, wh); err != nil {
		return nil, fmt.Errorf("creating webhook: %w", err)
	}

	return wh, nil
}

// ListWebhooks returns all webhooks for an application.
func (s *WebhookService) ListWebhooks(ctx context.Context, appID uuid.UUID) ([]models.Webhook, error) {
	return s.repo.GetWebhooksByApp(ctx, appID)
}

// DeleteWebhook removes a webhook.
func (s *WebhookService) DeleteWebhook(ctx context.Context, webhookID, appID uuid.UUID) error {
	return s.repo.DeleteWebhook(ctx, webhookID, appID)
}

// WebhookPayload is the standard structure sent to webhook endpoints.
type WebhookPayload struct {
	Event     string    `json:"event"`
	Data      any       `json:"data"`
	Timestamp int64     `json:"timestamp"`
	ID        string    `json:"id"`
}

// Dispatch fires an event to all webhooks subscribed to it for the given app.
// This is fire-and-forget; failures are logged but don't block the caller.
func (s *WebhookService) Dispatch(ctx context.Context, appID uuid.UUID, event string, data any) {
	go func() {
		hooks, err := s.repo.GetWebhooksForEvent(context.Background(), appID, event)
		if err != nil {
			log.Printf("ERROR webhook dispatch: failed to get webhooks for event %s: %v", event, err)
			return
		}

		for _, hook := range hooks {
			go s.deliverWebhook(hook, event, data)
		}
	}()
}

// deliverWebhook sends a single webhook payload with HMAC-SHA256 signature.
func (s *WebhookService) deliverWebhook(hook models.Webhook, event string, data any) {
	timestamp := time.Now().Unix()
	payload := WebhookPayload{
		Event:     event,
		Data:      data,
		Timestamp: timestamp,
		ID:        uuid.New().String(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR webhook: failed to marshal payload for %s: %v", hook.URL, err)
		return
	}

	// Compute HMAC-SHA256 signature (Stripe style)
	timestampStr := strconv.FormatInt(timestamp, 10)
	signedPayload := timestampStr + "." + string(body)

	mac := hmac.New(sha256.New, []byte(hook.Secret))
	mac.Write([]byte(signedPayload))
	signatureHex := hex.EncodeToString(mac.Sum(nil))

	stripeStyleSignature := fmt.Sprintf("t=%s,v1=%s", timestampStr, signatureHex)

	req, err := http.NewRequest("POST", hook.URL, bytes.NewReader(body))
	if err != nil {
		log.Printf("ERROR webhook: failed to create request for %s: %v", hook.URL, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Exchange-Signature", stripeStyleSignature)
	req.Header.Set("X-Exchange-Event", event)
	req.Header.Set("X-Exchange-Delivery", payload.ID)

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("WARN webhook delivery failed for %s (event=%s): %v", hook.URL, event, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("WARN webhook %s returned status %d for event %s", hook.URL, resp.StatusCode, event)
	}
}
