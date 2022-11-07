package email

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	"github.com/Anthony-Bible/password-exchange/app/message"
	"github.com/rs/zerolog/log"
)

// if !ok {
//  log.Fatalf("Invalid type assertion for %s", envname)
// }
// type MyMessage message.MessagePost

func (conf Config) Deliver(msg *message.MessagePost) error {
	//set neccessary info for environment variables

	// Sender data.
	// Receiver email address.
	to := msg.OtherEmail
	// smtp server configuration.
	fullhost := fmt.Sprintf("%s:%d", conf.EmailHost, conf.EmailPort)
	// Authentication.
	auth := smtp.PlainAuth("", conf.EmailUser, conf.EmailPass, conf.EmailHost)

	t, err := template.ParseFiles("/templates/email_template.html")
	if err != nil {
		log.Error().Err(err).Msg("template not found")

		return err
	}

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := []byte("From: Password Exchange <server@password.exchange>\r\n" + "To: " + strings.Join(to, "") + "\r\n" +
		fmt.Sprintf("Subject: Encrypted Messsage from Password exchange from %s \r\n", msg.FirstName) +
		mimeHeaders)
	buf := bytes.NewBuffer(body)
	err = t.Execute(buf, struct {
		Body    string
		Message string
	}{
		Body:    fmt.Sprintf("Hi %s, \n %s used our service at <a href=\"https://password.exchange\"> Password Exchange </a> to send a secure message to you. We've included a link to view the message below, to find out more information go to https://password.exchange/about", msg.OtherFirstName, msg.FirstName),
		Message: msg.Content,
	})

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with rendering email template")
		return err
	}
	// Sending email.
	if err = smtp.SendMail(fullhost, auth, conf.EmailFrom, to, buf.Bytes()); err != nil {
		log.Error().Err(err).Msgf("emailhost: %s from: %s to: %s authHost: %s", conf.EmailHost, conf.EmailFrom, to, conf.EmailHost)
	}

	return err
}
