package database

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMigrator is a mock of the migrations.IMigrator interface.
type MockMigrator struct {
	mock.Mock
}

func (m *MockMigrator) Up() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMigrator) Down() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMigrator) Version() (uint, bool, error) {
	args := m.Called()
	return args.Get(0).(uint), args.Bool(1), args.Error(2)
}

func (m *MockMigrator) Create(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockMigrator) Force(version int) error {
	args := m.Called(version)
	return args.Error(0)
}

func (m *MockMigrator) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMigrator) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestMigrateUpCommand(t *testing.T) {
	mockMigrator := new(MockMigrator)
	mockMigrator.On("Up").Return(nil)

	// Inject mock
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	b := bytes.NewBufferString("")
	migrateUpCmd.SetOut(b)
	migrateUpCmd.SetArgs([]string{})

	err := migrateUpCmd.RunE(migrateUpCmd, []string{})
	require.NoError(t, err)

	mockMigrator.AssertExpectations(t)
}

func TestMigrateUpCommandError(t *testing.T) {
	mockMigrator := new(MockMigrator)
	mockMigrator.On("Up").Return(errors.New("migration failed"))

	// Inject mock
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	err := migrateUpCmd.RunE(migrateUpCmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "migration failed")

	mockMigrator.AssertExpectations(t)
}

func TestMigrateDownCommand(t *testing.T) {
	mockMigrator := new(MockMigrator)
	mockMigrator.On("Down").Return(nil)

	// Inject mock
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	err := migrateDownCmd.RunE(migrateDownCmd, []string{})
	require.NoError(t, err)

	mockMigrator.AssertExpectations(t)
}

func TestMigrateStatusCommand(t *testing.T) {
	mockMigrator := new(MockMigrator)
	mockMigrator.On("Version").Return(uint(1), false, nil)

	// Inject mock
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	err := migrateStatusCmd.RunE(migrateStatusCmd, []string{})
	require.NoError(t, err)

	mockMigrator.AssertExpectations(t)
}

func TestMigrateCreateCommand(t *testing.T) {
	mockMigrator := new(MockMigrator)
	migrationName := "test_migration"
	mockMigrator.On("Create", migrationName).Return(nil)

	// Inject mock
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	err := migrateCreateCmd.RunE(migrateCreateCmd, []string{migrationName})
	require.NoError(t, err)

	mockMigrator.AssertExpectations(t)
}

func TestDatabaseCmdRun_RunsMigrationsOnStartup(t *testing.T) {
	mockMigrator := new(MockMigrator)
	mockMigrator.On("Up").Return(nil)

	// Inject mock migrator
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	// Enable auto-migrate
	oldCfg := cfg
	cfg.PassConfig.AutoMigrate = true
	defer func() { cfg = oldCfg }()

	// Mock the server start function to prevent blocking and verify call
	serverStarted := false
	oldRunDatabaseServer := runDatabaseServer
	runDatabaseServer = func() {
		serverStarted = true
	}
	defer func() { runDatabaseServer = oldRunDatabaseServer }()

	// Execute the command
	databaseCmd.Run(databaseCmd, []string{})

	// Verify Up() was called
	mockMigrator.AssertExpectations(t)
	assert.True(t, serverStarted, "Server should have been started after migrations")
}

func TestDatabaseCmdRun_NoMigrationsWhenDisabled(t *testing.T) {
	mockMigrator := new(MockMigrator)
	// We don't expect Up() to be called

	// Inject mock migrator
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	// Disable auto-migrate
	oldCfg := cfg
	cfg.PassConfig.AutoMigrate = false
	defer func() { cfg = oldCfg }()

	// Mock server start
	serverStarted := false
	oldRunDatabaseServer := runDatabaseServer
	runDatabaseServer = func() {
		serverStarted = true
	}
	defer func() { runDatabaseServer = oldRunDatabaseServer }()

	// Execute
	databaseCmd.Run(databaseCmd, []string{})

	// Verify Up() was NOT called
	mockMigrator.AssertNotCalled(t, "Up")
	assert.True(t, serverStarted, "Server should have been started without migrations")
}

func TestDatabaseCmdRun_MigrationsBeforeServerStart(t *testing.T) {
	mockMigrator := new(MockMigrator)

	var callOrder []string

	mockMigrator.On("Up").Return(nil).Run(func(args mock.Arguments) {
		callOrder = append(callOrder, "migrations")
	})

	// Inject mock migrator
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	// Enable auto-migrate
	oldCfg := cfg
	cfg.PassConfig.AutoMigrate = true
	defer func() { cfg = oldCfg }()

	// Mock server start
	oldRunDatabaseServer := runDatabaseServer
	runDatabaseServer = func() {
		callOrder = append(callOrder, "server")
	}
	defer func() { runDatabaseServer = oldRunDatabaseServer }()

	// Execute
	databaseCmd.Run(databaseCmd, []string{})

	// Verify order
	require.Len(t, callOrder, 2, "Both migrations and server should be called")
	assert.Equal(t, "migrations", callOrder[0], "Migrations should be called first")
	assert.Equal(t, "server", callOrder[1], "Server should be called after migrations")
}

func TestDatabaseCmdRun_NoServerOnMigrationFailure(t *testing.T) {
	mockMigrator := new(MockMigrator)
	mockMigrator.On("Up").Return(errors.New("migration error"))

	// Inject mock migrator
	oldMigrator := migrator
	migrator = mockMigrator
	defer func() { migrator = oldMigrator }()

	// Enable auto-migrate
	oldCfg := cfg
	cfg.PassConfig.AutoMigrate = true
	defer func() { cfg = oldCfg }()

	// Mock server start
	serverStarted := false
	oldRunDatabaseServer := runDatabaseServer
	runDatabaseServer = func() {
		serverStarted = true
	}
	defer func() { runDatabaseServer = oldRunDatabaseServer }()

	// Execute
	databaseCmd.Run(databaseCmd, []string{})

	// Verify server NOT started on failure
	assert.False(t, serverStarted, "Server should NOT have been started if migration failed")
}
