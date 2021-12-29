package commons

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

func GetViperVariable(envname string) (string, error) {
	viper.SetEnvPrefix("passwordexchange") // will be uppercased automatically
	viper.AutomaticEnv()                   //will automatically load every env variable with PASSWORDEXCHANGE_
	if viper.IsSet(envname) {
		viperReturn := viper.GetString(envname)
		return viperReturn, nil
	} else {
		err := errors.New(fmt.Sprintf("Environment  variable not set %s", envname))
		log.Fatal().Err(err).Msg("")
		return "not right", err
	}
	// if !ok {
	//  log.Fatalf("Invalid type assertion for %s", envname)
	// }

}
