/**
 * API Service with Load Balancing
 * Handles all API requests with automatic failover
 */

import loadBalancer from '../utils/loadBalancer';

class ApiService {
  constructor() {
    this.basePath = '/api';
  }

  /**
   * Generic API request method
   */
  async request(endpoint, options = {}) {
    const path = `${this.basePath}${endpoint}`;
    return await loadBalancer.apiRequest(path, options);
  }

  /**
   * GET request
   */
  async get(endpoint, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const path = queryString ? `${endpoint}?${queryString}` : endpoint;
    
    return await this.request(path, {
      method: 'GET'
    });
  }

  /**
   * POST request
   */
  async post(endpoint, data = {}) {
    return await this.request(endpoint, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  /**
   * PUT request
   */
  async put(endpoint, data = {}) {
    return await this.request(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  /**
   * DELETE request
   */
  async delete(endpoint) {
    return await this.request(endpoint, {
      method: 'DELETE'
    });
  }

  // User related APIs

  /**
   * Get user profile
   */
  async getUserProfile(userId) {
    return await this.get(`/user/${userId}/profile`);
  }

  /**
   * Update user profile
   */
  async updateUserProfile(userId, data) {
    return await this.put(`/user/${userId}/profile`, data);
  }

  /**
   * Get user balance
   */
  async getUserBalance(userId) {
    return await this.get(`/user/${userId}/balance`);
  }

  /**
   * Get user energy
   */
  async getUserEnergy(userId) {
    return await this.get(`/user/${userId}/energy`);
  }

  // Game related APIs

  /**
   * Process tap action
   */
  async processTap(userId, tapAmount) {
    return await this.post('/game/tap', {
      userId,
      tapAmount,
      timestamp: Date.now()
    });
  }

  /**
   * Get daily reward
   */
  async claimDailyReward(userId) {
    return await this.post('/game/daily-reward', {
      userId
    });
  }

  /**
   * Get user stats
   */
  async getUserStats(userId) {
    return await this.get(`/user/${userId}/stats`);
  }

  /**
   * Get leaderboard
   */
  async getLeaderboard(category = 'balance', limit = 100) {
    return await this.get('/leaderboard', {
      category,
      limit
    });
  }

  // P2P Market APIs

  /**
   * Create P2P transaction
   */
  async createP2PTransaction(transactionData) {
    return await this.post('/p2p/create', transactionData);
  }

  /**
   * Get P2P transaction details
   */
  async getP2PTransaction(transactionId) {
    return await this.get(`/p2p/transaction/${transactionId}`);
  }

  /**
   * Get user P2P transactions
   */
  async getUserP2PTransactions(userId, page = 1, limit = 20) {
    return await this.get(`/p2p/user/${userId}/transactions`, {
      page,
      limit
    });
  }

  /**
   * Confirm P2P payment
   */
  async confirmP2PPayment(transactionId, adminId) {
    return await this.post(`/p2p/confirm`, {
      transactionId,
      adminId
    });
  }

  /**
   * Dispute P2P transaction
   */
  async disputeP2PTransaction(transactionId, reason, adminId) {
    return await this.post(`/p2p/dispute`, {
      transactionId,
      reason,
      adminId
    });
  }

  // Bank APIs

  /**
   * Get bank loan offers
   */
  async getBankLoanOffers(userId) {
    return await this.get(`/bank/offers/${userId}`);
  }

  /**
   * Apply for bank loan
   */
  async applyBankLoan(userId, amount) {
    return await this.post('/bank/loan/apply', {
      userId,
      amount
    });
  }

  /**
   * Get user loans
   */
  async getUserLoans(userId) {
    return await this.get(`/bank/loans/${userId}`);
  }

  /**
   * Repay bank loan
   */
  async repayBankLoan(loanId, amount) {
    return await this.post('/bank/loan/repay', {
      loanId,
      amount
    });
  }

  // Admin APIs

  /**
   * Get system statistics
   */
  async getSystemStats() {
    return await this.get('/admin/stats');
  }

  /**
   * Get cluster health
   */
  async getClusterHealth() {
    return await this.get('/admin/cluster-health');
  }

  /**
   * Get user list (admin)
   */
  async getUserList(page = 1, limit = 50, filters = {}) {
    return await this.get('/admin/users', {
      page,
      limit,
      ...filters
    });
  }

  /**
   * Ban user (admin)
   */
  async banUser(userId, reason, duration) {
    return await this.post('/admin/users/ban', {
      userId,
      reason,
      duration
    });
  }

  /**
   * Get transaction logs (admin)
   */
  async getTransactionLogs(page = 1, limit = 100, filters = {}) {
    return await this.get('/admin/transactions', {
      page,
      limit,
      ...filters
    });
  }

  // WebSocket connection

  /**
   * Get WebSocket URL with load balancing
   */
  async getWebSocketUrl() {
    const node = loadBalancer.getCurrentNode();
    return node.replace('https://', 'wss://').replace('http://', 'ws://') + '/ws';
  }

  // Health check

  /**
   * Check API health
   */
  async healthCheck() {
    try {
      const response = await this.get('/health');
      return { healthy: true, data: response };
    } catch (error) {
      return { healthy: false, error: error.message };
    }
  }

  /**
   * Get load balancer stats
   */
  getLoadBalancerStats() {
    return loadBalancer.getStats();
  }
}

// Create singleton instance
const apiService = new ApiService();

export default apiService;
