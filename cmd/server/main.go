package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/ridopark/jonbu-ohlcv/internal/config"
	"github.com/ridopark/jonbu-ohlcv/internal/database"
	"github.com/ridopark/jonbu-ohlcv/internal/fetcher/alpaca"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/ridopark/jonbu-ohlcv/internal/stream"
	"github.com/ridopark/jonbu-ohlcv/internal/worker"
	"github.com/ridopark/jonbu-ohlcv/pkg/api/handlers"
)

// REQ-016: HTTP server for API endpoints
// REQ-017: WebSocket server for real-time streaming
// REQ-026: Graceful shutdown handling
// REQ-031: High-performance event processing pipeline

// Server represents the main application server
type Server struct {
	// Core components
	config *config.Config
	logger zerolog.Logger
	db     *database.DB

	// Streaming components
	streamServer *stream.Server
	workerPool   *worker.Pool
	alpacaStream alpaca.StreamInterface

	// HTTP server
	httpServer *http.Server
	router     *mux.Router

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

func main() {
	// Initialize application
	server, err := initializeServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize server: %v\n", err)
		os.Exit(1)
	}

	// Start server
	if err := server.Start(); err != nil {
		server.logger.Fatal().Err(err).Msg("Failed to start server")
	}

	// Wait for shutdown signal
	server.WaitForShutdown()
}

// initializeServer creates and configures the server
func initializeServer() (*Server, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Environment, cfg.LogLevel)

	appLogger.Info().
		Str("version", "2.0.0").
		Str("phase", "Phase 2 - Real-time Streaming").
		Msg("Initializing jonbu-ohlcv server")

	// Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test database connection
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize streaming components
	streamServer := stream.NewServer(appLogger)

	// Initialize worker pool
	poolConfig := worker.DefaultPoolConfig()
	poolConfig.MaxWorkers = cfg.Worker.MaxWorkersPerSymbol * 10 // Scale for multiple symbols
	poolConfig.EventBufferSize = cfg.Worker.BufferSize
	poolConfig.UseMockMode = cfg.Alpaca.UseMock // Pass mock mode to workers
	workerPool := worker.NewPool(poolConfig, appLogger)

	// Initialize Alpaca stream client using factory
	streamFactory := alpaca.NewStreamClientFactory(cfg.Alpaca.UseMock)
	alpacaStream := streamFactory.Create(
		cfg.Alpaca.APIKey,
		cfg.Alpaca.SecretKey,
		cfg.Alpaca.BaseURL,
		appLogger,
	)

	// Create HTTP router
	router := mux.NewRouter()

	server := &Server{
		config:       cfg,
		logger:       appLogger,
		db:           db,
		streamServer: streamServer,
		workerPool:   workerPool,
		alpacaStream: alpacaStream,
		router:       router,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler:      server.router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	return server, nil
}

// setupRoutes configures all HTTP and WebSocket routes
func (s *Server) setupRoutes() {
	// REQ-041: CORS middleware
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// REQ-042: Request logging middleware
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)

			s.logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Dur("duration", time.Since(start)).
				Msg("HTTP request")
		})
	})

	// Health check endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// API routes
	apiRouter := s.router.PathPrefix("/api/v1").Subrouter()

	// OHLCV endpoints
	repo, err := database.NewOHLCVRepository(s.db)
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed to create OHLCV repository")
	}
	ohlcvHandler := handlers.NewOHLCVHandler(repo)
	apiRouter.HandleFunc("/ohlcv/{symbol}", ohlcvHandler.GetOHLCV).Methods("GET")
	apiRouter.HandleFunc("/ohlcv/{symbol}/history", ohlcvHandler.GetOHLCVHistory).Methods("GET")

	// Stream management endpoints
	apiRouter.HandleFunc("/stream/symbols", s.handleAddSymbols).Methods("POST")
	apiRouter.HandleFunc("/stream/symbols/{symbol}", s.handleRemoveSymbol).Methods("DELETE")
	apiRouter.HandleFunc("/stream/status", s.handleStreamStatus).Methods("GET")

	// Worker pool metrics
	apiRouter.HandleFunc("/workers/metrics", s.handleWorkerMetrics).Methods("GET")

	// Register WebSocket routes
	s.streamServer.RegisterRoutes(s.router)

	s.logger.Info().Msg("Routes configured")
}

// Start begins all server components
func (s *Server) Start() error {
	s.logger.Info().
		Str("address", s.httpServer.Addr).
		Msg("Starting server")

	// Start streaming components
	s.streamServer.Start()
	s.workerPool.Start()

	// Start Alpaca stream
	if err := s.alpacaStream.Start(); err != nil {
		s.logger.Error().Err(err).Msg("Failed to start Alpaca stream - continuing without streaming")
		// Don't fail server startup, just log the error
	} else {
		s.logger.Info().Msg("Alpaca stream started successfully")

		// Connect data pipeline: Alpaca → Worker Pool → WebSocket Hub
		go s.runDataPipeline()
	}

	// Start HTTP server
	go func() {
		s.logger.Info().Msg("HTTP server listening")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	return nil
}

// runDataPipeline connects the data flow from Alpaca to WebSocket clients
func (s *Server) runDataPipeline() {
	s.logger.Info().Msg("Starting data pipeline")

	// Create OHLCV repository for database storage
	repo, err := database.NewOHLCVRepository(s.db)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create OHLCV repository for pipeline")
		return
	}

	// Stream events from Alpaca to worker pool
	go func() {
		for event := range s.alpacaStream.GetOutput() {
			s.workerPool.ProcessEvent(event)
		}
	}()

	// Stream candles from worker pool to WebSocket clients AND database
	go func() {
		hub := s.streamServer.GetHub()
		for candle := range s.workerPool.GetCandleOutput() {
			// Broadcast to WebSocket clients
			hub.BroadcastCandle(candle.Symbol, candle.Interval, &candle)

			// Store in database for historical data
			go s.storeCandleToDatabase(repo, &candle)
		}
	}()
}

// WaitForShutdown waits for shutdown signals and handles graceful shutdown
func (s *Server) WaitForShutdown() {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info().Msg("Shutdown signal received")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("HTTP server shutdown error")
	}

	// Stop streaming components
	s.alpacaStream.Stop()
	s.workerPool.Stop()
	s.streamServer.Stop()

	// Close database connection
	if err := s.db.Close(); err != nil {
		s.logger.Error().Err(err).Msg("Database close error")
	}

	s.logger.Info().Msg("Server shutdown complete")
}

// storeCandleToDatabase persists aggregated candles to PostgreSQL
func (s *Server) storeCandleToDatabase(repo *database.OHLCVRepository, candle *models.Candle) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convert timeframe from internal format to database format
	dbTimeframe := s.convertTimeframeForDB(candle.Interval)

	// Convert candle to OHLCV model for database storage
	ohlcv := &models.OHLCV{
		Symbol:    candle.Symbol,
		Timestamp: candle.Timestamp,
		Open:      candle.Open,
		High:      candle.High,
		Low:       candle.Low,
		Close:     candle.Close,
		Volume:    candle.Volume,
		Timeframe: dbTimeframe, // Use converted timeframe
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert candle into database
	if err := repo.Insert(ctx, ohlcv); err != nil {
		s.logger.Error().Err(err).
			Str("symbol", candle.Symbol).
			Str("internal_timeframe", candle.Interval).
			Str("db_timeframe", dbTimeframe).
			Time("timestamp", candle.Timestamp).
			Msg("Failed to store candle to database")
	} else {
		s.logger.Debug().
			Str("symbol", candle.Symbol).
			Str("timeframe", dbTimeframe).
			Time("timestamp", candle.Timestamp).
			Float64("close", candle.Close).
			Msg("Successfully stored candle to database")
	}
}

// convertTimeframeForDB converts internal timeframe format to database format
func (s *Server) convertTimeframeForDB(timeframe string) string {
	switch timeframe {
	case "1min":
		return "1m"
	case "5min":
		return "5m"
	case "15min":
		return "15m"
	case "30min":
		return "30m"
	case "1hour":
		return "1h"
	case "4hour":
		return "4h"
	case "1day":
		return "1d"
	default:
		s.logger.Warn().
			Str("timeframe", timeframe).
			Msg("Unknown timeframe format, defaulting to 1m")
		return "1m"
	}
}

// HTTP Handlers

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "2.0.0",
		"phase":     "Phase 2 - Real-time Streaming",
	}

	// Check database health
	ctx := context.Background()
	if err := s.db.Ping(ctx); err != nil {
		status["status"] = "unhealthy"
		status["database"] = "disconnected"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		status["database"] = "connected"
	}

	// Add component status
	status["stream_server"] = "running"
	status["worker_pool"] = s.workerPool.GetStatus()
	status["alpaca_stream"] = s.alpacaStream.GetConnectionStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleAddSymbols(w http.ResponseWriter, r *http.Request) {
	// Parse request body for symbols to add
	var request struct {
		Symbols []string `json:"symbols"`
	}

	// Try to parse JSON body, fall back to default symbols if no body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Use default symbols for testing
		request.Symbols = []string{"AAPL", "GOOGL", "MSFT"}
		s.logger.Info().Msg("Using default symbols for testing")
	}

	if len(request.Symbols) == 0 {
		request.Symbols = []string{"AAPL", "GOOGL", "MSFT"}
	}

	// Add symbols to worker pool
	for _, symbol := range request.Symbols {
		timeframes := []string{"1min", "5min", "15min"}
		for _, timeframe := range timeframes {
			if err := s.workerPool.AddSymbol(symbol, timeframe); err != nil {
				s.logger.Error().Err(err).
					Str("symbol", symbol).
					Str("timeframe", timeframe).
					Msg("Failed to add symbol worker")
			}
		}
	}

	// Subscribe to Alpaca stream
	if err := s.alpacaStream.Subscribe(request.Symbols); err != nil {
		s.logger.Error().Err(err).Msg("Failed to subscribe to Alpaca stream")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to subscribe to Alpaca stream",
			"error":   err.Error(),
		})
		return
	}

	s.logger.Info().
		Strs("symbols", request.Symbols).
		Msg("Successfully added symbols and subscribed to stream")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"symbols": request.Symbols,
		"message": "Symbols added and subscribed to stream",
	})
}

func (s *Server) handleRemoveSymbol(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	// Remove from worker pool
	timeframes := []string{"1min", "5min", "15min"}
	for _, timeframe := range timeframes {
		if err := s.workerPool.RemoveSymbol(symbol, timeframe); err != nil {
			s.logger.Error().Err(err).
				Str("symbol", symbol).
				Str("timeframe", timeframe).
				Msg("Failed to remove symbol worker")
		}
	}

	// Unsubscribe from Alpaca stream
	if err := s.alpacaStream.Unsubscribe([]string{symbol}); err != nil {
		s.logger.Error().Err(err).Msg("Failed to unsubscribe from Alpaca stream")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "success", "symbol": "%s"}`, symbol)
}

func (s *Server) handleStreamStatus(w http.ResponseWriter, r *http.Request) {
	clientCount, messageCount, subscriptionCount := s.streamServer.GetHub().GetMetrics()

	response := map[string]interface{}{
		"websocket_clients":    clientCount,
		"messages_sent":        messageCount,
		"active_subscriptions": subscriptionCount,
		"worker_pool":          s.workerPool.GetMetrics(),
		"alpaca_connection":    s.alpacaStream.GetConnectionStatus(),
		"status":               "active",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleWorkerMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.workerPool.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"worker_metrics": "active", "total_workers": %v}`,
		metrics["active_workers"])
}
