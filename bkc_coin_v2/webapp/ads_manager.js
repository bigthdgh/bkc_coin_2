// BKC Coin - Smart Ads Manager
class AdsManager {
    constructor() {
        this.isInitialized = false;
        this.currentAd = null;
        this.adsgramLoaded = false;
        this.userStats = null;
        this.availableRewards = [];
    }

    async initialize() {
        try {
            await this.loadAdsgramSDK();
            await this.loadUserStats();
            await this.loadAvailableRewards();
            this.setupEventListeners();
            this.isInitialized = true;
            console.log('Ads Manager initialized successfully');
        } catch (error) {
            console.error('Failed to initialize Ads Manager:', error);
        }
    }

    async loadAdsgramSDK() {
        return new Promise((resolve, reject) => {
            // Check if already loaded
            if (window.Adsgram) {
                this.adsgramLoaded = true;
                resolve();
                return;
            }

            // Load Adsgram SDK
            const script = document.createElement('script');
            script.src = 'https://static.adsgram.com/js/adsgram.js';
            script.async = true;
            script.onload = () => {
                this.adsgramLoaded = true;
                console.log('Adsgram SDK loaded');
                resolve();
            };
            script.onerror = () => {
                console.error('Failed to load Adsgram SDK');
                reject(new Error('Failed to load Adsgram SDK'));
            };
            document.head.appendChild(script);
        });
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

    // Smart trigger handlers - only show ads when user wants them
    async handleEnergyDepleted() {
        if (!this.shouldShowAd('energy_depleted')) return;

        const energyAd = this.availableRewards.find(r => r.trigger === 'energy_depleted');
        if (!energyAd) return;

        this.showEnergyAdOption(energyAd);
    }

    showEnergyAdOption(adConfig) {
        // Create non-intrusive energy restoration option
        const energyButton = document.createElement('button');
        energyButton.className = 'energy-restore-btn';
        energyButton.innerHTML = `
            <div class="btn-content">
                <span class="icon">⚡</span>
                <span class="text">Восстановить энергию (${adConfig.reward})</span>
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
            this.showRewardedAd(adConfig);
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

        // Show achievement bonus option
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
            this.showRewardedAd(adConfig);
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

    async handleLevelUp() {
        // Show interstitial ad (non-rewarded, less intrusive)
        const levelUpAd = this.availableRewards.find(r => r.trigger === 'level_up');
        if (!levelUpAd) return;

        // Only show interstitial (no reward)
        this.showInterstitialAd(levelUpAd);
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
            this.showRewardedAd(adConfig);
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
            this.showRewardedAd(adConfig);
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
                <span class="discount">${(adConfig.reward * 100).toFixed(0)}%</span>
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
            this.showRewardedAd(adConfig);
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
        if (this.userStats.ads_today >= 20) return false; // Max 20 ads per day
        
        // Check if user recently watched an ad (avoid spam)
        const lastAdTime = localStorage.getItem('lastAdTime');
        if (lastAdTime) {
            const timeDiff = Date.now() - parseInt(lastAdTime);
            if (timeDiff < 60000) return false; // Wait at least 1 minute
        }

        return true;
    }

    async showRewardedAd(adConfig) {
        if (!this.adsgramLoaded) {
            console.error('Adsgram SDK not loaded');
            return;
        }

        try {
            // Record ad start
            await fetch('/api/v1/ads/start', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    placement_id: adConfig.placement_id,
                    trigger: adConfig.trigger
                })
            });

            // Show the ad
            await this.showAdsgramRewardedAd(adConfig.placement_id);

            // Record completion and grant reward
            const response = await fetch('/api/v1/ads/complete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    placement_id: adConfig.placement_id,
                    trigger: adConfig.trigger
                })
            });

            const result = await response.json();
            
            // Show reward notification
            this.showRewardNotification(adConfig, result.reward);

            // Update stats
            await this.loadUserStats();
            
            // Store last ad time
            localStorage.setItem('lastAdTime', Date.now().toString());

        } catch (error) {
            console.error('Failed to show rewarded ad:', error);
            this.showAdErrorNotification();
        }
    }

    async showInterstitialAd(adConfig) {
        if (!this.adsgramLoaded) {
            console.error('Adsgram SDK not loaded');
            return;
        }

        try {
            // Show interstitial ad (no reward)
            await this.showAdsgramInterstitial(adConfig.placement_id);
        } catch (error) {
            console.error('Failed to show interstitial ad:', error);
        }
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

    showRewardNotification(adConfig, reward) {
        const notification = document.createElement('div');
        notification.className = 'ad-reward-notification';
        
        let rewardText = '';
        switch (adConfig.reward_type) {
            case 'energy':
                rewardText = `⚡ +${reward} энергии`;
                break;
            case 'coins':
                rewardText = `💰 +${reward} BKC`;
                break;
            case 'tournament_entry':
                rewardText = '🏆 Бесплатный вход в турнир';
                break;
            case 'discount':
                rewardText = `🎨 Скидка ${(reward * 100).toFixed(0)}% на NFT`;
                break;
            default:
                rewardText = `🎁 Награда получена!`;
        }

        notification.innerHTML = `
            <div class="notification-content">
                <span class="reward-text">${rewardText}</span>
                <span class="thank-you">Спасибо за просмотр рекламы!</span>
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

    showAdErrorNotification() {
        const notification = document.createElement('div');
        notification.className = 'ad-error-notification';
        notification.innerHTML = `
            <div class="notification-content">
                <span class="error-text">Не удалось загрузить рекламу</span>
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

    // Public API for manual ad triggers
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

    // Get current ad stats
    getAdStats() {
        return this.userStats;
    }

    // Check if ads are available
    hasAvailableAds() {
        return this.availableRewards.length > 0;
    }
}

// Add CSS animations
const adStyles = `
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

@keyframes fadeOut {
    from {
        opacity: 1;
    }
    to {
        opacity: 0;
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
.ad-error-notification .notification-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
}

.ad-reward-notification .reward-text,
.ad-error-notification .error-text {
    font-weight: 600;
    font-size: 1.1em;
}

.ad-reward-notification .thank-you,
.ad-error-notification .retry-text {
    font-size: 0.9em;
    opacity: 0.9;
}
`;

// Inject styles
const styleSheet = document.createElement('style');
styleSheet.textContent = adStyles;
document.head.appendChild(styleSheet);

// Initialize ads manager when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.adsManager = new AdsManager();
    window.adsManager.initialize();
});

// Export for manual usage
window.BKCAds = {
    triggerAd: (trigger) => window.adsManager?.triggerAd(trigger),
    getStats: () => window.adsManager?.getAdStats(),
    hasAds: () => window.adsManager?.hasAvailableAds()
};
