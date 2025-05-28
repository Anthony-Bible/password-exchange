/*
Copyright © 2024 Anthony Bible <anthony@anthony-bible.com>
*/
package reminder

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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

// Config represents the reminder command configuration
type Config struct {
	config.Config `mapstructure:",squash"`
}

// bindenvs is required due to viper not automatically mapping env to marshal https://github.com/spf13/viper/issues/584
func bindenvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		v := ifv.Field(i)
		t := ifv.Type().Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		if tv == ",squash" {
			bindenvs(v.Interface(), parts...)
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindenvs(v.Interface(), append(parts, tv)...)
		default:
			key := strings.Join(append(parts, tv), ".")
			// Build environment variable name from key
			envKey := "PASSWORDEXCHANGE_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
			viper.BindEnv(key, envKey)
		}
	}
}

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
		var cfg Config
		bindenvs(&cfg)
		viper.Unmarshal(&cfg)

		// Apply default values for any unset reminder configuration
		cfg.Reminder.WithDefaults()

		// Apply CLI flag overrides with validation
		if err := applyFlagOverrides(&cfg); err != nil {
			log.Error().Err(err).Str("operation", "flag_validation").Msg("Failed to validate configuration flags")
			return
		}

		// Convert shared config to notification domain-specific config
		// This maintains separation between CLI configuration and domain logic
		reminderConfig := notificationDomain.ReminderConfig{
			Enabled:         cfg.Reminder.Enabled,
			CheckAfterHours: cfg.Reminder.CheckAfterHours,
			MaxReminders:    cfg.Reminder.MaxReminders,
			Interval:        cfg.Reminder.ReminderInterval,
		}

		// Initialize storage adapter with database connection
		// Uses MySQL adapter as the concrete implementation of the storage port
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

		// Create reminder service with storage adapter
		// Email sender is nil for now as reminders only log without actually sending emails
		// This follows dependency injection pattern for hexagonal architecture
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

// applyFlagOverrides applies command-line flag overrides with validation
func applyFlagOverrides(cfg *Config) error {
	// Override with command-line flags if provided
	if olderThanHoursFlag := viper.GetString("older-than-hours"); olderThanHoursFlag != "" {
		hours, err := strconv.Atoi(olderThanHoursFlag)
		if err != nil {
			return fmt.Errorf("invalid value for older-than-hours flag '%s': %w", olderThanHoursFlag, err)
		}
		if hours < MinCheckAfterHours || hours > MaxCheckAfterHours {
			return fmt.Errorf("older-than-hours value %d must be between %d and %d", hours, MinCheckAfterHours, MaxCheckAfterHours)
		}
		cfg.Reminder.CheckAfterHours = hours
	}
	if maxRemindersFlag := viper.GetString("max-reminders"); maxRemindersFlag != "" {
		maxRemindersValue, err := strconv.Atoi(maxRemindersFlag)
		if err != nil {
			return fmt.Errorf("invalid value for max-reminders flag '%s': %w", maxRemindersFlag, err)
		}
		if maxRemindersValue < MinMaxReminders || maxRemindersValue > MaxMaxReminders {
			return fmt.Errorf("max-reminders value %d must be between %d and %d", maxRemindersValue, MinMaxReminders, MaxMaxReminders)
		}
		cfg.Reminder.MaxReminders = maxRemindersValue
	}
	if intervalHoursFlag := viper.GetString("interval-hours"); intervalHoursFlag != "" {
		intervalValue, err := strconv.Atoi(intervalHoursFlag)
		if err != nil {
			return fmt.Errorf("invalid value for interval-hours flag '%s': %w", intervalHoursFlag, err)
		}
		if intervalValue < MinReminderInterval || intervalValue > MaxReminderInterval {
			return fmt.Errorf("interval-hours value %d must be between %d and %d", intervalValue, MinReminderInterval, MaxReminderInterval)
		}
		cfg.Reminder.ReminderInterval = intervalValue
	}

	log.Info().
		Bool("enabled", cfg.Reminder.Enabled).
		Int("checkAfterHours", cfg.Reminder.CheckAfterHours).
		Int("maxReminders", cfg.Reminder.MaxReminders).
		Str("operation", "config_loaded").
		Msg("Reminder configuration loaded")

	return nil
}

func init() {
	cmd.RootCmd.AddCommand(reminderCmd)

	// Set defaults for reminder configuration
	viper.SetDefault("reminder.enabled", true)
	viper.SetDefault("reminder.checkafterhours", 24)
	viper.SetDefault("reminder.maxreminders", 3)
	viper.SetDefault("reminder.reminderinterval", 24)

	// Command-line flags
	reminderCmd.Flags().String("older-than-hours", "", "Hours to wait before sending first reminder (1-8760, default: 24)")
	reminderCmd.Flags().String("max-reminders", "", "Maximum number of reminders per message (1-10, default: 3)")
	reminderCmd.Flags().String("interval-hours", "", "Hours between reminders (1-720, default: 24)")

	// Bind flags to viper
	viper.BindPFlag("older-than-hours", reminderCmd.Flags().Lookup("older-than-hours"))
	viper.BindPFlag("max-reminders", reminderCmd.Flags().Lookup("max-reminders"))
	viper.BindPFlag("interval-hours", reminderCmd.Flags().Lookup("interval-hours"))
}