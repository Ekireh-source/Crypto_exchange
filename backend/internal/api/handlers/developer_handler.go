package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"exchange/internal/api/middleware"
	"exchange/internal/models"
	"exchange/internal/services"
)

// DeveloperHandler serves the developer portal management endpoints.
// These are protected by JWT (the same auth as regular UI routes).
type DeveloperHandler struct {
	apiKeySvc  *services.APIKeyService
	webhookSvc *services.WebhookService
}

// NewDeveloperHandler creates a new DeveloperHandler.
func NewDeveloperHandler(apiKeySvc *services.APIKeyService, webhookSvc *services.WebhookService) *DeveloperHandler {
	return &DeveloperHandler{
		apiKeySvc:  apiKeySvc,
		webhookSvc: webhookSvc,
	}
}

// ── Application Management ───────────────────────────────────────────────────

// CreateApp creates a new developer application and returns its initial key pair.
// The raw keys are shown ONCE in this response — they cannot be retrieved again.
//
// POST /v1/developer/apps
func (h *DeveloperHandler) CreateApp(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	app, keyPair, err := h.apiKeySvc.CreateApp(c.Request().Context(), userID, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"app":  app,
		"keys": keyPair,
		"message": "Save these keys now. The secret key will not be shown again.",
	})
}

// ListApps returns all applications owned by the authenticated user.
//
// GET /v1/developer/apps
func (h *DeveloperHandler) ListApps(c echo.Context) error {
	userID := middleware.GetUserID(c)

	apps, err := h.apiKeySvc.GetAppsByUser(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if apps == nil {
		apps = []models.APIApplication{}
	}

	return c.JSON(http.StatusOK, apps)
}

// DeleteApp removes an application and all its keys.
//
// DELETE /v1/developer/apps/:appID
func (h *DeveloperHandler) DeleteApp(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	if err := h.apiKeySvc.DeleteApp(c.Request().Context(), appID, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateAppStatus updates the is_live status of an application.
//
// PUT /v1/developer/apps/:appID/status
func (h *DeveloperHandler) UpdateAppStatus(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	var req struct {
		IsLive bool `json:"is_live"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.apiKeySvc.UpdateAppLiveStatus(c.Request().Context(), userID, appID, req.IsLive); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "application status updated successfully",
		"is_live": req.IsLive,
	})
}

// ── API Keys ─────────────────────────────────────────────────────────────────

// ListKeys returns all keys for an application (hashes + hints only, never raw keys).
//
// GET /v1/developer/apps/:appID/keys
func (h *DeveloperHandler) ListKeys(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	keys, err := h.apiKeySvc.ListKeysByApp(c.Request().Context(), appID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, keys)
}

// RegenerateKeys generates a new key pair for an application, revoking the old ones.
//
// POST /v1/developer/apps/:appID/keys/regenerate
func (h *DeveloperHandler) RegenerateKeys(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	// Revoke all existing keys for this app
	existingKeys, err := h.apiKeySvc.ListKeysByApp(c.Request().Context(), appID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	for _, k := range existingKeys {
		_ = h.apiKeySvc.RevokeKey(c.Request().Context(), k.ID, appID)
	}

	// Generate new pair
	keyPair, err := h.apiKeySvc.GenerateKeyPair(c.Request().Context(), appID, app.IsLive)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"keys":    keyPair,
		"message": "New keys generated. Old keys have been revoked. Save these keys now.",
	})
}

// RevokeKey deactivates a specific API key.
//
// DELETE /v1/developer/apps/:appID/keys/:keyID
func (h *DeveloperHandler) RevokeKey(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}
	keyID, err := uuid.Parse(c.Param("keyID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid key ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	if err := h.apiKeySvc.RevokeKey(c.Request().Context(), keyID, appID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// ── Request Logs ──────────────────────────────────────────────────────────────

// GetLogs returns paginated request logs for an application.
//
// GET /v1/developer/apps/:appID/logs
func (h *DeveloperHandler) GetLogs(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	logs, total, err := h.apiKeySvc.GetLogsByApp(c.Request().Context(), appID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ── Webhooks ──────────────────────────────────────────────────────────────────

// CreateWebhook registers a new webhook endpoint.
//
// POST /v1/developer/apps/:appID/webhooks
func (h *DeveloperHandler) CreateWebhook(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	var req struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := c.Bind(&req); err != nil || req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "url is required")
	}

	wh, err := h.webhookSvc.RegisterWebhook(c.Request().Context(), appID, req.URL, req.Events)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"webhook": wh,
		"message": "Save the webhook secret. You'll need it to verify webhook signatures.",
	})
}

// ListWebhooks returns all webhooks for an application.
//
// GET /v1/developer/apps/:appID/webhooks
func (h *DeveloperHandler) ListWebhooks(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	webhooks, err := h.webhookSvc.ListWebhooks(c.Request().Context(), appID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, webhooks)
}

// DeleteWebhook removes a webhook endpoint.
//
// DELETE /v1/developer/apps/:appID/webhooks/:webhookID
func (h *DeveloperHandler) DeleteWebhook(c echo.Context) error {
	userID := middleware.GetUserID(c)
	appID, err := uuid.Parse(c.Param("appID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid app ID")
	}
	webhookID, err := uuid.Parse(c.Param("webhookID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid webhook ID")
	}

	// Verify ownership
	app, err := h.apiKeySvc.GetAppByID(c.Request().Context(), appID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if app == nil {
		return echo.NewHTTPError(http.StatusNotFound, "app not found")
	}

	if err := h.webhookSvc.DeleteWebhook(c.Request().Context(), webhookID, appID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
