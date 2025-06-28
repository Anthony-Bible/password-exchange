/*
Copyright © 2024 Anthony Bible <anthony@anthony-bible.com>
*/
package reminder

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/rabbitmq"
	sharedConfig "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/shared"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/storage"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/adapters/secondary/validator"
	notificationDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Validation constants for reminder configuration
	MinCheckAfterHours  = 1    // Minimum 1 hour
	MaxCheckAfterHours  = 8760 // Maximum 1 year (365 * 24)
	MinMaxReminders     = 1    // Minimum 1 reminder
	MaxMaxReminders     = 10   // Maximum 10 reminders
	MinReminderInterval = 1    // Minimum 1 hour between reminders
	MaxReminderInterval = 720  // Maximum 30 days (30 * 24)
)


// Simple logger adapter
type loggerAdapter struct {
	logger zerolog.Logger
}

func (l *loggerAdapter) Debug() contracts.LogEvent { return &logEvent{l.logger.Debug()} }
func (l *loggerAdapter) Info() contracts.LogEvent  { return &logEvent{l.logger.Info()} }
func (l *loggerAdapter) Warn() contracts.LogEvent  { return &logEvent{l.logger.Warn()} }
func (l *loggerAdapter) Error() contracts.LogEvent { return &logEvent{l.logger.Error()} }

type logEvent struct {
	event *zerolog.Event
}

func (e *logEvent) Err(err error) contracts.LogEvent              { e.event = e.event.Err(err); return e }
func (e *logEvent) Str(key, value string) contracts.LogEvent     { e.event = e.event.Str(key, value); return e }
func (e *logEvent) Int(key string, value int) contracts.LogEvent { e.event = e.event.Int(key, value); return e }
func (e *logEvent) Bool(key string, value bool) contracts.LogEvent { e.event = e.event.Bool(key, value); return e }
func (e *logEvent) Dur(key string, value time.Duration) contracts.LogEvent { e.event = e.event.Dur(key, value); return e }
func (e *logEvent) Float64(key string, value float64) contracts.LogEvent { e.event = e.event.Float64(key, value); return e }
func (e *logEvent) Msg(msg string) { e.event.Msg(msg) }

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
			// Set up defer before attempting connection
			defer mysqlAdapter.Close()
			
			if err := mysqlAdapter.Connect(); err != nil {
				log.Error().
					Err(err).
					Str("operation", "database_connect").
					Str("host", cfg.DbHost).
					Str("database", cfg.DbName).
					Msg("Failed to connect to database")
				return
			}
		}

		// Initialize storage service
		storageService := storageDomain.NewStorageService(storageAdapter)

		// Create notification storage adapter
		notificationStorageAdapter := storage.NewGRPCStorageAdapter(storageService)

		// Create RabbitMQ notification publisher
		rabbitConfig := rabbitmq.NotificationConfig{
			Host:      cfg.RabHost,
			Port:      cfg.RabPort,
			User:      cfg.RabUser,
			Password:  cfg.RabPass,
			QueueName: cfg.RabQName,
		}
		
		notificationPublisher, err := rabbitmq.NewNotificationPublisher(rabbitConfig)
		if err != nil {
			log.Error().
				Err(err).
				Str("operation", "rabbitmq_connect").
				Str("host", cfg.RabHost).
				Int("port", cfg.RabPort).
				Msg("Failed to connect to RabbitMQ")
			return
		}
		defer notificationPublisher.Close()
       
		// Create port adapters
		configPort := sharedConfig.NewSharedConfigAdapter(cfg.PassConfig)
		loggerPort := &loggerAdapter{logger: log.Logger}
		validationPort := validator.NewValidationAdapter()

		// Create reminder service with storage adapter and notification publisher
		// Uses RabbitMQ to publish reminder notifications instead of sending emails directly
		reminderService := notificationDomain.NewReminderService(notificationStorageAdapter, notificationPublisher, loggerPort, configPort, validationPort)

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
		
		// Signal Istio sidecar to shut down for cronjob completion
		shutdownSidecar("http://localhost:15020/quitquitquit", "http://localhost:4191/shutdown")
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

// shutdownSidecar sends a shutdown signal to the service mesh sidecar.
// This is necessary for cronjobs to complete properly.
func shutdownSidecar(istioURL, linkerdURL string) {
	// Try Istio first
	if shutdown(istioURL, "Istio") {
		return
	}

	// If Istio fails, try Linkerd
	shutdown(linkerdURL, "Linkerd")
}

func shutdown(fullURL, sidecarName string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to create shutdown request for %s sidecar", sidecarName)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Debug().Err(err).Msgf("Failed to shutdown %s sidecar (may not be present)", sidecarName)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Info().Msgf("Successfully signaled %s sidecar shutdown", sidecarName)
		// Give the sidecar a moment to shut down gracefully
		time.Sleep(2 * time.Second)
		return true
	}

	log.Debug().Int("status", resp.StatusCode).Msgf("Unexpected response from %s sidecar shutdown endpoint", sidecarName)
	return false
}
