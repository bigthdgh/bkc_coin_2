// BKC Coin - Secure Ads Manager with WebSocket Integration
class SecureAdsManager {
    constructor() {
        this.isInitialized = false;
        this.currentAd = null;
        this.websocket = null;
        this.userStats = null;
        this.availableRewards = [];
        this.adSessions = new Map(); // Track active ad sessions
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
    }

    async initialize() {
        try {
            await this.loadUserStats();
            await this.loadAvailableRewards();
            this.setupEventListeners();
            this.connectWebSocket();
            this.isInitialized = true;
            console.log('Secure Ads Manager initialized successfully');
        } catch (error) {
            console.error('Failed to initialize Secure Ads Manager:', error);
        }
    }

    connectWebSocket() {
        const userID = this.getCurrentUserID();
        if (!userID) {
            console.error('User ID not found');
            return;
        }

        const wsURL = `ws://localhost:8081/ws?user_id=${userID}`;
        
        try {
            this.websocket = new WebSocket(wsURL);
            
            this.websocket.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
            };

            this.websocket.onmessage = (event) => {
                this.handleWebSocketMessage(event);
            };

            this.websocket.onclose = () => {
                console.log('WebSocket disconnected');
                this.attemptReconnect();
            };

            this.websocket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('Failed to connect WebSocket:', error);
            this.attemptReconnect();
        }
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            
            setTimeout(() => {
                this.connectWebSocket();
            }, 1000 * this.reconnectAttempts); // Exponential backoff
        } else {
            console.error('Max reconnection attempts reached');
        }
    }

    handleWebSocketMessage(event) {
        try {
            const data = JSON.parse(event.data);
            
            switch (data.type) {
                case 'ad_completed':
                    this.handleAdCompleted(data);
                    break;
                case 'ad_expired':
                    this.handleAdExpired(data);
                    break;
                default:
                    console.log('Unknown WebSocket message type:', data.type);
            }
        } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
        }
    }

    handleAdCompleted(data) {
        console.log('Ad completed:', data);
        
        // Remove loading state
        this.hideAdLoading();
        
        // Show reward notification
        this.showRewardNotification(data);
        
        // Update user stats
        this.loadUserStats();
        
        // Remove from active sessions
        this.adSessions.delete(data.trigger);
        
        // Update UI
        this.updateUIAfterReward(data);
    }

    handleAdExpired(data) {
        console.log('Ad expired:', data);
        
        // Remove loading state
        this.hideAdLoading();
        
        // Show expiration notification
        this.showExpirationNotification(data);
        
        // Remove from active sessions
        this.adSessions.delete(data.trigger);
        
        // Allow user to try again
        this.enableAdTrigger(data.trigger);
    }

    async loadUserStats() {
        try {
            const response = await fetch('/api/v1/ads/stats');
            this.userStats = await response.json();
        } catch (error) {
            console.error('Failed to load user stats:', error);
            this.userStats = {
                ads_today: 0,
                total_ads: 0,
                completed_ads: 0,
                total_rewards: 0
            };
        }
    }

    async loadAvailableRewards() {
        try {
            const response = await fetch('/api/v1/ads/rewards');
            this.availableRewards = await response.json();
        } catch (error) {
            console.error('Failed to load available rewards:', error);
            this.availableRewards = [];
        }
    }

    setupEventListeners() {
        // Listen for game events
        document.addEventListener('energyDepleted', () => this.handleEnergyDepleted());
        document.addEventListener('achievementUnlocked', () => this.handleAchievement());
        document.addEventListener('levelUp', () => this.handleLevelUp());
        document.addEventListener('dailyBonusClaimed', () => this.handleDailyBonus());
        document.addEventListener('tournamentEntry', () => this.handleTournamentEntry());
        document.addEventListener('nftPurchase', () => this.handleNFTPurchase());
    }

    // Smart trigger handlers
    async handleEnergyDepleted() {
        if (!this.shouldShowAd('energy_depleted')) return;

        const energyAd = this.availableRewards.find(r => r.trigger === 'energy_depleted');
        if (!energyAd) return;

        this.showEnergyAdOption(energyAd);
    }

    showEnergyAdOption(adConfig) {
        // Check if energy is actually depleted
        const currentEnergy = this.getCurrentEnergy();
        if (currentEnergy > 100) {
            return; // Energy not depleted enough
        }

        // Create energy restoration button
        const energyButton = document.createElement('button');
        energyButton.className = 'energy-restore-btn';
        energyButton.innerHTML = `
            <div class="btn-content">
                <span class="icon">⚡</span>
                <span class="text">Восстановить энергию</span>
                <span class="subtext">Посмотреть рекламу</span>
            </div>
        `;

        // Add styles
        energyButton.style.cssText = `
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: linear-gradient(135deg, #3cd6c6, #22d3ee);
            color: white;
            border: none;
            border-radius: 12px;
            padding: 12px 20px;
            cursor: pointer;
            box-shadow: 0 4px 15px rgba(60, 214, 198, 0.3);
            z-index: 1000;
            font-family: 'Inter', sans-serif;
            font-weight: 500;
            transition: all 0.3s ease;
            animation: slideInRight 0.5s ease;
        `;

        // Add hover effects
        energyButton.addEventListener('mouseenter', () => {
            energyButton.style.transform = 'translateY(-2px)';
            energyButton.style.boxShadow = '0 6px 20px rgba(60, 214, 198, 0.4)';
        });

        energyButton.addEventListener('mouseleave', () => {
            energyButton.style.transform = 'translateY(0)';
            energyButton.style.boxShadow = '0 4px 15px rgba(60, 214, 198, 0.3)';
        });

        // Handle click
        energyButton.addEventListener('click', () => {
            this.startSecureAdSession(adConfig);
            energyButton.remove();
        });

        // Auto-hide after 10 seconds
        setTimeout(() => {
            if (energyButton.parentNode) {
                energyButton.style.animation = 'slideOutRight 0.5s ease';
                setTimeout(() => energyButton.remove(), 500);
            }
        }, 10000);

        document.body.appendChild(energyButton);
    }

    async handleAchievement() {
        if (!this.shouldShowAd('achievement')) return;

        const achievementAd = this.availableRewards.find(r => r.trigger === 'achievement');
        if (!achievementAd) return;

        this.showAchievementBonusOption(achievementAd);
    }

    showAchievementBonusOption(adConfig) {
        const bonusOption = document.createElement('div');
        bonusOption.className = 'achievement-bonus-option';
        bonusOption.innerHTML = `
            <div class="bonus-content">
                <span class="icon">🎁</span>
                <span class="text">Увеличить награду за достижение</span>
                <span class="reward">+${adConfig.reward} BKC</span>
                <button class="watch-ad-btn">Посмотреть рекламу</button>
            </div>
        `;

        bonusOption.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: linear-gradient(135deg, #0891b2, #06b6d4);
            color: white;
            border-radius: 16px;
            padding: 20px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            z-index: 10000;
            text-align: center;
            animation: achievementPop 0.5s ease;
        `;

        const watchBtn = bonusOption.querySelector('.watch-ad-btn');
        watchBtn.addEventListener('click', () => {
            this.startSecureAdSession(adConfig);
            bonusOption.remove();
        });

        // Auto-hide after 8 seconds
        setTimeout(() => {
            if (bonusOption.parentNode) {
                bonusOption.style.animation = 'fadeOut 0.5s ease';
                setTimeout(() => bonusOption.remove(), 500);
            }
        }, 8000);

        document.body.appendChild(bonusOption);
    }

    async handleDailyBonus() {
        if (!this.shouldShowAd('daily_bonus')) return;

        const dailyAd = this.availableRewards.find(r => r.trigger === 'daily_bonus');
        if (!dailyAd) return;

        this.showDailyBonusOption(dailyAd);
    }

    showDailyBonusOption(adConfig) {
        const bonusOption = document.createElement('div');
        bonusOption.className = 'daily-bonus-option';
        bonusOption.innerHTML = `
            <div class="bonus-content">
                <span class="icon">💰</span>
                <span class="text">Увеличить ежедневный бонус</span>
                <span class="reward">+${adConfig.reward} BKC</span>
                <button class="watch-ad-btn">Посмотреть рекламу</button>
            </div>
        `;

        bonusOption.style.cssText = `
            position: fixed;
            bottom: 80px;
            right: 20px;
            background: linear-gradient(135deg, #feca57, #ff6b6b);
            color: white;
            border-radius: 12px;
            padding: 15px;
            box-shadow: 0 4px 15px rgba(254, 202, 87, 0.3);
            z-index: 1000;
            animation: slideInRight 0.5s ease;
        `;

        const watchBtn = bonusOption.querySelector('.watch-ad-btn');
        watchBtn.addEventListener('click', () => {
            this.startSecureAdSession(adConfig);
            bonusOption.remove();
        });

        // Auto-hide after 12 seconds
        setTimeout(() => {
            if (bonusOption.parentNode) {
                bonusOption.style.animation = 'slideOutRight 0.5s ease';
                setTimeout(() => bonusOption.remove(), 500);
            }
        }, 12000);

        document.body.appendChild(bonusOption);
    }

    async handleLevelUp() {
        // Show interstitial ad (non-rewarded)
        const levelUpAd = this.availableRewards.find(r => r.trigger === 'level_up');
        if (!levelUpAd) return;

        this.startSecureAdSession(levelUpAd);
    }

    async handleTournamentEntry() {
        if (!this.shouldShowAd('tournament_entry')) return;

        const tournamentAd = this.availableRewards.find(r => r.trigger === 'tournament_entry');
        if (!tournamentAd) return;

        this.showTournamentEntryOption(tournamentAd);
    }

    showTournamentEntryOption(adConfig) {
        const entryOption = document.createElement('div');
        entryOption.className = 'tournament-entry-option';
        entryOption.innerHTML = `
            <div class="entry-content">
                <span class="icon">🏆</span>
                <span class="text">Бесплатный вход в турнир</span>
                <button class="watch-ad-btn">Посмотреть рекламу</button>
            </div>
        `;

        entryOption.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: linear-gradient(135deg, #ff6b6b, #feca57);
            color: white;
            border-radius: 12px;
            padding: 12px 16px;
            box-shadow: 0 4px 15px rgba(255, 107, 107, 0.3);
            z-index: 1000;
            animation: slideInRight 0.5s ease;
        `;

        const watchBtn = entryOption.querySelector('.watch-ad-btn');
        watchBtn.addEventListener('click', () => {
            this.startSecureAdSession(adConfig);
            entryOption.remove();
        });

        // Auto-hide after 15 seconds
        setTimeout(() => {
            if (entryOption.parentNode) {
                entryOption.style.animation = 'slideOutRight 0.5s ease';
                setTimeout(() => entryOption.remove(), 500);
            }
        }, 15000);

        document.body.appendChild(entryOption);
    }

    async handleNFTPurchase() {
        if (!this.shouldShowAd('nft_purchase')) return;

        const nftAd = this.availableRewards.find(r => r.trigger === 'nft_purchase');
        if (!nftAd) return;

        this.showNFTDiscountOption(nftAd);
    }

    showNFTDiscountOption(adConfig) {
        const discountOption = document.createElement('div');
        discountOption.className = 'nft-discount-option';
        discountOption.innerHTML = `
            <div class="discount-content">
                <span class="icon">🎨</span>
                <span class="text">Скидка на NFT</span>
                <span class="discount">+${adConfig.reward} BKC</span>
                <button class="watch-ad-btn">Посмотреть рекламу</button>
            </div>
        `;

        discountOption.style.cssText = `
            position: fixed;
            bottom: 20px;
            left: 20px;
            background: linear-gradient(135deg, #8b5cf6, #ec4899);
            color: white;
            border-radius: 12px;
            padding: 12px 16px;
            box-shadow: 0 4px 15px rgba(139, 92, 246, 0.3);
            z-index: 1000;
            animation: slideInLeft 0.5s ease;
        `;

        const watchBtn = discountOption.querySelector('.watch-ad-btn');
        watchBtn.addEventListener('click', () => {
            this.startSecureAdSession(adConfig);
            discountOption.remove();
        });

        // Auto-hide after 10 seconds
        setTimeout(() => {
            if (discountOption.parentNode) {
                discountOption.style.animation = 'slideOutLeft 0.5s ease';
                setTimeout(() => discountOption.remove(), 500);
            }
        }, 10000);

        document.body.appendChild(discountOption);
    }

    shouldShowAd(trigger) {
        // Check if user has reached daily limits
        if (this.userStats.ads_today >= 24) return false; // Max 24 ads per day
        
        // Check if user recently watched an ad
        if (this.adSessions.has(trigger)) return false; // Already watching

        return true;
    }

    async startSecureAdSession(adConfig) {
        try {
            // Start ad session on backend
            const response = await fetch('/api/v1/ads/session/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    trigger: adConfig.trigger
                })
            });

            if (!response.ok) {
                const error = await response.json();
                this.showAdErrorNotification(error.message || 'Не удалось начать сессию');
                return;
            }

            const sessionData = await response.json();
            
            // Track active session
            this.adSessions.set(adConfig.trigger, sessionData);
            
            // Show loading state
            this.showAdLoading(adConfig);
            
            // Load Adsgram and show ad
            await this.loadAdsgramAndShowAd(adConfig);

        } catch (error) {
            console.error('Failed to start ad session:', error);
            this.showAdErrorNotification('Не удалось начать сессию рекламы');
        }
    }

    showAdLoading(adConfig) {
        const loadingOverlay = document.createElement('div');
        loadingOverlay.className = 'ad-loading-overlay';
        loadingOverlay.innerHTML = `
            <div class="loading-content">
                <div class="spinner"></div>
                <div class="loading-text">Проверяем просмотр...</div>
                <div class="loading-subtext">Пожалуйста, подождите</div>
            </div>
        `;

        loadingOverlay.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.8);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 10000;
            animation: fadeIn 0.3s ease;
        `;

        document.body.appendChild(loadingOverlay);
    }

    hideAdLoading() {
        const loadingOverlay = document.querySelector('.ad-loading-overlay');
        if (loadingOverlay) {
            loadingOverlay.style.animation = 'fadeOut 0.3s ease';
            setTimeout(() => loadingOverlay.remove(), 300);
        }
    }

    async loadAdsgramAndShowAd(adConfig) {
        // Load Adsgram SDK if not already loaded
        if (!window.Adsgram) {
            await this.loadAdsgramSDK();
        }

        try {
            if (adConfig.type === 'rewarded') {
                await this.showAdsgramRewardedAd(adConfig.placement_id);
            } else {
                await this.showAdsgramInterstitial(adConfig.placement_id);
            }
        } catch (error) {
            console.error('Failed to show ad:', error);
            this.hideAdLoading();
            this.showAdErrorNotification('Не удалось загрузить рекламу');
        }
    }

    loadAdsgramSDK() {
        return new Promise((resolve, reject) => {
            if (window.Adsgram) {
                resolve();
                return;
            }

            const script = document.createElement('script');
            script.src = 'https://static.adsgram.com/js/adsgram.js';
            script.async = true;
            script.onload = () => resolve();
            script.onerror = () => reject(new Error('Failed to load Adsgram SDK'));
            document.head.appendChild(script);
        });
    }

    showAdsgramRewardedAd(placementId) {
        return new Promise((resolve, reject) => {
            if (!window.Adsgram) {
                reject(new Error('Adsgram SDK not loaded'));
                return;
            }

            window.Adsgram.showRewardedAd({
                placementId: placementId,
                onReward: (reward) => {
                    console.log('Ad completed, reward:', reward);
                    // Don't grant reward here - wait for webhook
                    resolve(reward);
                },
                onError: (error) => {
                    console.error('Ad error:', error);
                    reject(error);
                },
                onAdClosed: () => {
                    console.log('Ad closed by user');
                    reject(new Error('Ad closed by user'));
                }
            });
        });
    }

    showAdsgramInterstitial(placementId) {
        return new Promise((resolve, reject) => {
            if (!window.Adsgram) {
                reject(new Error('Adsgram SDK not loaded'));
                return;
            }

            window.Adsgram.showInterstitial({
                placementId: placementId,
                onAdClosed: () => {
                    console.log('Interstitial ad closed');
                    resolve();
                },
                onError: (error) => {
                    console.error('Interstitial ad error:', error);
                    reject(error);
                }
            });
        });
    }

    showRewardNotification(data) {
        const notification = document.createElement('div');
        notification.className = 'ad-reward-notification';
        
        let rewardText = '';
        switch (data.rewardType) {
            case 'energy':
                rewardText = `⚡ Энергия восстановлена!`;
                break;
            case 'coins':
                rewardText = `💰 +${data.reward} BKC`;
                break;
            default:
                rewardText = `🎁 Награда получена!`;
        }

        notification.innerHTML = `
            <div class="notification-content">
                <span class="reward-text">${rewardText}</span>
                <span class="thank-you">${data.data.message}</span>
            </div>
        `;

        notification.style.cssText = `
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: linear-gradient(135deg, #10b981, #059669);
            color: white;
            border-radius: 12px;
            padding: 16px 24px;
            box-shadow: 0 4px 15px rgba(16, 185, 129, 0.3);
            z-index: 10000;
            animation: slideInTop 0.5s ease;
        `;

        document.body.appendChild(notification);

        // Auto-hide after 3 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOutTop 0.5s ease';
            setTimeout(() => notification.remove(), 500);
        }, 3000);
    }

    showExpirationNotification(data) {
        const notification = document.createElement('div');
        notification.className = 'ad-expiration-notification';
        notification.innerHTML = `
            <div class="notification-content">
                <span class="error-text">${data.message}</span>
                <span class="retry-text">Попробуйте еще раз</span>
            </div>
        `;

        notification.style.cssText = `
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: linear-gradient(135deg, #f59e0b, #d97706);
            color: white;
            border-radius: 12px;
            padding: 16px 24px;
            box-shadow: 0 4px 15px rgba(245, 158, 11, 0.3);
            z-index: 10000;
            animation: slideInTop 0.5s ease;
        `;

        document.body.appendChild(notification);

        // Auto-hide after 3 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOutTop 0.5s ease';
            setTimeout(() => notification.remove(), 500);
        }, 3000);
    }

    showAdErrorNotification(message) {
        const notification = document.createElement('div');
        notification.className = 'ad-error-notification';
        notification.innerHTML = `
            <div class="notification-content">
                <span class="error-text">${message}</span>
                <span class="retry-text">Попробуйте позже</span>
            </div>
        `;

        notification.style.cssText = `
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: linear-gradient(135deg, #ef4444, #dc2626);
            color: white;
            border-radius: 12px;
            padding: 16px 24px;
            box-shadow: 0 4px 15px rgba(239, 68, 68, 0.3);
            z-index: 10000;
            animation: slideInTop 0.5s ease;
        `;

        document.body.appendChild(notification);

        // Auto-hide after 3 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOutTop 0.5s ease;
            setTimeout(() => notification.remove(), 500);
        }, 3000);
    }

    updateUIAfterReward(data) {
        switch (data.rewardType) {
            case 'energy':
                // Update energy bar
                this.updateEnergyBar(data.data.energy);
                break;
            case 'coins':
                // Update coin balance
                this.updateCoinBalance(data.data.coins);
                break;
        }
    }

    updateEnergyBar(energy) {
        const energyBar = document.querySelector('.energy-bar');
        if (energyBar) {
            energyBar.style.width = `${(energy / 1000) * 100}%`;
        }
        
        const energyText = document.querySelector('.energy-text');
        if (energyText) {
            energyText.textContent = `${energy}/1000`;
        }
    }

    updateCoinBalance(coins) {
        const coinBalance = document.querySelector('.coin-balance');
        if (coinBalance) {
            coinBalance.textContent = coins.toFixed(2);
        }
    }

    enableAdTrigger(trigger) {
        // Re-enable the trigger button if needed
        // This depends on your UI implementation
        console.log(`Ad trigger ${trigger} re-enabled`);
    }

    // Helper methods
    getCurrentUserID() {
        // Get user ID from localStorage, sessionStorage, or cookie
        return localStorage.getItem('user_id') || sessionStorage.getItem('user_id');
    }

    getCurrentEnergy() {
        // Get current energy from your game state
        const energyBar = document.querySelector('.energy-bar');
        if (energyBar) {
            const width = parseFloat(energyBar.style.width);
            return Math.round((width / 100) * 1000);
        }
        return 0;
    }

    // Public API
    async triggerAd(trigger) {
        const adConfig = this.availableRewards.find(r => r.trigger === trigger);
        if (!adConfig) {
            console.log('No ad available for trigger:', trigger);
            return;
        }

        switch (trigger) {
            case 'energy_depleted':
                this.showEnergyAdOption(adConfig);
                break;
            case 'daily_bonus':
                this.showDailyBonusOption(adConfig);
                break;
            case 'achievement':
                this.showAchievementBonusOption(adConfig);
                break;
            case 'tournament_entry':
                this.showTournamentEntryOption(adConfig);
                break;
            case 'nft_purchase':
                this.showNFTDiscountOption(adConfig);
                break;
            default:
                console.log('Unknown trigger:', trigger);
        }
    }

    getAdStats() {
        return this.userStats;
    }

    hasAvailableAds() {
        return this.availableRewards.length > 0;
    }
}

// Add CSS animations and styles
const secureAdStyles = `
@keyframes slideInRight {
    from {
        transform: translateX(100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}

@keyframes slideOutRight {
    from {
        transform: translateX(0);
        opacity: 1;
    }
    to {
        transform: translateX(100%);
        opacity: 0;
    }
}

@keyframes slideInLeft {
    from {
        transform: translateX(-100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}

@keyframes slideOutLeft {
    from {
        transform: translateX(0);
        opacity: 1;
    }
    to {
        transform: translateX(-100%);
        opacity: 0;
    }
}

@keyframes slideInTop {
    from {
        transform: translate(-50%, -100%);
        opacity: 0;
    }
    to {
        transform: translate(-50%, 0);
        opacity: 1;
    }
}

@keyframes slideOutTop {
    from {
        transform: translate(-50%, 0);
        opacity: 1;
    }
    to {
        transform: translate(-50%, -100%);
        opacity: 0;
    }
}

@keyframes achievementPop {
    0% {
        transform: translate(-50%, -50%) scale(0);
        opacity: 0;
    }
    50% {
        transform: translate(-50%, -50%) scale(1.1);
        opacity: 1;
    }
    100% {
        transform: translate(-50%, -50%) scale(1);
        opacity: 1;
    }
}

@keyframes fadeIn {
    from {
        opacity: 0;
    }
    to {
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

@keyframes spin {
    0% {
        transform: rotate(0deg);
    }
    100% {
        transform: rotate(360deg);
    }
}

.energy-restore-btn .btn-content {
    display: flex;
    align-items: center;
    gap: 8px;
}

.energy-restore-btn .icon {
    font-size: 1.2em;
}

.energy-restore-btn .text {
    font-weight: 600;
}

.energy-restore-btn .subtext {
    font-size: 0.8em;
    opacity: 0.9;
}

.achievement-bonus-option .bonus-content,
.daily-bonus-option .bonus-content,
.tournament-entry-option .entry-content,
.nft-discount-option .discount-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
}

.achievement-bonus-option .icon,
.daily-bonus-option .icon,
.tournament-entry-option .icon,
.nft-discount-option .icon {
    font-size: 2em;
}

.achievement-bonus-option .reward,
.daily-bonus-option .reward,
.nft-discount-option .discount {
    font-weight: bold;
    font-size: 1.1em;
}

.watch-ad-btn {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.3);
    color: white;
    border-radius: 8px;
    padding: 8px 16px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.3s ease;
}

.watch-ad-btn:hover {
    background: rgba(255, 255, 255, 0.3);
    transform: translateY(-1px);
}

.ad-reward-notification .notification-content,
.ad-expiration-notification .notification-content,
.ad-error-notification .notification-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
}

.ad-reward-notification .reward-text,
.ad-expiration-notification .error-text,
.ad-error-notification .error-text {
    font-weight: 600;
    font-size: 1.1em;
}

.ad-reward-notification .thank-you,
.ad-expiration-notification .retry-text,
.ad-error-notification .retry-text {
    font-size: 0.9em;
    opacity: 0.9;
}

.ad-loading-overlay .loading-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    color: white;
}

.ad-loading-overlay .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid rgba(255, 255, 255, 0.3);
    border-top: 4px solid white;
    border-radius: 50%;
    animation: spin 1s linear infinite;
}

.ad-loading-overlay .loading-text {
    font-size: 1.2em;
    font-weight: 600;
}

.ad-loading-overlay .loading-subtext {
    font-size: 0.9em;
    opacity: 0.8;
}
`;

// Inject styles
const styleSheet = document.createElement('style');
styleSheet.textContent = secureAdStyles;
document.head.appendChild(styleSheet);

// Initialize secure ads manager when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.secureAdsManager = new SecureAdsManager();
    window.secureAdsManager.initialize();
});

// Export for manual usage
window.BKCSecureAds = {
    triggerAd: (trigger) => window.secureAdsManager?.triggerAd(trigger),
    getStats: () => window.secureAdsManager?.getAdStats(),
    hasAds: () => window.secureAdsManager?.hasAvailableAds()
};
