package email

import (
  "bytes"
  "fmt"
  "net/smtp"
  "text/template"
  "password.exchange/commons"   
  "password.exchange/message"
)

func  Deliver(msg *message.MessagePost) error {
   //set neccessary info for environment variables

  // Sender data.
  password := commons.GetViperVariable("emailpass")
  from := "server@password.exchange"
  AWS_ACCESS_KEY_ID := commons.GetViperVariable("emailuser")

  // Receiver email address.
  to := msg.Email
  fmt.Println(commons.GetViperVariable("emailhost"))
  // smtp server configuration.
  smtpHost := commons.GetViperVariable("emailhost") + ":" + commons.GetViperVariable("emailport")
  // smtpPort := commons.GetViperVariable("emailport")
  fmt.Println(smtpHost)


  // Authentication.
  auth := smtp.PlainAuth("", AWS_ACCESS_KEY_ID, password, commons.GetViperVariable("emailhost"))

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
  err := smtp.SendMail(smtpHost, auth, from, to, body.Bytes())
  fmt.Println(err)
  return err
}
