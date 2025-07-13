#!/bin/bash

# Simple multi-symbol WebSocket test with websocat
echo "🚀 Testing multi-symbol WebSocket subscriptions with websocat"
echo ""

# Method 1: Send multiple subscription messages sequentially
echo "Method 1: Sequential subscriptions"
echo ""

(
    echo '{"type":"subscription","symbol":"AAPL","timeframe":"1min","action":"subscribe"}'
    echo '{"type":"subscription","symbol":"GOOGL","timeframe":"1min","action":"subscribe"}'
    echo '{"type":"subscription","symbol":"MSFT","timeframe":"1min","action":"subscribe"}'
    sleep 60  # Wait for some candles
) | websocat --text ws://localhost:8080/ws/ohlcv

echo ""
echo "Test completed!"
