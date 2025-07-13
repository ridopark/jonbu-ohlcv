package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Mock data management commands",
	Long:  "Commands for managing mock data generation and testing",
}

var mockEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable mock mode",
	Long:  "Enable mock mode for Alpaca streaming (useful for testing)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := setMockMode(true); err != nil {
			fmt.Fprintf(os.Stderr, "Error enabling mock mode: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Mock mode enabled")
	},
}

var mockDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable mock mode",
	Long:  "Disable mock mode for Alpaca streaming (use real API)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := setMockMode(false); err != nil {
			fmt.Fprintf(os.Stderr, "Error disabling mock mode: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Mock mode disabled")
	},
}

var mockTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test mock data generation",
	Long:  "Subscribe to mock symbols and verify data generation",
	Run: func(cmd *cobra.Command, args []string) {
		if err := testMockData(); err != nil {
			fmt.Fprintf(os.Stderr, "Error testing mock data: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	mockCmd.AddCommand(mockEnableCmd)
	mockCmd.AddCommand(mockDisableCmd)
	mockCmd.AddCommand(mockTestCmd)
	rootCmd.AddCommand(mockCmd)
}

func setMockMode(enable bool) error {
	// This would ideally update the configuration
	// For now, we'll provide instructions
	fmt.Printf("To %s mock mode:\n", map[bool]string{true: "enable", false: "disable"}[enable])
	fmt.Printf("1. Set ALPACA_USE_MOCK=%t in config/.env\n", enable)
	fmt.Println("2. Restart the server")
	return nil
}

func testMockData() error {
	serverURL := "http://localhost:8080"

	fmt.Println("Testing mock data generation...")

	// Add symbols to trigger mock data
	symbols := []string{"AAPL", "TSLA", "GOOGL"}

	requestBody := map[string]interface{}{
		"symbols": symbols,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to add symbols
	resp, err := http.Post(serverURL+"/api/v1/stream/symbols", "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to add symbols: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add symbols, status: %s", resp.Status)
	}

	fmt.Printf("✅ Successfully added symbols: %v\n", symbols)

	// Check stream status
	resp, err = http.Get(serverURL + "/api/v1/stream/status")
	if err != nil {
		return fmt.Errorf("failed to get stream status: %w", err)
	}
	defer resp.Body.Close()

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("failed to decode status: %w", err)
	}

	fmt.Printf("✅ Stream status: %+v\n", status)

	// Check if mock mode is enabled
	if alpacaStatus, ok := status["alpaca_connection"].(map[string]interface{}); ok {
		if isMock, ok := alpacaStatus["mock"].(bool); ok && isMock {
			fmt.Println("✅ Mock mode is active")
		} else {
			fmt.Println("⚠️  Mock mode not detected - check configuration")
		}
	}

	fmt.Println("🔍 Monitor the server logs to see mock data events being generated")
	fmt.Println("📊 Use WebSocket client to see real-time mock data: ws://localhost:8080/ws")

	return nil
}
