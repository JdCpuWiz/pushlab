#!/bin/bash
# Setup automatic git sync with cron

SCRIPT_PATH="/home/shad/pushlab/.github/sync.sh"
CRON_CMD="*/15 * * * * $SCRIPT_PATH >> /var/log/pushlab-sync.log 2>&1"

echo "Setting up automatic git sync..."

# Check if cron job already exists
if ! crontab -l 2>/dev/null | grep -q "$SCRIPT_PATH"; then
    # Add cron job (sync every 15 minutes)
    (crontab -l 2>/dev/null; echo "$CRON_CMD") | crontab -
    echo "✅ Cron job added - will sync every 15 minutes"
else
    echo "ℹ️  Cron job already exists"
fi

echo ""
echo "Manual sync: $SCRIPT_PATH"
echo "Check sync log: tail -f /var/log/pushlab-sync.log"
echo "Remove sync: crontab -e (then delete the line)"
