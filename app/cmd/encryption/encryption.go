/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package encryption

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
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
		bindenvs(cfg)
		viper.Unmarshal(&cfg.PassConfig)
		fmt.Printf("CFG Is %+v\n", cfg)
		cfg.startServer()
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
	cmd.RootCmd.AddCommand(encryptionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// encryptionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// encryptionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
