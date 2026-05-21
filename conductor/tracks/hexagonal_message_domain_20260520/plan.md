# Implementation Plan: Refactor Message Domain to Strict Hexagonal Architecture

## Background & Motivation
The `message` domain currently violates strict hexagonal architecture by importing shared implementations (`app/internal/shared/logging`, `config`, `validation`) directly into the domain service. Additionally, it defines secondary port interfaces within its `domain/message_entities.go` instead of centralizing them in `ports/secondary/`. Refactoring this domain to match the `notification` domain will improve isolation, testability, and architectural consistency.

## Scope & Impact
**Scope:**
- `app/internal/domains/message/domain/message_entities.go`
- `app/internal/domains/message/domain/message_service.go`
- `app/internal/domains/message/domain/message_service_test.go`
- `app/internal/domains/message/ports/secondary/` (Add missing cross-cutting ports)
- Primary adapters and tests initializing `NewMessageService` (e.g. `adapters/primary/api/...`)

**Impact:**
- The domain layer will have zero dependencies on external or shared infrastructure.
- All configurations, logging, and validation functions will be injected as secondary ports.
- Tests will need to be updated to provide mocked versions of these cross-cutting dependencies.

## Proposed Solution
1. **Clean up Interfaces:** Remove duplicate interfaces (`EncryptionService`, `StorageService`, `NotificationService`, `PasswordHasher`, `URLBuilder`, `TurnstileValidator`, `WebRenderer`) from `message_entities.go`.
2. **Create New Secondary Ports:** Create `LoggerPort`, `ConfigPort`, `ValidationPort`, and `TurnstileValidatorPort` inside `ports/secondary/`.
3. **Decouple the Domain Service:** Refactor `MessageService` to accept the new secondary ports via dependency injection and replace direct function calls (e.g., `logging.Info()`, `config.AppConfig`, `validation.SanitizeEmailForLogging()`) with calls to the injected ports.
4. **Update Adapters & Tests:** Update all `NewMessageService` invocations in adapters and tests to provide the required ports. The tests will rely on mocked port implementations to isolate business logic testing.

## Phased Implementation Plan

### Phase 0: Branch Creation
- Ensure the local workspace is synchronized with the `master` branch.
- Create and checkout a new feature branch (e.g., `refactor/hexagonal-message-domain`) based on `master`.

### Phase 1: Define Secondary Ports
- Move or clean up redundant interface definitions in `message_entities.go`.
- Ensure existing domain interfaces match their counterparts in `ports/secondary/` (e.g., `StorageServicePort`).
- Create `LoggerPort`, `ConfigPort`, and `ValidationPort` in `app/internal/domains/message/ports/secondary/` mapping to the required behavior from the shared packages.
- Move `TurnstileValidator` to a proper secondary port file.

### Phase 2: Refactor Domain Service
- Modify the `MessageService` struct and `NewMessageService` constructor to require `LoggerPort`, `ConfigPort`, and `ValidationPort` along with the existing dependencies.
- Go through `message_service.go` and replace all direct `github.com/Anthony-Bible/password-exchange/app/internal/shared/*` calls with the injected ports.
- Ensure the file only imports standard libraries and its own secondary ports.

### Phase 3: Update Tests
- Generate or implement mocks for the newly created ports.
- Update `message_service_test.go` to provide these mocks during `NewMessageService` initialization.
- Verify tests fail (Red Phase), implement fixes, and pass (Green Phase).

### Phase 4: Wire Adapters
- Locate all instantiations of `NewMessageService` in the primary adapters (like REST API endpoints or gRPC server initializations).
- Create corresponding adapter implementations for the new cross-cutting ports and inject them.

## Verification & Testing
- **Unit Tests:** Verify `go test ./internal/domains/message/domain/...` runs green.
- **Build Verification:** Run `./test-build.sh` to ensure full Go compilation and no missing dependencies.
- **Dependency Audit:** Use static analysis or IDE tools to confirm that `app/internal/domains/message/domain` strictly relies on standard library + ports, and has NO imports to `shared/logging`, `config`, or other domains.

## Migration & Rollback Strategies
- **Migration:** No data migration is required, as this is purely a code restructuring.
- **Rollback:** In the event of catastrophic failures, use standard git version control to revert back to the commit before the refactoring. Because no database schemas are altered, codebase rollback is 100% safe.