# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

Password Exchange is a microservices-based secure password sharing platform using **hexagonal architecture** (ports and adapters pattern) with:

### Core Services
- **Web Service** (`app web`): Gin-based HTTP server serving frontend and REST API
- **Database Service** (`app database`): gRPC service for all database operations  
- **Encryption Service** (`app encryption`): gRPC service handling encryption/decryption
- **Email Service** (`app email`): RabbitMQ consumer for sending notification emails
- **Slackbot**: Python Flask app with Slack Bolt framework for Slack integration

### Hexagonal Architecture Structure
```
internal/domains/{domain}/
├── domain/           # Business logic (entities, services, errors)
├── ports/
│   ├── primary/      # Inbound interfaces (REST API, gRPC, web handlers)
│   └── secondary/    # Outbound interfaces (databases, external services)
└── adapters/
    ├── primary/      # Inbound implementations (API handlers, gRPC servers)
    └── secondary/    # Outbound implementations (MySQL, RabbitMQ, SMTP)
```

**Key Domains:**
- `message/`: Core password sharing logic with submission, retrieval, encryption
- `storage/`: Database operations via gRPC
- `encryption/`: Key generation and cryptographic operations via gRPC
- `notification/`: Email notifications via RabbitMQ

### Key Technologies
- **Go 1.23+**: Main application with Cobra CLI, Gin web framework, gRPC services
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

### Primary Build Script
```bash
# Complete build verification (ALWAYS run before commits)
./test-build.sh
```
This script performs:
- Protobuf generation for Go and Python
- Go build compilation
- Swagger documentation generation
- Docker image builds (main app + slackbot)
- Kubernetes manifest generation with variable substitution

### Testing
```bash
# Run all Go tests
cd app && go test ./... -v

# Run specific domain tests
cd app && go test ./internal/domains/message/... -v

# Run tests with coverage
cd app && go test ./... -cover

# Run single test file
cd app && go test ./internal/domains/message/adapters/primary/api -v
```

### Building Individual Components
```bash
# Build Go application only
cd app && go mod tidy && go build -o app

# Generate protobuf files manually
protoc --proto_path=protos --go_out=./app --go_opt=module=github.com/Anthony-Bible/password-exchange/app --go-grpc_out=./app --go-grpc_opt=module=github.com/Anthony-Bible/password-exchange/app protos/*.proto

# Generate Swagger docs
cd app && swag init -g internal/domains/message/adapters/primary/api/docs.go -o docs --parseDependency

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

## Hexagonal Architecture Patterns

### Domain Layer (`domain/`)
- **Entities**: Core business objects (`entities.go`, `message_entities.go`)
- **Services**: Business logic implementations (`message_service.go`)
- **Errors**: Domain-specific error types (`message_errors.go`)
- NO external dependencies - only standard library and other domain objects

### Ports (`ports/`)
- **Primary ports**: Inbound interfaces defining what the domain offers
- **Secondary ports**: Outbound interfaces defining what the domain needs
- Define contracts WITHOUT implementations

### Adapters (`adapters/`)
- **Primary adapters**: Handle inbound requests (REST API, gRPC servers, web handlers)
- **Secondary adapters**: Implement outbound calls (databases, message queues, external APIs)

### Key Implementation Patterns
- All inter-service communication uses gRPC with protobuf definitions
- Dependency injection: Domain services receive ports, adapters implement ports
- Web templates in `app/templates/` with assets in `app/assets/`
- Database operations centralized in database service via gRPC
- Encryption/decryption handled by dedicated encryption service via gRPC
- Email notifications use RabbitMQ message queue pattern
- Testing uses mocks for ports to isolate domain logic
- Slackbot OAuth tokens stored in separate database tables via SQLAlchemy

### Adding New Features
1. Define business logic in `domain/` layer
2. Create ports for external dependencies
3. Implement primary adapters for inbound requests
4. Implement secondary adapters for outbound calls
5. Wire dependencies in main application

## Testing

### Test-Driven Development (TDD)

**MUST follow strict Test-Driven Development practices** for all code changes:

1. **Red Phase**: Write failing tests first that describe the expected behavior
2. **Green Phase**: Write minimal code to make tests pass
3. **Refactor Phase**: Improve code while keeping tests green
4. **Iterate**: Repeat cycle for each feature/requirement

**TDD Workflow:**
```bash
# 1. Write failing test
go test ./... -v  # Expect failures

# 2. Write minimal implementation
# 3. Run tests again
go test ./... -v  # Expect success

# 4. Refactor and verify tests still pass
go test ./... -v  # Must remain green
```

**Testing Requirements:**
- Always use test driven development
- All new functionality MUST have tests written before implementation
- Tests must initially fail to prove they test the right behavior
- No code should be written without corresponding tests
- All tests must pass before code is considered complete

### Build Verification

**ALWAYS run `./test-build.sh` before commits** to verify:
- Protobuf generation for Go and Python
- Go compilation without errors
- Swagger documentation generation with validation
- Docker image builds for both main app and slackbot
- Kubernetes manifest generation with proper variable substitution

### Deployment Information
- **Dev environment**: https://dev.password.exchange
- **Production**: https://password.exchange  
- Deployment is manual - **MUST confirm with user before assuming deployment**
- Generated `combined.yaml` contains all Kubernetes manifests with substituted variables 

## Commit Standards

**MUST use conventional commits** for all commits to this repository. Format:
```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
```