package migrations

import "embed"

// PostgresFS holds the embedded Postgres migrations (under postgres/).
//
//go:embed postgres/*.sql
var PostgresFS embed.FS
