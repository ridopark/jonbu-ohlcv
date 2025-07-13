#!/usr/bin/env python3
"""
WebSocket client to test multi-symbol OHLCV data streaming
"""
import asyncio
import json
import websockets
import sys

async def test_multi_symbol_websocket():
    uri = "ws://localhost:8080/ws/ohlcv"
    
    # Define multiple symbols and timeframes to subscribe to
    subscriptions = [
        {"symbol": "AAPL", "timeframe": "1min"},
        {"symbol": "GOOGL", "timeframe": "1min"},
        {"symbol": "MSFT", "timeframe": "1min"},
        {"symbol": "AAPL", "timeframe": "5min"},
        {"symbol": "GOOGL", "timeframe": "5min"},
    ]
    
    try:
        print(f"Connecting to {uri}...")
        async with websockets.connect(uri) as websocket:
            print("Connected successfully!")
            
            # Send subscription messages for multiple symbols
            for sub in subscriptions:
                subscription = {
                    "type": "subscription",
                    "symbol": sub["symbol"],
                    "timeframe": sub["timeframe"], 
                    "action": "subscribe"
                }
                
                print(f"Subscribing to: {sub['symbol']} {sub['timeframe']}")
                await websocket.send(json.dumps(subscription))
                await asyncio.sleep(0.1)  # Small delay between subscriptions
            
            print(f"\nSubscribed to {len(subscriptions)} symbol/timeframe combinations!")
            print("Listening for candle data from all subscriptions... (Press Ctrl+C to stop)")
            
            timeout_count = 0
            max_timeouts = 20  # Increased for multiple symbols
            candle_count = {}  # Track candles per symbol
            
            while timeout_count < max_timeouts:
                try:
                    # Wait for message with timeout
                    message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                    
                    # Handle multiple JSON objects in one message (split by newlines)
                    messages = message.strip().split('\n') if '\n' in message else [message]
                    
                    for single_message in messages:
                        if not single_message.strip():
                            continue
                            
                        try:
                            data = json.loads(single_message)
                            
                            if data.get("type") == "candle":
                                symbol = data.get("symbol", "Unknown")
                                timeframe = data.get("timeframe", "Unknown")
                                key = f"{symbol}:{timeframe}"
                                
                                # Track candle counts
                                candle_count[key] = candle_count.get(key, 0) + 1
                                
                                # Display candle info
                                candle_data = data.get("data", {})
                                print(f"🕯️  {symbol} {timeframe} #{candle_count[key]}: "
                                      f"O:{candle_data.get('open', 0):.2f} "
                                      f"H:{candle_data.get('high', 0):.2f} "
                                      f"L:{candle_data.get('low', 0):.2f} "
                                      f"C:{candle_data.get('close', 0):.2f} "
                                      f"V:{candle_data.get('volume', 0)}")
                            else:
                                print(f"📡 {data.get('type', 'unknown')}: {data.get('timestamp', '')}")
                            
                        except json.JSONDecodeError:
                            print(f"❌ JSON decode error for: {single_message[:100]}...")
                    
                    timeout_count = 0  # Reset timeout counter on successful message
                        
                except asyncio.TimeoutError:
                    timeout_count += 1
                    print(f"⏰ No message received (timeout {timeout_count}/{max_timeouts})")
                    
                    # Send ping to keep connection alive
                    ping_msg = {"type": "ping"}
                    await websocket.send(json.dumps(ping_msg))
                    
            print(f"\n📊 Final candle counts:")
            for key, count in candle_count.items():
                print(f"  {key}: {count} candles")
            print("Max timeouts reached, ending test")
                    
    except Exception as e:
        print(f"❌ Error: {e}")
        return 1
        
    return 0

if __name__ == "__main__":
    try:
        exit_code = asyncio.run(test_multi_symbol_websocket())
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\n\n⛔ Test interrupted by user")
        sys.exit(0)
