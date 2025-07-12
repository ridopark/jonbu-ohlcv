package main

import (
	"fmt"
	"regexp"
	"strings"
)

// REQ-025: Input validation helper functions

// validateSymbol validates a stock symbol format
func validateSymbol(symbol string) error {
	// Basic validation - uppercase letters, 1-5 characters
	symbolRegex := regexp.MustCompile(`^[A-Z]{1,5}$`)

	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if !symbolRegex.MatchString(symbol) {
		return fmt.Errorf("symbol must be 1-5 uppercase letters")
	}

	return nil
}

// validateTimeframe validates timeframe parameter
func validateTimeframe(timeframe string) error {
	validTimeframes := map[string]bool{
		"1min":  true,
		"5min":  true,
		"15min": true,
		"30min": true,
		"1hour": true,
		"1day":  true,
	}

	if !validTimeframes[timeframe] {
		return fmt.Errorf("invalid timeframe: %s (valid: 1min, 5min, 15min, 30min, 1hour, 1day)", timeframe)
	}

	return nil
}

// validateOutputFormat validates output format parameter
func validateOutputFormat(format string) error {
	validFormats := map[string]bool{
		"table": true,
		"json":  true,
		"csv":   true,
	}

	format = strings.ToLower(format)
	if !validFormats[format] {
		return fmt.Errorf("invalid format: %s (valid: table, json, csv)", format)
	}

	return nil
}
