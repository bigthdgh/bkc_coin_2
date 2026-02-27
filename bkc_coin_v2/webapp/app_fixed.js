// BKC Coin - Fixed Application with Minimal Economy
class BKCApp {
    constructor() {
        this.userID = null;
        this.userEconomy = null;
        this.isInitialized = false;
        this.tapCount = 0;
        this.lastTapTime = 0;
        this.energy = 1000;
        this.maxEnergy = 1000;
        this.balance = 0;
        this.level = 1;
        this.experience = 0;
        this.tapsToday = 0;
        this.maxTapsPerDay = 300;
        this.streakDays = 0;
        this.referralCount = 0;
        this.maxReferrals = 10;
        
        // Initialize minimal ads manager
        this.adsManager = null;
        
        this.init();
    }

    async init() {
        try {
            // Get user ID from localStorage or generate new
            this.userID = this.getUserID();
            
            // Initialize ads manager
            if (window.minimalAdsManager) {
                this.adsManager = window.minimalAdsManager;
            }
            
            // Load user economy
            await this.loadUserEconomy();
            
            // Setup UI
            this.setupUI();
            
            // Start energy regeneration
            this.startEnergyRegeneration();
            
            // Setup tap handlers
            this.setupTapHandlers();
            
            // Setup periodic updates
            this.startPeriodicUpdates();
            
            this.isInitialized = true;
            console.log('BKC App initialized successfully');
            
        } catch (error) {
            console.error('Failed to initialize BKC App:', error);
        }
    }

    getUserID() {
        let userID = localStorage.getItem('bkc_user_id');
        if (!userID) {
            userID = 'user_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            localStorage.setItem('bkc_user_id', userID);
        }
        return userID;
    }

    async loadUserEconomy() {
        try {
            const response = await fetch(`/api/v1/economy/user/${this.userID}`);
            const result = await response.json();
            
            if (result.success) {
                this.userEconomy = result.data;
                this.updateLocalEconomy();
            } else {
                console.error('Failed to load user economy:', result.error);
            }
        } catch (error) {
            console.error('Error loading user economy:', error);
        }
    }

    updateLocalEconomy() {
        if (!this.userEconomy) return;
        
        this.balance = this.userEconomy.balance || 0;
        this.energy = this.userEconomy.energy || 1000;
        this.maxEnergy = this.userEconomy.max_energy || 1000;
        this.level = this.userEconomy.level || 1;
        this.experience = this.userEconomy.experience || 0;
        this.tapsToday = this.userEconomy.taps_today || 0;
        this.streakDays = this.userEconomy.streak_days || 0;
        this.referralCount = this.userEconomy.referral_count || 0;
        
        this.updateUI();
    }

    setupUI() {
        // Create main UI elements
        this.createMainInterface();
        this.updateUI();
    }

    createMainInterface() {
        // Clear existing content
        document.body.innerHTML = '';
        
        // Create main container
        const mainContainer = document.createElement('div');
        mainContainer.className = 'bkc-main-container';
        mainContainer.innerHTML = `
            <div class="bkc-header">
                <div class="bkc-title">BKC Coin</div>
                <div class="bkc-level">Level ${this.level}</div>
            </div>
            
            <div class="bkc-stats">
                <div class="bkc-balance">
                    <div class="bkc-balance-label">Balance</div>
                    <div class="bkc-balance-value">${this.formatNumber(this.balance)} BKC</div>
                </div>
                
                <div class="bkc-energy">
                    <div class="bkc-energy-label">Energy</div>
                    <div class="bkc-energy-bar">
                        <div class="bkc-energy-fill" style="width: ${(this.energy / this.maxEnergy) * 100}%"></div>
                    </div>
                    <div class="bkc-energy-text">${this.energy}/${this.maxEnergy}</div>
                </div>
                
                <div class="bkc-taps">
                    <div class="bkc-taps-label">Taps Today</div>
                    <div class="bkc-taps-value">${this.tapsToday}/${this.maxTapsPerDay}</div>
                </div>
            </div>
            
            <div class="bkc-tap-area">
                <div class="bkc-tap-button">
                    <div class="bkc-tap-icon">⚡</div>
                    <div class="bkc-tap-text">TAP</div>
                </div>
                <div class="bkc-tap-effect"></div>
            </div>
            
            <div class="bkc-actions">
                <button class="bkc-btn bkc-btn-primary" id="dailyBonusBtn">
                    🎁 Daily Bonus
                </button>
                
                <button class="bkc-btn bkc-btn-secondary" id="coinsAdBtn">
                    💰 Get 30 BKC
                </button>
                
                <button class="bkc-btn bkc-btn-info" id="referralBtn">
                    👥 Referrals (${this.referralCount}/${this.maxReferrals})
                </button>
            </div>
            
            <div class="bkc-footer">
                <div class="bkc-streak">🔥 ${this.streakDays} day streak</div>
                <div class="bkc-experience">XP: ${this.experience}/${this.level * 100}</div>
            </div>
        `;
        
        // Add styles
        this.addStyles();
        
        document.body.appendChild(mainContainer);
        
        // Setup event listeners
        this.setupEventListeners();
    }

    addStyles() {
        const styles = `
            .bkc-main-container {
                max-width: 400px;
                margin: 0 auto;
                padding: 20px;
                font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                min-height: 100vh;
                color: white;
            }
            
            .bkc-header {
                text-align: center;
                margin-bottom: 30px;
            }
            
            .bkc-title {
                font-size: 2.5em;
                font-weight: bold;
                margin-bottom: 10px;
            }
            
            .bkc-level {
                font-size: 1.2em;
                opacity: 0.9;
            }
            
            .bkc-stats {
                display: grid;
                gap: 15px;
                margin-bottom: 30px;
            }
            
            .bkc-balance, .bkc-energy, .bkc-taps {
                background: rgba(255, 255, 255, 0.1);
                padding: 15px;
                border-radius: 15px;
                backdrop-filter: blur(10px);
            }
            
            .bkc-balance-label, .bkc-energy-label, .bkc-taps-label {
                font-size: 0.9em;
                opacity: 0.8;
                margin-bottom: 5px;
            }
            
            .bkc-balance-value {
                font-size: 1.8em;
                font-weight: bold;
            }
            
            .bkc-energy-bar {
                width: 100%;
                height: 8px;
                background: rgba(255, 255, 255, 0.2);
                border-radius: 4px;
                overflow: hidden;
                margin: 10px 0;
            }
            
            .bkc-energy-fill {
                height: 100%;
                background: linear-gradient(90deg, #00ff88, #00cc66);
                transition: width 0.3s ease;
            }
            
            .bkc-energy-text {
                font-size: 0.9em;
                opacity: 0.9;
            }
            
            .bkc-taps-value {
                font-size: 1.2em;
                font-weight: bold;
            }
            
            .bkc-tap-area {
                position: relative;
                display: flex;
                justify-content: center;
                align-items: center;
                height: 200px;
                margin-bottom: 30px;
            }
            
            .bkc-tap-button {
                width: 150px;
                height: 150px;
                border-radius: 50%;
                background: linear-gradient(135deg, #ff6b6b, #ff8e53);
                display: flex;
                flex-direction: column;
                justify-content: center;
                align-items: center;
                cursor: pointer;
                transition: transform 0.1s ease;
                box-shadow: 0 10px 30px rgba(255, 107, 107, 0.3);
                user-select: none;
            }
            
            .bkc-tap-button:active {
                transform: scale(0.95);
            }
            
            .bkc-tap-icon {
                font-size: 3em;
                margin-bottom: 5px;
            }
            
            .bkc-tap-text {
                font-size: 1.2em;
                font-weight: bold;
            }
            
            .bkc-tap-effect {
                position: absolute;
                pointer-events: none;
                opacity: 0;
            }
            
            .bkc-actions {
                display: grid;
                gap: 15px;
                margin-bottom: 30px;
            }
            
            .bkc-btn {
                padding: 15px;
                border: none;
                border-radius: 12px;
                font-size: 1em;
                font-weight: bold;
                cursor: pointer;
                transition: all 0.3s ease;
                text-align: center;
            }
            
            .bkc-btn-primary {
                background: linear-gradient(135deg, #667eea, #764ba2);
                color: white;
            }
            
            .bkc-btn-secondary {
                background: linear-gradient(135deg, #feca57, #ff6b6b);
                color: white;
            }
            
            .bkc-btn-info {
                background: rgba(255, 255, 255, 0.1);
                color: white;
                backdrop-filter: blur(10px);
            }
            
            .bkc-btn:hover {
                transform: translateY(-2px);
                box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
            }
            
            .bkc-footer {
                display: flex;
                justify-content: space-between;
                align-items: center;
                font-size: 0.9em;
                opacity: 0.8;
            }
            
            .bkc-notification {
                position: fixed;
                top: 20px;
                left: 50%;
                transform: translateX(-50%);
                background: rgba(0, 0, 0, 0.8);
                color: white;
                padding: 15px 25px;
                border-radius: 10px;
                z-index: 10000;
                animation: slideDown 0.3s ease;
            }
            
            @keyframes slideDown {
                from {
                    transform: translate(-50%, -100%);
                    opacity: 0;
                }
                to {
                    transform: translate(-50%, 0);
                    opacity: 1;
                }
            }
            
            @keyframes fadeOut {
                from {
                    opacity: 1;
                }
                to {
                    opacity: 0;
                }
            }
        `;
        
        const styleSheet = document.createElement('style');
        styleSheet.textContent = styles;
        document.head.appendChild(styleSheet);
    }

    setupEventListeners() {
        // Tap button
        const tapButton = document.querySelector('.bkc-tap-button');
        if (tapButton) {
            tapButton.addEventListener('click', () => this.handleTap());
        }
        
        // Daily bonus button
        const dailyBonusBtn = document.getElementById('dailyBonusBtn');
        if (dailyBonusBtn) {
            dailyBonusBtn.addEventListener('click', () => this.claimDailyBonus());
        }
        
        // Coins ad button
        const coinsAdBtn = document.getElementById('coinsAdBtn');
        if (coinsAdBtn) {
            coinsAdBtn.addEventListener('click', () => this.showCoinsAd());
        }
        
        // Referral button
        const referralBtn = document.getElementById('referralBtn');
        if (referralBtn) {
            referralBtn.addEventListener('click', () => this.showReferralInfo());
        }
    }

    setupTapHandlers() {
        // Prevent context menu on tap button
        document.addEventListener('contextmenu', (e) => {
            if (e.target.closest('.bkc-tap-button')) {
                e.preventDefault();
            }
        });
    }

    async handleTap() {
        // Check limits
        if (this.tapsToday >= this.maxTapsPerDay) {
            this.showNotification('Daily tap limit reached! Come back tomorrow.');
            return;
        }
        
        if (this.energy <= 0) {
            this.showNotification('No energy! Wait for restoration or watch ad.');
            return;
        }
        
        // Prevent spam tapping
        const now = Date.now();
        if (now - this.lastTapTime < 100) {
            return;
        }
        
        this.lastTapTime = now;
        
        // Process tap
        try {
            const response = await fetch('/api/v1/economy/tap', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    user_id: this.userID
                })
            });
            
            const result = await response.json();
            
            if (result.success) {
                this.handleTapResult(result.data);
                this.showTapEffect();
            } else {
                this.showNotification('Tap failed: ' + result.error);
            }
        } catch (error) {
            console.error('Tap error:', error);
            this.showNotification('Tap failed. Please try again.');
        }
    }

    handleTapResult(tapResult) {
        if (tapResult.success) {
            this.balance = tapResult.new_balance;
            this.energy = tapResult.new_energy;
            this.tapsToday = this.maxTapsPerDay - tapResult.taps_remaining;
            
            if (tapResult.level_up) {
                this.level++;
                this.showNotification('🎉 Level Up! You are now level ' + this.level);
            }
            
            this.updateUI();
        }
    }

    showTapEffect() {
        const effect = document.querySelector('.bkc-tap-effect');
        if (!effect) return;
        
        effect.innerHTML = '+1 BKC';
        effect.style.opacity = '1';
        effect.style.transform = 'translateY(0)';
        
        setTimeout(() => {
            effect.style.opacity = '0';
            effect.style.transform = 'translateY(-50px)';
        }, 1000);
    }

    async claimDailyBonus() {
        try {
            const response = await fetch(`/api/v1/economy/daily-bonus/${this.userID}`);
            const result = await response.json();
            
            if (result.success) {
                const bonus = result.data;
                this.balance += bonus.bonus;
                this.streakDays = bonus.streak;
                this.updateUI();
                this.showNotification(`🎁 Daily bonus: +${bonus.bonus} BKC! 🔥 ${bonus.streak} day streak!`);
            } else {
                this.showNotification('Daily bonus already claimed today!');
            }
        } catch (error) {
            console.error('Daily bonus error:', error);
            this.showNotification('Failed to claim daily bonus');
        }
    }

    showCoinsAd() {
        if (this.adsManager) {
            this.adsManager.showCoinsAd();
        } else {
            this.showNotification('Ads manager not available');
        }
    }

    showReferralInfo() {
        const referralLink = `https://bkc-coin.com/ref/${this.userID}`;
        const message = `Refer friends and earn 10 BKC for each referral!\n\nYour link: ${referralLink}\n\nReferrals: ${this.referralCount}/${this.maxReferrals}`;
        
        this.showNotification(message);
        
        // Copy to clipboard
        navigator.clipboard.writeText(referralLink).then(() => {
            console.log('Referral link copied to clipboard');
        }).catch(err => {
            console.error('Failed to copy referral link:', err);
        });
    }

    startEnergyRegeneration() {
        setInterval(() => {
            if (this.energy < this.maxEnergy) {
                this.energy = Math.min(this.energy + 1, this.maxEnergy);
                this.updateUI();
            }
        }, 1000); // 1 energy per second
    }

    startPeriodicUpdates() {
        // Update economy data every 30 seconds
        setInterval(() => {
            this.loadUserEconomy();
        }, 30000);
    }

    updateUI() {
        // Update balance
        const balanceElement = document.querySelector('.bkc-balance-value');
        if (balanceElement) {
            balanceElement.textContent = this.formatNumber(this.balance) + ' BKC';
        }
        
        // Update energy
        const energyFill = document.querySelector('.bkc-energy-fill');
        const energyText = document.querySelector('.bkc-energy-text');
        if (energyFill) {
            energyFill.style.width = `${(this.energy / this.maxEnergy) * 100}%`;
        }
        if (energyText) {
            energyText.textContent = `${this.energy}/${this.maxEnergy}`;
        }
        
        // Update taps
        const tapsValue = document.querySelector('.bkc-taps-value');
        if (tapsValue) {
            tapsValue.textContent = `${this.tapsToday}/${this.maxTapsPerDay}`;
        }
        
        // Update level
        const levelElement = document.querySelector('.bkc-level');
        if (levelElement) {
            levelElement.textContent = `Level ${this.level}`;
        }
        
        // Update streak
        const streakElement = document.querySelector('.bkc-streak');
        if (streakElement) {
            streakElement.textContent = `🔥 ${this.streakDays} day streak`;
        }
        
        // Update experience
        const expElement = document.querySelector('.bkc-experience');
        if (expElement) {
            expElement.textContent = `XP: ${this.experience}/${this.level * 100}`;
        }
        
        // Update referral button
        const referralBtn = document.getElementById('referralBtn');
        if (referralBtn) {
            referralBtn.textContent = `👥 Referrals (${this.referralCount}/${this.maxReferrals})`;
        }
    }

    showNotification(message) {
        const notification = document.createElement('div');
        notification.className = 'bkc-notification';
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.style.animation = 'fadeOut 0.3s ease';
            setTimeout(() => {
                notification.remove();
            }, 300);
        }, 3000);
    }

    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toFixed(0);
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.bkcApp = new BKCApp();
});

// Export for external access
window.BKCApp = {
    showCoinsAd: () => window.bkcApp?.showCoinsAd(),
    checkEnergy: () => window.bkcApp?.adsManager?.checkEnergy(),
    getStats: () => window.bkcApp?.userEconomy
};
