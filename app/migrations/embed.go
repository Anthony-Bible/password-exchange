// Package migrations embeds the SQL migrations shipped inside the binary.
//
// The MySQL migrations at the root of this directory are still read from disk
// until they are removed; golang-migrate's file source skips subdirectories,
// so postgres/ does not affect it.
package migrations

import "embed"

// PostgresFS holds the Postgres migrations under the "postgres" directory.
// Consumers use iofs.New(PostgresFS, "postgres") or fs.Sub(PostgresFS, "postgres").
//
//go:embed postgres/*.sql
var PostgresFS embed.FS
