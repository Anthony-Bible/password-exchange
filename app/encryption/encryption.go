package encryption

import (
    "crypto/rand"
    "crypto/aes"
	"encoding/base64"
    "github.com/rs/xid"
	"crypto/cipher"
	"errors"
	"io"
	"net"
	"log"
	"fmt"
	// "password.exchange/message"
	// b "password.exchange/aws"
	"github.com/Anthony-Bible/password-exchange/app/encryptionpb"
	"google.golang.org/grpc"
)


// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.

func GenerateRandomBytes(n int) (*[32]byte) {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}


// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (*[32]byte, string) {
    b := GenerateRandomBytes(s)
    return b, base64.URLEncoding.EncodeToString((b[:]))
}

func Generateid() (string) {
    guid := xid.New()
    return guid.String()
}
func MessageEncrypt(plaintext []byte, key *[32]byte) (ciphertext string) {

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		panic(err)
	}
	urlEncodedString := base64.URLEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil))
	return urlEncodedString
}

func MessageDecrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}


// func main() {
// 	msgEncrypted := &message.Message{
// 	Email:   string(MessageEncrypt([]byte(c.PostForm("email")), encryptionbytes)),
//     FirstName: string(MessageEncrypt([]byte(c.PostForm("firstname")), encryptionbytes)),
//     OtherFirstName: string(MessageEncrypt([]byte(c.PostForm("other_firstname")), encryptionbytes)),
//     OtherLastName: string(MessageEncrypt([]byte(c.PostForm("other_lastname")), encryptionbytes)),
//     OtherEmail: string(MessageEncrypt([]byte(c.PostForm("other_email")), encryptionbytes)),
//     Content: string(MessageEncrypt([]byte(c.PostForm("content")), encryptionbytes)),
//     Uniqueid: guid.String(),
//   }
//   sess := b.BuildSession()
//   // queueurl, _ := b.GetQueueURL(sess, "arn:aws:sns:us-west-2:842805395457:my-test.fifo")
//   fmt.Println(len(urlEncodedString))
//   fmt.Println(urlEncodedString)

//   b.SendSNS(sess, "arn:aws:sns:us-west-2:842805395457:my-test.fifo", msgEncrypted)
// }


type server struct {

}

func (*server) encryptMessage(ctx context.Context, request *encryption.Message) (*encryption.Message, error) {
	// name := request.Name
	// response := &hellopb.HelloResponse{
	// 	Greeting: "Hello " + name,
	// }
	// return response, nil

	msgEncrypted := &encryption.Message{
		Email:   string(MessageEncrypt([]byte(c.PostForm("email")), encryptionbytes)),
		FirstName: string(MessageEncrypt([]byte(c.PostForm("firstname")), encryptionbytes)),
		OtherFirstName: string(MessageEncrypt([]byte(c.PostForm("other_firstname")), encryptionbytes)),
		OtherLastName: string(MessageEncrypt([]byte(c.PostForm("other_lastname")), encryptionbytes)),
		OtherEmail: string(MessageEncrypt([]byte(c.PostForm("other_email")), encryptionbytes)),
		Content: string(MessageEncrypt([]byte(c.PostForm("content")), encryptionbytes)),
		Uniqueid: guid.String(),
	  }
	  return msgEncrypted
}


func main() {
	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("Server is listening on %v ...", address)

	s := grpc.NewServer()
	encryption.RegisterMessageServiceServer(s, &server{})

	s.Serve(lis)
}