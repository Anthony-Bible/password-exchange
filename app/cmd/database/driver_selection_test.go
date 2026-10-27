package database

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An unknown PASSWORDEXCHANGE_DBDRIVER must stop `app database` with a
// non-zero exit and a message that names the variable and the bad value.
func TestBinary_UnknownDBDriverExitsNonZeroNamingTheVariable(t *testing.T) {
	bin := buildAppBinary(t)

	res := runBinary(t, bin, t.TempDir(), 60, map[string]string{
		"PASSWORDEXCHANGE_DBDRIVER": "bogus-db",
		"PASSWORDEXCHANGE_DBHOST":   "127.0.0.1",
		"PASSWORDEXCHANGE_DBPORT":   "5432",
		"PASSWORDEXCHANGE_DBUSER":   "u",
		"PASSWORDEXCHANGE_DBPASS":   "p",
		"PASSWORDEXCHANGE_DBNAME":   "d",
	})

	require.False(t, res.timedOut, "database service must fail fast, not start serving")
	assert.NotZero(t, res.exitCode)
	assert.Contains(t, res.output, "PASSWORDEXCHANGE_DBDRIVER")
	assert.Contains(t, res.output, "bogus-db")
}

// The pool budget (#568: 50 connections shared by dev and prod) requires a
// conservative default of 5 open connections, overridable per environment.
func TestDBMaxOpenConns_DefaultsToFive(t *testing.T) {
	resetConfig(t)
	viper.Set("dbhost", "127.0.0.1")
	viper.Set("dbport", 3306)

	require.NoError(t, initConfigAndMigrator())

	assert.Equal(t, 5, cfg.PassConfig.DbMaxOpenConns)
}

func TestDBMaxOpenConns_ReadFromEnvironment(t *testing.T) {
	resetConfig(t)
	t.Setenv("PASSWORDEXCHANGE_DBMAXOPENCONNS", "3")
	t.Setenv("PASSWORDEXCHANGE_DBHOST", "127.0.0.1")
	t.Setenv("PASSWORDEXCHANGE_DBPORT", "3306")

	require.NoError(t, initConfigAndMigrator())

	assert.Equal(t, 3, cfg.PassConfig.DbMaxOpenConns)
}
