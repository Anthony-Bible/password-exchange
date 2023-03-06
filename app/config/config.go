// Package config has the config struct that makes it a centralized place to put
// user configurable values
package config

// Config is the struct that holds all the user configurable values
var Config PassConfig

//PassConfig These are all options configurable by the user
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
	DatabaseProdService   string `mapstructure:"encryptionprodservice"`
	EncryptionDevService  string `mapstructure:"encryptiondevservice"`
	DatabaseDevService    string `mapstructure:"databasedevservice"`
	Loglevel              string `mapstructure:"loglevel"`
	RunningEnvironment    string `mapstructure:"runningenvironment"`
	S3Key                 string `mapstructure:"s3key"`
	S3ID                  string `mapstructure:"s3id"`
	S3Endpoint            string `mapstructure:"s3endpoint"`
	S3Bucket              string `mapstructure:"s3bucket"`
	S3Region              string `mapstructure:"s3region"`
	GroupcacheSvcName     string `mapstructure:"groupcachesvcname"`
	S3Ssl                 bool   `mapstructure:"s3ssl"`
	DbPort                int    `mapstructure:"dbport"`
	EmailPort             int    `mapstructure:"emailport"`
	RabPort               int    `mapstructure:"rabport"`
}
