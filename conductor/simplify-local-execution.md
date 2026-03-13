# Plan: Simplify Local Execution with Docker Compose

This plan outlines the steps to create a `docker-compose.yaml` file and an updated `Dockerfile` to allow for a single-command local setup of the entire Password Exchange system.

## Objective
Enable users to run the complete microservices architecture (Web, Database, Encryption, Email) along with their dependencies (MariaDB, RabbitMQ) using `docker compose up`.

## Proposed Solution

### 1. Unified Dockerfile
Update the existing `Dockerfile` to be more flexible, allowing it to start different components based on an environment variable or command argument. However, since the current `app` binary uses Cobra subcommands, we can just point to the same binary with different arguments in `docker-compose.yaml`.

### 2. Docker Compose Configuration
Create a `docker-compose.yaml` in the root directory with the following services:
- **db**: MariaDB instance with `passwordexchange.sql` initialized.
- **rabbitmq**: RabbitMQ instance for email notifications.
- **encryption-service**: Runs `./app encryption`.
- **database-service**: Runs `./app database`.
- **email-service**: Runs `./app email`.
- **web-service**: Runs `./app web`.

### 3. Environment Configuration
Provide a `.env.example` or bake default local development environment variables into the `docker-compose.yaml`.

## Implementation Steps

### Phase 1: Preparation
1. Verify `passwordexchange.sql` is suitable for MariaDB initialization (placed in `/docker-entrypoint-initdb.d/`).
2. Ensure `app/templates` and `app/assets` are correctly mapped in the Docker image.

### Phase 2: Dockerfile Enhancements (if needed)
1. The current `Dockerfile` already builds the `app` binary and copies templates/assets. It sets `ENTRYPOINT ["./app"]`. This is perfect because we can pass subcommands as `command` in `docker-compose.yaml`.

### Phase 3: Create Docker Compose File
1. Define the services, networks, and volumes.
2. Use internal gRPC service names for communication (e.g., `encryption-service:50051`).

### Phase 4: Verification
1. Run `docker compose up --build`.
2. Check logs for service health.
3. Access `localhost:8080`.

## Verification & Testing
- **Connectivity**: Verify gRPC communication between services.
- **Database**: Ensure schema is applied.
- **Functionality**: Submit a password and retrieve it via the web UI.

## Migration & Rollback
- This is a new feature; no migration required.
- If it fails, users can still use the manual method described in `GEMINI.md`.
