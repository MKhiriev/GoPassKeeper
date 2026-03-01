# GoPassKeeper

**GoPassKeeper** is a client-server password manager with a terminal UI (TUI), local encryption, and client-server synchronization.

The project stores sensitive records (logins/passwords, secure notes, card data, binary metadata) in encrypted form and supports offline-first usage with periodic sync.

## Features

- TUI client based on Bubble Tea (login, register, CRUD, manual sync, quick copy of sensitive values).
- REST API server based on chi (`/api/auth`, `/api/data`, `/api/sync`, `/api/version`).
- Local client storage in SQLite with automatic migrations.
- Server storage in PostgreSQL with automatic migrations.
- Client-side encryption (Argon2id + AES-GCM) and integrity checks (HMAC-SHA256).
- Conflict-safe synchronization by `client_side_id + version + hash + deleted`.
- Soft-delete model to preserve deletion semantics during sync.

## Supported Data Types

- `LoginPassword` (username/password/URIs/TOTP)
- `Text` (secure notes)
- `Binary` (encrypted binary metadata, storage hooks are present)
- `BankCard` (cardholder, PAN, expiry, CVV)

## High-Level Architecture

- `cmd/server`: API server bootstrap.
- `cmd/client`: TUI client bootstrap.
- `internal/handler/http`: REST routes and middleware.
- `internal/service`: business logic (auth, private data, sync).
- `internal/store`: repositories and DB abstractions (PostgreSQL + SQLite).
- `internal/tui`: terminal interface and flows.
- `migrations/`: embedded SQL migrations for PostgreSQL and SQLite.
- `models/`: request/response/data contracts.

## Quick Start

### 1. Prerequisites

- Go `1.26+`
- PostgreSQL `14+` (or compatible)

### 2. Configure and run server

Create a database before running the server (example):

```sql
CREATE DATABASE gopasskeeper;
```

Create a server config file from template:

```bash
cp settings.template.json settings.json
```

Minimal required fields in `settings.json`:

- `storage.db.dsn` (PostgreSQL DSN)
- `app.password_hash_key`
- `app.token_sign_key`
- `app.token_issuer`
- `app.token_duration`
- `app.hash_key`
- `server.http_address`
- `server.request_timeout`

Run server:

```bash
go run ./cmd/server -config ./settings.json
```

### 3. Configure and run client

Create a client config file from template:

```bash
cp client-settings.template.json client-settings.json
```

Important client fields:

- `storage.db.dsn`: local SQLite file path (for example `data.db`)
- `adapter.http_address`: server address (for example `localhost:8080`)
- `adapter.request_timeout`: request timeout
- `workers.sync_interval`: background sync interval
- `app.hash_key`: must match server hash key

Run client:

```bash
go run ./cmd/client -config ./client-settings.json
```

## Configuration Sources

The app supports three configuration sources:

1. Environment variables
2. CLI flags
3. JSON config file (`-c` or `-config`)

Common CLI flags:

- `-a` (server HTTP address)
- `-grpc-address`
- `-d` (database DSN)
- `-f` (binary files directory)
- `-password-hash-key`
- `-token-sign-key`
- `-token-issuer`
- `-token-duration`
- `-request-timeout`
- `-hash-key`
- `-v` / `-version`
- `-c` / `-config`

Environment examples:

- `APP_PASSWORD_HASH_KEY`
- `APP_TOKEN_SIGN_KEY`
- `APP_TOKEN_ISSUER`
- `APP_TOKEN_DURATION`
- `APP_HASH_KEY`
- `STORAGE_DB_DATABASE_URI`
- `SERVER_ADDRESS`
- `SERVER_REQUEST_TIMEOUT`
- `ADAPTER_ADDRESS`
- `ADAPTER_REQUEST_TIMEOUT`
- `WORKERS_SYNC_INTERVAL`

## HTTP API Overview

Public endpoints:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/params`
- `GET /api/version/`

Protected endpoints (JWT):

- `POST /api/data/`
- `GET /api/data/all`
- `POST /api/data/download`
- `PUT /api/data/update`
- `DELETE /api/data/delete`
- `GET /api/sync/`
- `GET /api/sync/specific`
- `POST /api/auth/settings/password/change`
- `POST /api/auth/settings/otp`
- `DELETE /api/auth/settings/otp`

## Sync Model

The sync planner compares server and client item states and produces actions:

- `Download`
- `Upload`
- `Update`
- `DeleteClient`
- `DeleteServer`

Decision inputs:

- `client_side_id` (stable identity)
- `version` (optimistic locking)
- `hash` (payload integrity/change detection)
- `deleted` (soft-delete marker)

Detailed matrices and pseudo-code are available in [docs/sync algorithm.md](docs/sync%20algorithm.md).

## Development

Run tests:

```bash
go test ./...
```

Build binaries:

```bash
go build -o ./bin/gopass-server ./cmd/server
go build -o ./bin/gopass-client ./cmd/client
```

Optional build metadata:

```bash
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.buildCommit=$(git rev-parse --short HEAD)" -o ./bin/gopass-server ./cmd/server
go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.buildCommit=$(git rev-parse --short HEAD)" -o ./bin/gopass-client ./cmd/client
```

## Documentation

- [docs/summary.md](docs/summary.md)
- [docs/sync algorithm.md](docs/sync%20algorithm.md)

## Attribution / Credits

This project is licensed under Apache License 2.0. If you redistribute this project or derivative works, keep the `LICENSE` and `NOTICE` files and preserve attribution to the original author and repository.

- Author: Rasul Khiriev (MKhiriev)
- Repository: https://github.com/MKhiriev/GoPassKeeper
- Citation metadata: `CITATION.cff`

## License and Citation

- License: Apache License 2.0 (`LICENSE`)
- Notices: `NOTICE`
- Citation format: `CITATION.cff`
