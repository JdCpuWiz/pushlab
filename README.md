# PushLab - Self-Hosted iOS Push Notification Service

A self-hosted push notification service for homelab users to send notifications to iOS devices via Apple Push Notification service (APNs). Built with Go and designed for easy deployment on Ubuntu servers via Docker.

## Features

- 🔐 **Multi-user support** with JWT authentication and API keys
- 📱 **Device management** with flexible tag-based routing
- 🎯 **Tag-based notifications** - send to specific device groups
- 🚀 **RESTful API** for easy integration with automation tools
- 📊 **Delivery tracking** with detailed status reporting
- ⚡ **Async processing** with RabbitMQ for reliable delivery
- 🔄 **Automatic retries** with exponential backoff
- 🐳 **Docker deployment** with docker-compose
- 📝 **Comprehensive logging** and health checks

## Architecture

```
┌─────────────┐
│ iOS Client  │ (Swift app)
└─────┬───────┘
      │ HTTPS
      │
┌─────▼─────────────────┐
│   Backend API (Go)    │
│  - REST endpoints     │
│  - JWT auth           │
│  - Device mgmt        │
└───┬───────────────────┘
    │
┌───┴──┬──────┬─────────┐
│ PG   │ Redis│ RabbitMQ│
└──────┴──────┴────┬────┘
                   │
           ┌───────▼───────┐
           │ Worker Service│
           │ - APNs sender │
           │ - Retries     │
           └───────┬───────┘
                   │ HTTP/2
           ┌───────▼───────┐
           │  Apple APNs   │
           └───────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Apple Developer Account
- APNs authentication key (.p8 file)
- iOS device for testing

### 1. Clone and Setup

```bash
git clone https://github.com/JdCpuWiz/pushlab
cd pushlab

# Copy environment file and configure
cp .env.example .env
nano .env  # Edit with your values
```

### 2. Configure Environment Variables

Edit `.env` file:

```env
DB_PASSWORD=your_secure_password
RABBITMQ_USER=guest
RABBITMQ_PASS=guest
REDIS_PASSWORD=your_redis_password
JWT_SECRET=your_secure_jwt_secret_at_least_32_characters_long
```

### 3. Start Services

```bash
cd docker
docker-compose up -d
```

This will start:
- PostgreSQL (port 5432)
- RabbitMQ (port 5672, management UI: 15672)
- Redis (port 6379)
- API server (port 8080)
- Worker service (2 replicas)

### 4. Verify Services

```bash
# Check health
curl http://localhost:8080/health

# View logs
docker-compose logs -f api
docker-compose logs -f worker
```

## API Usage

### Authentication

#### Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "john",
    "email": "john@example.com",
    "api_key": "generated_api_key",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "securepassword123"
  }'
```

### Upload APNs Credentials

Before sending notifications, upload your APNs authentication key:

```bash
curl -X POST http://localhost:8080/api/v1/credentials/apns \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "team_id": "YOUR_TEAM_ID",
    "key_id": "YOUR_KEY_ID",
    "bundle_id": "com.example.app",
    "environment": "production",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
  }'
```

### Device Registration

Register an iOS device (typically done by the iOS app):

```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "device_name": "iPhone 15 Pro",
    "device_identifier": "unique-device-id",
    "device_token": "apns-device-token-from-ios",
    "bundle_id": "com.example.app",
    "environment": "production",
    "tags": ["personal", "critical"]
  }'
```

### Send Notifications

#### Send to Devices with Specific Tags

```bash
curl -X POST http://localhost:8080/api/v1/notify \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Server Alert",
    "body": "CPU usage above 90%",
    "tags": ["server", "critical"],
    "badge": 1,
    "sound": "default",
    "priority": "high",
    "data": {
      "server": "web-01",
      "cpu": "92%"
    }
  }'
```

#### Send to All Devices

```bash
curl -X POST http://localhost:8080/api/v1/notify \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "General Notification",
    "body": "This goes to all your devices",
    "sound": "default"
  }'
```

#### Send to Specific Device

```bash
curl -X POST http://localhost:8080/api/v1/notify/device/{device_id} \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Direct Message",
    "body": "This is sent to one device"
  }'
```

#### Using API Key (for automation)

```bash
curl -X POST http://localhost:8080/api/v1/notify \
  -H "X-API-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Automation Alert",
    "body": "Backup completed successfully",
    "tags": ["automation"]
  }'
```

### View Notification History

```bash
# List notifications
curl http://localhost:8080/api/v1/notifications?limit=10 \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get notification details with delivery status
curl http://localhost:8080/api/v1/notifications/{notification_id} \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## iOS Client App

The iOS client app (in `ios-client/`) provides:

- User authentication
- Automatic APNs registration
- Device token management
- Tag configuration
- Notification history

### Building the iOS App

1. Open `ios-client/PushLab.xcodeproj` in Xcode
2. Configure your Team ID and Bundle ID
3. Enable Push Notifications capability
4. Build and run on your device

### APNs Certificate Setup

1. Go to [Apple Developer Portal](https://developer.apple.com)
2. Create an App ID with Push Notifications enabled
3. Create an APNs Authentication Key (.p8 file)
4. Note your Team ID and Key ID
5. Upload the key via the API (see above)

## Configuration

The main configuration file is `config/config.yaml`:

```yaml
server:
  api_port: 8080
  worker_count: 4

database:
  host: postgres
  port: 5432
  user: pushlab
  password: ${DB_PASSWORD}

rabbitmq:
  url: amqp://guest:guest@rabbitmq:5672/
  queue_name: notifications

jwt:
  secret: ${JWT_SECRET}
  expiry_hours: 24

apns:
  default_environment: production
```

Environment variables are expanded using `${VAR}` syntax.

## Database Migrations

The database schema is automatically initialized when PostgreSQL starts using the migration file in `migrations/001_initial_schema.sql`.

For manual migration:

```bash
docker exec -i pushlab-postgres psql -U pushlab -d pushlab < migrations/001_initial_schema.sql
```

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### RabbitMQ Management UI

Access at http://localhost:15672 (default credentials: guest/guest)

### Logs

```bash
# API logs
docker-compose logs -f api

# Worker logs
docker-compose logs -f worker

# All services
docker-compose logs -f
```

## Integration Examples

### Home Assistant

```yaml
rest_command:
  notify_pushlab:
    url: http://localhost:8080/api/v1/notify
    method: POST
    headers:
      X-API-Key: "your-api-key"
      Content-Type: "application/json"
    payload: >
      {
        "title": "{{ title }}",
        "body": "{{ message }}",
        "tags": ["homeassistant"]
      }
```

### Bash Script

```bash
#!/bin/bash
API_KEY="your-api-key"
API_URL="http://localhost:8080/api/v1/notify"

send_notification() {
  local title="$1"
  local body="$2"

  curl -X POST "$API_URL" \
    -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"$title\",\"body\":\"$body\",\"tags\":[\"automation"]}"
}

# Usage
send_notification "Backup Complete" "Server backup finished successfully"
```

### Python

```python
import requests

class PushLabClient:
    def __init__(self, api_key, base_url="http://localhost:8080"):
        self.api_key = api_key
        self.base_url = base_url

    def send_notification(self, title, body, tags=None, **kwargs):
        headers = {
            "X-API-Key": self.api_key,
            "Content-Type": "application/json"
        }

        payload = {
            "title": title,
            "body": body,
            "tags": tags or []
        }
        payload.update(kwargs)

        response = requests.post(
            f"{self.base_url}/api/v1/notify",
            headers=headers,
            json=payload
        )
        return response.json()

# Usage
client = PushLabClient("your-api-key")
client.send_notification(
    "Server Alert",
    "CPU usage is high",
    tags=["server", "critical"],
    priority="high"
)
```

## Tag-Based Routing

Tags allow flexible notification routing:

```bash
# Device with tags: ["personal", "critical", "iphone"]
# Device with tags: ["work", "iphone"]
# Device with tags: ["personal", "ipad"]

# Send to all personal devices
curl -X POST .../notify -d '{"body":"...", "tags":["personal"]}'

# Send to critical iPhones
curl -X POST .../notify -d '{"body":"...", "tags":["critical","iphone"]}'
```

Devices receive notifications if they have **all** specified tags (AND logic).

## Troubleshooting

### Notifications Not Delivering

1. Check worker logs: `docker-compose logs worker`
2. Verify APNs credentials are correct
3. Ensure device token is valid
4. Check RabbitMQ queue depth
5. Verify iOS app has notification permissions

### Database Connection Issues

```bash
# Check database status
docker-compose ps postgres

# Test connection
docker exec pushlab-postgres pg_isready -U pushlab
```

### Worker Not Processing

```bash
# Check RabbitMQ
docker-compose ps rabbitmq

# Check queue status
curl -u guest:guest http://localhost:15672/api/queues
```

## Security Considerations

1. **Change default passwords** in `.env` file
2. **Use strong JWT secret** (minimum 32 characters)
3. **Enable SSL/TLS** for production (use reverse proxy)
4. **Restrict network access** using firewall rules
5. **Regular backups** of PostgreSQL database
6. **Rotate API keys** periodically
7. **Monitor logs** for suspicious activity

## Production Deployment

For production use:

1. Use a reverse proxy (nginx/Caddy) with SSL/TLS
2. Set up proper firewall rules
3. Enable database backups
4. Use secrets management (Vault, etc.)
5. Monitor resource usage
6. Set up log aggregation
7. Configure rate limiting

## Development

### Local Development

```bash
# Start dependencies only
docker-compose up -d postgres rabbitmq redis

# Run API locally
cd backend
export CONFIG_PATH=../config/config.yaml
export DB_PASSWORD=pushlab_password
export JWT_SECRET=your_secret
go run cmd/api/main.go

# Run worker locally
go run cmd/worker/main.go
```

### Running Tests

```bash
cd backend
go test ./...
```

## API Reference

Full API documentation:

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/apikey` - Generate API key
- `POST /api/v1/devices` - Register device
- `GET /api/v1/devices` - List devices
- `PUT /api/v1/devices/{id}` - Update device
- `DELETE /api/v1/devices/{id}` - Delete device
- `POST /api/v1/notify` - Send notification
- `POST /api/v1/notify/device/{id}` - Send to device
- `GET /api/v1/notifications` - List notifications
- `GET /api/v1/notifications/{id}` - Get notification details
- `POST /api/v1/credentials/apns` - Upload APNs credentials
- `GET /api/v1/credentials/apns` - List credentials
- `DELETE /api/v1/credentials/apns/{id}` - Delete credentials
- `GET /health` - Health check

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- GitHub Issues: [Report bugs](https://github.com/yourusername/pushlab/issues)
- Documentation: [Wiki](https://github.com/yourusername/pushlab/wiki)

## Acknowledgments

- Built with [apns2](https://github.com/sideshow/apns2) library
- Inspired by [Apprise](https://github.com/caronc/apprise)
- Thanks to the Go and homelab communities

## Roadmap

- [ ] Web dashboard for device management
- [ ] Scheduled notifications
- [ ] Notification templates
- [ ] Rate limiting per user
- [ ] Webhook callbacks for delivery status
- [ ] Multi-channel support (email, SMS)
- [ ] Analytics dashboard
- [ ] Silent notifications support
- [ ] Notification grouping
- [ ] Localization support
