// forms.go
package main

import (
	"github.com/rs/zerolog/log"
	"github.com/gin-gonic/gin"
    // "net/http"
    // "net/url"
    // "password.exchange/slackbot/drivers"
    "password.exchange/slackbot/controllers"
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


  func failedtoFind(c *gin.Context) {
	render(c, "404.html", 404, nil)
  }
  
  
  func doAction(c *gin.Context) {
	// for key, value := range c.Request.PostForm {
	// 	log.Printf("%v = %v \n",key,value)
	// }
    controllers.SlashCommandHandler(c)
  }


  func render(c *gin.Context, filename string, status int, data interface{}) {

    
	if status == 0{
	  status=200
	}

   // Call the HTML method of the Context to render a template
   c.HTML(
	 // Set the HTTP status to 200 (OK)
	 //TODO: have this be settable
	 status,
	 // Use the index.html template
	 filename,
	 // Pass the data that the page uses (in this case, 'title')
	 data,
   )
 
 
 
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