# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

Password Exchange is a microservices-based secure password sharing platform consisting of:

### Core Services
- **Web Service** (`app web`): Gin-based HTTP server serving frontend and REST API
- **Database Service** (`app database`): gRPC service for all database operations  
- **Encryption Service** (`app encryption`): gRPC service handling encryption/decryption
- **Email Service** (`app email`): RabbitMQ consumer for sending notification emails
- **Slackbot**: Python Flask app with Slack Bolt framework for Slack integration

### Key Technologies
- **Go**: Main application with Cobra CLI, Gin web framework, gRPC services
- **Python**: Slackbot using Flask, Slack Bolt, SQLAlchemy
- **Protocol Buffers**: Service definitions in `protos/` generate Go and Python clients
- **RabbitMQ**: Message queue for email notifications
- **MySQL/MariaDB**: Primary database for encrypted content and OAuth tokens
- **Kubernetes**: Container orchestration with manifests in `k8s/`

### Communication Flow
1. Web service receives password submission, calls encryption service via gRPC
2. Encryption service generates unique ID and encryption key, stores in database service
3. Email service (if enabled) sends notification via RabbitMQ
4. Recipient accesses unique URL with decryption key
5. Slackbot provides same functionality within Slack workspaces

## Development Commands

### Building and Testing
```bash
# Generate protobuf files and test all builds
./test-build.sh

# Build Go application only
cd app && go mod tidy && go build -o app

# Build Docker images
docker build -t passwordexchange-test .
docker build -t slackbot-test -f slackbot/Dockerfile .
```

### Running Services Locally
```bash
# Start individual services (requires config file)
./app/app web --config=config.yaml
./app/app database --config=config.yaml  
./app/app encryption --config=config.yaml
./app/app email --config=config.yaml

# Run slackbot
cd slackbot && python program.py
```

### Kubernetes Deployment
```bash
# Generate combined manifest
./test-build.sh  # Creates combined.yaml with proper substitutions

# Deploy to cluster
kubectl apply -f combined.yaml
```

### Protobuf Generation
```bash
# Manual generation (usually handled by test-build.sh)
# Generates to standard locations then moves to pkg/pb structure
protoc --proto_path=protos \
       --go_out=./app --go_opt=module=github.com/Anthony-Bible/password-exchange/app \
       --go-grpc_out=./app --go-grpc_opt=module=github.com/Anthony-Bible/password-exchange/app \
       protos/database.proto protos/encryption.proto protos/message.proto

# Move generated files to hexagonal structure
mv ./app/databasepb/* ./app/pkg/pb/database/
mv ./app/encryptionpb/* ./app/pkg/pb/encryption/ 
mv ./app/messagepb/* ./app/pkg/pb/message/

# Python protobuf for slackbot
protoc --proto_path=protos \
       --python_out=./python_protos \
       --python_grpc_out=./python_protos \
       protos/database.proto protos/encryption.proto protos/message.proto
```

## Configuration

- Services use Viper for configuration with environment variable support (prefix: `PASSWORDEXCHANGE_`)
- Required config file for CLI commands (specify with `--config` flag)
- Kubernetes secrets in `kubernetes/secrets.encrypted.yaml` (SOPS encrypted)
- Environment variables documented in project wiki

## Important Patterns

- All inter-service communication uses gRPC with protobuf definitions
- Web templates in `app/templates/` with assets in `app/assets/`
- Database operations centralized in database service
- Encryption/decryption handled by dedicated encryption service
- Email notifications use RabbitMQ message queue pattern
- Slackbot OAuth tokens stored in separate database tables via SQLAlchemy

## Testing

Run `./test-build.sh` to verify:
- Protobuf generation
- Go compilation
- Docker image builds for both main app and slackbot
- Kubernetes manifest generation with proper variable substitution