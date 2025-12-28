# Temp Mail 365 - Temporary Email Service

A modern, privacy-focused temporary email service built with Next.js, Go, MongoDB, and Redis.

## 🌟 Features

- **Zero Registration**: No signup required, instant email generation
- **Real-time Updates**: WebSocket integration for instant email notifications
- **Customizable Lifetime**: Choose between 5, 10, 30, or 60 minutes
- **Privacy First**: Zero-logging policy, no IP addresses stored
- **End-to-End Encryption**: AES-256-GCM encryption for email content
- **Auto-Deletion**: Emails and mailboxes automatically deleted after expiration
- **Multi-language Support**: 9 languages (English, Russian, Chinese, German, French, Spanish, Korean, Japanese, Italian)
- **Responsive Design**: Beautiful gradient UI with smooth animations
- **Comprehensive Documentation**: Tutorials, FAQ, Blog, Privacy Policy, Terms of Service

## 🏗️ Architecture

### Frontend
- **Framework**: Next.js 14 with React 18 and App Router
- **Styling**: TailwindCSS with custom gradients
- **Language**: TypeScript
- **State Management**: React Hooks + localStorage
- **Real-time**: WebSocket client with auto-reconnection
- **Icons**: Lucide React

### Backend
- **Language**: Go 1.21+
- **SMTP Server**: Custom implementation using go-smtp
- **REST API**: Gin web framework
- **WebSocket**: Gorilla WebSocket
- **Databases**: MongoDB (emails) + Redis (caching, pub/sub)
- **Encryption**: AES-256-GCM for email content

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Security**: HTTPS/WSS (TLS 1.3), zero-logging
- **Auto-cleanup**: MongoDB TTL indexes

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-username/temp-mail-365.git
   cd temp-mail-365
   ```

2. **Configure environment variables**

   Backend (.env in project root):
   ```bash
   # Copy from .env.example
   cp .env.example .env
   
   # Edit and set your values
   nano .env
   ```

   Frontend (frontend/.env.local):
   ```bash
   # Copy from example
   cp frontend/.env.local.example frontend/.env.local
   
   # Edit if needed (defaults work for local development)
   nano frontend/.env.local
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - SMTP Server: localhost:2525

### Manual Setup (Without Docker)

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Set environment variables (see .env.example)
export SMTP_PORT=2525
export BACKEND_PORT=8080
export MONGO_URL=mongodb://localhost:27017
export MONGO_DATABASE=tempmail
export REDIS_URL=redis://localhost:6379
export EMAIL_DOMAIN=your-tempmail.info
export ENCRYPTION_KEY=$(openssl rand -hex 32)
export FRONTEND_URL=http://localhost:3000

# Run the server
go run main.go
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Set environment variables
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
echo "NEXT_PUBLIC_WS_URL=ws://localhost:8080" >> .env.local
echo "NEXT_PUBLIC_EMAIL_DOMAIN=your-tempmail.info" >> .env.local

# Run development server
npm run dev

# Or build for production
npm run build
npm start
```

## 📁 Project Structure

```
temp-mail-365/
├── backend/
│   ├── internal/
│   │   ├── api/           # REST API handlers and routing
│   │   ├── config/        # Configuration management
│   │   ├── crypto/        # Encryption service
│   │   ├── database/      # MongoDB and Redis clients
│   │   ├── models/        # Data models
│   │   ├── repository/    # Data access layer
│   │   ├── smtp/          # SMTP server implementation
│   │   └── websocket/     # WebSocket hub and clients
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── app/           # Next.js pages and layouts
│   │   │   ├── page.tsx   # Homepage
│   │   │   ├── faq/       # FAQ page
│   │   │   ├── tutorials/ # Tutorials page
│   │   │   ├── blog/      # Blog listing
│   │   │   ├── contact/   # Contact page
│   │   │   ├── privacy/   # Privacy Policy
│   │   │   └── terms/     # Terms of Service
│   │   └── lib/           # Utilities and API clients
│   ├── package.json
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   └── Dockerfile
├── docker-compose.yml
├── .env.example
├── .gitignore
└── README.md
```

## 🔐 Security

- **Zero Logging**: No IP addresses, browser fingerprints, or personal data stored
- **Encryption**: AES-256-GCM for email content at rest
- **Secure Transmission**: HTTPS/WSS with TLS 1.3
- **Auto-Deletion**: All data permanently deleted after expiration
- **No Third-Party Tracking**: No analytics, no cookies, no tracking
- **Input Validation**: All API inputs validated and sanitized

## 🌐 API Documentation

### REST API Endpoints

#### Mailbox

- `POST /api/mailbox` - Create new temporary mailbox
  ```json
  {
    "expiryMinutes": 10
  }
  ```

- `GET /api/mailbox/:address` - Get mailbox information
- `DELETE /api/mailbox/:address` - Delete mailbox
- `GET /api/mailbox/:address/emails` - List all emails

#### Email

- `GET /api/email/:id` - Get email content

#### WebSocket

- `GET /api/ws/mailbox/:address` - WebSocket connection for real-time updates

### SMTP Server

- **Host**: localhost (or your domain)
- **Port**: 2525 (configurable)
- **Accepts**: Any email to `*@your-tempmail.info`

## 🧪 Testing

Send a test email to your temporary address:

```bash
# Using swaks (SMTP test tool)
swaks --to test123@your-tempmail.info \
      --from sender@example.com \
      --server localhost:2525 \
      --header "Subject: Test Email" \
      --body "This is a test message"
```

## 🌍 Multi-language Support

The application supports 9 languages:
- English (en)
- Russian (ru)
- Chinese (zh)
- German (de)
- French (fr)
- Spanish (es)
- Korean (ko)
- Japanese (ja)
- Italian (it)

## 📝 License

This project is open source and available under the MIT License.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📧 Contact

For questions or support:
- Email: contact.tempmail365@gmail.com
- GitHub Issues: [Create an issue](https://github.com/your-username/temp-mail-365/issues)

## 🙏 Acknowledgments

- Inspired by Temp Mail 365 service
- Built with love for privacy and security
- Community-driven development

---

**⚠️ Important Notice**: This service is for receiving temporary emails only. Do not use it for:
- Banking or financial accounts
- Important account registrations
- Password recovery
- Any service requiring long-term access
- Sensitive or confidential information

**Your privacy is our priority. All data is automatically deleted and never recoverable.**

© 2025 Temp Mail 365. All rights reserved.

