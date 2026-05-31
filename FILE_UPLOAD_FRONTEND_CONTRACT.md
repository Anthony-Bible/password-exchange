# File Upload Frontend Contract

This document is the authoritative spec for building a JavaScript client that interoperates with the encrypted chunked file upload/download endpoints. All citations were verified against source before being written here.

---

## 1. Route Registration

`WithFileHandler(...)` is conditionally called during server construction (`app/cmd/web/forms.go:138`):

```go
webServer := webAdapter.NewWebServer(messageService, encryptionClient, storageClient).WithFileHandler(fileHandler)
```

`fileHandler` is only non-nil when all four object storage environment variables are set (`ObjectStorageEndpoint`, `ObjectStorageAccessKey`, `ObjectStorageSecretKey`, `ObjectStorageBucket`). When any is absent the route guard at `server.go:134` (`if s.fileHandler != nil`) silently skips all three file routes — they return 404 with no distinguishing log message.

```go
if s.fileHandler != nil {
    files := v1.Group("/files")
    files.POST("/initiate", middleware.FileInitiateRateLimit(), s.fileHandler.InitiateUpload)
    files.POST("/:fileID/chunks", middleware.FileUploadRateLimit(), s.fileHandler.UploadChunk)
    files.GET("/:fileID", middleware.MessageAccessRateLimit(), s.fileHandler.DownloadFile)
}
```

Verify all four S3 env vars are set in every target environment before testing these routes.

---

## 2. Endpoint Table

All endpoints sit under `/api/v1/files`. The `/api` group applies CORS (`Access-Control-Allow-Origin: *`), `X-Correlation-ID` tracking, and panic-recovery middleware to every request. The server uses `gin.New()` + `RedactingLogger` instead of `gin.Default()` (`server.go:42–43`). The `?key=` query parameter has been removed from the download endpoint — the key travels only via the `X-File-Key` header (`file_handlers.go:131`).

---

### 2.1 Initiate Upload

| Property | Value |
|---|---|
| Method | `POST` |
| Path | `/api/v1/files/initiate` |
| Rate limit | `FileInitiateRateLimit` (30 req/hr per IP) |
| Request Content-Type | `application/json` |

**Request body** (`file_handlers.go:36–42`):
```json
{
  "filename":    "report.pdf",
  "contentType": "application/pdf",
  "totalSize":   1048576,
  "chunkSize":   262144,
  "messageID":   "<parent-message-uuid>"
}
```

All fields except `contentType` are required. `filename` and `messageID` must be non-empty. `totalSize` and `chunkSize` must be `> 0`. `totalSize` must not exceed the configured maximum (default 100 MiB = `104857600` bytes; `forms.go:27`).

**Response — 201 Created** (`file_handlers.go:74–75`):
```json
{
  "fileID":     "<uuid-string>",
  "sessionID":  "<uuid-string>",
  "encodedKey": "<base64url-with-padding>"
}
```

- `fileID`: permanent object-storage key. Used in all chunk upload paths and the download URL.
- `sessionID`: resumable upload token. Required in every chunk upload request.
- `encodedKey`: the 32-byte AES-256 encryption key encoded with `base64.URLEncoding` (base64url **with** `=` padding, `-`/`_` alphabet). This is the **only time the server exposes the key** — it is cleared from server state when the upload completes (see Section 5). Store it immediately and treat it as a secret; it must never appear in a server-visible URL.

**Error responses**:

| Condition | HTTP | `error` code |
|---|---|---|
| Malformed JSON | 400 | `validation_failed` |
| Missing/invalid fields | 400 | `validation_failed` |
| `totalSize` > max | 400 | `validation_failed` |
| Object storage failure | 500 | `internal_error` |

---

### 2.2 Upload Chunk

| Property | Value |
|---|---|
| Method | `POST` |
| Path | `/api/v1/files/:fileID/chunks` |
| Rate limit | `FileUploadRateLimit` (500 req/hr per IP) |
| Request Content-Type | `multipart/form-data` |

**Multipart form fields** (`file_handlers.go:79–113`):

| Field | Type | Required | Notes |
|---|---|---|---|
| `sessionID` | string | yes | Value from `InitiateUpload` response |
| `chunkIndex` | string (integer) | yes | **One-based.** Must be `>= 1` and `<= totalChunks` |
| `totalChunks` | string (integer) | yes | Must equal `ceil(totalSize / chunkSize)` — see Section 6 |
| `data` | file part | yes | Raw **plaintext** chunk bytes as a file attachment |

`:fileID` in the URL path must match the `fileID` from `InitiateUpload`. Chunk bytes are plaintext — the server performs AES-256-GCM encryption before storing (see Section 4).

**Response — 200 OK**:
```json
{
  "fileID":     "<uuid>",
  "chunkIndex": 3,
  "done":       false
}
```

`done` is `true` on the final chunk. At that point the server calls `CompleteMultipartUpload`, sets `session.Status = "complete"`, and **clears the stored encryption key** from session state (see Section 5).

**Error responses**:

| Condition | HTTP | `error` code |
|---|---|---|
| Missing/invalid form fields | 400 | `validation_failed` |
| `chunkIndex <= 0` or `totalChunks` mismatch | 400 | `validation_failed` |
| Session not found | 404 | `message_not_found` |
| Session already complete | 400 | `validation_failed` |
| Session expired (> 24 h) | 410 | `session_expired` |
| Encryption or storage failure | 500 | `internal_error` |

---

### 2.3 Download File

| Property | Value |
|---|---|
| Method | `GET` |
| Path | `/api/v1/files/:fileID` |
| Rate limit | `MessageAccessRateLimit` |

**Key transport — `X-File-Key` request header only** (`file_handlers.go:123–144`):

```
GET /api/v1/files/<fileID>
X-File-Key: <encodedKey verbatim from InitiateUpload response>
```

The `X-File-Key` header is the sole key transport mechanism. The `?key=` query parameter was removed (`file_handlers.go:131`) — a request without `X-File-Key` returns 400 regardless of query params. `X-File-Key` is included in the CORS `Access-Control-Allow-Headers` list (`web/server.go:125`), so cross-origin `fetch()` calls work without preflight rejection.

**IMPORTANT — use `fetch()` + Blob, not `<a href>` or browser navigation.** Browsers cannot carry custom headers on top-level navigations or anchor clicks. A direct `href` to this endpoint omits `X-File-Key` and returns 400. See Section 5.3 for the complete JS flow.

**Response — 200 OK**:
- `Content-Type`: the `contentType` supplied at initiation
- `Content-Disposition`: `attachment; filename="<sanitized-filename>"`
- Body: raw decrypted file bytes

**Error responses**:

| Condition | HTTP | `error` code |
|---|---|---|
| `X-File-Key` header absent | 400 | `validation_failed` |
| Key value not valid base64url | 400 | `validation_failed` |
| `fileID` not found | 404 | `message_not_found` |
| Wrong key / tampered ciphertext | 500 | `internal_error` |

---

## 3. Error Response Shape

All errors follow the RFC 7807-inspired envelope (`models/error.go`):

```json
{
  "error":     "validation_failed",
  "message":   "Human-readable description",
  "details":   { "field": "value" },
  "timestamp": "2026-05-30T12:00:00Z",
  "path":      "/api/v1/files/initiate"
}
```

`details` is omitted when null. Known `error` codes: `validation_failed`, `message_not_found`, `session_expired`, `internal_error`, `rate_limit_exceeded`.

---

## 4. Encryption — Server-Side Only

The client does not encrypt chunks. All AES-GCM operations happen on the server. The client sends plaintext bytes and receives the encryption key in the `InitiateUpload` response for use at download time only.

### 4.1 Algorithm

| Parameter | Value | Source |
|---|---|---|
| Algorithm | AES-256-GCM | `file_encryption.go:12–13` |
| Key length | 32 bytes | `file_encryption.go:25` |
| Nonce size | 12 bytes (standard GCM) | `file_encryption.go:75` |
| Tag length | 16 bytes (standard GCM) | `cipher.NewGCM` default |
| Nonce source | `crypto/rand` per chunk | `file_encryption.go:76` |

### 4.2 Per-Chunk Wire Format

Each chunk is stored as a length-prefixed frame (`file_encryption.go:64–87`):

```
┌──────────────────┬──────────────┬────────────────────────────┐
│  uint32 BE (4 B) │  nonce (12 B)│  ciphertext + tag (N+16 B) │
│  = 12 + N + 16   │              │                            │
└──────────────────┴──────────────┴────────────────────────────┘
```

- Bytes 0–3: big-endian `uint32` = payload length that follows (`12 + plaintext_len + 16`)
- Bytes 4–15: 12-byte random nonce
- Bytes 16–end: AES-GCM ciphertext followed immediately by the 16-byte authentication tag

Produced by `gcm.Seal(nonce, nonce, data, aad)` — first argument is the output prefix, second is the nonce — yielding `[nonce ‖ ciphertext ‖ tag]`. All parts are stored contiguously in object storage; the server reads the full blob on download and parses frames sequentially.

### 4.3 AAD Construction (Per Chunk)

Additional authenticated data binds each chunk to its one-based position and owning file (`file_encryption.go:167–173`):

```
[ uint32 BE chunkIndex (1-based) ][ uint32 BE totalChunks ][ fileID bytes (UTF-8) ]
  bytes 0–3                         bytes 4–7                bytes 8+
```

```go
aad := make([]byte, 8, 8+len(fileID))
binary.BigEndian.PutUint32(aad[0:4], uint32(chunkIndex))   // 1-based
binary.BigEndian.PutUint32(aad[4:8], uint32(totalChunks))
aad = append(aad, fileID...)
```

On decryption the server reconstructs the identical AAD for each frame using the frame's loop position (`i+1`), session `TotalChunks`, and session `FileID`. Reordering, truncating, relocating, or tampering any frame fails GCM authentication.

### 4.4 Web Crypto Compatibility

Client-side encryption is not required. For a future client-side mode this framing is compatible with `SubtleCrypto.encrypt({name:"AES-GCM", iv: nonce, additionalData: aad, tagLength: 128}, key, data)`. The 4-byte length prefix and nonce-prepend must be assembled manually with `DataView`/`Uint8Array`.

---

## 5. Key Handling and Security

### 5.1 Key Lifetime — Generated Once, Cleared on Completion

The server generates a 32-byte AES-256 key during `InitiateUpload` (`file_service.go:60–63`, `file_service.go:97`) and returns it **once** in the `201` response as `encodedKey`.

The key is stored in the upload session only for the duration of the active upload, so each incoming chunk can be encrypted server-side. When the final chunk completes the upload, `CompleteSession` is called and the stored key is immediately cleared:

- **Memory adapter** (`memory/upload_state.go:91`): `session.EncryptionKey = nil`
- **Database service (SQL)** (`storage/adapters/secondary/mysql/upload_session.go`): on completion, `UPDATE file_upload_sessions SET status = ..., encryption_key = NULL WHERE session_id = ...`

After completion the server holds **no copy of the key**. The download endpoint relies entirely on the key the client supplies via `X-File-Key`. If the client loses the key, the file cannot be decrypted.

### 5.2 Share-URL Format — Fragment `#k=`

The key travels to the recipient via a **URL fragment**. Fragments are never sent in HTTP requests, so the key never reaches the server, never appears in proxy logs, and never leaks through `Referer` headers.

**Share URL format** (`file-upload.js:184–186`):
```
https://<host>/files/<fileID>#k=<encodedKey>
```

- Fragment parameter name: **`k`** (`file-upload.js:186`: `return base + encodeURIComponent(fileID) + '#k=' + encodedKey`)
- `<encodedKey>`: the `encodedKey` value from `InitiateUpload`, verbatim
- The `#k=...` portion is stripped by the browser before any HTTP request

**Relationship to the existing message decrypt-URL convention:**

The message domain (`url/builder.go:24–25`) uses `{baseURL}decrypt/{messageID}/{base64url-key}` — key as a path segment, parsed server-side via `c.Param("key")`, read client-side from `window.location.pathname`. The file design uses a URL fragment instead: the key never reaches the server. The encoding (`base64.URLEncoding` with `=` padding) is identical in both.

### 5.3 Recipient Download Flow — `file-download.html`

The share URL resolves to `file-download.html`, whose JS reads the fragment and drives the download via `fetch()`. It must not be a direct link to the API endpoint — browsers cannot carry `X-File-Key` on navigation.

**Step 1 — read fragment on page load** (`file-upload.js:199–210`):
```js
const pathParts = window.location.pathname.replace(/\/$/, '').split('/');
const fileID = pathParts[pathParts.length - 1];
const hash = window.location.hash.slice(1); // strip '#'
const params = new URLSearchParams(hash);
const encodedKey = params.get('k');          // param name is 'k'
if (!fileID || !encodedKey) { /* show error */ }
```

**Step 2 — fetch with `X-File-Key` header** (`file-upload.js:223+`):
```js
const response = await fetch(`/api/v1/files/${fileID}`, {
  headers: { 'X-File-Key': encodedKey }
});
if (!response.ok) {
  const err = await response.json(); // Section 3 error shape
}
```

**Step 3 — Blob download using `Content-Disposition` filename**:
```js
const blob = await response.blob();
const cd = response.headers.get('Content-Disposition') ?? '';
const filename = cd.match(/filename="?([^"]+)"?/)?.[1] ?? 'download';
const url = URL.createObjectURL(blob);
const a = document.createElement('a');
a.href = url;
a.download = filename;
a.click();
URL.revokeObjectURL(url);
```

### 5.4 Key Encoding Reference

`base64.URLEncoding` (Go): base64url with `=` padding, `-`/`_` alphabet. JavaScript's `btoa`/`atob` use `+`/`/` — do not use them on `encodedKey`. Pass `encodedKey` from the server verbatim as the `X-File-Key` header value and in the `#k=` fragment; no re-encoding is needed.

### 5.5 QA Verification — Key Must Not Appear in Any Log or URL

Verify in staging that `encodedKey` never appears in:

- **Server access logs**: `X-File-Key` is not logged by the access logger. The `?key=` query param no longer exists on the download endpoint so there is nothing to redact there. Confirm no other middleware logs request headers containing the key value.
- **Application structured logs (slog)**: grep log output for any substring of `encodedKey` during a full upload + download cycle.
- **Browser network inspector URL**: the `GET /api/v1/files/:fileID` request URL must contain no `key=` parameter.
- **Referer headers**: navigate away from the download page; confirm `Referer` on the next request omits the key (fragments are excluded from `Referer` by spec — verify empirically).
- **Server state after completion**: query the session table after `done: true`; `encryption_key` must be NULL (MySQL) or nil (memory). The server must not be able to reconstruct the key from state.

---

## 6. Chunk Size, File Size, and Session Rules

| Parameter | Value | Source |
|---|---|---|
| Default max file size | 100 MiB (104,857,600 bytes) | `forms.go:27` |
| Configurable max | `FileUploadMaxSize` config field | `forms.go:114` |
| Session TTL | 24 hours from creation | `file_service.go:16` |
| Session cleanup interval | 1 hour background goroutine | `forms.go:128` |
| `chunkIndex` | One-based (1 = first chunk) | `file_handlers.go:83`, `file_service.go:297` |
| `totalChunks` | `ceil(totalSize / chunkSize)` — integer ceiling | `file_service.go:70` |
| Chunk size | Client-chosen; no per-chunk byte-count enforcement | last chunk may be smaller |

### Session Persistence

**Durable — database service over gRPC** (`forms.go:161` `buildUploadStateAdapter` → `storageClient.NewUploadStateAdapter()`): when a database-service StorageClient is configured, upload-session state is persisted via the database service over gRPC (`storage` domain: `adapters/primary/grpc/upload_session.go` server, `adapters/secondary/mysql/upload_session.go` SQL). Durable across web-service restarts and visible to all replicas. Lookup by either `session_id` OR `file_id`. The `encryption_key` is stored only while the session is active and is cleared (set to `NULL`) on completion — after completion the server holds no copy of the key, and download relies solely on the client-supplied `X-File-Key`. The web service holds NO direct MySQL connection or `go-sql-driver` dependency; all DB access is centralized in the database service.

**Fallback — in-memory** (`MemoryUploadStateAdapter`, `forms.go:164`): used when no StorageClient is configured. Single-replica; in-progress sessions lost on pod restart (completed files already assembled in object storage are unaffected). The in-memory adapter mirrors the same key-cleared-on-completion semantics.

### `totalChunks` formula

Server: `(totalSize + chunkSize - 1) / chunkSize` (integer division; `file_service.go:70`). JS: `Math.ceil(totalSize / chunkSize)`. Every `UploadChunk` call must send this exact value or the server returns `ErrInvalidChunkIndex`.

### Chunk retry

Uploading the same `chunkIndex` twice replaces the existing part entry (`memory/upload_state.go:63–68`). Safe to retry any chunk.

### Session expiry enforcement

`UploadChunk` checks `time.Now().After(session.ExpiresAt)` on every call (`file_service.go:124`). Returns HTTP 410 `session_expired`. The hourly background goroutine also aborts the S3 multipart upload for expired incomplete sessions.

---

## 7. Upload + Download Flow Summary

```
Client                                      Server
  │                                            │
  │  POST /api/v1/files/initiate               │
  │  {filename, contentType, totalSize,        │
  │   chunkSize, messageID}                    │
  │───────────────────────────────────────────>│ generates fileID, sessionID, 32-byte key
  │                                            │ stores key in session (active only)
  │                                            │ starts S3 multipart upload
  │  201 {fileID, sessionID, encodedKey}       │
  │<───────────────────────────────────────────│
  │  (store encodedKey — only copy that exists)│
  │                                            │
  │  for each chunk i = 1 .. totalChunks:      │
  │  POST /api/v1/files/:fileID/chunks         │
  │  multipart: sessionID, chunkIndex=i,       │
  │  totalChunks=N, data=<plaintext bytes>     │
  │───────────────────────────────────────────>│ AES-256-GCM encrypts chunk
  │                                            │ stores [uint32-BE len][nonce][ct+tag]
  │                                            │   as S3 multipart part i
  │  200 {fileID, chunkIndex, done}            │
  │<───────────────────────────────────────────│ done=true on final chunk:
  │                                            │   → CompleteMultipartUpload
  │                                            │   → session.Status = "complete"
  │                                            │   → encryption_key CLEARED (NULL/nil)
  │                                            │
  │  construct share URL:                      │
  │  https://<host>/files/<fileID>#k=<encodedKey>
  │                                            │
  │  recipient opens share URL                 │
  │  file-download.html loads                  │
  │  JS reads hash: params.get('k')            │
  │                                            │
  │  fetch GET /api/v1/files/:fileID           │
  │  X-File-Key: <encodedKey>                  │
  │───────────────────────────────────────────>│ reads all S3 parts
  │                                            │ parses frames, AES-256-GCM decrypts
  │  200 Content-Disposition: attachment       │   using client-supplied key only
  │<───────────────────────────────────────────│ raw decrypted file bytes
  │                                            │
  │  blob → createObjectURL → <a>.click()      │
```

---

## 8. Notes for Frontend-Dev and QA

1. **Response field is `encodedKey`.** Read `response.encodedKey` from the 201 JSON. (`file_handlers.go:75`)

2. **Fragment param is `#k=`.** Share URL: `...#k=<encodedKey>`. Read with `params.get('k')`. (`file-upload.js:186`, `file-upload.js:205`)

3. **Download uses `X-File-Key` header only.** The `?key=` query parameter was removed. (`file_handlers.go:131, 136–138`)

4. **`X-File-Key` is pre-authorized for CORS.** Cross-origin `fetch()` with this header works without preflight issues. (`web/server.go:125`)

5. **Key is cleared on completion — client is the only holder after upload finishes.** If the client loses `encodedKey`, the file cannot be decrypted. (`memory/upload_state.go:91`, and the database-service SQL path `storage/adapters/secondary/mysql/upload_session.go`)

6. **`chunkIndex` is one-based.** First chunk is `1`. Sending `0` returns 400.

7. **`totalChunks` must match exactly.** Use `Math.ceil(totalSize / chunkSize)`. Any mismatch is rejected.

8. **Last chunk may be smaller than `chunkSize`.** No padding needed.

9. **`contentType` is echoed verbatim.** Use a valid MIME type; `application/octet-stream` is a safe fallback.

10. **Pass `encodedKey` verbatim — no re-encoding.** It is already base64url with `=` padding. Use it directly as the `X-File-Key` value and in the `#k=` fragment.

11. **The download page route (`/files/:fileID`) must be registered in the web server.** `file-download.html` needs a web page route — analogous to `/decrypt/:uuid/*key`. This is not a REST route.

12. **File routes return 404 silently without S3 config.** Verify all four env vars in every environment.

13. **Session state is durable via the database service (gRPC) when configured**, falling back to in-memory (single-replica, lost on pod restart) when no database-service StorageClient is set. The web service holds no direct MySQL connection. The encryption key is cleared on completion in both adapters.

14. **QA: key must never appear in any server log or request URL.** See Section 5.5 checklist. Specifically verify `encryption_key` is NULL in the session table after upload completes.
