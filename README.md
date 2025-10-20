# Notiflow

A minimal email notification API written in Go. It exposes HTTP endpoints to queue and send emails via SMTP, persists message metadata to MongoDB, and provides simple health and metrics endpoints.

- Language: Go 1.25+
- HTTP: chi router with middleware (logging, recovery, compression, timeout)
- Storage: MongoDB (emails collection with schema validation and indexes)
- SMTP: gomail-based sender with configurable SMTP server(s)


## Features
- Send email via POST /api/v1/email with optional CC/BCC and attachments
- Asynchronous delivery; request returns immediately with pending status
- MongoDB persistence with validation, indexes, and 90‑day TTL for cleanup
- Health check at /api/health and simple runtime metrics at /api/metrics
- Configuration via environment variables or YAML file, with .env support
- Ready-to-run via docker compose or Makefile targets


## Quick start

### Using Docker Compose
Prerequisites: Docker Desktop or Docker Engine + Compose plugin

```bash
# From repo root
make docker-dev        # or: docker compose up --build -d

# Tail logs
docker compose logs -f app
```

The API will be available at http://localhost:8080.

Note: The docker-compose stack starts MongoDB and the app. You still need to provide SMTP configuration (see Configuration) to actually deliver emails.

### Run locally (Go)
Prerequisites: Go 1.25+, running MongoDB instance

```bash
# 1) Export environment variables (see Configuration section)
export DB_HOST=localhost
export DB_PORT=27017
export DB_USER=mongo
export DB_PASSWORD=mongo
export DB_NAME=notiflow
# SMTP (example using Gmail SMTP – you likely need an app password)
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587
export SMTP_USERNAME=you@example.com
export SMTP_PASSWORD=app-password
export FROM_EMAIL=you@example.com

# 2) Run
go run ./cmd/notiflow

# or use Makefile
make build && ./build/notiflow
```


## API

Base URL: http://localhost:8080

- GET /api/health
  - 200 OK: {"status":"healthy","service":"notiflow"}

- GET /api/metrics
  - 200 OK: returns application start time and total uptime

- POST /api/v1/email
  - Description: Queues an email for sending. Returns pending status; actual delivery happens asynchronously.
  - Request body (application/json):
    - to: array of email addresses (required, 1–100)
    - cc: array of email addresses (optional, 0–50)
    - bcc: array of email addresses (optional, 0–50)
    - subject: string (required, 1–255)
    - body: string (required, up to 1 MB). Set is_html accordingly.
    - is_html: boolean (default false)
    - attachments: array (optional, max 10). Each attachment:
      - filename: string
      - content: base64-encoded data (JSON maps base64 string to bytes in Go)
      - content_type: string (e.g., "text/plain", "application/pdf")

  - Response 201 Created:
    {
      "id": "<emailId>",
      "status": "pending",
      "message": "Email queued for sending",
      "created_at": "2025-01-01T12:00:00Z"
    }

  - Possible errors:
    - 400 Bad Request: invalid JSON or validation errors
    - 500 Internal Server Error: persistence or SMTP configuration error

### Example requests

Health check:
```bash
curl -s http://localhost:8080/api/health | jq
```

Send a plain-text email:
```bash
curl -s -X POST http://localhost:8080/api/v1/email \
  -H 'Content-Type: application/json' \
  -d '{
    "to": ["recipient@example.com"],
    "subject": "Hello from Notiflow",
    "body": "This is a test.",
    "is_html": false
  }'
```

Send an HTML email with an attachment:
```bash
# Prepare a base64 payload for an attachment
BASE64_CONTENT=$(printf 'Hello file' | base64)

curl -s -X POST http://localhost:8080/api/v1/email \
  -H 'Content-Type: application/json' \
  -d "{
    \"to\": [\"recipient@example.com\"],
    \"subject\": \"Report\",
    \"body\": \"<strong>See attachment</strong>\",
    \"is_html\": true,
    \"attachments\": [
      {
        \"filename\": \"hello.txt\",
        \"content\": \"$BASE64_CONTENT\",
        \"content_type\": \"text/plain\"
      }
    ]
  }"
```


## Data model and constraints
- Collection: emails
- Important constraints (enforced by MongoDB validator):
  - to: 1–100 valid email addresses
  - cc, bcc: up to 50 valid email addresses each
  - subject: 1–255 chars
  - body: up to 1 MB
  - attachments: up to 10 items, each requiring filename, content (binary/base64), content_type
  - status: one of pending | sent | failed
- Indexes: created_at (desc), status, to, text index on subject+body
- TTL: documents expire ~90 days after created_at


## Configuration
Configuration can come from environment variables and/or a YAML file. A config file path can be provided via the flag:

```bash
./notiflow -config.file=config.yaml
```

Notiflow loads environment variables first (including from a .env file if present), then merges values from the YAML file (if found). Logging is configurable (level and format).

Environment variables and defaults:

- General
  - APP_ENV: development | production (default: development)
  - TZ: IANA timezone, e.g., UTC, America/New_York (default: UTC)

- Server
  - HOST: bind host (default: 0.0.0.0)
  - PORT: bind port (default: 8080)

- MongoDB
  - DB_HOST: host (default: localhost)
  - DB_PORT: port (default: 27017)
  - DB_USER: username (default: mongo)
  - DB_PASSWORD: password (default: mongo)
  - DB_NAME: database name (default: notiflow)

- Redis (optional – not currently used by the running code, placeholder for future features)
  - REDIS_HOST (default: localhost)
  - REDIS_PORT (default: 6379)
  - REDIS_USER (default: empty)
  - REDIS_PASS (default: empty)
  - REDIS_DB (default: 0)

- Logging
  - LOG_LEVEL: debug | info | warn | error (default: info)
  - LOG_FORMAT: json | text (default: text)

- SMTP (at least one server required to actually send email)
  - SMTP_HOST (default: smtp.gmail.com)
  - SMTP_PORT (default: 587)
  - SMTP_USERNAME
  - SMTP_PASSWORD
  - FROM_EMAIL: the From header used for messages

You can also provide multiple SMTP servers in config.yaml, for example:

```yaml
smtp_servers:
  - name: primary
    host: smtp.gmail.com
    port: 587
    username: you@example.com
    password: app-password
    from_email: you@example.com
  - name: backup
    host: smtp.mailgun.org
    port: 587
    username: postmaster@sandbox.mailgun.org
    password: secret
    from_email: notifications@example.com
```

If no SMTP servers are configured, POST /api/v1/email will fail with "no SMTP servers configured".


## Development
- Makefile targets:
  - make help
  - make dev (requires air for hot reload)
  - make test, make test-coverage, make test-race
  - make build, make build-release
  - make docker-build, docker-up, docker-down, docker-restart, docker-clean, docker-dev

- Dockerfile builds a minimal image for the app service.
- docker-compose.yml includes MongoDB and maps app port 8080.


## Troubleshooting
- MongoDB connection errors: confirm DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME, and that MongoDB is reachable.
- SMTP errors: verify SMTP_* and FROM_EMAIL; some providers require app passwords or TLS/port specifics.
- Validation errors (400): ensure request JSON matches the schema, email addresses are valid, and limits are respected.
- Logs: set LOG_LEVEL=debug and LOG_FORMAT=json for detailed structured logging.


## License
This project is licensed under the MIT License. See LICENSE for details.
