package web

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"time"

	"golang.org/x/crypto/bcrypt"

	// "net/url"
	"strings"

	"github.com/Anthony-Bible/password-exchange/app/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"

	"encoding/json"
	// "io/ioutil"

	"github.com/Anthony-Bible/password-exchange/app/commons"
	"github.com/Anthony-Bible/password-exchange/app/message"

	db "github.com/Anthony-Bible/password-exchange/app/databasepb"
	pb "github.com/Anthony-Bible/password-exchange/app/encryptionpb"
	"github.com/Anthony-Bible/password-exchange/app/messagepb"
	amqp "github.com/rabbitmq/amqp091-go"

	servertiming "github.com/p768lwy3/gin-server-timing"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
	Channel           *amqp.Channel
	EncryptionClient
}

// TODO add a size limit for messages
type htmlHeaders struct {
	Title            string
	Url              string
	DecryptedMessage string
	Errors           map[string]string
	HasPassword      bool
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

func (conf *Config) GetConn(rabbitUrl string) error {
	conn, err := amqp.Dial(rabbitUrl)
	if err != nil {
		log.Err(err).Msg("Problem with connecting")
	}
	ch, err := conn.Channel()
	conf.Channel = ch
	return err
}

// constructor for server context
func newServerContext(endpoint1 string, endpoint2 string) (EncryptionClient, error) {
	userConn, err := grpc.Dial(endpoint1, grpc.WithInsecure())
	if err != nil {
		return EncryptionClient{}, err
	}
	userConn2, err := grpc.Dial(endpoint2, grpc.WithInsecure())
	if err != nil {
		return EncryptionClient{}, err
	}
	dbclient := db.NewDbServiceClient(userConn2)

	client := pb.NewMessageServiceClient(userConn)
	ctx := EncryptionClient{
		Client:   client,
		DbClient: dbclient,
		conn:     userConn,
		dbconn:   userConn2,
	}
	fmt.Println("in function", ctx)
	return ctx, nil
}

func (conf Config) StartServer() {
	//TODO put port in environment variable
	encryptionServiceName, dbServiceName := conf.getServiceNames()
	s, err := newServerContext(encryptionServiceName, dbServiceName)
	conf.EncryptionClient = s
	if err != nil {
		log.Error().Err(err).Msg("something went wrong 	with contacting encryption grpc server")
	}

	router := gin.Default()
	router.Use(servertiming.Middleware())
	router.LoadHTMLGlob("/templates/*.html")
	router.Static("/assets", "/templates/assets")
	router.GET("/", home)
	router.POST("/", conf.send)
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

func (conf Config) getServiceNames() (string, string) {
	encryptionServiceName, err := commons.GetViperVariable(fmt.Sprintf("Encryption%sService", conf.RunningEnvironment))
	dbServiceName, err := commons.GetViperVariable(fmt.Sprintf("Database%sService", conf.RunningEnvironment))
	log.Debug().Msg(dbServiceName)

	encryptionServiceName += ":50051"
	log.Debug().Msg(encryptionServiceName)

	if err != nil {
		log.Fatal().Err(err).Msg("something went wrong with getting the encryption-service address")
	}
	return encryptionServiceName, dbServiceName
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
	uuid := c.Param("uuid")
	hashedPassword, _ := s.getHashedPassphrase(context.Background(), uuid)
	hasPassword := true
	if hashedPassword == "" {
		hasPassword = false
		s.displaydecryptedWithPassword(c)
	} else {
		extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted", HasPassword: hasPassword}

		render(c, "decryption.html", 0, extraHeaders)
	}
}
func (s *EncryptionClient) getHashedPassphrase(ctx context.Context, uuid string) (string, *db.SelectResponse) {

	selectResult, err := s.DbClient.Select(ctx, &db.SelectRequest{Uuid: uuid})
	hashedPassword := selectResult.GetPassphrase()

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with select from db")
	}
	return hashedPassword, selectResult
}
func (s *EncryptionClient) displaydecryptedWithPassword(c *gin.Context) {
	printPost(c)
	ctx := context.Background()
	uuid := c.Param("uuid")
	key := c.Param("key")
	hashedPassword, selectResult := s.getHashedPassphrase(ctx, uuid)
	decodedKey := decodeString(key)
	inputtedPassphrase := c.PostForm("passphrase")
	// print warning if password is empty
	// print hashedPassword
	if checkPassword([]byte(hashedPassword), []byte(inputtedPassphrase)) || hashedPassword == "" {

		// bytesDecodedContent, err := b64.URLEncoding.DecodeString(selectResult.Content)
		if len(selectResult.GetContent()) == 0 {
			render(c, "404.html", 404, nil)

    if checkPassword([]byte(hashedPassword), []byte(inputtedPassphrase)) {
        displayDecryptedContent(c, selectResult, decodedKey)
    } else {
        extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted", DecryptedMessage: "Wrong Passphrase/Lastname. Please try again(can be empty)"}
        render(c, "decryption.html", 0, extraHeaders)
    }
}

func displayDecryptedContent(c *gin.Context, selectResult *db.SelectResponse, decodedKey []byte) {
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
func (conf Config) sendEmailtoQueue(ch chan message.MessagePost, c *gin.Context) {
	rabUrl := fmt.Sprintf("amqp://%s:%s@%s", conf.RabUser, conf.RabPass, conf.RabHost)
	go func() {
		if strings.ToLower(c.PostForm("color")) == "blue" {
			if len(c.PostForm("skipEmail")) <= 0 {
				err := conf.GetConn(rabUrl)
				if err != nil {
					log.Fatal().Err(err)
				}
				msg := <-ch
				isokay := verifyEmail(msg, c)
				log.Debug().Msg("verified email")
				if isokay {
					log.Error().Msg("email is malformed")
				} else {

					log.Debug().Msg("start 1 push email")
					conf.publishToQueue(msg)
					log.Debug().Msg("finished push")
				}
			}
		} else {
			log.Debug().Msg("no color")
			log.Debug().Msgf("%+v", <-ch)
		}
	}()
}

func (conf Config) publishToQueue(msg message.MessagePost) {
	log.Info().Msg("Starting push")
	q, err := conf.Channel.QueueDeclare(
		conf.RabQName, //name
		true,          //durable
		false,         //delete when unused
		false,         //exclusive
		false,         //no-wait
		nil,           //arguments
	)
	if err != nil {
		log.Fatal().Msgf("Couldn't declare queue: %s\n", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	body := messagepb.Message{
		Email:          strings.Join(msg.Email, ""),
		Firstname:      msg.FirstName,
		Otherfirstname: msg.OtherFirstName,
		OtherLastName:  msg.OtherLastName,
		OtherEmail:     strings.Join(msg.OtherEmail, ""),
		Uniqueid:       msg.Uniqueid,
		Content:        msg.Content,
		Url:            msg.Url,
		Hidden:         msg.Hidden,
		Captcha:        msg.Captcha,
	}
	data, err := proto.Marshal(&body)
	if err != nil {
		log.Error().Msg("Error with marshaling body")
	}
	log.Info().Msg("before publish")
	err = conf.Channel.PublishWithContext(ctx,
		"",     //exchange
		q.Name, //routing key
		false,  //mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(data),
		})
}

//func sendEmail(c *gin.Context, msg *message.MessagePost) {
//		}
//	}
//}

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

func (conf Config) send(c *gin.Context) {
	// Step 1: Validate form
	// Step 2: Send message in an email
	// // Step 3: Redirect to confirmation page
	// FOR DEBUGGING HTTP POST:
	// printPost(c)
	log.Info().Msg("sending")
	ctx := context.Background()
	timing := servertiming.FromContext(c)
	var msgStream chan message.MessagePost
	msgStream = make(chan message.MessagePost)
	go conf.sendEmailtoQueue(msgStream, c)
	encryptionbytes, err := conf.EncryptionClient.Client.GenerateRandomString(context.Background(), &pb.Randomrequest{RandomLength: 32})
	if err != nil {
		log.Error().Err(err).Msg("Problem with generating random string")
	}
	log.Info().Msg(string(encryptionbytes.GetEncryptionbytes()))
	guid := xid.New()
	environment := conf.RunningEnvironment
	siteHost, err := commons.GetViperVariable(environment + "Host")
	if err != nil {
		log.Error().Err(err).Msg("Problem with env variable")
	}
	//TODO: pass in struct & Handle two return values
	//TODO LATER: Find more effecient way to encrypt rather than contact encrypt everytime,
	encryptionRequest := &pb.EncryptedMessageRequest{
		Key: []byte(encryptionbytes.GetEncryptionbytes()),
	}
	encryptionRequest.PlainText = append(encryptionRequest.PlainText, c.PostForm("content"))
	m := timing.NewMetric("grpc-encrypt").Start()
	encryptedStrings, err := conf.EncryptionClient.Client.EncryptMessage(ctx, encryptionRequest)
	m.Stop()
	encryptedStringSlice := encryptedStrings.GetCiphertext()
	metric2 := timing.NewMetric("messageFromPost").Start()
	msg := createMessageFromPost(c, siteHost, guid, encryptionRequest)
	metric2.Stop()

	metric3 := timing.NewMetric("hashpassphrase").Start()
	// only hash if it's not empty
	if len(msg.OtherLastName) > 0 {

		msg.OtherLastName = string(hashPassphrase([]byte(msg.OtherLastName)))
	}
	metric3.Stop()
	metric4 := timing.NewMetric("insert").Start()
	_, err = conf.DbClient.Insert(ctx, &db.InsertRequest{Uuid: guid.String(), Content: strings.Join(encryptedStringSlice, ""), Passphrase: msg.OtherLastName})
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with insert")
	}
	metric4.Stop()
	log.Info().Msg("before stream")
	msgStream <- msg
	log.Info().Msg("after stream")
	servertiming.WriteHeader(c)
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
	hashed, err := bcrypt.GenerateFromPassword(passphrase, 11)
	if err != nil {
		log.Error().Err(err).Msg("something went wrong with hashing passphrase")
	}
	return hashed
}
func checkPassword(hashedPassword []byte, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	if strings.TrimSpace(string(hashedPassword)) == "" {
		log.Debug().Msg("password is empty")
		return true
	}
	log.Debug().Err(err).Msg("error is")
	log.Debug().Msgf("error==nil: %t", err == nil)
	return err == nil

}
func printPost(c *gin.Context) {
	//used for debugging
	//	c.MultipartForm()
	//	for key, value := range c.Request.PostForm {
	//		log.Info().Msgf("%v = %v \n", key, value)
	//	}
	body, _ := ioutil.ReadAll(c.Request.Body)
	println(string(body))

	c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
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
