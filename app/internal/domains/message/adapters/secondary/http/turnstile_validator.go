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

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/domain"
)

const (
	turnstileValidationURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	requestTimeout         = 10 * time.Second
)

// TurnstileValidator implements the TurnstileValidationPort interface
type TurnstileValidator struct {
	secret     string
	httpClient *http.Client
}

// TurnstileRequest represents the request payload for Turnstile validation
type TurnstileRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

// TurnstileResponse represents the response from Cloudflare Turnstile API
type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
	CData       string   `json:"cdata,omitempty"`
}

// NewTurnstileValidator creates a new TurnstileValidator instance
func NewTurnstileValidator(secret string) domain.TurnstileValidator {
	return &TurnstileValidator{
		secret: secret,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// ValidateToken validates a Turnstile token with Cloudflare's API
func (t *TurnstileValidator) ValidateToken(ctx context.Context, token string, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("turnstile token is required")
	}

	if t.secret == "" {
		return false, fmt.Errorf("turnstile secret is not configured")
	}

	// Prepare request payload
	requestData := TurnstileRequest{
		Secret:   t.secret,
		Response: token,
		RemoteIP: remoteIP,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", turnstileValidationURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PasswordExchange/1.0")

	// Make the request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected HTTP status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var turnstileResp TurnstileResponse
	if err := json.Unmarshal(body, &turnstileResp); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Return validation result
	if !turnstileResp.Success {
		// Log error codes for debugging (in production, you might want to use a proper logger)
		if len(turnstileResp.ErrorCodes) > 0 {
			return false, fmt.Errorf("turnstile validation failed with error codes: %v", turnstileResp.ErrorCodes)
		}
		return false, fmt.Errorf("turnstile validation failed")
	}

	return true, nil
}

// ValidateTokenWithFormData is an alternative implementation using form data
// This can be used if you prefer form-encoded requests over JSON
func (t *TurnstileValidator) ValidateTokenWithFormData(ctx context.Context, token string, remoteIP string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("turnstile token is required")
	}

	if t.secret == "" {
		return false, fmt.Errorf("turnstile secret is not configured")
	}

	// Prepare form data
	formData := url.Values{}
	formData.Set("secret", t.secret)
	formData.Set("response", token)
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", turnstileValidationURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "PasswordExchange/1.0")

	// Make the request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected HTTP status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var turnstileResp TurnstileResponse
	if err := json.Unmarshal(body, &turnstileResp); err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Return validation result
	if !turnstileResp.Success {
		if len(turnstileResp.ErrorCodes) > 0 {
			return false, fmt.Errorf("turnstile validation failed with error codes: %v", turnstileResp.ErrorCodes)
		}
		return false, fmt.Errorf("turnstile validation failed")
	}

	return true, nil
}