import { config } from './config';
import type { OHLCVCandle, EnrichedCandle, SymbolInfo, APIResponse } from '../types';

class APIClient {
  private baseURL: string;

  constructor() {
    this.baseURL = config.api.baseURL;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<APIResponse<T>> {
    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        ...options,
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      return {
        success: true,
        data,
      };
    } catch (error) {
      console.error('API Request failed:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  // OHLCV Candles
  async getCandles(symbol: string, interval: string, limit?: number): Promise<APIResponse<OHLCVCandle[]>> {
    const params = new URLSearchParams({
      symbol,
      interval,
      ...(limit && { limit: limit.toString() }),
    });
    
    return this.request<OHLCVCandle[]>(`${config.api.endpoints.candles}?${params}`);
  }

  // Enriched Candles
  async getEnrichedCandles(symbol: string, interval: string, limit?: number): Promise<APIResponse<EnrichedCandle[]>> {
    const params = new URLSearchParams({
      symbol,
      interval,
      ...(limit && { limit: limit.toString() }),
    });
    
    return this.request<EnrichedCandle[]>(`${config.api.endpoints.enriched}?${params}`);
  }

  // Symbols
  async getSymbols(): Promise<APIResponse<SymbolInfo[]>> {
    return this.request<SymbolInfo[]>(config.api.endpoints.symbols);
  }

  async addSymbol(symbol: string): Promise<APIResponse<SymbolInfo>> {
    return this.request<SymbolInfo>(config.api.endpoints.symbols, {
      method: 'POST',
      body: JSON.stringify({ symbol }),
    });
  }

  async removeSymbol(symbol: string): Promise<APIResponse<void>> {
    return this.request<void>(`${config.api.endpoints.symbols}/${symbol}`, {
      method: 'DELETE',
    });
  }

  // Health & Status
  async getHealth(): Promise<APIResponse<{ status: string; timestamp: string }>> {
    return this.request<{ status: string; timestamp: string }>(config.api.endpoints.health);
  }

  async getStatus(): Promise<APIResponse<any>> {
    return this.request<any>(config.api.endpoints.status);
  }
}

export const apiClient = new APIClient();
