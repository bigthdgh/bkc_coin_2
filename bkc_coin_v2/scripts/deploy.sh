#!/bin/bash

# 🚀 BKC Coin - Deployment Script for 10K Users

set -e

echo "🚀 Starting BKC Coin deployment for 10K concurrent users..."

# 📋 Configuration
PROJECT_NAME="bkc-coin"
CLUSTER_SIZE=3
MAX_USERS=10000

# 🎯 Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 📊 Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# 🔍 Check prerequisites
check_prerequisites() {
    log "🔍 Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker first."
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed. Please install Docker Compose first."
    fi
    
    # Check Git
    if ! command -v git &> /dev/null; then
        error "Git is not installed. Please install Git first."
    fi
    
    log "✅ All prerequisites are installed"
}

# 📁 Create necessary directories
create_directories() {
    log "📁 Creating necessary directories..."
    
    mkdir -p logs
    mkdir -p monitoring/grafana/dashboards
    mkdir -p monitoring/grafana/datasources
    mkdir -p nginx/ssl
    
    log "✅ Directories created"
}

# 🔧 Build Docker images
build_images() {
    log "🔧 Building Docker images..."
    
    docker-compose build --no-cache
    
    log "✅ Docker images built successfully"
}

# 🚀 Start services
start_services() {
    log "🚀 Starting BKC Coin services..."
    
    # Start core services
    docker-compose up -d redis
    
    # Wait for Redis
    sleep 5
    
    # Start application servers
    docker-compose up -d bkc-server-1 bkc-server-2 bkc-server-3
    
    # Wait for servers
    sleep 10
    
    # Start load balancer
    docker-compose up -d nginx-lb
    
    # Start monitoring
    docker-compose up -d prometheus grafana redis-commander
    
    log "✅ All services started"
}

# 🏥 Health check
health_check() {
    log "🏥 Performing health checks..."
    
    # Check individual servers
    for i in {1..3}; do
        if curl -f http://localhost:808$(($i-1))/health > /dev/null 2>&1; then
            log "✅ Server $i is healthy"
        else
            warn "⚠️ Server $i is not responding"
        fi
    done
    
    # Check load balancer
    if curl -f http://localhost/health > /dev/null 2>&1; then
        log "✅ Load balancer is healthy"
    else
        warn "⚠️ Load balancer is not responding"
    fi
    
    # Check monitoring
    if curl -f http://localhost:9093/targets > /dev/null 2>&1; then
        log "✅ Prometheus is healthy"
    else
        warn "⚠️ Prometheus is not responding"
    fi
    
    if curl -f http://localhost:3000 > /dev/null 2>&1; then
        log "✅ Grafana is healthy"
    else
        warn "⚠️ Grafana is not responding"
    fi
}

# 📊 Performance optimization
optimize_performance() {
    log "📊 Applying performance optimizations..."
    
    # Set ulimits for high concurrency
    echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
    echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf
    
    # Optimize network settings
    echo "net.core.somaxconn = 65536" | sudo tee -a /etc/sysctl.conf
    echo "net.ipv4.tcp_max_syn_backlog = 65536" | sudo tee -a /etc/sysctl.conf
    echo "net.ipv4.tcp_fin_timeout = 30" | sudo tee -a /etc/sysctl.conf
    
    sudo sysctl -p
    
    log "✅ Performance optimizations applied"
}

# 🔧 Configure monitoring
setup_monitoring() {
    log "🔧 Setting up monitoring..."
    
    # Create Prometheus configuration
    cat > monitoring/prometheus.yml << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: 'bkc-servers'
    static_configs:
      - targets: 
        - 'bkc-server-1:9090'
        - 'bkc-server-2:9091'
        - 'bkc-server-3:9092'
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx-lb:80']
    metrics_path: '/nginx_status'

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']

alerting:
  alertmanagers:
    - static_configs:
        - targets: []
EOF

    log "✅ Monitoring configured"
}

# 🧪 Load test
load_test() {
    log "🧪 Running load test for $MAX_USERS concurrent users..."
    
    # Install Apache Bench if not present
    if ! command -v ab &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y apache2-utils
    fi
    
    # Run load test
    ab -n 50000 -c $MAX_USERS http://localhost/health
    
    log "✅ Load test completed"
}

# 📋 Show status
show_status() {
    log "📋 Deployment Status:"
    echo ""
    echo "🌐 Load Balancer: http://localhost"
    echo "📊 Prometheus: http://localhost:9093"
    echo "📈 Grafana: http://localhost:3000 (admin/admin123)"
    echo "🗄️ Redis Commander: http://localhost:8083"
    echo ""
    echo "🚀 Server Endpoints:"
    echo "  Server 1: http://localhost:8080"
    echo "  Server 2: http://localhost:8081"
    echo "  Server 3: http://localhost:8082"
    echo ""
    echo "📊 Monitoring Endpoints:"
    echo "  Metrics: http://localhost/metrics"
    echo "  Health: http://localhost/health"
    echo ""
    echo "🎯 Capacity: $MAX_USERS concurrent users"
    echo "🖥️  Cluster Size: $CLUSTER_SIZE servers"
    echo "💾 Database: 3x PostgreSQL (Render)"
    echo "🗄️  Cache: Redis"
}

# 🧹 Cleanup function
cleanup() {
    log "🧹 Cleaning up..."
    docker-compose down -v
    docker system prune -f
    log "✅ Cleanup completed"
}

# 🔄 Main execution
main() {
    log "🚀 BKC Coin Deployment Started"
    
    check_prerequisites
    create_directories
    setup_monitoring
    build_images
    start_services
    health_check
    optimize_performance
    show_status
    
    log "🎉 BKC Coin deployment completed successfully!"
    log "🎯 Ready for $MAX_USERS concurrent users"
    
    # Optional: Run load test
    read -p "🧪 Run load test? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        load_test
    fi
}

# 🛠️ Command line options
case "${1:-}" in
    "start")
        main
        ;;
    "stop")
        docker-compose down
        log "🛑 Services stopped"
        ;;
    "restart")
        docker-compose restart
        log "🔄 Services restarted"
        ;;
    "status")
        docker-compose ps
        show_status
        ;;
    "logs")
        docker-compose logs -f
        ;;
    "cleanup")
        cleanup
        ;;
    "test")
        load_test
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs|cleanup|test}"
        echo ""
        echo "Commands:"
        echo "  start   - Start all services"
        echo "  stop    - Stop all services"
        echo "  restart - Restart all services"
        echo "  status  - Show service status"
        echo "  logs    - Show service logs"
        echo "  cleanup - Clean up containers and images"
        echo "  test    - Run load test"
        exit 1
        ;;
esac
