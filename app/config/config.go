// Package config has the config struct that makes it a centralized place to put
// user configurable values
package config

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
	S3apiKey              string `mapstructure:"s3apikey"`
	S3apiID               string `mapstructure:"s3apiid"`
	S3apiEndpoint         string `mapstructure:"s3apiendpoint"`
	S3apiBucket           string `mapstructure:"s3apibucket"`
	S3apiRegion           string `mapstructure:"s3apiregion"`
	GroupcacheSvcName     string `mapstructure:"groupcachesvcname"`
	S3apiSsl              bool   `mapstructure:"s3apissl"`
	DbPort                int    `mapstructure:"dbport"`
	EmailPort             int    `mapstructure:"emailport"`
	RabPort               int    `mapstructure:"rabport"`
}
