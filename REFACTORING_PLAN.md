# Hexagonal Architecture Refactoring Plan

This document outlines the complete plan for refactoring the Password Exchange application to use hexagonal (ports and adapters) architecture.

## Overview

**Goal:** Transform the current monolithic command structure into a clean hexagonal architecture with clear separation of concerns, improved testability, and better maintainability.

**Approach:** Incremental refactoring maintaining backward compatibility throughout the process.

---

## ✅ COMPLETED PHASES

### Phase 1: Create Hexagonal Directory Structure ✅
**Status:** COMPLETED

Created the foundation directory structure following hexagonal architecture principles:

```
app/
├── internal/                    # Private application code
│   ├── shared/                  # Cross-cutting concerns
│   │   ├── config/             # Configuration management
│   │   ├── logging/            # Structured logging (placeholder)
│   │   └── validation/         # Common validation utilities
│   └── domains/                # Business domains
│       ├── message/            # Core message sharing domain
│       ├── encryption/         # Cryptographic operations domain
│       ├── storage/            # Data persistence domain
│       └── notification/       # Communication domain
└── pkg/                        # Public interfaces
    ├── pb/                     # Generated protobuf code
    └── clients/                # Service clients
```

Each domain follows the hexagonal pattern:
- `domain/` - Business entities & rules
- `ports/primary/` - Inbound interfaces (web, gRPC)
- `ports/secondary/` - Outbound interfaces (database, external services)
- `adapters/primary/` - Inbound implementations
- `adapters/secondary/` - Outbound implementations

### Phase 2: Move Shared Components ✅
**Status:** COMPLETED

Relocated shared utilities and configuration:
- ✅ `app/config/` → `app/internal/shared/config/`
- ✅ `app/commons/` → `app/internal/shared/validation/` (renamed to env.go)
- ✅ `app/message/` → `app/internal/domains/message/domain/entities.go`
- ✅ Updated all import paths across codebase
- ✅ Verified builds work after moves

### Phase 3: Organize Generated Code ✅
**Status:** COMPLETED

Moved protobuf files to clean pkg structure:
- ✅ `app/databasepb/` → `app/pkg/pb/database/`
- ✅ `app/encryptionpb/` → `app/pkg/pb/encryption/`
- ✅ `app/messagepb/` → `app/pkg/pb/message/`
- ✅ Updated import paths in all consuming files
- ✅ Updated build scripts for new structure

### Phase 4: Update Build System ✅
**Status:** COMPLETED

Updated build scripts and Docker configuration:
- ✅ Modified `test-build.sh` for new protobuf generation paths
- ✅ Verified Docker builds work with new structure
- ✅ Updated documentation (CLAUDE.md)
- ✅ Added proper .gitignore entries
- ✅ All builds tested and working

### Phase 5.3: Storage Domain Implementation ✅
**Status:** COMPLETED

Successfully extracted Storage domain with complete hexagonal architecture:

**Domain Layer:**
- ✅ `domain/entities.go` - Message entity and MessageRepository interface
- ✅ `domain/service.go` - Business logic with validation and error handling
- ✅ `domain/errors.go` - Domain-specific error definitions

**Ports (Interfaces):**
- ✅ `ports/primary/service.go` - StorageServicePort interface for inbound operations
- ✅ MessageRepository interface in domain for outbound data access

**Adapters:**
- ✅ `adapters/secondary/mysql/adapter.go` - MySQL implementation of MessageRepository
- ✅ `adapters/primary/grpc/server.go` - gRPC server implementing protobuf interface

**Integration:**
- ✅ Updated `cmd/database/database2.go` to use hexagonal architecture
- ✅ Dependency injection with proper service composition
- ✅ Backward compatibility maintained
- ✅ All builds and tests passing
- ✅ Successfully deployed and tested

**Key Achievements:**
- Clean separation of business logic from infrastructure
- Testable domain logic with clear interfaces
- Database technology independence through repository pattern
- gRPC presentation layer properly decoupled
- Foundation established for remaining domain extractions

### Phase 5.2: Encryption Domain Implementation ✅
**Status:** COMPLETED

Successfully extracted Encryption domain with complete hexagonal architecture:

**Domain Layer:**
- ✅ `domain/entities.go` - EncryptionKey, encryption/decryption entities with KeyGenerator interface
- ✅ `domain/service.go` - AES-GCM cryptographic business logic with comprehensive error handling
- ✅ `domain/errors.go` - Domain-specific error definitions for encryption operations

**Ports (Interfaces):**
- ✅ `ports/primary/service.go` - EncryptionServicePort interface for inbound operations
- ✅ `ports/secondary/keygen.go` - KeyGeneratorPort interface for outbound key generation

**Adapters:**
- ✅ `adapters/secondary/memory/keygen.go` - In-memory key generation using crypto/rand
- ✅ `adapters/primary/grpc/server.go` - gRPC server implementing protobuf MessageService interface

**Integration:**
- ✅ Updated `cmd/encryption/encryption2.go` to use hexagonal architecture
- ✅ Dependency injection with proper service composition
- ✅ Backward compatibility maintained with legacy server methods
- ✅ Context.Context parameters added for consistency and best practices
- ✅ All builds and tests passing
- ✅ Successfully deployed and tested via PR #341

**Key Achievements:**
- Clean separation of cryptographic business logic from infrastructure concerns
- Testable encryption service with clear interfaces and comprehensive error handling
- Technology independence through adapter pattern for key generation
- gRPC presentation layer properly decoupled from domain logic
- Context propagation following Go best practices
- Foundation ready for comprehensive unit and integration testing

---

## 🚧 PLANNED PHASES

### Phase 5: Extract Domain Logic
**Status:** IN PROGRESS (Storage ✅, Encryption ✅ COMPLETED)
**Priority:** HIGH

Extract business logic from current services into domain layers:

#### 5.1: Message Domain
- [ ] Extract message validation logic from web handlers
- [ ] Create message sharing business rules
- [ ] Define message lifecycle management
- [ ] Move from `cmd/web/forms.go` to `internal/domains/message/domain/`

#### 5.2: Encryption Domain ✅
- [x] Extract encryption algorithms from `cmd/encryption/encryption2.go`
- [x] Create key generation and management logic with KeyGenerator interface
- [x] Define cryptographic policies and error handling in domain layer
- [x] Move to `internal/domains/encryption/domain/`
- [x] Implement hexagonal architecture with ports and adapters
- [x] Create memory-based secondary adapter for key generation
- [x] Create gRPC primary adapter implementing protobuf interface
- [x] Maintain backward compatibility with legacy server methods
- [x] Add context.Context parameters for consistency and best practices
- [x] Test and deploy successfully

#### 5.3: Storage Domain ✅
- [x] Extract repository patterns from `cmd/database/database2.go`
- [x] Create data access abstractions with MessageRepository interface
- [x] Define query composition logic in MySQL adapter
- [x] Move to `internal/domains/storage/domain/`
- [x] Implement hexagonal architecture with ports and adapters
- [x] Create MySQL secondary adapter
- [x] Create gRPC primary adapter
- [x] Maintain backward compatibility
- [x] Test and deploy successfully

#### 5.4: Notification Domain
- [ ] Extract email logic from `cmd/email/`
- [ ] Create notification templates and routing
- [ ] Define delivery policies
- [ ] Move to `internal/domains/notification/domain/`

### Phase 6: Define Ports (Interfaces)
**Status:** PENDING
**Priority:** HIGH

Create port interfaces for each domain:

#### 6.1: Primary Ports (Inbound)
- [ ] `message/ports/primary/web.go` - HTTP handlers interface
- [ ] `message/ports/primary/grpc.go` - gRPC server interface
- [ ] `encryption/ports/primary/service.go` - Encryption service interface
- [ ] `storage/ports/primary/service.go` - Storage service interface
- [ ] `notification/ports/primary/service.go` - Notification service interface

#### 6.2: Secondary Ports (Outbound)
- [ ] `message/ports/secondary/storage.go` - Storage interface
- [ ] `message/ports/secondary/encryption.go` - Encryption interface  
- [ ] `message/ports/secondary/notification.go` - Notification interface
- [ ] `storage/ports/secondary/database.go` - Database interface
- [ ] `encryption/ports/secondary/keystore.go` - Key storage interface
- [ ] `notification/ports/secondary/mailer.go` - Email sending interface
- [ ] `notification/ports/secondary/queue.go` - Message queue interface

### Phase 7: Implement Adapters
**Status:** PENDING
**Priority:** HIGH

Create adapter implementations:

#### 7.1: Secondary Adapters (Infrastructure)
- [ ] `storage/adapters/secondary/mysql/` - MySQL database adapter
- [ ] `encryption/adapters/secondary/memory/` - In-memory key storage
- [ ] `notification/adapters/secondary/smtp/` - SMTP email adapter
- [ ] `notification/adapters/secondary/rabbitmq/` - RabbitMQ adapter
- [ ] `message/adapters/secondary/grpc_clients/` - External service clients

#### 7.2: Primary Adapters (Presentation)
- [ ] `message/adapters/primary/web/` - HTTP/Gin web adapter
- [ ] `message/adapters/primary/grpc/` - gRPC server adapter
- [ ] `encryption/adapters/primary/grpc/` - Encryption gRPC server
- [ ] `storage/adapters/primary/grpc/` - Storage gRPC server
- [ ] `notification/adapters/primary/consumer/` - RabbitMQ consumer

### Phase 8: Move Static Assets
**Status:** PENDING
**Priority:** MEDIUM

Relocate web-specific assets to appropriate domain:
- [ ] `app/templates/` → `message/adapters/primary/web/templates/`
- [ ] `app/assets/` → `message/adapters/primary/web/assets/`
- [ ] Update Docker and build scripts

### Phase 9: Create Service Clients
**Status:** PENDING
**Priority:** MEDIUM

Implement reusable service clients:
- [ ] `pkg/clients/encryption.go` - Encryption service client
- [ ] `pkg/clients/storage.go` - Storage service client
- [ ] Update inter-service communication

### Phase 10: Wire Dependencies
**Status:** PENDING
**Priority:** HIGH

Implement dependency injection:
- [ ] Create application assemblers for each domain
- [ ] Wire ports to adapters with dependency injection
- [ ] Update main.go to use hexagonal structure
- [ ] Simplify cmd/ to just wire domains together

### Phase 11: Testing Strategy
**Status:** PENDING
**Priority:** HIGH

Implement comprehensive testing:
- [ ] Unit tests for domain logic (isolated)
- [ ] Integration tests for adapters (with real implementations)
- [ ] Contract tests for ports (verify implementations)
- [ ] End-to-end tests for complete workflows

### Phase 12: Documentation & Examples
**Status:** PENDING
**Priority:** MEDIUM

Create comprehensive documentation:
- [ ] Architecture decision records (ADRs)
- [ ] Domain interaction diagrams
- [ ] Developer onboarding guide
- [ ] Testing examples and patterns

---

## Benefits Achieved So Far

✅ **Clean Code Organization** - Clear separation between business logic, infrastructure, and interfaces
✅ **Build System Compatibility** - All existing build processes work with new structure
✅ **Backward Compatibility** - No breaking changes to existing functionality
✅ **Documentation** - Architecture guidance in CLAUDE.md
✅ **Testability Foundation** - Structure ready for comprehensive testing
✅ **Technology Independence** - Generated code separated from business logic
✅ **Hexagonal Architecture Implementation** - Storage and Encryption domains fully extracted with proper ports and adapters
✅ **Domain-Driven Design** - Business logic separated from infrastructure concerns across multiple domains
✅ **Dependency Injection** - Clean service composition with interface-based dependencies
✅ **Repository Pattern** - Data access abstraction with MySQL implementation
✅ **Cryptographic Abstraction** - Encryption business logic separated from infrastructure with AES-GCM implementation
✅ **Context Propagation** - Consistent context.Context usage following Go best practices

## Benefits After Completion

🎯 **Independent Deployability** - Each domain can be deployed separately
🎯 **Technology Flexibility** - Easy to swap databases, frameworks, etc.
🎯 **Enhanced Testability** - Mock adapters for isolated unit testing
🎯 **Clear Boundaries** - Explicit separation of concerns
🎯 **Maintainability** - Changes in one domain don't affect others
🎯 **Team Scalability** - Different teams can own different domains

---

## Migration Strategy

**Incremental Approach:**
1. Extract one domain at a time (start with least coupled)
2. Create tests for ports before implementing adapters  
3. Keep old structure working during migration
4. Validate each step with comprehensive testing

**Recommended Order:**
1. ✅ Storage domain (most foundational) - COMPLETED
2. ✅ Encryption domain (clear boundaries) - COMPLETED 
3. Notification domain (fewer dependencies) - NEXT
4. Message domain (most complex, do last) - REMAINING

**Risk Mitigation:**
- Each phase can be done independently
- Rollback capability at each step
- Maintain existing functionality throughout
- Comprehensive testing at each milestone