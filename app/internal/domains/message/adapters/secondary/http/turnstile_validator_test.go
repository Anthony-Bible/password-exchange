package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time proof that NewTurnstileValidator satisfies the port interface.
var _ secondary.TurnstileValidatorPort = NewTurnstileValidator("secret")

func turnstileSuccessResponse() string {
	return `{"success":true,"challenge_ts":"2024-01-01T00:00:00Z","hostname":"example.com"}`
}

func turnstileFailureResponse(codes ...string) string {
	b, _ := json.Marshal(TurnstileResponse{
		Success:    false,
		ErrorCodes: codes,
	})
	return string(b)
}

// newTestValidator bypasses NewTurnstileValidator to inject an arbitrary validation URL.
func newTestValidator(secret, serverURL string) *TurnstileValidator {
	return &TurnstileValidator{
		secret:        secret,
		httpClient:    &http.Client{},
		validationURL: serverURL,
	}
}

func TestValidateToken_ValidToken_ReturnsTrue(t *testing.T) {
	t.Parallel()

	type capturedRequest struct {
		Secret   string `json:"secret"`
		Response string `json:"response"`
		RemoteIP string `json:"remoteip"`
	}

	var captured capturedRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if jsonErr := json.Unmarshal(body, &captured); jsonErr != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, turnstileSuccessResponse())
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateToken(context.Background(), "valid-token", "1.2.3.4")

	require.NoError(t, err)
	assert.True(t, ok)

	// Verify the server received the correct payload fields.
	assert.Equal(t, "my-secret", captured.Secret, "secret field must be forwarded to Cloudflare")
	assert.Equal(t, "valid-token", captured.Response, "token must be sent as 'response' field")
	assert.Equal(t, "1.2.3.4", captured.RemoteIP, "remoteIP must be forwarded to Cloudflare")
}

func TestValidateToken_InvalidToken_ReturnsErrorWithCodes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, turnstileFailureResponse("invalid-input-response"))
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateToken(context.Background(), "bad-token", "")

	assert.False(t, ok)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid-input-response", "error must mention the error codes returned by Cloudflare")
}

func TestValidateToken_EmptyToken_ReturnsErrorWithoutHTTPCall(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprint(w, turnstileSuccessResponse())
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateToken(context.Background(), "", "1.2.3.4")

	assert.False(t, ok)
	require.Error(t, err)
	assert.Equal(t, 0, callCount, "no HTTP request must be made when the token is empty")
}

func TestValidateToken_EmptySecret_ReturnsErrorWithoutHTTPCall(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprint(w, turnstileSuccessResponse())
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("", server.URL)
	ok, err := v.ValidateToken(context.Background(), "some-token", "1.2.3.4")

	assert.False(t, ok)
	require.Error(t, err)
	assert.Equal(t, 0, callCount, "no HTTP request must be made when the secret is empty")
}

func TestValidateToken_NetworkError_ReturnsError(t *testing.T) {
	t.Parallel()

	// Start and immediately shut down the server so all connections are refused.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateToken(context.Background(), "some-token", "")

	assert.False(t, ok)
	require.Error(t, err)
}

func TestValidateToken_Non200Response_ReturnsErrorMentioningStatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
	}{
		{"InternalServerError", http.StatusInternalServerError},
		{"BadGateway", http.StatusBadGateway},
		{"ServiceUnavailable", http.StatusServiceUnavailable},
		{"TooManyRequests", http.StatusTooManyRequests},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				fmt.Fprint(w, "upstream error")
			}))
			t.Cleanup(server.Close)

			v := newTestValidator("my-secret", server.URL)
			ok, err := v.ValidateToken(context.Background(), "some-token", "")

			assert.False(t, ok)
			require.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("%d", tc.statusCode),
				"error must mention the HTTP status code received")
		})
	}
}

func TestValidateToken_MalformedJSONResponse_ReturnsError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{this is not valid json}`)
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateToken(context.Background(), "some-token", "")

	assert.False(t, ok)
	require.Error(t, err)
}

func TestValidateToken_SuccessFalseNoErrorCodes_ReturnsGenericError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success":false}`)
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateToken(context.Background(), "some-token", "")

	assert.False(t, ok)
	require.Error(t, err)
}

func TestValidateTokenWithFormData_ValidToken_ReturnsTrue(t *testing.T) {
	t.Parallel()

	var (
		capturedContentType string
		capturedSecret      string
		capturedResponse    string
		capturedRemoteIP    string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		values, err := url.ParseQuery(string(body))
		if err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		capturedSecret = values.Get("secret")
		capturedResponse = values.Get("response")
		capturedRemoteIP = values.Get("remoteip")

		fmt.Fprint(w, turnstileSuccessResponse())
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateTokenWithFormData(context.Background(), "valid-token", "1.2.3.4")

	require.NoError(t, err)
	assert.True(t, ok)

	// Verify the request used form encoding and carried the correct fields.
	assert.True(t,
		strings.HasPrefix(capturedContentType, "application/x-www-form-urlencoded"),
		"Content-Type must be application/x-www-form-urlencoded, got: %s", capturedContentType,
	)
	assert.Equal(t, "my-secret", capturedSecret, "secret form field must be forwarded")
	assert.Equal(t, "valid-token", capturedResponse, "token must be sent as 'response' form field")
	assert.Equal(t, "1.2.3.4", capturedRemoteIP, "remoteIP must be sent as 'remoteip' form field")
}

func TestValidateTokenWithFormData_EmptyToken_ReturnsErrorWithoutHTTPCall(t *testing.T) {
	t.Parallel()

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprint(w, turnstileSuccessResponse())
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateTokenWithFormData(context.Background(), "", "1.2.3.4")

	assert.False(t, ok)
	require.Error(t, err)
	assert.Equal(t, 0, callCount, "no HTTP request must be made when the token is empty")
}

func TestValidateTokenWithFormData_InvalidToken_ReturnsError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, turnstileFailureResponse("invalid-input-response", "timeout-or-duplicate"))
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	ok, err := v.ValidateTokenWithFormData(context.Background(), "expired-token", "")

	assert.False(t, ok)
	require.Error(t, err)
}

func TestValidateTokenWithFormData_OmitsRemoteIPWhenEmpty(t *testing.T) {
	t.Parallel()

	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedBody = string(body)
		fmt.Fprint(w, turnstileSuccessResponse())
	}))
	t.Cleanup(server.Close)

	v := newTestValidator("my-secret", server.URL)
	_, _ = v.ValidateTokenWithFormData(context.Background(), "some-token", "")

	values, err := url.ParseQuery(capturedBody)
	require.NoError(t, err)
	assert.False(t, values.Has("remoteip"),
		"remoteip form field must be omitted when remoteIP is empty, body was: %s", capturedBody)
}
