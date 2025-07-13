#!/bin/bash

echo "🔗 Connecting to WebSocket server..."
echo "📊 You should see real-time OHLCV data streaming below:"
echo "🟢 Press Ctrl+C to disconnect"
echo ""

# Use websocat with explicit text protocol
websocat --text ws://localhost:8080/ws/ohlcv
