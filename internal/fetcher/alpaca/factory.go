package alpaca

import (
	"github.com/ridopark/jonbu-ohlcv/internal/models"
	"github.com/rs/zerolog"
)

// StreamInterface defines the common interface for both real and mock stream clients
type StreamInterface interface {
	Start() error
	Stop()
	Subscribe(symbols []string) error
	Unsubscribe(symbols []string) error
	GetOutput() <-chan models.MarketEvent
	GetConnectionStatus() map[string]interface{}
}

// StreamClientFactory creates either a real or mock stream client
type StreamClientFactory struct {
	useMock bool
}

// NewStreamClientFactory creates a new factory
func NewStreamClientFactory(useMock bool) *StreamClientFactory {
	return &StreamClientFactory{
		useMock: useMock,
	}
}

// Create returns either a real or mock stream client
func (f *StreamClientFactory) Create(apiKey, secretKey, baseURL string, logger zerolog.Logger) StreamInterface {
	if f.useMock {
		logger.Info().Msg("Creating mock Alpaca stream client")
		return NewMockStreamClient(apiKey, secretKey, baseURL, logger)
	}

	logger.Info().Msg("Creating real Alpaca stream client")
	return NewStreamClient(apiKey, secretKey, baseURL, logger)
}

// SetMockMode enables or disables mock mode
func (f *StreamClientFactory) SetMockMode(useMock bool) {
	f.useMock = useMock
}

// IsMockMode returns whether mock mode is enabled
func (f *StreamClientFactory) IsMockMode() bool {
	return f.useMock
}
