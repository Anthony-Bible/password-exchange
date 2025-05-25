# View Count Feature Deployment Guide

This document outlines the changes made to implement automatic deletion of encrypted content after 5 views and the steps required for deployment.

## Feature Overview

- Messages are now automatically deleted after being viewed 5 times
- View counting is atomic and handles concurrent access safely  
- Users see current view count and remaining views in the UI
- Database uses transactions to prevent race conditions

## Database Migration Required

Before deploying this feature, you MUST run the database migration:

```sql
-- Run this against your production database
ALTER TABLE messages 
ADD COLUMN view_count INT NOT NULL DEFAULT 0;

CREATE INDEX idx_messages_view_count ON messages(view_count);

UPDATE messages SET view_count = 0 WHERE view_count IS NULL;
```

**Note:** The migration is in `migration_add_view_count.sql`

## Breaking Changes

⚠️ **IMPORTANT:** This is a breaking change that requires:

1. **Database migration** (see above)
2. **Protobuf regeneration** - The gRPC interface has been updated
3. **Coordinated deployment** - All services must be updated together

## Files Modified

### Core Implementation
- `app/internal/domains/storage/domain/entities.go` - Added ViewCount field to Message entity
- `app/internal/domains/storage/domain/service.go` - Updated to use atomic view counting
- `app/internal/domains/storage/adapters/secondary/mysql/adapter.go` - Implemented atomic increment and deletion
- `protos/database.proto` - Added view_count field to SelectResponse

### Service Layer Updates  
- `app/internal/domains/message/domain/message_entities.go` - Added ViewCount to response types
- `app/internal/domains/message/domain/message_service.go` - Updated to pass ViewCount
- `app/internal/domains/message/adapters/secondary/grpc_clients/storage_client.go` - Handle ViewCount field
- `app/internal/domains/storage/adapters/primary/grpc/server.go` - Return ViewCount in gRPC responses

### Web Interface Updates
- `app/internal/domains/message/adapters/primary/web/handlers.go` - Pass ViewCount to templates
- `app/internal/domains/message/adapters/primary/web/server.go` - Added template functions
- `app/templates/home.html` - Updated messaging to reflect 5-view limit
- `app/templates/confirmation.html` - Updated security notes
- `app/templates/decryption.html` - Dynamic view count display

## Deployment Steps

1. **Stop all services** to prevent data inconsistency
2. **Run database migration** using `migration_add_view_count.sql`
3. **Regenerate protobuf files** using `./test-build.sh`
4. **Deploy all services** with the new code
5. **Verify functionality** with test messages

## Testing the Feature

After deployment, test with a new message:

1. Create a message through the web interface
2. Access the decrypt URL - should show "viewed 1 time(s)"
3. Access 4 more times - should show increasing count
4. On the 5th access, message should be deleted
5. 6th access should return 404/not found

## Rollback Plan

If rollback is necessary:

1. Deploy previous version of code
2. Database rollback: `ALTER TABLE messages DROP COLUMN view_count;`

**Note:** Rolling back will lose view count data for existing messages but won't break functionality.