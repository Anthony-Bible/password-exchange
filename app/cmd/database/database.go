/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package database

import (
	"context"
	"log" // For critical startup errors
	"os"  // For os.Exit
	"reflect"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// databaseCmd represents the database command
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmdCobra *cobra.Command, args []string) { // Renamed cmd to cmdCobra to avoid conflict with package name
		appFullConfig, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("CRITICAL: Failed to load application configuration: %v", err)
		}

		// Initialize the new logger
		serviceLogger, err := logging.NewBridgeLogger(appFullConfig.Log, appFullConfig.EnableSlog, "database-service")
		if err != nil {
			log.Fatalf("CRITICAL: Failed to initialize logger: %v", err)
		}
		serviceLogger.Info(context.Background(), "Database command initiated, logger active")

		var cfg Config // This is the local Config type defined in database2.go

		// The existing bindenvs and Unmarshal logic for the command's specific config (cfg.PassConfig)
		// It's assumed that `bindenvs` correctly sets up viper for the subsequent Unmarshal into cfg.PassConfig
		bindenvs(cfg)
		if err := viper.Unmarshal(&cfg.PassConfig); err != nil {
			 serviceLogger.Error(context.Background(), "Failed to unmarshal PassConfig specific to database command", "error", err)
			 os.Exit(1)
		}

		// Call startServer with the logger
		if err := cfg.startServer(serviceLogger); err != nil {
			serviceLogger.Error(context.Background(), "Database service command execution failed", "error", err)
			os.Exit(1)
		}
		serviceLogger.Info(context.Background(), "Database service command completed successfully.")
	},
}

// this is required due to viper not automatically mapping env to marshal https://github.com/spf13/viper/issues/584
func bindenvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	if ifv.Kind() == reflect.Ptr {
		ifv = ifv.Elem()
	}
	for i := 0; i < ifv.NumField(); i++ {
		v := ifv.Field(i)
		t := ifv.Type().Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		if tv == ",squash" {
			bindenvs(v.Interface(), parts...)
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindenvs(v.Interface(), append(parts, tv)...)
		default:
			viper.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}
func init() {
	cmd.RootCmd.AddCommand(databaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// databaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// databaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
