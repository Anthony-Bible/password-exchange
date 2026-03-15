# Implementation Plan: Formalizing Migrations

## Phase 1: Setup and Infrastructure [checkpoint: e9d5635]
- [x] Task: Create `app/migrations` directory to store migration files. 01b3180
- [x] Task: Add `github.com/golang-migrate/migrate/v4` to `app/go.mod`. 0e8e8db
- [x] Task: Create a migration utility package in `app/internal/shared/database/migrations` to encapsulate `golang-migrate` logic. 01b3180 (original commit, refactored in e9d5635)
- [x] Task: Conductor - User Manual Verification 'Phase 1: Setup and Infrastructure' (Protocol in workflow.md) e9d5635

## Phase 2: Migration Consolidation [checkpoint: 16377f8]
- [x] Task: Rename and move `passwordexchange.sql` to `app/migrations/20260311000000_initial_schema.up.sql`. 34f09ae
- [x] Task: Create `app/migrations/20260311000000_initial_schema.down.sql` with the corresponding drop table commands. efb37bc
- [x] Task: Rename and move `migration_add_view_count.sql` to `app/migrations/20260311000001_add_view_count.up.sql`. 50688f2
- [x] Task: Create `app/migrations/20260311000001_add_view_count.down.sql`. 35e8313
- [x] Task: Rename and move `migration_add_max_view_count.sql` to `app/migrations/20260311000002_add_max_view_count.up.sql`. 3a93fa5
- [x] Task: Create `app/migrations/20260311000002_add_max_view_count.down.sql`. 727bb7a
- [x] Task: Rename and move `migration_add_expires_at.sql` to `app/migrations/20260311000003_add_expires_at.up.sql`. 1412c6d
- [x] Task: Create `app/migrations/20260311000003_add_expires_at.down.sql`. c3696a8
- [x] Task: Rename and move `migration_add_email_reminders.sql` to `app/migrations/20260311000004_add_email_reminders.up.sql`. 83dbf85
- [x] Task: Create `app/migrations/20260311000004_add_email_reminders.down.sql`. fe1d303
- [x] Task: Conductor - User Manual Verification 'Phase 2: Migration Consolidation' (Protocol in workflow.md) 16377f8


## Phase 3: CLI Command Implementation [checkpoint: 7857842]
- [x] Task: Define the `migrate` subcommand in `app/cmd/database/database.go` with `up`, `down`, `status`, and `create` subcommands. 7857842
- [x] Task: Use **red-phase-tester** to generate failing unit tests for the CLI commands (mocking the database connection and migration utility). 7857842
- [x] Task: Implement the migration logic in the subcommands using the utility package from Phase 1. 7857842
- [x] Task: Use **tdd-refactor-specialist** to clean up and refactor the CLI implementation while keeping tests green. 7857842
- [x] Task: Use **tdd-review-agent** to verify the implementation completeness and quality of the CLI commands. 7857842
- [x] Task: Conductor - User Manual Verification 'Phase 3: CLI Command Implementation' (Protocol in workflow.md) 7857842

## Phase 4: Auto-Execution in Database Service [checkpoint: 1c9387e]
- [x] Task: Use **red-phase-tester** to generate failing integration tests for auto-migration logic in the `database` service startup. 7857842 (original refactor in 7c52a1e, test in 7c52a1e)
- [x] Task: Update the `database` command's `Run` function (in `app/cmd/database/database.go`) to trigger the `up` migration before starting the gRPC server. 940627
- [x] Task: Implement configuration to enable/disable auto-migration. 948499
- [x] Task: Use **tdd-refactor-specialist** to refactor the startup integration while maintaining test coverage. 985625
- [x] Task: Use **tdd-review-agent** to ensure the auto-migration logic is robust and follows hexagonal architecture principles. 1c9387e
- [x] Task: Conductor - User Manual Verification 'Phase 4: Auto-Execution in Database Service' (Protocol in workflow.md) 1c9387e

## Phase 5: Verification and Cleanup [checkpoint: 00a8523]
- [x] Task: Verify the migration process on a clean database (using `app database migrate up`). 1005444
- [x] Task: Verify the migration process on an existing database (ensuring it handles the existing schema). 1006404
- [x] Task: Remove the original SQL files from the project root. b378f56
- [x] Task: Update project documentation (README or a dedicated migrations document) to reflect the new process. 1c9387e
- [x] Task: Conductor - User Manual Verification 'Phase 5: Verification and Cleanup' (Protocol in workflow.md) 00a8523
