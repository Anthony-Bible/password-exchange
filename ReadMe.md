|  [Documentation](https://github.com/Anthony-Bible/password-exchange/wiki) | [API Docs](https://password.exchange/api/v1/docs/index.html) | [Building](#building-from-source) | [Running](#running)

# Password.exchange

A secure, open-source password sharing platform built with microservices architecture. Share passwords and sensitive text securely without compromising security - no phone calls or complicated GPG setup required.

🔒 **Secure by design**: Messages are encrypted and automatically expire  
🌐 **Multiple interfaces**: Web UI, REST API, and Slack integration  
⚡ **Modern architecture**: Go microservices with hexagonal architecture  
📱 **Developer-friendly**: Complete REST API with Swagger documentation  

**Visit our [wiki](https://github.com/Anthony-Bible/password-exchange/wiki) for detailed information**

---

## How it works
### [Website](https://password.exchange)

You fill out [the form](https://password.exchange) with the necessary information including both of your names and emails (optional). We use email to send the link to the content, but there is an option to disable emails. For your name(s), this is used to personalize and let the recipient know who sent them the link. There is no verification on names so you can use whatever to remain anonymous. 

**NOTE:** All messages expire after 7 days. This means you won't be able to view your message after 7 days and will have to resend it. 

### REST API

The platform provides a complete REST API for programmatic access:

📖 **Interactive Documentation**: [https://password.exchange/api/v1/docs](https://password.exchange/api/v1/docs)

**Quick API Examples:**
```bash
# Submit a message
curl -X POST https://password.exchange/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "my secret", "maxViewCount": 3}'

# Check message status
curl "https://password.exchange/api/v1/messages/{id}?key={key}"

# Decrypt message
curl -X POST https://password.exchange/api/v1/messages/{id}/decrypt \
  -H "Content-Type: application/json" \
  -d '{"decryptionKey": "{key}"}'

# Trigger a deferred email notification for an existing message
curl -X POST https://password.exchange/api/v1/messages/{id}/notify \
  -H "Content-Type: application/json" \
  -d '{"senderName": "Alice", "recipientName": "Bob", "recipientEmail": "bob@example.com"}'

# Query API capabilities and current version
curl https://password.exchange/api/v1/info
```

**Health probes** (used by Kubernetes; not rate-limited):
```bash
curl https://password.exchange/livez   # process alive
curl https://password.exchange/readyz  # upstream gRPC services reachable
```

#### File Upload API

Files can be attached to messages using a chunked, client-side-encrypted upload protocol. Files are encrypted with AES-256 before leaving the browser; the server never sees the plaintext key.

```bash
# 1. Initiate an upload session (returns fileID, sessionID, and a base64url AES-256 key)
curl -X POST https://password.exchange/api/v1/files/initiate \
  -H "Content-Type: application/json" \
  -d '{"filename": "secret.txt", "contentType": "text/plain", "totalSize": 1024, "chunkSize": 512, "messageID": "{messageID}"}'

# 2. Upload each chunk as multipart/form-data (repeat for each chunk, chunkIndex is 1-based)
curl -X POST https://password.exchange/api/v1/files/{fileID}/chunks \
  -F "sessionID={sessionID}" \
  -F "chunkIndex=1" \
  -F "totalChunks=2" \
  -F "data=@chunk1.bin"

# 3. Download — decryption key must be in the X-File-Key header (not a query param,
#    to keep it out of server access logs)
curl https://password.exchange/api/v1/files/{fileID} \
  -H "X-File-Key: {base64url-encoded-key}" \
  --output secret.txt
```

The back-end stores encrypted chunks in S3-compatible object storage (AWS S3, MinIO, etc.). See [configuration options](https://github.com/Anthony-Bible/password-exchange/wiki/Environment-Variables) for the required `objectstorage*` keys.

### Slackbot
To install our Slackbot go to [https://api.password.exchange/slack/install](https://api.password.exchange/slack/install). If you have set up your own version of this app, you can go to `https://yoursite.com/slack/install`.

Once installed to your organization, you can use the `/encrypt` command which will send the text to the bot and the bot will send a link to access the unencrypted text.

The bot also monitors messages for unencrypted passwords — if it detects a plain-text `password: ...` pattern in a channel it will nudge the user to use `/encrypt` instead.

**NOTE:** Slackbot relies on the database and encryption services and deployments. You can remove the website deployment/service from the yaml if you only intend to deploy the slackbot.

---

## Architecture

Password Exchange uses a **microservices architecture** with **hexagonal (ports and adapters) pattern**:

### Core Services
- **Web Service**: Gin-based HTTP server serving frontend and REST API
- **Database Service**: gRPC service for all database operations
- **Encryption Service**: gRPC service handling encryption/decryption
- **Email Service**: RabbitMQ consumer for sending notification emails (fully hexagonal)
- **Slackbot**: Python Flask app with Slack integration

### Key Technologies
- **Go 1.25+**: Main application with Cobra CLI, Gin web framework, gRPC services
- **Python**: Slackbot using Flask, Slack Bolt, SQLAlchemy
- **Protocol Buffers**: Service definitions generate Go and Python clients
- **RabbitMQ**: Message queue for email notifications
- **MySQL/MariaDB**: Primary database for encrypted content and OAuth tokens
- **S3-compatible object storage**: Encrypted file attachments (AWS S3, MinIO, etc.)
- **Kubernetes**: Container orchestration

### Communication Flow
1. Web service receives password submission → calls encryption service via gRPC
2. Encryption service generates unique ID and key → stores in database service
3. Email service sends notifications via RabbitMQ (if enabled)
4. Recipient accesses unique URL with decryption key
5. Slackbot provides same functionality within Slack workspaces

### Hexagonal Architecture
The codebase follows hexagonal architecture principles:
- **Domain Layer**: Pure business logic with no external dependencies
- **Ports**: Define interfaces for inbound/outbound interactions
- **Adapters**: Implement infrastructure concerns (databases, messaging, APIs)
- **Clean separation**: Business logic isolated from technical implementation details

---
### Extensions/Tools

The current tools are available:

1. [Slack Bot](https://api.password.exchange/slack/install)
2. [Password-Store](https://www.passwordstore.org/) Extension - https://github.com/Anthony-Bible/password-store-extension

_If you have a tool or extension that interacts with Password Exchange please make a PR and we'll add it_

---
## Features

### ✅ Current Features
- **Secure message sharing**: Server-side encrypted password and text sharing
- **Encrypted file attachments**: AES-256 chunked file upload with S3-compatible storage backend
- **Automatic expiration**: Messages expire after 7 days by default
- **Configurable view limits**: Set maximum number of times a message can be viewed; default is driven by server config
- **Multiple interfaces**: Web UI, REST API, and Slack bot
- **Email notifications**: Optional email alerts when messages are sent
- **Email reminders**: Automated reminders for unviewed messages with configurable timing
- **Passphrase protection**: Additional security layer with optional passphrases
- **Password generator**: Generate strong, secure passwords and passphrases directly in the web UI
- **Cloudflare Turnstile**: Optional bot-protection CAPTCHA on form submissions and API requests
- **Rate limiting**: Built-in per-endpoint protection against abuse
- **Prometheus metrics**: Monitoring and observability support
- **Swagger documentation**: Complete API documentation at `/api/v1/docs`
- **Slack pattern detection**: Bot warns users when unencrypted passwords are posted in channels

#### Password Generator
The web interface includes a comprehensive, client-side password generator to help users create strong, secure passwords and passphrases.

- **Generation Methods**: Supports both randomly generated passwords and memorable passphrases based on the [EFF Large Wordlist](https://www.eff.org/dice).
- **Secure Randomness**: Utilizes the browser's `crypto.getRandomValues` API for cryptographically secure random number generation.
- **Strength Analysis**: Integrates the `zxcvbn` library to provide real-time password strength analysis, including crack time estimations and actionable feedback.
- **Customization**:
    - **Random Passwords**: Configure length, and inclusion of uppercase, lowercase, numbers, and symbols. Also provides an option to exclude ambiguous characters (e.g., `O`, `0`, `l`, `1`).
    - **Passphrases**: Configure the number of words and the separator character.
- **User Experience**:
    - Real-time generation and strength analysis as options are changed.
    - Ability to generate multiple password options for users to choose from.
    - Easy copy-to-clipboard and insertion into the message field.

### 🚧 Planned Features
1. Send message to both users
2. Email/page visit notifications
3. Configurable expiration times
4. **Client-side encryption**: End-to-end encryption option (would limit bot integrations)

### 🔮 Future Integrations
- Bitwarden integration
- Google Drive file sharing
- Salesforce integration
- LastPass integration
- PGP email integration

---


## Building from Source

### Prerequisites
- **Go 1.25+**: Main application language
- **Python 3.8+**: For slackbot and protobuf generation
- **Docker**: For containerized builds
- **protoc**: Protocol buffer compiler
- **grpcio-tools**: Python protobuf tools (`pip install grpcio-tools`)

### Quick Build
```bash
# Complete build verification (recommended)
./test-build.sh
```

`./test-build.sh` will:
- Generate protobuf files for Go and Python
- Build the Go application (version injected via ldflags)
- Generate Swagger documentation
- Build Docker images for main app and slackbot
- Generate Kubernetes manifests with proper variable substitution

### E2E Testing

End-to-end tests live in the `e2e/` directory and use [Playwright](https://playwright.dev/). They run against a live environment (default: `https://dev.password.exchange`).

```bash
cd e2e
npm install

# Run all tests (Chromium + Firefox)
npx playwright test

# Run against a custom URL
BASE_URL=https://localhost:8080 npx playwright test

# Open the HTML report after a run
npx playwright show-report
```

### Manual Build Steps
```bash
# Build Go application only (version injected via ldflags)
cd app && go mod tidy && go build -ldflags "-X github.com/Anthony-Bible/password-exchange/app/internal/shared/buildinfo.Version=$(git describe --tags --always --dirty)" -o app

# Generate protobuf files
protoc --proto_path=protos \
       --go_out=./app --go_opt=module=github.com/Anthony-Bible/password-exchange/app \
       --go-grpc_out=./app --go-grpc_opt=module=github.com/Anthony-Bible/password-exchange/app \
       protos/*.proto

# Generate Swagger docs
cd app && swag init -g internal/domains/message/adapters/primary/api/docs.go -o docs --parseDependency

# Build Docker images
docker build -t passwordexchange .
docker build -t slackbot -f slackbot/Dockerfile .
```

### Available Commands
```bash
# Web service with frontend and REST API
./app web --config=config.yaml

# Database service (gRPC) - automatically runs migrations if enabled in config
./app database --config=config.yaml

# Database migration management
./app database migrate status --config=config.yaml
./app database migrate up --config=config.yaml
./app database migrate down --config=config.yaml
./app database migrate create my_migration --config=config.yaml

# Encryption service (gRPC)  
./app encryption --config=config.yaml

# Email service (RabbitMQ consumer)
./app email --config=config.yaml

# Send email reminders for unviewed messages
# All flags are optional; values shown are the defaults
./app reminder --config=config.yaml \
  --older-than-hours=24 \
  --max-reminders=3 \
  --interval-hours=24

# Print the current application version
./app version
```



---
## Running

### Prerequisites
Currently we only support Kubernetes. If you don't have a Kubernetes cluster, you have two options:
- **Docker Desktop**: [Enable local Kubernetes](https://docs.docker.com/desktop/kubernetes/)
- **Minikube**: [Set up Kubernetes locally](https://minikube.sigs.k8s.io/docs/start/)

### Deployment Steps
1. **Database Setup**
   - The database service handles schema initialization and migrations automatically on startup if `automigrate: true` is set in your configuration.
   - For manual setup, use the migration commands: `./app database migrate up --config=config.yaml`

2. **Configuration**
   - Edit `kubernetes/secrets.yaml` with your information
   - [View available options](https://github.com/Anthony-Bible/password-exchange/wiki/Environment-Variables)
   - To enable file attachments, supply the `objectstorage*` config keys pointing at an S3-compatible bucket
   - To enable Cloudflare Turnstile bot protection, set `turnstile_secret` in your config

3. **Deploy**
   - Download the latest manifest from releases
   - Apply to cluster: `kubectl apply -f password-exchange.yaml`
   - Or use the generated `combined.yaml` from `./test-build.sh`

The Kubernetes manifests include a daily CronJob that purges expired messages, gRPC-native liveness/readiness probes, and ArgoCD sync-wave annotations for ordered rollout.
