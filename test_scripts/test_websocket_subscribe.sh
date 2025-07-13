#!/bin/bash

echo "🔗 Connecting to WebSocket and subscribing to AAPL 1min data..."
echo "📊 Streaming real-time mock OHLCV data:"
echo "🟢 Press Ctrl+C to disconnect"
echo ""

# Create a named pipe for bidirectional communication
PIPE=$(mktemp -u)
mkfifo $PIPE

# Start websocat in background, reading from pipe
websocat ws://localhost:8080/ws/ohlcv < $PIPE &
WEBSOCAT_PID=$!

# Send subscription message
echo '{"action": "subscribe", "symbol": "AAPL", "timeframe": "1min"}' > $PIPE &

# Keep pipe open and wait for user to exit
echo "Subscribed to AAPL 1min data. Streaming should start now..."
wait $WEBSOCAT_PID

# Cleanup
rm -f $PIPE
