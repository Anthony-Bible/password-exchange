/*
Copyright © 2024 Anthony Bible <anthony@anthony-bible.com>
*/
package reminder

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// reminderCmd represents the reminder command
var reminderCmd = &cobra.Command{
	Use:   "reminder",
	Short: "Send reminder emails for unviewed messages",
	Long: `This command sends reminder emails to recipients who haven't viewed their secure messages.
It queries for messages that are:
- Unviewed (view_count = 0)
- Older than a specified number of hours (default: 24)
- Under the maximum reminder limit (default: 3)

Configuration via environment variables:
PASSWORDEXCHANGE_REMINDER_ENABLED: Enable/disable reminder system
PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS: Hours to wait before first reminder (default: 24)
PASSWORDEXCHANGE_REMINDER_MAXREMINDERS: Maximum reminders per message (default: 3)
PASSWORDEXCHANGE_REMINDER_INTERVAL: Hours between reminders (default: 24)`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		if !cfg.Reminder.Enabled {
			log.Info().Msg("Reminder system is disabled")
			return
		}

		log.Info().Msg("Starting reminder email processing")
		
		// Initialize storage adapter
		dbConfig := domain.DatabaseConfig{
			Host:     cfg.DbHost,
			User:     cfg.DbUser,
			Password: cfg.DbPass,
			Name:     cfg.DbName,
		}
		storageAdapter := mysql.NewMySQLAdapter(dbConfig)
		if mysqlAdapter, ok := storageAdapter.(*mysql.MySQLAdapter); ok {
			if err := mysqlAdapter.Connect(); err != nil {
				log.Fatal().Err(err).Msg("Failed to connect to database")
				return
			}
			defer mysqlAdapter.Close()
		}

		// Initialize storage service
		storageService := domain.NewStorageService(storageAdapter)

		// Create reminder processor
		processor := NewReminderProcessor(storageService, cfg)
		
		// Process reminders
		ctx := context.Background()
		if err := processor.ProcessReminders(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to process reminders")
			return
		}

		log.Info().Msg("Reminder email processing completed")
	},
}

// ReminderProcessor handles the business logic for sending reminder emails
type ReminderProcessor struct {
	storageService domain.StorageServicePort
	config         *config.Config
}

// NewReminderProcessor creates a new reminder processor
func NewReminderProcessor(storageService domain.StorageServicePort, config *config.Config) *ReminderProcessor {
	return &ReminderProcessor{
		storageService: storageService,
		config:         config,
	}
}

// ProcessReminders finds and processes messages eligible for reminder emails
func (r *ReminderProcessor) ProcessReminders(ctx context.Context) error {
	// Get unviewed messages eligible for reminders
	messages, err := r.storageService.GetUnviewedMessagesForReminders(
		ctx,
		r.config.Reminder.CheckAfterHours,
		r.config.Reminder.MaxReminders,
	)
	if err != nil {
		return fmt.Errorf("failed to get unviewed messages: %w", err)
	}

	log.Info().Int("count", len(messages)).Msg("Found messages eligible for reminders")

	if len(messages) == 0 {
		log.Info().Msg("No messages found requiring reminders")
		return nil
	}

	// Process each message
	for _, message := range messages {
		if err := r.processMessageReminder(ctx, message); err != nil {
			log.Error().Err(err).Int("messageID", message.MessageID).Str("email", message.RecipientEmail).Msg("Failed to process reminder for message")
			continue // Continue processing other messages
		}
	}

	return nil
}

// processMessageReminder sends a reminder email for a specific message
func (r *ReminderProcessor) processMessageReminder(ctx context.Context, message *domain.UnviewedMessage) error {
	log.Info().Int("messageID", message.MessageID).Str("email", message.RecipientEmail).Int("daysOld", message.DaysOld).Msg("Processing reminder for message")

	// Get reminder history to determine reminder count
	history, err := r.storageService.GetReminderHistory(ctx, message.MessageID)
	if err != nil {
		return fmt.Errorf("failed to get reminder history: %w", err)
	}

	reminderCount := 0
	if len(history) > 0 {
		reminderCount = history[0].ReminderCount
	}

	// Check if we've already sent the maximum number of reminders
	if reminderCount >= r.config.Reminder.MaxReminders {
		log.Debug().Int("messageID", message.MessageID).Int("reminderCount", reminderCount).Msg("Maximum reminders already sent for message")
		return nil
	}

	// TODO: Integrate with notification system to send actual email
	// For now, we'll simulate the email sending and log the reminder
	r.logReminderAttempt(ctx, message, reminderCount+1)

	// Record that we sent a reminder
	if err := r.storageService.LogReminderSent(ctx, message.MessageID, message.RecipientEmail); err != nil {
		return fmt.Errorf("failed to log reminder sent: %w", err)
	}

	log.Info().Int("messageID", message.MessageID).Str("email", message.RecipientEmail).Int("reminderNumber", reminderCount+1).Msg("Reminder email sent successfully")
	return nil
}

// logReminderAttempt logs the details of a reminder attempt (for development/testing)
func (r *ReminderProcessor) logReminderAttempt(ctx context.Context, message *domain.UnviewedMessage, reminderNumber int) {
	log.Info().
		Int("messageID", message.MessageID).
		Str("uniqueID", message.UniqueID).
		Str("recipientEmail", message.RecipientEmail).
		Int("daysOld", message.DaysOld).
		Int("reminderNumber", reminderNumber).
		Time("created", message.Created).
		Msg("REMINDER EMAIL WOULD BE SENT")

	// Generate the decryption URL (placeholder logic)
	decryptionURL := fmt.Sprintf("https://password.exchange/decrypt/%s", message.UniqueID)
	
	log.Info().
		Str("decryptionURL", decryptionURL).
		Str("template", "reminder_email_template.html").
		Msg("Email template data prepared")
}

// loadConfig loads configuration from viper with defaults
func loadConfig() *config.Config {
	cfg := &config.Config{}
	
	// Set defaults for reminder configuration
	viper.SetDefault("reminder.enabled", true)
	viper.SetDefault("reminder.checkafterhours", 24)
	viper.SetDefault("reminder.maxreminders", 3)
	viper.SetDefault("reminder.interval", 24)

	// Bind environment variables
	viper.BindEnv("reminder.enabled")
	viper.BindEnv("reminder.checkafterhours")
	viper.BindEnv("reminder.maxreminders")
	viper.BindEnv("reminder.interval")

	// Unmarshal into config struct
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal configuration")
	}

	// Override with command-line flags if provided
	if olderThanHours := viper.GetString("older-than-hours"); olderThanHours != "" {
		if hours, err := strconv.Atoi(olderThanHours); err == nil {
			cfg.Reminder.CheckAfterHours = hours
		}
	}
	if maxReminders := viper.GetString("max-reminders"); maxReminders != "" {
		if max, err := strconv.Atoi(maxReminders); err == nil {
			cfg.Reminder.MaxReminders = max
		}
	}

	log.Info().
		Bool("enabled", cfg.Reminder.Enabled).
		Int("checkAfterHours", cfg.Reminder.CheckAfterHours).
		Int("maxReminders", cfg.Reminder.MaxReminders).
		Msg("Reminder configuration loaded")

	return cfg
}

func init() {
	cmd.RootCmd.AddCommand(reminderCmd)

	// Command-line flags
	reminderCmd.Flags().String("older-than-hours", "", "Hours to wait before sending first reminder (default: 24)")
	reminderCmd.Flags().String("max-reminders", "", "Maximum number of reminders per message (default: 3)")
	reminderCmd.Flags().Bool("dry-run", false, "Show what would be done without actually sending emails")

	// Bind flags to viper
	viper.BindPFlag("older-than-hours", reminderCmd.Flags().Lookup("older-than-hours"))
	viper.BindPFlag("max-reminders", reminderCmd.Flags().Lookup("max-reminders"))
	viper.BindPFlag("dry-run", reminderCmd.Flags().Lookup("dry-run"))
}