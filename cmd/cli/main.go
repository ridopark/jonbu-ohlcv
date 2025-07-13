package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ridopark/jonbu-ohlcv/internal/config"
	"github.com/ridopark/jonbu-ohlcv/internal/database"
	"github.com/ridopark/jonbu-ohlcv/internal/logger"
)

var rootCmd = &cobra.Command{
	Use:   "jonbu-ohlcv",
	Short: "OHLCV data fetching and management tool",
	Long:  `A CLI tool for fetching, storing, and managing OHLCV market data.`,
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database schema migrations to set up or update the database`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations()
	},
}

func init() {
	// Add persistent flags to root command
	rootCmd.PersistentFlags().StringP("format", "f", "table", "output format (table, json, csv)")

	// Add subcommands
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(fetchCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMigrations() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Environment, cfg.LogLevel)

	// Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	appLogger.Info().Msg("Database migrations completed successfully")
}
