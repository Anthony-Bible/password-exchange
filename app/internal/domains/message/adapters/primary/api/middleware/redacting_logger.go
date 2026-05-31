package middleware

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// sensitiveQueryParams lists query parameter names whose values must be
// redacted from access logs. These are cryptographic secrets that would
// leak to anyone with log access if printed verbatim.
var sensitiveQueryParams = []string{"key"}

// RedactingLogger returns a gin middleware that logs each request using
// the provided writer, with the values of sensitiveQueryParams replaced
// by "[REDACTED]". It is a drop-in replacement for gin.Logger() that
// prevents encryption keys from appearing in access logs.
func RedactingLogger(out io.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := redactQueryParams(c.Request.URL.RawQuery)

		c.Next()

		latency := time.Since(start)
		method := c.Request.Method
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		loggedPath := path
		if rawQuery != "" {
			loggedPath = path + "?" + rawQuery
		}

		fmt.Fprintf(out, "[GIN] %v | %3d | %13v | %15s | %-7s %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			loggedPath,
		)
	}
}

// redactQueryParams returns a query string with the values of any
// sensitiveQueryParams entries replaced by "[REDACTED]". Non-sensitive
// parameters are preserved verbatim. An empty input returns an empty string.
func redactQueryParams(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		// Malformed query — redact the whole thing rather than risk leaking anything.
		return "[REDACTED-MALFORMED]"
	}

	for _, name := range sensitiveQueryParams {
		if _, present := values[name]; present {
			values[name] = []string{"[REDACTED]"}
		}
	}

	return values.Encode()
}
