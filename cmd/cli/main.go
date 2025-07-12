package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ridopark/jonbu-ohlcv/internal/config"
	"github.com/ridopark/jonbu-ohlcv/internal/database"
	"github.com/ridopark/jonbu-ohlcv/internal/fetcher/alpaca"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
)

// REQ-021: CLI for historical data fetching
// REQ-022: Multiple output formats support
// REQ-023: Symbol management commands
// REQ-024: Database migration commands
// REQ-025: Input validation and helpful error messages

var (
	rootCmd = &cobra.Command{
		Use:   "jonbu-ohlcv",
		Short: "OHLCV data fetching and management tool",
		Long:  `A CLI tool for fetching, storing, and managing OHLCV (Open, High, Low, Close, Volume) market data.`,
	}

	// Global flags
	configFile string
	logLevel   string
	format     string
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is config/.env)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "table", "output format (table, json, csv)")

	// Add subcommands
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(symbolsCmd)
	rootCmd.AddCommand(migrateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// initializeApp initializes the application configuration and dependencies
func initializeApp() (*config.Config, *database.DB, *alpaca.AlpacaProvider, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override log level if specified
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Initialize logger
	logger.InitLogger(cfg.LogLevel, cfg.Environment)

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize Alpaca provider
	provider := alpaca.NewAlpacaProvider(cfg.Alpaca)
	if err := provider.Connect(context.Background()); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to Alpaca: %w", err)
	}

	return cfg, db, provider, nil
}

// validateDateString validates a date string in YYYY-MM-DD format
func validateDateString(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date cannot be empty")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: use YYYY-MM-DD")
	}

	return date, nil
}
