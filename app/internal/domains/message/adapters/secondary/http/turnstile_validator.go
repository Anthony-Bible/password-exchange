package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
)

const (
	turnstileValidationURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	requestTimeout         = 10 * time.Second
)

// TurnstileValidator implements the TurnstileValidatorPort interface.
type TurnstileValidator struct {
	secret        string
	httpClient    *http.Client
	validationURL string
}

// TurnstileRequest represents the JSON request payload for Turnstile validation.
type TurnstileRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

// TurnstileResponse represents the response from Cloudflare Turnstile API.
type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
	CData       string   `json:"cdata,omitempty"`
}

// NewTurnstileValidator creates a new TurnstileValidator instance.
func NewTurnstileValidator(secret string) secondary.TurnstileValidatorPort {
	return &TurnstileValidator{
		secret:        secret,
		httpClient:    &http.Client{Timeout: requestTimeout},
		validationURL: turnstileValidationURL,
	}
}

// ValidateToken validates a Turnstile token by sending a JSON request to Cloudflare.
func (t *TurnstileValidator) ValidateToken(ctx context.Context, token string, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("turnstile token is required")
	}
	if t.secret == "" {
		return false, fmt.Errorf("turnstile secret is not configured")
	}

	jsonData, err := json.Marshal(TurnstileRequest{
		Secret:   t.secret,
		Response: token,
		RemoteIP: remoteIP,
	})
	if err != nil {
		return false, fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.validationURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PasswordExchange/1.0")

	return t.doValidationRequest(req)
}

// ValidateTokenWithFormData validates a Turnstile token using form-encoded data.
// Cloudflare accepts both JSON and form encoding; use this variant when integrating
// with proxies or legacy tooling that normalises the Content-Type.
func (t *TurnstileValidator) ValidateTokenWithFormData(ctx context.Context, token string, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("turnstile token is required")
	}
	if t.secret == "" {
		return false, fmt.Errorf("turnstile secret is not configured")
	}

	formData := url.Values{}
	formData.Set("secret", t.secret)
	formData.Set("response", token)
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.validationURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "PasswordExchange/1.0")

	return t.doValidationRequest(req)
}

// doValidationRequest executes a pre-built request and interprets the Cloudflare response.
func (t *TurnstileValidator) doValidationRequest(req *http.Request) (bool, error) {
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected HTTP status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result TurnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			return false, fmt.Errorf("turnstile validation failed with error codes: %v", result.ErrorCodes)
		}
		return false, fmt.Errorf("turnstile validation failed")
	}

	return true, nil
}
