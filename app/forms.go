// forms.go
package main

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	"github.com/Anthony-Bible/password-exchange/app/commons"
	"github.com/Anthony-Bible/password-exchange/app/message"

	"encoding/json"
	"io/ioutil"
)

type htmlHeaders struct {
	Title            string
	Url              string
	DecryptedMessage string
	Errors           map[string]string
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.GET("/", home)
	router.POST("/", send)
	router.GET("/confirmation", confirmation)
	router.GET("/decrypt/:uuid/*key", displaydecrypted)
	router.POST("/api/:app/*action", doAction)
	router.POST("/api/:app", doAction)

	router.NoRoute(failedtoFind)
	log.Info().Msg("Listening...")

	// Build a Slack App Home in Golang Using Socket Mode

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
	content, err := MessageDecrypt([]byte(decodedContent), &arr)
	if err != nil {
		panic(err)
	}
	msg := &Message.MessagePost{
		Content: string(content),
	}
	extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted", DecryptedMessage: msg.Content}

	render(c, "decryption.html", 0, extraHeaders)
}
func doAction(c *gin.Context) {
	c.MultipartForm()
	for key, value := range c.Request.PostForm {
		log.Printf("%v = %v \n", key, value)
	}

}
func send(c *gin.Context) {
	// Step 1: Validate form
	// Step 2: Send message in an email
	// Step 3: Redirect to confirmation page
	encryptionbytes, encryptionstring := GenerateRandomString(32)
	guid := xid.New()
	siteHost, err := commons.GetViperVariable("host")
	if err != nil {
		panic(err)
	}
	msgEncrypted := &message.Message{
		Email:          string(MessageEncrypt([]byte(c.PostForm("email")), encryptionbytes)),
		FirstName:      string(MessageEncrypt([]byte(c.PostForm("firstname")), encryptionbytes)),
		OtherFirstName: string(MessageEncrypt([]byte(c.PostForm("other_firstname")), encryptionbytes)),
		OtherLastName:  string(MessageEncrypt([]byte(c.PostForm("other_lastname")), encryptionbytes)),
		OtherEmail:     string(MessageEncrypt([]byte(c.PostForm("other_email")), encryptionbytes)),
		Content:        string(MessageEncrypt([]byte(c.PostForm("content")), encryptionbytes)),
		Uniqueid:       guid.String(),
	}

	msg := &message.MessagePost{
		Email:          []string{c.PostForm("email")},
		FirstName:      c.PostForm("firstname"),
		OtherFirstName: c.PostForm("other_firstname"),
		OtherLastName:  c.PostForm("other_lastname"),
		OtherEmail:     []string{c.PostForm("other_email")},
		Content:        c.PostForm("content"),
		Url:            siteHost + "decrypt/" + msgEncrypted.Uniqueid + "/" + string(encryptionstring[:]),
		hidden:         c.PostForm("other_information"),
		captcha:        c.PostForm("h-captcha-response"),
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
	Insert(msgEncrypted)
	if checkBot(msg.captcha) {
		if err := msg.Deliver(); err != nil {
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
	secret, err := GetViperVariable("hcaptcha_secret")
	if err != nil {
		panic(err)
	}
	sitekey, err := GetViperVariable("hcaptcha_sitekey")
	if err != nil {
		panic(err)
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
		panic(err)

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
