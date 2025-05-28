package config

import (
	"testing"
)

func TestNewReminderConfig(t *testing.T) {
	config := NewReminderConfig()

	if !config.Enabled {
		t.Errorf("Expected Enabled to be true, got %v", config.Enabled)
	}
	if config.CheckAfterHours != 24 {
		t.Errorf("Expected CheckAfterHours to be 24, got %d", config.CheckAfterHours)
	}
	if config.MaxReminders != 3 {
		t.Errorf("Expected MaxReminders to be 3, got %d", config.MaxReminders)
	}
	if config.ReminderInterval != 24 {
		t.Errorf("Expected ReminderInterval to be 24, got %d", config.ReminderInterval)
	}
}

func TestReminderConfigWithDefaults_ZeroValues(t *testing.T) {
	config := ReminderConfig{}
	config.WithDefaults()

	if config.CheckAfterHours != 24 {
		t.Errorf("Expected CheckAfterHours to be 24, got %d", config.CheckAfterHours)
	}
	if config.MaxReminders != 3 {
		t.Errorf("Expected MaxReminders to be 3, got %d", config.MaxReminders)
	}
	if config.ReminderInterval != 24 {
		t.Errorf("Expected ReminderInterval to be 24, got %d", config.ReminderInterval)
	}
}

func TestReminderConfigWithDefaults_ValidValues(t *testing.T) {
	config := ReminderConfig{
		Enabled:          false,
		CheckAfterHours:  48,
		MaxReminders:     5,
		ReminderInterval: 12,
	}
	config.WithDefaults()

	// Valid values should not be changed
	if config.Enabled != false {
		t.Errorf("Expected Enabled to remain false, got %v", config.Enabled)
	}
	if config.CheckAfterHours != 48 {
		t.Errorf("Expected CheckAfterHours to remain 48, got %d", config.CheckAfterHours)
	}
	if config.MaxReminders != 5 {
		t.Errorf("Expected MaxReminders to remain 5, got %d", config.MaxReminders)
	}
	if config.ReminderInterval != 12 {
		t.Errorf("Expected ReminderInterval to remain 12, got %d", config.ReminderInterval)
	}
}

func TestReminderConfigWithDefaults_InvalidValues(t *testing.T) {
	config := ReminderConfig{
		CheckAfterHours:  -1,   // Invalid: below minimum
		MaxReminders:     15,   // Invalid: above maximum
		ReminderInterval: 1000, // Invalid: above maximum
	}
	config.WithDefaults()

	// Invalid values should be reset to defaults
	if config.CheckAfterHours != 24 {
		t.Errorf("Expected CheckAfterHours to be reset to 24, got %d", config.CheckAfterHours)
	}
	if config.MaxReminders != 3 {
		t.Errorf("Expected MaxReminders to be reset to 3, got %d", config.MaxReminders)
	}
	if config.ReminderInterval != 24 {
		t.Errorf("Expected ReminderInterval to be reset to 24, got %d", config.ReminderInterval)
	}
}

func TestReminderConfigWithDefaults_BoundaryValues(t *testing.T) {
	tests := []struct {
		name           string
		config         ReminderConfig
		expectedConfig ReminderConfig
	}{
		{
			name: "minimum valid values",
			config: ReminderConfig{
				CheckAfterHours:  1,
				MaxReminders:     1,
				ReminderInterval: 1,
			},
			expectedConfig: ReminderConfig{
				CheckAfterHours:  1,
				MaxReminders:     1,
				ReminderInterval: 1,
			},
		},
		{
			name: "maximum valid values",
			config: ReminderConfig{
				CheckAfterHours:  8760,
				MaxReminders:     10,
				ReminderInterval: 720,
			},
			expectedConfig: ReminderConfig{
				CheckAfterHours:  8760,
				MaxReminders:     10,
				ReminderInterval: 720,
			},
		},
		{
			name: "boundary invalid values",
			config: ReminderConfig{
				CheckAfterHours:  8761, // Just above max
				MaxReminders:     0,    // Below min
				ReminderInterval: 721,  // Just above max
			},
			expectedConfig: ReminderConfig{
				CheckAfterHours:  24, // Reset to default
				MaxReminders:     3,  // Reset to default
				ReminderInterval: 24, // Reset to default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			config.WithDefaults()

			if config.CheckAfterHours != tt.expectedConfig.CheckAfterHours {
				t.Errorf("Expected CheckAfterHours to be %d, got %d", tt.expectedConfig.CheckAfterHours, config.CheckAfterHours)
			}
			if config.MaxReminders != tt.expectedConfig.MaxReminders {
				t.Errorf("Expected MaxReminders to be %d, got %d", tt.expectedConfig.MaxReminders, config.MaxReminders)
			}
			if config.ReminderInterval != tt.expectedConfig.ReminderInterval {
				t.Errorf("Expected ReminderInterval to be %d, got %d", tt.expectedConfig.ReminderInterval, config.ReminderInterval)
			}
		})
	}
}