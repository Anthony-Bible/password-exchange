//Package web This package starts the web server as the primary interface for interaction
package web

import (
	"bufio"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	servertiming "github.com/p768lwy3/gin-server-timing"
	amqp "github.com/rabbitmq/amqp091-go"
	gcache "github.com/vimeo/galaxycache"
	ghttp "github.com/vimeo/galaxycache/http"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
	Channel           *amqp.Channel
	S3Session         *session.Session
	EncryptionClient
	Galaxy *gcache.Galaxy
}

// TODO add a size limit for messages
type htmlHeaders struct {
	Title            string
	URL              string
	DecryptedMessage string
	Errors           map[string]string
}

//EncryptionClient this type contains state of the server
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

//type Result struct {
//	Email string
//	Error error
//}

//GetConn adds rabitmq connection to config struct
func (conf *Config) GetConn(rabbitURL string) error {
	conn, err := amqp.Dial(rabbitURL)
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

// StartServer starts the web server using gin-gonic
func (conf Config) StartServer() {
	//TODO put port in environment variable
	encryptionServiceName, dbServiceName := conf.getServiceNames()
	s, err := newServerContext(encryptionServiceName, dbServiceName)
	conf.EncryptionClient = s
	if err != nil {
		log.Error().Err(err).Msg("something went wrong 	with contacting encryption grpc server")
	}

	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20 // 8 MiB
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

	extraHeaders := htmlHeaders{Title: "passwordExchange Decrypted"}

	render(c, "decryption.html", 0, extraHeaders)
}

func (s *EncryptionClient) displaydecryptedWithPassword(c *gin.Context) {
	printPost(c)
	ctx := context.Background()
	uuid := c.Param("uuid")
	key := c.Param("key")
	decodedKey := decodeString(key)
	inputtedPassphrase := c.PostForm("passphrase")
	selectResult, err := s.DbClient.Select(ctx, &db.SelectRequest{Uuid: uuid})
	hashedPassword := selectResult.GetPassphrase()
	if checkPassword([]byte(hashedPassword), []byte(inputtedPassphrase)) || hashedPassword == "" {

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
func (conf Config) sendEmailtoQueue(ch chan message.MessagePost, c *gin.Context) {
	rabURL := fmt.Sprintf("amqp://%s:%s@%s", conf.RabUser, conf.RabPass, conf.RabHost)
	err := conf.GetConn(rabURL)
	if err != nil {
		log.Fatal().Err(err)
	}
	go func() {
		if strings.ToLower(c.PostForm("color")) == "blue" {
			if len(c.PostForm("skipEmail")) <= 0 {
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
		URL:            msg.URL,
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
	// pri(conf ntPost(c)
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
	guid := generateUniqueID()
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
	msg.OtherLastName = string(hashPassphrase([]byte(msg.OtherLastName)))
	metric3.Stop()
	metric4 := timing.NewMetric("insert").Start()
	_, err = conf.DbClient.Insert(ctx, &db.InsertRequest{Uuid: guid, Content: strings.Join(encryptedStringSlice, ""), Passphrase: msg.OtherLastName})
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with insert")
	}
	metric4.Stop()
	log.Info().Msg("before stream")
	msgStream <- msg
	log.Info().Msg("after stream")
	servertiming.WriteHeader(c)
	// TODO Figure out how to use a fucntion from another package on a struct on another package
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("something went wrong with getting file from request")
	}
	conf.createS3Connection()
	defer file.Close()
	var fileid string
	fileid = getFileID(c)
	totalChunks, err := strconv.Atoi(c.Request.FormValue("totalChunks"))
	if err != nil {
		// Handle error
		// ...
		log.Error().Err(err).Msg("Couldn't save totalchunks")
	}
	currentChunk, err := strconv.Atoi(c.Request.FormValue("currentChunk"))
	if err != nil {
		// Handle error
		// ...
		log.Error().Err(err).Msg("Couldn't save currentChunk")
	}
	encryptedFilePath, err := filepath.Abs("/tmp/" + fileid + ".enc")
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with getting absolute path")

	}
	encryptedChunk, err := encryptChunk(file, encryptionRequest.Key)
	//todo: Perhaps we want to move this to the encryption service
	// this is the final chunk time to create the final encrypted files
	encryptedFile, err := os.OpenFile(encryptedFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with opening encryptefile")
	}
	defer encryptedFile.Close()
	encryptedWriter := bufio.NewWriter(encryptedFile)

	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with encrypting chunk")

	}
	// read the encrypted file chunk one block at a tijme and write it to the final encrytped file
	block := make([]byte, aes.BlockSize)
	for {
		n, err := encryptedChunk.Read(block)
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("Something went wrong with reading encrypted block")
		}
		if n == 0 {
			break

		}
		_, err = encryptedWriter.Write(block[:n])
		if err != nil {
			log.Error().Err(err).Msg("Something went wrong with writing encyrpted block to file")

		}
	}
	//Flush the buffered writer to ensure the data is writen to file
	encryptedWriter.Flush()
	c.JSON(http.StatusOK, gin.H{"fileID": fileid})
	//check if this is the final chunk
	if currentChunk == totalChunks-1 {
		encryptedFile.Close()
	}
	if len(c.PostForm("api")) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"url": msg.URL,
		})
	} else {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/confirmation?content=%s", msg.URL))
	}
}

//get file id
func getFileID(c *gin.Context) string {
	var fileid string
	if len(c.Request.FormValue("fileID")) > 0 {
		fileIDLength := len(c.Request.FormValue("fileID"))
		fileid = c.Request.FormValue("fileID")
		log.Debug().Msgf("Length of fileid is %d", fileIDLength)
	} else {
		fileid = generateUniqueID()
	}
	return fileid
}

//Initialize s3 connection
func (conf *Config) createS3Connection() {
	// Create a single AWS session (we can re use this if we're uploading many files)
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(conf.S3apiID, conf.S3apiKey, ""),
		Endpoint:         aws.String(conf.S3apiEndpoint),
		Region:           aws.String(conf.S3apiRegion),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		log.Error().Err(err).Msg("Something went wrong with creating session")
	}
	conf.S3Session = sess
}

//query kubernetes headless service for endpoints
// Uses multiple A records to  get the endpoints
func getEndpoints(svcName string) []string {
	//query kubernetes headless service for endpoints
	endpoints, err := net.LookupIP(svcName)
	if err != nil {
		log.Error().Err(err).Msgf("Something went wrong with looking up %s service", svcName)
	}

	var endpointsString []string
	for _, endpoint := range endpoints {
		endpointsString = append(endpointsString, endpoint.String())
	}
	return endpointsString

}

//Initialize galaxy cache
func (conf *Config) initGalaxyCache() {
	endpoints := getEndpoints("password-exchange")
	modifiedEndpoints := modifyStringSlice(endpoints)
	httpProto := ghttp.NewHTTPFetchProtocol(nil)
	universe := gcache.NewUniverse(httpProto, endpoints[0])
	//set peers of universe
	universe.Set(modifiedEndpoints)
	getter := gcache.GetterFunc(func(ctx context.Context, key string, dest gcache.Codec) error {
		uploadID := initiateS3MultipartUpload(key)
		return dest.UnmarshalBinary([]byte(uploadID))
	})
	//Create a new galaxy
	galaxy := universe.NewGalaxy("password-exchange", 1<<20, getter)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveMux := http.NewServeMux()
	ghttp.RegisterHTTPHandler(universe, nil, serveMux)
	//store galaxy in config struct so I can just call conf.g.get()
	conf.Galaxy = galaxy
	var srv http.Server
	go func() {
		log.Info().Msg("Starting HTTP server on :8080")
		httpAltListener, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
		srv.Handler = serveMux
		if err := srv.Serve(httpAltListener); err != nil {
			log.Error().Err(err).Msg("Something went wrong with starting http server")
		}
	}()
	<-ctx.Done()
	srv.Shutdown(ctx)

}

//Initialize s3 mulitpart uplaod
func initiateS3MultipartUpload(key string) string {

}

//Modify string slice to modify each string
func modifyStringSlice(slice []string) []string {
	for i, s := range slice {
		slice[i] = "http://" + s + ":8080"
	}
	return slice
}

func encryptChunk(file io.Reader, key []byte) (io.Reader, error) {
	log.Debug().Msg("encrypting chunk")
	buffer := bytes.NewBuffer(nil)
	nonce := make([]byte, aes.BlockSize)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, err

	}
	//read the file chunk one block at a time
	block := make([]byte, aes.BlockSize)
	for {
		n, err := file.Read(block)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		newAesBlock, err := aes.NewCipher(key)
		if err != nil {
			log.Error().Err(err).Msg("Something went wrong with creating a new cipher")

		}
		encryptedBlock := make([]byte, aes.BlockSize)
		// Encrypt the block using the key and nonce
		stream := cipher.NewCTR(newAesBlock, nonce)
		stream.XORKeyStream(encryptedBlock, block[:n])
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(encryptedBlock)
		if err != nil {
			return nil, err
		}
	}
	log.Debug().Msg("Chunk should now be encrypted")
	return buffer, nil
}
func generateUniqueID() string {
	guid := xid.New()
	return guid.String()
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

func createMessageFromPost(c *gin.Context, siteHost string, guid string, encryptionRequest *pb.EncryptedMessageRequest) message.MessagePost {
	msg := message.MessagePost{
		Email:          []string{c.PostForm("email")},
		FirstName:      c.PostForm("firstname"),
		OtherFirstName: c.PostForm("other_firstname"),
		OtherLastName:  c.PostForm("other_lastname"),
		OtherEmail:     []string{c.PostForm("other_email")},
		Uniqueid:       "",
		Content:        "",
		Errors:         map[string]string{},
		Url:            siteHost + "decrypt/" + guid + "/" + string(b64.URLEncoding.EncodeToString(encryptionRequest.Key)),
		Hidden:         c.PostForm("other_information"),
		Captcha:        c.PostForm("h-captcha-response"),
	}
	msg.Content = "please click this link to get your encrypted message" + "\n <a href=\"" + msg.URL + "\"> here</a>"
	return msg
}

func confirmation(c *gin.Context) {
	content := c.Query("content")
	extraHeaders := htmlHeaders{Title: "passwordExchange", URL: content}

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
