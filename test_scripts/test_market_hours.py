#!/usr/bin/env python3
"""
Test market hours simulation in the mock service
"""
import asyncio
import json
import websockets
import sys
from datetime import datetime, timezone
import pytz

async def test_market_hours():
    uri = "ws://localhost:8080/ws/ohlcv"
    
    # Get current time in different timezones
    utc_now = datetime.now(timezone.utc)
    est = pytz.timezone('America/New_York')
    est_now = utc_now.astimezone(est)
    
    print(f"🕐 Current Time Analysis:")
    print(f"   UTC: {utc_now.strftime('%Y-%m-%d %H:%M:%S %Z')}")
    print(f"   EST: {est_now.strftime('%Y-%m-%d %H:%M:%S %Z')}")
    print(f"   Day: {est_now.strftime('%A')}")
    
    # Determine if we're in market hours
    is_weekday = est_now.weekday() < 5  # 0-4 are Mon-Fri
    hour = est_now.hour
    minute = est_now.minute
    
    is_market_hours = (
        is_weekday and 
        (hour > 9 or (hour == 9 and minute >= 30)) and 
        hour < 16
    )
    
    print(f"   📈 Market Status: {'🟢 OPEN' if is_market_hours else '🔴 CLOSED'}")
    if not is_weekday:
        print(f"   📅 Reason: Weekend")
    elif hour < 9 or (hour == 9 and minute < 30):
        print(f"   📅 Reason: Before market open (9:30 AM EST)")
    elif hour >= 16:
        print(f"   📅 Reason: After market close (4:00 PM EST)")
    
    print("\n🚀 Testing volume differences during market hours...")
    
    try:
        async with websockets.connect(uri) as websocket:
            print("📡 Connected to WebSocket")
            
            # Subscribe to AAPL 1min
            subscription = {
                "type": "subscription",
                "symbol": "AAPL",
                "timeframe": "1min",
                "action": "subscribe"
            }
            
            await websocket.send(json.dumps(subscription))
            print("📊 Subscribed to AAPL 1min candles")
            
            print("⏳ Monitoring volume levels...")
            print("💡 Note: Mock service generates 5x higher volume during market hours")
            print("🔍 Collecting samples (this may take 1-2 minutes for weekend data)...")
            
            volume_samples = []
            sample_count = 0
            max_samples = 5  # Reduced for faster testing
            timeout_attempts = 0
            max_timeouts = 15  # Allow more timeouts for weekend
            
            while sample_count < max_samples and timeout_attempts < max_timeouts:
                try:
                    message = await asyncio.wait_for(websocket.recv(), timeout=10.0)
                    
                    try:
                        data = json.loads(message)
                        
                        if data.get("type") == "candle":
                            candle_data = data.get("data", {})
                            volume = candle_data.get("volume", 0)
                            volume_samples.append(volume)
                            sample_count += 1
                            
                            print(f"🕯️  Sample {sample_count}: Volume = {volume:,}")
                            
                        elif data.get("type") == "connected":
                            print("✅ WebSocket connection confirmed")
                        elif data.get("type") == "pong":
                            # Send ping to keep connection alive during quiet periods
                            ping_msg = {"type": "ping"}
                            await websocket.send(json.dumps(ping_msg))
                            
                    except json.JSONDecodeError:
                        continue
                        
                except asyncio.TimeoutError:
                    timeout_attempts += 1
                    print(f"⏰ Timeout {timeout_attempts}/{max_timeouts} - waiting for weekend data...")
                    # Send ping during timeouts
                    try:
                        ping_msg = {"type": "ping"}
                        await websocket.send(json.dumps(ping_msg))
                    except:
                        break
            
            if volume_samples:
                avg_volume = sum(volume_samples) / len(volume_samples)
                min_volume = min(volume_samples)
                max_volume = max(volume_samples)
                
                print(f"\n📊 Volume Analysis ({len(volume_samples)} samples):")
                print(f"   Average: {avg_volume:,.0f}")
                print(f"   Range: {min_volume:,} - {max_volume:,}")
                
                expected_range = "25,000-300,000" if is_market_hours else "5,000-60,000"
                print(f"   Expected range for {'market hours' if is_market_hours else 'after hours'}: {expected_range}")
                
                # Analyze if volumes match expectations
                if is_market_hours:
                    if avg_volume > 15000:
                        print("✅ Volume levels consistent with market hours (higher volume)")
                    else:
                        print("⚠️  Volume seems low for market hours")
                else:
                    if avg_volume < 15000:
                        print("✅ Volume levels consistent with after hours (lower volume)")
                    else:
                        print("⚠️  Volume seems high for after hours")
            
    except Exception as e:
        print(f"❌ Error: {e}")
        return 1
        
    return 0

if __name__ == "__main__":
    try:
        exit_code = asyncio.run(test_market_hours())
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\n⛔ Test interrupted by user")
        sys.exit(0)
