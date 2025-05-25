# Password Exchange REST API

Secure password sharing service with REST API endpoints.

## Documentation

📖 **Complete API Documentation**: Visit `/api/v1/docs/index.html` for interactive Swagger documentation

🔗 **Live Documentation**: 
- Production: https://password.exchange/api/v1/docs/index.html
- Development: http://localhost:8080/api/v1/docs/index.html

## Quick Start

```bash
# Submit a message
curl -X POST https://password.exchange/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"content": "my secret"}'

# Check message status  
curl "https://password.exchange/api/v1/messages/{id}?key={key}"

# Decrypt message
curl -X POST https://password.exchange/api/v1/messages/{id}/decrypt \
  -H "Content-Type: application/json" \
  -d '{"decryptionKey": "{key}"}'
```

## Migration from Old API

The old form-based API (`api=true` parameter) has been replaced with proper REST endpoints:

- **Old**: `POST /` with form data and `api=true`
- **New**: `POST /api/v1/messages` with JSON

See the interactive documentation for complete migration details.

## Resources

- **OpenAPI Spec**: `/api/v1/docs/doc.json`
- **Health Check**: `/api/v1/health` 
- **API Info**: `/api/v1/info`