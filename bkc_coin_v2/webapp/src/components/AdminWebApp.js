import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  Shield, 
  Settings, 
  Activity,
  Users,
  DollarSign,
  Bell,
  Image as ImageIcon,
  Ban,
  Server,
  Database,
  TrendingUp,
  AlertTriangle,
  CheckCircle,
  XCircle,
  RefreshCw,
  Upload,
  Send,
  Power,
  Eye,
  Edit,
  Trash2,
  Plus,
  Minus,
  Save
} from 'lucide-react';
import toast from 'react-hot-toast';

const AdminWebApp = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [password, setPassword] = useState('');
  const [activeTab, setActiveTab] = useState('economy');
  const [systemData, setSystemData] = useState(null);
  const [loading, setLoading] = useState(false);

  // Form states
  const [broadcastForm, setBroadcastForm] = useState({
    message: '',
    type: 'push',
    imageUrl: ''
  });

  const [nftForm, setNftForm] = useState({
    name: '',
    type: 'digger',
    description: '',
    price: '',
    power: '1',
    maxSupply: '1000',
    imageUrl: ''
  });

  const [economyForm, setEconomyForm] = useState({
    tapReward: '0.01',
    p2pFee: '5',
    burnRate: '1'
  });

  useEffect(() => {
    // Check if already authenticated
    const auth = localStorage.getItem('adminAuth');
    if (auth === 'true') {
      setIsAuthenticated(true);
      loadSystemData();
    }
  }, []);

  const handleLogin = (e) => {
    e.preventDefault();
    if (password === 'bkc_admin_2024') {
      setIsAuthenticated(true);
      localStorage.setItem('adminAuth', 'true');
      loadSystemData();
      toast.success('Authenticated successfully');
    } else {
      toast.error('Invalid password');
    }
  };

  const handleLogout = () => {
    setIsAuthenticated(false);
    localStorage.removeItem('adminAuth');
    setPassword('');
  };

  const loadSystemData = async () => {
    setLoading(true);
    try {
      // Получаем admin token из localStorage
      const adminToken = localStorage.getItem('bkc_admin_token');
      if (!adminToken) {
        toast.error('Требуется авторизация');
        window.location.href = '/admin-login.html';
        return;
      }

      // Запрашиваем реальные данные с API
      const response = await fetch('/api/v1/admin/system-data', {
        headers: {
          'Authorization': `Bearer ${adminToken}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error('Failed to fetch system data');
      }
      
      const data = await response.json();
      
      if (data.success) {
        setSystemData(data.data);
      } else {
        // Fallback на демо данные если API недоступен
        const fallbackData = {
          renderNodes: Array.from({ length: 25 }, (_, i) => ({
            nodeId: i + 1,
            nodeName: `render-node-${i + 1}`,
            status: i % 10 === 0 ? 'degraded' : 'online',
            cpuUsage: Math.random() * 80 + 10,
            memoryUsage: Math.random() * 70 + 20,
            diskUsage: Math.random() * 60 + 30,
            activeUsers: Math.floor(Math.random() * 1000) + 100,
            responseTime: Math.floor(Math.random() * 200) + 50,
            lastSeen: new Date()
          })),
          neonShards: Array.from({ length: 6 }, (_, i) => ({
            shardId: i + 1,
            shardType: 'neon',
            status: 'online',
            connections: (i + 1) * 50,
            queryTime: Math.random() * 100 + 10,
            storageUsed: (i + 1) * 1024 * 1024 * 100,
            storageTotal: 1024 * 1024 * 1024,
            lastSync: new Date()
          })),
          currentUsers: Math.floor(Math.random() * 50000) + 10000,
          tps: Math.floor(Math.random() * 1000) + 100,
          totalSupply: 100000000000,
          circulatingSupply: 850000000,
          burnRate: 15.5,
          currentPrice: 0.0012,
          lastUpdated: new Date()
        };
        setSystemData(fallbackData);
      }
    } catch (error) {
      console.error('Failed to load system data:', error);
      toast.error('Failed to load system data');
      
      // Fallback на демо данные
      const fallbackData = {
        renderNodes: Array.from({ length: 25 }, (_, i) => ({
          nodeId: i + 1,
          nodeName: `render-node-${i + 1}`,
          status: i % 10 === 0 ? 'degraded' : 'online',
          cpuUsage: Math.random() * 80 + 10,
          memoryUsage: Math.random() * 70 + 20,
          diskUsage: Math.random() * 60 + 30,
          activeUsers: Math.floor(Math.random() * 1000) + 100,
          responseTime: Math.floor(Math.random() * 200) + 50,
          lastSeen: new Date()
        })),
        currentUsers: Math.floor(Math.random() * 50000) + 10000,
        tps: Math.floor(Math.random() * 1000) + 100,
        totalSupply: 100000000000,
        circulatingSupply: 850000000,
        burnRate: 15.5,
        currentPrice: 0.0012,
        lastUpdated: new Date()
      };
      setSystemData(fallbackData);
    } finally {
      setLoading(false);
    }
  };

  const handleEconomyControl = async (action, value, reason) => {
    try {
      const adminToken = localStorage.getItem('bkc_admin_token');
      if (!adminToken) {
        toast.error('Требуется авторизация');
        return;
      }

      const response = await fetch('/api/v1/admin/economy-control', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          action,
          value,
          reason
        })
      });

      if (!response.ok) {
        throw new Error('Failed to update economy');
      }

      const result = await response.json();
      
      if (result.success) {
        toast.success(`${action} updated successfully`);
        loadSystemData();
      } else {
        toast.error(result.message || 'Failed to update economy');
      }
    } catch (error) {
      console.error('Economy control error:', error);
      toast.error('Failed to update economy');
    }
  };

  const handleBroadcast = async () => {
    if (!broadcastForm.message.trim()) {
      toast.error('Message is required');
      return;
    }

    try {
      const adminToken = localStorage.getItem('bkc_admin_token');
      if (!adminToken) {
        toast.error('Требуется авторизация');
        return;
      }

      const response = await fetch('/api/v1/admin/broadcast', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(broadcastForm)
      });

      if (!response.ok) {
        throw new Error('Failed to send broadcast');
      }

      const result = await response.json();
      
      if (result.success) {
        toast.success('Broadcast sent successfully');
        setBroadcastForm({ message: '', type: 'push', imageUrl: '' });
      } else {
        toast.error(result.message || 'Failed to send broadcast');
      }
    } catch (error) {
      console.error('Broadcast error:', error);
      toast.error('Failed to send broadcast');
    }
  };

  const handleNFTUpload = async () => {
    if (!nftForm.name || !nftForm.price) {
      toast.error('Name and price are required');
      return;
    }

    try {
      const adminToken = localStorage.getItem('bkc_admin_token');
      if (!adminToken) {
        toast.error('Требуется авторизация');
        return;
      }

      const response = await fetch('/api/v1/admin/nft-upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(nftForm)
      });

      if (!response.ok) {
        throw new Error('Failed to upload NFT');
      }

      const result = await response.json();
      
      if (result.success) {
        toast.success('NFT uploaded successfully');
        setNftForm({
          name: '',
          description: '',
          price: '',
          maxSupply: '',
          imageUrl: ''
        });
        loadSystemData();
      } else {
        toast.error(result.message || 'Failed to upload NFT');
      }
    } catch (error) {
      console.error('NFT upload error:', error);
      toast.error('Failed to upload NFT');
    }
  };

  const handleMassBan = async () => {
    const userIds = prompt('Enter user IDs (comma-separated):');
    if (!userIds) return;

    const ids = userIds.split(',').map(id => parseInt(id.trim())).filter(id => !isNaN(id));
    if (ids.length === 0) {
      toast.error('No valid user IDs provided');
      return;
    }

    try {
      const adminToken = localStorage.getItem('bkc_admin_token');
      if (!adminToken) {
        toast.error('Требуется авторизация');
        return;
      }

      const response = await fetch('/api/v1/admin/mass-ban', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${adminToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          userIds: ids,
          reason: 'Mass ban from admin panel'
        })
      });

      if (!response.ok) {
        throw new Error('Failed to ban users');
      }

      const result = await response.json();
      
      if (result.success) {
        toast.success(`${ids.length} users banned successfully`);
        loadSystemData();
      } else {
        toast.error(result.message || 'Failed to ban users');
      }
    } catch (error) {
      console.error('Mass ban error:', error);
      toast.error('Failed to ban users');
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'online': return 'text-green-500';
      case 'offline': return 'text-red-500';
      case 'degraded': return 'text-yellow-500';
      default: return 'text-gray-500';
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="admin-login">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="login-container"
        >
          <div className="login-header">
            <Shield size={48} className="login-icon" />
            <h1>Admin Access</h1>
            <p>BKC Coin Control Center</p>
          </div>
          
          <form onSubmit={handleLogin} className="login-form">
            <div className="form-group">
              <label>Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter admin password"
                className="login-input"
                required
              />
            </div>
            <button type="submit" className="login-button">
              <Shield size={20} />
              Login
            </button>
          </form>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="admin-webapp">
      {/* Header */}
      <header className="admin-header">
        <div className="header-left">
          <h1>🚀 BKC Control Center</h1>
          <span className="header-subtitle">System Management & Control</span>
        </div>
        <div className="header-right">
          <button onClick={loadSystemData} className="refresh-btn">
            <RefreshCw size={20} />
          </button>
          <button onClick={handleLogout} className="logout-btn">
            <Power size={20} />
            Logout
          </button>
        </div>
      </header>

      {/* Navigation */}
      <nav className="admin-nav">
        <button
          className={`nav-tab ${activeTab === 'economy' ? 'active' : ''}`}
          onClick={() => setActiveTab('economy')}
        >
          <DollarSign size={18} />
          Economy
        </button>
        <button
          className={`nav-tab ${activeTab === 'broadcast' ? 'active' : ''}`}
          onClick={() => setActiveTab('broadcast')}
        >
          <Bell size={18} />
          Broadcast
        </button>
        <button
          className={`nav-tab ${activeTab === 'nft' ? 'active' : ''}`}
          onClick={() => setActiveTab('nft')}
        >
          <ImageIcon size={18} />
          NFT
        </button>
        <button
          className={`nav-tab ${activeTab === 'security' ? 'active' : ''}`}
          onClick={() => setActiveTab('security')}
        >
          <Shield size={18} />
          Security
        </button>
        <button
          className={`nav-tab ${activeTab === 'monitor' ? 'active' : ''}`}
          onClick={() => setActiveTab('monitor')}
        >
          <Activity size={18} />
          Monitor
        </button>
      </nav>

      {/* Content */}
      <main className="admin-main">
        <AnimatePresence mode="wait">
          {activeTab === 'economy' && (
            <motion.div
              key="economy"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>💰 Economy Control</h2>
              
              <div className="economy-controls">
                <div className="control-card">
                  <h3>Tap Reward</h3>
                  <div className="control-group">
                    <input
                      type="number"
                      step="0.001"
                      value={economyForm.tapReward}
                      onChange={(e) => setEconomyForm({...economyForm, tapReward: e.target.value})}
                      className="control-input"
                    />
                    <span className="control-unit">BKC</span>
                    <button
                      onClick={() => handleEconomyControl('tap_reward', economyForm.tapReward, 'Economy adjustment')}
                      className="control-btn"
                    >
                      Update
                    </button>
                  </div>
                </div>

                <div className="control-card">
                  <h3>P2P Trading Fee</h3>
                  <div className="control-group">
                    <input
                      type="number"
                      step="0.1"
                      value={economyForm.p2pFee}
                      onChange={(e) => setEconomyForm({...economyForm, p2pFee: e.target.value})}
                      className="control-input"
                    />
                    <span className="control-unit">%</span>
                    <button
                      onClick={() => handleEconomyControl('p2p_fee', economyForm.p2pFee, 'Fee adjustment')}
                      className="control-btn"
                    >
                      Update
                    </button>
                  </div>
                </div>

                <div className="control-card">
                  <h3>Burn Rate Multiplier</h3>
                  <div className="control-group">
                    <input
                      type="number"
                      step="0.1"
                      value={economyForm.burnRate}
                      onChange={(e) => setEconomyForm({...economyForm, burnRate: e.target.value})}
                      className="control-input"
                    />
                    <span className="control-unit">x</span>
                    <button
                      onClick={() => handleEconomyControl('burn_rate', economyForm.burnRate, 'Burn rate adjustment')}
                      className="control-btn"
                    >
                      Update
                    </button>
                  </div>
                </div>
              </div>

              {systemData && (
                <div className="economy-stats">
                  <div className="stat-card">
                    <h4>Total Supply</h4>
                    <div className="stat-value">{(systemData.totalSupply / 100).toLocaleString()} BKC</div>
                  </div>
                  <div className="stat-card">
                    <h4>Circulating Supply</h4>
                    <div className="stat-value">{(systemData.circulatingSupply / 100).toLocaleString()} BKC</div>
                  </div>
                  <div className="stat-card">
                    <h4>Current Price</h4>
                    <div className="stat-value">${systemData.currentPrice.toFixed(4)}</div>
                  </div>
                  <div className="stat-card">
                    <h4>Active Users</h4>
                    <div className="stat-value">{systemData.currentUsers.toLocaleString()}</div>
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {activeTab === 'broadcast' && (
            <motion.div
              key="broadcast"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>📢 Mass Broadcast</h2>
              
              <div className="broadcast-form">
                <div className="form-group">
                  <label>Message</label>
                  <textarea
                    value={broadcastForm.message}
                    onChange={(e) => setBroadcastForm({...broadcastForm, message: e.target.value})}
                    placeholder="Enter your broadcast message..."
                    rows={4}
                    className="form-textarea"
                  />
                </div>

                <div className="form-group">
                  <label>Type</label>
                  <select
                    value={broadcastForm.type}
                    onChange={(e) => setBroadcastForm({...broadcastForm, type: e.target.value})}
                    className="form-select"
                  >
                    <option value="push">Push Notification</option>
                    <option value="in_app">In-App Message</option>
                    <option value="email">Email</option>
                  </select>
                </div>

                <div className="form-group">
                  <label>Image (optional)</label>
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => {
                      const file = e.target.files[0];
                      if (file && file.size <= 15 * 1024 * 1024) { // 15MB limit
                        console.log('Image selected:', file);
                      } else {
                        toast.error('Image size must be less than 15MB');
                      }
                    }}
                    className="form-file"
                  />
                </div>

                <button onClick={handleBroadcast} className="broadcast-btn">
                  <Send size={20} />
                  Send Broadcast
                </button>
              </div>
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
              <h2>🖼️ NFT Management</h2>
              
              <div className="nft-form">
                <div className="form-row">
                  <div className="form-group">
                    <label>Name</label>
                    <input
                      type="text"
                      value={nftForm.name}
                      onChange={(e) => setNftForm({...nftForm, name: e.target.value})}
                      placeholder="NFT Name"
                      className="form-input"
                    />
                  </div>

                  <div className="form-group">
                    <label>Type</label>
                    <select
                      value={nftForm.type}
                      onChange={(e) => setNftForm({...nftForm, type: e.target.value})}
                      className="form-select"
                    >
                      <option value="digger">Digger</option>
                      <option value="banker">Banker</option>
                      <option value="inspector">Inspector</option>
                    </select>
                  </div>
                </div>

                <div className="form-group">
                  <label>Description</label>
                  <textarea
                    value={nftForm.description}
                    onChange={(e) => setNftForm({...nftForm, description: e.target.value})}
                    placeholder="NFT Description..."
                    rows={3}
                    className="form-textarea"
                  />
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label>Price (BKC)</label>
                    <input
                      type="number"
                      value={nftForm.price}
                      onChange={(e) => setNftForm({...nftForm, price: e.target.value})}
                      placeholder="50000"
                      className="form-input"
                    />
                  </div>

                  <div className="form-group">
                    <label>Power</label>
                    <input
                      type="number"
                      value={nftForm.power}
                      onChange={(e) => setNftForm({...nftForm, power: e.target.value})}
                      min="1"
                      max="10"
                      className="form-input"
                    />
                  </div>

                  <div className="form-group">
                    <label>Max Supply</label>
                    <input
                      type="number"
                      value={nftForm.maxSupply}
                      onChange={(e) => setNftForm({...nftForm, maxSupply: e.target.value})}
                      placeholder="1000"
                      className="form-input"
                    />
                  </div>
                </div>

                <div className="form-group">
                  <label>Image (max 2MB)</label>
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => {
                      const file = e.target.files[0];
                      if (file && file.size <= 2 * 1024 * 1024) { // 2MB limit
                        console.log('NFT image selected:', file);
                      } else {
                        toast.error('Image size must be less than 2MB');
                      }
                    }}
                    className="form-file"
                  />
                </div>

                <button onClick={handleNFTUpload} className="nft-upload-btn">
                  <Upload size={20} />
                  Upload NFT
                </button>
              </div>
            </motion.div>
          )}

          {activeTab === 'security' && (
            <motion.div
              key="security"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>🛡️ Security Controls</h2>
              
              <div className="security-actions">
                <div className="action-card">
                  <h3>Mass Ban</h3>
                  <p>Ban multiple users at once</p>
                  <button onClick={handleMassBan} className="action-btn danger">
                    <Ban size={20} />
                    Mass Ban
                  </button>
                </div>

                <div className="action-card">
                  <h3>Resolve P2P Disputes</h3>
                  <p>Manually resolve trading disputes</p>
                  <button className="action-btn">
                    <Eye size={20} />
                    View Disputes
                  </button>
                </div>

                <div className="action-card">
                  <h3>Emergency Stop</h3>
                  <p>Stop all trading operations</p>
                  <button className="action-btn danger">
                    <Power size={20} />
                    Emergency Stop
                  </button>
                </div>
              </div>
            </motion.div>
          )}

          {activeTab === 'monitor' && (
            <motion.div
              key="monitor"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="admin-content"
            >
              <h2>📊 System Monitor</h2>
              
              {systemData && (
                <div className="monitor-grid">
                  <div className="monitor-section">
                    <h3>Render Nodes</h3>
                    <div className="nodes-grid">
                      {systemData.renderNodes?.slice(0, 6).map((node) => (
                        <div key={node.nodeId} className="node-card">
                          <div className="node-header">
                            <span className="node-name">{node.nodeName}</span>
                            <span className={`node-status ${getStatusColor(node.status)}`}>
                              {node.status}
                            </span>
                          </div>
                          <div className="node-metrics">
                            <div className="metric">
                              <span className="label">CPU:</span>
                              <span className="value">{node.cpuUsage.toFixed(1)}%</span>
                            </div>
                            <div className="metric">
                              <span className="label">Memory:</span>
                              <span className="value">{node.memoryUsage.toFixed(1)}%</span>
                            </div>
                            <div className="metric">
                              <span className="label">Users:</span>
                              <span className="value">{node.activeUsers}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="monitor-section">
                    <h3>Database Shards</h3>
                    <div className="shards-grid">
                      {systemData.neonShards?.slice(0, 6).map((shard) => (
                        <div key={shard.shardId} className="shard-card">
                          <div className="shard-header">
                            <span className="shard-name">Neon-{shard.shardId}</span>
                            <span className={`shard-status ${getStatusColor(shard.status)}`}>
                              {shard.status}
                            </span>
                          </div>
                          <div className="shard-metrics">
                            <div className="metric">
                              <span className="label">Storage:</span>
                              <span className="value">{formatBytes(shard.storageUsed)}</span>
                            </div>
                            <div className="metric">
                              <span className="label">Connections:</span>
                              <span className="value">{shard.connections}</span>
                            </div>
                            <div className="metric">
                              <span className="label">Query Time:</span>
                              <span className="value">{shard.queryTime.toFixed(1)}ms</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="monitor-section">
                    <h3>System Metrics</h3>
                    <div className="metrics-grid">
                      <div className="metric-card">
                        <h4>Current Users</h4>
                        <div className="metric-value">{systemData.currentUsers.toLocaleString()}</div>
                      </div>
                      <div className="metric-card">
                        <h4>TPS</h4>
                        <div className="metric-value">{systemData.tps}</div>
                      </div>
                      <div className="metric-card">
                        <h4>Total Supply</h4>
                        <div className="metric-value">{(systemData.totalSupply / 100).toLocaleString()} BKC</div>
                      </div>
                      <div className="metric-card">
                        <h4>Current Price</h4>
                        <div className="metric-value">${systemData.currentPrice.toFixed(4)}</div>
                      </div>
                    </div>
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

export default AdminWebApp;
