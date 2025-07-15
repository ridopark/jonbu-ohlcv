import { config } from './config';

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = config.websocket.maxReconnectAttempts;
  private reconnectInterval = config.websocket.reconnectInterval;
  private listeners: Map<string, Set<(data: any) => void>> = new Map();
  private isConnecting = false;

  constructor() {
    this.connect();
  }

  private connect(): void {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.isConnecting = true;

    try {
      this.ws = new WebSocket(config.websocket.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.emit('connection', { status: 'connected' });
      };

      this.ws.onmessage = (event) => {
        try {
          // Handle potentially concatenated JSON messages
          this.parseWebSocketMessages(event.data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error, 'Raw data:', event.data);
        }
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        this.isConnecting = false;
        this.ws = null;
        this.emit('connection', { status: 'disconnected' });
        this.handleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnecting = false;
        this.emit('connection', { status: 'error', error });
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.isConnecting = false;
      this.handleReconnect();
    }
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.emit('connection', { status: 'failed' });
      return;
    }

    this.reconnectAttempts++;
    console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);

    setTimeout(() => {
      this.connect();
    }, this.reconnectInterval);
  }

  public subscribe(event: string, callback: (data: any) => void): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    
    this.listeners.get(event)!.add(callback);

    // Return unsubscribe function
    return () => {
      const eventListeners = this.listeners.get(event);
      if (eventListeners) {
        eventListeners.delete(callback);
        if (eventListeners.size === 0) {
          this.listeners.delete(event);
        }
      }
    };
  }

  private emit(event: string, data: any): void {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      eventListeners.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in WebSocket event listener for ${event}:`, error);
        }
      });
    }
  }

  public send(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message);
    }
  }

  // WebSocket subscription methods
  public subscribeToSymbol(symbol: string, timeframe: string): void {
    const subscription = {
      type: 'subscription',
      symbol: symbol,
      timeframe: timeframe,
      action: 'subscribe'
    };
    
    console.log(`🔔 Subscribing to ${symbol} ${timeframe}`, {
      subscription,
      timestamp: new Date().toISOString(),
      connectionState: this.readyState === WebSocket.OPEN ? 'OPEN' : 'NOT_OPEN'
    });
    this.send(subscription);
  }

  public unsubscribeFromSymbol(symbol: string, timeframe: string): void {
    const subscription = {
      type: 'subscription', 
      symbol: symbol,
      timeframe: timeframe,
      action: 'unsubscribe'
    };
    
    console.log(`🔕 Unsubscribing from ${symbol} ${timeframe}`, {
      subscription,
      timestamp: new Date().toISOString(),
      connectionState: this.readyState === WebSocket.OPEN ? 'OPEN' : 'NOT_OPEN'
    });
    this.send(subscription);
  }

  // Parse potentially concatenated JSON messages from WebSocket
  private parseWebSocketMessages(rawData: string): void {
    // Handle single JSON object (most common case)
    if (rawData.trim().startsWith('{') && rawData.trim().endsWith('}')) {
      try {
        const message = JSON.parse(rawData);
        this.handleWebSocketMessage(message);
        return;
      } catch (error) {
        // If single parse fails, try splitting approach
        console.warn('Single JSON parse failed, attempting message splitting...');
      }
    }

    // Handle concatenated JSON messages by splitting on "}{"
    const messages: string[] = [];
    let currentMessage = '';
    let braceCount = 0;
    let inString = false;
    let escapeNext = false;

    for (let i = 0; i < rawData.length; i++) {
      const char = rawData[i];
      currentMessage += char;

      if (escapeNext) {
        escapeNext = false;
        continue;
      }

      if (char === '\\') {
        escapeNext = true;
        continue;
      }

      if (char === '"') {
        inString = !inString;
        continue;
      }

      if (!inString) {
        if (char === '{') {
          braceCount++;
        } else if (char === '}') {
          braceCount--;
          
          // Complete JSON object found
          if (braceCount === 0) {
            messages.push(currentMessage.trim());
            currentMessage = '';
          }
        }
      }
    }

    // Add any remaining message
    if (currentMessage.trim()) {
      messages.push(currentMessage.trim());
    }

    // Parse and handle each message
    console.log(`📦 Split WebSocket data into ${messages.length} messages`);
    
    messages.forEach((messageStr, index) => {
      try {
        if (messageStr) {
          const message = JSON.parse(messageStr);
          this.handleWebSocketMessage(message);
        }
      } catch (error) {
        console.error(`Failed to parse message ${index + 1}:`, error, 'Message:', messageStr);
      }
    });
  }

  // Handle individual parsed WebSocket message
  private handleWebSocketMessage(message: any): void {
    // Log raw message received
    console.log('📨 WebSocket message received:', {
      type: message.type,
      timestamp: new Date().toISOString(),
      data: message
    });
    
    // Handle different message types from the backend
    if (message.type === 'candle') {
      console.log('📈 Processing candle data:', {
        symbol: message.symbol || message.data?.symbol,
        timeframe: message.timeframe || message.data?.interval,
        timestamp: message.data?.timestamp,
        close: message.data?.close,
        volume: message.data?.volume
      });
      this.emit('candle', message);
    } else if (message.type === 'error') {
      console.error('❌ WebSocket error received:', message);
      this.emit('error', message);
    } else if (message.type === 'status') {
      console.log('ℹ️ WebSocket status update:', message);
      this.emit('status', message);
    } else if (message.type === 'connected') {
      console.log('🔗 WebSocket connection confirmed:', message);
      this.emit('connected', message);
    } else {
      // Fallback for any other message format
      console.log('🔄 WebSocket unknown message type:', message.type, message);
      this.emit(message.type || 'message', message);
    }
  }

  public close(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.listeners.clear();
  }

  public get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }

  public get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
export const wsClient = new WebSocketClient();
