# Track Specification: Service Health and Swagger Documentation

## Description
This track focuses on improving the observability and documentation of the Password Exchange platform by implementing comprehensive service health checks and fully automating the generation of Swagger (OpenAPI) documentation for the REST API.

## Goals
- **Service Health Checks**:
  - Replace placeholder health check responses in `app/internal/domains/message/adapters/primary/api/handlers.go` with actual checks for the Database, Encryption, and Email services.
  - Implement a standard `/health` endpoint that aggregates health status from all dependent gRPC services and RabbitMQ.
  - Ensure the system provides meaningful error codes and status messages when a dependency is down.

- **Swagger Documentation**:
  - Fully integrate the `swaggo/swag` library to generate OpenAPI documentation from Go source code.
  - Ensure all existing REST API endpoints (e.g., `/messages`, `/messages/:id/decrypt`) are properly documented with correct request/response types and error codes.
  - Automate the generation of `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml` as part of the build process.
  - Verify that the Swagger UI is accessible at `/api/v1/docs/index.html`.

## Requirements
- Use existing Go patterns for hexagonal architecture.
- Adhere to the `go.md` and `general.md` code style guides.
- Implement tests for new health check logic using mocks where appropriate.
- Ensure no regression in existing API functionality.

## Context
- **Files of Interest**:
  - `app/internal/domains/message/adapters/primary/api/handlers.go`: Current health check placeholders.
  - `app/internal/domains/message/adapters/primary/api/docs.go`: Swagger entry point.
  - `test-build.sh`: Existing build automation.
  - `app/go.mod`: Dependency management.
