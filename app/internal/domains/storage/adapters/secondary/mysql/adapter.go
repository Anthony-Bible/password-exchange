package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
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
func (m *MySQLAdapter) InsertMessage(content, uniqueID, passphrase string) error {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return err
		}
	}
	
	query := "INSERT INTO messages (message, uniqueid, other_lastname) VALUES (?, ?, ?)"
	_, err := m.db.Exec(query, content, uniqueID, passphrase)
	if err != nil {
		log.Error().Err(err).Str("uniqueID", uniqueID).Msg("Failed to insert message")
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOperation, err)
	}
	
	log.Info().Str("uniqueID", uniqueID).Msg("Message stored successfully")
	return nil
}

// SelectMessageByUniqueID retrieves a message by its unique identifier
func (m *MySQLAdapter) SelectMessageByUniqueID(uniqueID string) (*domain.Message, error) {
	if m.db == nil {
		if err := m.Connect(); err != nil {
			return nil, err
		}
	}
	
	query := "SELECT message, uniqueid, other_lastname FROM messages WHERE uniqueid = ?"
	row := m.db.QueryRow(query, uniqueID)
	
	var message domain.Message
	err := row.Scan(&message.Content, &message.UniqueID, &message.Passphrase)
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