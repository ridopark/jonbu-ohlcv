import React from 'react';
import { useThemeStore } from '../stores/themeStore';

const Settings: React.FC = () => {
  const { theme, setTheme } = useThemeStore();
  
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-card-foreground">Settings</h1>
        <p className="text-muted-foreground">
          Configure application preferences and settings
        </p>
      </div>
      
      {/* Theme Settings */}
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-xl font-semibold text-card-foreground mb-4">
          Appearance
        </h2>
        
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-card-foreground mb-2">
              Theme
            </label>
            <div className="flex gap-2">
              <button
                onClick={() => setTheme('light')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  theme === 'light'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                }`}
              >
                ☀️ Light
              </button>
              
              <button
                onClick={() => setTheme('dark')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  theme === 'dark'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                }`}
              >
                🌙 Dark
              </button>
              
              <button
                onClick={() => setTheme('system')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  theme === 'system'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                }`}
              >
                🌓 System
              </button>
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Choose your preferred color theme or follow system preference
            </p>
          </div>
        </div>
      </div>
      
      {/* API Settings */}
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-xl font-semibold text-card-foreground mb-4">
          API Configuration
        </h2>
        
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-card-foreground mb-2">
              Backend Server URL
            </label>
            <input
              type="text"
              value="http://localhost:8080"
              disabled
              className="w-full px-3 py-2 bg-muted border border-input rounded-md text-muted-foreground"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Backend server endpoint (configured via environment)
            </p>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-card-foreground mb-2">
              WebSocket URL
            </label>
            <input
              type="text"
              value="ws://localhost:8080/ws"
              disabled
              className="w-full px-3 py-2 bg-muted border border-input rounded-md text-muted-foreground"
            />
            <p className="text-xs text-muted-foreground mt-1">
              WebSocket endpoint for real-time data
            </p>
          </div>
        </div>
      </div>
      
      {/* Chart Settings */}
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-xl font-semibold text-card-foreground mb-4">
          Chart Configuration
        </h2>
        
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-card-foreground mb-2">
              Default Timeframe
            </label>
            <select className="w-full px-3 py-2 bg-background border border-input rounded-md text-card-foreground">
              <option value="1m">1 Minute</option>
              <option value="5m">5 Minutes</option>
              <option value="15m">15 Minutes</option>
              <option value="1h">1 Hour</option>
              <option value="1d">1 Day</option>
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-card-foreground mb-2">
              Default Symbol
            </label>
            <input
              type="text"
              placeholder="AAPL"
              className="w-full px-3 py-2 bg-background border border-input rounded-md text-card-foreground"
            />
          </div>
          
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="autoRefresh"
              className="rounded border-input"
            />
            <label htmlFor="autoRefresh" className="text-sm text-card-foreground">
              Enable auto-refresh (real-time updates)
            </label>
          </div>
          
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="showIndicators"
              defaultChecked
              className="rounded border-input"
            />
            <label htmlFor="showIndicators" className="text-sm text-card-foreground">
              Show technical indicators by default
            </label>
          </div>
        </div>
      </div>
      
      {/* About */}
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-xl font-semibold text-card-foreground mb-4">
          About
        </h2>
        
        <div className="space-y-2 text-sm text-muted-foreground">
          <div className="flex justify-between">
            <span>Version:</span>
            <span className="font-medium">1.0.0</span>
          </div>
          <div className="flex justify-between">
            <span>Build:</span>
            <span className="font-medium">Phase 4 - Web Frontend</span>
          </div>
          <div className="flex justify-between">
            <span>Stack:</span>
            <span className="font-medium">Go + React + TypeScript</span>
          </div>
          <div className="flex justify-between">
            <span>License:</span>
            <span className="font-medium">MIT</span>
          </div>
        </div>
        
        <div className="mt-4 pt-4 border-t border-border">
          <p className="text-xs text-muted-foreground">
            Jonbu OHLCV is a comprehensive financial data streaming and analysis platform
            built with modern technologies for real-time market data processing.
          </p>
        </div>
      </div>
    </div>
  );
};

export default Settings;
