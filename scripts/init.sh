#!/bin/bash
# PushLab initialization script - Complete setup in one command

set -e

echo "🚀 PushLab Initialization Script"
echo "================================"
echo ""

# Check for required commands
command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed. Aborting." >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed. Aborting." >&2; exit 1; }

# Step 1: Setup environment
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env

    # Generate secure secrets
    JWT_SECRET=$(openssl rand -base64 48)
    DB_PASSWORD=$(openssl rand -base64 24)
    REDIS_PASSWORD=$(openssl rand -base64 24)

    sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|g" .env
    sed -i "s|DB_PASSWORD=.*|DB_PASSWORD=$DB_PASSWORD|g" .env
    sed -i "s|REDIS_PASSWORD=.*|REDIS_PASSWORD=$REDIS_PASSWORD|g" .env

    echo "✅ Environment file created with secure secrets"
else
    echo "ℹ️  .env file already exists"
fi

# Step 2: Create necessary directories
echo "📁 Creating directories..."
mkdir -p certs logs

# Step 3: Start services
echo "🐳 Starting Docker services..."
cd docker
docker-compose up -d

echo "⏳ Waiting for services to be healthy (30s)..."
sleep 30

# Step 4: Check health
echo "🏥 Checking service health..."
if curl -sf http://localhost:8080/health > /dev/null; then
    echo "✅ API server is healthy"
else
    echo "⚠️  API server may not be ready yet, check logs: docker-compose logs api"
fi

# Step 5: Display information
echo ""
echo "🎉 PushLab is ready!"
echo "===================="
echo ""
echo "📍 Access Points:"
echo "   API Server:       http://localhost:8080"
echo "   Health Check:     http://localhost:8080/health"
echo "   RabbitMQ UI:      http://localhost:15672 (guest/guest)"
echo ""
echo "📚 Next Steps:"
echo "   1. Create a user:  curl -X POST http://localhost:8080/api/v1/auth/register \\"
echo "                        -H 'Content-Type: application/json' \\"
echo "                        -d '{\"username\":\"admin\",\"email\":\"admin@example.com\",\"password\":\"secure123\"}'"
echo ""
echo "   2. Upload APNs credentials (see README.md)"
echo "   3. Build and install iOS app"
echo ""
echo "📖 Documentation:"
echo "   README.md       - Full documentation"
echo "   QUICKSTART.md   - Quick start guide"
echo ""
echo "🔧 Management:"
echo "   View logs:      cd docker && docker-compose logs -f"
echo "   Stop services:  cd docker && docker-compose down"
echo "   Restart:        cd docker && docker-compose restart"
