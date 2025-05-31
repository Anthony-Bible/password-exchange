package mysql

import (
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLAdapter_InsertMessage_WithRecipientEmail(t *testing.T) {
	// Test that recipient email is properly stored in other_email field
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{db: db}

	// Test data
	message := &domain.Message{
		UniqueID:       "test-uuid-123",
		Content:        "encrypted-content",
		Passphrase:     "test-passphrase",
		MaxViewCount:   3,
		RecipientEmail: "test@example.com",
	}

	// Expected SQL should store recipient email in other_email field
	mock.ExpectExec(`INSERT INTO messages \(message, uniqueid, other_lastname, other_email, view_count, max_view_count\) VALUES \(\?, \?, \?, \?, 0, \?\)`).
		WithArgs(message.Content, message.UniqueID, message.Passphrase, message.RecipientEmail, message.MaxViewCount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	err = adapter.InsertMessage(message)

	// Assert
	if err != nil {
		t.Errorf("InsertMessage() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations were not met: %v", err)
	}
}

func TestMySQLAdapter_GetUnviewedMessagesForReminders_WithInterval(t *testing.T) {
	// Test that reminder interval is properly applied in the query
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{db: db}

	// Test parameters
	olderThanHours := 24
	maxReminders := 3
	reminderIntervalHours := 48

	// Expected columns for the query
	columns := []string{"messageid", "uniqueid", "other_email", "created", "days_old"}
	
	// Mock rows with test data
	rows := sqlmock.NewRows(columns).
		AddRow(123, "test-uuid-123", "test@example.com", time.Now().Add(-72*time.Hour), 3).
		AddRow(124, "test-uuid-124", "test2@example.com", time.Now().Add(-48*time.Hour), 2)

	// The query should include the reminder interval check
	mock.ExpectQuery(`SELECT m.messageid, m.uniqueid, m.other_email, m.created,
         TIMESTAMPDIFF\(DAY, m.created, NOW\(\)\) as days_old
  FROM messages m 
  LEFT JOIN email_reminders er ON m.messageid = er.message_id
  WHERE m.view_count = 0 
    AND m.created < NOW\(\) - INTERVAL \? HOUR
    AND m.other_email IS NOT NULL
    AND m.other_email != ''
    AND \(er.reminder_count IS NULL OR er.reminder_count < \?\)
    AND \(er.last_reminder_sent IS NULL OR er.last_reminder_sent < NOW\(\) - INTERVAL \? HOUR\)`).
		WithArgs(olderThanHours, maxReminders, reminderIntervalHours).
		WillReturnRows(rows)

	// Act
	messages, err := adapter.GetUnviewedMessagesForReminders(olderThanHours, maxReminders, reminderIntervalHours)

	// Assert
	if err != nil {
		t.Errorf("GetUnviewedMessagesForReminders() error = %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations were not met: %v", err)
	}
}

func TestMySQLAdapter_GetUnviewedMessagesForReminders(t *testing.T) {
	// Test that unviewed messages are properly retrieved for reminders
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{db: db}

	// Test data
	olderThanHours := 24
	maxReminders := 3

	// Mock rows - use time.Time for created column
	mockTime1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	mockTime2 := time.Date(2023, 1, 2, 15, 30, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"messageid", "uniqueid", "other_email", "created", "days_old"}).
		AddRow(1, "uuid-1", "user1@example.com", mockTime1, 2).
		AddRow(2, "uuid-2", "user2@example.com", mockTime2, 1)

	// Expected SQL query for unviewed messages
	expectedSQL := `SELECT m\.messageid, m\.uniqueid, m\.other_email, m\.created,
         TIMESTAMPDIFF\(DAY, m\.created, NOW\(\)\) as days_old
  FROM messages m 
  LEFT JOIN email_reminders er ON m\.messageid = er\.message_id
  WHERE m\.view_count = 0 
    AND m\.created < NOW\(\) - INTERVAL \? HOUR
    AND m\.other_email IS NOT NULL
    AND m\.other_email != ''
    AND \(er\.reminder_count IS NULL OR er\.reminder_count < \?\)
    AND \(er\.last_reminder_sent IS NULL OR er\.last_reminder_sent < NOW\(\) - INTERVAL \? HOUR\)`

	mock.ExpectQuery(expectedSQL).
		WithArgs(olderThanHours, maxReminders, 24).
		WillReturnRows(rows)

	// Act
	messages, err := adapter.GetUnviewedMessagesForReminders(olderThanHours, maxReminders, 24) // Default 24 hour interval

	// Assert
	if err != nil {
		t.Errorf("GetUnviewedMessagesForReminders() error = %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Verify first message
	if messages[0].MessageID != 1 {
		t.Errorf("Expected MessageID 1, got %d", messages[0].MessageID)
	}
	if messages[0].UniqueID != "uuid-1" {
		t.Errorf("Expected UniqueID 'uuid-1', got %s", messages[0].UniqueID)
	}
	if messages[0].RecipientEmail != "user1@example.com" {
		t.Errorf("Expected RecipientEmail 'user1@example.com', got %s", messages[0].RecipientEmail)
	}
	if messages[0].DaysOld != 2 {
		t.Errorf("Expected DaysOld 2, got %d", messages[0].DaysOld)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations were not met: %v", err)
	}
}

func TestMySQLAdapter_LogReminderSent(t *testing.T) {
	// Test that reminder attempts are properly logged
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{db: db}

	messageID := 1
	emailAddress := "user@example.com"

	// Expected SQL for inserting or updating reminder log
	mock.ExpectExec(`INSERT INTO email_reminders \(message_id, email_address, reminder_count, last_reminder_sent\)
		VALUES \(\?, \?, 1, NOW\(\)\)
		ON DUPLICATE KEY UPDATE 
		reminder_count = reminder_count \+ 1,
		last_reminder_sent = NOW\(\)`).
		WithArgs(messageID, emailAddress).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	err = adapter.LogReminderSent(messageID, emailAddress)

	// Assert
	if err != nil {
		t.Errorf("LogReminderSent() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations were not met: %v", err)
	}
}

func TestMySQLAdapter_GetReminderHistory(t *testing.T) {
	// Test that reminder history is properly retrieved
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	adapter := &MySQLAdapter{db: db}

	messageID := 1

	// Mock rows
	rows := sqlmock.NewRows([]string{"message_id", "email_address", "reminder_count", "last_reminder_sent"}).
		AddRow(1, "user@example.com", 2, time.Now())

	mock.ExpectQuery(`SELECT message_id, email_address, reminder_count, last_reminder_sent 
		FROM email_reminders WHERE message_id = \?`).
		WithArgs(messageID).
		WillReturnRows(rows)

	// Act
	history, err := adapter.GetReminderHistory(messageID)

	// Assert
	if err != nil {
		t.Errorf("GetReminderHistory() error = %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}

	if history[0].MessageID != 1 {
		t.Errorf("Expected MessageID 1, got %d", history[0].MessageID)
	}
	if history[0].EmailAddress != "user@example.com" {
		t.Errorf("Expected EmailAddress 'user@example.com', got %s", history[0].EmailAddress)
	}
	if history[0].ReminderCount != 2 {
		t.Errorf("Expected ReminderCount 2, got %d", history[0].ReminderCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations were not met: %v", err)
	}
}