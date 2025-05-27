/*
Copyright © 2024 Anthony Bible <anthony@anthony-bible.com>
*/
package reminder

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/storage"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Validation constants for reminder configuration
	MinCheckAfterHours   = 1     // Minimum 1 hour
	MaxCheckAfterHours   = 8760  // Maximum 1 year (365 * 24)
	MinMaxReminders      = 1     // Minimum 1 reminder
	MaxMaxReminders      = 10    // Maximum 10 reminders
	MinReminderInterval  = 1     // Minimum 1 hour between reminders
	MaxReminderInterval  = 720   // Maximum 30 days (30 * 24)
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
PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS: Hours to wait before first reminder (1-8760, default: 24)
PASSWORDEXCHANGE_REMINDER_MAXREMINDERS: Maximum reminders per message (1-10, default: 3)
PASSWORDEXCHANGE_REMINDER_INTERVAL: Hours between reminders (1-720, default: 24)`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfig()
		if err != nil {
			log.Error().Err(err).Str("operation", "load_config").Msg("Failed to load configuration")
			return
		}

		// Convert config to notification domain config
		reminderConfig := notificationDomain.ReminderConfig{
			Enabled:         cfg.Reminder.Enabled,
			CheckAfterHours: cfg.Reminder.CheckAfterHours,
			MaxReminders:    cfg.Reminder.MaxReminders,
			Interval:        cfg.Reminder.ReminderInterval,
		}

		// Initialize storage adapter
		dbConfig := storageDomain.DatabaseConfig{
			Host:     cfg.DbHost,
			User:     cfg.DbUser,
			Password: cfg.DbPass,
			Name:     cfg.DbName,
		}
		storageAdapter := mysql.NewMySQLAdapter(dbConfig)
		if mysqlAdapter, ok := storageAdapter.(*mysql.MySQLAdapter); ok {
			if err := mysqlAdapter.Connect(); err != nil {
				log.Error().
					Err(err).
					Str("operation", "database_connect").
					Str("host", cfg.DbHost).
					Str("database", cfg.DbName).
					Msg("Failed to connect to database")
				return
			}
			defer mysqlAdapter.Close()
		}

		// Initialize storage service
		storageService := storageDomain.NewStorageService(storageAdapter)

		// Create notification storage adapter
		notificationStorageAdapter := storage.NewGRPCStorageAdapter(storageService)

		// Create reminder service (for now using nil for email sender - will be implemented later)
		reminderService := notificationDomain.NewReminderService(notificationStorageAdapter, nil)
		
		// Process reminders
		ctx := context.Background()
		if err := reminderService.ProcessReminders(ctx, reminderConfig); err != nil {
			log.Error().
				Err(err).
				Str("operation", "process_reminders").
				Msg("Failed to process reminders")
			return
		}

		log.Info().
			Str("operation", "processing_completed").
			Msg("Reminder email processing completed")
	},
}

// loadConfig loads configuration from viper with defaults
func loadConfig() (*config.Config, error) {
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
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Override with command-line flags if provided
	if olderThanHours := viper.GetString("older-than-hours"); olderThanHours != "" {
		hours, err := strconv.Atoi(olderThanHours)
		if err != nil {
			return nil, fmt.Errorf("invalid value for older-than-hours flag '%s': %w", olderThanHours, err)
		}
		if hours < MinCheckAfterHours || hours > MaxCheckAfterHours {
			return nil, fmt.Errorf("older-than-hours value %d must be between %d and %d", hours, MinCheckAfterHours, MaxCheckAfterHours)
		}
		cfg.Reminder.CheckAfterHours = hours
	}
	if maxReminders := viper.GetString("max-reminders"); maxReminders != "" {
		max, err := strconv.Atoi(maxReminders)
		if err != nil {
			return nil, fmt.Errorf("invalid value for max-reminders flag '%s': %w", maxReminders, err)
		}
		if max < MinMaxReminders || max > MaxMaxReminders {
			return nil, fmt.Errorf("max-reminders value %d must be between %d and %d", max, MinMaxReminders, MaxMaxReminders)
		}
		cfg.Reminder.MaxReminders = max
	}

	log.Info().
		Bool("enabled", cfg.Reminder.Enabled).
		Int("checkAfterHours", cfg.Reminder.CheckAfterHours).
		Int("maxReminders", cfg.Reminder.MaxReminders).
		Str("operation", "config_loaded").
		Msg("Reminder configuration loaded")

	return cfg, nil
}

func init() {
	cmd.RootCmd.AddCommand(reminderCmd)

	// Command-line flags
	reminderCmd.Flags().String("older-than-hours", "", "Hours to wait before sending first reminder (1-8760, default: 24)")
	reminderCmd.Flags().String("max-reminders", "", "Maximum number of reminders per message (1-10, default: 3)")
	reminderCmd.Flags().Bool("dry-run", false, "Show what would be done without actually sending emails")

	// Bind flags to viper
	viper.BindPFlag("older-than-hours", reminderCmd.Flags().Lookup("older-than-hours"))
	viper.BindPFlag("max-reminders", reminderCmd.Flags().Lookup("max-reminders"))
	viper.BindPFlag("dry-run", reminderCmd.Flags().Lookup("dry-run"))
}