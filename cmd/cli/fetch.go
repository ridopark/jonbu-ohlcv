package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/ridopark/jonbu-ohlcv/internal/database"
	"github.com/ridopark/jonbu-ohlcv/internal/models"
)

// REQ-021: CLI for historical data fetching
// REQ-022: Multiple output formats (JSON, table, CSV)

var (
	fetchCmd = &cobra.Command{
		Use:   "fetch [symbol]",
		Short: "Fetch historical OHLCV data",
		Long:  `Fetch historical OHLCV data for a given symbol and store it in the database.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runFetch,
	}

	// Fetch command flags
	timeframe string
	startDate string
	endDate   string
	store     bool
)

func init() {
	fetchCmd.Flags().StringVar(&timeframe, "timeframe", "1d", "timeframe (1m, 5m, 15m, 1h, 4h, 1d)")
	fetchCmd.Flags().StringVar(&startDate, "start", "", "start date (YYYY-MM-DD)")
	fetchCmd.Flags().StringVar(&endDate, "end", "", "end date (YYYY-MM-DD)")
	fetchCmd.Flags().BoolVar(&store, "store", false, "store data in database")
}

func runFetch(cmd *cobra.Command, args []string) error {
	symbol := strings.ToUpper(args[0])

	// REQ-025: Input validation with helpful error messages
	if err := validateSymbol(symbol); err != nil {
		return fmt.Errorf("invalid symbol '%s': %w", symbol, err)
	}

	if err := validateTimeframe(timeframe); err != nil {
		return fmt.Errorf("invalid timeframe '%s': %w", timeframe, err)
	}

	// Parse dates
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = validateDateString(startDate)
		if err != nil {
			return fmt.Errorf("invalid start date '%s': %w", startDate, err)
		}
	} else {
		// Default to 30 days ago
		start = time.Now().AddDate(0, 0, -30)
	}

	if endDate != "" {
		end, err = validateDateString(endDate)
		if err != nil {
			return fmt.Errorf("invalid end date '%s': %w", endDate, err)
		}
	} else {
		end = time.Now()
	}

	if start.After(end) {
		return fmt.Errorf("start date must be before end date")
	}

	// Initialize application
	_, db, provider, err := initializeApp()
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}
	defer db.Close()
	defer provider.Close()

	fmt.Printf("Fetching OHLCV data for %s (%s) from %s to %s...\n",
		symbol, timeframe, start.Format("2006-01-02"), end.Format("2006-01-02"))

	// Fetch data from Alpaca
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ohlcvs, err := provider.GetHistoricalOHLCV(ctx, symbol, timeframe, start, end)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	if len(ohlcvs) == 0 {
		fmt.Println("No data found for the specified period.")
		return nil
	}

	// Store in database if requested
	if store {
		repo, err := database.NewOHLCVRepository(db)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}
		defer repo.Close()

		if err := repo.InsertBatch(ctx, ohlcvs); err != nil {
			return fmt.Errorf("failed to store data: %w", err)
		}

		fmt.Printf("Successfully stored %d records in database.\n", len(ohlcvs))
	}

	// Display results
	return displayOHLCVData(ohlcvs, format)
}

// displayOHLCVData displays OHLCV data in the specified format
func displayOHLCVData(ohlcvs []*models.OHLCV, format string) error {
	switch format {
	case "json":
		return displayJSON(ohlcvs)
	case "csv":
		return displayCSV(ohlcvs)
	case "table":
		return displayTable(ohlcvs)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, csv)", format)
	}
}

func displayJSON(ohlcvs []*models.OHLCV) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ohlcvs)
}

func displayCSV(ohlcvs []*models.OHLCV) error {
	fmt.Println("Symbol,Timestamp,Open,High,Low,Close,Volume,Timeframe")
	for _, ohlcv := range ohlcvs {
		fmt.Printf("%s,%s,%.4f,%.4f,%.4f,%.4f,%d,%s\n",
			ohlcv.Symbol,
			ohlcv.Timestamp.Format("2006-01-02T15:04:05Z"),
			ohlcv.Open,
			ohlcv.High,
			ohlcv.Low,
			ohlcv.Close,
			ohlcv.Volume,
			ohlcv.Timeframe,
		)
	}
	return nil
}

func displayTable(ohlcvs []*models.OHLCV) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "SYMBOL\tTIMESTAMP\tOPEN\tHIGH\tLOW\tCLOSE\tVOLUME\tTIMEFRAME")
	fmt.Fprintln(w, "------\t---------\t----\t----\t---\t-----\t------\t---------")

	// Print data rows
	for _, ohlcv := range ohlcvs {
		fmt.Fprintf(w, "%s\t%s\t%.4f\t%.4f\t%.4f\t%.4f\t%s\t%s\n",
			ohlcv.Symbol,
			ohlcv.Timestamp.Format("2006-01-02 15:04"),
			ohlcv.Open,
			ohlcv.High,
			ohlcv.Low,
			ohlcv.Close,
			formatVolume(ohlcv.Volume),
			ohlcv.Timeframe,
		)
	}

	return nil
}

// formatVolume formats volume for better readability
func formatVolume(volume int64) string {
	if volume >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(volume)/1000000)
	} else if volume >= 1000 {
		return fmt.Sprintf("%.1fK", float64(volume)/1000)
	}
	return strconv.FormatInt(volume, 10)
}
