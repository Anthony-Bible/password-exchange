package main

import (
    "log"
    "github.com/gin-gonic/gin"
    b64 "encoding/base64"
    "net/http"
    "fmt"
    "github.com/rs/xid"
    "github.com/Anthony-Bible/password-exchange/app/encryption"
    "github.com/Anthony-Bible/password-exchange/app/message"
    "github.com/Anthony-Bible/password-exchange/app/commons"
	"github.com/Anthony-Bible/password-exchange/app/encryptionpb"
	"google.golang.org/grpc"
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
//   router.GET("/", home)
  router.POST("/encrypt", encrypt)
//   router.GET("/confirmation", confirmation)
  router.GET("/decrypt/:uuid/*key", displaydecrypted)
  router.NoRoute(failedtoFind)
  log.Println("Listening...")

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
  router.Run()

}




func encrypt(c *gin.Context) {
	// Step 1: Validate form
	// Step 2: Send message in an email
	// Step 3: Redirect to confirmation page
	encryptionbytes, encryptionstring := encryption.GenerateRandomString(32)
	guid := xid.New()
	siteHost := commons.GetViperVariable("host")
	cc, err := grpc.Dial("encryption-svc:50051", opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Close()

	client := encryptionpb.NewHelloServiceClient(cc)
	request := &encryptionpb.PlainMessage{	  
	Email: []string{c.PostForm("email")},
	FirstName: c.PostForm("firstname"),
	OtherFirstName: c.PostForm("other_firstname"),
	OtherLastName: c.PostForm("other_lastname"),
	OtherEmail: c.PostForm("other_email"),
	Content: c.PostForm("content"),
	Url: siteHost + "encrypt/" + msgEncrypted.Uniqueid + "/" + string(encryptionstring[:]),
  }

	resp, _ := client.Hello(context.Background(), request)
	fmt.Printf("Receive response => [%v]", resp.Greeting)
	msg := &MessagePost{
	  Email: []string{c.PostForm("email")},
	  FirstName: c.PostForm("firstname"),
	  OtherFirstName: c.PostForm("other_firstname"),
	  OtherLastName: c.PostForm("other_lastname"),
	  OtherEmail: c.PostForm("other_email"),
	  Content: c.PostForm("content"),
	  Url: siteHost + "encrypt/" + msgEncrypted.Uniqueid + "/" + string(encryptionstring[:]),
	}
  
}
