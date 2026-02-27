/**
 * Load Balancer for 25 Render Nodes
 * Implements client-side load distribution with failover
 */

// Render node URLs - these should be your actual Render deployment URLs
const RENDER_NODES = [
  'https://bkc-node-1.onrender.com',
  'https://bkc-node-2.onrender.com',
  'https://bkc-node-3.onrender.com',
  'https://bkc-node-4.onrender.com',
  'https://bkc-node-5.onrender.com',
  'https://bkc-node-6.onrender.com',
  'https://bkc-node-7.onrender.com',
  'https://bkc-node-8.onrender.com',
  'https://bkc-node-9.onrender.com',
  'https://bkc-node-10.onrender.com',
  'https://bkc-node-11.onrender.com',
  'https://bkc-node-12.onrender.com',
  'https://bkc-node-13.onrender.com',
  'https://bkc-node-14.onrender.com',
  'https://bkc-node-15.onrender.com',
  'https://bkc-node-16.onrender.com',
  'https://bkc-node-17.onrender.com',
  'https://bkc-node-18.onrender.com',
  'https://bkc-node-19.onrender.com',
  'https://bkc-node-20.onrender.com',
  'https://bkc-node-21.onrender.com',
  'https://bkc-node-22.onrender.com',
  'https://bkc-node-23.onrender.com',
  'https://bkc-node-24.onrender.com',
  'https://bkc-node-25.onrender.com'
];

class LoadBalancer {
  constructor() {
    this.nodes = [...RENDER_NODES];
    this.currentIndex = 0;
    this.healthStatus = new Map();
    this.lastHealthCheck = 0;
    this.healthCheckInterval = 30000; // 30 seconds
    
    // Initialize all nodes as healthy
    this.nodes.forEach(node => {
      this.healthStatus.set(node, {
        healthy: true,
        lastCheck: Date.now(),
        responseTime: 0,
        errorCount: 0
      });
    });

    // Select initial node
    this.selectInitialNode();
    
    // Start health monitoring
    this.startHealthMonitoring();
  }

  /**
   * Select initial node using round-robin with random start
   */
  selectInitialNode() {
    // Random start to distribute load evenly across restarts
    this.currentIndex = Math.floor(Math.random() * this.nodes.length);
    this.currentNode = this.nodes[this.currentIndex];
  }

  /**
   * Get current active node
   */
  getCurrentNode() {
    return this.currentNode;
  }

  /**
   * Get next healthy node (round-robin)
   */
  getNextNode() {
    const healthyNodes = this.nodes.filter(node => 
      this.healthStatus.get(node)?.healthy !== false
    );

    if (healthyNodes.length === 0) {
      // All nodes are unhealthy, return current node anyway
      console.warn('All nodes are unhealthy, using current node');
      return this.currentNode;
    }

    // Find current index in healthy nodes
    const currentHealthyIndex = healthyNodes.indexOf(this.currentNode);
    const nextIndex = (currentHealthyIndex + 1) % healthyNodes.length;
    
    this.currentNode = healthyNodes[nextIndex];
    this.currentIndex = this.nodes.indexOf(this.currentNode);
    
    return this.currentNode;
  }

  /**
   * Mark node as unhealthy
   */
  markNodeUnhealthy(node) {
    const status = this.healthStatus.get(node);
    if (status) {
      status.healthy = false;
      status.errorCount++;
      status.lastCheck = Date.now();
      
      console.warn(`Node ${node} marked as unhealthy (error count: ${status.errorCount})`);
      
      // Switch to next healthy node
      if (node === this.currentNode) {
        this.getNextNode();
      }
    }
  }

  /**
   * Mark node as healthy
   */
  markNodeHealthy(node, responseTime = 0) {
    const status = this.healthStatus.get(node);
    if (status) {
      status.healthy = true;
      status.responseTime = responseTime;
      status.lastCheck = Date.now();
    }
  }

  /**
   * Get API base URL with automatic failover
   */
  async getApiUrl(path = '') {
    const maxRetries = 3;
    let lastError;

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      const node = this.getCurrentNode();
      const url = `${node}${path}`;
      
      try {
        // Quick health check
        const startTime = Date.now();
        const response = await fetch(`${node}/health`, {
          method: 'GET',
          timeout: 3000
        });
        
        const responseTime = Date.now() - startTime;
        
        if (response.ok) {
          this.markNodeHealthy(node, responseTime);
          return url;
        } else {
          throw new Error(`Health check failed: ${response.status}`);
        }
      } catch (error) {
        lastError = error;
        this.markNodeUnhealthy(node);
        
        if (attempt < maxRetries - 1) {
          // Try next node
          this.getNextNode();
          console.warn(`Node ${node} failed, trying next node...`);
        }
      }
    }

    throw new Error(`All nodes failed after ${maxRetries} attempts. Last error: ${lastError.message}`);
  }

  /**
   * Make API request with automatic failover
   */
  async apiRequest(path, options = {}) {
    const maxRetries = 2;
    let lastError;

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        const url = await this.getApiUrl(path);
        
        const response = await fetch(url, {
          ...options,
          headers: {
            'Content-Type': 'application/json',
            'X-Client-Node': this.currentIndex.toString(),
            ...options.headers
          }
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        return await response.json();
      } catch (error) {
        lastError = error;
        
        if (attempt < maxRetries - 1) {
          // Try next node for next attempt
          this.getNextNode();
          console.warn(`API request failed, retrying with next node...`);
        }
      }
    }

    throw new Error(`API request failed after ${maxRetries} attempts. Last error: ${lastError.message}`);
  }

  /**
   * Start health monitoring for all nodes
   */
  startHealthMonitoring() {
    setInterval(async () => {
      await this.checkAllNodesHealth();
    }, this.healthCheckInterval);
  }

  /**
   * Check health of all nodes
   */
  async checkAllNodesHealth() {
    const healthPromises = this.nodes.map(async (node) => {
      try {
        const startTime = Date.now();
        const response = await fetch(`${node}/health`, {
          method: 'GET',
          timeout: 5000
        });
        
        const responseTime = Date.now() - startTime;
        
        if (response.ok) {
          this.markNodeHealthy(node, responseTime);
        } else {
          this.markNodeUnhealthy(node);
        }
      } catch (error) {
        this.markNodeUnhealthy(node);
      }
    });

    await Promise.allSettled(healthPromises);
    
    // Log health status
    const healthyCount = Array.from(this.healthStatus.values())
      .filter(status => status.healthy).length;
    
    console.log(`Node health status: ${healthyCount}/${this.nodes.length} nodes healthy`);
  }

  /**
   * Get load balancer statistics
   */
  getStats() {
    const stats = {
      totalNodes: this.nodes.length,
      healthyNodes: 0,
      currentNode: this.currentNode,
      currentIndex: this.currentIndex,
      nodes: []
    };

    this.healthStatus.forEach((status, node) => {
      if (status.healthy) stats.healthyNodes++;
      
      stats.nodes.push({
        url: node,
        healthy: status.healthy,
        responseTime: status.responseTime,
        errorCount: status.errorCount,
        lastCheck: status.lastCheck
      });
    });

    return stats;
  }

  /**
   * Reset error counts (useful for maintenance)
   */
  resetErrorCounts() {
    this.healthStatus.forEach((status) => {
      status.errorCount = 0;
      if (!status.healthy) {
        status.healthy = true; // Give nodes another chance
      }
    });
  }
}

// Create singleton instance
const loadBalancer = new LoadBalancer();

export default loadBalancer;

// Export for testing
export { LoadBalancer };
