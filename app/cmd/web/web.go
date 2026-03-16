/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package web

import (
	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start main web server",
	Long: `This component is the main entry point into the application. It starts
    the webserver which is responsible for displaying the webpage.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg Config
		config.BindEnvs(cfg)
		viper.Unmarshal(&cfg.PassConfig)
		cfg.StartServer()
	},
}

func init() {
	cmd.RootCmd.AddCommand(webCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// webCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// webCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
