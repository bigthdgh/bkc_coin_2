import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Users, 
  Activity,
  AlertTriangle,
  CheckCircle,
  XCircle,
  RefreshCw,
  Eye,
  Ban,
  CreditCard,
  Flame,
  BarChart3,
  PieChart,
  Settings
} from 'lucide-react';
import toast from 'react-hot-toast';

const AdminDashboard = ({ user, onNavigate, apiService }) => {
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // Data states
  const [systemStats, setSystemStats] = useState(null);
  const [bankStats, setBankStats] = useState(null);
  const [p2pStats, setP2pStats] = useState(null);
  const [nftStats, setNFTStats] = useState(null);
  const [currentPrice, setCurrentPrice] = useState(0);
  const [burnRate, setBurnRate] = useState(0);

  useEffect(() => {
    if (user?.isAdmin) {
      loadDashboardData();
    }
  }, [user]);

  const loadDashboardData = async () => {
    try {
      const [
        statsRes,
        bankRes,
        p2pRes,
        nftRes,
        priceRes
      ] = await Promise.all([
        apiService.getSystemStats(),
        apiService.getBankStats(),
        apiService.getP2PStats(),
        apiService.getNFTStats(),
        apiService.getCurrentPrice()
      ]);

      setSystemStats(statsRes);
      setBankStats(bankRes);
      setP2pStats(p2pRes);
      setNFTStats(nftRes);
      setCurrentPrice(priceRes.currentPrice);
      setBurnRate(priceRes.burnRate);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      toast.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const refreshData = async () => {
    setRefreshing(true);
    await loadDashboardData();
    setRefreshing(false);
    toast.success('Dashboard refreshed');
  };

  const formatNumber = (num) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  };

  const formatCurrency = (amount) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 4,
      maximumFractionDigits: 4
    }).format(amount);
  };

  const getStatusColor = (value, threshold, inverse = false) => {
    if (inverse) {
      return value > threshold ? 'text-red-500' : 'text-green-500';
    }
    return value > threshold ? 'text-green-500' : 'text-red-500';
  };

  const getStatusIcon = (positive) => {
    return positive ? <TrendingUp size={16} /> : <TrendingDown size={16} />;
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner" />
        <p>Loading admin dashboard...</p>
      </div>
    );
  }

  if (!user?.isAdmin) {
    return (
      <div className="error-container">
        <h2>Access Denied</h2>
        <p>You don't have permission to access this panel.</p>
        <button onClick={() => onNavigate('/')}>Go Back</button>
      </div>
    );
  }

  return (
    <div className="admin-dashboard">
      {/* Header */}
      <header className="admin-header">
        <div className="header-left">
          <h1>Crypto-Bank Control Panel</h1>
          <span className="header-subtitle">Real-time Financial Management</span>
        </div>
        <div className="header-right">
          <button 
            className="refresh-button"
            onClick={refreshData}
            disabled={refreshing}
          >
            <RefreshCw size={20} className={refreshing ? 'spinning' : ''} />
            Refresh
          </button>
        </div>
      </header>

      {/* Navigation */}
      <nav className="admin-nav">
        <button
          className={`nav-tab ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          <BarChart3 size={18} />
          Overview
        </button>
        <button
          className={`nav-tab ${activeTab === 'bank' ? 'active' : ''}`}
          onClick={() => setActiveTab('bank')}
        >
          <CreditCard size={18} />
          Bank System
        </button>
        <button
          className={`nav-tab ${activeTab === 'p2p' ? 'active' : ''}`}
          onClick={() => setActiveTab('p2p')}
        >
          <Users size={18} />
          P2P Market
        </button>
        <button
          className={`nav-tab ${activeTab === 'nft' ? 'active' : ''}`}
          onClick={() => setActiveTab('nft')}
        >
          <Flame size={18} />
          NFT Market
        </button>
        <button
          className={`nav-tab ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => setActiveTab('settings')}
        >
          <Settings size={18} />
          Settings
        </button>
      </nav>

      {/* Content */}
      <main className="admin-main">
        <AnimatePresence mode="wait">
          {activeTab === 'overview' && (
            <motion.div
              key="overview"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>Financial Overview</h2>
              
              {/* Current Price Display */}
              <div className="price-display">
                <div className="price-card">
                  <h3>Current BKC Price</h3>
                  <div className="price-value">
                    <span className="price-amount">{formatCurrency(currentPrice)}</span>
                    <span className="price-change">
                      {getStatusIcon(burnRate > 0)}
                      <span className={getStatusColor(burnRate, 0, true)}>
                        {burnRate > 0 ? '+' : ''}{burnRate.toFixed(2)}%
                      </span>
                    </span>
                  </div>
                </div>
              </div>

              {/* Key Metrics */}
              {systemStats && (
                <div className="metrics-grid">
                  <div className="metric-card">
                    <div className="metric-header">
                      <DollarSign size={24} className="metric-icon" />
                      <span className="metric-label">Total Supply</span>
                    </div>
                    <div className="metric-value">{formatNumber(systemStats.totalSupply)} BKC</div>
                    <div className="metric-change">
                      <span className={getStatusColor(systemStats.circulatingSupply, systemStats.totalSupply * 0.8)}>
                        {formatNumber(systemStats.circulatingSupply)} circulating
                      </span>
                    </div>
                  </div>

                  <div className="metric-card">
                    <div className="metric-header">
                      <Flame size={24} className="metric-icon" />
                      <span className="metric-label">Total Burned</span>
                    </div>
                    <div className="metric-value">{formatNumber(systemStats.totalBurned)} BKC</div>
                    <div className="metric-change text-red-500">
                      {systemStats.burnRate.toFixed(2)}% burn rate
                    </div>
                  </div>

                  <div className="metric-card">
                    <div className="metric-header">
                      <Users size={24} className="metric-icon" />
                      <span className="metric-label">Active Users</span>
                    </div>
                    <div className="metric-value">{formatNumber(systemStats.activeUsers)}</div>
                    <div className="metric-change text-green-500">
                      +{systemStats.newUsersToday} today
                    </div>
                  </div>

                  <div className="metric-card">
                    <div className="metric-header">
                      <Activity size={24} className="metric-icon" />
                      <span className="metric-label">Daily Volume</span>
                    </div>
                    <div className="metric-value">${formatNumber(systemStats.dailyVolume)}</div>
                    <div className="metric-change">
                      {getStatusIcon(systemStats.volumeChange > 0)}
                      <span className={getStatusColor(systemStats.volumeChange, 0)}>
                        {systemStats.volumeChange > 0 ? '+' : ''}{systemStats.volumeChange.toFixed(1)}%
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'bank' && (
            <motion.div
              key="bank"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>Bank System Statistics</h2>
              
              {bankStats && (
                <div className="bank-stats">
                  <div className="stat-row">
                    <div className="stat-card">
                      <h4>Total Loans</h4>
                      <div className="stat-value">{formatNumber(bankStats.totalLoans)}</div>
                    </div>
                    <div className="stat-card">
                      <h4>Active Loans</h4>
                      <div className="stat-value">{formatNumber(bankStats.activeLoans)}</div>
                    </div>
                  </div>

                  <div className="stat-row">
                    <div className="stat-card">
                      <h4>Defaulted Loans</h4>
                      <div className="stat-value text-red-500">{formatNumber(bankStats.defaultedLoans)}</div>
                    </div>
                    <div className="stat-card">
                      <h4>Overdue Loans</h4>
                      <div className="stat-value text-yellow-500">{formatNumber(bankStats.overdueLoans)}</div>
                    </div>
                  </div>

                  <div className="stat-row">
                    <div className="stat-card">
                      <h4>Total Loaned</h4>
                      <div className="stat-value">{formatNumber(bankStats.totalLoaned)} BKC</div>
                    </div>
                    <div className="stat-card">
                      <h4>Total Penalties</h4>
                      <div className="stat-value text-red-500">{formatNumber(bankStats.totalPenalties)} BKC</div>
                    </div>
                  </div>

                  <div className="default-rate">
                    <h4>Default Rate</h4>
                    <div className="rate-value">
                      {bankStats.totalLoans > 0 
                        ? ((bankStats.defaultedLoans / bankStats.totalLoans) * 100).toFixed(1)
                        : '0.0'
                      }%
                    </div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'p2p' && (
            <motion.div
              key="p2p"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>P2P Marketplace Statistics</h2>
              
              {p2pStats && (
                <div className="p2p-stats">
                  <div className="stat-row">
                    <div className="stat-card">
                      <h4>Total Trades</h4>
                      <div className="stat-value">{formatNumber(p2pStats.totalTrades)}</div>
                    </div>
                    <div className="stat-card">
                      <h4>24h Volume</h4>
                      <div className="stat-value">${formatNumber(p2pStats.volume24h)}</div>
                    </div>
                  </div>

                  <div className="stat-row">
                    <div className="stat-card">
                      <h4>Open Orders</h4>
                      <div className="stat-value">{formatNumber(p2pStats.openOrders)}</div>
                    </div>
                    <div className="stat-card">
                      <h4>Total Fees</h4>
                      <div className="stat-value text-green-500">{formatNumber(p2pStats.totalFees)} BKC</div>
                    </div>
                  </div>

                  <div className="order-book">
                    <h4>Order Book Summary</h4>
                    <div className="book-summary">
                      <div className="book-side">
                        <h5>Best Buy</h5>
                        <div className="price">${p2pStats.bestBuy.toFixed(4)}</div>
                      </div>
                      <div className="book-side">
                        <h5>Best Sell</h5>
                        <div className="price">${p2pStats.bestSell.toFixed(4)}</div>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'nft' && (
            <motion.div
              key="nft"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>NFT Marketplace Statistics</h2>
              
              {nftStats && (
                <div className="nft-stats">
                  <div className="nft-grid">
                    <div className="nft-type-card">
                      <h4>Diggers</h4>
                      <div className="nft-metrics">
                        <div className="metric">
                          <span className="label">Supply:</span>
                          <span className="value">{nftStats.diggerSupply}/{nftStats.diggerCount}</span>
                        </div>
                        <div className="metric">
                          <span className="label">Holders:</span>
                          <span className="value">{formatNumber(nftStats.diggerHolders)}</span>
                        </div>
                      </div>
                    </div>

                    <div className="nft-type-card">
                      <h4>Bankers</h4>
                      <div className="nft-metrics">
                        <div className="metric">
                          <span className="label">Supply:</span>
                          <span className="value">{nftStats.bankerSupply}/{nftStats.bankerCount}</span>
                        </div>
                        <div className="metric">
                          <span className="label">Holders:</span>
                          <span className="value">{formatNumber(nftStats.bankerHolders)}</span>
                        </div>
                      </div>
                    </div>

                    <div className="nft-type-card">
                      <h4>Inspectors</h4>
                      <div className="nft-metrics">
                        <div className="metric">
                          <span className="label">Supply:</span>
                          <span className="value">{nftStats.inspectorSupply}/{nftStats.inspectorCount}</span>
                        </div>
                        <div className="metric">
                          <span className="label">Holders:</span>
                          <span className="value">{formatNumber(nftStats.inspectorHolders)}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="nft-summary">
                    <div className="summary-card">
                      <h4>Total NFT Holders</h4>
                      <div className="summary-value">{formatNumber(nftStats.totalHolders)}</div>
                    </div>
                    <div className="summary-card">
                      <h4>Floor Price</h4>
                      <div className="summary-value">${formatCurrency(nftStats.floorPrice)}</div>
                    </div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'settings' && (
            <motion.div
              key="settings"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>System Settings</h2>
              
              <div className="settings-grid">
                <div className="setting-card">
                  <h4>Loan Settings</h4>
                  <div className="setting-controls">
                    <label>
                      <span>Max Loan Percentage:</span>
                      <input type="number" defaultValue="20" min="10" max="50" />
                      <span>% of total taps</span>
                    </label>
                    <label>
                      <span>Daily Interest Rate:</span>
                      <input type="number" defaultValue="5" min="1" max="20" />
                      <span>%</span>
                    </label>
                  </div>
                </div>

                <div className="setting-card">
                  <h4>P2P Settings</h4>
                  <div className="setting-controls">
                    <label>
                      <span>Trading Fee:</span>
                      <input type="number" defaultValue="5" min="1" max="10" />
                      <span>%</span>
                    </label>
                    <label>
                      <span>Max Order Duration:</span>
                      <input type="number" defaultValue="24" min="1" max="168" />
                      <span>hours</span>
                    </label>
                  </div>
                </div>

                <div className="setting-card">
                  <h4>Price Oracle</h4>
                  <div className="setting-controls">
                    <label>
                      <span>Price Calculation Period:</span>
                      <select defaultValue="100">
                        <option value="50">Last 50 trades</option>
                        <option value="100">Last 100 trades</option>
                        <option value="200">Last 200 trades</option>
                      </select>
                    </label>
                  </div>
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>
    </div>
  );
};

export default AdminDashboard;
