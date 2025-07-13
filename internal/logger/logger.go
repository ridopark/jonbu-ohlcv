package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// REQ-046: All components MUST use zerolog for structured logging
// REQ-047: All operations MUST include correlation IDs for tracing
// REQ-048: All errors MUST be logged with appropriate context
// REQ-049: All performance metrics MUST be logged for monitoring
// REQ-050: Log levels MUST be configurable per environment

// InitLogger initializes the global logger with the specified configuration
func InitLogger(level string, environment string) {
	// Set global time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Configure log level
	logLevel := parseLogLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Configure output format based on environment
	if environment == "development" {
		// Human-readable console output for development
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).With().
			Timestamp().
			Caller().
			Logger()
	} else {
		// JSON output for production
		log.Logger = log.With().
			Timestamp().
			Caller().
			Logger()
	}

	log.Info().
		Str("level", level).
		Str("environment", environment).
		Msg("Logger initialized")
}

// New creates a new configured logger instance
func New(environment, level string) zerolog.Logger {
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Configure log level
	logLevel := parseLogLevel(level)

	// Configure output based on environment
	var logger zerolog.Logger

	if environment == "production" {
		// JSON output for production
		logger = zerolog.New(os.Stdout).
			Level(logLevel).
			With().
			Timestamp().
			Str("service", "jonbu-ohlcv").
			Logger()
	} else {
		// Pretty console output for development
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).
			Level(logLevel).
			With().
			Timestamp().
			Str("service", "jonbu-ohlcv").
			Logger()
	}

	return logger
}

// NewContextLogger creates a new logger with context fields
func NewContextLogger(component string) zerolog.Logger {
	return log.With().
		Str("component", component).
		Logger()
}

// NewRequestLogger creates a logger for HTTP requests with correlation ID
func NewRequestLogger(correlationID, method, path string) zerolog.Logger {
	return log.With().
		Str("correlation_id", correlationID).
		Str("method", method).
		Str("path", path).
		Str("component", "http").
		Logger()
}

// NewServiceLogger creates a logger for service operations
func NewServiceLogger(service, operation string) zerolog.Logger {
	return log.With().
		Str("service", service).
		Str("operation", operation).
		Logger()
}

// NewWorkerLogger creates a logger for worker processes
func NewWorkerLogger(workerType, symbol string) zerolog.Logger {
	return log.With().
		Str("worker_type", workerType).
		Str("symbol", symbol).
		Str("component", "worker").
		Logger()
}

// LogPerformance logs performance metrics for monitoring
func LogPerformance(logger zerolog.Logger, operation string, start time.Time, success bool) {
	duration := time.Since(start)

	event := logger.Info()
	if !success {
		event = logger.Error()
	}

	event.
		Str("operation", operation).
		Dur("duration", duration).
		Bool("success", success).
		Msg("Performance metric")
}

// LogError logs errors with proper context and wrapping
func LogError(logger zerolog.Logger, err error, message string, fields map[string]interface{}) {
	event := logger.Error().Err(err)

	for key, value := range fields {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case int64:
			event = event.Int64(key, v)
		case float64:
			event = event.Float64(key, v)
		case bool:
			event = event.Bool(key, v)
		case time.Duration:
			event = event.Dur(key, v)
		default:
			event = event.Interface(key, v)
		}
	}

	event.Msg(message)
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
