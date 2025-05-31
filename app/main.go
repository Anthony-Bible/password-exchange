/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/Anthony-Bible/password-exchange/app/cmd"
	_ "github.com/Anthony-Bible/password-exchange/app/cmd/database"
	_ "github.com/Anthony-Bible/password-exchange/app/cmd/email"
	_ "github.com/Anthony-Bible/password-exchange/app/cmd/encryption"
	_ "github.com/Anthony-Bible/password-exchange/app/cmd/reminder"
	_ "github.com/Anthony-Bible/password-exchange/app/cmd/web"
)

func main() {
	cmd.Execute()
}
