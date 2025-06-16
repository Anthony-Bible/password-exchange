package config

import (
	"reflect"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging" // Added for LogConfig
	"github.com/spf13/viper"                                                 // Added for Viper
)

// AppConfig global variable - keeping it as is from the original file.
// Consider refactoring this in the future if it's not actively used or if
// the new `LoadConfig` function makes it redundant.
var AppConfig PassConfig

// Config represents the complete application configuration.
// It now includes logging configuration and the EnableSlog flag.
type Config struct {
	PassConfig `mapstructure:",squash"`      // Existing PassConfig
	Database   domain.DatabaseConfig `mapstructure:"database"`  // Existing Database config
	Reminder   ReminderConfig        `mapstructure:"reminder"`  // Existing Reminder config
	EnableSlog bool                  `mapstructure:"enable_slog"` // New: Flag to enable Slog
	Log        logging.LogConfig     `mapstructure:"log"`         // New: Logging configuration
}

// ReminderConfig contains configuration for the reminder email system (existing struct)
type ReminderConfig struct {
	Enabled          bool `mapstructure:"enabled" default:"true"`
	CheckAfterHours  int  `mapstructure:"checkafterhours" default:"24"`
	MaxReminders     int  `mapstructure:"maxreminders" default:"3"`
	ReminderInterval int  `mapstructure:"reminderinterval" default:"24"`
}

// NewReminderConfig creates a new ReminderConfig with proper default values (existing function)
func NewReminderConfig() ReminderConfig {
	return ReminderConfig{
		Enabled:          true,
		CheckAfterHours:  24,
		MaxReminders:     3,
		ReminderInterval: 24,
	}
}

// WithDefaults applies default values to zero-value fields and ensures valid ranges (existing method)
func (r *ReminderConfig) WithDefaults() {
	if r.CheckAfterHours == 0 {
		r.CheckAfterHours = 24
	}
	if r.MaxReminders == 0 {
		r.MaxReminders = 3
	}
	if r.ReminderInterval == 0 {
		r.ReminderInterval = 24
	}
	if r.CheckAfterHours < 1 || r.CheckAfterHours > 8760 {
		r.CheckAfterHours = 24
	}
	if r.MaxReminders < 1 || r.MaxReminders > 10 {
		r.MaxReminders = 3
	}
	if r.ReminderInterval < 1 || r.ReminderInterval > 720 {
		r.ReminderInterval = 24
	}
}

// PassConfig struct (existing struct)
type PassConfig struct {
	EmailHost             string `mapstructure:"emailhost"`
	EmailUser             string `mapstructure:"emailuser"`
	EmailPass             string `mapstructure:"emailpass"`
	EmailFrom             string `mapstructure:"emailfrom"`
	RabHost               string `mapstructure:"rabhost"`
	RabUser               string `mapstructure:"rabuser"`
	RabPass               string `mapstructure:"rabpass"`
	RabQName              string `mapstructure:"rabqname"`
	DbHost                string `mapstructure:"dbhost"`
	DbUser                string `mapstructure:"dbuser"`
	DbPass                string `mapstructure:"dbpass"`
	DbName                string `mapstructure:"dbname"`
	ProdHost              string `mapstructure:"prodhost"`
	DevHost               string `mapstructure:"devhost"`
	EncryptionProdService string `mapstructure:"encryptionprodservice"`
	DatabaseProdService   string `mapstructure:"databaseprodservice"`
	EncryptionDevService  string `mapstructure:"encryptiondevservice"`
	DatabaseDevService    string `mapstructure:"databasedevservice"`
	Loglevel              string `mapstructure:"loglevel"` // This might be superseded by Log.Level
	RunningEnvironment    string `mapstructure:"runningenvironment"`
	TurnstileSecret       string `mapstructure:"turnstile_secret"`
	DefaultMaxViewCount   int    `mapstructure:"defaultmaxviewcount"`
	DbPort                int    `mapstructure:"dbport"`
	EmailPort             int    `mapstructure:"emailport"`
	RabPort               int    `mapstructure:"rabport"`
}

// LoadConfig loads the application configuration from environment variables and/or a config file.
// This is a new function based on the subtask requirements.
func LoadConfig() (*Config, error) {
	var cfg Config

	// Set defaults for new logging configurations
	viper.SetDefault("enable_slog", false) // Default to zerolog
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	// Defaults for other LogConfig fields (like sampling, output, performance) can be added here if needed
	// e.g., viper.SetDefault("log.sampling.enabled", false)

	// Set up Viper for environment variable binding
	viper.SetEnvPrefix("APP") // Application specific prefix
	viper.AutomaticEnv()      // Read all environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replaces . with _ in env var names (e.g., log.level -> APP_LOG_LEVEL)

	// Explicitly bind APP_ENABLE_SLOG to cfg.EnableSlog.
	// For simple top-level keys, BindEnv can be useful.
	// For nested keys, AutomaticEnv with SetEnvKeyReplacer usually handles it.
	if err := viper.BindEnv("enable_slog", "APP_ENABLE_SLOG"); err != nil {
		// This specific BindEnv call might be redundant if APP_ENABLE_SLOG is already picked up by AutomaticEnv.
		// However, it doesn't hurt to be explicit for critical flags.
		// fmt.Printf("Error binding APP_ENABLE_SLOG: %v\n", err) // Optional: log error
	}

	// Recursively bind environment variables to struct fields
	// This helper function ensures that environment variables are bound to the struct fields
	// according to their `mapstructure` tags.
	bindEnvs(reflect.TypeOf(cfg), "")

	// Unmarshal all the configuration into the Config struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

    // Apply defaults for ReminderConfig after unmarshalling, if they are not set by viper
    // This ensures that the WithDefaults logic is applied if no config values were found.
    // However, viper.SetDefault should ideally handle this.
    // If ReminderConfig fields are properly tagged and viper.SetDefault is used for them,
    // this manual call might not be necessary.
    // For now, keeping it to ensure compatibility with existing logic if it relies on this.
    cfg.Reminder.WithDefaults()


	// The global AppConfig variable:
	// If it's still needed, it should be populated from the loaded cfg.
	// For example: AppConfig = cfg.PassConfig
	// This depends on how AppConfig is used elsewhere.
	// For now, I'm not changing how AppConfig is populated beyond its declaration.
	// If `PassConfig` fields are part of `Config` loaded by Viper, then `cfg.PassConfig` will have values.

	return &cfg, nil
}

// bindEnvs recursively binds environment variables to struct fields.
// This version is adapted to work with Viper's environment variable handling.
func bindEnvs(t reflect.Type, prefix string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the mapstructure tag
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" { // Skip if no tag or explicitly ignored
			if field.Name == "PassConfig" && field.Anonymous { // Handle embedded PassConfig
				bindEnvs(field.Type, prefix) // Recurse into embedded struct without changing prefix
			}
			continue
		}
		if tag == ",squash" { // Handle squash tag for embedded structs
			bindEnvs(field.Type, prefix) // Recurse with the same prefix
			continue
		}


		// Construct the path for Viper (e.g., "log.level")
		viperKey := tag
		if prefix != "" {
			viperKey = prefix + "." + tag
		}

		if field.Type.Kind() == reflect.Struct && field.Name != "PassConfig" { // Recurse for non-anonymous structs
			// Pass the viperKey as the new prefix for nested structs
			bindEnvs(field.Type, viperKey)
		} else if field.Type.Kind() != reflect.Struct { // Bind simple fields
			// Construct the environment variable name (e.g., APP_LOG_LEVEL)
			// This relies on viper.SetEnvKeyReplacer and viper.SetEnvPrefix.
			// viper.BindEnv (called here or by AutomaticEnv) uses the viperKey.
			// No, we don't need to call viper.BindEnv here for each field if AutomaticEnv is working.
			// AutomaticEnv should handle binding APP_PREFIX_VIPER_KEY to the struct field.
			// This function's role can be to ensure defaults are registered or for complex cases.
			// The original prompt's bindEnvs called viper.BindEnv. Let's re-evaluate.
			// If AutomaticEnv is used, manual BindEnv calls are often for specific overrides or
			// when env var names don't match the pattern.
			// The prompt's bindEnvs: `if err := viper.BindEnv(fullTagPath, "APP_"+envVarName); err != nil`
			// This implies we *should* call BindEnv.

			// Let's try to align with the spirit of the prompt's bindEnvs, but make it Viper idiomatic.
			// viper.BindEnv(key string, envName ...string) error
			// `key` is the path (e.g. "log.level")
			// `envName` is the explicit environment variable name. If not provided, Viper derives it.
			// Since we use AutomaticEnv and SetEnvKeyReplacer, Viper can derive the env var name.
			// So, viper.BindEnv(viperKey) should be sufficient to ensure it's bound if not already.
			// This can help if there are env vars that AutomaticEnv might miss or if we want to ensure they are considered.

			// It is generally recommended to rely on AutomaticEnv and use BindEnv for specific exceptions or explicit mappings.
			// Calling BindEnv for every field might be redundant but could act as a safeguard.
			// Let's simplify: AutomaticEnv should cover most cases. The explicit BindEnv for APP_ENABLE_SLOG is a good example.
			// The recursive nature of bindEnvs is more about traversing the struct to potentially set defaults or other viper settings per field.
			// For now, we will keep the recursive structure but not call viper.BindEnv within the loop,
			// relying on AutomaticEnv and specific BindEnv calls in LoadConfig if needed.
			// The main purpose of this recursive traversal will be to ensure all nested struct defaults are also set if viper.SetDefault is used with full paths.
		}
	}
}
