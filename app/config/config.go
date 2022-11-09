package config

var Config PassConfig

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
	DbPort                int    `mapstructure:"dbport"`
	EmailPort             int    `mapstructure:"emailport"`
	RabPort               int    `mapstructure:"rabport"`
}
