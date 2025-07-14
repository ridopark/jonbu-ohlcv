package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ridopark/jonbu-ohlcv/internal/fetcher/alpaca"
	"github.com/rs/zerolog"
)

func main() {
	// Set up logging
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Set environment variables for testing
	os.Setenv("ALPACA_MOCK_SPEED_MULTIPLIER", "12.0") // 12x speed
	os.Setenv("ALPACA_MOCK_CANDLE_INTERVAL_SEC", "5") // 5 seconds per candle

	fmt.Println("🚀 Testing Mock OHLCV Candle Generation")
	fmt.Println("📊 Configuration:")
	fmt.Println("   - Speed: 12x (1-minute candles every 5 seconds)")
	fmt.Println("   - Symbols: AAPL, TSLA, NVDA")
	fmt.Println("   - Duration: 30 seconds")
	fmt.Println("")

	// Create mock client
	client := alpaca.NewMockStreamClient("test", "test", "test", logger)

	// Start the client
	if err := client.Start(); err != nil {
		fmt.Printf("❌ Error starting mock client: %v\n", err)
		return
	}
	defer client.Stop()

	// Subscribe to test symbols
	symbols := []string{"AAPL", "TSLA", "NVDA"}
	if err := client.Subscribe(symbols); err != nil {
		fmt.Printf("❌ Error subscribing to symbols: %v\n", err)
		return
	}

	fmt.Printf("✅ Subscribed to symbols: %v\n", symbols)
	fmt.Println("📈 Generated OHLCV Candles:")
	fmt.Println("")

	// Listen for events
	eventCount := 0
	timeout := time.After(30 * time.Second)

	for {
		select {
		case event := <-client.GetOutput():
			eventCount++
			fmt.Printf("[%02d] %s | %s | Price: $%.2f | Volume: %d | %s\n",
				eventCount,
				event.Timestamp.Format("15:04:05"),
				event.Symbol,
				event.Price,
				event.Volume,
				event.Type,
			)

		case <-timeout:
			fmt.Println("")
			fmt.Printf("⏰ Test completed after 30 seconds\n")
			fmt.Printf("📊 Total candles generated: %d\n", eventCount)
			fmt.Printf("🔥 Expected ~18 candles (3 symbols × 6 intervals)\n")

			if eventCount >= 15 {
				fmt.Println("✅ Mock OHLCV generation working correctly!")
			} else {
				fmt.Println("⚠️  Lower than expected candle count")
			}
			return
		}
	}
}
