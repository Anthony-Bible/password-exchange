package reminder

import (
	"os"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestViper initializes viper for tests exactly like the application does
func setupTestViper() {
	viper.Reset()
	viper.SetEnvPrefix("passwordexchange")
	viper.AutomaticEnv()
}

// setupTestViperWithDefaults sets up viper and includes the default values
func setupTestViperWithDefaults() {
	setupTestViper()
	// Set same defaults as in reminder.go init()
	viper.SetDefault("reminder.enabled", true)
	viper.SetDefault("reminder.checkafterhours", 24)
	viper.SetDefault("reminder.maxreminders", 3)
	viper.SetDefault("reminder.reminderinterval", 24)
}

func TestConfigurationLoading(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		teardown func(t *testing.T)
		verify   func(t *testing.T, cfg Config)
	}{
		{
			name: "loads default configuration",
			setup: func(t *testing.T) {
				setupTestViperWithDefaults()
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			verify: func(t *testing.T, cfg Config) {
				assert.True(t, cfg.Reminder.Enabled)
				assert.Equal(t, 24, cfg.Reminder.CheckAfterHours)
				assert.Equal(t, 3, cfg.Reminder.MaxReminders)
				assert.Equal(t, 24, cfg.Reminder.ReminderInterval)
			},
		},
		{
			name: "loads configuration from environment variables",
			setup: func(t *testing.T) {
				// Set environment variables BEFORE setting up viper
				os.Setenv("PASSWORDEXCHANGE_REMINDER_ENABLED", "false")
				os.Setenv("PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS", "48")
				os.Setenv("PASSWORDEXCHANGE_REMINDER_MAXREMINDERS", "5")
				os.Setenv("PASSWORDEXCHANGE_REMINDER_REMINDERINTERVAL", "72")
				setupTestViper()
			},
			teardown: func(t *testing.T) {
				os.Unsetenv("PASSWORDEXCHANGE_REMINDER_ENABLED")
				os.Unsetenv("PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS")
				os.Unsetenv("PASSWORDEXCHANGE_REMINDER_MAXREMINDERS")
				os.Unsetenv("PASSWORDEXCHANGE_REMINDER_REMINDERINTERVAL")
				viper.Reset()
			},
			verify: func(t *testing.T, cfg Config) {
				assert.False(t, cfg.Reminder.Enabled)
				assert.Equal(t, 48, cfg.Reminder.CheckAfterHours)
				assert.Equal(t, 5, cfg.Reminder.MaxReminders)
				assert.Equal(t, 72, cfg.Reminder.ReminderInterval)
			},
		},
		{
			name: "loads database configuration from environment variables",
			setup: func(t *testing.T) {
				setupTestViper()
				os.Setenv("PASSWORDEXCHANGE_DBHOST", "testhost")
				os.Setenv("PASSWORDEXCHANGE_DBUSER", "testuser")
				os.Setenv("PASSWORDEXCHANGE_DBPASS", "testpass")
				os.Setenv("PASSWORDEXCHANGE_DBNAME", "testdb")
				os.Setenv("PASSWORDEXCHANGE_DBPORT", "3307")
			},
			teardown: func(t *testing.T) {
				os.Unsetenv("PASSWORDEXCHANGE_DBHOST")
				os.Unsetenv("PASSWORDEXCHANGE_DBUSER")
				os.Unsetenv("PASSWORDEXCHANGE_DBPASS")
				os.Unsetenv("PASSWORDEXCHANGE_DBNAME")
				os.Unsetenv("PASSWORDEXCHANGE_DBPORT")
				viper.Reset()
			},
			verify: func(t *testing.T, cfg Config) {
				assert.Equal(t, "testhost", cfg.DbHost)
				assert.Equal(t, "testuser", cfg.DbUser)
				assert.Equal(t, "testpass", cfg.DbPass)
				assert.Equal(t, "testdb", cfg.DbName)
				assert.Equal(t, 3307, cfg.DbPort)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			defer tt.teardown(t)

			var cfg Config
			bindenvs(&cfg)
			err := viper.Unmarshal(&cfg)
			require.NoError(t, err)

			tt.verify(t, cfg)
		})
	}
}

func TestBindEnvs(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedKey string
		expectedVal interface{}
	}{
		{
			name: "binds simple reminder field",
			envVars: map[string]string{
				"PASSWORDEXCHANGE_REMINDER_ENABLED": "false",
			},
			expectedKey: "reminder.enabled",
			expectedVal: "false",
		},
		{
			name: "binds nested database field",
			envVars: map[string]string{
				"PASSWORDEXCHANGE_DBHOST": "localhost",
			},
			expectedKey: "dbhost",
			expectedVal: "localhost",
		},
		{
			name: "binds multiple reminder fields",
			envVars: map[string]string{
				"PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS": "48",
				"PASSWORDEXCHANGE_REMINDER_MAXREMINDERS":    "5",
			},
			expectedKey: "reminder.checkafterhours",
			expectedVal: "48",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			setupTestViper()
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			// Test
			var cfg Config
			bindenvs(&cfg)

			// Verify environment binding worked
			assert.Equal(t, tt.expectedVal, viper.GetString(tt.expectedKey))

			// Cleanup
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
			viper.Reset()
		})
	}
}

func TestApplyFlagOverrides(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) *Config
		teardown  func(t *testing.T)
		expectErr bool
		errorMsg  string
		verify    func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid older-than-hours flag override",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("older-than-hours", "48")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
							MaxReminders:    3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: false,
			verify: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 48, cfg.Reminder.CheckAfterHours)
			},
		},
		{
			name: "valid max-reminders flag override",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("max-reminders", "5")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
							MaxReminders:    3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: false,
			verify: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 5, cfg.Reminder.MaxReminders)
			},
		},
		{
			name: "invalid older-than-hours flag - non-numeric",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("older-than-hours", "invalid")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: true,
			errorMsg:  "invalid value for older-than-hours flag 'invalid'",
		},
		{
			name: "invalid older-than-hours flag - below minimum",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("older-than-hours", "0")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: true,
			errorMsg:  "older-than-hours value 0 must be between 1 and 8760",
		},
		{
			name: "invalid older-than-hours flag - above maximum",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("older-than-hours", "8761")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: true,
			errorMsg:  "older-than-hours value 8761 must be between 1 and 8760",
		},
		{
			name: "invalid max-reminders flag - non-numeric",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("max-reminders", "invalid")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							MaxReminders: 3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: true,
			errorMsg:  "invalid value for max-reminders flag 'invalid'",
		},
		{
			name: "invalid max-reminders flag - below minimum",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("max-reminders", "0")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							MaxReminders: 3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: true,
			errorMsg:  "max-reminders value 0 must be between 1 and 10",
		},
		{
			name: "invalid max-reminders flag - above maximum",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("max-reminders", "11")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							MaxReminders: 3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: true,
			errorMsg:  "max-reminders value 11 must be between 1 and 10",
		},
		{
			name: "no flag overrides - config unchanged",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
							MaxReminders:    3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: false,
			verify: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 24, cfg.Reminder.CheckAfterHours)
				assert.Equal(t, 3, cfg.Reminder.MaxReminders)
			},
		},
		{
			name: "valid boundary values - minimum",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("older-than-hours", "1")
				viper.Set("max-reminders", "1")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
							MaxReminders:    3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: false,
			verify: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1, cfg.Reminder.CheckAfterHours)
				assert.Equal(t, 1, cfg.Reminder.MaxReminders)
			},
		},
		{
			name: "valid boundary values - maximum",
			setup: func(t *testing.T) *Config {
				viper.Reset()
				viper.Set("older-than-hours", "8760")
				viper.Set("max-reminders", "10")
				return &Config{
					Config: config.Config{
						Reminder: config.ReminderConfig{
							CheckAfterHours: 24,
							MaxReminders:    3,
						},
					},
				}
			},
			teardown: func(t *testing.T) {
				viper.Reset()
			},
			expectErr: false,
			verify: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 8760, cfg.Reminder.CheckAfterHours)
				assert.Equal(t, 10, cfg.Reminder.MaxReminders)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setup(t)
			defer tt.teardown(t)

			err := applyFlagOverrides(cfg)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				if tt.verify != nil {
					tt.verify(t, cfg)
				}
			}
		})
	}
}

func TestValidationConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"MinCheckAfterHours", MinCheckAfterHours, 1},
		{"MaxCheckAfterHours", MaxCheckAfterHours, 8760},
		{"MinMaxReminders", MinMaxReminders, 1},
		{"MaxMaxReminders", MaxMaxReminders, 10},
		{"MinReminderInterval", MinReminderInterval, 1},
		{"MaxReminderInterval", MaxReminderInterval, 720},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestConfigStructure(t *testing.T) {
	t.Run("config embeds config.Config correctly", func(t *testing.T) {
		cfg := Config{}
		
		// Test that we can access embedded fields
		cfg.DbHost = "testhost"
		cfg.DbUser = "testuser"
		cfg.Reminder.Enabled = true
		cfg.Reminder.CheckAfterHours = 48
		
		assert.Equal(t, "testhost", cfg.DbHost)
		assert.Equal(t, "testuser", cfg.DbUser)
		assert.True(t, cfg.Reminder.Enabled)
		assert.Equal(t, 48, cfg.Reminder.CheckAfterHours)
	})
}

func TestViperFlagIntegration(t *testing.T) {
	t.Run("command flags are properly bound to viper", func(t *testing.T) {
		// Reset viper state
		viper.Reset()
		
		// Create a test command to simulate CLI usage
		testCmd := &cobra.Command{
			Use: "test",
		}
		
		// Add the same flags as reminder command
		testCmd.Flags().String("older-than-hours", "", "Test flag")
		testCmd.Flags().String("max-reminders", "", "Test flag")
		testCmd.Flags().Bool("dry-run", false, "Test flag")
		
		// Bind flags to viper
		viper.BindPFlag("older-than-hours", testCmd.Flags().Lookup("older-than-hours"))
		viper.BindPFlag("max-reminders", testCmd.Flags().Lookup("max-reminders"))
		viper.BindPFlag("dry-run", testCmd.Flags().Lookup("dry-run"))
		
		// Set flag values
		testCmd.Flags().Set("older-than-hours", "48")
		testCmd.Flags().Set("max-reminders", "5")
		testCmd.Flags().Set("dry-run", "true")
		
		// Verify viper can read the flag values
		assert.Equal(t, "48", viper.GetString("older-than-hours"))
		assert.Equal(t, "5", viper.GetString("max-reminders"))
		assert.True(t, viper.GetBool("dry-run"))
	})
}

func TestCompleteConfigurationFlow(t *testing.T) {
	t.Run("complete configuration loading flow", func(t *testing.T) {
		// Setup environment
		setupTestViper()
		os.Setenv("PASSWORDEXCHANGE_REMINDER_ENABLED", "true")
		os.Setenv("PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS", "24")
		os.Setenv("PASSWORDEXCHANGE_DBHOST", "testhost")
		defer func() {
			os.Unsetenv("PASSWORDEXCHANGE_REMINDER_ENABLED")
			os.Unsetenv("PASSWORDEXCHANGE_REMINDER_CHECKAFTERHOURS")
			os.Unsetenv("PASSWORDEXCHANGE_DBHOST")
			viper.Reset()
		}()
		
		// Set viper flag override
		viper.Set("older-than-hours", "48")
		
		// Load configuration
		var cfg Config
		bindenvs(&cfg)
		err := viper.Unmarshal(&cfg)
		require.NoError(t, err)
		
		// Apply flag overrides
		err = applyFlagOverrides(&cfg)
		require.NoError(t, err)
		
		// Verify final configuration
		assert.True(t, cfg.Reminder.Enabled)
		assert.Equal(t, 48, cfg.Reminder.CheckAfterHours) // Overridden by flag
		assert.Equal(t, "testhost", cfg.DbHost)
	})
}