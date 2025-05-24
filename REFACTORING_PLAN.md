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

---

## 🚧 PLANNED PHASES

### Phase 5: Extract Domain Logic
**Status:** PENDING
**Priority:** HIGH

Extract business logic from current services into domain layers:

#### 5.1: Message Domain
- [ ] Extract message validation logic from web handlers
- [ ] Create message sharing business rules
- [ ] Define message lifecycle management
- [ ] Move from `cmd/web/forms.go` to `internal/domains/message/domain/`

#### 5.2: Encryption Domain  
- [ ] Extract encryption algorithms from `cmd/encryption/encryption2.go`
- [ ] Create key generation and management logic
- [ ] Define cryptographic policies
- [ ] Move to `internal/domains/encryption/domain/`

#### 5.3: Storage Domain
- [ ] Extract repository patterns from `cmd/database/database2.go`
- [ ] Create data access abstractions
- [ ] Define query composition logic
- [ ] Move to `internal/domains/storage/domain/`

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
1. Storage domain (most foundational)
2. Encryption domain (clear boundaries) 
3. Notification domain (fewer dependencies)
4. Message domain (most complex, do last)

**Risk Mitigation:**
- Each phase can be done independently
- Rollback capability at each step
- Maintain existing functionality throughout
- Comprehensive testing at each milestone