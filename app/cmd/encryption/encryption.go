/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package encryption

import (
	"context" // Added
	"log"     // Added
	"os"      // Added
	"reflect"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config" // Added
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging" // Added
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// encryptionCmd represents the encryption command
var encryptionCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Starts the encryption service.", // Updated
	Long:  `Starts the gRPC server for the encryption service, which handles message encryption and decryption.`, // Updated
	Run: func(cmdcobra *cobra.Command, args []string) { // Renamed cmd to cmdcobra
		appFullConfig, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("CRITICAL: Failed to load application configuration: %v", err)
		}

		serviceLogger, err := logging.NewBridgeLogger(appFullConfig.Log, appFullConfig.EnableSlog, "encryption-service")
		if err != nil {
			log.Fatalf("CRITICAL: Failed to initialize logger: %v", err)
		}
		serviceLogger.Info(context.Background(), "Encryption command initiated, logger active")

		var cmdLocalConfig Config // This is the Config from encryption2.go

		// The existing bindenvs and Unmarshal logic.
		// Note: The original `viper.Unmarshal(&cfg.PassConfig)` was very specific.
		// The new version `viper.Unmarshal(&cmdLocalConfig)` assumes `Config` struct in encryption2.go
		// is structured to be unmarshalled directly or `PassConfig` is squashed and other fields are handled.
		// This was part of the prompt's new content for encryption.go.
		bindenvs(cmdLocalConfig) // Call existing bindenvs
		if err := viper.Unmarshal(&cmdLocalConfig); err != nil {
			serviceLogger.Error(context.Background(), "Failed to unmarshal command-specific config for encryption", "error", err)
			os.Exit(1)
		}
		// TODO: Review if cmdLocalConfig.Channel needs specific initialization here.

		if err := cmdLocalConfig.startServer(serviceLogger); err != nil {
			serviceLogger.Error(context.Background(), "Encryption service command execution failed", "error", err)
			os.Exit(1)
		}
		serviceLogger.Info(context.Background(), "Encryption service command completed successfully.")
	},
}

// this is required due to viper not automatically mapping env to marshal https://github.com/spf13/viper/issues/584
// Keeping existing bindenvs function
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
	cmd.RootCmd.AddCommand(encryptionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// encryptionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// encryptionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
