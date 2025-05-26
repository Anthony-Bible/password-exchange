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
```

### Slackbot
To install our Slackbot go to [https://api.password.exchange/slack/install](https://api.password.exchange/slack/install). If you have set up your own version of this app, you can go to `https://yoursite.com/slack/install`.

Once installed to your organization, you can use the `/encrypt` command which will send the text to the bot and the bot will send a link to access the unencrypted text. 

**NOTE:** Slackbot relies on the database and encryption services and deployments. You can remove the website deployment/service from the yaml if you only intend to deploy the slackbot.

---

## Architecture

Password Exchange uses a **microservices architecture** with **hexagonal (ports and adapters) pattern**:

### Core Services
- **Web Service**: Gin-based HTTP server serving frontend and REST API
- **Database Service**: gRPC service for all database operations
- **Encryption Service**: gRPC service handling encryption/decryption
- **Email Service**: RabbitMQ consumer for sending notification emails
- **Slackbot**: Python Flask app with Slack integration

### Key Technologies
- **Go 1.23+**: Main application with Cobra CLI, Gin web framework, gRPC services
- **Python**: Slackbot using Flask, Slack Bolt, SQLAlchemy
- **Protocol Buffers**: Service definitions generate Go and Python clients
- **RabbitMQ**: Message queue for email notifications
- **MySQL/MariaDB**: Primary database for encrypted content and OAuth tokens
- **Kubernetes**: Container orchestration

### Communication Flow
1. Web service receives password submission → calls encryption service via gRPC
2. Encryption service generates unique ID and key → stores in database service
3. Email service sends notifications via RabbitMQ (if enabled)
4. Recipient accesses unique URL with decryption key
5. Slackbot provides same functionality within Slack workspaces

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
- **Automatic expiration**: Messages expire after 7 days by default
- **Configurable view limits**: Set maximum number of times a message can be viewed
- **Multiple interfaces**: Web UI, REST API, and Slack bot
- **Email notifications**: Optional email alerts when messages are sent
- **Passphrase protection**: Additional security layer with optional passphrases
- **Rate limiting**: Built-in protection against abuse
- **Prometheus metrics**: Monitoring and observability support
- **Swagger documentation**: Complete API documentation at `/api/v1/docs`

### 🚧 Planned Features
1. Send message to both users
2. Remind users after a day of not opening
3. Email/page visit notifications
4. Configurable expiration times
5. User-generated passwords
6. **Client-side encryption**: End-to-end encryption option (would limit bot integrations)

### 🔮 Future Integrations
- Bitwarden integration
- Google Drive file sharing
- Salesforce integration
- LastPass integration
- PGP email integration

---


## Building from Source

### Prerequisites
- **Go 1.23+**: Main application language
- **Python 3.8+**: For slackbot and protobuf generation
- **Docker**: For containerized builds
- **protoc**: Protocol buffer compiler
- **grpcio-tools**: Python protobuf tools (`pip install grpcio-tools`)

### Quick Build
```bash
# Complete build verification (recommended)
./test-build.sh
```

This script will:
- Generate protobuf files for Go and Python
- Build the Go application
- Generate Swagger documentation
- Build Docker images for main app and slackbot
- Generate Kubernetes manifests with proper variable substitution

### Manual Build Steps
```bash
# Build Go application only
cd app && go mod tidy && go build -o app

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



---
## Running

### Prerequisites
Currently we only support Kubernetes. If you don't have a Kubernetes cluster, you have two options:
- **Docker Desktop**: [Enable local Kubernetes](https://docs.docker.com/desktop/kubernetes/)
- **Minikube**: [Set up Kubernetes locally](https://minikube.sigs.k8s.io/docs/start/)

### Deployment Steps
1. **Database Setup**
   - Download the MySQL file from the root of the project
   - Update password in create user statements
   - Import the MySQL schema: `mysql -u user -p < passwordexchange.sql`

2. **Configuration**
   - Edit `kubernetes/secrets.yaml` with your information
   - [View available options](https://github.com/Anthony-Bible/password-exchange/wiki/Environment-Variables)

3. **Deploy**
   - Download the latest manifest from releases
   - Apply to cluster: `kubectl apply -f password-exchange.yaml`
   - Or use the generated `combined.yaml` from `./test-build.sh` 
