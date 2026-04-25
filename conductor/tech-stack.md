# Tech Stack

This document defines the technology stack and architectural patterns for Password Exchange.

## Core Languages
- **Go 1.24+**: The primary language for the main microservices (Web, Database, Encryption, Email).
- **Python**: Used for the Slackbot integration.

## Frameworks & Libraries
- **Gin**: Go web framework for the REST API and web UI.
- **gRPC**: High-performance inter-service communication using Protocol Buffers.
- **Flask**: Python web framework for the Slackbot.
- **Slack Bolt**: Framework for building Slack applications.
- **Cobra & Viper**: For CLI commands and configuration management in Go.
- **golang-migrate**: For automated database schema migrations and version tracking.
- **Zerolog**: For structured logging.
- **Prometheus**: For metrics and monitoring.
- **Swaggo/swag**: For automated Swagger/OpenAPI documentation generation.

## Data Storage & Messaging
- **MySQL/MariaDB**: Primary database for storing encrypted content and metadata.
- **SQLAlchemy**: Python ORM for database access in the Slackbot.
- **RabbitMQ**: Message queue for handling asynchronous notifications and background tasks.

## Infrastructure & Deployment
- **Kubernetes (K8s)**: Container orchestration for deploying all microservices.
- **Docker**: Containerization for application packaging.
- **RabbitMQ Operator**: Managing RabbitMQ clusters within Kubernetes.

## Architecture
- **Hexagonal Architecture (Ports and Adapters)**: Ensuring a clean separation between business logic and infrastructure.
- **Microservices**: Decomposing the system into independent, modular services (Web, Database, Encryption, Notification, Slackbot).
