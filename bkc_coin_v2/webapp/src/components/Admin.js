import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Database, 
  Server, 
  Users, 
  TrendingUp, 
  AlertTriangle, 
  CheckCircle, 
  XCircle, 
  Activity,
  DollarSign,
  Zap,
  BarChart3,
  RefreshCw,
  Settings,
  LogOut,
  Menu
} from 'lucide-react';
import toast from 'react-hot-toast';

const Admin = ({ user, onNavigate, apiService }) => {
  const [activeTab, setActiveTab] = useState('overview');
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  // Data states
  const [systemStats, setSystemStats] = useState(null);
  const [clusterHealth, setClusterHealth] = useState(null);
  const [userList, setUserList] = useState([]);
  const [transactionLogs, setTransactionLogs] = useState([]);
  const [loadBalancerStats, setLoadBalancerStats] = useState(null);

  // Pagination
  const [userPage, setUserPage] = useState(1);
  const [logPage, setLogPage] = useState(1);

  useEffect(() => {
    if (user?.isAdmin) {
      loadAdminData();
    }
  }, [user]);

  const loadAdminData = async () => {
    try {
      setLoading(true);
      const [
        statsRes,
        healthRes,
        usersRes,
        logsRes
      ] = await Promise.all([
        apiService.getSystemStats(),
        apiService.getClusterHealth(),
        apiService.getUserList(userPage),
        apiService.getTransactionLogs(logPage)
      ]);

      setSystemStats(statsRes);
      setClusterHealth(healthRes);
      setUserList(usersRes.users || []);
      setTransactionLogs(logsRes.logs || []);
      setLoadBalancerStats(apiService.getLoadBalancerStats());
    } catch (error) {
      console.error('Failed to load admin data:', error);
      toast.error('Failed to load admin dashboard');
    } finally {
      setLoading(false);
    }
  };

  const refreshData = async () => {
    setRefreshing(true);
    await loadAdminData();
    setRefreshing(false);
    toast.success('Dashboard refreshed');
  };

  const handleBanUser = async (userId, reason) => {
    try {
      await apiService.banUser(userId, reason, 7); // 7 days ban
      toast.success('User banned successfully');
      loadAdminData();
    } catch (error) {
      console.error('Failed to ban user:', error);
      toast.error('Failed to ban user');
    }
  };

  const formatNumber = (num) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  };

  const formatDate = (timestamp) => {
    return new Date(timestamp).toLocaleString();
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'healthy': return 'text-green-500';
      case 'warning': return 'text-yellow-500';
      case 'error': return 'text-red-500';
      default: return 'text-gray-500';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'healthy': return <CheckCircle size={16} />;
      case 'warning': return <AlertTriangle size={16} />;
      case 'error': return <XCircle size={16} />;
      default: return <Activity size={16} />;
    }
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
    <div className="admin-container">
      {/* Header */}
      <header className="admin-header">
        <div className="header-left">
          <button 
            className="menu-button"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
          <h1>Admin Dashboard</h1>
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
          <button 
            className="logout-button"
            onClick={() => onNavigate('/')}
          >
            <LogOut size={20} />
          </button>
        </div>
      </header>

      {/* Navigation Tabs */}
      <nav className="admin-nav">
        <button
          className={`nav-tab ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          <BarChart3 size={18} />
          Overview
        </button>
        <button
          className={`nav-tab ${activeTab === 'cluster' ? 'active' : ''}`}
          onClick={() => setActiveTab('cluster')}
        >
          <Database size={18} />
          Cluster Health
        </button>
        <button
          className={`nav-tab ${activeTab === 'users' ? 'active' : ''}`}
          onClick={() => setActiveTab('users')}
        >
          <Users size={18} />
          Users
        </button>
        <button
          className={`nav-tab ${activeTab === 'transactions' ? 'active' : ''}`}
          onClick={() => setActiveTab('transactions')}
        >
          <DollarSign size={18} />
          Transactions
        </button>
        <button
          className={`nav-tab ${activeTab === 'loadbalancer' ? 'active' : ''}`}
          onClick={() => setActiveTab('loadbalancer')}
        >
          <Server size={18} />
          Load Balancer
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
              <h2>System Overview</h2>
              
              {systemStats && (
                <div className="stats-grid">
                  <div className="stat-card">
                    <div className="stat-header">
                      <Users size={24} className="stat-icon" />
                      <span className="stat-label">Total Users</span>
                    </div>
                    <div className="stat-value">{formatNumber(systemStats.totalUsers)}</div>
                    <div className="stat-change">+{systemStats.newUsersToday} today</div>
                  </div>

                  <div className="stat-card">
                    <div className="stat-header">
                      <DollarSign size={24} className="stat-icon" />
                      <span className="stat-label">Total Supply</span>
                    </div>
                    <div className="stat-value">{formatNumber(systemStats.totalSupply)} BKC</div>
                    <div className="stat-change">+{systemStats.mintedToday} today</div>
                  </div>

                  <div className="stat-card">
                    <div className="stat-header">
                      <Activity size={24} className="stat-icon" />
                      <span className="stat-label">Active Sessions</span>
                    </div>
                    <div className="stat-value">{formatNumber(systemStats.activeSessions)}</div>
                    <div className="stat-change">Peak: {systemStats.peakSessions}</div>
                  </div>

                  <div className="stat-card">
                    <div className="stat-header">
                      <Zap size={24} className="stat-icon" />
                      <span className="stat-label">Taps/Second</span>
                    </div>
                    <div className="stat-value">{systemStats.tapsPerSecond}</div>
                    <div className="stat-change">24h avg: {systemStats.avgTps}</div>
                  </div>

                  <div className="stat-card">
                    <div className="stat-header">
                      <TrendingUp size={24} className="stat-icon" />
                      <span className="stat-label">P2P Volume</span>
                    </div>
                    <div className="stat-value">{formatNumber(systemStats.p2pVolume)} BKC</div>
                    <div className="stat-change">{systemStats.p2pTransactions} trades</div>
                  </div>

                  <div className="stat-card">
                    <div className="stat-header">
                      <Server size={24} className="stat-icon" />
                      <span className="stat-label">Server Load</span>
                    </div>
                    <div className="stat-value">{systemStats.avgCpuUsage}%</div>
                    <div className="stat-change">Memory: {systemStats.avgMemoryUsage}%</div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'cluster' && (
            <motion.div
              key="cluster"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>Cluster Health</h2>
              
              {clusterHealth && (
                <div className="cluster-grid">
                  {/* Neon Shards */}
                  <div className="cluster-section">
                    <h3>Neon Shards (User Balances)</h3>
                    <div className="shard-grid">
                      {clusterHealth.neonShards?.map((shard, index) => (
                        <div key={`neon-${index}`} className="shard-card">
                          <div className="shard-header">
                            <span className="shard-name">Shard {index + 1}</span>
                            <span className={`shard-status ${getStatusColor(shard.status)}`}>
                              {getStatusIcon(shard.status)}
                            </span>
                          </div>
                          <div className="shard-metrics">
                            <div className="metric">
                              <span className="metric-label">Connections:</span>
                              <span className="metric-value">{shard.connections}</span>
                            </div>
                            <div className="metric">
                              <span className="metric-label">Latency:</span>
                              <span className="metric-value">{shard.latency}ms</span>
                            </div>
                            <div className="metric">
                              <span className="metric-label">Users:</span>
                              <span className="metric-value">{formatNumber(shard.userCount)}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Redis Shards */}
                  <div className="cluster-section">
                    <h3>Redis Shards (Energy/Anti-Cheat)</h3>
                    <div className="shard-grid">
                      {clusterHealth.redisShards?.map((shard, index) => (
                        <div key={`redis-${index}`} className="shard-card">
                          <div className="shard-header">
                            <span className="shard-name">Shard {index + 1}</span>
                            <span className={`shard-status ${getStatusColor(shard.status)}`}>
                              {getStatusIcon(shard.status)}
                            </span>
                          </div>
                          <div className="shard-metrics">
                            <div className="metric">
                              <span className="metric-label">Memory:</span>
                              <span className="metric-value">{shard.memoryUsage}%</span>
                            </div>
                            <div className="metric">
                              <span className="metric-label">Commands/s:</span>
                              <span className="metric-value">{shard.commandsPerSec}</span>
                            </div>
                            <div className="metric">
                              <span className="metric-label">Keys:</span>
                              <span className="metric-value">{formatNumber(shard.keyCount)}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* CockroachDB Shards */}
                  <div className="cluster-section">
                    <h3>CockroachDB Shards (Audit Logs)</h3>
                    <div className="shard-grid">
                      {clusterHealth.cockroachShards?.map((shard, index) => (
                        <div key={`cockroach-${index}`} className="shard-card">
                          <div className="shard-header">
                            <span className="shard-name">Shard {index + 1}</span>
                            <span className={`shard-status ${getStatusColor(shard.status)}`}>
                              {getStatusIcon(shard.status)}
                            </span>
                          </div>
                          <div className="shard-metrics">
                            <div className="metric">
                              <span className="metric-label">Storage:</span>
                              <span className="metric-value">{shard.storageUsage}GB</span>
                            </div>
                            <div className="metric">
                              <span className="metric-label">Queries/s:</span>
                              <span className="metric-value">{shard.queriesPerSec}</span>
                            </div>
                            <div className="metric">
                              <span className="metric-label">Replicas:</span>
                              <span className="metric-value">{shard.replicas}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'users' && (
            <motion.div
              key="users"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>User Management</h2>
              
              <div className="user-table-container">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>User ID</th>
                      <th>Username</th>
                      <th>Balance</th>
                      <th>Level</th>
                      <th>Status</th>
                      <th>Last Active</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {userList.map((user) => (
                      <tr key={user.id}>
                        <td>{user.id}</td>
                        <td>{user.username || user.first_name}</td>
                        <td>{formatNumber(user.balance)} BKC</td>
                        <td>{user.level}</td>
                        <td>
                          <span className={`status-badge ${user.isBanned ? 'banned' : 'active'}`}>
                            {user.isBanned ? 'Banned' : 'Active'}
                          </span>
                        </td>
                        <td>{formatDate(user.lastActive)}</td>
                        <td>
                          <div className="action-buttons">
                            {!user.isBanned && (
                              <button
                                className="ban-button"
                                onClick={() => handleBanUser(user.id, 'Admin action')}
                              >
                                Ban
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                
                <div className="pagination">
                  <button
                    onClick={() => setUserPage(Math.max(1, userPage - 1))}
                    disabled={userPage === 1}
                  >
                    Previous
                  </button>
                  <span>Page {userPage}</span>
                  <button onClick={() => setUserPage(userPage + 1)}>
                    Next
                  </button>
                </div>
              </div>
            </motion.div>
          )}

          {activeTab === 'transactions' && (
            <motion.div
              key="transactions"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>Transaction Logs</h2>
              
              <div className="transaction-table-container">
                <table className="admin-table">
                  <thead>
                    <tr>
                      <th>Transaction ID</th>
                      <th>User ID</th>
                      <th>Type</th>
                      <th>Amount</th>
                      <th>Status</th>
                      <th>Timestamp</th>
                    </tr>
                  </thead>
                  <tbody>
                    {transactionLogs.map((log) => (
                      <tr key={log.id}>
                        <td>{log.id.substring(0, 8)}...</td>
                        <td>{log.userId}</td>
                        <td>{log.type}</td>
                        <td>{formatNumber(log.amount)} BKC</td>
                        <td>
                          <span className={`status-badge ${log.status}`}>
                            {log.status}
                          </span>
                        </td>
                        <td>{formatDate(log.timestamp)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                
                <div className="pagination">
                  <button
                    onClick={() => setLogPage(Math.max(1, logPage - 1))}
                    disabled={logPage === 1}
                  >
                    Previous
                  </button>
                  <span>Page {logPage}</span>
                  <button onClick={() => setLogPage(logPage + 1)}>
                    Next
                  </button>
                </div>
              </div>
            </motion.div>
          )}

          {activeTab === 'loadbalancer' && (
            <motion.div
              key="loadbalancer"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>Load Balancer Status</h2>
              
              {loadBalancerStats && (
                <div className="loadbalancer-stats">
                  <div className="lb-overview">
                    <div className="lb-stat">
                      <span className="lb-label">Total Nodes:</span>
                      <span className="lb-value">{loadBalancerStats.totalNodes}</span>
                    </div>
                    <div className="lb-stat">
                      <span className="lb-label">Healthy Nodes:</span>
                      <span className="lb-value text-green-500">{loadBalancerStats.healthyNodes}</span>
                    </div>
                    <div className="lb-stat">
                      <span className="lb-label">Current Node:</span>
                      <span className="lb-value">{loadBalancerStats.currentNode}</span>
                    </div>
                  </div>

                  <h3>Node Status</h3>
                  <div className="node-grid">
                    {loadBalancerStats.nodes?.map((node, index) => (
                      <div key={index} className="node-card">
                        <div className="node-header">
                          <span className="node-name">Node {index + 1}</span>
                          <span className={`node-status ${getStatusColor(node.healthy ? 'healthy' : 'error')}`}>
                            {getStatusIcon(node.healthy ? 'healthy' : 'error')}
                          </span>
                        </div>
                        <div className="node-details">
                          <div className="node-metric">
                            <span>URL:</span>
                            <span className="text-xs">{node.url}</span>
                          </div>
                          <div className="node-metric">
                            <span>Response Time:</span>
                            <span>{node.responseTime}ms</span>
                          </div>
                          <div className="node-metric">
                            <span>Error Count:</span>
                            <span>{node.errorCount}</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </main>
    </div>
  );
};

export default Admin;
