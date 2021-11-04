package drivers

import (
	"errors"
	"github.com/rs/zerolog/log"
	// "os"
	// "strings"
    "fmt"
	// "github.com/slack-go/slack"
	// "github.com//slack-go/slack/socketmode"
	"github.com/spf13/viper"
)
func GetViperVariable(envname string) (string,error) {
    viper.SetEnvPrefix("passwordexchange") // will be uppercased automatically
    viper.AutomaticEnv() //will automatically load every env variable with PASSWORDEXCHANGE_
    if viper.IsSet(envname){
      viperReturn := viper.GetString(envname)
      return viperReturn, nil
    }else{
		err := errors.New(fmt.Sprintf("Environment  variable not set %s", envname))
		log.Fatal().Err(err).Msg("")
		return "not right", err
	}
    // if !ok {
    //  log.Fatalf("Invalid type assertion for %s", envname)
    // }


}

// func ConnectToSlackViaSocketmode() (*socketmode.Client, error) {

// 	appToken := GetViperVariable("slackkey")
// 	if appToken == "" {
// 		return nil, errors.New("SLACK_APP_TOKEN must be set")
// 	}

// 	if !strings.HasPrefix(appToken, "xapp-") {
// 		return nil, errors.New("SLACK_APP_TOKEN must have the prefix \"xapp-\".")
// 	}

// 	botToken := GetViperVariable("botkey")
// 	if botToken == "" {
// 		return nil, errors.New("SLACK_BOT_TOKEN must be set.")
// 	}

// 	if !strings.HasPrefix(botToken, "xoxb-") {
// 		return nil, errors.New("SLACK_BOT_TOKEN must have the prefix \"xoxb-\".")
// 	}

// 	api := slack.New(
// 		botToken,
// 		slack.OptionDebug(true),
// 		slack.OptionAppLevelToken(appToken),
// 		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
// 	)

// 	client := socketmode.New(
// 		api,
// 		socketmode.OptionDebug(true),
// 		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
// 	)

// 	return client, nil
// }