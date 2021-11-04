// forms.go
package main

import (
	"github.com/rs/zerolog/log"
	"github.com/gin-gonic/gin"
    // "net/http"
    // "net/url"
    // "password.exchange/slackbot/drivers"
    // "password.exchange/slackbot/controllers"
    // "github.com/slack-go/slack/socketmode"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.POST("/api/:app/*action",doAction)
	router.POST("/api/:app",doAction)
  
	router.NoRoute(failedtoFind)
	  // By default it serves on :8080 unless a
	  // PORT environment variable was defined.
	router.Run()
	log.Info().Msg("Listening...")

  
  
  }
// log.Info().Msg("connected to slack")
// socketmodeHandler := socketmode.NewsSocketmodeHandler(client)
// log.Info().Msg("Listening...")

// // Build a Slack App Home in Golang Using Socket Mode
//  controllers.NewAppHomeController(socketmodeHandler)
// log.Info().Msg("New app home")

// // // Properly Welcome Users in Slack with Golang using Socket Mode
//  controllers.NewGreetingController(socketmodeHandler)
// log.Info().Msg("Greeting controller")

// // Build Slack Slash Command in Golang Using Socket Mode
// controllers.NewSlashCommandController(socketmodeHandler)
// log.Info().Msg("slash controller")


// socketmodeHandler.RunEventLoop()
// }