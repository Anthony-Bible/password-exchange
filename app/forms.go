// forms.go
package main

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	"encoding/json"
	"io/ioutil"

	"github.com/Anthony-Bible/password-exchange/app/email"
	"github.com/Anthony-Bible/password-exchange/app/message"

	"github.com/Anthony-Bible/password-exchange/app/commons"

	db "github.com/Anthony-Bible/password-exchange/app/database"
	pb "github.com/Anthony-Bible/password-exchange/app/encryptionpb"

	"google.golang.org/grpc"
)

type htmlHeaders struct {
	Title            string
	Url              string
	DecryptedMessage string
	Errors           map[string]string
}

// this type contains state of the server
type EncryptionClient struct {
	// client to GRPC service
	Client pb.MessageServiceClient
	conn   *grpc.ClientConn

	// default timeout
	// Timeout time.Duration

	// some other useful objects, like config
	// or logger (to replace global logging)
	// (...)
}

// constructor for server context
func newServerContext(endpoint string) (*EncryptionClient, error) {
	userConn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewMessageServiceClient(userConn)
	ctx := &EncryptionClient{
		Client: client,
		conn:   userConn,
	}
	return ctx, nil
}

func main() {

	s, err := newServerContext(os.Getenv("USER_SERVICE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong with contacting grpc server")
	}
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.GET("/", home)
	router.POST("/", s.send)
	router.GET("/confirmation", confirmation)
	router.GET("/decrypt/:uuid/*key", s.displaydecrypted)
	router.POST("/api/:app/*action", s.doAction)
	router.POST("/api/:app", s.doAction)

	router.NoRoute(failedtoFind)
	log.Info().Msg("Listening...")

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()

}

func home(c *gin.Context) {
	render(c, "home.html", 0, nil)
}
func failedtoFind(c *gin.Context) {
	render(c, "404.html", 404, nil)
}
func (s *EncryptionClient) displaydecrypted(c *gin.Context) {
	ctx := context.Background()

	uuid := c.Param("uuid")
	key := c.Param("key")
	decodedKey, err := b64.URLEncoding.DecodeString(key[1:])
	if err != nil {
		log.Fatal().Err(err).Msg("Something went wrong with b64 decoding")
	}
	selectResult := db.Select(uuid)
	bytesDecodedContent, err := b64.URLEncoding.DecodeString(selectResult.Content)
	if err != nil {
		log.Fatal().Err(err).Msg("Something went wrong with base64 decoding")
	}
	var decodedContent []string
	decodedContent = append(decodedContent, string(bytesDecodedContent))
	var arr [32]byte
	copy(arr[:], decodedKey)
	content, err := s.Client.DecryptMessage(ctx, &pb.DecryptedMessageRequest{Ciphertext: decodedContent, Key: decodedKey})
	if err != nil {
		log.Fatal().Err(err).Msg("Something went wrong with decryption")
	}
	msg := &message.MessagePost{
		Content: strings.Join((content.GetPlaintext()), ""),
	}
	extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted", DecryptedMessage: msg.Content}

	render(c, "decryption.html", 0, extraHeaders)
}
func (s *EncryptionClient) doAction(c *gin.Context) {
	c.MultipartForm()
	for key, value := range c.Request.PostForm {
		log.Info().Msgf("%v = %v \n", key, value)
	}

}
func (s *EncryptionClient) send(c *gin.Context) {
	// Step 1: Validate form
	// Step 2: Send message in an email
	// Step 3: Redirect to confirmation page
	ctx := context.Background()
	encryptionbytes, err := s.Client.GenerateRandomString(ctx, &pb.Randomrequest{RandomLength: 32})
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with generating random string")
	}
	guid := xid.New()
	siteHost, err := commons.GetViperVariable("host")
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with env variable")
	}
	//TODO: pass in struct & Handle two return values
	//TODO LATER: Find more effecient way to encrypt rather than contact encrypt everytime
	encryptionRequest := &pb.EncryptedMessageRequest{
		Key: encryptionbytes.GetEncryptionbytes(),
	}
	encryptionRequest.PlainText = append(encryptionRequest.PlainText, c.PostForm("content"))

	encryptedStrings, err := s.Client.EncryptMessage(ctx, encryptionRequest)
	encryptedStringSlice := encryptedStrings.GetCiphertext()
	// msgEncrypted.Uniqueid = guid.String()
	msg := &message.MessagePost{
		Email:          []string{c.PostForm("email")},
		FirstName:      c.PostForm("firstname"),
		OtherFirstName: c.PostForm("other_firstname"),
		OtherLastName:  c.PostForm("other_lastname"),
		OtherEmail:     []string{c.PostForm("other_email")},
		Url:            siteHost + "decrypt/" + guid.String() + "/" + strings.Join(encryptedStringSlice, ""),
		Hidden:         c.PostForm("other_information"),
		Captcha:        c.PostForm("h-captcha-response"),
	}

	if msg.Validate() == false {
		log.Debug().Msgf("errors: %s", msg.Errors)
		htmlHeaders := htmlHeaders{
			Title:  "Password Exchange",
			Errors: msg.Errors,
		}
		render(c, "home.html", 500, htmlHeaders)
		return
	}

	msg.Content = "please click this link to get your encrypted message" + "\n <a href=\"" + msg.Url + "\"> here</a>"
	db.Insert(&message.Message{Uniqueid: guid.String(), Content: strings.Join(encryptedStringSlice, "")})
	if checkBot(msg.Captcha) {
		// TODO Figure out how to use a fucntion from another package on a struct on another package
		if err := email.Deliver(msg); err != nil {
			log.Error().Err(err).Msg("")
			c.String(http.StatusInternalServerError, fmt.Sprintf("something went wrong: %s", err))

			return
		}
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/confirmation?content=%s", msg.Url))

}

func confirmation(c *gin.Context) {
	content := c.Query("content")
	extraHeaders := htmlHeaders{Title: "passwordExchange", Url: content}

	render(c, "confirmation.html", 0, extraHeaders)
}
func render(c *gin.Context, filename string, status int, data interface{}) {

	if status == 0 {
		status = 200
	}

	// Call the HTML method of the Context to render a template
	c.HTML(
		// Set the HTTP status to 200 (OK)
		//TODO: have this be settable
		status,
		// Use the index.html template
		filename,
		// Pass the data that the page uses (in this case, 'title')
		data,
	)

}

type test_struct struct {
	Success      bool   `json:"success"`
	Challenge_ts string `json:"challenge_ts"`
	Hostname     string `json:"hostname"`
}

func checkBot(hcaptchaResponse string) (returnstatus bool) {
	secret, err := commons.GetViperVariable("hcaptcha_secret")
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with env variable")
	}
	sitekey, err := commons.GetViperVariable("hcaptcha_sitekey")
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with env variable")
	}
	u := make(url.Values)
	u.Set("secret", secret)
	u.Set("response", hcaptchaResponse)
	u.Set("sitekey", sitekey)
	response, err := http.PostForm("https://hcaptcha.com/siteverify", u)

	if err != nil {
		log.Error().
			Str("error", err.Error()).
			Msg("Something went wrong with hcaptcha")
		return false
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal().Err(err).Msg("Problem with env variable")

	}
	var t test_struct
	err = json.Unmarshal(body, &t)
	if err != nil {

		log.Error().
			Msg("Can't Unmarshal json")
		return false
	}
	return t.Success
}

//   if err := tmpl.Execute(w, data); err != nil {
//     log.Println(err)
//     http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
//   }
// }
