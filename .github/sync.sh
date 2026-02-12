#!/bin/bash
# Git Sync Helper - Automatically sync changes to GitHub

set -e

REPO_DIR="/home/shad/pushlab"
BRANCH="main"

cd "$REPO_DIR"

# Check if there are any changes
if [[ -z $(git status -s) ]]; then
    echo "✅ No changes to sync"
    exit 0
fi

echo "📦 Changes detected, syncing to GitHub..."

# Stage all changes
git add .

# Commit with timestamp
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
git commit -m "Auto-sync: $TIMESTAMP

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

# Push to GitHub
git push origin "$BRANCH"

echo "✅ Successfully synced to GitHub!"
