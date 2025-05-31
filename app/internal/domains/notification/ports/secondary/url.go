package secondary

import "context"

// URLBuilderPort defines the secondary port for URL construction operations
type URLBuilderPort interface {
	BuildDecryptionURL(ctx context.Context, uniqueID string) (string, error)
	BuildPasswordExchangeURL() string
	BuildAboutURL() string
}