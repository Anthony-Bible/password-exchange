# Encrypted Chunked File Upload

This document describes how chunked file uploads are encrypted, stored, and
decrypted, and the security properties (and operational caveats) of the design.

## Overview

Files are uploaded in chunks, each sealed locally with **AES-256-GCM** before
being stored as a part of an S3-compatible multipart upload. Each chunk binds
its file ID, one-based index, and total chunk count as **additional
authenticated data (AAD)**, so reordered, truncated, relocated, or tampered
ciphertext fails authentication on decrypt.

Encryption and decryption happen **in-process** in the web/file service. The
shared encryption gRPC service is used only as an ID/key oracle (`GenerateID`,
`GenerateKey`) — chunk bytes never round-trip to it.

Key source files:

- `app/internal/domains/message/adapters/secondary/crypto/file_encryption.go` — local AES-256-GCM seal/open + framing
- `app/internal/domains/message/domain/file_service.go` — upload/download/abort orchestration
- `app/internal/domains/message/adapters/secondary/s3/object_storage.go` — S3 multipart adapter
- `app/internal/domains/message/adapters/secondary/memory/upload_state.go` — in-memory session/state
- `app/internal/domains/message/adapters/primary/api/file_handlers.go` — REST endpoints

## Upload — sealing the vault

```mermaid
sequenceDiagram
    autonumber
    actor C as Client
    participant W as Web / FileService
    participant E as EncryptionSvc (gRPC)
    participant S as ObjectStore (S3)

    Note over C,S: Initiate
    C->>W: POST /files/initiate {filename, size, chunkSize}
    W->>E: GenerateID() ×2  (fileID, sessionID)
    W->>E: GenerateKey(32)
    E-->>W: 32-byte AES-256 key
    W->>S: InitiateMultipartUpload(fileID)
    S-->>W: uploadID
    W->>W: store session {key, fileID, totalChunks…} (in-memory)
    W-->>C: {fileID, sessionID}

    Note over C,S: Per chunk (repeat)
    loop each chunk
        C->>W: POST /:fileID/chunks {sessionID, idx, total, PLAINTEXT}
        rect rgb(235, 245, 255)
        Note right of W: EncryptChunk — LOCAL, in-process
        W->>W: nonce ← 12 random bytes (crypto/rand)
        W->>W: aad ← idx‖total‖fileID
        W->>W: sealed ← AES-256-GCM.Seal(nonce, data, aad)
        W->>W: frame ← [uint32 len][nonce][ct+tag]
        end
        W->>S: UploadPart(idx, frame)
        S-->>W: ETag
        W-->>C: {done: false}
    end

    Note over C,S: Finalize
    W->>S: CompleteMultipartUpload(ordered parts)
    W-->>C: {done: true}
```

## Download — cracking it open

```mermaid
sequenceDiagram
    autonumber
    actor R as Recipient
    participant W as Web / FileService
    participant S as ObjectStore (S3)
    actor R as Recipient
    participant W as Web / FileService
    participant S as ObjectStore (S3)

    R->>W: GET /files/:fileID#key=<base64url>
    Note right of R: Browser JS calls GET /api/v1/files/:fileID with X-File-Key header
    S-->>W: concatenated frames
    rect rgb(235, 245, 255)
    Note right of W: DecryptFile — LOCAL
    W->>W: parseFrames(data)
    W->>W: assert len(frames) == TotalChunks  (truncation guard)
    loop each frame i
        W->>W: aad ← (i+1)‖total‖fileID
        W->>W: pt ← AES-256-GCM.Open(nonce, ct, aad)
        Note right of W: any auth failure ⇒ whole download aborts
    end
    end
    W-->>R: decrypted file bytes
```

## Frame format + what each field defends

```mermaid
flowchart TB
    subgraph FRAME["One frame = one chunk (frames concatenated → 1 S3 object)"]
        direction LR
        L["uint32 BE<br/>length"] --> N["12-byte<br/>nonce"] --> CT["ciphertext + 16-byte GCM tag"]
    end

    subgraph AAD["AAD bound into every chunk (authenticated, not encrypted)"]
        direction LR
        IDX["chunkIndex<br/>(uint32 BE)"] --- TOT["totalChunks<br/>(uint32 BE)"] --- FID["fileID<br/>(variable)"]
    end

    L -. "len vs remaining bytes,<br/>reject size≤0 → overflow-safe parsing" .-> P[Parser hardening]
    N -. "fresh per chunk → no GCM nonce reuse" .-> CONF[Confidentiality]
    CT -. "tag verify → tamper detection" .-> INT[Per-chunk integrity]
    IDX -. "reorder → AAD mismatch" .-> RO[Anti-reorder]
    TOT -. "drop/append → count + AAD fail" .-> TR[Anti-truncation]
    FID -. "splice A→B → AAD fail" .-> RL[Anti-relocation]
```

The on-disk layout per chunk:

```
 ┌──────────┬──────────────┬───────────────────────────────┐
 │ uint32   │  12-byte     │  ciphertext + 16-byte GCM tag  │
 │ BE length│  nonce       │                                │
 └──────────┴──────────────┴───────────────────────────────┘
 └───────────────── one frame, one chunk ───────────────────┘
   frames are simply concatenated → stored as one S3 object
```

The AAD is constructed as fixed-width `chunkIndex ‖ totalChunks` (each a
big-endian `uint32`) followed by the variable-length `fileID`, so the encoding
is unambiguous and both the encrypt and decrypt sides reconstruct identical
bytes. The AAD is authenticated, not encrypted: tampering with position, count,
or file identity causes GCM tag verification to fail.

## Why it's secure

| Property | Mechanism | What it stops |
|---|---|---|
| **Confidentiality** | AES-256-GCM, 32-byte key | Anyone reading the S3 object (or the storage provider) sees only ciphertext |
| **Per-chunk integrity** | 16-byte GCM auth tag on every frame | Bit-flipping / tampering — `Open` fails, download aborts |
| **Anti-reorder** | `chunkIndex` (1-based) bound as AAD | Swapping chunk 3 and 7 → AAD mismatch → auth failure |
| **Anti-truncation / padding** | `totalChunks` in AAD **+** `len(frames) == TotalChunks` check | Dropping or appending chunks fails both the count check and AAD auth |
| **Anti-relocation** | `fileID` bound as AAD | Splicing a chunk from file A into file B fails auth |
| **Nonce safety** | Fresh 12-byte CSPRNG nonce per chunk (`crypto/rand`) | GCM nonce reuse, which catastrophically breaks GCM |
| **Parser hardening** | Frame length compared against *remaining* bytes; rejects `size <= 0` | `uint32`→`int` overflow on 32-bit platforms and malformed/hostile framing |

## Security model at a glance

```mermaid
flowchart LR
    K["AES-256 key<br/>GenerateKey(32)"]:::good
    K --> SEAL["Seal in-process<br/>(no per-chunk round-trip)"]:::good
    SEAL --> STORE[(S3: ciphertext only)]:::good

    K -. delivered as .-> URL["?key= query param"]:::warn
    K -. copied into .-> MEM["in-memory session<br/>EncryptionKey, 24h TTL (enforced)"]:::warn
    MEM -. single replica .-> REPL["MemoryUploadState<br/>(no multi-pod)"]:::warn
    STORE -. decrypt path .-> ALLOC["append all chunks<br/>→ one in-RAM buffer"]:::warn

    classDef good fill:#dcfce7,stroke:#16a34a,color:#000;
    classDef warn fill:#fef9c3,stroke:#ca8a04,color:#000;
```

Green is the cryptographic core, which is solid. Yellow marks the operational
edges where the "zero-knowledge" promise is softer than it first appears.

## Operational caveats

These are not breaks in the AES-256-GCM core — they are edges in the surrounding
secrecy model worth tracking:

1. **The key is delivered client-side and sent via an HTTP header.** The share link
   carries the key in the URL fragment (`#/...`) so it never reaches the server on
   initial navigation. The browser then calls `GET /api/v1/files/:fileID` with
   `X-File-Key: <base64url>` to decrypt and download the file.

2. **The symmetric key is stored server-side** in the in-memory upload session
   (`EncryptionKey`, `contracts.FileUploadSession`). During an active upload the
   design is not purely zero-knowledge — anyone with web-process memory access
   has the key. Download itself uses the client-supplied key, but the session
   copy still lives in RAM. The 24h TTL (`ExpiresAt`) **is now enforced** for
   incomplete sessions: `UploadChunk` rejects an expired session with
   `ErrUploadSessionExpired` (HTTP 410 Gone), and a background janitor
   (`FileService.RunSessionCleanup`, hourly) sweeps expired incomplete sessions
   — freeing their key from memory and best-effort aborting the dangling S3
   multipart upload.

   The remaining limitation: **completed** sessions are retained in memory (and
   keep their key) indefinitely so finalized files stay downloadable — completed
   S3 objects don't expire today, so there is no lifecycle to evict them against.

3. **`MemoryUploadStateAdapter` is single-replica only.** Sessions vanish on
   restart and are not shared across pods — a blocker for horizontal scaling.
   The expiry janitor above only sweeps the local replica's map.

4. **`DecryptFile` buffers the whole file in memory**, appending all chunks into
   one `[]byte`. `maxFileSize` guards initiation, but the decrypt path trusts
   stored data and allocates proportionally to file size.
