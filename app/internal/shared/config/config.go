package config

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
)

var AppConfig PassConfig

// Config represents the complete application configuration
type Config struct {
	PassConfig `mapstructure:",squash"`
	Database   domain.DatabaseConfig `mapstructure:"database"`
	Reminder   ReminderConfig        `mapstructure:"reminder"`
}

// ReminderConfig contains configuration for the reminder email system
type ReminderConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	CheckAfterHours   int  `mapstructure:"checkafterhours"`   // Default: 24
	MaxReminders      int  `mapstructure:"maxreminders"`      // Default: 3
	ReminderInterval  int  `mapstructure:"reminderinterval"`  // Default: 24 hours
}

type PassConfig struct {
	EmailHost             string `mapstructure:"emailhost"`
	EmailUser             string `mapstructure:"emailuser"`
	EmailPass             string `mapstructure:"emailpass"`
	EmailFrom             string `mapstructure:"emailfrom"`
	RabHost               string `mapstructure:"rabhost"`
	RabUser               string `mapstructure:"rabuser"`
	RabPass               string `mapstructure:"rabpass"`
	RabQName              string `mapstructure:"rabqname"`
	DbHost                string `mapstructure:"dbhost"`
	DbUser                string `mapstructure:"dbuser"`
	DbPass                string `mapstructure:"dbpass"`
	DbName                string `mapstructure:"dbname"`
	ProdHost              string `mapstructure:"prodhost"`
	DevHost               string `mapstructure:"devhost"`
	EncryptionProdService string `mapstructure:"encryptionprodservice"`
	DatabaseProdService   string `mapstructure:"databaseprodservice"`
	EncryptionDevService  string `mapstructure:"encryptiondevservice"`
	DatabaseDevService    string `mapstructure:"databasedevservice"`
	Loglevel              string `mapstructure:"loglevel"`
	RunningEnvironment    string `mapstructure:"runningenvironment"`
	DefaultMaxViewCount   int    `mapstructure:"defaultmaxviewcount"`
	DbPort                int    `mapstructure:"dbport"`
	EmailPort             int    `mapstructure:"emailport"`
	RabPort               int    `mapstructure:"rabport"`
}
