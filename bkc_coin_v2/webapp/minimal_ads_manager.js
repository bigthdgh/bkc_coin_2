// BKC Coin - Minimal Ads Manager (ONLY Energy + 30 BKC)
class MinimalAdsManager {
    constructor() {
        this.isInitialized = false;
        this.websocket = null;
        this.userStats = null;
        this.availableAds = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
    }

    async initialize() {
        try {
            await this.loadUserStats();
            await this.loadAvailableAds();
            this.connectWebSocket();
            this.isInitialized = true;
            console.log('Minimal Ads Manager initialized successfully');
        } catch (error) {
            console.error('Failed to initialize Minimal Ads Manager:', error);
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
            }, 1000 * this.reconnectAttempts);
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
        
        // Update UI
        this.updateUIAfterReward(data);
    }

    handleAdExpired(data) {
        console.log('Ad expired:', data);
        
        // Remove loading state
        this.hideAdLoading();
        
        // Show expiration notification
        this.showExpirationNotification(data);
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

    async loadAvailableAds() {
        try {
            const response = await fetch('/api/v1/ads/available');
            this.availableAds = await response.json();
        } catch (error) {
            console.error('Failed to load available ads:', error);
            this.availableAds = null;
        }
    }

    // Check and show energy ad if needed
    checkEnergyAd() {
        if (!this.availableAds || !this.availableAds.energy) {
            return;
        }

        // Check if energy is actually depleted
        const currentEnergy = this.getCurrentEnergy();
        if (currentEnergy > 100) {
            return; // Energy not depleted enough
        }

        this.showEnergyAdOption();
    }

    showEnergyAdOption() {
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
            this.startEnergyAd();
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

    // Show coins ad option (manual trigger)
    showCoinsAdOption() {
        if (!this.availableAds || !this.availableAds.coins) {
            this.showAdErrorNotification('Реклама за монеты недоступна');
            return;
        }

        const coinsButton = document.createElement('button');
        coinsButton.className = 'coins-reward-btn';
        coinsButton.innerHTML = `
            <div class="btn-content">
                <span class="icon">💰</span>
                <span class="text">Получить 30 BKC</span>
                <span class="subtext">Посмотреть рекламу</span>
            </div>
        `;

        coinsButton.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: linear-gradient(135deg, #feca57, #ff6b6b);
            color: white;
            border: none;
            border-radius: 12px;
            padding: 12px 20px;
            cursor: pointer;
            box-shadow: 0 4px 15px rgba(254, 202, 87, 0.3);
            z-index: 1000;
            font-family: 'Inter', sans-serif;
            font-weight: 500;
            transition: all 0.3s ease;
            animation: slideInRight 0.5s ease;
        `;

        coinsButton.addEventListener('mouseenter', () => {
            coinsButton.style.transform = 'translateY(-2px)';
            coinsButton.style.boxShadow = '0 6px 20px rgba(254, 202, 87, 0.4)';
        });

        coinsButton.addEventListener('mouseleave', () => {
            coinsButton.style.transform = 'translateY(0)';
            coinsButton.style.boxShadow = '0 4px 15px rgba(254, 202, 87, 0.3)';
        });

        coinsButton.addEventListener('click', () => {
            this.startCoinsAd();
            coinsButton.remove();
        });

        // Auto-hide after 15 seconds
        setTimeout(() => {
            if (coinsButton.parentNode) {
                coinsButton.style.animation = 'slideOutRight 0.5s ease';
                setTimeout(() => coinsButton.remove(), 500);
            }
        }, 15000);

        document.body.appendChild(coinsButton);
    }

    async startEnergyAd() {
        try {
            // Start ad session on backend
            const response = await fetch('/api/v1/ads/session/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ad_type: 'energy'
                })
            });

            if (!response.ok) {
                const error = await response.json();
                this.showAdErrorNotification(error.message || 'Не удалось начать сессию');
                return;
            }

            const sessionData = await response.json();
            
            // Show loading state
            this.showAdLoading('energy');
            
            // Load Adsgram and show ad
            await this.loadAdsgramAndShowAd(sessionData.placement_id);

        } catch (error) {
            console.error('Failed to start energy ad:', error);
            this.showAdErrorNotification('Не удалось начать сессию рекламы');
        }
    }

    async startCoinsAd() {
        try {
            // Start ad session on backend
            const response = await fetch('/api/v1/ads/session/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ad_type: 'coins'
                })
            });

            if (!response.ok) {
                const error = await response.json();
                this.showAdErrorNotification(error.message || 'Не удалось начать сессию');
                return;
            }

            const sessionData = await response.json();
            
            // Show loading state
            this.showAdLoading('coins');
            
            // Load Adsgram and show ad
            await this.loadAdsgramAndShowAd(sessionData.placement_id);

        } catch (error) {
            console.error('Failed to start coins ad:', error);
            this.showAdErrorNotification('Не удалось начать сессию рекламы');
        }
    }

    showAdLoading(adType) {
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

    async loadAdsgramAndShowAd(placementId) {
        // Load Adsgram SDK if not already loaded
        if (!window.Adsgram) {
            await this.loadAdsgramSDK();
        }

        try {
            await this.showAdsgramRewardedAd(placementId);
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

    showRewardNotification(data) {
        const notification = document.createElement('div');
        notification.className = 'ad-reward-notification';
        
        let rewardText = '';
        switch (data.ad_type) {
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
            notification.style.animation = 'slideOutTop 0.5s ease';
            setTimeout(() => notification.remove(), 500);
        }, 3000);
    }

    updateUIAfterReward(data) {
        switch (data.ad_type) {
            case 'energy':
                // Update energy bar
                this.updateEnergyBar(data.data.energy);
                break;
            case 'coins':
                // Update coin balance
                this.updateCoinBalance(data.reward);
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
            const currentBalance = parseFloat(coinBalance.textContent) || 0;
            coinBalance.textContent = (currentBalance + coins).toFixed(2);
        }
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
    showCoinsAd() {
        this.showCoinsAdOption();
    }

    checkEnergy() {
        this.checkEnergyAd();
    }

    getAdStats() {
        return this.userStats;
    }

    hasAvailableAds() {
        return this.availableAds && (this.availableAds.energy || this.availableAds.coins);
    }
}

// Add CSS animations and styles
const minimalAdStyles = `
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

.energy-restore-btn .btn-content,
.coins-reward-btn .btn-content {
    display: flex;
    align-items: center;
    gap: 8px;
}

.energy-restore-btn .icon,
.coins-reward-btn .icon {
    font-size: 1.2em;
}

.energy-restore-btn .text,
.coins-reward-btn .text {
    font-weight: 600;
}

.energy-restore-btn .subtext,
.coins-reward-btn .subtext {
    font-size: 0.8em;
    opacity: 0.9;
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
styleSheet.textContent = minimalAdStyles;
document.head.appendChild(styleSheet);

// Initialize minimal ads manager when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.minimalAdsManager = new MinimalAdsManager();
    window.minimalAdsManager.initialize();
    
    // Auto-check energy every 5 seconds
    setInterval(() => {
        if (window.minimalAdsManager) {
            window.minimalAdsManager.checkEnergy();
        }
    }, 5000);
});

// Export for manual usage
window.BKCMinimalAds = {
    showCoinsAd: () => window.minimalAdsManager?.showCoinsAd(),
    checkEnergy: () => window.minimalAdsManager?.checkEnergy(),
    getStats: () => window.minimalAdsManager?.getAdStats(),
    hasAds: () => window.minimalAdsManager?.hasAvailableAds()
};
