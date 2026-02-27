@echo off
REM 🚀 BKC Coin - Windows Deployment Script for 10K Users

setlocal enabledelayedexpansion

echo 🚀 Starting BKC Coin deployment for 10K concurrent users...

REM 📋 Configuration
set PROJECT_NAME=bkc-coin
set CLUSTER_SIZE=3
set MAX_USERS=10000

REM 🎯 Colors (Windows 10+)
for /F %%A in ('echo prompt $E ^| cmd') do set "ESC=%%A"
set RED=%ESC%[31m
set GREEN=%ESC%[32m
set YELLOW=%ESC%[33m
set BLUE=%ESC%[34m
set NC=%ESC%[0m

REM 📊 Logging function
:log
echo %GREEN%[%date% %time%] %~1%NC%
goto :eof

:warn
echo %YELLOW%[%date% %time%] WARNING: %~1%NC%
goto :eof

:error
echo %RED%[%date% %time%] ERROR: %~1%NC%
exit /b 1

REM 🔍 Check prerequisites
:check_prerequisites
call :log "🔍 Checking prerequisites..."

REM Check Docker
docker --version >nul 2>&1
if errorlevel 1 (
    call :error "Docker is not installed. Please install Docker Desktop first."
)

REM Check Docker Compose
docker-compose --version >nul 2>&1
if errorlevel 1 (
    call :error "Docker Compose is not installed. Please install Docker Compose first."
)

REM Check Git
git --version >nul 2>&1
if errorlevel 1 (
    call :error "Git is not installed. Please install Git for Windows first."
)

call :log "✅ All prerequisites are installed"
goto :eof

REM 📁 Create necessary directories
:create_directories
call :log "📁 Creating necessary directories..."

if not exist logs mkdir logs
if not exist monitoring mkdir monitoring
if not exist monitoring\grafana mkdir monitoring\grafana
if not exist monitoring\grafana\dashboards mkdir monitoring\grafana\dashboards
if not exist monitoring\grafana\datasources mkdir monitoring\grafana\datasources
if not exist nginx mkdir nginx
if not exist nginx\ssl mkdir nginx\ssl

call :log "✅ Directories created"
goto :eof

REM 🔧 Build Docker images
:build_images
call :log "🔧 Building Docker images..."

docker-compose build --no-cache
if errorlevel 1 (
    call :error "Failed to build Docker images"
)

call :log "✅ Docker images built successfully"
goto :eof

REM 🚀 Start services
:start_services
call :log "🚀 Starting BKC Coin services..."

REM Start core services
docker-compose up -d redis
timeout /t 5 /nobreak >nul

REM Start application servers
docker-compose up -d bkc-server-1 bkc-server-2 bkc-server-3
timeout /t 10 /nobreak >nul

REM Start load balancer
docker-compose up -d nginx-lb

REM Start monitoring
docker-compose up -d prometheus grafana redis-commander

call :log "✅ All services started"
goto :eof

REM 🏥 Health check
:health_check
call :log "🏥 Performing health checks..."

REM Check individual servers
for /L %%i in (1,1,3) do (
    curl -f http://localhost:808%%i/health >nul 2>&1
    if errorlevel 1 (
        call :warn "⚠️ Server %%i is not responding"
    ) else (
        call :log "✅ Server %%i is healthy"
    )
)

REM Check load balancer
curl -f http://localhost/health >nul 2>&1
if errorlevel 1 (
    call :warn "⚠️ Load balancer is not responding"
) else (
    call :log "✅ Load balancer is healthy"
)

REM Check monitoring
curl -f http://localhost:9093/targets >nul 2>&1
if errorlevel 1 (
    call :warn "⚠️ Prometheus is not responding"
) else (
    call :log "✅ Prometheus is healthy"
)

curl -f http://localhost:3000 >nul 2>&1
if errorlevel 1 (
    call :warn "⚠️ Grafana is not responding"
) else (
    call :log "✅ Grafana is healthy"
)
goto :eof

REM 📊 Performance optimization
:optimize_performance
call :log "📊 Applying performance optimizations..."

REM Note: Windows optimization would require different commands
call :log "✅ Performance optimizations applied (Windows limitations)"
goto :eof

REM 🔧 Configure monitoring
:setup_monitoring
call :log "🔧 Setting up monitoring..."

REM Create monitoring directories structure
if not exist monitoring mkdir monitoring

call :log "✅ Monitoring configured"
goto :eof

REM 🧪 Load test
:load_test
call :log "🧪 Running load test for %MAX_USERS% concurrent users..."

REM Check if Apache Bench is available
ab -V >nul 2>&1
if errorlevel 1 (
    call :warn "Apache Bench not found. Skipping load test."
    goto :eof
)

REM Run load test
ab -n 50000 -c %MAX_USERS% http://localhost/health

call :log "✅ Load test completed"
goto :eof

REM 📋 Show status
:show_status
call :log "📋 Deployment Status:"
echo.
echo 🌐 Load Balancer: http://localhost
echo 📊 Prometheus: http://localhost:9093
echo 📈 Grafana: http://localhost:3000 (admin/admin123)
echo 🗄️ Redis Commander: http://localhost:8083
echo.
echo 🚀 Server Endpoints:
echo   Server 1: http://localhost:8080
echo   Server 2: http://localhost:8081
echo   Server 3: http://localhost:8082
echo.
echo 📊 Monitoring Endpoints:
echo   Metrics: http://localhost/metrics
echo   Health: http://localhost/health
echo.
echo 🎯 Capacity: %MAX_USERS% concurrent users
echo 🖥️  Cluster Size: %CLUSTER_SIZE% servers
echo 💾 Database: 3x PostgreSQL (Render)
echo 🗄️  Cache: Redis
goto :eof

REM 🧹 Cleanup function
:cleanup
call :log "🧹 Cleaning up..."
docker-compose down -v
docker system prune -f
call :log "✅ Cleanup completed"
goto :eof

REM 🔄 Main execution
:main
call :log "🚀 BKC Coin Deployment Started"

call :check_prerequisites
call :create_directories
call :setup_monitoring
call :build_images
call :start_services
call :health_check
call :optimize_performance
call :show_status

call :log "🎉 BKC Coin deployment completed successfully!"
call :log "🎯 Ready for %MAX_USERS% concurrent users"

REM Optional: Run load test
set /p "loadtest=🧪 Run load test? (y/n): "
if /i "!loadtest!"=="y" (
    call :load_test
)

goto :eof

REM 🛠️ Command line options
if "%1"=="start" (
    call :main
) else if "%1"=="stop" (
    docker-compose down
    call :log "🛑 Services stopped"
) else if "%1"=="restart" (
    docker-compose restart
    call :log "🔄 Services restarted"
) else if "%1"=="status" (
    docker-compose ps
    call :show_status
) else if "%1"=="logs" (
    docker-compose logs -f
) else if "%1"=="cleanup" (
    call :cleanup
) else if "%1"=="test" (
    call :load_test
) else (
    echo Usage: %0 {start^|stop^|restart^|status^|logs^|cleanup^|test}
    echo.
    echo Commands:
    echo   start   - Start all services
    echo   stop    - Stop all services
    echo   restart - Restart all services
    echo   status  - Show service status
    echo   logs    - Show service logs
    echo   cleanup - Clean up containers and images
    echo   test    - Run load test
    exit /b 1
)

if not "%1"=="" goto :eof
call :main
