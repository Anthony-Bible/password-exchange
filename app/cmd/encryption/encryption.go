/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package encryption

import (
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// encryptionCmd represents the encryption command
var encryptionCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Component to do the encryption",
	Long: `
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("encryption called")
		var cfg Config
		config.BindEnvs(cfg)
		viper.Unmarshal(&cfg.PassConfig)
		cfg.startServer()
	},
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
