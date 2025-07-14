# Jonbu OHLCV Web Frontend

A modern React + TypeScript frontend for the Jonbu OHLCV financial data streaming platform.

## 🚀 Features

- **Real-time Charts**: Live candlestick charts with technical indicators
- **Symbol Management**: Add/remove symbols for tracking
- **Historical Data**: View and analyze historical OHLCV data
- **Web CLI**: Execute backend commands through browser interface
- **System Monitoring**: Real-time server health and performance metrics
- **Dark/Light Themes**: Configurable appearance with system theme support
- **Responsive Design**: Mobile-friendly interface with Tailwind CSS

## 🛠️ Technology Stack

- **React 18.2.0** - Modern React with hooks and concurrent features
- **TypeScript 5.0.2** - Type-safe development
- **Vite 4.4.5** - Fast build tool and dev server
- **Tailwind CSS 3.3.0** - Utility-first CSS framework
- **Less** - CSS preprocessor for component styling
- **Zustand** - Lightweight state management
- **React Query** - Server state management and caching
- **React Router** - Client-side routing
- **Socket.io** - Real-time WebSocket communication

## 📁 Project Structure

```
web/
├── src/
│   ├── components/
│   │   ├── charts/
│   │   │   └── CandlestickChart.tsx
│   │   ├── forms/
│   │   │   ├── SymbolSelector.tsx
│   │   │   └── TimeframeSelector.tsx
│   │   ├── layout/
│   │   │   └── Layout.tsx
│   │   └── ui/
│   │       ├── ErrorFallback.tsx
│   │       ├── MarketStatus.tsx
│   │       └── StatsCards.tsx
│   ├── pages/
│   │   ├── Dashboard.tsx
│   │   ├── Symbols.tsx
│   │   ├── History.tsx
│   │   ├── CLI.tsx
│   │   ├── Monitoring.tsx
│   │   └── Settings.tsx
│   ├── stores/
│   │   ├── themeStore.ts
│   │   └── chartStore.ts
│   ├── styles/
│   │   ├── globals.less
│   │   ├── components.less
│   │   └── themes.less
│   ├── types/
│   │   └── index.ts
│   ├── utils/
│   │   ├── api.ts
│   │   ├── config.ts
│   │   └── websocket.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── vite-env.d.ts
├── public/
├── package.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
└── tsconfig.node.json
```

## 🏗️ Development Setup

### Prerequisites

- Node.js 18+ and npm
- Go backend server running on `localhost:8080`

### Installation

```bash
# Navigate to web directory
cd web

# Install dependencies
npm install

# Start development server
npm run dev
```

The application will be available at `http://localhost:3000`

### Build for Production

```bash
# Build optimized production bundle
npm run build

# Preview production build
npm run preview
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the web directory:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

### Backend Integration

The frontend expects the Go backend to be running with the following endpoints:

- `GET /api/candles` - OHLCV candle data
- `GET /api/enriched` - Enriched candle data with indicators
- `GET /api/symbols` - Symbol management
- `GET /api/health` - Health check
- `GET /api/status` - Server status
- `WebSocket /ws` - Real-time data streaming

## 📊 Features Implementation

### Real-time Charts

- SVG-based candlestick charts with volume
- Technical indicators (SMA, EMA, RSI, MACD)
- Responsive design with zoom and pan
- Real-time updates via WebSocket

### State Management

- **Theme Store**: Dark/light/system theme preference
- **Chart Store**: Symbol, timeframe, chart data, and streaming state
- Persistent storage with Zustand middleware

### WebSocket Integration

- Auto-reconnecting WebSocket client
- Real-time candle updates
- Connection status monitoring
- Error handling and recovery

## 🎨 Styling System

### Tailwind CSS Integration

- Utility-first styling approach
- Dark mode support with CSS variables
- Responsive design patterns
- Custom color palette for financial data

### Less Preprocessor

- Component-specific styling
- Theme-aware color variables
- Responsive mixins and utilities
- Tailwind CSS integration

### Theme System

- Light/Dark/System themes
- CSS custom properties for dynamic theming
- Smooth theme transitions
- Financial chart color schemes

## 🧪 Code Quality

### TypeScript

- Strict type checking enabled
- Comprehensive type definitions
- API response typing
- Component prop interfaces

### Development Tools

- ESLint for code linting
- Prettier for code formatting
- Hot module replacement
- Source maps for debugging

## 🚀 Deployment

### Production Build

```bash
npm run build
```

Creates optimized production files in `dist/`:

- Minified JavaScript and CSS
- Asset optimization and compression
- Service worker for caching
- Static file serving ready

### Deployment Options

- **Static Hosting**: Vercel, Netlify, GitHub Pages
- **Docker**: Container-ready build
- **CDN**: Serve static assets via CDN
- **Reverse Proxy**: Nginx/Apache integration

## 📈 Performance

### Optimization Features

- Code splitting with dynamic imports
- Tree shaking for minimal bundle size
- Asset compression and optimization
- Service worker caching strategy
- Lazy loading for charts and components

### Real-time Performance

- Efficient WebSocket message handling
- Canvas/SVG chart rendering optimization
- State update batching
- Memory management for large datasets

## 🔒 Security

### Best Practices

- CORS configuration for API calls
- Input validation and sanitization
- Secure WebSocket connections
- Environment variable protection
- XSS prevention measures

## 📝 API Integration

### REST API Client

```typescript
import { apiClient } from './utils/api';

// Get candle data
const response = await apiClient.getCandles('AAPL', '1m', 100);

// Get enriched data
const enriched = await apiClient.getEnrichedCandles('AAPL', '1m');
```

### WebSocket Client

```typescript
import { wsClient } from './utils/websocket';

// Subscribe to real-time candles
const unsubscribe = wsClient.subscribe('candle', (candle) => {
  console.log('New candle:', candle);
});

// Clean up subscription
unsubscribe();
```

## 🧩 Component Architecture

### Page Components

- **Dashboard**: Real-time charts and market overview
- **Symbols**: Symbol management interface
- **History**: Historical data analysis
- **CLI**: Web-based command interface
- **Monitoring**: System health dashboard
- **Settings**: Application configuration

### UI Components

- **Layout**: Navigation and page structure
- **Charts**: Candlestick and technical indicators
- **Forms**: Symbol and timeframe selectors
- **Status**: Market status and connection state

## 🔄 State Flow

```
User Action → Store Update → Component Re-render → API/WebSocket → Backend
     ↑                                                              ↓
UI Update ← State Change ← WebSocket Message ← Real-time Data ← Backend
```

## 🎯 Phase 4 Completion

This frontend completes **Phase 4 - Web Frontend** of the Jonbu OHLCV platform:

✅ **Modern React Application** - TypeScript, hooks, error boundaries  
✅ **Real-time Data Visualization** - Charts, indicators, live updates  
✅ **Responsive UI/UX** - Mobile-friendly, dark/light themes  
✅ **WebSocket Integration** - Live streaming, auto-reconnect  
✅ **State Management** - Zustand stores, persistence  
✅ **API Integration** - RESTful endpoints, error handling  
✅ **Build System** - Vite, TypeScript, Tailwind, Less  
✅ **Production Ready** - Optimized builds, deployment config  

## 📞 Support

For issues and questions:

- Check the backend logs at `logs/` directory
- Verify WebSocket connection in browser dev tools
- Review API responses in Network tab
- Check console for JavaScript errors

---

**Built with ❤️ for modern financial data streaming**
