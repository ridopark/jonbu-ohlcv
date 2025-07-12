package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/ridopark/jonbu-ohlcv/internal/database"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
	"github.com/ridopark/jonbu-ohlcv/pkg/api/types"
)

// REQ-081: Health check endpoints for monitoring

type HealthHandler struct {
	db      *database.DB
	logger  zerolog.Logger
	version string
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *database.DB, version string) *HealthHandler {
	return &HealthHandler{
		db:      db,
		logger:  logger.NewContextLogger("health_handler"),
		version: version,
	}
}

// GetHealth handles GET /api/v1/health
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	correlationID := uuid.New().String()
	reqLogger := logger.NewRequestLogger(correlationID, r.Method, r.URL.Path)

	reqLogger.Info().Msg("Processing health check request")

	// Check database health
	ctx := r.Context()
	dbHealth := h.db.HealthCheck(ctx)

	// Determine overall status
	status := "healthy"
	if dbStatus, ok := dbHealth["status"].(string); ok && dbStatus != "healthy" {
		status = "unhealthy"
	}

	// Build response
	response := &types.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   h.version,
		Components: map[string]interface{}{
			"database": dbHealth,
		},
	}

	// Set appropriate status code
	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(statusCode)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		reqLogger.Error().Err(err).Msg("Failed to encode health response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	reqLogger.Info().
		Str("status", status).
		Msg("Health check completed")
}
