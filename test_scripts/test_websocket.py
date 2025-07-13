#!/usr/bin/env python3
"""
Simple WebSocket client to test OHLCV data streaming
"""
import asyncio
import json
import websockets
import sys

async def test_websocket():
    uri = "ws://localhost:8080/ws/ohlcv"
    
    try:
        print(f"Connecting to {uri}...")
        async with websockets.connect(uri) as websocket:
            print("Connected successfully!")
            
            # Send subscription messages for multiple symbols
            subscriptions = [
                {"symbol": "AAPL", "timeframe": "1min"},
                {"symbol": "GOOGL", "timeframe": "1min"},
                {"symbol": "MSFT", "timeframe": "1min"}
            ]
            
            for sub in subscriptions:
                subscription = {
                    "type": "subscription",
                    "symbol": sub["symbol"],
                    "timeframe": sub["timeframe"], 
                    "action": "subscribe"
                }
                
                print(f"Sending subscription: {subscription}")
                await websocket.send(json.dumps(subscription))
                await asyncio.sleep(0.1)  # Small delay between subscriptions
            
            print(f"All {len(subscriptions)} subscriptions sent!")
            
            # Listen for messages
            print("Listening for candle data... (Press Ctrl+C to stop)")
            timeout_count = 0
            max_timeouts = 10
            
            while timeout_count < max_timeouts:
                try:
                    # Wait for message with timeout
                    message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                    
                    # Parse and display message
                    try:
                        data = json.loads(message)
                        print(f"Received: {json.dumps(data, indent=2)}")
                        timeout_count = 0  # Reset timeout counter on successful message
                    except json.JSONDecodeError:
                        print(f"Received non-JSON message: {message}")
                        
                except asyncio.TimeoutError:
                    timeout_count += 1
                    print(f"No message received (timeout {timeout_count}/{max_timeouts})")
                    
                    # Send ping to keep connection alive
                    ping_msg = {"type": "ping"}
                    await websocket.send(json.dumps(ping_msg))
                    print("Sent ping message")
                    
            print("Max timeouts reached, ending test")
                    
    except Exception as e:
        print(f"Error: {e}")
        return 1
        
    return 0

if __name__ == "__main__":
    try:
        exit_code = asyncio.run(test_websocket())
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\nTest interrupted by user")
        sys.exit(0)
