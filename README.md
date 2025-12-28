# 📧 Temp Mail 365 - Temporary Email Service

A modern, secure temporary email service built with Go, Next.js, MongoDB, and Redis. Provides disposable email addresses with real-time notifications and end-to-end encryption.

## ✨ Features

- **🔒 Secure & Private**: AES-256-GCM encryption, zero-logging policy, auto-deletion
- **⚡ Real-time**: WebSocket notifications for instant email updates
- **📬 Full SMTP**: Complete email reception with MIME parsing
- **🎨 Modern UI**: Clean, responsive interface built with Next.js and TailwindCSS
- **🐳 Docker Ready**: Complete containerized setup with Docker Compose
- **🌍 Multi-language**: Support for 9+ languages
- **🚀 Production Ready**: Rate limiting, health checks, graceful shutdown

## 🏗️ Architecture

```
┌─────────────────┐
│   Frontend      │
│   (Next.js)     │
└────────┬────────┘
         │
         ↓
┌─────────────────┐      ┌──────────────┐
│   API Gateway   │←────→│  WebSocket   │
│   (Gin Router)  │      │     Hub      │
└────────┬────────┘      └──────┬───────┘
         │                      │
         ├──────────────────────┤
         ↓                      ↓
┌─────────────────┐      ┌──────────────┐
│  Repositories   │      │ SMTP Server  │
│ (Email/Mailbox) │      │   + Parser   │
└────────┬────────┘      └──────┬───────┘
         │                      │
         ├──────────────────────┤
         ↓                      ↓
┌─────────────────┬──────────────────┐
│    MongoDB      │      Redis       │
│  (Email Store)  │  (Mailbox State) │
└─────────────────┴──────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- (Optional) Go 1.21+ for local development
- (Optional) Node.js 18+ for local development

### 1. Clone the Repository

```bash
git clone https://github.com/ARTEMKOPIK/Temp-Mail-NEW-365.git
cd Temp-Mail-NEW-365
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

**Important**: Change these values in production:
- `ENCRYPTION_KEY`: 32-byte encryption key
- `JWT_SECRET`: Secret for JWT tokens
- `EMAIL_DOMAIN`: Your domain name

### 3. Start Services

```bash
docker-compose up -d
```

This will start:
- MongoDB on `localhost:27017`
- Redis on `localhost:6379`
- Backend API on `localhost:8080`
- SMTP Server on `localhost:2525`
- Frontend on `localhost:3000`

### 4. Access the Application

Open your browser: **http://localhost:3000**

## 📡 API Endpoints

### Mailbox Management
- `POST /api/mailbox` - Create new mailbox
- `GET /api/mailbox/:address` - Get mailbox info
- `DELETE /api/mailbox/:address` - Delete mailbox
- `GET /api/mailbox/:address/emails` - Get all emails

### Email Operations
- `GET /api/email/:id` - Get full email
- `DELETE /api/email/:id` - Delete email

### WebSocket
- `GET /api/ws/mailbox/:address` - WebSocket connection for real-time updates

### Health Check
- `GET /health` - Service health status

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_PORT` | 8080 | Backend API port |
| `SMTP_PORT` | 2525 | SMTP server port |
| `MONGO_URI` | mongodb://mongo:27017 | MongoDB connection string |
| `MONGO_DB` | tempmail | MongoDB database name |
| `REDIS_URL` | redis://redis:6379 | Redis connection string |
| `EMAIL_DOMAIN` | your-tempmail.info | Email domain |
| `EMAIL_EXPIRY_MINUTES` | 60 | Email lifetime in minutes |
| `ENCRYPTION_KEY` | *required* | 32-byte encryption key |
| `RATE_LIMIT_REQUESTS` | 100 | Max requests per window |
| `RATE_LIMIT_WINDOW` | 60 | Rate limit window (seconds) |

## 🛠️ Development

### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Run locally
go run main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run dev server
npm run dev
```

## 🧪 Testing

### Test SMTP Server

```bash
# Using swaks
swaks --to test@your-tempmail.info \
      --from sender@example.com \
      --server localhost:2525 \
      --body "Test email"
```

### Test API

```bash
# Create mailbox
curl -X POST http://localhost:8080/api/mailbox \
  -H "Content-Type: application/json" \
  -d '{"expiryMinutes": 30}'

# Get emails
curl http://localhost:8080/api/mailbox/YOUR_ADDRESS/emails
```

## 🔒 Security Features

- **AES-256-GCM Encryption**: Email bodies encrypted at rest
- **PBKDF2 Key Derivation**: 100,000 iterations with SHA-256
- **Zero-Logging**: No IP addresses or personal data stored
- **Auto-Deletion**: TTL indexes for automatic cleanup
- **Rate Limiting**: Per-IP request throttling
- **Input Validation**: All endpoints protected
- **HTTPS/WSS**: TLS 1.3 in production

## 📦 Project Structure

```
.
├── backend/
│   ├── internal/
│   │   ├── api/              # HTTP handlers & router
│   │   ├── config/           # Configuration management
│   │   ├── crypto/           # Encryption service
│   │   ├── database/         # MongoDB & Redis clients
│   │   ├── models/           # Data structures
│   │   ├── repository/       # Data access layer
│   │   ├── smtp/             # SMTP server & parser
│   │   └── websocket/        # WebSocket hub & client
│   ├── main.go               # Application entry point
│   ├── go.mod                # Go dependencies
│   └── Dockerfile            # Backend container
├── frontend/
│   ├── src/
│   │   ├── app/              # Next.js pages
│   │   ├── components/       # React components
│   │   └── lib/              # API & WebSocket clients
│   ├── package.json          # Node dependencies
│   └── Dockerfile            # Frontend container
├── docker-compose.yml        # Service orchestration
└── .env.example              # Environment template
```

## 🚢 Production Deployment

### 1. Security Checklist

- [ ] Change `ENCRYPTION_KEY` to secure random 32-byte key
- [ ] Change `JWT_SECRET` to secure random string
- [ ] Set `EMAIL_DOMAIN` to your domain
- [ ] Enable HTTPS/TLS for frontend and backend
- [ ] Enable WSS for WebSocket connections
- [ ] Configure firewall rules
- [ ] Set up MongoDB authentication
- [ ] Set up Redis password
- [ ] Configure proper CORS origins

### 2. DNS Configuration

```
# A Records
your-domain.com        → Your_Server_IP
*.your-domain.com      → Your_Server_IP (for wildcard emails)

# MX Record
your-domain.com   10   mail.your-domain.com
```

### 3. SMTP Configuration

Update MX records to point to your server for receiving emails.

### 4. Monitoring

- Backend health: `http://your-domain:8080/health`
- MongoDB: Use MongoDB Compass or mongosh
- Redis: Use redis-cli or RedisInsight
- Logs: `docker-compose logs -f`

## 📊 Performance

- **Connection Pooling**: MongoDB (10-100), Redis configured
- **TTL Indexes**: Automatic email cleanup
- **Rate Limiting**: 100 requests/minute per IP
- **WebSocket**: Buffered channels (256 messages)
- **SMTP**: 10MB max message size, 50 max recipients

## 🐛 Troubleshooting

### Backend won't start
```bash
# Check logs
docker-compose logs backend

# Verify MongoDB connection
docker exec -it tempmail-mongo mongosh

# Verify Redis connection
docker exec -it tempmail-redis redis-cli ping
```

### Emails not receiving
```bash
# Check SMTP server logs
docker-compose logs backend | grep SMTP

# Test SMTP connection
telnet localhost 2525
```

### Frontend build fails
```bash
# Clear cache
cd frontend
rm -rf .next node_modules
npm install
npm run build
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgements

- Built with [Go](https://golang.org/), [Next.js](https://nextjs.org/), [MongoDB](https://www.mongodb.com/), [Redis](https://redis.io/)
- SMTP server powered by [go-smtp](https://github.com/emersion/go-smtp)
- WebSocket implementation using [Gorilla WebSocket](https://github.com/gorilla/websocket)

## 📞 Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/ARTEMKOPIK/Temp-Mail-NEW-365/issues)
- Email: contact.tempmail365@gmail.com

---

**⭐ Star this repo if you find it useful!**

