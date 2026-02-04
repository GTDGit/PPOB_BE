# Nginx Configuration for PPOB

## Structure

```
nginx/
├── nginx.conf          # Production config (with SSL)
├── nginx.dev.conf      # Development config (HTTP only)
├── ssl/                # SSL certificates directory
│   └── ppob.id/
│       ├── fullchain.pem
│       └── privkey.pem
└── certbot/            # Let's Encrypt challenge directory
```

## Domains

| Domain | Description | Upstream |
|--------|-------------|----------|
| `ppob.id` / `www.ppob.id` | Main frontend website | `ppob_frontend:3000` |
| `admin.ppob.id` | Admin panel | `ppob_admin:3000` |
| `api.ppob.id` | Backend API | `ppob_backend:8080` |

## Development Usage

For development, use the simplified config without SSL:

```bash
# In docker-compose.yml, change the nginx volume to:
# - ./nginx/nginx.dev.conf:/etc/nginx/nginx.conf:ro
```

Or run directly:
```bash
docker-compose up -d
```

## Production Usage

### 1. Generate SSL Certificates (Let's Encrypt)

```bash
# Create directories
mkdir -p nginx/ssl/ppob.id nginx/certbot

# First time: Get certificates using certbot
docker run -it --rm \
  -v $(pwd)/nginx/certbot:/var/www/certbot \
  -v $(pwd)/nginx/ssl/ppob.id:/etc/letsencrypt/live/ppob.id \
  certbot/certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d ppob.id -d www.ppob.id -d admin.ppob.id -d api.ppob.id
```

### 2. Run with Production Config

```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 3. Certificate Renewal (Cron Job)

Add to crontab:
```bash
0 0 1 * * docker run --rm -v /path/to/nginx/certbot:/var/www/certbot -v /path/to/nginx/ssl:/etc/letsencrypt certbot/certbot renew && docker exec ppob_nginx nginx -s reload
```

## Rate Limiting

| Zone | Rate | Usage |
|------|------|-------|
| `api_limit` | 30 req/s | General API endpoints |
| `auth_limit` | 5 req/s | Auth & OTP endpoints |
| `conn_limit` | Varies | Connection limits per IP |

## Security Features

- ✅ HTTPS redirect
- ✅ TLS 1.2/1.3 only
- ✅ HSTS enabled
- ✅ Security headers (X-Frame-Options, X-Content-Type-Options, etc.)
- ✅ Rate limiting
- ✅ Connection limits
- ✅ CORS configured
- ✅ Gzip compression

## Troubleshooting

### Test configuration
```bash
docker exec ppob_nginx nginx -t
```

### Reload configuration
```bash
docker exec ppob_nginx nginx -s reload
```

### View logs
```bash
docker logs ppob_nginx -f
```

### Check certificate expiry
```bash
openssl s_client -connect api.ppob.id:443 -servername api.ppob.id 2>/dev/null | openssl x509 -noout -dates
```
