package logging

// LogConfig holds the configuration for the logging setup.
type LogConfig struct {
	Level       string            `mapstructure:"level"`       // Default log level (e.g., "info", "debug", "error")
	Format      string            `mapstructure:"format"`      // Log format (e.g., "json", "text")
	Services    map[string]string `mapstructure:"services"`    // Per-service log levels
	Sampling    SamplingConfig    `mapstructure:"sampling"`    // Log sampling configuration
	Output      OutputConfig      `mapstructure:"output"`      // Log output configuration
	Performance PerformanceConfig `mapstructure:"performance"` // Performance-related logging settings
}

// SamplingConfig configures log sampling.
type SamplingConfig struct {
	Enabled   bool    `mapstructure:"enabled"`   // Whether sampling is enabled
	Rate      float64 `mapstructure:"rate"`      // Sampling rate (e.g., 0.1 for 10%)
	Burst     int     `mapstructure:"burst"`     // Number of messages to allow in a burst
	Interval  string  `mapstructure:"interval"`  // Interval for sampling (e.g., "1s", "1m")
}

// OutputConfig defines where logs should be written.
type OutputConfig struct {
	Type      string   `mapstructure:"type"`       // Output type (e.g., "stdout", "file", "multi")
	Paths     []string `mapstructure:"paths"`      // File paths for "file" or "multi" output
	Rotation  LogRotationConfig `mapstructure:"rotation"`   // Log rotation settings for file output
}

// LogRotationConfig configures log rotation for file outputs.
type LogRotationConfig struct {
	Enabled    bool   `mapstructure:"enabled"`    // Whether log rotation is enabled
	MaxSize    int    `mapstructure:"maxsize"`    // Maximum size in megabytes before rotation
	MaxBackups int    `mapstructure:"maxbackups"` // Maximum number of old log files to retain
	MaxAge     int    `mapstructure:"maxage"`     // Maximum number of days to retain old log files
	Compress   bool   `mapstructure:"compress"`   // Whether to compress rotated log files
}

// PerformanceConfig holds settings related to logging performance.
type PerformanceConfig struct {
	Async           bool `mapstructure:"async"`            // Enable asynchronous logging
	BufferSize      int  `mapstructure:"buffersize"`       // Buffer size for async logging
	OverflowHandling string `mapstructure:"overflowhandling"` // How to handle buffer overflow (e.g., "drop", "block")
}
