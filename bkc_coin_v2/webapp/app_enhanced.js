/* global React, ReactDOM, Chart */

const { useEffect, useMemo, useRef, useState, useCallback } = React;

const tg = window.Telegram?.WebApp;
if (tg) {
  tg.ready();
  tg.expand();
  try {
    tg.setHeaderColor("#0a3a4e");
    tg.setBackgroundColor("#12253f");
  } catch {}
}

// Enhanced formatting functions
const format = (value) => Number(value || 0).toLocaleString("ru-RU");
const formatCurrency = (value, currency = "BKC") => `${format(value)} ${currency}`;
const formatPercent = (value) => `${(value * 100).toFixed(2)}%`;

// API configuration
const qp = new URLSearchParams(window.location.search);
const apiParam = qp.get("api");
const fallbackApi = window.location.hostname === "127.0.0.1" || window.location.hostname === "localhost"
  ? "http://127.0.0.1:8080"
  : window.location.origin;

const API_BASE = apiParam || fallbackApi;

// TonConnect integration
const [tonConnect, setTonConnect] = useState(null);
const [walletConnected, setWalletConnected] = useState(false);

// Initialize TonConnect
useEffect(() => {
  if (window.TonConnectUI) {
    const tonConnectUI = new window.TonConnectUI({
      manifestUrl: `${window.location.origin}/tonconnect-manifest.json`,
      buttonRootId: 'tonconnect-button'
    });
    
    setTonConnect(tonConnectUI);
    
    tonConnectUI.onStatusChange(wallet => {
      setWalletConnected(!!wallet);
      if (wallet) {
        console.log('Wallet connected:', wallet);
        showNotification('Кошелек успешно подключен!', 'success');
      }
    });
  }
}, []);

// Enhanced state management
const [user, setUser] = useState({
  id: tg?.initDataUnsafe?.user?.id || 12345,
  balance: 1000,
  energy: 1000,
  maxEnergy: 1000,
  tapPower: 1,
  level: 1,
  experience: 0,
  referrals: 0,
  achievements: []
});

const [rates, setRates] = useState({
  ton_usd: 5.0,
  bkc_usd: 0.001,
  ton_bkc: 5000,
  last_update: new Date().toISOString()
});

const [nfts, setNfts] = useState([]);
const [transactions, setTransactions] = useState([]);
const [loading, setLoading] = useState(false);
const [theme, setTheme] = useState(localStorage.getItem('bkc-theme') || 'dark');

// Enhanced API functions
const api = {
  // User operations
  getUser: async () => {
    try {
      const response = await fetch(`${API_BASE}/api/v1/user/${user.id}`);
      return await response.json();
    } catch (error) {
      console.error('Error fetching user:', error);
      return null;
    }
  },
  
  // Tap operations
  tap: async (count) => {
    try {
      const response = await fetch(`${API_BASE}/api/v1/tap`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: user.id, tap_count: count })
      });
      return await response.json();
    } catch (error) {
      console.error('Error tapping:', error);
      return null;
    }
  },
  
  // Rates operations
  getRates: async () => {
    try {
      const response = await fetch(`${API_BASE}/api/v1/rates/current`);
      return await response.json();
    } catch (error) {
      console.error('Error fetching rates:', error);
      return null;
    }
  },
  
  // NFT operations
  getNFTs: async () => {
    try {
      const response = await fetch(`${API_BASE}/api/v1/nfts`);
      return await response.json();
    } catch (error) {
      console.error('Error fetching NFTs:', error);
      return [];
    }
  },
  
  // Transaction operations
  getTransactions: async () => {
    try {
      const response = await fetch(`${API_BASE}/api/v1/transactions/${user.id}`);
      return await response.json();
    } catch (error) {
      console.error('Error fetching transactions:', error);
      return [];
    }
  }
};

// Enhanced notification system
const showNotification = (message, type = 'info') => {
  if (tg) {
    tg.showAlert(message);
  } else {
    // Fallback notification
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    notification.style.cssText = `
      position: fixed;
      top: 20px;
      left: 50%;
      transform: translateX(-50%);
      background: ${type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : '#3b82f6'};
      color: white;
      padding: 12px 20px;
      border-radius: 8px;
      z-index: 10000;
      animation: slideDown 0.3s ease;
    `;
    
    document.body.appendChild(notification);
    setTimeout(() => notification.remove(), 3000);
  }
};

// Enhanced tap handler with animations
const handleTap = useCallback(async () => {
  if (user.energy < 1) {
    showNotification('Недостаточно энергии!', 'error');
    return;
  }
  
  setLoading(true);
  
  try {
    const result = await api.tap(user.tapPower);
    
    if (result.success) {
      setUser(prev => ({
        ...prev,
        balance: prev.balance + result.reward,
        energy: prev.energy - user.tapPower,
        experience: prev.experience + result.experience
      }));
      
      // Add tap animation
      const tapButton = document.querySelector('.tap-button');
      tapButton.classList.add('success-animation');
      setTimeout(() => tapButton.classList.remove('success-animation'), 600);
      
      // Show floating reward
      showFloatingReward(result.reward);
    }
  } catch (error) {
    showNotification('Ошибка при тапе!', 'error');
  } finally {
    setLoading(false);
  }
}, [user.energy, user.tapPower]);

// Floating reward animation
const showFloatingReward = (reward) => {
  const reward = document.createElement('div');
  reward.textContent = `+${format(reward)} BKC`;
  reward.style.cssText = `
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: #3cd6c6;
    font-weight: bold;
    font-size: 24px;
    z-index: 10000;
    animation: floatUp 1s ease-out forwards;
    pointer-events: none;
  `;
  
  document.body.appendChild(reward);
  setTimeout(() => reward.remove(), 1000);
};

// Enhanced charts
const ChartsSection = () => {
  const balanceChartRef = useRef(null);
  const rateChartRef = useRef(null);
  
  useEffect(() => {
    // Initialize balance chart
    if (balanceChartRef.current) {
      new Chart(balanceChartRef.current, {
        type: 'line',
        data: {
          labels: ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'],
          datasets: [{
            label: 'Баланс BKC',
            data: [1000, 1200, 1150, 1300, 1250, 1400, 1500],
            borderColor: '#3cd6c6',
            backgroundColor: 'rgba(60, 214, 198, 0.1)',
            tension: 0.4,
            fill: true
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: { display: false }
          },
          scales: {
            y: {
              beginAtZero: true,
              grid: { color: 'rgba(255, 255, 255, 0.1)' },
              ticks: { color: '#94a3b8' }
            },
            x: {
              grid: { display: false },
              ticks: { color: '#94a3b8' }
            }
          }
        }
      });
    }
    
    // Initialize rate chart
    if (rateChartRef.current) {
      new Chart(rateChartRef.current, {
        type: 'line',
        data: {
          labels: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'],
          datasets: [{
            label: 'TON/USD',
            data: [5.0, 5.1, 5.05, 5.15, 5.08, 5.12, 5.10],
            borderColor: '#22d3ee',
            backgroundColor: 'rgba(34, 211, 238, 0.1)',
            tension: 0.4,
            fill: true
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: { display: false }
          },
          scales: {
            y: {
              beginAtZero: false,
              grid: { color: 'rgba(255, 255, 255, 0.1)' },
              ticks: { color: '#94a3b8' }
            },
            x: {
              grid: { display: false },
              ticks: { color: '#94a3b8' }
            }
          }
        }
      });
    }
  }, []);
  
  return (
    <div className="charts-grid">
      <div className="chart-container">
        <div className="chart-title">📈 История баланса</div>
        <canvas ref={balanceChartRef} height="150"></canvas>
      </div>
      
      <div className="chart-container">
        <div className="chart-title">💱 Курс TON/USD</div>
        <canvas ref={rateChartRef} height="150"></canvas>
      </div>
    </div>
  );
};

// Enhanced NFT marketplace
const NFTMarketplace = () => {
  const [selectedNFT, setSelectedNFT] = useState(null);
  const [pricingMode, setPricingMode] = useState('bkc'); // 'bkc' or 'ton'
  
  const handleCreateListing = async (nftId, price) => {
    if (!walletConnected) {
      showNotification('Сначала подключите кошелек!', 'error');
      return;
    }
    
    try {
      // Convert price based on mode
      const finalPrice = pricingMode === 'ton' 
        ? price * rates.ton_bkc 
        : price;
      
      const response = await fetch(`${API_BASE}/api/v1/nfts/${nftId}/list`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          price: finalPrice,
          currency: pricingMode,
          user_id: user.id
        })
      });
      
      if (response.ok) {
        showNotification('NFT успешно выставлен на продажу!', 'success');
        setSelectedNFT(null);
      }
    } catch (error) {
      showNotification('Ошибка при выставлении NFT!', 'error');
    }
  };
  
  return (
    <div className="card">
      <h3>🎨 NFT Маркетплейс</h3>
      
      <div className="tonconnect-container">
        <button 
          className="tonconnect-button"
          onClick={() => tonConnect?.openModal()}
        >
          {walletConnected ? '🟢 Кошелек подключен' : '🔗 Подключить TON кошелек'}
        </button>
      </div>
      
      <div className="pricing-toggle">
        <button 
          className={`btn ${pricingMode === 'bkc' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setPricingMode('bkc')}
        >
          💰 BKC
        </button>
        <button 
          className={`btn ${pricingMode === 'ton' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setPricingMode('ton')}
        >
          💎 TON
        </button>
      </div>
      
      <div className="nft-grid">
        {nfts.map(nft => (
          <div key={nft.id} className="nft-card">
            <div className="nft-image">
              <img src={nft.image} alt={nft.name} />
            </div>
            <div className="nft-info">
              <h4>{nft.name}</h4>
              <p>{nft.description}</p>
              <div className="nft-price">
                {pricingMode === 'bkc' ? (
                  <span>💰 {format(nft.price_bkc)} BKC</span>
                ) : (
                  <span>💎 {format(nft.price_ton)} TON</span>
                )}
              </div>
              <button 
                className="btn btn-primary"
                onClick={() => handleCreateListing(nft.id, pricingMode === 'bkc' ? nft.price_bkc : nft.price_ton)}
              >
                Выставить на продажу
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// Enhanced main app component
const App = () => {
  const [activeTab, setActiveTab] = useState('tap');
  
  // Load initial data
  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      
      try {
        const [userData, ratesData, nftsData, transactionsData] = await Promise.all([
          api.getUser(),
          api.getRates(),
          api.getNFTs(),
          api.getTransactions()
        ]);
        
        if (userData) setUser(userData);
        if (ratesData) setRates(ratesData);
        if (nftsData) setNfts(nftsData);
        if (transactionsData) setTransactions(transactionsData);
      } catch (error) {
        console.error('Error loading data:', error);
      } finally {
        setLoading(false);
      }
    };
    
    loadData();
    
    // Auto-refresh rates every 5 minutes
    const interval = setInterval(async () => {
      const ratesData = await api.getRates();
      if (ratesData) setRates(ratesData);
    }, 5 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, []);
  
  return (
    <div id="root">
      <div className="app">
        {/* Header */}
        <div className="card">
          <div className="user-header">
            <div className="user-info">
              <h2>BKC Coin Pro</h2>
              <div className="user-stats">
                <div className="stat">
                  <span className="label">Баланс:</span>
                  <span className="value">{formatCurrency(user.balance)}</span>
                </div>
                <div className="stat">
                  <span className="label">Энергия:</span>
                  <span className="value">{user.energy}/{user.maxEnergy}</span>
                </div>
                <div className="stat">
                  <span className="label">Уровень:</span>
                  <span className="value">{user.level}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        {/* Navigation */}
        <div className="card">
          <div className="nav-tabs">
            <button 
              className={`nav-tab ${activeTab === 'tap' ? 'active' : ''}`}
              onClick={() => setActiveTab('tap')}
            >
              👆 Тап
            </button>
            <button 
              className={`nav-tab ${activeTab === 'market' ? 'active' : ''}`}
              onClick={() => setActiveTab('market')}
            >
              🛍️ Маркет
            </button>
            <button 
              className={`nav-tab ${activeTab === 'stats' ? 'active' : ''}`}
              onClick={() => setActiveTab('stats')}
            >
              📊 Статистика
            </button>
            <button 
              className={`nav-tab ${activeTab === 'wallet' ? 'active' : ''}`}
              onClick={() => setActiveTab('wallet')}
            >
              💳 Кошелек
            </button>
          </div>
        </div>
        
        {/* Content */}
        {activeTab === 'tap' && (
          <div className="card">
            <div className="tap-section">
              <button 
                className="tap-button"
                onClick={handleTap}
                disabled={loading || user.energy < 1}
              >
                <span className="tap-icon">👆</span>
              </button>
              <div className="tap-info">
                <p>Сила тапа: {user.tapPower}</p>
                <p>Энергия: {user.energy}/{user.maxEnergy}</p>
              </div>
            </div>
          </div>
        )}
        
        {activeTab === 'market' && <NFTMarketplace />}
        {activeTab === 'stats' && <ChartsSection />}
        {activeTab === 'wallet' && (
          <div className="card">
            <h3>💳 Кошелек</h3>
            <div className="wallet-info">
              <p>TON/USD: ${rates.ton_usd}</p>
              <p>BKC/USD: ${rates.bkc_usd}</p>
              <p>TON/BKC: {format(rates.ton_bkc)}</p>
              <p>Последнее обновление: {new Date(rates.last_update).toLocaleString()}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// Render app
ReactDOM.render(<App />, document.getElementById('root'));

// Add global styles
const style = document.createElement('style');
style.textContent = `
  @keyframes slideDown {
    from { transform: translate(-50%, -100%); opacity: 0; }
    to { transform: translate(-50%, 0); opacity: 1; }
  }
  
  @keyframes floatUp {
    from { transform: translate(-50%, -50%); opacity: 1; }
    to { transform: translate(-50%, -150%); opacity: 0; }
  }
  
  .nav-tabs {
    display: flex;
    gap: 8px;
  }
  
  .nav-tab {
    flex: 1;
    padding: 12px;
    background: var(--bg-glass);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.3s ease;
  }
  
  .nav-tab.active {
    background: var(--accent-gradient);
    color: var(--bg-primary);
    border-color: var(--accent-primary);
  }
  
  .tap-section {
    text-align: center;
    padding: 20px;
  }
  
  .tap-icon {
    font-size: 48px;
  }
  
  .nft-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
    margin-top: 16px;
  }
  
  .nft-card {
    background: var(--bg-glass);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 12px;
    transition: all 0.3s ease;
  }
  
  .nft-card:hover {
    transform: translateY(-2px);
    border-color: var(--accent-primary);
  }
  
  .nft-image img {
    width: 100%;
    height: 120px;
    object-fit: cover;
    border-radius: var(--radius-sm);
  }
  
  .pricing-toggle {
    display: flex;
    gap: 8px;
    margin: 16px 0;
  }
  
  .pricing-toggle .btn {
    flex: 1;
  }
  
  .user-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
  }
  
  .user-stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
    margin-top: 12px;
  }
  
  .stat {
    text-align: center;
  }
  
  .stat .label {
    display: block;
    font-size: 12px;
    color: var(--text-muted);
    margin-bottom: 4px;
  }
  
  .stat .value {
    display: block;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }
  
  .wallet-info {
    padding: 16px;
  }
  
  .wallet-info p {
    margin: 8px 0;
    color: var(--text-secondary);
  }
`;
document.head.appendChild(style);
