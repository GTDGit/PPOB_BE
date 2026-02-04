# PPOB.ID Backend

Backend API service untuk aplikasi PPOB (Payment Point Online Banking) dan pembayaran digital.

## Features

- **Authentication**: OTP via WhatsApp/SMS, PIN verification, Multi-device support
- **Prepaid Transactions**: Pulsa, Paket Data, Token PLN, Voucher Game
- **Postpaid Payments**: PLN Pascabayar, PDAM, BPJS, Telkom
- **Bank Transfers**: Transfer antar bank, e-wallet
- **Deposit**: Bank Transfer, QRIS, Virtual Account, Retail (Alfamart/Indomaret)
- **Transaction History**: Riwayat lengkap dengan filter dan export
- **Notifications**: Push notification dan in-app notifications

## Tech Stack

- **Language**: Go 1.22+
- **Framework**: Gin (HTTP), sqlx (Database)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **External APIs**: Gerbang (Payment Gateway), WhatsApp Business API, Fazpass (SMS OTP), Brevo (Email)

## Quick Start

### Prerequisites

- Go 1.22 or later
- PostgreSQL 15 or later
- Redis 7 or later

### Installation

1. Clone repository:
   ```bash
   git clone https://github.com/GTDGit/PPOB_BE.git
   cd PPOB_BE
   ```

2. Copy environment file:
   ```bash
   cp .env.example .env
   ```

3. Configure `.env` with your settings (see Configuration section)

4. Run database migrations:
   ```bash
   # Using golang-migrate
   migrate -path migrations -database "postgres://user:pass@localhost/ppob?sslmode=disable" up
   ```

5. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

### Using Docker

See [README.Docker.md](README.Docker.md) for Docker deployment.

```bash
docker compose up -d
```

## Configuration

See [.env.example](.env.example) for all available configuration options.

### Required Environment Variables

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | Secret key for JWT signing (required) |
| `DB_PASSWORD` | PostgreSQL password (required in production) |
| `DB_HOST` | PostgreSQL host |
| `REDIS_HOST` | Redis host |

### External Services

Configure at least one OTP provider:
- **WhatsApp**: `WA_ACCESS_TOKEN`, `WA_PHONE_NUMBER_ID`
- **SMS (Fazpass)**: `FAZPASS_MERCHANT_KEY`, `FAZPASS_GATEWAY_KEY`

For payment gateway:
- `GERBANG_CLIENT_ID`, `GERBANG_CLIENT_SECRET`

For email:
- `BREVO_API_KEY`

## API Documentation

API documentation is available in [docs/api-v1/](docs/api-v1/):

| File | Description |
|------|-------------|
| [auth.md](docs/api-v1/auth.md) | Authentication endpoints |
| [deposit.md](docs/api-v1/deposit.md) | Deposit operations |
| [history.md](docs/api-v1/history.md) | Transaction history |
| [transaction.md](docs/api-v1/transaction.md) | Transaction flows |
| [user.md](docs/api-v1/user.md) | User profile |

## Health Check

```bash
curl http://localhost:8080/health
# {"status":"ok","timestamp":"2026-02-01T10:00:00+07:00"}
```

## Project Structure

```
.
├── cmd/api/           # Application entrypoint
├── internal/
│   ├── config/        # Configuration
│   ├── domain/        # Domain models & errors
│   ├── handler/       # HTTP handlers
│   ├── middleware/    # Gin middlewares
│   ├── repository/    # Data access layer
│   ├── service/       # Business logic
│   ├── external/      # External API clients
│   └── job/           # Background jobs
├── pkg/               # Shared packages
├── migrations/        # Database migrations
└── docs/              # Documentation
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o ppob-api cmd/api/main.go
```

### Code Quality

```bash
go vet ./...
go build ./...
```

## License

MIT License - see [LICENSE](LICENSE) file.
