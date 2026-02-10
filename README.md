# Password Vault Backend (Fiber + SQLite)

## Overview
REST API for a password vault using Go Fiber and SQLite. Passwords are encrypted at rest using AES-GCM. User passwords are hashed with bcrypt.

## Setup

### 1. Database (SQLite)
SQLite is **already integrated**. It automatically:
- Creates `./data/vault.db` on first run
- Runs migrations from `migrations/001_init.sql`
- Creates `users` and `vault_entries` tables

No additional setup needed—it's file-based and starts automatically.

### 2. Environment Variables
Copy `.env` file values to your shell before running, or source the file:
```bash
source .env
```

Key variables:
- **JWT_SECRET**: Random string for signing JWT tokens (change in production)
- **VAULT_ENC_KEY**: Base64-encoded 32-byte encryption key (use `openssl rand -base64 32` to generate)
- **PORT**: Server port (default `:8080`)
- **DB_PATH**: SQLite database file path (default `./data/vault.db`)
- **TOKEN_TTL_MIN**: JWT token lifetime in minutes (default `60`)
- **WORKER_POOL_SIZE**: Max concurrent workers for API handlers (default `8`)

### 3. Generate Encryption Key (Production)
```bash
openssl rand -base64 32
```
Copy output and set as `VAULT_ENC_KEY` in `.env`

## Run
```bash
source .env
go mod tidy
go run .
```

## Endpoints
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login (returns JWT token)
- `GET /api/vault/entries` - List all vault entries (auth required)
- `POST /api/vault/entries` - Create entry (auth required)
- `GET /api/vault/entries/:id` - Get decrypted password (auth required)
- `PUT /api/vault/entries/:id` - Update entry (auth required)
- `DELETE /api/vault/entries/:id` - Delete entry (auth required)
- `GET /api/vault/search?q=gmail` - Search by website/URL/username (auth required)

## Sample API Calls

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"mypassword123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"mypassword123"}'
```

### Create Entry (replace TOKEN)
```bash
curl -X POST http://localhost:8080/api/vault/entries \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer TOKEN" \
    -d '{"title":"Gmail","username":"user@gmail.com","password":"mypass123","url":"https://mail.google.com","category":"personal"}'
```

### List Entries
```bash
curl http://localhost:8080/api/vault/entries \
    -H "Authorization: Bearer TOKEN"
```

### Get Entry by ID
```bash
curl http://localhost:8080/api/vault/entries/1 \
    -H "Authorization: Bearer TOKEN"
```

### Search by Website/URL/Username
```bash
curl "http://localhost:8080/api/vault/search?q=gmail" \
    -H "Authorization: Bearer TOKEN"
```

## Go Concepts Implemented

### 1. **Goroutines & Concurrency**
- Fiber automatically spawns goroutines per HTTP request
- **AuditService**: Background worker goroutine processes audit events from a buffered channel
- **WorkerPool**: Concurrent job processing pattern for batch operations

### 2. **Channels**
- **Buffered channel** (`eventChan`) for audit event queue
- Non-blocking event logging with `select` statement
- Worker pattern for job distribution

### 3. **Custom Error Types**
- `VaultError` struct with code, message, and wrapped error
- Centralized error handling in `internal/errors/errors.go`

### 4. **Context**
- Request context passed through search operations
- Graceful shutdown with `context.WithTimeout`
- Context-aware operation cancellation

### 5. **Sync Package**
- `sync.WaitGroup` for coordinating goroutine lifecycle
- Thread-safe background task management

### 6. **Defer & Signal Handling**
- Graceful shutdown on SIGINT/SIGTERM
- Proper resource cleanup during shutdown
- 5-second timeout for audit service shutdown

### 7. **Repository Pattern**
- Interface-based design for database operations
- Separation of concerns (data, business logic, handlers)

### 8. **Interfaces**
- Extensible service layers
- Type-safe implementations

## Architecture Diagram
```
HTTP Request
    ↓
Fiber Handler (runs in goroutine)
    ↓
Service Layer (business logic + audit logging)
    ↓
Audit Service (background worker) ← Channel ← Non-blocking LogEvent()
    ↓
Repository Layer (database)
    ↓
SQLite
```

## Notes
- List endpoint omits decrypted passwords for security
- Get endpoint returns the full decrypted password
- All passwords encrypted with AES-GCM before storage
- Audit logging happens in background without blocking responses
- Search supports wildcard queries on title, URL, and username
