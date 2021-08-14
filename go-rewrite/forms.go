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
  encryptionstring := GenerateRandomString()
  guid := xid.New()
  siteHost := GetViperVariable("host")
  fmt.Printf("type of postform email: %T\n", c.PostForm("email"))

  msgEncrypted := &Message{
		Email:   string(MessageEncrypt([]byte(c.PostForm("email")), encryptionstring)),
    FirstName: string(MessageEncrypt([]byte(c.PostForm("firstname")), encryptionstring)),
    OtherFirstName: string(MessageEncrypt([]byte(c.PostForm("other_firstname")), encryptionstring)),
    OtherLastName: string(MessageEncrypt([]byte(c.PostForm("other_lastname")), encryptionstring)),
    OtherEmail: string(MessageEncrypt([]byte(c.PostForm("other_email")), encryptionstring)),
    Uniqueid: guid.String(),
  }
  fmt.Printf("this is the encrypted Message %s", msgEncrypted)
  msg := &MessagePost{
    Email: []string{c.PostForm("email")},
    FirstName: c.PostForm("firstname"),
    OtherFirstName: c.PostForm("other_firstname"),
    OtherLastName: c.PostForm("other_lastname"),
    OtherEmail: c.PostForm("other_email"),

  }
  
  


	if msg.Validate() == false {
		render(c, "home.html", msg)
		return
	}
  msg.Content = "please click this link to get your encrypted message" +  "\n" + siteHost + "encrypt/" + msgEncrypted.Uniqueid + "/" + string(encryptionstring[:])
  Insert(msgEncrypted)
  fmt.Sprintf("this is the msgEncrypted: %s", msgEncrypted)


	if err := msg.Deliver(); err != nil {
		log.Println(err)
    c.String(http.StatusInternalServerError, fmt.Sprintf("something went wrong: %s", err))

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

