package stream

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// REQ-017: WebSocket endpoints for real-time data streaming
// REQ-019: Connection lifecycle handling

// Server represents the WebSocket streaming server
type Server struct {
	hub    *Hub
	logger zerolog.Logger
}

// NewServer creates a new WebSocket server
func NewServer(logger zerolog.Logger) *Server {
	hub := NewHub(logger)

	return &Server{
		hub: hub,
		logger: logger.With().
			Str("component", "websocket_server").
			Logger(),
	}
}

// Start begins the WebSocket server
func (s *Server) Start() {
	s.hub.Start()
	s.logger.Info().Msg("WebSocket server started")
}

// Stop gracefully shuts down the WebSocket server
func (s *Server) Stop() {
	s.hub.Stop()
	s.logger.Info().Msg("WebSocket server stopped")
}

// RegisterRoutes adds WebSocket routes to the router
func (s *Server) RegisterRoutes(router *mux.Router) {
	// REQ-017: WebSocket endpoint for streaming
	router.HandleFunc("/ws/ohlcv", s.handleWebSocket).Methods("GET")

	// Hub metrics endpoint
	router.HandleFunc("/api/v1/stream/metrics", s.handleMetrics).Methods("GET")

	s.logger.Info().Msg("WebSocket routes registered")
}

// handleWebSocket handles WebSocket connection upgrades
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	correlationID := r.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = generateClientID()
	}

	logger := s.logger.With().
		Str("correlation_id", correlationID).
		Str("remote_addr", r.RemoteAddr).
		Logger()

	logger.Info().Msg("WebSocket connection attempt")

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Create new client
	client := NewClient(conn, s.hub, logger)

	// Register client with hub
	s.hub.RegisterClient(client)

	// Start client goroutines with hub context (not request context)
	// Request context gets cancelled after upgrade, but client should persist
	client.Start(s.hub.ctx)
}

// handleMetrics returns WebSocket hub metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	clientCount, messageCount, subscriptionCount := s.hub.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response
	response := fmt.Sprintf(`{
		"clients": %d,
		"messages_sent": %d,
		"active_subscriptions": %d,
		"status": "healthy"
	}`, clientCount, messageCount, subscriptionCount)

	w.Write([]byte(response))
}

// GetHub returns the hub for external access
func (s *Server) GetHub() *Hub {
	return s.hub
}
