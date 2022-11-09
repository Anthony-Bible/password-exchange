/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "passwordexchange",
	Short: "Cli to start password exchange",
	Long: `This cli helps start the seperate components of password Exchange
      The purpose of this app is to make it easy to share one time secure 
    information like passwords without prior requisites (installation or accounts). 

    To use run passwordexchange <component>:
      passwordexchange email - start up the email consumer to send emails
      passwordexchange server - start up the web component
      passwordexchange database - start up the component that interacts with the
        database
      passwordexchange encryption - start up the component that does encrytion`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

var loglevel string
var cfgFile string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	cobra.OnInitialize(initConfig)
	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.app.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	RootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "info", "Logging level for the application, Default: info")
	viper.BindPFlag("loglevel", RootCmd.PersistentFlags().Lookup("loglevel"))
}
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		log.Error().Err(errors.New("config file not set")).Msg("Config file not set")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Info().Msg("no config file can be found")
	} else {
		log.Info().Msgf("Using config file: %s", cfgFile)
	}
	viper.SetEnvPrefix("passwordexchange")
	viper.AutomaticEnv()
}
