package logging

// SamplingConfig defines the configuration for log sampling.
type SamplingConfig struct {
	Interval int `mapstructure:"interval"` // Number of logs to skip before logging one
	Hook     func(map[string]any) bool `mapstructure:"-"` // Optional hook to decide if a log should be sampled
}

// OutputConfig defines the configuration for log output.
type OutputConfig struct {
	Path       string `mapstructure:"path"`       // File path to write logs to, if any
	MaxSize    int    `mapstructure:"max_size"`    // Maximum size of log files in megabytes
	MaxBackups int    `mapstructure:"max_backups"` // Maximum number of old log files to retain
	MaxAge     int    `mapstructure:"max_age"`     // Maximum number of days to retain old log files
	Compress   bool   `mapstructure:"compress"`   // Whether to compress retired log files
}

// PerformanceConfig defines performance-related logging configurations.
type PerformanceConfig struct {
	Async         bool `mapstructure:"async"`          // Enable asynchronous logging
	BufferSize    int  `mapstructure:"buffer_size"`    // Buffer size for async logging
	FlushInterval int  `mapstructure:"flush_interval"` // How often to flush the buffer in milliseconds
}

// LogConfig defines the overall logging configuration structure.
// It supports per-service log levels, sampling, multiple outputs,
// and performance tuning options.
type LogConfig struct {
	Level       string            `mapstructure:"level"`                 // Default log level (e.g., "info", "debug", "error")
	Format      string            `mapstructure:"format"`                // Log format ("json", "text")
	Services    map[string]string `mapstructure:"services"`              // Per-service log levels, e.g., {"service-name": "debug"}
	Sampling    SamplingConfig    `mapstructure:"sampling,omitempty"`    // Log sampling configuration
	Output      OutputConfig      `mapstructure:"output,omitempty"`      // Log output configuration (for file logging)
	Performance PerformanceConfig `mapstructure:"performance,omitempty"` // Performance-related configurations
}
