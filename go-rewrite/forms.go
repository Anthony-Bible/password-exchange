// forms.go
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    b64 "encoding/base64"
    "net/http"
    "fmt"
    "github.com/rs/xid"
)

type htmlHeaders struct{
  Title string
  Url string
  DecryptedMessage string
  
}
func main() {
  router := gin.Default()
  router.LoadHTMLGlob("templates/*")
  router.Static("/assets", "./assets")
  router.GET("/", home)
  router.POST("/", send)
  router.GET("/confirmation", confirmation)
  router.GET("/encrypt/:uuid/*key", displaydecrypted)
  router.NoRoute(failedtoFind)
  log.Println("Listening...")

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
  router.Run()

}

func home(c *gin.Context) {
  render(c, "home.html", nil)
}
func failedtoFind(c *gin.Context) {
  render(c, "404.html", nil)
}
func displaydecrypted(c *gin.Context) {
  uuid := c.Param("uuid")
  key := c.Param("key")
  decodedKey, err := b64.URLEncoding.DecodeString(key[1:])
  if err != nil {
		panic(err)
	}
  message := Select(uuid)
  decodedContent, err := b64.URLEncoding.DecodeString(message.Content)
  if err != nil {
		panic(err)
	}
  var arr [32]byte
  copy(arr[:], decodedKey)
  content, err :=MessageDecrypt([]byte(decodedContent), &arr )
  if err != nil {
		panic(err)
	}
  msg := &MessagePost{
    Content: string(content),
  }
  extraHeaders :=htmlHeaders{Title: "passwordExchange Decrypted", DecryptedMessage: msg.Content,}

  render(c, "decryption.html", extraHeaders)
}
func send(c *gin.Context) {
  // Step 1: Validate form
  // Step 2: Send message in an email
  // Step 3: Redirect to confirmation page
  encryptionbytes, encryptionstring := GenerateRandomString(32)
  guid := xid.New()
  siteHost := GetViperVariable("host")
  msgEncrypted := &Message{
		Email:   string(MessageEncrypt([]byte(c.PostForm("email")), encryptionbytes)),
    FirstName: string(MessageEncrypt([]byte(c.PostForm("firstname")), encryptionbytes)),
    OtherFirstName: string(MessageEncrypt([]byte(c.PostForm("other_firstname")), encryptionbytes)),
    OtherLastName: string(MessageEncrypt([]byte(c.PostForm("other_lastname")), encryptionbytes)),
    OtherEmail: string(MessageEncrypt([]byte(c.PostForm("other_email")), encryptionbytes)),
    Content: string(MessageEncrypt([]byte(c.PostForm("content")), encryptionbytes)),
    Uniqueid: guid.String(),
  }

  msg := &MessagePost{
    Email: []string{c.PostForm("email")},
    FirstName: c.PostForm("firstname"),
    OtherFirstName: c.PostForm("other_firstname"),
    OtherLastName: c.PostForm("other_lastname"),
    OtherEmail: c.PostForm("other_email"),
    Content: c.PostForm("content"),
    Url: siteHost + "encrypt/" + msgEncrypted.Uniqueid + "/" + string(encryptionstring[:]),
  }


	if msg.Validate() == false {
    fmt.Println("unvalidated")
    fmt.Println("errors: %s", msg.Errors)
    htmlHeaders :=htmlHeaders{
      Title: "Password Exchange",
    }
		render(c, "home.html", htmlHeaders)
		return
	}

  msg.Content = "please click this link to get your encrypted message" +  "\n <a href=\"" + msg.Url + "\"> here</a>"
  Insert(msgEncrypted)

	if err := msg.Deliver(); err != nil {
		log.Println(err)
    c.String(http.StatusInternalServerError, fmt.Sprintf("something went wrong: %s", err))

		return
	}
	c.Redirect( http.StatusSeeOther, fmt.Sprintf("/confirmation?content=%s", msg.Url) )

}
  
func confirmation(c *gin.Context) {
  content := c.Query("content")
  extraHeaders :=htmlHeaders{Title: "passwordExchange", Url: content,}

  render(c, "confirmation.html", extraHeaders)
}
func render(c *gin.Context, filename string, data interface{}) {

    
   

      // Call the HTML method of the Context to render a template
      c.HTML(
        // Set the HTTP status to 200 (OK)
        //TODO: have this be settable
        http.StatusOK,
        // Use the index.html template
        filename,
        // Pass the data that the page uses (in this case, 'title')
        data,
      )
    
    
    
  }

//   if err := tmpl.Execute(w, data); err != nil {
//     log.Println(err)
//     http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
//   }
// }

