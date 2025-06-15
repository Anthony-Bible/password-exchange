package features

import (
	"os"
	"strconv"
	"sync"
)

// FeatureFlags holds the state of all feature flags in the application.
type FeatureFlags struct {
	UseSlog bool // If true, slog is used; otherwise, zerolog is used.
}

var (
	currentFlags *FeatureFlags
	once         sync.Once
)

const (
	// EnvUseSlog is the environment variable to control the slog feature flag.
	// Set to "true" to enable slog, "false" or unset to use zerolog.
	EnvUseSlog = "FEATURE_USE_SLOG"
)

// GetFlags returns the current feature flag settings.
// It initializes the flags based on environment variables on its first call.
func GetFlags() *FeatureFlags {
	once.Do(func() {
		useSlogStr := os.Getenv(EnvUseSlog)
		useSlog, _ := strconv.ParseBool(useSlogStr) // Defaults to false if parsing fails or env var is not set

		currentFlags = &FeatureFlags{
			UseSlog: useSlog,
		}
	})
	return currentFlags
}

// IsSlogEnabled checks if the slog logger is enabled via feature flag.
func IsSlogEnabled() bool {
	return GetFlags().UseSlog
}
