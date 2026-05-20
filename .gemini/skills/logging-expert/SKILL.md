---
name: logging-expert
description: Expert guidance for logging in the Password Exchange project. Use when adding or modifying Go code to ensure consistent structured logging using the project's Fluent API and Hexagonal LoggerPort patterns.
---

# Logging Expert

This skill ensures consistent, structured, and idiomatic logging across the Password Exchange Go services.

## Core Principles

1.  **Always use structured logging:** Never use `fmt.Println` or the standard `log` package.
2.  **Use the Fluent API:** Prefer the chained method calls for readability and consistency.
3.  **Context Matters:** Always include relevant metadata (IDs, errors, durations) using the provided enricher methods.
4.  **Hexagonal Alignment:** In domain layers, always use the `LoggerPort` interface rather than the global logger.

## Go Logging Patterns

### 1. Global Fluent API (`internal/shared/logging`)

For general application code, use the global logger provided in `app/internal/shared/logging`.

**Basic Usage:**

```go
import "github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"

// Simple info message
logging.Info().Msg("starting the web server")

// Error with metadata
if err != nil {
    logging.Error().
        Err(err).
        Str("component", "database").
        Int("retry_count", 3).
        Msg("failed to connect to database")
}
```

**Available Levels:**
- `logging.Debug()`
- `logging.Info()`
- `logging.Warn()`
- `logging.Error()`
- `logging.Fatal()` (calls `os.Exit(1)` after logging)

**Enricher Methods:**
- `.Err(error)`
- `.Str(key, value string)`
- `.Int(key string, value int)`
- `.Bool(key string, value bool)`
- `.Dur(key string, value time.Duration)`
- `.Interface(key string, value any)`

### 2. Hexagonal LoggerPort

When working within a domain (e.g., `internal/domains/notification`), you MUST use the `LoggerPort` interface. This allows the domain logic to remain independent of the specific logging implementation.

**Interface Definition (`ports/secondary/logger.go`):**

```go
type LoggerPort interface {
    Debug() contracts.LogEvent
    Info() contracts.LogEvent
    Warn() contracts.LogEvent
    Error() contracts.LogEvent
}
```

**Domain Usage Example:**

```go
func (s *Service) ProcessNotification(ctx context.Context, req contracts.NotificationRequest) error {
    s.logger.Info().
        Str("recipient", req.To).
        Msg("processing notification request")
        
    if err := s.emailPort.Send(ctx, req); err != nil {
        s.logger.Error().
            Err(err).
            Str("recipient", req.To).
            Msg("failed to send notification")
        return err
    }
    
    return nil
}
```

## Best Practices

- **Errors:** Always use `.Err(err)` when an error occurs. Don't just put the error message in `.Msg()`.
- **Naming:** Use `snake_case` for attribute keys (e.g., `user_id`, `request_duration_ms`).
- **Messages:** Keep `.Msg()` static and descriptive. Use attributes for dynamic data.
- **Leveling:** 
    - `Debug`: High-volume technical details.
    - `Info`: Key lifecycle events (starts, stops, successful requests).
    - `Warn`: Non-critical issues that might need attention.
    - `Error`: Serious issues requiring immediate investigation.
    - `Fatal`: System cannot continue.
