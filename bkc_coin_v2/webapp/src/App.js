import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Toaster } from 'react-hot-toast';

// Components
import Game from './components/Game';
import Profile from './components/Profile';
import Leaderboard from './components/Leaderboard';
import P2PMarket from './components/P2PMarket';
import Bank from './components/Bank';
import Admin from './components/Admin';
import Loading from './components/Loading';
import ErrorBoundary from './components/ErrorBoundary';

// Services
import apiService from './services/api';
import { useTelegram } from './hooks/useTelegram';

// Styles
import './App.css';

// Create React Query client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      staleTime: 5 * 60 * 1000, // 5 minutes
      cacheTime: 10 * 60 * 1000, // 10 minutes
    },
  },
});

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [currentNode, setCurrentNode] = useState(null);

  // Initialize Telegram Web App
  const { tg, user: tgUser, ready } = useTelegram();

  useEffect(() => {
    if (!ready || !tgUser) {
      return;
    }

    initializeApp();
  }, [ready, tgUser]);

  const initializeApp = async () => {
    try {
      setLoading(true);
      
      // Get current node info
      const stats = apiService.getLoadBalancerStats();
      setCurrentNode(stats.currentNode);

      // Initialize user data
      const userProfile = await apiService.getUserProfile(tgUser.id);
      
      setUser({
        ...tgUser,
        ...userProfile
      });

      // Configure Telegram Web App
      if (tg) {
        tg.ready();
        tg.expand();
        
        // Set theme colors
        tg.setHeaderColor('#1a1a1a');
        tg.setBackgroundColor('#0f0f0f');
        
        // Enable back button
        tg.onEvent('backButtonClicked', () => {
          window.history.back();
        });
      }

    } catch (err) {
      console.error('Failed to initialize app:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle navigation
  const handleNavigation = (path) => {
    if (tg) {
      if (path === '/') {
        tg.BackButton.hide();
      } else {
        tg.BackButton.show();
      }
    }
  };

  if (loading) {
    return <Loading message="Connecting to BKC Coin Network..." />;
  }

  if (error) {
    return (
      <div className="error-container">
        <h2>Connection Error</h2>
        <p>{error}</p>
        <button onClick={initializeApp}>Retry</button>
      </div>
    );
  }

  if (!user) {
    return <Loading message="Loading user data..." />;
  }

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <Router>
          <div className="app">
            {/* Node indicator */}
            {currentNode && (
              <div className="node-indicator">
                <span>Node: {currentNode.split('//')[1]}</span>
              </div>
            )}

            {/* Main Routes */}
            <Routes>
              <Route 
                path="/" 
                element={
                  <Game 
                    user={user} 
                    onNavigate={handleNavigation}
                    apiService={apiService}
                  />
                } 
              />
              <Route 
                path="/profile" 
                element={
                  <Profile 
                    user={user} 
                    onNavigate={handleNavigation}
                    apiService={apiService}
                  />
                } 
              />
              <Route 
                path="/leaderboard" 
                element={
                  <Leaderboard 
                    user={user} 
                    onNavigate={handleNavigation}
                    apiService={apiService}
                  />
                } 
              />
              <Route 
                path="/p2p" 
                element={
                  <P2PMarket 
                    user={user} 
                    onNavigate={handleNavigation}
                    apiService={apiService}
                  />
                } 
              />
              <Route 
                path="/bank" 
                element={
                  <Bank 
                    user={user} 
                    onNavigate={handleNavigation}
                    apiService={apiService}
                  />
                } 
              />
              <Route 
                path="/admin" 
                element={
                  <Admin 
                    user={user} 
                    onNavigate={handleNavigation}
                    apiService={apiService}
                  />
                } 
              />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>

            {/* Toast notifications */}
            <Toaster
              position="top-center"
              toastOptions={{
                duration: 4000,
                style: {
                  background: '#1a1a1a',
                  color: '#ffffff',
                  border: '1px solid #333',
                },
                success: {
                  iconTheme: {
                    primary: '#10b981',
                    secondary: '#ffffff',
                  },
                },
                error: {
                  iconTheme: {
                    primary: '#ef4444',
                    secondary: '#ffffff',
                  },
                },
              }}
            />
          </div>
        </Router>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}

export default App;
