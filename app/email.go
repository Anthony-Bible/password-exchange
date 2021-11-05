package main

import (
  "github.com/rs/zerolog/log"
  "bytes"
  "fmt"
  "errors"
  "net/smtp"
  "text/template"
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
      log.Error().Err(err).Msg("")
      return "not right", err

    }
    // if !ok {
    //  log.Fatalf("Invalid type assertion for %s", envname)
    // }


}
func  (msg *MessagePost) Deliver() error {
   //set neccessary info for environment variables

  // Sender data.
  password,err := GetViperVariable("emailpass")
  if err != nil {
		panic(err)
	}
  from := "server@password.exchange"
  AWS_ACCESS_KEY_ID,err := GetViperVariable("emailuser")
  if err != nil {
		panic(err)
	}
  // Receiver email address.
  to := msg.OtherEmail
  // smtp server configuration.
  emailhost, err := GetViperVariable("emailhost") 
  if err != nil {
		panic(err)
	}
  // smtpPort := GetViperVariable("emailport")


  // Authentication.
  auth := smtp.PlainAuth("", AWS_ACCESS_KEY_ID, password, emailhost)

  t, _ := template.ParseFiles("templates/email_template.html")

  var body bytes.Buffer

  mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
  body.Write([]byte(fmt.Sprintf("Subject: This is a test subject \n%s\n\n", mimeHeaders)))
  
  t.Execute(&body, struct {
    Body    string
    Message string
  }{
    Body:    fmt.Sprintf("Hi %s, \n %s used our service at https://passsword.exchange to send a secure message to you. We've included a link to view the message below, to find out more information go to https://password.exchange/about", msg.OtherFirstName, msg.FirstName),
    Message:  msg.Content,
  })

  // Sending email.
  err = smtp.SendMail(emailhost, auth, from, to, body.Bytes())
  fmt.Println(err)
  return err
}
