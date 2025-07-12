package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano}).
		With().
		Timestamp().
		Logger()
}

func main() {
	initLogger()

	// Basic logging
	log.Info().Msg("Logger initialized with zerolog")

	// Structured logging with fields
	log.Info().
		Str("service", "jonbu-ohlcv").
		Str("version", "1.0.0").
		Msg("Service starting")

	// Simulated error logging
	log.Error().
		Str("symbol", "AAPL").
		Str("operation", "fetch").
		Msg("Simulated error for testing")

	// Performance logging
	start := time.Now()
	time.Sleep(10 * time.Millisecond) // Simulate work

	log.Info().
		Str("operation", "data_fetch").
		Dur("duration", time.Since(start)).
		Int("records", 1000).
		Msg("Operation completed")
}
