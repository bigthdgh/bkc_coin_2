import React, { useState, useEffect, useCallback, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Coins, Zap, Trophy, Users, TrendingUp, Menu, X } from 'lucide-react';
import toast from 'react-hot-toast';

const Game = ({ user, onNavigate, apiService }) => {
  // Game state
  const [balance, setBalance] = useState(0);
  const [energy, setEnergy] = useState(1000);
  const [maxEnergy, setMaxEnergy] = useState(1000);
  const [level, setLevel] = useState(1);
  const [tapsCount, setTapsCount] = useState(0);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  
  // Animation state
  const [taps, setTaps] = useState([]);
  const [isTapping, setIsTapping] = useState(false);
  
  // Refs
  const tapTimeoutRef = useRef(null);
  const lastTapTime = useRef(Date.now());

  // Load initial data
  useEffect(() => {
    loadGameData();
    
    // Set up periodic updates
    const interval = setInterval(loadGameData, 10000); // Update every 10 seconds
    return () => clearInterval(interval);
  }, [user.id]);

  const loadGameData = async () => {
    try {
      const [balanceData, energyData, statsData] = await Promise.all([
        apiService.getUserBalance(user.id),
        apiService.getUserEnergy(user.id),
        apiService.getUserStats(user.id)
      ]);

      setBalance(balanceData.balance || 0);
      setEnergy(energyData.energy || 1000);
      setMaxEnergy(energyData.maxEnergy || 1000);
      setLevel(statsData.level || 1);
      setTapsCount(statsData.total_taps || 0);
    } catch (error) {
      console.error('Failed to load game data:', error);
      toast.error('Failed to load game data');
    }
  };

  const handleTap = useCallback(async (event) => {
    if (energy <= 0) {
      toast.error('No energy left! Wait for regeneration.');
      return;
    }

    const now = Date.now();
    const timeSinceLastTap = now - lastTapTime.current;
    
    // Prevent too rapid tapping (anti-cheat)
    if (timeSinceLastTap < 50) {
      return;
    }

    lastTapTime.current = now;
    setIsTapping(true);

    try {
      const response = await apiService.processTap(user.id, 1);
      
      if (response.success) {
        // Update local state immediately for responsiveness
        setBalance(response.newBalance);
        setEnergy(response.newEnergy);
        setTapsCount(prev => prev + 1);

        // Create tap animation
        const rect = event.currentTarget.getBoundingClientRect();
        const tap = {
          id: Date.now() + Math.random(),
          x: event.clientX - rect.left,
          y: event.clientY - rect.top,
          value: 1
        };
        
        setTaps(prev => [...prev, tap]);
        
        // Remove tap animation after delay
        setTimeout(() => {
          setTaps(prev => prev.filter(t => t.id !== tap.id));
        }, 1000);

        // Clear tapping state
        if (tapTimeoutRef.current) {
          clearTimeout(tapTimeoutRef.current);
        }
        tapTimeoutRef.current = setTimeout(() => {
          setIsTapping(false);
        }, 100);

      } else {
        toast.error(response.message || 'Tap failed');
      }
    } catch (error) {
      console.error('Tap failed:', error);
      toast.error('Network error. Please try again.');
      setIsTapping(false);
    }
  }, [user.id, energy, apiService]);

  const claimDailyReward = async () => {
    try {
      const response = await apiService.claimDailyReward(user.id);
      if (response.success) {
        toast.success(`Daily reward claimed: ${response.reward} BKC!`);
        setBalance(prev => prev + response.reward);
      } else {
        toast.error(response.message || 'Failed to claim reward');
      }
    } catch (error) {
      console.error('Failed to claim daily reward:', error);
      toast.error('Failed to claim daily reward');
    }
  };

  const formatNumber = (num) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const energyPercentage = (energy / maxEnergy) * 100;

  return (
    <div className="game-container">
      {/* Header */}
      <header className="game-header">
        <div className="header-left">
          <button 
            className="menu-button"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
          <div className="user-info">
            <span className="username">{user.first_name}</span>
            <span className="level">Level {level}</span>
          </div>
        </div>
        
        <div className="header-right">
          <div className="balance-display">
            <Coins size={20} className="coin-icon" />
            <span className="balance-amount">{formatNumber(balance)}</span>
          </div>
        </div>
      </header>

      {/* Side Menu */}
      <AnimatePresence>
        {isMenuOpen && (
          <motion.div
            initial={{ x: -300 }}
            animate={{ x: 0 }}
            exit={{ x: -300 }}
            className="side-menu"
          >
            <nav className="menu-nav">
              <button 
                className="menu-item"
                onClick={() => {
                  onNavigate('/profile');
                  setIsMenuOpen(false);
                }}
              >
                <Users size={20} />
                <span>Profile</span>
              </button>
              
              <button 
                className="menu-item"
                onClick={() => {
                  onNavigate('/leaderboard');
                  setIsMenuOpen(false);
                }}
              >
                <Trophy size={20} />
                <span>Leaderboard</span>
              </button>
              
              <button 
                className="menu-item"
                onClick={() => {
                  onNavigate('/p2p');
                  setIsMenuOpen(false);
                }}
              >
                <TrendingUp size={20} />
                <span>P2P Market</span>
              </button>
              
              <button 
                className="menu-item"
                onClick={() => {
                  onNavigate('/bank');
                  setIsMenuOpen(false);
                }}
              >
                <Coins size={20} />
                <span>Bank</span>
              </button>

              {user.isAdmin && (
                <button 
                  className="menu-item admin-item"
                  onClick={() => {
                    onNavigate('/admin');
                    setIsMenuOpen(false);
                  }}
                >
                  <Users size={20} />
                  <span>Admin Panel</span>
                </button>
              )}
            </nav>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Game Area */}
      <main className="game-main">
        {/* Energy Bar */}
        <div className="energy-container">
          <div className="energy-header">
            <Zap size={16} className="energy-icon" />
            <span className="energy-text">Energy</span>
            <span className="energy-count">{energy}/{maxEnergy}</span>
          </div>
          <div className="energy-bar">
            <motion.div 
              className="energy-fill"
              style={{ width: `${energyPercentage}%` }}
              animate={{ width: `${energyPercentage}%` }}
              transition={{ duration: 0.3 }}
            />
          </div>
        </div>

        {/* Tap Area */}
        <div className="tap-container">
          <motion.button
            className="tap-button"
            onClick={handleTap}
            disabled={energy <= 0}
            whileTap={{ scale: 0.95 }}
            animate={{
              scale: isTapping ? 1.05 : 1,
            }}
            transition={{
              type: "spring",
              stiffness: 400,
              damping: 17
            }}
          >
            <div className="tap-button-content">
              <Coins size={60} className="tap-coin" />
              <span className="tap-text">TAP!</span>
            </div>
          </motion.button>

          {/* Tap Animations */}
          <div className="tap-animations">
            <AnimatePresence>
              {taps.map(tap => (
                <motion.div
                  key={tap.id}
                  className="tap-animation"
                  initial={{ 
                    x: tap.x, 
                    y: tap.y, 
                    opacity: 1, 
                    scale: 1 
                  }}
                  animate={{ 
                    y: tap.y - 50, 
                    opacity: 0, 
                    scale: 1.5 
                  }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 1 }}
                >
                  +{tap.value}
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        </div>

        {/* Stats */}
        <div className="game-stats">
          <div className="stat-item">
            <span className="stat-label">Total Taps</span>
            <span className="stat-value">{formatNumber(tapsCount)}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Daily Reward</span>
            <button 
              className="reward-button"
              onClick={claimDailyReward}
            >
              Claim
            </button>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="game-footer">
        <div className="footer-stats">
          <div className="footer-stat">
            <span className="footer-label">BKC Balance</span>
            <span className="footer-value">{formatNumber(balance)}</span>
          </div>
          <div className="footer-stat">
            <span className="footer-label">Energy</span>
            <span className="footer-value">{energy}</span>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default Game;
