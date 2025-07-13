#!/bin/bash

# Create a named pipe for two-way communication
pipe="/tmp/websocket_pipe"
mkfifo "$pipe" 2>/dev/null || true

echo "Starting WebSocket connection to ws://localhost:8080/ws/ohlcv"
echo "Sending subscription for AAPL 1min..."

# Start websocat in the background with the pipe
cat "$pipe" | websocat --text ws://localhost:8080/ws/ohlcv &
websocket_pid=$!

# Send subscription message
echo '{"type":"subscription","symbol":"AAPL","timeframe":"1min","action":"subscribe"}' > "$pipe"

echo "Subscription sent. Listening for candle data..."
echo "Press Ctrl+C to stop"

# Keep the connection alive and display any output
sleep 10

# Clean up
kill $websocket_pid 2>/dev/null
rm -f "$pipe"
