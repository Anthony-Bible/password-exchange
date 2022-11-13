/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package email

import (
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/rs/zerolog/log"

	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// emailCmd represents the email command
var emailCmd = &cobra.Command{
	Use:        "email",
	Aliases:    []string{},
	SuggestFor: []string{},
	Short:      "Run the component in charge of sending emails",
	GroupID:    "",
	Long: `This component consumes from rabbitmq the emails to send. It uses
      configurable options to connect via SMTP to send emails:
    PASSWORDEXCHANGE_EMAILUSER: User to connect as
    PASSWORDEXCHANGE_EMAILPASS: Password for email user
    PASSWORDEXCHANGE_EMAILHOST: Email host to use as a relay
    PASSWORDEXCHANGE_EMAILPORT: Port for email host`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("email called")
		log.Debug().Msgf("the value of loglevel is %s\n", viper.Get("loglevel"))
		var cfg Config
		bindenvs(cfg)
		viper.Unmarshal(&cfg.PassConfig)
		cfg.StartProcessing()
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
	cmd.RootCmd.AddCommand(emailCmd)

	// Here you will define your flags and configuration settings.
	emailCmd.Flags().String("emailuser", "", "User to log in with for SMTP authentication")
	emailCmd.Flags().String("emailpass", "", "pass to log in with for SMTP authentication")
	emailCmd.Flags().String("emailhost", "", "host to log in with for SMTP authentication")
	emailCmd.Flags().String("emailport", "", "port to log in with for SMTP authentication")
	emailCmd.Flags().String("rabuser", "", "User to log in with for rabbitmq authentication")
	emailCmd.Flags().String("rabhost", "", "host to log in with for rabbitmq authentication")
	emailCmd.Flags().String("rabpass", "", "password to log in with for rabbitmq authentication")
	emailCmd.Flags().String("rabport", "", "port to log in with for rabbitmq authentication")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// emailCmd.PersistentFlags().String("foo", "", "A help for foo")

	viper.BindPFlag("emailuser", emailCmd.Flags().Lookup("emailuser"))
	viper.BindPFlag("emailpass", emailCmd.Flags().Lookup("emailpass"))
	viper.BindPFlag("emailport", emailCmd.Flags().Lookup("emailport"))
	viper.BindPFlag("emailhost", emailCmd.Flags().Lookup("emailhost"))
	viper.BindPFlag("rabuser", emailCmd.Flags().Lookup("rabuser"))
	viper.BindPFlag("rabpass", emailCmd.Flags().Lookup("rabpass"))
	viper.BindPFlag("rabport", emailCmd.Flags().Lookup("rabport"))
	viper.BindPFlag("rabhost", emailCmd.Flags().Lookup("rabhost"))
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// emailCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
