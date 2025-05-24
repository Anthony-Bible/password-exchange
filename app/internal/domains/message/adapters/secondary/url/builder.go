package url

import (
	"encoding/base64"
	"fmt"

	"github.com/rs/zerolog/log"
)

// URLBuilder implements the URLBuilderPort
type URLBuilder struct {
	baseURL string
}

// NewURLBuilder creates a new URL builder
func NewURLBuilder(baseURL string) *URLBuilder {
	return &URLBuilder{
		baseURL: baseURL,
	}
}

// BuildDecryptURL builds a URL for decrypting a message
func (u *URLBuilder) BuildDecryptURL(messageID string, encryptionKey []byte) string {
	encodedKey := base64.URLEncoding.EncodeToString(encryptionKey)
	decryptURL := fmt.Sprintf("%sdecrypt/%s/%s", u.baseURL, messageID, encodedKey)
	
	log.Debug().Str("messageId", messageID).Str("url", decryptURL).Msg("Built decrypt URL")
	return decryptURL
}