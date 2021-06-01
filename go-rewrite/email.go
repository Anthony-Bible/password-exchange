package main

import (
  "bytes"
  "fmt"
  "log"
  "net/smtp"
  "text/template"
  "github.com/spf13/viper"
)
func GetViperVariable(envname string) string {
    viperReturn, ok := viper.Get(envname).(string)

    if !ok {
     log.Fatalf("Invalid type assertion for %s", envname)
   }
   return viperReturn

}
func  (msg *Message) Deliver() error {
   //set neccessary info for environment variables
  viper.SetEnvPrefix("passwordexchange") // will be uppercased automatically
  viper.AutomaticEnv() //will automatically load every env variable with PASSWORDEXCHANGE_
  // Sender data.
  password := GetViperVariable("emailpass")
  from := "server@password.exchange"
  AWS_ACCESS_KEY_ID := GetViperVariable("emailuser")

  // Receiver email address.
  to := []string{
    msg.Email,
  }
  fmt.Println(GetViperVariable("emailhost"))
  // smtp server configuration.
  smtpHost := GetViperVariable("emailhost") + ":" + GetViperVariable("emailport")
  // smtpPort := GetViperVariable("emailport")
  fmt.Println(smtpHost)


  // Authentication.
  auth := smtp.PlainAuth("", AWS_ACCESS_KEY_ID, password, GetViperVariable("emailhost"))

  t, _ := template.ParseFiles("templates/email_template.html")

  var body bytes.Buffer

  mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
  body.Write([]byte(fmt.Sprintf("Subject: This is a test subject \n%s\n\n", mimeHeaders)))

  // t.Execute(&body, struct {
  //   Name    string
  //   Message string
  // }{
  //   Name:    "Puneet Singh",
  //   Message: "This is a test message in a HTML template",
  // })
  b := []byte(msg.Content)

  body.Write(b)
  // Sending email.
  err := smtp.SendMail(smtpHost, auth, from, to, body.Bytes())
  fmt.Println(err)
  return err
}
