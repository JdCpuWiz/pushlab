#!/bin/bash
# PushLab update script - Pull latest changes and restart services

set -e

echo "🔄 Updating PushLab..."
echo "====================="
echo ""

# Pull latest changes
echo "📥 Pulling latest changes from GitHub..."
git pull origin main

# Rebuild Docker images if backend changed
if git diff --name-only HEAD@{1} HEAD | grep -q "^backend/"; then
    echo "🔨 Backend changes detected, rebuilding Docker images..."
    cd docker
    docker-compose build
    docker-compose up -d
    echo "✅ Services restarted with new images"
else
    echo "ℹ️  No backend changes, skipping rebuild"
fi

# Update Go dependencies if needed
if git diff --name-only HEAD@{1} HEAD | grep -q "go.mod\|go.sum"; then
    echo "📦 Updating Go dependencies..."
    cd backend
    go mod download
    go mod tidy
fi

echo ""
echo "✅ Update complete!"
echo ""
echo "📊 Check status: cd docker && docker-compose ps"
echo "📝 View logs:    cd docker && docker-compose logs -f"
