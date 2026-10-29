package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/dsn"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// defaultMaxOpenConns is the pool size used when DatabaseConfig.MaxOpenConns is unset.
	defaultMaxOpenConns = 5

	// connectPingTimeout bounds the initial connectivity check in Connect.
	connectPingTimeout = 5 * time.Second

	// defaultMessageTTL is the default time-to-live for messages stored in Postgres.
	// Must match the message domain's DefaultMessageTTL.
	defaultMessageTTL = 7 * 24 * time.Hour
)

// selectMessageQuery is the standard SELECT statement used to retrieve a message row.
const selectMessageQuery = "SELECT message, uniqueid, is_client_encrypted, other_lastname, other_email, view_count, max_view_count, expires_at FROM messages WHERE uniqueid = $1"

// scanMessageRow scans a single message row into a contracts.Message, handling the nullable expires_at field.
func scanMessageRow(row *sql.Row) (*contracts.Message, error) {
	var message contracts.Message
	var expiresAt sql.NullTime
	err := row.Scan(
		&message.Content,
		&message.UniqueID,
		&message.IsClientEncrypted,
		&message.Passphrase,
		&message.RecipientEmail,
		&message.ViewCount,
		&message.MaxViewCount,
		&expiresAt,
	)
	if err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		message.ExpiresAt = &expiresAt.Time
	}
	return &message, nil
}

// PostgresAdapter implements the MessageRepository interface for Postgres.
// It depends only on secondary ports (LoggerPort, ValidationPort) so it can
// be exercised without touching the shared logging/validation globals.
type PostgresAdapter struct {
	db        *sql.DB
	config    contracts.DatabaseConfig
	logger    secondary.LoggerPort
	validator secondary.ValidationPort
	closed    atomic.Bool
	connMu    sync.Mutex
	opMu      sync.RWMutex
}

// NewPostgresAdapter creates a new Postgres database adapter. The logger and
// validator are required; passing nil for either panics so wiring mistakes
// surface at boot instead of as a nil deref deep inside a query.
func NewPostgresAdapter(
	config contracts.DatabaseConfig,
	logger secondary.LoggerPort,
	validator secondary.ValidationPort,
) secondary.MessageRepository {
	if logger == nil {
		panic("storage/postgres: NewPostgresAdapter requires a non-nil LoggerPort")
	}
	if validator == nil {
		panic("storage/postgres: NewPostgresAdapter requires a non-nil ValidationPort")
	}
	return &PostgresAdapter{
		config:    config,
		logger:    logger,
		validator: validator,
	}
}

// requireOpen enforces the MessageRepository.Close contract: once Close has
// been called, no further operations are permitted. Otherwise it lazily
// connects on first use.
func (m *PostgresAdapter) requireOpen() (*sql.DB, func(), error) {
	if m.closed.Load() {
		return nil, nil, domain.ErrRepositoryClosed
	}
	if err := m.Connect(); err != nil {
		return nil, nil, err
	}
	// Hold the read lock for the whole operation so Close cannot tear the
	// pool down underneath an in-flight call.
	m.opMu.RLock()
	m.connMu.Lock()
	db := m.db
	m.connMu.Unlock()
	if m.closed.Load() || db == nil {
		m.opMu.RUnlock()
		return nil, nil, domain.ErrRepositoryClosed
	}
	return db, m.opMu.RUnlock, nil
}

// Connect establishes a connection to the Postgres database
// It is idempotent and safe for concurrent use: only one pool is ever opened.
func (m *PostgresAdapter) Connect() error {
	m.connMu.Lock()
	defer m.connMu.Unlock()
	if m.closed.Load() {
		return domain.ErrRepositoryClosed
	}
	if m.db != nil {
		return nil
	}
	db, err := sql.Open("pgx", dsn.PostgresURL(m.config.Host, m.config.Port, m.config.User, m.config.Password, m.config.Name))
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to open Postgres connection")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseConnection, err)
	}

	maxOpen := m.config.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = defaultMaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Test the connection; bounded so a hung database cannot hold connMu indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), connectPingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close() // Close connection if ping fails to prevent leak
		m.logger.Error().Err(err).Msg("Failed to ping Postgres database")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseConnection, err)
	}

	m.db = db
	return nil
}

// InsertMessage stores a new encrypted message in the database
func (m *PostgresAdapter) InsertMessage(message *contracts.Message) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	// Column mapping: recipient email -> other_email, passphrase -> other_lastname.
	var expiresAt time.Time
	if message.ExpiresAt != nil {
		expiresAt = *message.ExpiresAt
	} else {
		expiresAt = time.Now().Add(defaultMessageTTL)
	}
	query := "INSERT INTO messages (message, uniqueid, is_client_encrypted, other_lastname, other_email, view_count, max_view_count, expires_at) VALUES ($1, $2, $3, $4, $5, 0, $6, $7)"
	_, err := db.Exec(
		query,
		message.Content,
		message.UniqueID,
		message.IsClientEncrypted,
		message.Passphrase,
		message.RecipientEmail,
		message.MaxViewCount,
		expiresAt,
	)
	if err != nil {
		if isInvalidInputError(err) {
			m.logger.Warn().Err(err).Str("uniqueID", message.UniqueID).Msg("Rejected message with invalid input")
			return fmt.Errorf("%w: %v", domain.ErrInvalidParameter, err)
		}
		m.logger.Error().Err(err).Str("uniqueID", message.UniqueID).Msg("Failed to insert message")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	m.logger.Info().
		Str("uniqueID", message.UniqueID).
		Int("maxViewCount", message.MaxViewCount).
		Str("recipientEmail", m.validator.SanitizeEmailForLogging(message.RecipientEmail)).
		Msg("Message stored successfully")
	return nil
}

// SelectMessageByUniqueID retrieves a message by its unique identifier
func (m *PostgresAdapter) SelectMessageByUniqueID(uniqueID string) (*contracts.Message, error) {
	return m.fetchMessage(uniqueID)
}

// GetMessage retrieves a message by its unique identifier without incrementing the view count
func (m *PostgresAdapter) GetMessage(uniqueID string) (*contracts.Message, error) {
	return m.fetchMessage(uniqueID)
}

// fetchMessage is the shared read-only lookup behind SelectMessageByUniqueID and GetMessage.
func (m *PostgresAdapter) fetchMessage(uniqueID string) (*contracts.Message, error) {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return nil, openErr
	}
	defer release()

	message, err := scanMessageRow(db.QueryRow(selectMessageQuery, uniqueID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			m.logger.Debug().Str("uniqueID", uniqueID).Msg("Message not found")
			return nil, domain.ErrMessageNotFound
		}
		m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to select message")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	m.logger.Info().Str("uniqueID", uniqueID).Msg("Message retrieved successfully")
	return message, nil
}

// IncrementViewCountAndGet atomically increments the view count and returns the message
// If the view count reaches the message's MaxViewCount, the message is deleted
func (m *PostgresAdapter) IncrementViewCountAndGet(uniqueID string) (*contracts.Message, error) {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return nil, openErr
	}
	defer release()

	// Start a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to begin transaction")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	defer tx.Rollback() // Will be ignored if transaction is committed

	// First, increment the view count
	updateQuery := "UPDATE messages SET view_count = view_count + 1 WHERE uniqueid = $1"
	result, err := tx.Exec(updateQuery, uniqueID)
	if err != nil {
		m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to increment view count")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	// Check if the message exists
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to get rows affected")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	if rowsAffected == 0 {
		m.logger.Debug().Str("uniqueID", uniqueID).Msg("Message not found for view count increment")
		return nil, domain.ErrMessageNotFound
	}

	// Get the updated message with the new view count
	message, err := scanMessageRow(tx.QueryRow(selectMessageQuery, uniqueID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			m.logger.Debug().Str("uniqueID", uniqueID).Msg("Message not found after increment")
			return nil, domain.ErrMessageNotFound
		}
		m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to select message after increment")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	// If view count has reached max view count, delete the message
	if message.ViewCount >= message.MaxViewCount {
		deleteQuery := "DELETE FROM messages WHERE uniqueid = $1"
		_, err = tx.Exec(deleteQuery, uniqueID)
		if err != nil {
			m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to delete message after reaching view limit")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}
		m.logger.Info().
			Str("uniqueID", uniqueID).
			Int("viewCount", message.ViewCount).
			Int("maxViewCount", message.MaxViewCount).
			Msg("Message deleted after reaching view limit")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		m.logger.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to commit transaction")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	m.logger.Info().Str("uniqueID", uniqueID).Int("viewCount", message.ViewCount).Msg("View count incremented successfully")
	return message, nil
}

// DeleteExpiredMessages removes messages that have exceeded their TTL
func (m *PostgresAdapter) DeleteExpiredMessages() error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	query := "DELETE FROM messages WHERE expires_at < NOW()"
	result, err := db.Exec(query)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to delete expired messages")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		m.logger.Warn().Err(err).Msg("Expired messages cleaned up; rows affected unavailable")
		return nil
	}
	m.logger.Info().Int64("rowsDeleted", rowsAffected).Msg("Expired messages cleaned up")
	return nil
}

// GetUnviewedMessagesForReminders retrieves messages that are unviewed and eligible for reminder emails
func (m *PostgresAdapter) GetUnviewedMessagesForReminders(
	olderThanHours, maxReminders, reminderIntervalHours int,
) ([]*contracts.UnviewedMessage, error) {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return nil, openErr
	}
	defer release()

	query := `SELECT m.messageid, m.uniqueid, m.other_email, m.created,
         EXTRACT(DAY FROM now() - m.created)::int as days_old
  FROM messages m 
  LEFT JOIN email_reminders er ON m.messageid = er.message_id
  WHERE m.view_count = 0 
    AND m.created < now() - make_interval(hours => $1::int)
    AND m.other_email IS NOT NULL
    AND m.other_email != ''
    AND (er.reminder_count IS NULL OR er.reminder_count < $2)
    AND (er.last_reminder_sent IS NULL OR er.last_reminder_sent < now() - make_interval(hours => $3::int))`

	rows, err := db.Query(query, olderThanHours, maxReminders, reminderIntervalHours)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to query unviewed messages for reminders")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var messages []*contracts.UnviewedMessage
	for rows.Next() {
		var msg contracts.UnviewedMessage
		err := rows.Scan(&msg.MessageID, &msg.UniqueID, &msg.RecipientEmail, &msg.Created, &msg.DaysOld)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to scan unviewed message")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		m.logger.Error().Err(err).Msg("Error iterating over unviewed messages")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	m.logger.Info().Int("count", len(messages)).Msg("Retrieved unviewed messages for reminders")
	return messages, nil
}

// LogReminderSent records that a reminder email was sent for a message
func (m *PostgresAdapter) LogReminderSent(messageID int, emailAddress string) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()

	query := `INSERT INTO email_reminders (message_id, email_address, reminder_count, last_reminder_sent)
		VALUES ($1, $2, 1, now())
		ON CONFLICT (message_id, email_address) DO UPDATE
		SET reminder_count = email_reminders.reminder_count + 1,
		last_reminder_sent = now()`

	_, err := db.Exec(query, messageID, emailAddress)
	if err != nil {
		m.logger.Error().
			Err(err).
			Int("messageID", messageID).
			Str("emailAddress", m.validator.SanitizeEmailForLogging(emailAddress)).
			Msg("Failed to log reminder sent")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	m.logger.Info().
		Int("messageID", messageID).
		Str("emailAddress", m.validator.SanitizeEmailForLogging(emailAddress)).
		Msg("Reminder sent logged successfully")
	return nil
}

// GetReminderHistory retrieves the reminder history for a specific message
func (m *PostgresAdapter) GetReminderHistory(messageID int) ([]*contracts.ReminderLogEntry, error) {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return nil, openErr
	}
	defer release()

	query := `SELECT message_id, email_address, reminder_count, last_reminder_sent 
		FROM email_reminders WHERE message_id = $1`

	rows, err := db.Query(query, messageID)
	if err != nil {
		m.logger.Error().Err(err).Int("messageID", messageID).Msg("Failed to query reminder history")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var history []*contracts.ReminderLogEntry
	for rows.Next() {
		var entry contracts.ReminderLogEntry
		err := rows.Scan(&entry.MessageID, &entry.EmailAddress, &entry.ReminderCount, &entry.LastReminderSent)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to scan reminder history entry")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}
		history = append(history, &entry)
	}

	if err = rows.Err(); err != nil {
		m.logger.Error().Err(err).Msg("Error iterating over reminder history")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	m.logger.Info().Int("messageID", messageID).Int("count", len(history)).Msg("Retrieved reminder history")
	return history, nil
}

// Ping verifies the underlying *sql.DB pool can reach Postgres. The supplied
// context bounds the call so callers (notably the gRPC health-status loop)
// don't wedge waiting on a hung database.
func (m *PostgresAdapter) Ping(ctx context.Context) error {
	db, release, openErr := m.requireOpen()
	if openErr != nil {
		return openErr
	}
	defer release()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrDatabaseConnection, err)
	}
	return nil
}

// Close closes the database connection and marks the adapter as closed so
// no further operations can resurrect the pool. Calling Close more than once
// is a no-op.
func (m *PostgresAdapter) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}
	m.opMu.Lock()
	defer m.opMu.Unlock()
	m.connMu.Lock()
	defer m.connMu.Unlock()
	if m.db == nil {
		return nil
	}
	if err := m.db.Close(); err != nil {
		m.logger.Error().Err(err).Msg("Failed to close database connection")
		return err
	}
	m.db = nil
	return nil
}

// isInvalidInputError reports whether err is a Postgres string-too-long (22001),
// invalid-byte-sequence (22021, including NUL) or untranslatable-character (22P05) error.
func isInvalidInputError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.Code {
	case "22001", "22021", "22P05":
		return true
	}
	return false
}
