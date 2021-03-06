// forms.go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "net/http"
    "fmt"
    "github.com/rs/xid"
)


func main() {
  router := gin.Default()
  router.LoadHTMLGlob("templates/*")
  router.GET("/", home)
  router.POST("/", send)
  router.GET("/confirmation", confirmation)
  router.GET("/encrypt")
  router.NoRoute(failedtoFind)
  log.Println("Listening...")

  // if err != nil {
  //   log.Fatal(err)
  // }
  router.Run(":8080")

}

func home(c *gin.Context) {
  render(c, "home.html", nil)
}
func failedtoFind(c *gin.Context) {
  render(c, "404.html", nil)
}
func send(c *gin.Context) {
  // Step 1: Validate form
  // Step 2: Send message in an email
  // Step 3: Redirect to confirmation page
  encryptionstring, randmError := GenerateRandomString(32)
  if randmError != nil {
    log.Fatal(randmError)
  }
  guid := xid.New()
  siteHost := GetViperVariable("host")
  Connect()
  msgEncrypted := &Message{
		Email:   Encrypt(c.PostForm("email"), encryptionstring),
    FirstName: Encrypt(c.PostForm("firstname"), encryptionstring),
    OtherFirstName: Encrypt(c.PostForm("other_firstname"), encryptionstring),
    OtherLastName: Encrypt(c.PostForm("other_lastname"), encryptionstring),
    OtherEmail: Encrypt(c.PostForm("other_email"), encryptionstring),
    Uniqueid: guid.String(),
  }
  msg := &Message{
    c.PostForm("email"),
    c.PostForm("firstname"),
    c.PostForm("other_firstname"),
    c.PostForm("other_lastname"),
    c.PostForm("other_email"),

  }
  
  


	if msg.Validate() == false {
		render(c, "home.html", msg)
		return
	}
  msg.Content = "please click this link to get your encrypted message" +  "\n" + siteHost + "encrypt/" + msgEncrypted.Uniqueid + "/" + encryptionstring
  Insert(msgEncrypted)

	if err := msg.Deliver(); err != nil {
		log.Println(err)
    c.String(http.StatusInternalServerError, fmt.Sprintf("something wwnet wrong: %s", err))

		return
	}
	c.Redirect( http.StatusSeeOther, "/confirmation")

}
  
func confirmation(c *gin.Context) {
  render(c, "confirmation.html", nil)
}
func render(c *gin.Context, filename string, data interface{}) {

    

      // Call the HTML method of the Context to render a template
      c.HTML(
        // Set the HTTP status to 200 (OK)
        http.StatusOK,
        // Use the index.html template
        filename,
        // Pass the data that the page uses (in this case, 'title')
        gin.H{
            "title": "Password Exchange", 
        })
    
    
    
  }

//   if err := tmpl.Execute(w, data); err != nil {
//     log.Println(err)
//     http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
//   }
// }

