#!/bin/bash

# Quick deployment script for Docker

echo "🐳 Deploying Connect Four Game with Docker..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop."
    exit 1
fi

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down

# Build and start services
echo "🔨 Building and starting services..."
docker-compose -f docker-compose.prod.yml up -d --build

# Wait for services to be healthy
echo "⏳ Waiting for services to start..."
sleep 5

# Check status
echo "📊 Service status:"
docker-compose -f docker-compose.prod.yml ps

# Test backend health
echo "🏥 Testing backend health..."
sleep 3
curl -f http://localhost:3001/api/health && echo "✅ Backend is healthy!" || echo "⚠️  Backend might still be starting..."

echo ""
echo "✅ Deployment complete!"
echo "📝 Backend: http://localhost:3001"
echo "📝 API Health: http://localhost:3001/api/health"
echo "📝 WebSocket: ws://localhost:3001/ws"
echo ""
echo "To view logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "To stop: docker-compose -f docker-compose.prod.yml down"


