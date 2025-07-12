package alpaca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"

	"github.com/ridopark/jonbu-ohlcv/internal/config"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-029: Data provider abstraction interface
type MarketDataProvider interface {
	Connect(ctx context.Context) error
	GetHistoricalOHLCV(ctx context.Context, symbol, timeframe string, start, end time.Time) ([]*models.OHLCV, error)
	ValidateSymbol(symbol string) error
	Close() error
}

// REQ-001, REQ-002: Alpaca API implementation
type AlpacaProvider struct {
	cfg        config.AlpacaConfig
	httpClient *http.Client
	logger     zerolog.Logger
	baseURL    string
}

// AlpacaBar represents Alpaca's bar data format
type AlpacaBar struct {
	Symbol    string    `json:"S"`
	Timestamp time.Time `json:"t"`
	Open      float64   `json:"o"`
	High      float64   `json:"h"`
	Low       float64   `json:"l"`
	Close     float64   `json:"c"`
	Volume    int64     `json:"v"`
}

// AlpacaBarsResponse represents the response from Alpaca bars API
type AlpacaBarsResponse struct {
	Bars     map[string][]AlpacaBar `json:"bars"`
	NextPage string                 `json:"next_page_token,omitempty"`
}

// NewAlpacaProvider creates a new Alpaca data provider
func NewAlpacaProvider(cfg config.AlpacaConfig) *AlpacaProvider {
	return &AlpacaProvider{
		cfg:     cfg,
		logger:  logger.NewContextLogger("alpaca_provider"),
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// REQ-003: Connect with validation and error handling
func (a *AlpacaProvider) Connect(ctx context.Context) error {
	a.logger.Info().Msg("Connecting to Alpaca API")

	// Validate configuration
	if a.cfg.APIKey == "" || a.cfg.SecretKey == "" {
		return fmt.Errorf("Alpaca API credentials are required")
	}

	// Test connection with account endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/v2/account", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	a.setAuthHeaders(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Alpaca API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Alpaca API connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	a.logger.Info().
		Str("base_url", a.baseURL).
		Bool("is_paper", a.cfg.IsPaper).
		Msg("Successfully connected to Alpaca API")

	return nil
}

// REQ-001: Get historical OHLCV data from Alpaca
func (a *AlpacaProvider) GetHistoricalOHLCV(ctx context.Context, symbol, timeframe string, start, end time.Time) ([]*models.OHLCV, error) {
	startTime := time.Now()
	defer func() {
		logger.LogPerformance(a.logger, "get_historical_ohlcv", startTime, true)
	}()

	// REQ-005: Validate input parameters
	if err := a.ValidateSymbol(symbol); err != nil {
		return nil, fmt.Errorf("invalid symbol: %w", err)
	}

	if err := validateTimeframe(timeframe); err != nil {
		return nil, fmt.Errorf("invalid timeframe: %w", err)
	}

	// Build request URL
	requestURL := fmt.Sprintf("%s/v2/stocks/%s/bars", a.baseURL, symbol)

	params := url.Values{}
	params.Set("timeframe", convertTimeframe(timeframe))
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("limit", "10000")
	params.Set("adjustment", "raw")

	fullURL := requestURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.setAuthHeaders(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from Alpaca: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Alpaca API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response AlpacaBarsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Alpaca response: %w", err)
	}

	// Convert Alpaca bars to OHLCV models
	symbolBars, exists := response.Bars[symbol]
	if !exists {
		a.logger.Warn().Str("symbol", symbol).Msg("No bars found for symbol")
		return []*models.OHLCV{}, nil
	}

	ohlcvs := make([]*models.OHLCV, 0, len(symbolBars))
	for _, bar := range symbolBars {
		// REQ-005: Validate incoming data
		if err := a.validateBarData(&bar); err != nil {
			a.logger.Warn().
				Err(err).
				Str("symbol", symbol).
				Time("timestamp", bar.Timestamp).
				Msg("Skipping invalid bar data")
			continue
		}

		ohlcv := &models.OHLCV{
			Symbol:    symbol,
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
			Timeframe: timeframe,
		}

		ohlcvs = append(ohlcvs, ohlcv)
	}

	a.logger.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Int("count", len(ohlcvs)).
		Time("start", start).
		Time("end", end).
		Msg("Successfully fetched historical OHLCV data")

	return ohlcvs, nil
}

// REQ-080: Validate symbol format
func (a *AlpacaProvider) ValidateSymbol(symbol string) error {
	if symbol == "" {
		return models.ErrInvalidSymbol
	}

	// Symbol should be uppercase letters only, 1-10 characters
	if len(symbol) > 10 {
		return fmt.Errorf("symbol too long: maximum 10 characters")
	}

	for _, char := range symbol {
		if char < 'A' || char > 'Z' {
			return fmt.Errorf("symbol must contain only uppercase letters")
		}
	}

	return nil
}

// Close cleanup resources
func (a *AlpacaProvider) Close() error {
	a.logger.Info().Msg("Closing Alpaca provider")
	// HTTP client doesn't require explicit closing
	return nil
}

// setAuthHeaders sets the required authentication headers for Alpaca API
func (a *AlpacaProvider) setAuthHeaders(req *http.Request) {
	req.Header.Set("APCA-API-KEY-ID", a.cfg.APIKey)
	req.Header.Set("APCA-API-SECRET-KEY", a.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")
}

// validateTimeframe validates the timeframe parameter
func validateTimeframe(timeframe string) error {
	validTimeframes := map[string]bool{
		"1m":  true,
		"5m":  true,
		"15m": true,
		"1h":  true,
		"4h":  true,
		"1d":  true,
	}

	if !validTimeframes[timeframe] {
		return models.ErrInvalidTimeframe
	}

	return nil
}

// convertTimeframe converts our timeframe format to Alpaca's format
func convertTimeframe(timeframe string) string {
	switch timeframe {
	case "1m":
		return "1Min"
	case "5m":
		return "5Min"
	case "15m":
		return "15Min"
	case "1h":
		return "1Hour"
	case "4h":
		return "4Hour"
	case "1d":
		return "1Day"
	default:
		return "1Min"
	}
}

// REQ-005: Validate bar data integrity
func (a *AlpacaProvider) validateBarData(bar *AlpacaBar) error {
	if bar.Symbol == "" {
		return fmt.Errorf("empty symbol")
	}

	if bar.Open <= 0 || bar.High <= 0 || bar.Low <= 0 || bar.Close <= 0 {
		return fmt.Errorf("invalid prices: all prices must be positive")
	}

	if bar.High < bar.Low {
		return fmt.Errorf("invalid price range: high (%.4f) < low (%.4f)", bar.High, bar.Low)
	}

	if bar.Volume < 0 {
		return fmt.Errorf("invalid volume: volume cannot be negative")
	}

	if bar.Timestamp.IsZero() {
		return fmt.Errorf("invalid timestamp: timestamp cannot be zero")
	}

	return nil
}
