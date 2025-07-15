import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import Layout from './components/layout/Layout';
import Dashboard from './pages/Dashboard';
import Symbols from './pages/Symbols';
import History from './pages/History';
import CLI from './pages/CLI';
import Monitoring from './pages/Monitoring';
import Settings from './pages/Settings';
import ErrorFallback from './components/ui/ErrorFallback';
import { useThemeStore } from './stores/themeStore';

const App: React.FC = () => {
  const { theme, resolvedTheme, setTheme } = useThemeStore();

  React.useEffect(() => {
    // Initialize theme on app start
    setTheme(theme);
  }, []);

  React.useEffect(() => {
    // Apply resolved theme to document root
    if (resolvedTheme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [resolvedTheme]);

  return (
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <div className="min-h-screen bg-background text-foreground transition-colors duration-200">
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/symbols" element={<Symbols />} />
            <Route path="/history" element={<History />} />
            <Route path="/cli" element={<CLI />} />
            <Route path="/monitoring" element={<Monitoring />} />
            <Route path="/settings" element={<Settings />} />
          </Routes>
        </Layout>
      </div>
    </ErrorBoundary>
  );
};

export default App;
