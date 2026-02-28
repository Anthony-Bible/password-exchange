package api

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// OpenAPISpec represents the basic structure of an OpenAPI specification
type OpenAPISpec struct {
	OpenAPI string `yaml:"openapi"`
	Info    struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	} `yaml:"info"`
	Paths      map[string]interface{} `yaml:"paths"`
	Components struct {
		Schemas map[string]interface{} `yaml:"schemas"`
	} `yaml:"components"`
}

func TestOpenAPISpecExists(t *testing.T) {
	// Test that the OpenAPI specification file exists and is valid YAML
	specData, err := ioutil.ReadFile("../../../../../../api/openapi.yaml")
	require.NoError(t, err, "OpenAPI specification file should exist")

	var spec OpenAPISpec
	err = yaml.Unmarshal(specData, &spec)
	require.NoError(t, err, "OpenAPI specification should be valid YAML")

	// Basic validation
	assert.Equal(t, "3.0.3", spec.OpenAPI, "Should use OpenAPI 3.0.3")
	assert.Equal(t, "Password Exchange API", spec.Info.Title, "Should have correct title")
	assert.Equal(t, "1.0.0", spec.Info.Version, "Should have correct version")
	assert.NotEmpty(t, spec.Info.Description, "Should have description")

	// Check required paths exist
	requiredPaths := []string{
		"/messages",
		"/messages/{messageId}",
		"/messages/{messageId}/decrypt",
		"/health",
		"/info",
	}

	for _, path := range requiredPaths {
		assert.Contains(t, spec.Paths, path, "OpenAPI spec should contain path: %s", path)
	}

	// Check required schemas exist
	requiredSchemas := []string{
		"MessageSubmissionRequest",
		"MessageSubmissionResponse",
		"MessageAccessInfoResponse",
		"MessageDecryptRequest",
		"MessageDecryptResponse",
		"HealthCheckResponse",
		"APIInfoResponse",
		"StandardErrorResponse",
		"Sender",
		"Recipient",
	}

	for _, schema := range requiredSchemas {
		assert.Contains(t, spec.Components.Schemas, schema, "OpenAPI spec should contain schema: %s", schema)
	}
}

func TestMessageSubmissionRequestSchema(t *testing.T) {
	// Test that MessageSubmissionRequest can be properly serialized/deserialized
	tests := []struct {
		name    string
		request *models.MessageSubmissionRequest
		valid   bool
	}{
		{
			name: "complete valid request",
			request: &models.MessageSubmissionRequest{
				Content: "Test message content",
				Sender: &models.Sender{
					Name:  "John Doe",
					Email: "john@example.com",
				},
				Recipient: &models.Recipient{
					Name:  "Jane Smith",
					Email: "jane@example.com",
				},
				Passphrase:       "secure-passphrase",
				AdditionalInfo:   "Please access within 24 hours",
				SendNotification: true,
				AntiSpamAnswer:   "blue",
			},
			valid: true,
		},
		{
			name: "minimal valid request",
			request: &models.MessageSubmissionRequest{
				Content: "Test message",
				Recipient: &models.Recipient{
					Name: "Jane Smith",
				},
				SendNotification: false,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err, "Should be able to marshal to JSON")

			// Test JSON deserialization
			var decoded models.MessageSubmissionRequest
			err = json.Unmarshal(jsonData, &decoded)
			require.NoError(t, err, "Should be able to unmarshal from JSON")

			// Verify roundtrip consistency
			assert.Equal(t, tt.request.Content, decoded.Content)
			assert.Equal(t, tt.request.SendNotification, decoded.SendNotification)
			assert.Equal(t, tt.request.AntiSpamAnswer, decoded.AntiSpamAnswer)

			if tt.request.Sender != nil {
				require.NotNil(t, decoded.Sender)
				assert.Equal(t, tt.request.Sender.Name, decoded.Sender.Name)
				assert.Equal(t, tt.request.Sender.Email, decoded.Sender.Email)
			}

			if tt.request.Recipient != nil {
				require.NotNil(t, decoded.Recipient)
				assert.Equal(t, tt.request.Recipient.Name, decoded.Recipient.Name)
				assert.Equal(t, tt.request.Recipient.Email, decoded.Recipient.Email)
			}
		})
	}
}

func TestMessageSubmissionResponseSchema(t *testing.T) {
	response := &models.MessageSubmissionResponse{
		MessageID:        "123e4567-e89b-12d3-a456-426614174000",
		DecryptURL:       "https://api.password.exchange/api/v1/messages/123e4567-e89b-12d3-a456-426614174000/decrypt?key=YWJjZGVmZ2hpams=",
		WebURL:           "https://password.exchange/decrypt/123e4567-e89b-12d3-a456-426614174000/YWJjZGVmZ2hpams=",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		NotificationSent: true,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Should be able to marshal to JSON")

	// Test JSON deserialization
	var decoded models.MessageSubmissionResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err, "Should be able to unmarshal from JSON")

	// Verify roundtrip consistency
	assert.Equal(t, response.MessageID, decoded.MessageID)
	assert.Equal(t, response.DecryptURL, decoded.DecryptURL)
	assert.Equal(t, response.WebURL, decoded.WebURL)
	assert.Equal(t, response.NotificationSent, decoded.NotificationSent)
	// Note: Time comparison might need tolerance for precision differences
}

func TestMessageAccessInfoResponseSchema(t *testing.T) {
	response := &models.MessageAccessInfoResponse{
		MessageID:          "123e4567-e89b-12d3-a456-426614174000",
		Exists:             true,
		RequiresPassphrase: true,
		HasBeenAccessed:    false,
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Should be able to marshal to JSON")

	// Test JSON deserialization
	var decoded models.MessageAccessInfoResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err, "Should be able to unmarshal from JSON")

	// Verify roundtrip consistency
	assert.Equal(t, response.MessageID, decoded.MessageID)
	assert.Equal(t, response.Exists, decoded.Exists)
	assert.Equal(t, response.RequiresPassphrase, decoded.RequiresPassphrase)
	assert.Equal(t, response.HasBeenAccessed, decoded.HasBeenAccessed)
}

func TestMessageDecryptRequestSchema(t *testing.T) {
	tests := []struct {
		name    string
		request *models.MessageDecryptRequest
	}{
		{
			name: "with passphrase",
			request: &models.MessageDecryptRequest{
				DecryptionKey: "YWJjZGVmZ2hpams=",
				Passphrase:    "secure-passphrase",
			},
		},
		{
			name: "without passphrase",
			request: &models.MessageDecryptRequest{
				DecryptionKey: "YWJjZGVmZ2hpams=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err, "Should be able to marshal to JSON")

			// Test JSON deserialization
			var decoded models.MessageDecryptRequest
			err = json.Unmarshal(jsonData, &decoded)
			require.NoError(t, err, "Should be able to unmarshal from JSON")

			// Verify roundtrip consistency
			assert.Equal(t, tt.request.DecryptionKey, decoded.DecryptionKey)
			assert.Equal(t, tt.request.Passphrase, decoded.Passphrase)
		})
	}
}

func TestMessageDecryptResponseSchema(t *testing.T) {
	response := &models.MessageDecryptResponse{
		MessageID:   "123e4567-e89b-12d3-a456-426614174000",
		Content:     "This is the secret message content",
		DecryptedAt: time.Now(),
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Should be able to marshal to JSON")

	// Test JSON deserialization
	var decoded models.MessageDecryptResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err, "Should be able to unmarshal from JSON")

	// Verify roundtrip consistency
	assert.Equal(t, response.MessageID, decoded.MessageID)
	assert.Equal(t, response.Content, decoded.Content)
}

func TestHealthCheckResponseSchema(t *testing.T) {
	response := &models.HealthCheckResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Services: map[string]string{
			"database":   "healthy",
			"encryption": "healthy",
			"email":      "healthy",
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Should be able to marshal to JSON")

	// Test JSON deserialization
	var decoded models.HealthCheckResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err, "Should be able to unmarshal from JSON")

	// Verify roundtrip consistency
	assert.Equal(t, response.Status, decoded.Status)
	assert.Equal(t, response.Version, decoded.Version)
	assert.Equal(t, response.Services, decoded.Services)
}

func TestAPIInfoResponseSchema(t *testing.T) {
	response := &models.APIInfoResponse{
		Version:       "1.0.0",
		Documentation: "/api/v1/docs",
		Endpoints: map[string]string{
			"submit":  "POST /api/v1/messages",
			"access":  "GET /api/v1/messages/{id}",
			"decrypt": "POST /api/v1/messages/{id}/decrypt",
		},
		Features: map[string]bool{
			"emailNotifications":   true,
			"passphraseProtection": true,
			"antiSpamProtection":   true,
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Should be able to marshal to JSON")

	// Test JSON deserialization
	var decoded models.APIInfoResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err, "Should be able to unmarshal from JSON")

	// Verify roundtrip consistency
	assert.Equal(t, response.Version, decoded.Version)
	assert.Equal(t, response.Documentation, decoded.Documentation)
	assert.Equal(t, response.Endpoints, decoded.Endpoints)
	assert.Equal(t, response.Features, decoded.Features)
}

func TestStandardErrorResponseSchema(t *testing.T) {
	response := &models.StandardErrorResponse{
		Error:     "validation_failed",
		Message:   "Request validation failed",
		Timestamp: time.Now(),
		Path:      "/api/v1/messages",
		Details: map[string]interface{}{
			"content":      "Content is required",
			"sender.email": "Valid email required when notifications enabled",
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	require.NoError(t, err, "Should be able to marshal to JSON")

	// Test JSON deserialization
	var decoded models.StandardErrorResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err, "Should be able to unmarshal from JSON")

	// Verify roundtrip consistency
	assert.Equal(t, response.Error, decoded.Error)
	assert.Equal(t, response.Message, decoded.Message)
	assert.Equal(t, response.Path, decoded.Path)
	assert.Equal(t, response.Details, decoded.Details)
}

func TestJSONFieldNaming(t *testing.T) {
	// Test that JSON field names match OpenAPI specification
	request := &models.MessageSubmissionRequest{
		Content:          "Test",
		SendNotification: true,
		AntiSpamAnswer:   "blue",
	}

	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	require.NoError(t, err)

	// Verify expected JSON field names
	expectedFields := []string{
		"content",
		"sendNotification",
		"antiSpamAnswer",
	}

	for _, field := range expectedFields {
		assert.Contains(t, jsonMap, field, "JSON should contain field: %s", field)
	}

	// Verify camelCase naming convention
	assert.Contains(t, jsonMap, "sendNotification", "Should use camelCase for sendNotification")
	assert.Contains(t, jsonMap, "antiSpamAnswer", "Should use camelCase for antiSpamAnswer")
}
