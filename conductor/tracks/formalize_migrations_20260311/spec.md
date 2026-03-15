# Specification: Formalizing Migrations with golang-migrate

## Overview
This track aims to replace the current manual SQL script-based migration process with a formal, automated system using `golang-migrate`. It will provide both a CLI interface for manual management and automatic execution during service startup.

## Functional Requirements
- **Migration Consolidation**: All existing `.sql` migration scripts in the root directory will be moved and renamed to follow the `golang-migrate` convention (e.g., `YYYYMMDDHHMMSS_description.up.sql`).
- **CLI Management**: Implement subcommands under the `app database` command (or a new `app migrate` command) to handle:
  - `up`: Apply all pending migrations.
  - `down`: Roll back the last migration.
  - `status`: Show the current migration version and pending migrations.
  - `create`: Generate a new migration template file.
- **Auto-Execution**: The database service will automatically attempt to run pending migrations upon startup.
- **Version Tracking**: Use the standard `schema_migrations` table (default for `golang-migrate`) to track applied migrations.

## Non-Functional Requirements
- **Robustness**: Migration failures should prevent the database service from starting to avoid inconsistent states.
- **Compatibility**: Ensure existing data is preserved during the transition to the formal migration system.

## Acceptance Criteria
- [ ] Existing SQL scripts are moved to a dedicated directory (e.g., `app/migrations/`).
- [ ] CLI commands for `up`, `down`, `status`, and `create` are implemented and functional.
- [ ] Database service successfully runs pending migrations on startup.
- [ ] A `schema_migrations` table exists and correctly tracks migration history.
- [ ] Documentation is updated to explain how to create and manage migrations.

## Out of Scope
- Migrating non-SQL data.
- Handling database backups (this is assumed to be handled at the infrastructure level).
