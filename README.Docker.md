# 🐳 Docker Quick Start Guide

## Prerequisites

- Docker >= 20.10
- Docker Compose >= 2.0

## Quick Start (Development)

### 1. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add your API keys (optional for basic testing)
nano .env
```

### 2. Start Services

```bash
# Start all services (PostgreSQL, Redis, API)
docker-compose up -d

# Or use Makefile
make docker-up
```

### 3. Check Status

```bash
# Check running containers
docker-compose ps

# Check logs
docker-compose logs -f

# Check API health
curl http://localhost:8080/health
```

### 4. Access Services

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | - |
| PostgreSQL | localhost:5432 | User: `ppob_user`, Pass: `ppob_secret`, DB: `ppob_db` |
| Redis | localhost:6379 | No password (development) |

## Common Commands

### Service Management

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# Restart services
docker-compose restart

# View logs
docker-compose logs -f

# View API logs only
docker-compose logs -f ppob_backend
```

### Database Operations

```bash
# Connect to PostgreSQL
docker-compose exec ppob_postgres psql -U ppob_user -d ppob_db

# Backup database
docker-compose exec ppob_postgres pg_dump -U ppob_user ppob_db > backup.sql

# Restore database
cat backup.sql | docker-compose exec -T ppob_postgres psql -U ppob_user -d ppob_db
```

### Redis Operations

```bash
# Connect to Redis CLI
docker-compose exec ppob_redis redis-cli

# Flush cache
docker-compose exec ppob_redis redis-cli FLUSHALL

# Monitor Redis
docker-compose exec ppob_redis redis-cli MONITOR
```

### Cleanup

```bash
# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes (⚠️ deletes data)
docker-compose down -v

# Stop and remove everything including images
docker-compose down -v --rmi all
```

## Production Deployment

### 1. Prepare Environment

```bash
# Copy and configure production environment
cp .env.example .env

# Edit with production values
nano .env
```

**⚠️ Important Production Settings:**

```env
APP_ENV=production
JWT_SECRET=<strong-random-32-char-secret>
DB_PASSWORD=<strong-database-password>
REDIS_PASSWORD=<strong-redis-password>
DB_SSL_MODE=require

# Add your real API keys
WA_ACCESS_TOKEN=<your-whatsapp-token>
FAZPASS_MERCHANT_KEY=<your-fazpass-key>
BREVO_API_KEY=<your-brevo-key>
GERBANG_CLIENT_ID=<your-gerbang-client-id>
GERBANG_CLIENT_SECRET=<your-gerbang-secret>
```

### 2. Start Production Services

```bash
# Build and start with production settings
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Or use Makefile
make docker-prod-up
```

### 3. Monitor Production

```bash
# View logs
docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

# Check container stats
docker stats

# Check health
curl https://your-domain.com/health
```

## Troubleshooting

### Containers won't start

```bash
# Check logs
docker-compose logs

# Check specific service
docker-compose logs ppob_backend
docker-compose logs ppob_postgres
docker-compose logs ppob_redis
```

### Database connection issues

```bash
# Check if PostgreSQL is ready
docker-compose exec ppob_postgres pg_isready -U ppob_user

# Check PostgreSQL logs
docker-compose logs ppob_postgres
```

### API returning 502/503

```bash
# Check if API is running
docker-compose ps ppob_backend

# Check API logs
docker-compose logs ppob_backend

# Restart API
docker-compose restart ppob_backend
```

### Port already in use

```bash
# Check what's using the port
lsof -i :8080
lsof -i :5432
lsof -i :6379

# Or change ports in docker-compose.yml
ports:
  - "8081:8080"  # Change host port to 8081
```

### Reset everything

```bash
# Complete reset (⚠️ deletes all data)
docker-compose down -v --rmi all
docker system prune -a --volumes
```

## Architecture

```
┌─────────────────────────────────────────┐
│         Docker Compose Network          │
│              (ppob_network)             │
│                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────┐ │
│  │   API    │  │PostgreSQL│  │Redis │ │
│  │  :8080   │  │  :5432   │  │:6379 │ │
│  └──────────┘  └──────────┘  └──────┘ │
│       │             │           │      │
│       │             │           │      │
│  ┌────▼─────────────▼───────────▼───┐ │
│  │      Persistent Volumes           │ │
│  │  - ppob_postgres_data             │ │
│  │  - ppob_redis_data                │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## Environment Variables Reference

See `.env.example` for complete list and documentation.

### Required Variables

```env
# Database
DB_PASSWORD=<strong-password>

# JWT
JWT_SECRET=<strong-secret>
```

### Optional Variables (for full functionality)

```env
# WhatsApp OTP
WA_PHONE_NUMBER_ID=<your-phone-id>
WA_ACCESS_TOKEN=<your-token>

# Fazpass OTP
FAZPASS_MERCHANT_KEY=<your-key>
FAZPASS_GATEWAY_KEY=<your-key>

# Email (Brevo)
BREVO_API_KEY=<your-key>

# Payment Gateway (Gerbang)
GERBANG_CLIENT_ID=<your-client-id>
GERBANG_CLIENT_SECRET=<your-secret>
GERBANG_CALLBACK_SECRET=<your-webhook-secret>
```

## Makefile Commands

We provide a Makefile for convenience:

```bash
# Setup project
make setup

# Start development environment
make dev

# Docker operations
make docker-up
make docker-down
make docker-logs
make docker-clean

# Production
make docker-prod-up
make docker-prod-down

# Database
make db-shell
make db-backup
make redis-cli

# See all commands
make help
```

## Next Steps

1. ✅ Services running → Check `/health` endpoint
2. 📝 Configure API keys → Edit `.env`
3. 🔐 Test authentication → POST `/v1/auth/register`
4. 💰 Test transactions → Use Prepaid/Postpaid APIs
5. 📊 Monitor logs → `docker-compose logs -f`

## Support

- 📚 Full documentation: See main `README.md`
- 🐛 Issues: Check logs with `docker-compose logs`
- 💬 Questions: Review `.env.example` for configuration help
