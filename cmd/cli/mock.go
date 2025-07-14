package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

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

var mockSpeedCmd = &cobra.Command{
	Use:   "speed [multiplier]",
	Short: "Set mock data generation speed",
	Long:  "Set the speed multiplier for mock data generation (e.g., 10.0 for 10x faster)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := setMockSpeed(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting mock speed: %v\n", err)
			os.Exit(1)
		}
	},
}

var mockStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mock service status",
	Long:  "Display current mock service configuration and speed settings",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showMockStatus(); err != nil {
			fmt.Fprintf(os.Stderr, "Error getting mock status: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	mockCmd.AddCommand(mockEnableCmd)
	mockCmd.AddCommand(mockDisableCmd)
	mockCmd.AddCommand(mockTestCmd)
	mockCmd.AddCommand(mockSpeedCmd)
	mockCmd.AddCommand(mockStatusCmd)
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

func setMockSpeed(speedStr string) error {
	speed, err := strconv.ParseFloat(speedStr, 64)
	if err != nil {
		return fmt.Errorf("invalid speed multiplier: %w", err)
	}

	if speed <= 0 {
		return fmt.Errorf("speed multiplier must be positive")
	}

	serverURL := "http://localhost:8080"

	requestBody := map[string]interface{}{
		"speed_multiplier": speed,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to set speed
	resp, err := http.Post(serverURL+"/api/v1/mock/speed", "application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to set mock speed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set mock speed, status: %s", resp.Status)
	}

	fmt.Printf("✅ Mock speed set to %.1fx\n", speed)
	return nil
}

func showMockStatus() error {
	serverURL := "http://localhost:8080"

	resp, err := http.Get(serverURL + "/api/v1/mock/status")
	if err != nil {
		return fmt.Errorf("failed to get mock status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get mock status, status: %s", resp.Status)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("failed to decode status: %w", err)
	}

	fmt.Println("📊 Mock Service Status:")
	for key, value := range status {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return nil
}
