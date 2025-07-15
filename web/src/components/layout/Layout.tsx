import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import ThemeToggle from '../ui/ThemeToggle';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation();

  const navigationItems = [
    { path: '/dashboard', name: 'Dashboard', icon: '📊' },
    { path: '/symbols', name: 'Symbols', icon: '🏷️' },
    { path: '/history', name: 'History', icon: '📈' },
    { path: '/cli', name: 'CLI', icon: '💻' },
    { path: '/monitoring', name: 'Monitoring', icon: '🔍' },
    { path: '/settings', name: 'Settings', icon: '⚙️' },
  ];

  const isActivePath = (path: string): boolean => {
    return location.pathname === path || (path === '/dashboard' && location.pathname === '/');
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border sticky top-0 z-50">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <div className="flex items-center space-x-4">
              <Link to="/" className="flex items-center space-x-2">
                <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                  <span className="text-primary-foreground font-bold text-lg">J</span>
                </div>
                <span className="font-semibold text-card-foreground text-xl">
                  Jonbu OHLCV
                </span>
              </Link>
            </div>

            {/* Navigation */}
            <nav className="hidden md:flex items-center space-x-1">
              {navigationItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                    isActivePath(item.path)
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:text-card-foreground hover:bg-accent'
                  }`}
                >
                  <span className="mr-2">{item.icon}</span>
                  {item.name}
                </Link>
              ))}
            </nav>

            {/* Theme Toggle */}
            <div className="flex items-center space-x-4">
              <ThemeToggle />
            </div>
          </div>
        </div>
      </header>

      {/* Mobile Navigation */}
      <nav className="md:hidden bg-card border-b border-border">
        <div className="container mx-auto px-4">
          <div className="flex overflow-x-auto py-2 space-x-1">
            {navigationItems.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`flex-shrink-0 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isActivePath(item.path)
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-card-foreground hover:bg-accent'
                }`}
              >
                <span className="mr-1">{item.icon}</span>
                {item.name}
              </Link>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6">
        {children}
      </main>

      {/* Footer */}
      <footer className="bg-card border-t border-border mt-auto">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <div>
              © 2024 Jonbu OHLCV. Built with Go + React + TypeScript.
            </div>
            <div className="flex items-center space-x-4">
              <span>v1.0.0</span>
              <span>•</span>
              <a 
                href="https://github.com/ridopark/jonbu-ohlcv" 
                className="hover:text-card-foreground transition-colors"
                target="_blank"
                rel="noopener noreferrer"
              >
                GitHub
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Layout;
