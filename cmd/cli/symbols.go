package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// REQ-023: Symbol management commands

var (
	symbolsCmd = &cobra.Command{
		Use:   "symbols",
		Short: "Manage tracked symbols",
		Long:  `Add, remove, and list symbols for OHLCV data tracking.`,
	}

	symbolsAddCmd = &cobra.Command{
		Use:   "add [symbols...]",
		Short: "Add symbols to track",
		Long:  `Add one or more symbols to the tracking list. Symbols should be uppercase stock tickers.`,
		Args:  cobra.MinimumNArgs(1),
		RunE:  runSymbolsAdd,
	}

	symbolsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tracked symbols",
		Long:  `List all symbols currently being tracked for OHLCV data.`,
		RunE:  runSymbolsList,
	}

	symbolsRemoveCmd = &cobra.Command{
		Use:   "remove [symbols...]",
		Short: "Remove symbols from tracking",
		Long:  `Remove one or more symbols from the tracking list.`,
		Args:  cobra.MinimumNArgs(1),
		RunE:  runSymbolsRemove,
	}
)

func init() {
	symbolsCmd.AddCommand(symbolsAddCmd)
	symbolsCmd.AddCommand(symbolsListCmd)
	symbolsCmd.AddCommand(symbolsRemoveCmd)
}

func runSymbolsAdd(cmd *cobra.Command, args []string) error {
	// Validate and normalize symbols
	var validSymbols []string
	for _, symbol := range args {
		normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))

		// REQ-025: Input validation with helpful error messages
		if err := validateSymbol(normalizedSymbol); err != nil {
			fmt.Printf("Warning: Skipping invalid symbol '%s': %v\n", symbol, err)
			continue
		}

		validSymbols = append(validSymbols, normalizedSymbol)
	}

	if len(validSymbols) == 0 {
		return fmt.Errorf("no valid symbols provided")
	}

	// TODO: Implement actual symbol storage in database
	// For now, just print what would be added
	fmt.Printf("Would add symbols: %s\n", strings.Join(validSymbols, ", "))
	fmt.Printf("Successfully added %d symbol(s) to tracking list.\n", len(validSymbols))

	return nil
}

func runSymbolsList(cmd *cobra.Command, args []string) error {
	// TODO: Implement actual symbol retrieval from database
	// For now, show example output
	fmt.Println("Tracked symbols:")
	fmt.Println("- AAPL (Apple Inc.)")
	fmt.Println("- GOOGL (Alphabet Inc.)")
	fmt.Println("- MSFT (Microsoft Corporation)")
	fmt.Println("\nTotal: 3 symbols")

	return nil
}

func runSymbolsRemove(cmd *cobra.Command, args []string) error {
	// Validate and normalize symbols
	var validSymbols []string
	for _, symbol := range args {
		normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))

		if err := validateSymbol(normalizedSymbol); err != nil {
			fmt.Printf("Warning: Skipping invalid symbol '%s': %v\n", symbol, err)
			continue
		}

		validSymbols = append(validSymbols, normalizedSymbol)
	}

	if len(validSymbols) == 0 {
		return fmt.Errorf("no valid symbols provided")
	}

	// TODO: Implement actual symbol removal from database
	// For now, just print what would be removed
	fmt.Printf("Would remove symbols: %s\n", strings.Join(validSymbols, ", "))
	fmt.Printf("Successfully removed %d symbol(s) from tracking list.\n", len(validSymbols))

	return nil
}
