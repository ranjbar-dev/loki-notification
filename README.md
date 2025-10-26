# Loki Notification Service

A Go-based microservice that intercepts Grafana Loki push requests and sends intelligent Telegram notifications for critical log events (errors, warnings, and fatal messages).

## 🚀 Features

- **Real-time Log Monitoring**: Intercepts and parses Grafana Loki push messages
- **Smart Filtering**: Automatically detects error, warning, and fatal log entries
- **Telegram Integration**: Sends formatted notifications to Telegram channels/groups
- **Channel Routing**: Route notifications to different Telegram chats based on container/service names
- **Snappy Compression Support**: Handles Loki's Snappy-compressed protobuf format
- **Docker Ready**: Fully containerized with multi-stage Docker builds
- **Nginx Mirror Support**: Works seamlessly with Nginx's `mirror` directive for non-intrusive monitoring

## 📋 Table of Contents

- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Nginx Setup](#nginx-setup)
- [Docker Deployment](#docker-deployment)
- [Development](#development)
- [API Reference](#api-reference)

## 🏗️ Architecture

```
┌─────────────┐       ┌─────────────┐       ┌─────────────────────┐
│   Grafana   │──────▶│    Nginx    │──────▶│   Loki (Primary)    │
│   Promtail  │       │   (Mirror)  │       └─────────────────────┘
└─────────────┘       └─────────────┘
                             │
                             │ Mirror
                             ▼
                      ┌─────────────────────┐       ┌─────────────┐
                      │ Loki Notification   │──────▶│  Telegram   │
                      │      Service        │       │     Bot     │
                      └─────────────────────┘       └─────────────┘
```

The service operates as a mirror endpoint, receiving a copy of all Loki push requests without affecting the primary log ingestion pipeline.

## 📦 Installation

### Prerequisites

- Go 1.24.2 or higher
- Docker & Docker Compose (for containerized deployment)
- Telegram Bot Token ([Create a bot](https://core.telegram.org/bots#6-botfather))
- Telegram Chat ID ([Get your chat ID](https://t.me/userinfobot))

### Build from Source

```bash
# Clone the repository
git clone https://github.com/ranjbar-dev/loki-notification.git
cd loki-notification

# Download dependencies
go mod download

# Build the binary
go build -o loki-notification ./cmd

# Run the service
./loki-notification -config ./config/config.yaml
```

## ⚙️ Configuration

Create a `config.yaml` file:

```yaml
app:
  name: "loki-notification"
  environment: "production"
  log_level: "info"

api:
  host: "0.0.0.0"
  port: "7777"
  cert_location: ""  # Optional: path to SSL cert
  key_location: ""   # Optional: path to SSL key

telegram:
  bot_token: "YOUR_BOT_TOKEN"
  chat_id: 0  # Default chat for unmatched logs

channels:
  - name: "Payment Service"
    needle: "payment-service"  # Match container_name or service_name
    telegram_token: "YOUR_BOT_TOKEN"
    telegram_chat_id: 0
  
  - name: "User Service"
    needle: "user-service"
    telegram_token: "YOUR_BOT_TOKEN"
    telegram_chat_id: 0
```

### Configuration Parameters

| Parameter | Description | Required |
|-----------|-------------|----------|
| `app.name` | Application name | Yes |
| `app.environment` | Environment (development/production) | Yes |
| `app.log_level` | Log level (debug/info/warn/error) | Yes |
| `api.host` | API server bind address | Yes |
| `api.port` | API server port | Yes |
| `telegram.bot_token` | Default Telegram bot token | Yes |
| `telegram.chat_id` | Default Telegram chat ID | Yes |
| `channels[].name` | Channel description | No |
| `channels[].needle` | Match string for container/service name | Yes |
| `channels[].telegram_token` | Channel-specific bot token | Yes |
| `channels[].telegram_chat_id` | Channel-specific chat ID | Yes |

## 🎯 Usage

### Running Locally

```bash
# With default config path (./config/config.yaml)
./loki-notification

# With custom config path
./loki-notification -config /path/to/config.yaml
```

### Testing the Endpoint

```bash
# Send a test Loki push request (with Snappy compression)
curl -X POST http://localhost:7777/loki/api/v1/push \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @test-data.bin
```

## 🔧 Nginx Setup

Configure Nginx to mirror Loki push requests to the notification service:

```nginx
server {
    listen 80;
    server_name loki.example.com;

    # Limit request body size
    client_max_body_size 10M;

    # Primary Loki endpoint
    location / {
        proxy_pass http://loki:3100;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Mirror request to notification service
        mirror /mirror;
        mirror_request_body on;
    }

    # Mirror endpoint (internal)
    location /mirror {
        internal;
        proxy_pass http://loki-notification:7777$request_uri;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Don't wait for mirrored response
        proxy_read_timeout 5;
        
        # Optional: separate logs for mirror
        access_log /var/log/nginx/mirror_access.log;
        error_log /var/log/nginx/mirror_error.log;
    }
}
```

## 🐳 Docker Deployment

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop the service
docker-compose down
```

### Docker Compose Configuration

```yaml
services:
  app:
    build:
      context: .
      dockerfile: deployments/app/Dockerfile
      target: production
    container_name: loki-notification
    restart: unless-stopped
    ports:
      - "7777:7777"
    volumes:
      - ./config/config.yaml:/app/config/config.yaml:ro
    environment:
      - APP_ENVIRONMENT=production
```

### Building Docker Image

```bash
# Development build
docker build -f deployments/app/Dockerfile --target dev -t loki-notification:dev .

# Production build
docker build -f deployments/app/Dockerfile --target production -t loki-notification:latest .
```

## 💻 Development

### Project Structure

```
loki-notification/
├── cmd/                    # Application entry point
│   └── main.go
├── internal/
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── httpserver/        # HTTP server implementation
│   │   ├── server.go
│   │   └── methods.go
│   └── logger/            # Logging utilities
│       └── logger.go
├── srv/                   # Service logic
│   ├── main.go           # Service struct and initialization
│   ├── routes.go         # Route registration
│   └── handlers.go       # Request handlers
├── config/
│   └── config.yaml       # Configuration file
├── deployments/
│   └── app/
│       └── Dockerfile    # Multi-stage Docker build
├── docker-compose.yml    # Docker Compose configuration
└── go.mod                # Go module dependencies
```

### Running in Development Mode

```bash
# Run with hot reload (using air or similar)
go run ./cmd -config ./config/config.yaml

# Run tests
go test ./...

# Run with verbose logging
LOG_LEVEL=debug go run ./cmd -config ./config/config.yaml
```

### Adding Dependencies

```bash
go get <package-name>
go mod tidy
```

## 📡 API Reference

### POST `/loki/api/v1/push`

Receives Grafana Loki push requests in Snappy-compressed protobuf format.

**Content-Type:** `application/x-protobuf`

**Request Body:** Snappy-compressed Loki `PushRequest` protobuf message

**Response:**
```json
{
  "message": "OK"
}
```

**Error Response:**
```json
{
  "error": "error description"
}
```

## 🔍 Log Filtering

The service automatically filters and sends notifications for log entries containing:

- `error` (case-insensitive)
- `warning` (case-insensitive)
- `fatal` (case-insensitive)

### Telegram Message Format

```
*Level:* `error`
*Container:* `payment-service`
*Service:* `api`
```
Payment processing failed: insufficient funds
```
*File:* `/var/log/docker/abc123.json`
*Host:* `server-01`
*IpAddress:* `192.168.1.100`
*Time:* `2025-10-26T10:33:28.680Z`
```

## 🛠️ Troubleshooting

### Common Issues

**Issue: "Failed to decompress snappy data"**
- Ensure the incoming request body is Snappy-compressed
- Check that Grafana/Promtail is configured to use Snappy compression

**Issue: "Failed to send telegram message: Bad Request"**
- Verify your Telegram bot token is correct
- Ensure the chat ID is valid and the bot is a member of the chat/channel
- Check that the bot has permission to send messages

**Issue: "Failed to unmarshal protobuf"**
- Verify the protobuf data format matches Loki's PushRequest schema
- Check Loki version compatibility

### Debug Logging

Enable debug logging to troubleshoot issues:

```yaml
app:
  log_level: "debug"
```

## 📄 License

This project is licensed under the MIT License.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

For issues and questions, please open an issue on GitHub.

---

**Built with ❤️ using Go and Grafana Loki**

