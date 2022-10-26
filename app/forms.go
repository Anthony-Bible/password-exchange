package main

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	// "net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	"encoding/json"
	// "io/ioutil"

	"github.com/Anthony-Bible/password-exchange/app/commons"
	"github.com/Anthony-Bible/password-exchange/app/email"
	"github.com/Anthony-Bible/password-exchange/app/message"
	"github.com/Anthony-Bible/password-exchange/app/rabbitmq"

	db "github.com/Anthony-Bible/password-exchange/app/databasepb"
	pb "github.com/Anthony-Bible/password-exchange/app/encryptionpb"

	"google.golang.org/grpc"
)

// TODO add a size limit for messages
type htmlHeaders struct {
	Title            string
	Url              string
	DecryptedMessage string
	Errors           map[string]string
}

// this type contains state of the server
type EncryptionClient struct {
	// client to GRPC service
	Client   pb.MessageServiceClient
	DbClient db.DbServiceClient
	conn     *grpc.ClientConn
	dbconn   *grpc.ClientConn

	// default timeout
	// Timeout time.Duration

	// some other useful objects, like config
	// or logger (to replace global logging)
	// (...)
}
type Result struct {
	Email string
	Error error
}

// constructor for server context
func newServerContext(endpoint1 string, endpoint2 string) (*EncryptionClient, error) {
	userConn, err := grpc.Dial(endpoint1, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	userConn2, err := grpc.Dial(endpoint2, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	dbclient := db.NewDbServiceClient(userConn2)

	client := pb.NewMessageServiceClient(userConn)
	ctx := &EncryptionClient{
		Client:   client,
		DbClient: dbclient,
		conn:     userConn,
		dbconn:   userConn2,
	}
	fmt.Println("in function", ctx)
	return ctx, nil
}

func main() {
	//TODO put port in environment variable
	encryptionServiceName, dbServiceName := getServiceNames()
	s, err := newServerContext(encryptionServiceName, dbServiceName)
	if err != nil {
		log.Error().Err(err).Msg("something went wrong 	with contacting encryption grpc server")
	}

	router := gin.Default()
	router.LoadHTMLGlob("/templates/*.html")
	router.Static("/assets", "/templates/assets")
	router.GET("/", home)
	router.POST("/", s.send)
	router.GET("/confirmation", confirmation)
	router.GET("/decrypt/:uuid/*key", s.displaydecrypted)
	router.POST("/decrypt/:uuid/*key", s.displaydecryptedWithPassword)
	router.GET("/about", about)

	router.NoRoute(failedtoFind)
	log.Info().Msg("Listening...")

	// By default it serves on :8080 unless a
	// PORT environment variable was defined.
	router.Run()

}

func getServiceNames() (string, string) {
	environment := getEnvironment()
	encryptionServiceName, err := commons.GetViperVariable("encryption_" + environment + "_service")
	dbServiceName, err := commons.GetViperVariable("database_" + environment + "_service")
	log.Debug().Msg(dbServiceName)

	encryptionServiceName += ":50051"
	log.Debug().Msg(encryptionServiceName)

	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong with getting the encryption-service address")
	}
	return encryptionServiceName, dbServiceName
}

func getEnvironment() string {
	environment, err := commons.GetViperVariable("running_environment")
	if err != nil {
		log.Error().Err(err).Msg("couldn't get running_environment")
	}
	return environment
}

func home(c *gin.Context) {
	render(c, "home.html", 0, nil)
}
func about(c *gin.Context) {
	render(c, "about.html", 0, nil)
}
func failedtoFind(c *gin.Context) {
	render(c, "404.html", 404, nil)
}
func (s *EncryptionClient) displaydecrypted(c *gin.Context) {

	extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted"}

	render(c, "decryption.html", 0, extraHeaders)
}

func (s *EncryptionClient) displaydecryptedWithPassword(c *gin.Context) {
	ctx := context.Background()

	uuid := c.Param("uuid")
	key := c.Param("key")
	decodedKey := decodeString(key)
	inputtedPassphrase := c.PostForm("passphrase")
	selectResult, err := s.DbClient.Select(ctx, &db.SelectRequest{Uuid: uuid})
	hashedPassword := selectResult.GetPassphrase()
	if checkPassword([]byte(hashedPassword), []byte(inputtedPassphrase)) {

		// bytesDecodedContent, err := b64.URLEncoding.DecodeString(selectResult.Content)
		if err != nil {
			log.Error().Err(err).Msg("Something went wrong with select from db")
		}
		if len(selectResult.GetContent()) == 0 {
			render(c, "404.html", 404, nil)

			return
		}
		var decodedContent []string
		decodedContent = append(decodedContent, string(selectResult.GetContent()))
		var arr [32]byte
		copy(arr[:], decodedKey)
		content := s.decryptMessage(ctx, decodedContent, decodedKey, selectResult)
		msg := &message.MessagePost{
			Content: strings.Join((content.GetPlaintext()), ""),
		}
		decryptedContent, _ := b64.URLEncoding.DecodeString(msg.Content)
		decryptedContentString := string(decryptedContent)
		extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted", DecryptedMessage: decryptedContentString}

		render(c, "decryption.html", 0, extraHeaders)
	} else {
		extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted", DecryptedMessage: "Wrong Passphrase/Lastname. Please try again(can be empty)"}

		render(c, "decryption.html", 0, extraHeaders)
	}
}

func (s *EncryptionClient) decryptMessage(ctx context.Context, decodedContent []string, decodedKey []byte, selectResult *db.SelectResponse) *pb.DecryptedMessageResponse {
	content, err := s.Client.DecryptMessage(ctx, &pb.DecryptedMessageRequest{Ciphertext: decodedContent, Key: decodedKey})
	if err != nil {
		marshaledSelect, _ := json.Marshal(selectResult)
		marshaledStruct, _ := json.Marshal(&pb.DecryptedMessageRequest{Ciphertext: decodedContent, Key: decodedKey})
		log.Debug().Msg(string(marshaledStruct))
		log.Debug().Msg(string(marshaledSelect))

		log.Error().Err(err).Msg("Something went wrong with decryption")
	}
	return content
}

func decodeString(key string) []byte {
	decodedKey, err := b64.URLEncoding.DecodeString(key[1:])
	if err != nil {
		log.Error().Err(err).Msgf("Something went wrong with b64 decoding: %s Key: %s", decodedKey, key)
	}
	return decodedKey
}
func sendEmailtoQueue(ch chan message.MessagePost, c *gin.Context, done <-chan interface{}) <-chan Result {
	results := make(chan Result)
	go func() {
		defer close(results)
		if strings.ToLower(c.PostForm("color")) == "blue" {
			if len(c.PostForm("skipEmail")) <= 0 {
				rabbitmq_address, err := commons.GetViperVariable("rabbitmq_address")
				if err != nil {
					log.Fatal().Err(err).Msg("Rabbitmq address is not defined")
				}
				client := rabbitmq.NewRab("email", "email", rabbitmq_address, done)
				isokay := verifyEmail(<-ch, c)
				if isokay {
					log.Error().Msg("email is malformed")
				}
				err := client.Push([]byte(email))
			}
		}
	}()
}

//func sendEmail(c *gin.Context, msg *message.MessagePost) {
//		}
//	}
//}

func deliverEmail(msg *message.MessagePost, c *gin.Context) bool {
	if err := email.Deliver(msg); err != nil {
		marshaledMesage, _ := json.Marshal(msg)
		log.Error().Err(err).Msg(string(marshaledMesage))
		c.String(http.StatusInternalServerError, fmt.Sprintf("something went wrong on email delivery: %s", err))

		return true
	}
	return false
}

func verifyEmail(msg message.MessagePost, c *gin.Context) bool {
	if msg.Validate() == false {
		log.Debug().Msgf("errors: %s", msg.Errors)
		htmlHeaders := htmlHeaders{
			Title:  "Password Exchange",
			Errors: msg.Errors,
		}
		render(c, "home.html", 500, htmlHeaders)
		return true
	}
	return false
}

func (s *EncryptionClient) send(c *gin.Context, done <-chan interface{}) {
	// Step 1: Validate form
	// Step 2: Send message in an email
	// // Step 3: Redirect to confirmation page
	// FOR DEBUGGING HTTP POST:
	// printPost(c)
	msgStream := make(chan message.MessagePost)
	go sendEmailtoQueue(msgStream, c, done)
	encryptionbytes, err := s.Client.GenerateRandomString(context.Background(), &pb.Randomrequest{RandomLength: 32})
	if err != nil {
		log.Error().Err(err).Msg("Problem with generating random string")
	}
	guid := xid.New()
	environment := getEnvironment()
	siteHost, err := commons.GetViperVariable(environment + "_host")
	if err != nil {
		log.Error().Err(err).Msg("Problem with env variable")
	}
	//TODO: pass in struct & Handle two return values
	//TODO LATER: Find more effecient way to encrypt rather than contact encrypt everytime
	encryptionRequest := &pb.EncryptedMessageRequest{
		Key: []byte(encryptionbytes.GetEncryptionbytes()),
	}
	encryptionRequest.PlainText = append(encryptionRequest.PlainText, c.PostForm("content"))

	encryptedStrings, err := s.Client.EncryptMessage(ctx, encryptionRequest)
	encryptedStringSlice := encryptedStrings.GetCiphertext()
	msg := createMessageFromPost(c, siteHost, guid, encryptionRequest)
	msg.OtherLastName = string(hashPassphrase([]byte(msg.OtherLastName)))
	_, err = s.DbClient.Insert(ctx, &db.InsertRequest{Uuid: guid.String(), Content: strings.Join(encryptedStringSlice, ""), Passphrase: msg.OtherLastName})
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with insert")
	}
	msgStream <- msg
	// TODO Figure out how to use a fucntion from another package on a struct on another package
	if len(c.PostForm("api")) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"url": msg.Url,
		})
	} else {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/confirmation?content=%s", msg.Url))
	}
}
func hashPassphrase(passphrase []byte) []byte {
	hashed, err := bcrypt.GenerateFromPassword(passphrase, 14)
	if err != nil {
		log.Error().Err(err).Msg("something went wrong with hashing passphrase")
	}
	return hashed
}
func checkPassword(hashedPassword []byte, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}
func printPost(c *gin.Context) {
	//used for debugging
	c.MultipartForm()
	for key, value := range c.Request.PostForm {
		log.Info().Msgf("%v = %v \n", key, value)
	}
}

func createMessageFromPost(c *gin.Context, siteHost string, guid xid.ID, encryptionRequest *pb.EncryptedMessageRequest) message.MessagePost {
	msg := message.MessagePost{
		Email:          []string{c.PostForm("email")},
		FirstName:      c.PostForm("firstname"),
		OtherFirstName: c.PostForm("other_firstname"),
		OtherLastName:  c.PostForm("other_lastname"),
		OtherEmail:     []string{c.PostForm("other_email")},
		Uniqueid:       "",
		Content:        "",
		Errors:         map[string]string{},
		Url:            siteHost + "decrypt/" + guid.String() + "/" + string(b64.URLEncoding.EncodeToString(encryptionRequest.Key)),
		Hidden:         c.PostForm("other_information"),
		Captcha:        c.PostForm("h-captcha-response"),
	}
	msg.Content = "please click this link to get your encrypted message" + "\n <a href=\"" + msg.Url + "\"> here</a>"
	return msg
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
