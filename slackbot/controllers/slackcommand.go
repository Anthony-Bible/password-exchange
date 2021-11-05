package controllers

import (
	"github.com/rs/zerolog/log"
	// "time"
	// "password.exchange/slackbot/views"
	"password.exchange/slackbot/drivers"
	"net/http"
	 "fmt"
	 "github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)
func SlashCommandHandler(c *gin.Context) {
	s, err := slack.SlashCommandParse(c.Request)
	if err != nil {
		c.String(http.StatusUnauthorized, fmt.Sprintf("error: %s", err))
		return
	}
	token,err :=drivers.GetViperVariable("slack_verification_token")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error().Msg("something went wrong with getting the slack token variable")
		return
	}
	log.Debug().Msg(token)
	// verifier, err :=slack.NewSecretsVerifier(c.Request.Header, token)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Error().Msg("something went wrong with veryifying slack authentication")
		return
	}

	
	// if  err = verifier.Ensure(); err != nil {
	// log.Warn().Msg("Unauthorized attempt")
	// //   w.WriteHeader(http.StatusUnauthorized)
	//   c.String(http.StatusUnauthorized, fmt.Sprintf("error: %s", err))
	//   for k, vals := range c.Request.Header {
	// 	fmt.Printf("\t%s", k)
	// 	for _, v := range vals {
	// 		fmt.Printf("\t%s", v)
	// 	}
	// }
	
	
	switch  s.Command {
	  case "/encrypt":
	  params := &slack.Msg{Text: s.Text}
	  response := fmt.Sprintf("You asked for the weather for %v", params.Text)
	//   w.Write([]byte(response))
	log.Info().Msg("success")
	c.String(200,response)

	  default:
	  c.AbortWithStatus(http.StatusInternalServerError)

	  return
	}
  }