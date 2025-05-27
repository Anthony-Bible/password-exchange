package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

// MySQLAdapter implements the MessageRepository interface for MySQL
type MySQLAdapter struct {
	db     *sql.DB
	config domain.DatabaseConfig
}

// NewMySQLAdapter creates a new MySQL database adapter
func NewMySQLAdapter(config domain.DatabaseConfig) domain.MessageRepository {
	return &MySQLAdapter{
		config: config,
	}
}

// Connect establishes a connection to the MySQL database
func (m *MySQLAdapter) Connect() error {
	connectionString := fmt.Sprintf("%s:%s@(%s)/%s?parseTime=true",
		m.config.User, m.config.Password, m.config.Host, m.config.Name)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open MySQL connection")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseConnection, err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Error().Err(err).Msg("Failed to ping MySQL database")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseConnection, err)
	}

	m.db = db
	return nil
}

// InsertMessage stores a new encrypted message in the database
func (m *MySQLAdapter) InsertMessage(message *domain.Message) error {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return err
		}
	}

	// FIXED: Store recipient email in other_email field and passphrase in other_lastname field
	query := "INSERT INTO messages (message, uniqueid, other_lastname, other_email, view_count, max_view_count) VALUES (?, ?, ?, ?, 0, ?)"
	_, err := m.db.Exec(query, message.Content, message.UniqueID, message.Passphrase, message.RecipientEmail, message.MaxViewCount)
	if err != nil {
		log.Error().Err(err).Str("uniqueID", message.UniqueID).Msg("Failed to insert message")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Str("uniqueID", message.UniqueID).Int("maxViewCount", message.MaxViewCount).Str("recipientEmail", validation.SanitizeEmailForLogging(message.RecipientEmail)).Msg("Message stored successfully")
	return nil
}

// SelectMessageByUniqueID retrieves a message by its unique identifier
func (m *MySQLAdapter) SelectMessageByUniqueID(uniqueID string) (*domain.Message, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	query := "SELECT message, uniqueid, other_lastname, other_email, view_count, max_view_count FROM messages WHERE uniqueid = ?"
	row := m.db.QueryRow(query, uniqueID)

	var message domain.Message
	err := row.Scan(&message.Content, &message.UniqueID, &message.Passphrase, &message.RecipientEmail, &message.ViewCount, &message.MaxViewCount)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("uniqueID", uniqueID).Msg("Message not found")
			return nil, domain.ErrMessageNotFound
		}
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to select message")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Str("uniqueID", uniqueID).Msg("Message retrieved successfully")
	return &message, nil
}

// GetMessage retrieves a message by its unique identifier without incrementing the view count
func (m *MySQLAdapter) GetMessage(uniqueID string) (*domain.Message, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	query := "SELECT message, uniqueid, other_lastname, other_email, view_count, max_view_count FROM messages WHERE uniqueid = ?"
	row := m.db.QueryRow(query, uniqueID)

	var message domain.Message
	err := row.Scan(&message.Content, &message.UniqueID, &message.Passphrase, &message.RecipientEmail, &message.ViewCount, &message.MaxViewCount)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("uniqueID", uniqueID).Msg("Message not found")
			return nil, domain.ErrMessageNotFound
		}
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to select message")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Str("uniqueID", uniqueID).Msg("Message retrieved successfully")
	return &message, nil
}

// IncrementViewCountAndGet atomically increments the view count and returns the message
// If the view count reaches 5, the message is deleted
func (m *MySQLAdapter) IncrementViewCountAndGet(uniqueID string) (*domain.Message, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	// Start a transaction to ensure atomicity
	tx, err := m.db.Begin()
	if err != nil {
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to begin transaction")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	defer tx.Rollback() // Will be ignored if transaction is committed

	// First, increment the view count
	updateQuery := "UPDATE messages SET view_count = view_count + 1 WHERE uniqueid = ?"
	result, err := tx.Exec(updateQuery, uniqueID)
	if err != nil {
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to increment view count")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	// Check if the message exists
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to get rows affected")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	if rowsAffected == 0 {
		log.Debug().Str("uniqueID", uniqueID).Msg("Message not found for view count increment")
		return nil, domain.ErrMessageNotFound
	}

	// Get the updated message with the new view count
	selectQuery := "SELECT message, uniqueid, other_lastname, other_email, view_count, max_view_count FROM messages WHERE uniqueid = ?"
	row := tx.QueryRow(selectQuery, uniqueID)

	var message domain.Message
	err = row.Scan(&message.Content, &message.UniqueID, &message.Passphrase, &message.RecipientEmail, &message.ViewCount, &message.MaxViewCount)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().Str("uniqueID", uniqueID).Msg("Message not found after increment")
			return nil, domain.ErrMessageNotFound
		}
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to select message after increment")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	// If view count has reached max view count, delete the message
	if message.ViewCount >= message.MaxViewCount {
		deleteQuery := "DELETE FROM messages WHERE uniqueid = ?"
		_, err = tx.Exec(deleteQuery, uniqueID)
		if err != nil {
			log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to delete message after reaching view limit")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}
		log.Info().Str("uniqueID", uniqueID).Int("viewCount", message.ViewCount).Int("maxViewCount", message.MaxViewCount).Msg("Message deleted after reaching view limit")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to commit transaction")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Str("uniqueID", uniqueID).Int("viewCount", message.ViewCount).Msg("View count incremented successfully")
	return &message, nil
}

// DeleteExpiredMessages removes messages that have exceeded their TTL
func (m *MySQLAdapter) DeleteExpiredMessages() error {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return err
		}
	}

	// This is a placeholder - would need actual expiration logic based on business rules
	// For now, we'll assume messages older than 7 days should be deleted
	cutoffTime := time.Now().AddDate(0, 0, -7)

	query := "DELETE FROM messages WHERE created_at < ?"
	result, err := m.db.Exec(query, cutoffTime)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete expired messages")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info().Int64("rowsDeleted", rowsAffected).Msg("Expired messages cleaned up")
	return nil
}

// GetUnviewedMessagesForReminders retrieves messages that are unviewed and eligible for reminder emails
func (m *MySQLAdapter) GetUnviewedMessagesForReminders(olderThanHours, maxReminders int) ([]*domain.UnviewedMessage, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	query := `SELECT m.messageid, m.uniqueid, m.other_email, m.created,
         TIMESTAMPDIFF(DAY, m.created, NOW()) as days_old
  FROM messages m 
  LEFT JOIN email_reminders er ON m.messageid = er.message_id
  WHERE m.view_count = 0 
    AND m.created < NOW() - INTERVAL ? HOUR
    AND m.other_email IS NOT NULL
    AND m.other_email != ''
    AND (er.reminder_count IS NULL OR er.reminder_count < ?)`

	rows, err := m.db.Query(query, olderThanHours, maxReminders)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query unviewed messages for reminders")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var messages []*domain.UnviewedMessage
	for rows.Next() {
		var msg domain.UnviewedMessage
		var createdStr string
		err := rows.Scan(&msg.MessageID, &msg.UniqueID, &msg.RecipientEmail, &createdStr, &msg.DaysOld)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan unviewed message")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}

		// Parse the created time
		created, err := time.Parse("2006-01-02 15:04:05", createdStr)
		if err != nil {
			log.Error().Err(err).Str("createdStr", createdStr).Msg("Failed to parse created time")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}
		msg.Created = created

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error iterating over unviewed messages")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Int("count", len(messages)).Msg("Retrieved unviewed messages for reminders")
	return messages, nil
}

// LogReminderSent records that a reminder email was sent for a message
func (m *MySQLAdapter) LogReminderSent(messageID int, emailAddress string) error {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return err
		}
	}

	query := `INSERT INTO email_reminders (message_id, email_address, reminder_count, last_reminder_sent)
		VALUES (?, ?, 1, NOW())
		ON DUPLICATE KEY UPDATE 
		reminder_count = reminder_count + 1,
		last_reminder_sent = NOW()`

	_, err := m.db.Exec(query, messageID, emailAddress)
	if err != nil {
		log.Error().Err(err).Int("messageID", messageID).Str("emailAddress", validation.SanitizeEmailForLogging(emailAddress)).Msg("Failed to log reminder sent")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Int("messageID", messageID).Str("emailAddress", validation.SanitizeEmailForLogging(emailAddress)).Msg("Reminder sent logged successfully")
	return nil
}

// GetReminderHistory retrieves the reminder history for a specific message
func (m *MySQLAdapter) GetReminderHistory(messageID int) ([]*domain.ReminderLogEntry, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}

	query := `SELECT message_id, email_address, reminder_count, last_reminder_sent 
		FROM email_reminders WHERE message_id = ?`

	rows, err := m.db.Query(query, messageID)
	if err != nil {
		log.Error().Err(err).Int("messageID", messageID).Msg("Failed to query reminder history")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var history []*domain.ReminderLogEntry
	for rows.Next() {
		var entry domain.ReminderLogEntry
		err := rows.Scan(&entry.MessageID, &entry.EmailAddress, &entry.ReminderCount, &entry.LastReminderSent)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan reminder history entry")
			return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
		}
		history = append(history, &entry)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error iterating over reminder history")
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}

	log.Info().Int("messageID", messageID).Int("count", len(history)).Msg("Retrieved reminder history")
	return history, nil
}

// Close closes the database connection
func (m *MySQLAdapter) Close() error {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
			return err
		}
		m.db = nil
	}
	return nil
}
