# PushLab Quick Start Guide

Get PushLab up and running in 5 minutes!

## Prerequisites

- Docker and Docker Compose installed
- iOS device (for testing)
- Apple Developer Account (for APNs)

## Step 1: Clone and Configure

```bash
git clone https://github.com/yourusername/pushlab
cd pushlab

# Copy and edit environment file
cp .env.example .env
nano .env  # Set your passwords and JWT secret
```

**Important:** Set a strong JWT_SECRET (at least 32 characters)!

## Step 2: Start Services

```bash
cd docker
docker-compose up -d
```

Wait about 30 seconds for all services to start.

## Step 3: Verify Installation

```bash
# Check health
curl http://localhost:8080/health

# Should return: {"status":"ok","database":"ok"}
```

## Step 4: Create Your First User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "securepassword123"
  }'
```

Save the `token` and `api_key` from the response!

## Step 5: Upload APNs Credentials

You'll need your Apple Developer APNs key (.p8 file):

```bash
# Read your .p8 file
PRIVATE_KEY=$(cat /path/to/AuthKey_XXXXX.p8)

# Upload to PushLab
curl -X POST http://localhost:8080/api/v1/credentials/apns \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"team_id\": \"YOUR_TEAM_ID\",
    \"key_id\": \"YOUR_KEY_ID\",
    \"bundle_id\": \"com.example.app\",
    \"environment\": \"production\",
    \"private_key\": \"$PRIVATE_KEY\"
  }"
```

## Step 6: Build iOS App

1. Open `ios-client/PushLab.xcodeproj` in Xcode
2. Update `Utils/Constants.swift` with your server IP:
   ```swift
   static let apiBaseURL = "http://YOUR_SERVER_IP:8080"
   ```
3. Configure your Team ID and Bundle ID
4. Enable Push Notifications capability
5. Build and run on your iOS device
6. Login with credentials from Step 4
7. Grant notification permissions when prompted

## Step 7: Send Your First Notification

```bash
curl -X POST http://localhost:8080/api/v1/notify \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello PushLab!",
    "body": "Your first notification",
    "sound": "default"
  }'
```

You should receive a notification on your iOS device! 🎉

## Troubleshooting

### Notifications not arriving?

1. Check worker logs: `docker-compose logs worker`
2. Verify device is registered: `curl -H "Authorization: Bearer TOKEN" http://localhost:8080/api/v1/devices`
3. Check APNs credentials are correct
4. Ensure iOS app has notification permissions

### Can't connect to server?

1. Check all services are running: `docker-compose ps`
2. View API logs: `docker-compose logs api`
3. Verify port 8080 is accessible
4. Check firewall settings

### Database issues?

```bash
# Check database status
docker exec pushlab-postgres pg_isready -U pushlab

# View database logs
docker-compose logs postgres

# Re-run migrations if needed
docker exec -i pushlab-postgres psql -U pushlab -d pushlab < ../migrations/001_initial_schema.sql
```

## Next Steps

- **Configure tags**: Add tags to your devices for targeted notifications
- **Integrate with automation**: Use the API key with Home Assistant, scripts, etc.
- **Monitor**: Access RabbitMQ UI at http://localhost:15672 (guest/guest)
- **Scale**: Add more worker replicas in docker-compose.yml
- **Secure**: Set up SSL/TLS with a reverse proxy

## Useful Commands

```bash
# View all logs
docker-compose logs -f

# Restart API server
docker-compose restart api

# Stop everything
docker-compose down

# Stop and remove data
docker-compose down -v

# Access database
docker exec -it pushlab-postgres psql -U pushlab -d pushlab
```

## Integration Examples

### Bash Script

```bash
#!/bin/bash
API_KEY="your-api-key"

notify() {
  curl -X POST http://localhost:8080/api/v1/notify \
    -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"$1\",\"body\":\"$2\"}" -s
}

# Usage
notify "Backup Complete" "Server backup finished at $(date)"
```

### Cron Job

```bash
# Add to crontab: crontab -e
0 0 * * * /usr/local/bin/notify_backup.sh
```

### Home Assistant

```yaml
# configuration.yaml
rest_command:
  pushlab_notify:
    url: http://localhost:8080/api/v1/notify
    method: POST
    headers:
      X-API-Key: "your-api-key"
      Content-Type: "application/json"
    payload: '{"title":"{{ title }}","body":"{{ message }}"}'
```

## Getting Help

- Documentation: [Full README](README.md)
- Issues: [GitHub Issues](https://github.com/yourusername/pushlab/issues)
- Logs: `docker-compose logs -f`

Happy Pushing! 🚀
