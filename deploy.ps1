# PowerShell deployment script for Docker

Write-Host "🐳 Deploying Connect Four Game with Docker..." -ForegroundColor Cyan

# Check if Docker is running
try {
    docker info | Out-Null
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Stop existing containers
Write-Host "🛑 Stopping existing containers..." -ForegroundColor Yellow
docker-compose -f docker-compose.prod.yml down

# Build and start services
Write-Host "🔨 Building and starting services..." -ForegroundColor Yellow
docker-compose -f docker-compose.prod.yml up -d --build

# Wait for services to be healthy
Write-Host "⏳ Waiting for services to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check status
Write-Host "📊 Service status:" -ForegroundColor Green
docker-compose -f docker-compose.prod.yml ps

# Test backend health
Write-Host "🏥 Testing backend health..." -ForegroundColor Yellow
Start-Sleep -Seconds 3
try {
    $response = Invoke-WebRequest -Uri "http://localhost:3001/api/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "✅ Backend is healthy!" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Backend might still be starting..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "✅ Deployment complete!" -ForegroundColor Green
Write-Host "📝 Backend: http://localhost:3001" -ForegroundColor Cyan
Write-Host "📝 API Health: http://localhost:3001/api/health" -ForegroundColor Cyan
Write-Host "📝 WebSocket: ws://localhost:3001/ws" -ForegroundColor Cyan
Write-Host ""
Write-Host "To view logs: docker-compose -f docker-compose.prod.yml logs -f" -ForegroundColor Gray
Write-Host "To stop: docker-compose -f docker-compose.prod.yml down" -ForegroundColor Gray


