package controllers

import (
	"log"
	"time"
	"password.exchange/slackbot/views"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// We create a sctucture to let us use dependency injection
type SlashCommandController struct {
	EventHandler *socketmode.SocketmodeHandler
}


func NewSlashCommandController(eventhandler *socketmode.SocketmodeHandler) SlashCommandController {
	
	c := SlashCommandController{
		EventHandler: eventhandler,
	}
	
	// Register callback for the command /rocket
	c.EventHandler.HandleSlashCommand(
		"/rocket",
		c.launchRocketAnnoncement,
	)
	log.Printf("Slash command callback")
	// The rocket launch is approved
	c.EventHandler.HandleInteractionBlockAction(
		views.RocketAnnoncementActionID,
		c.launchRocket,
	)
	
	return c
}
func (c SlashCommandController) launchRocketAnnoncement(evt *socketmode.Event, clt *socketmode.Client) {
	// we need to cast our socketmode.Event into a Slash Command
	command, ok := evt.Data.(slack.SlashCommand)
	if ok != true {
		log.Printf("ERROR converting event to Slash Command: %v", ok)
	}
	// Make sure to respond to the server to avoid an error
	clt.Ack(*evt.Request)
	// parse the command line (Hardcoded in this example)
	count := 3
	// create the view using block-kit
	blocks := views.LaunchRocketAnnoncement(count)
	// Post ephemeral message
	_, _, err := clt.GetApiClient().PostMessage(
		command.ChannelID,
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionResponseURL(command.ResponseURL, slack.ResponseTypeEphemeral),
	)
	// Handle errors
	if err != nil {
		log.Printf("ERROR while sending message for /rocket: %v", err)
	}
}
func (c SlashCommandController) launchRocket(evt *socketmode.Event, clt *socketmode.Client) {
	// we need to cast our socketmode.Event into an Interaction Callback
	interaction := evt.Data.(slack.InteractionCallback)
	// Make sure to respond to the server to avoid an error
	clt.Ack(*evt.Request)
	// parse the command line
	count := 3
	for i := count; i >= 0; i-- {
		// create the view using block-kit
		blocks := views.LaunchRocket(i)
		// count down by steps of 1s
		time.Sleep(1000 * time.Millisecond)
		_, _, err := clt.GetApiClient().PostMessage(
			interaction.Container.ChannelID,
			slack.MsgOptionBlocks(blocks...),
			slack.MsgOptionResponseURL(interaction.ResponseURL, slack.ResponseTypeInChannel),
			slack.MsgOptionReplaceOriginal(interaction.ResponseURL),
		)
		// Handle errors
		if err != nil {
			log.Printf("ERROR while sending message for /rocket: %v", err)
		}
	}
}