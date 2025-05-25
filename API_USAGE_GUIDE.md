# Password Exchange API Usage Guide

The Password Exchange API provides a secure way to share sensitive information through encrypted, one-time access URLs. This guide shows you how to use the REST API programmatically.

## Quick Start

### Base URLs
- **Production**: `https://api.password.exchange/api/v1`
- **Development**: `http://localhost:8080/api/v1`

### Authentication
No authentication is required. The API is designed for public access with built-in rate limiting.

## Core Workflow

1. **Submit Message** → Get unique URL and message ID
2. **Check Message** → Verify message exists and requirements  
3. **Decrypt Message** → Retrieve content (one-time access)

## API Endpoints

### 1. Submit Message

Submit an encrypted message and optionally send email notifications.

```bash
curl -X POST https://api.password.exchange/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "This is my secret password: admin123",
    "sendNotification": false
  }'
```

**Response:**
```json
{
  "messageId": "550e8400-e29b-41d4-a716-446655440000",
  "decryptUrl": "https://password.exchange/decrypt/550e8400-e29b-41d4-a716-446655440000/eyJhbGciOiJIUzI1NiJ9...",
  "webUrl": "https://password.exchange/decrypt/550e8400-e29b-41d4-a716-446655440000/eyJhbGciOiJIUzI1NiJ9...",
  "expiresAt": "2024-01-01T12:00:00Z",
  "notificationSent": false
}
```

#### With Email Notification

```bash
curl -X POST https://api.password.exchange/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Database password: secretDB2024!",
    "sender": {
      "name": "John Smith",
      "email": "john@company.com"
    },
    "recipient": {
      "name": "Jane Doe", 
      "email": "jane@company.com"
    },
    "passphrase": "meeting-room-code",
    "sendNotification": true,
    "antiSpamAnswer": "blue"
  }'
```

### 2. Check Message Access

Check if a message exists and what's required to access it.

```bash
curl "https://api.password.exchange/api/v1/messages/550e8400-e29b-41d4-a716-446655440000?key=eyJhbGciOiJIUzI1NiJ9..."
```

**Response:**
```json
{
  "messageId": "550e8400-e29b-41d4-a716-446655440000",
  "exists": true,
  "requiresPassphrase": true,
  "hasBeenAccessed": false,
  "expiresAt": "2024-01-01T12:00:00Z"
}
```

### 3. Decrypt Message

Retrieve the message content (one-time access).

```bash
curl -X POST https://api.password.exchange/api/v1/messages/550e8400-e29b-41d4-a716-446655440000/decrypt \
  -H "Content-Type: application/json" \
  -d '{
    "decryptionKey": "eyJhbGciOiJIUzI1NiJ9...",
    "passphrase": "meeting-room-code"
  }'
```

**Response:**
```json
{
  "messageId": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Database password: secretDB2024!",
  "decryptedAt": "2024-01-01T10:30:00Z"
}
```

### 4. Health Check

Check API service status.

```bash
curl https://api.password.exchange/api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "encryption": "healthy",
    "email": "healthy"
  }
}
```

### 5. API Information

Get API version and capabilities.

```bash
curl https://api.password.exchange/api/v1/info
```

## Code Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

class PasswordExchangeClient {
  constructor(baseUrl = 'https://api.password.exchange/api/v1') {
    this.baseUrl = baseUrl;
  }

  async submitMessage(content, options = {}) {
    const payload = {
      content,
      sendNotification: false,
      ...options
    };

    const response = await axios.post(`${this.baseUrl}/messages`, payload);
    return response.data;
  }

  async getMessageInfo(messageId, decryptionKey) {
    const response = await axios.get(`${this.baseUrl}/messages/${messageId}?key=${encodeURIComponent(decryptionKey)}`);
    return response.data;
  }

  async decryptMessage(messageId, decryptionKey, passphrase = null) {
    const payload = { decryptionKey };
    if (passphrase) payload.passphrase = passphrase;

    const response = await axios.post(`${this.baseUrl}/messages/${messageId}/decrypt`, payload);
    return response.data;
  }
}

// Usage example
async function example() {
  const client = new PasswordExchangeClient();
  
  // Submit a message
  const result = await client.submitMessage("My secret password: admin123");
  console.log('Message submitted:', result.messageId);
  console.log('Share this URL:', result.webUrl);
  
  // Extract key from URL for API access
  const url = new URL(result.decryptUrl);
  const key = url.searchParams.get('key');
  
  // Check message
  const info = await client.getMessageInfo(result.messageId, key);
  console.log('Message exists:', info.exists);
  
  // Decrypt message
  const decrypted = await client.decryptMessage(result.messageId, key);
  console.log('Content:', decrypted.content);
}
```

### Python

```python
import requests
import json
from urllib.parse import urlparse, parse_qs

class PasswordExchangeClient:
    def __init__(self, base_url='https://api.password.exchange/api/v1'):
        self.base_url = base_url
        
    def submit_message(self, content, **kwargs):
        """Submit a new encrypted message"""
        payload = {
            'content': content,
            'sendNotification': False,
            **kwargs
        }
        
        response = requests.post(f'{self.base_url}/messages', json=payload)
        response.raise_for_status()
        return response.json()
    
    def get_message_info(self, message_id, decryption_key):
        """Check if message exists and get access requirements"""
        params = {'key': decryption_key}
        response = requests.get(f'{self.base_url}/messages/{message_id}', params=params)
        response.raise_for_status()
        return response.json()
    
    def decrypt_message(self, message_id, decryption_key, passphrase=None):
        """Decrypt and retrieve message content (one-time access)"""
        payload = {'decryptionKey': decryption_key}
        if passphrase:
            payload['passphrase'] = passphrase
            
        response = requests.post(f'{self.base_url}/messages/{message_id}/decrypt', json=payload)
        response.raise_for_status()
        return response.json()

# Usage example
def main():
    client = PasswordExchangeClient()
    
    # Submit message with email notification
    result = client.submit_message(
        content="Database connection: postgresql://user:secret@db.example.com/myapp",
        sender={'name': 'DevOps Team', 'email': 'devops@company.com'},
        recipient={'name': 'Developer', 'email': 'dev@company.com'},
        passphrase='deploy-key-2024',
        sendNotification=True,
        antiSpamAnswer='blue'
    )
    
    print(f"Message ID: {result['messageId']}")
    print(f"Web URL: {result['webUrl']}")
    
    # Extract decryption key from URL
    parsed_url = urlparse(result['decryptUrl'])
    key = parse_qs(parsed_url.query)['key'][0]
    
    # Check message status
    info = client.get_message_info(result['messageId'], key)
    print(f"Requires passphrase: {info['requiresPassphrase']}")
    
    # Decrypt message
    decrypted = client.decrypt_message(
        result['messageId'], 
        key, 
        passphrase='deploy-key-2024'
    )
    print(f"Content: {decrypted['content']}")

if __name__ == '__main__':
    main()
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

type PasswordExchangeClient struct {
    BaseURL string
    Client  *http.Client
}

type MessageSubmissionRequest struct {
    Content          string     `json:"content"`
    Sender           *Person    `json:"sender,omitempty"`
    Recipient        *Person    `json:"recipient,omitempty"`
    Passphrase       string     `json:"passphrase,omitempty"`
    SendNotification bool       `json:"sendNotification"`
    AntiSpamAnswer   string     `json:"antiSpamAnswer,omitempty"`
}

type Person struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type MessageSubmissionResponse struct {
    MessageID        string `json:"messageId"`
    DecryptURL       string `json:"decryptUrl"`
    WebURL           string `json:"webUrl"`
    ExpiresAt        string `json:"expiresAt"`
    NotificationSent bool   `json:"notificationSent"`
}

func NewClient(baseURL string) *PasswordExchangeClient {
    return &PasswordExchangeClient{
        BaseURL: baseURL,
        Client:  &http.Client{},
    }
}

func (c *PasswordExchangeClient) SubmitMessage(req MessageSubmissionRequest) (*MessageSubmissionResponse, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := c.Client.Post(c.BaseURL+"/messages", "application/json", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result MessageSubmissionResponse
    err = json.NewDecoder(resp.Body).Decode(&result)
    return &result, err
}

func main() {
    client := NewClient("https://api.password.exchange/api/v1")
    
    // Submit message
    req := MessageSubmissionRequest{
        Content:          "SSH Key: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
        Passphrase:       "server-access-2024",
        SendNotification: false,
    }
    
    result, err := client.SubmitMessage(req)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Message ID: %s\n", result.MessageID)
    fmt.Printf("Share URL: %s\n", result.WebURL)
}
```

### Shell Script (Bash)

```bash
#!/bin/bash

API_BASE="https://api.password.exchange/api/v1"

# Function to submit a message
submit_message() {
    local content="$1"
    local passphrase="$2"
    
    curl -s -X POST "$API_BASE/messages" \
        -H "Content-Type: application/json" \
        -d "{
            \"content\": \"$content\",
            \"passphrase\": \"$passphrase\",
            \"sendNotification\": false
        }"
}

# Function to decrypt a message
decrypt_message() {
    local message_id="$1"
    local key="$2"
    local passphrase="$3"
    
    curl -s -X POST "$API_BASE/messages/$message_id/decrypt" \
        -H "Content-Type: application/json" \
        -d "{
            \"decryptionKey\": \"$key\",
            \"passphrase\": \"$passphrase\"
        }"
}

# Example usage
main() {
    # Submit message
    echo "Submitting message..."
    RESULT=$(submit_message "Database password: mySecretPass123" "room-401")
    
    # Parse response (requires jq)
    MESSAGE_ID=$(echo "$RESULT" | jq -r '.messageId')
    WEB_URL=$(echo "$RESULT" | jq -r '.webUrl')
    
    echo "Message ID: $MESSAGE_ID"
    echo "Share this URL: $WEB_URL"
    
    # Extract key from decrypt URL
    DECRYPT_URL=$(echo "$RESULT" | jq -r '.decryptUrl')
    KEY=$(echo "$DECRYPT_URL" | sed 's/.*key=\([^&]*\).*/\1/')
    
    # Decrypt message
    echo "Decrypting message..."
    CONTENT=$(decrypt_message "$MESSAGE_ID" "$KEY" "room-401")
    echo "Content: $(echo "$CONTENT" | jq -r '.content')"
}

main "$@"
```

## Rate Limits

The API enforces rate limits per IP address:

- **Message Submission**: 10 requests/hour
- **Message Access**: 100 requests/hour  
- **Message Decryption**: 20 requests/hour
- **Health/Info**: 200 requests/hour

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 9
X-RateLimit-Reset: 1640995200
```

## Error Handling

All errors follow RFC 7807 format:

```json
{
  "error": "validation_failed",
  "message": "Request validation failed",
  "details": {
    "content": "Content is required",
    "antiSpamAnswer": "Anti-spam answer is required when notifications are enabled"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "path": "/api/v1/messages"
}
```

### Common Error Codes

- `validation_failed` (400) - Invalid request data
- `message_not_found` (404) - Message doesn't exist or expired
- `invalid_passphrase` (401) - Wrong passphrase provided
- `message_consumed` (410) - Message already accessed
- `rate_limit_exceeded` (429) - Too many requests

## Security Considerations

- **One-time access**: Messages are deleted after successful decryption
- **Encryption**: All content is encrypted before storage
- **Expiration**: Messages expire automatically
- **Rate limiting**: Prevents abuse and DoS attacks
- **No authentication**: Public service, use passphrases for sensitive data

## Interactive Documentation

Visit `/api/v1/docs` for interactive Swagger documentation where you can test API calls directly in your browser.

## Support

- **Issues**: https://github.com/Anthony-Bible/password-exchange/issues
- **Documentation**: Available at the `/api/v1/docs` endpoint
- **OpenAPI Spec**: Available at `/api/v1/docs/swagger.json`