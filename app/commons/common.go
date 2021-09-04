package commons
import (
	"fmt"
	"github.com/spf13/viper"
  )
func GetViperVariable(envname string) string {
    viper.SetEnvPrefix("passwordexchange") // will be uppercased automatically
    viper.AutomaticEnv() //will automatically load every env variable with PASSWORDEXCHANGE_
    if viper.IsSet(envname){
      viperReturn := viper.GetString(envname)
      return viperReturn
    }else{
      panic(fmt.Sprintf("Environment  variable not set %s", envname))
    }
    // if !ok {
    //  log.Fatalf("Invalid type assertion for %s", envname)
    // }


}