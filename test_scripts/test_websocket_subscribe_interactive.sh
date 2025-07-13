#!/bin/bash

# Test WebSocket connection with multiple symbol subscriptions
echo "Testing WebSocket connection with multi-symbol subscriptions..."

# Create subscription messages for multiple symbols
cat > /tmp/subscribe_aapl.json << 'EOF'
{"type": "subscription", "symbol": "AAPL", "timeframe": "1min", "action": "subscribe"}
EOF

cat > /tmp/subscribe_googl.json << 'EOF'
{"type": "subscription", "symbol": "GOOGL", "timeframe": "1min", "action": "subscribe"}
EOF

cat > /tmp/subscribe_msft.json << 'EOF'
{"type": "subscription", "symbol": "MSFT", "timeframe": "1min", "action": "subscribe"}
EOF

echo "Subscription messages:"
echo "AAPL 1min:"
cat /tmp/subscribe_aapl.json
echo ""
echo "GOOGL 1min:"
cat /tmp/subscribe_googl.json
echo ""
echo "MSFT 1min:"
cat /tmp/subscribe_msft.json
echo ""

echo "Connecting to WebSocket and sending multiple subscriptions..."
echo "Press Ctrl+C to stop"

# Combine all subscription messages
cat /tmp/subscribe_aapl.json /tmp/subscribe_googl.json /tmp/subscribe_msft.json > /tmp/all_subscriptions.json

# Use websocat to connect and send all subscriptions
websocat --text ws://localhost:8080/ws/ohlcv --binary-prefix=@ --text-prefix=% < /tmp/all_subscriptions.json

# Clean up
rm -f /tmp/subscribe_*.json /tmp/all_subscriptions.json
