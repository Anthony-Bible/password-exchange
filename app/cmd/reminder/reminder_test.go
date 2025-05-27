package reminder

import (
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
)

func TestValidateReminderConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  config.ReminderConfig
		wantErr bool
		errType error
	}{
		{
			name: "valid configuration",
			config: config.ReminderConfig{
				CheckAfterHours:  24,
				MaxReminders:     3,
				ReminderInterval: 24,
			},
			wantErr: false,
		},
		{
			name: "CheckAfterHours too small",
			config: config.ReminderConfig{
				CheckAfterHours:  0,
				MaxReminders:     3,
				ReminderInterval: 24,
			},
			wantErr: true,
			errType: ErrInvalidCheckAfterHours,
		},
		{
			name: "CheckAfterHours too large",
			config: config.ReminderConfig{
				CheckAfterHours:  10000,
				MaxReminders:     3,
				ReminderInterval: 24,
			},
			wantErr: true,
			errType: ErrInvalidCheckAfterHours,
		},
		{
			name: "MaxReminders too small",
			config: config.ReminderConfig{
				CheckAfterHours:  24,
				MaxReminders:     0,
				ReminderInterval: 24,
			},
			wantErr: true,
			errType: ErrInvalidMaxReminders,
		},
		{
			name: "MaxReminders too large",
			config: config.ReminderConfig{
				CheckAfterHours:  24,
				MaxReminders:     15,
				ReminderInterval: 24,
			},
			wantErr: true,
			errType: ErrInvalidMaxReminders,
		},
		{
			name: "ReminderInterval too small",
			config: config.ReminderConfig{
				CheckAfterHours:  24,
				MaxReminders:     3,
				ReminderInterval: 0,
			},
			wantErr: true,
			errType: ErrInvalidReminderInterval,
		},
		{
			name: "ReminderInterval too large",
			config: config.ReminderConfig{
				CheckAfterHours:  24,
				MaxReminders:     3,
				ReminderInterval: 1000,
			},
			wantErr: true,
			errType: ErrInvalidReminderInterval,
		},
		{
			name: "boundary values - minimum valid",
			config: config.ReminderConfig{
				CheckAfterHours:  MinCheckAfterHours,
				MaxReminders:     MinMaxReminders,
				ReminderInterval: MinReminderInterval,
			},
			wantErr: false,
		},
		{
			name: "boundary values - maximum valid",
			config: config.ReminderConfig{
				CheckAfterHours:  MaxCheckAfterHours,
				MaxReminders:     MaxMaxReminders,
				ReminderInterval: MaxReminderInterval,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReminderConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateReminderConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil {
				if !isExpectedError(err, tt.errType) {
					t.Errorf("validateReminderConfig() error = %v, expected error type %v", err, tt.errType)
				}
			}
		})
	}
}

// isExpectedError checks if the error contains the expected error type
func isExpectedError(got, expected error) bool {
	if got == nil || expected == nil {
		return got == expected
	}
	return got.Error() != "" && expected.Error() != ""
}