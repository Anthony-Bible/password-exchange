# REST/JSON API Implementation Plan

## Overview

This document outlines the implementation plan for converting the Password Exchange system from a template-based web interface to a REST/JSON API with OpenAPI specification, while maintaining backward compatibility.

## ✅ Implementation Checklist

### Phase 1: Core API Infrastructure (2-3 weeks)

#### Setup & Architecture
- [ ] Create API directory structure `app/internal/domains/message/adapters/primary/api/`
- [ ] Set up basic API server with Gin framework
- [ ] Create API models package `app/internal/domains/message/adapters/primary/api/models/`
- [ ] Implement RFC 7807 error response format
- [ ] Add request correlation ID middleware

#### Core Endpoints Implementation
- [ ] **POST /api/v1/messages** - Submit message endpoint
  - [ ] Request validation and parsing
  - [ ] Integration with existing MessageService
  - [ ] JSON response formatting
  - [ ] Error handling for validation failures
- [ ] **GET /api/v1/messages/{id}** - Get message access info
  - [ ] Parameter parsing and validation
  - [ ] Message existence check
  - [ ] Passphrase requirement detection
- [ ] **POST /api/v1/messages/{id}/decrypt** - Decrypt message
  - [ ] Decryption key handling
  - [ ] Passphrase validation
  - [ ] One-time access enforcement
- [ ] **GET /api/v1/health** - Health check endpoint
  - [ ] Service dependency health checks
  - [ ] JSON status response
- [ ] **GET /api/v1/info** - API information endpoint

#### Request/Response Models
- [ ] MessageSubmissionRequest struct with validation tags
- [ ] MessageSubmissionResponse struct
- [ ] MessageAccessInfoResponse struct  
- [ ] MessageDecryptRequest struct
- [ ] MessageDecryptResponse struct
- [ ] HealthCheckResponse struct
- [ ] StandardErrorResponse struct

#### Basic Testing
- [ ] Unit tests for each endpoint handler
- [ ] Request/response model validation tests
- [ ] Error handling test scenarios
- [ ] Integration tests with existing domain services

### Phase 2: Validation & Documentation (1-2 weeks)

#### Validation Middleware
- [ ] Install and configure `github.com/go-playground/validator/v10`
- [ ] Create custom validators for conditional fields
- [ ] Anti-spam answer validation (`oneof=blue`)
- [ ] Email conditional validation logic
- [ ] Request size and timeout limits

#### OpenAPI Specification
- [ ] Create `app/api/openapi.yaml` specification file
- [ ] Document all endpoint schemas with examples
- [ ] Define error response schemas
- [ ] Add security scheme definitions
- [ ] Include request/response examples

#### Swagger Documentation
- [ ] Install `github.com/swaggo/gin-swagger`
- [ ] Add swagger annotations to handlers
- [ ] Generate documentation at `/api/v1/docs`
- [ ] Auto-update docs in CI/CD pipeline

#### Enhanced Testing
- [ ] OpenAPI specification validation tests
- [ ] Request/response schema compliance tests
- [ ] Anti-spam validation test cases
- [ ] Email notification flow tests

### Phase 3: Security & Production Features (1-2 weeks)

#### Rate Limiting
- [ ] Install `github.com/ulule/limiter/v3`
- [ ] Configure per-endpoint rate limits:
  - [ ] POST /api/v1/messages: 10 requests/hour per IP
  - [ ] GET /api/v1/messages/*: 100 requests/hour per IP
  - [ ] POST /decrypt: 20 requests/hour per IP
- [ ] Implement sliding window algorithm
- [ ] Add rate limit headers to responses
- [ ] Rate limit exceeded error responses

#### Authentication Framework
- [ ] Design API key authentication system
- [ ] Create bearer token middleware
- [ ] Admin endpoint protection
- [ ] Token validation and error handling

#### CORS & Security Headers
- [ ] Configure CORS for API endpoints
- [ ] Add security headers middleware
- [ ] Request timeout configuration
- [ ] Request size limits

#### Monitoring & Logging
- [ ] Structured JSON logging with correlation IDs
- [ ] Request/response logging middleware
- [ ] Performance metrics collection
- [ ] Error tracking and alerting setup

### Phase 4: Integration & Deployment

#### Backward Compatibility
- [ ] Maintain existing web routes
- [ ] Update existing API detection in `handlers.go:55`
- [ ] Content-type based routing (`Accept: application/json`)
- [ ] Path-based API routing (`/api/*`)

#### Production Configuration
- [ ] Environment-based configuration
- [ ] Docker image updates with API support
- [ ] Kubernetes manifest updates
- [ ] Load balancer configuration for API traffic

#### Documentation & Examples
- [ ] API usage documentation
- [ ] Client SDK examples
- [ ] Postman collection creation
- [ ] Migration guide from web forms to API

### Phase 5: Future Enhancements (Optional)

#### Advanced Features
- [ ] **DELETE /api/v1/messages/{id}** - Admin message deletion
- [ ] **PATCH /api/v1/messages/{id}** - Extend message expiration
- [ ] **GET /api/v1/messages** - Bulk message listing (admin)
- [ ] Pagination support for bulk operations

#### Advanced Authentication
- [ ] JWT token implementation
- [ ] OAuth 2.0 integration
- [ ] API key rotation policies
- [ ] User-specific rate limits

#### Performance & Monitoring
- [ ] Redis-based rate limiting store
- [ ] Prometheus metrics integration
- [ ] Distributed tracing
- [ ] Performance optimization

## Current Architecture Analysis

### Current Web Interface
- Form-based HTML submission with complex conditional email functionality
- Template rendering for message display and decryption  
- Mixed GET/POST pattern for decryption flow
- Anti-spam verification with simple color question
- Located in: `app/internal/domains/message/adapters/primary/web/`

### Current Data Flow
1. **Submit**: Form → MessageSubmissionRequest → Encryption → Storage → Optional Notification
2. **Retrieve**: Access URL → CheckMessageAccess → Display/Decrypt → MessageRetrievalRequest

## REST API Endpoints Design

### Core Message Endpoints

#### 1. Submit Message
```http
POST /api/v1/messages
Content-Type: application/json

Request Body:
{
  "content": "string (required)",
  "sender": {
    "name": "string (conditional: required if sendNotification=true)",
    "email": "string (conditional: required if sendNotification=true)"
  },
  "recipient": {
    "name": "string (conditional: required if sendNotification=true)", 
    "email": "string (conditional: required if sendNotification=true)"
  },
  "passphrase": "string (optional)",
  "additionalInfo": "string (optional)",
  "sendNotification": "boolean (default: false)",
  "antiSpamAnswer": "string (conditional: required if sendNotification=true)"
}

Response: 201 Created
{
  "messageId": "uuid",
  "decryptUrl": "https://domain.com/api/v1/messages/{uuid}/decrypt?key={base64key}",
  "webUrl": "https://domain.com/decrypt/{uuid}/{base64key}",
  "expiresAt": "2024-01-01T12:00:00Z",
  "notificationSent": true
}

Error: 400 Bad Request
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": {
    "content": "Content is required",
    "sender.email": "Valid email required when notifications enabled"
  }
}
```

#### 2. Get Message Access Info
```http
GET /api/v1/messages/{messageId}?key={base64key}

Response: 200 OK
{
  "messageId": "uuid",
  "exists": true,
  "requiresPassphrase": true,
  "hasBeenAccessed": false,
  "expiresAt": "2024-01-01T12:00:00Z"
}

Response: 404 Not Found
{
  "error": "message_not_found",
  "message": "Message not found or has expired"
}
```

#### 3. Decrypt Message
```http
POST /api/v1/messages/{messageId}/decrypt
Content-Type: application/json

Request Body:
{
  "decryptionKey": "base64-encoded-key",
  "passphrase": "string (optional)"
}

Response: 200 OK
{
  "messageId": "uuid",
  "content": "decrypted message content",
  "decryptedAt": "2024-01-01T12:00:00Z"
}

Error: 401 Unauthorized
{
  "error": "invalid_passphrase", 
  "message": "Invalid passphrase provided"
}

Error: 410 Gone
{
  "error": "message_consumed",
  "message": "Message has already been accessed and deleted"
}
```

### Utility Endpoints

#### 4. Health Check
```http
GET /api/v1/health

Response: 200 OK
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "encryption": "healthy", 
    "email": "healthy"
  }
}
```

#### 5. API Information
```http
GET /api/v1/info

Response: 200 OK
{
  "version": "1.0.0",
  "documentation": "/api/v1/docs",
  "endpoints": {
    "submit": "POST /api/v1/messages",
    "access": "GET /api/v1/messages/{id}",
    "decrypt": "POST /api/v1/messages/{id}/decrypt"
  },
  "features": {
    "emailNotifications": true,
    "passphraseProtection": true,
    "antiSpamProtection": true
  }
}
```

### Future Enhancement Endpoints

#### 6. Message Management (Optional)
```http
DELETE /api/v1/messages/{messageId}
Authorization: Bearer {admin-token}

Response: 204 No Content

PATCH /api/v1/messages/{messageId}
Content-Type: application/json
Authorization: Bearer {admin-token}

Request Body:
{
  "expiresAt": "2024-01-02T12:00:00Z"
}

Response: 200 OK
{
  "messageId": "uuid",
  "expiresAt": "2024-01-02T12:00:00Z"
}
```

#### 7. Bulk Operations (Optional)
```http
GET /api/v1/messages
Authorization: Bearer {admin-token}
Query Parameters:
- limit: int (default: 50, max: 100)
- offset: int (default: 0)
- status: string (active|expired|accessed)

Response: 200 OK
{
  "messages": [
    {
      "messageId": "uuid",
      "createdAt": "2024-01-01T10:00:00Z",
      "expiresAt": "2024-01-01T12:00:00Z", 
      "hasBeenAccessed": false,
      "requiresPassphrase": true
    }
  ],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "total": 150
  }
}
```

## HTTP Status Code Mapping

### Success
- `200 OK` - Successful GET/POST operations
- `201 Created` - Message successfully submitted 
- `204 No Content` - Successful DELETE operations

### Client Errors
- `400 Bad Request` - Invalid request format/validation errors
- `401 Unauthorized` - Invalid passphrase or missing auth
- `404 Not Found` - Message not found or expired
- `410 Gone` - Message already accessed and deleted
- `422 Unprocessable Entity` - Anti-spam verification failed
- `429 Too Many Requests` - Rate limit exceeded

### Server Errors
- `500 Internal Server Error` - Encryption/database service errors
- `503 Service Unavailable` - Email service unavailable

## Error Response Standard

All API errors follow RFC 7807 Problem Details format:

```json
{
  "error": "error_code",
  "message": "Human-readable description", 
  "details": {}, // Optional validation details
  "timestamp": "2024-01-01T12:00:00Z",
  "path": "/api/v1/messages"
}
```

## Implementation Strategy

### Phase 1: Add API Layer
- Create new `/api/v1` routes alongside existing web routes
- Add `ApiHandler` struct in `app/internal/domains/message/adapters/primary/api/`
- Modify existing `handlers.go:55` API detection logic for full API support
- Add proper JSON error handling and HTTP status codes

### Phase 2: OpenAPI Integration
- Add OpenAPI 3.0 specification file
- Integrate Swaggo/Gin-Swagger for auto-documentation
- Add validation middleware using go-playground/validator
- Create API versioning strategy

### Phase 3: Enhanced Features
- Add rate limiting and authentication (API keys)
- Implement proper CORS handling
- Add structured logging with request tracing
- Create comprehensive error response format

## File Structure

```
app/
├── internal/domains/message/adapters/primary/
│   ├── web/                    # Existing web interface
│   │   ├── server.go
│   │   └── handlers.go
│   └── api/                    # New API interface
│       ├── server.go
│       ├── handlers.go
│       ├── models/
│       │   ├── message.go
│       │   ├── error.go
│       │   └── health.go
│       └── middleware/
│           ├── validation.go
│           ├── rate_limit.go
│           └── error_handler.go
├── api/
│   ├── openapi.yaml           # OpenAPI 3.0 specification
│   └── docs/                  # Generated swagger docs
└── pkg/
    └── api/
        └── validation/        # Shared validation logic
```

## Error Handling & Validation Strategy

### Input Validation
- Use `github.com/go-playground/validator/v10` for struct validation
- Custom validators for anti-spam answers and email conditionals
- JSON schema validation at middleware level

### Validation Rules
```go
type MessageSubmissionRequest struct {
    Content           string `json:"content" validate:"required,min=1,max=10000"`
    Sender            *Sender `json:"sender" validate:"required_if=SendNotification true"`
    Recipient         *Recipient `json:"recipient" validate:"required_if=SendNotification true"` 
    Passphrase        string `json:"passphrase" validate:"max=500"`
    SendNotification  bool   `json:"sendNotification"`
    AntiSpamAnswer    string `json:"antiSpamAnswer" validate:"required_if=SendNotification true,oneof=blue"`
}

type Sender struct {
    Name  string `json:"name" validate:"required,min=1,max=100"`
    Email string `json:"email" validate:"required,email"`
}
```

### Error Response Middleware
- Centralized error handling in `app/internal/domains/message/adapters/primary/api/middleware/`
- Structured logging with request correlation IDs
- Panic recovery with 500 responses

## Authentication & Rate Limiting Plan

### Rate Limiting Strategy
- Use `github.com/ulule/limiter/v3` with Redis/memory store
- Different limits per endpoint:
  - `POST /api/v1/messages`: 10 requests/hour per IP
  - `GET /api/v1/messages/*`: 100 requests/hour per IP  
  - `POST /decrypt`: 20 requests/hour per IP
- Sliding window algorithm with burst allowance

### Authentication Options
1. **API Keys** (Phase 1): Simple bearer tokens for basic access control
2. **JWT Tokens** (Phase 2): For user-specific operations and admin endpoints
3. **OAuth 2.0** (Future): For third-party integrations

### Implementation Example
```go
// Rate limiting middleware
func RateLimitMiddleware() gin.HandlerFunc {
    store := memory.NewStore()
    rate := limiter.Rate{
        Period: 1 * time.Hour,
        Limit:  10, // Per endpoint configuration
    }
    return gin_limiter.NewMiddleware(limiter.New(store, rate))
}
```

## OpenAPI Specification Plan

### Key Features
- RESTful endpoint documentation with examples
- JSON schema validation for all request/response bodies
- Security schemes for future API authentication
- Error response standardization (RFC 7807 Problem Details)
- Request/response examples and testing scenarios

### Documentation
- Interactive Swagger UI at `/api/v1/docs`
- OpenAPI 3.0 specification file
- Code generation for client SDKs
- Automated API testing integration

## Backward Compatibility

### Compatibility Strategy
- Maintain existing web routes for current users
- Detect API requests via `Accept: application/json` header or `/api` path prefix
- Gradual migration path from form-based to API-based integration
- Existing handlers in `app/internal/domains/message/adapters/primary/web/handlers.go:55` already have basic API detection

### Migration Path
1. Deploy API alongside existing web interface
2. Update documentation to include API examples
3. Provide client libraries and SDKs
4. Eventually deprecate form-based interface (optional)

## Dependencies

### Required Go Packages
```go
// API Framework
github.com/gin-gonic/gin

// Validation
github.com/go-playground/validator/v10

// Rate Limiting  
github.com/ulule/limiter/v3

// OpenAPI Documentation
github.com/swaggo/gin-swagger
github.com/swaggo/files

// Logging
github.com/rs/zerolog

// Testing
github.com/stretchr/testify
```

## Testing Strategy

### Unit Tests
- Test each API endpoint with various input scenarios
- Validate request/response schemas
- Test error handling and validation logic

### Integration Tests
- End-to-end API workflow testing
- Rate limiting and authentication testing
- OpenAPI specification validation

### Performance Tests
- Load testing for rate limiting
- Concurrent access testing
- Memory and CPU profiling

## Monitoring and Observability

### Metrics
- Request/response times per endpoint
- Error rates and types
- Rate limiting hits
- Message submission/retrieval counts

### Logging
- Structured JSON logging with correlation IDs
- Request/response logging for debugging
- Error tracking and alerting

### Health Checks
- Service dependency health monitoring
- Database connection health
- Encryption service availability
- Email service status

## Security Considerations

### Input Validation
- Strict validation of all user inputs
- Anti-spam protection for message submission
- Request size limits and timeouts

### Rate Limiting
- IP-based rate limiting per endpoint
- Burst protection for legitimate users
- DDoS protection strategies

### Authentication
- Secure API key management
- Token rotation policies
- Admin endpoint protection

### Data Protection
- Encryption key security
- Secure message storage
- PII handling compliance

## Deployment

### Development
- Docker containerization with API documentation
- Local development with hot reload
- Automated testing pipeline

### Production
- Kubernetes deployment with horizontal scaling
- Load balancer configuration for API traffic
- SSL/TLS termination
- Rate limiting at proxy level

## Timeline Estimation

### Phase 1 (2-3 weeks)
- Basic API endpoint implementation
- Request/response models
- Error handling middleware
- Unit tests

### Phase 2 (1-2 weeks) 
- OpenAPI specification
- Swagger documentation
- Validation middleware
- Integration tests

### Phase 3 (1-2 weeks)
- Rate limiting implementation
- Authentication framework
- Performance optimization
- Production deployment

**Total Estimated Time: 4-7 weeks**

## ✅ Success Criteria Checklist

- [ ] All existing functionality available via REST API
- [ ] OpenAPI specification with interactive documentation  
- [ ] Backward compatibility maintained with existing web interface
- [ ] Comprehensive error handling and validation implemented
- [ ] Rate limiting and basic security measures in place
- [ ] 95%+ test coverage for API endpoints achieved
- [ ] Performance equivalent to existing web interface
- [ ] Production-ready deployment configuration completed

## 📊 Progress Tracking

### Phase 1 Completion: __ / 19 tasks
### Phase 2 Completion: __ / 12 tasks  
### Phase 3 Completion: __ / 13 tasks
### Phase 4 Completion: __ / 8 tasks
### Phase 5 Completion: __ / 12 tasks (Optional)

**Overall Progress: __ / 64 total tasks**

**Estimated Timeline: 4-7 weeks**

---

This implementation plan leverages the existing domain logic in `app/internal/domains/message/ports/primary/message_service.go` while providing a clean REST interface that can coexist with the current web templates.