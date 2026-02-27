// BKC Coin - Enhanced Data Visualization Dashboard
class BKCDashboard {
    constructor() {
        this.charts = new Map();
        this.realTimeData = new Map();
        this.updateInterval = null;
        this.isInitialized = false;
    }

    async initialize() {
        try {
            await this.loadCharts();
            this.setupEventListeners();
            this.startRealTimeUpdates();
            this.isInitialized = true;
            console.log('Dashboard initialized successfully');
        } catch (error) {
            console.error('Failed to initialize dashboard:', error);
        }
    }

    async loadCharts() {
        // Load Chart.js if not already loaded
        if (typeof Chart === 'undefined') {
            await this.loadChartJS();
        }

        // Initialize all charts
        this.initializePriceChart();
        this.initializeVolumeChart();
        this.initializeUserGrowthChart();
        this.initializeEconomyHealthChart();
        this.initializeRealTimeMetrics();
        this.initializeLeaderboard();
        this.initializeHeatmap();
    }

    async loadChartJS() {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = 'https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js';
            script.onload = resolve;
            script.onerror = reject;
            document.head.appendChild(script);
        });
    }

    initializePriceChart() {
        const ctx = document.getElementById('priceChart');
        if (!ctx) return;

        this.charts.set('price', new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'BKC Price (USD)',
                    data: [],
                    borderColor: '#3cd6c6',
                    backgroundColor: 'rgba(60, 214, 198, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true
                }, {
                    label: 'Target Price',
                    data: [],
                    borderColor: '#ff6b6b',
                    borderWidth: 2,
                    borderDash: [5, 5],
                    fill: false
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        labels: {
                            color: '#ffffff'
                        }
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false,
                        callbacks: {
                            label: function(context) {
                                return `${context.dataset.label}: $${context.parsed.y.toFixed(6)}`;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8',
                            callback: function(value) {
                                return '$' + value.toFixed(6);
                            }
                        }
                    }
                }
            }
        }));
    }

    initializeVolumeChart() {
        const ctx = document.getElementById('volumeChart');
        if (!ctx) return;

        this.charts.set('volume', new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Trading Volume (BKC)',
                    data: [],
                    backgroundColor: 'rgba(34, 211, 238, 0.6)',
                    borderColor: '#22d3ee',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        labels: {
                            color: '#ffffff'
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return `Volume: ${context.parsed.y.toLocaleString()} BKC`;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8',
                            callback: function(value) {
                                return value.toLocaleString();
                            }
                        }
                    }
                }
            }
        }));
    }

    initializeUserGrowthChart() {
        const ctx = document.getElementById('userGrowthChart');
        if (!ctx) return;

        this.charts.set('userGrowth', new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Active Users',
                    data: [],
                    borderColor: '#3cd6c6',
                    backgroundColor: 'rgba(60, 214, 198, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true
                }, {
                    label: 'New Users',
                    data: [],
                    borderColor: '#feca57',
                    backgroundColor: 'rgba(254, 202, 87, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        labels: {
                            color: '#ffffff'
                        }
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8',
                            callback: function(value) {
                                return value.toLocaleString();
                            }
                        }
                    }
                }
            }
        }));
    }

    initializeEconomyHealthChart() {
        const ctx = document.getElementById('economyHealthChart');
        if (!ctx) return;

        this.charts.set('economyHealth', new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Healthy', 'Warning', 'Critical'],
                datasets: [{
                    data: [70, 20, 10],
                    backgroundColor: [
                        '#3cd6c6',
                        '#feca57',
                        '#ff6b6b'
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                        labels: {
                            color: '#ffffff'
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return `${context.label}: ${context.parsed}%`;
                            }
                        }
                    }
                }
            }
        }));
    }

    initializeRealTimeMetrics() {
        const ctx = document.getElementById('realTimeMetricsChart');
        if (!ctx) return;

        this.charts.set('realTimeMetrics', new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'TPS (Taps/Second)',
                    data: [],
                    borderColor: '#3cd6c6',
                    backgroundColor: 'rgba(60, 214, 198, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true
                }, {
                    label: 'Active Users',
                    data: [],
                    borderColor: '#22d3ee',
                    backgroundColor: 'rgba(34, 211, 238, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                animation: {
                    duration: 0 // Disable animations for real-time data
                },
                plugins: {
                    legend: {
                        display: true,
                        labels: {
                            color: '#ffffff'
                        }
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8'
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8'
                        }
                    }
                }
            }
        }));
    }

    initializeLeaderboard() {
        const ctx = document.getElementById('leaderboardChart');
        if (!ctx) return;

        this.charts.set('leaderboard', new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [{
                    label: 'Top Players',
                    data: [],
                    backgroundColor: 'rgba(60, 214, 198, 0.6)',
                    borderColor: '#3cd6c6',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                indexAxis: 'y',
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return `Score: ${context.parsed.x.toLocaleString()}`;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8',
                            callback: function(value) {
                                return value.toLocaleString();
                            }
                        }
                    },
                    y: {
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8'
                        }
                    }
                }
            }
        }));
    }

    initializeHeatmap() {
        const ctx = document.getElementById('activityHeatmap');
        if (!ctx) return;

        this.charts.set('heatmap', new Chart(ctx, {
            type: 'scatter',
            data: {
                datasets: [{
                    label: 'User Activity',
                    data: [],
                    backgroundColor: 'rgba(60, 214, 198, 0.6)'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                const point = context.raw;
                                return `Activity: ${point.value} at ${point.hour}:00`;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        type: 'linear',
                        position: 'bottom',
                        min: 0,
                        max: 23,
                        title: {
                            display: true,
                            text: 'Hour of Day',
                            color: '#ffffff'
                        },
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8',
                            stepSize: 1
                        }
                    },
                    y: {
                        type: 'linear',
                        position: 'left',
                        min: 0,
                        max: 6,
                        title: {
                            display: true,
                            text: 'Day of Week',
                            color: '#ffffff'
                        },
                        grid: {
                            color: 'rgba(255, 255, 255, 0.1)'
                        },
                        ticks: {
                            color: '#94a3b8',
                            callback: function(value) {
                                const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
                                return days[value];
                            }
                        }
                    }
                }
            }
        }));
    }

    setupEventListeners() {
        // Time range selector
        const timeRangeSelector = document.getElementById('timeRangeSelector');
        if (timeRangeSelector) {
            timeRangeSelector.addEventListener('change', (e) => {
                this.updateTimeRange(e.target.value);
            });
        }

        // Chart type selector
        const chartTypeSelector = document.getElementById('chartTypeSelector');
        if (chartTypeSelector) {
            chartTypeSelector.addEventListener('change', (e) => {
                this.updateChartType(e.target.value);
            });
        }

        // Export data button
        const exportButton = document.getElementById('exportData');
        if (exportButton) {
            exportButton.addEventListener('click', () => {
                this.exportData();
            });
        }

        // Refresh button
        const refreshButton = document.getElementById('refreshData');
        if (refreshButton) {
            refreshButton.addEventListener('click', () => {
                this.refreshAllData();
            });
        }
    }

    startRealTimeUpdates() {
        this.updateInterval = setInterval(() => {
            this.updateRealTimeData();
        }, 5000); // Update every 5 seconds
    }

    async updateRealTimeData() {
        try {
            const response = await fetch('/api/v1/analytics/realtime');
            const data = await response.json();

            // Update real-time metrics
            this.updateRealTimeMetricsChart(data);
            this.updateStatCards(data);

        } catch (error) {
            console.error('Failed to update real-time data:', error);
        }
    }

    updateRealTimeMetricsChart(data) {
        const chart = this.charts.get('realTimeMetrics');
        if (!chart) return;

        const now = new Date();
        const timeLabel = now.toLocaleTimeString();

        // Add new data point
        chart.data.labels.push(timeLabel);
        chart.data.datasets[0].data.push(data.tps || 0);
        chart.data.datasets[1].data.push(data.activeUsers || 0);

        // Keep only last 20 data points
        if (chart.data.labels.length > 20) {
            chart.data.labels.shift();
            chart.data.datasets[0].data.shift();
            chart.data.datasets[1].data.shift();
        }

        chart.update('none'); // Update without animation
    }

    updateStatCards(data) {
        // Update various stat cards
        this.updateStatCard('totalUsers', data.totalUsers?.toLocaleString() || '0');
        this.updateStatCard('totalTransactions', data.totalTransactions?.toLocaleString() || '0');
        this.updateStatCard('currentPrice', `$${data.currentPrice || '0.000000'}`);
        this.updateStatCard('marketCap', `$${(data.marketCap || 0).toLocaleString()}`);
        this.updateStatCard('totalSupply', `${(data.totalSupply || 0).toLocaleString()} BKC`);
        this.updateStatCard('burnRate', `${(data.burnRate || 0).toFixed(2)}%`);
    }

    updateStatCard(cardId, value) {
        const card = document.getElementById(cardId);
        if (card) {
            card.textContent = value;
            card.classList.add('updated');
            setTimeout(() => card.classList.remove('updated'), 300);
        }
    }

    async updateTimeRange(range) {
        try {
            const response = await fetch(`/api/v1/analytics/history?range=${range}`);
            const data = await response.json();

            // Update all charts with new data
            this.updateChartData('price', data.priceHistory);
            this.updateChartData('volume', data.volumeHistory);
            this.updateChartData('userGrowth', data.userGrowth);

        } catch (error) {
            console.error('Failed to update time range:', error);
        }
    }

    updateChartData(chartName, data) {
        const chart = this.charts.get(chartName);
        if (!chart) return;

        chart.data.labels = data.labels || [];
        chart.data.datasets[0].data = data.data || [];
        chart.update();
    }

    updateChartType(type) {
        // Update chart types based on user preference
        this.charts.forEach((chart, name) => {
            if (name === 'price' || name === 'volume') {
                chart.config.type = type;
                chart.update();
            }
        });
    }

    async refreshAllData() {
        try {
            // Show loading state
            this.showLoadingState();

            // Fetch fresh data
            const response = await fetch('/api/v1/analytics/dashboard');
            const data = await response.json();

            // Update all charts
            this.updateAllCharts(data);

            // Hide loading state
            this.hideLoadingState();

        } catch (error) {
            console.error('Failed to refresh data:', error);
            this.hideLoadingState();
        }
    }

    updateAllCharts(data) {
        // Update price chart
        if (data.priceHistory) {
            this.updateChartData('price', data.priceHistory);
        }

        // Update volume chart
        if (data.volumeHistory) {
            this.updateChartData('volume', data.volumeHistory);
        }

        // Update user growth chart
        if (data.userGrowth) {
            this.updateChartData('userGrowth', data.userGrowth);
        }

        // Update economy health
        if (data.economyHealth) {
            this.updateEconomyHealth(data.economyHealth);
        }

        // Update leaderboard
        if (data.leaderboard) {
            this.updateLeaderboard(data.leaderboard);
        }

        // Update heatmap
        if (data.activityHeatmap) {
            this.updateHeatmap(data.activityHeatmap);
        }
    }

    updateEconomyHealth(healthData) {
        const chart = this.charts.get('economyHealth');
        if (!chart) return;

        chart.data.datasets[0].data = [
            healthData.healthy || 0,
            healthData.warning || 0,
            healthData.critical || 0
        ];
        chart.update();
    }

    updateLeaderboard(leaderboardData) {
        const chart = this.charts.get('leaderboard');
        if (!chart) return;

        chart.data.labels = leaderboardData.labels || [];
        chart.data.datasets[0].data = leaderboardData.data || [];
        chart.update();
    }

    updateHeatmap(heatmapData) {
        const chart = this.charts.get('heatmap');
        if (!chart) return;

        chart.data.datasets[0].data = heatmapData || [];
        chart.update();
    }

    exportData() {
        try {
            // Collect all chart data
            const exportData = {
                timestamp: new Date().toISOString(),
                charts: {}
            };

            this.charts.forEach((chart, name) => {
                exportData.charts[name] = {
                    labels: chart.data.labels,
                    datasets: chart.data.datasets
                };
            });

            // Create download link
            const dataStr = JSON.stringify(exportData, null, 2);
            const dataBlob = new Blob([dataStr], { type: 'application/json' });
            const url = URL.createObjectURL(dataBlob);
            
            const link = document.createElement('a');
            link.href = url;
            link.download = `bkc-dashboard-${new Date().toISOString().split('T')[0]}.json`;
            link.click();
            
            URL.revokeObjectURL(url);

        } catch (error) {
            console.error('Failed to export data:', error);
        }
    }

    showLoadingState() {
        const loadingOverlay = document.getElementById('loadingOverlay');
        if (loadingOverlay) {
            loadingOverlay.style.display = 'flex';
        }
    }

    hideLoadingState() {
        const loadingOverlay = document.getElementById('loadingOverlay');
        if (loadingOverlay) {
            loadingOverlay.style.display = 'none';
        }
    }

    destroy() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
        }

        this.charts.forEach(chart => {
            if (chart) {
                chart.destroy();
            }
        });

        this.charts.clear();
        this.realTimeData.clear();
        this.isInitialized = false;
    }
}

// Initialize dashboard when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.bkcDashboard = new BKCDashboard();
    window.bkcDashboard.initialize();
});

// Handle page visibility changes
document.addEventListener('visibilitychange', () => {
    if (window.bkcDashboard) {
        if (document.hidden) {
            // Pause updates when page is hidden
            if (window.bkcDashboard.updateInterval) {
                clearInterval(window.bkcDashboard.updateInterval);
            }
        } else {
            // Resume updates when page is visible
            window.bkcDashboard.startRealTimeUpdates();
        }
    }
});

// Handle window resize
window.addEventListener('resize', () => {
    if (window.bkcDashboard && window.bkcDashboard.isInitialized) {
        // Resize all charts
        window.bkcDashboard.charts.forEach(chart => {
            if (chart) {
                chart.resize();
            }
        });
    }
});
