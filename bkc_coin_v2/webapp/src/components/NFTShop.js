import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  ShoppingCart, 
  TrendingUp, 
  Zap, 
  Shield, 
  Star,
  Info,
  ExternalLink,
  RefreshCw
} from 'lucide-react';
import toast from 'react-hot-toast';

const NFTShop = ({ user, onNavigate, apiService }) => {
  const [nftPrices, setNftPrices] = useState({});
  const [exchangeRate, setExchangeRate] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadNFTPrices();
  }, []);

  const loadNFTPrices = async () => {
    try {
      const data = await apiService.getNFTPrices();
      setNftPrices(data.prices);
      setExchangeRate(data.exchangeRate);
    } catch (error) {
      toast.error('Failed to load NFT prices');
    } finally {
      setLoading(false);
    }
  };

  const purchaseNFT = async (nftType, currency) => {
    if (!user) {
      toast.error('Please connect your wallet first');
      return;
    }

    try {
      const price = currency === 'bkc' 
        ? nftPrices[nftType]?.BKCPrice 
        : nftPrices[nftType]?.TONPrice;

      if (currency === 'bkc' && user.balance < price) {
        toast.error('Insufficient BKC balance');
        return;
      }

      await apiService.purchaseNFT({
        nftType,
        currency,
        price,
        userId: user.id
      });

      toast.success(`Successfully purchased ${nftType} NFT!`);
      
      // Refresh user data
      if (onNavigate) {
        onNavigate('/profile');
      }
    } catch (error) {
      toast.error(error.message || 'Failed to purchase NFT');
    }
  };

  const getNFTIcon = (type) => {
    switch (type) {
      case 'bronze': return '🥉';
      case 'silver': return '🥈';
      case 'gold': return '🥇';
      case 'rocket': return '🚀';
      default: return '💎';
    }
  };

  const getNFTDescription = (type) => {
    switch (type) {
      case 'bronze': 
        return 'Bronze License - +500 daily tap limit, +10% energy regeneration';
      case 'silver': 
        return 'Silver License - +2,000 daily tap limit, +25% energy regeneration';
      case 'gold': 
        return 'Gold License - +10,000 daily tap limit, +50% energy regeneration, P2P trading access';
      case 'rocket': 
        return 'Instant Energy Boost - Refill 1000 energy immediately';
      default: 
        return 'Premium NFT with exclusive benefits';
    }
  };

  const getNFTColor = (type) => {
    switch (type) {
      case 'bronze': return 'from-amber-600 to-amber-800';
      case 'silver': return 'from-gray-400 to-gray-600';
      case 'gold': return 'from-yellow-400 to-yellow-600';
      case 'rocket': return 'from-purple-500 to-purple-700';
      default: return 'from-blue-500 to-blue-700';
    }
  };

  if (loading) {
    return (
      <div className="nft-shop-loading">
        <div className="loading-spinner" />
        <p>Loading NFT Shop...</p>
      </div>
    );
  }

  return (
    <div className="nft-shop">
      {/* Header */}
      <header className="shop-header">
        <div className="header-left">
          <h1>🛍️ NFT Marketplace</h1>
          <p>Get powerful licenses and boosts for your BKC mining</p>
        </div>
        <div className="header-right">
          <button onClick={loadNFTPrices} className="refresh-btn">
            <RefreshCw size={20} />
          </button>
        </div>
      </header>

      {/* Exchange Rate Display */}
      <div className="exchange-rate-banner">
        <div className="rate-info">
          <TrendingUp size={16} />
          <span>{exchangeRate}</span>
        </div>
        <p className="rate-description">
          Prices automatically adjust based on market rates. You always pay the fair USD value!
        </p>
      </div>

      {/* NFT Grid */}
      <div className="nft-grid">
        <AnimatePresence>
          {Object.entries(nftPrices).map(([nftType, priceData]) => (
            <motion.div
              key={nftType}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="nft-card"
            >
              {/* NFT Icon and Type */}
              <div className={`nft-header bg-gradient-to-r ${getNFTColor(nftType)}`}>
                <div className="nft-icon">{getNFTIcon(nftType)}</div>
                <div className="nft-title">
                  <h3>{nftType.charAt(0).toUpperCase() + nftType.slice(1)}</h3>
                  {nftType === 'rocket' && <span className="boost-badge">BOOST</span>}
                </div>
              </div>

              {/* NFT Description */}
              <div className="nft-description">
                <p>{getNFTDescription(nftType)}</p>
              </div>

              {/* Price Display */}
              <div className="nft-pricing">
                <div className="price-row">
                  <div className="price-item">
                    <span className="currency-label">BKC</span>
                    <span className="price-value">{priceData.BKCPrice.toLocaleString()}</span>
                  </div>
                  <div className="price-divider">or</div>
                  <div className="price-item">
                    <span className="currency-label">TON</span>
                    <span className="price-value">{priceData.TONPrice.toFixed(2)}</span>
                  </div>
                </div>
                <div className="usd-equivalent">
                  ≈ ${priceData.TargetUSD.toFixed(2)} USD
                </div>
              </div>

              {/* Purchase Buttons */}
              <div className="nft-actions">
                <button
                  onClick={() => purchaseNFT(nftType, 'bkc')}
                  className="purchase-btn bkc-btn"
                  disabled={!user || user.balance < priceData.BKCPrice}
                >
                  <ShoppingCart size={16} />
                  Buy for {priceData.BKCPrice.toLocaleString()} BKC
                </button>
                
                <button
                  onClick={() => purchaseNFT(nftType, 'ton')}
                  className="purchase-btn ton-btn"
                >
                  <ExternalLink size={16} />
                  Buy for {priceData.TONPrice.toFixed(2)} TON
                </button>
              </div>

              {/* Benefits List */}
              <div className="nft-benefits">
                <h4>
                  <Star size={14} />
                  Benefits
                </h4>
                <ul>
                  {nftType === 'bronze' && (
                    <>
                      <li>+500 daily tap limit</li>
                      <li>+10% energy regeneration</li>
                      <li>Permanent ownership</li>
                    </>
                  )}
                  {nftType === 'silver' && (
                    <>
                      <li>+2,000 daily tap limit</li>
                      <li>+25% energy regeneration</li>
                      <li>Priority support</li>
                    </>
                  )}
                  {nftType === 'gold' && (
                    <>
                      <li>+10,000 daily tap limit</li>
                      <li>+50% energy regeneration</li>
                      <li>P2P trading access</li>
                      <li>VIP status</li>
                    </>
                  )}
                  {nftType === 'rocket' && (
                    <>
                      <li>+1000 instant energy</li>
                      <li>No cooldown</li>
                      <li>Stackable with other NFTs</li>
                    </>
                  )}
                </ul>
              </div>

              {/* User Balance Status */}
              {user && (
                <div className="balance-status">
                  <div className="balance-item">
                    <span className="balance-label">Your BKC:</span>
                    <span className="balance-value">{user.balance.toLocaleString()}</span>
                  </div>
                  {priceData.BKCPrice > user.balance && (
                    <div className="insufficient-warning">
                      <Info size={14} />
                      <span>Insufficient BKC balance</span>
                    </div>
                  )}
                </div>
              )}
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      {/* Footer Info */}
      <div className="shop-footer">
        <div className="footer-info">
          <Shield size={16} />
          <div>
            <h4>100% Fair Pricing</h4>
            <p>
              All NFT prices are calculated based on real market rates. 
              You always pay the exact USD value regardless of BKC or TON fluctuations.
            </p>
          </div>
        </div>
        <div className="footer-info">
          <Zap size={16} />
          <div>
            <h4>Instant Delivery</h4>
            <p>
              All NFTs are delivered instantly to your inventory after purchase. 
              No waiting time or manual processing required.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NFTShop;
