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

func GetViperVariable(envname string) (string, error) {
	viper.SetEnvPrefix("passwordexchange") // will be uppercased automatically
	viper.AutomaticEnv()                   //will automatically load every env variable with PASSWORDEXCHANGE_
	if viper.IsSet(envname) {
		viperReturn := viper.GetString(envname)
		return viperReturn, nil
	} else {
		err := errors.New(fmt.Sprintf("Environment  variable not set %s", envname))
		log.Error().Err(err).Msg("")
		return "not right", err

	}
}

// if !ok {
//  log.Fatalf("Invalid type assertion for %s", envname)
// }
// type MyMessage message.MessagePost

func Deliver(msg *message.MessagePost) error {
	//set neccessary info for environment variables

	// Sender data.
	password, err := GetViperVariable("emailpass")
	if err != nil {
		panic(err)
	}
	from := "server@password.exchange"
	AWS_ACCESS_KEY_ID, err := GetViperVariable("emailuser")
	if err != nil {
		panic(err)
	}
	// Receiver email address.
	to := msg.OtherEmail
	// smtp server configuration.
	authHost, err := GetViperVariable("emailhost")
	if err != nil {
		panic(err)
	}
	emailPort, err := GetViperVariable("emailport")
	if err != nil {
		return err
	}
	emailHost := authHost + ":" + emailPort
	// Authentication.
	auth := smtp.PlainAuth("", AWS_ACCESS_KEY_ID, password, authHost)

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
	if err = smtp.SendMail(emailHost, auth, from, to, buf.Bytes()); err != nil {
		log.Error().Err(err).Msgf("emailhost: %s from: %s to: %s authHost: %s", emailHost, from, to, authHost)
	}

	return err
}
