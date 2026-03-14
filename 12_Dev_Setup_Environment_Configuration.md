# Development Environment Setup Guide

**Document ID**: 12_Dev_Setup_Environment_Configuration
**Version**: 1.0
**Last Updated**: 2026-03-14
**Owner**: DevOps Lead

---

## 📋 Quick Start (5 minutes)

For experienced developers, just run:

```bash
# Clone repo
git clone https://github.com/pintuotuo/pintuotuo.git
cd pintuotuo

# Start everything with Docker Compose
docker-compose up -d

# Install dependencies
npm install (frontend)
go mod download (backend)

# Open http://localhost:3000 (frontend)
# API: http://localhost:8080
# Mock API: http://localhost:3001
```

---

## 🖥️ Full Setup Guide

### Prerequisites (Must Have)

#### 1. Git (All Platforms)
```bash
# Verify installation
git --version

# If not installed:
# macOS: brew install git
# Linux: sudo apt-get install git
# Windows: https://git-scm.com/download/win
```

#### 2. Docker & Docker Compose (All Platforms)
```bash
# Verify installation
docker --version
docker-compose --version

# If not installed:
# macOS: brew install docker docker-compose
# Linux: https://docs.docker.com/engine/install/
# Windows: https://www.docker.com/products/docker-desktop

# Start Docker daemon (macOS/Windows: start Docker Desktop)
# Linux: sudo systemctl start docker
```

#### 3. Node.js & npm (Frontend Development)
```bash
# Verify installation
node --version  # Should be v18+
npm --version   # Should be v8+

# If not installed:
# macOS: brew install node
# Linux: sudo apt-get install nodejs npm
# Windows: https://nodejs.org/

# Install nvm (Node Version Manager) - Optional but recommended
# https://github.com/nvm-sh/nvm
```

#### 4. Go (Backend Development)
```bash
# Verify installation
go version  # Should be 1.21+

# If not installed:
# macOS: brew install go
# Linux: https://go.dev/doc/install
# Windows: https://go.dev/dl/

# Install gvm (Go Version Manager) - Optional but recommended
# https://github.com/moovweb/gvm
```

#### 5. IDE/Editor (Choose One)

**Frontend Developers**:
- Visual Studio Code (recommended)
  - Install: https://code.visualstudio.com/
  - Extensions: ES7+ React/Redux/React-Native snippets, Prettier, ESLint
- WebStorm (if using JetBrains license)

**Backend Developers**:
- VS Code + Go extension
- GoLand (JetBrains)
- Vim/Neovim (advanced users)

**All**:
- Install Prettier code formatter
- Install ESLint (JavaScript)
- Install Golangci-lint (Go)

---

## 📁 Repository Setup

### 1. Clone Repository

```bash
# Using HTTPS
git clone https://github.com/pintuotuo/pintuotuo.git
cd pintuotuo

# Using SSH (if SSH key configured)
git clone git@github.com:pintuotuo/pintuotuo.git
cd pintuotuo

# Verify
git status
```

### 2. Configure Git

```bash
# Global configuration (one time)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Optional: Set default editor
git config --global core.editor "code"

# For this project
git config user.name "Your Name"
git config user.email "your.email@example.com"
```

### 3. Create Feature Branch

```bash
# Pull latest develop branch
git checkout develop
git pull origin develop

# Create your feature branch
git checkout -b feature/your-feature-name

# Example: feature/user-authentication
# Example: feature/api-payment-integration

# Naming: feature/, bugfix/, hotfix/, release/
```

---

## 🐳 Docker Environment Setup

### 1. Start All Services

```bash
# In project root directory
docker-compose up -d

# Verify all containers running
docker-compose ps

# Expected output:
# NAME              STATUS
# postgres          Up
# redis             Up
# kafka             Up
# api_backend       Up
# frontend_mock     Up
# elasticsearch     Up
```

### 2. Verify Services

```bash
# PostgreSQL (Port 5432)
psql -h localhost -U pintuotuo -d pintuotuo_db -c "SELECT version();"

# Redis (Port 6379)
redis-cli ping  # Should return PONG

# Kafka (Port 9092)
kafka-broker-api-versions --bootstrap-server localhost:9092

# Backend API (Port 8080)
curl http://localhost:8080/health

# Mock API (Port 3001)
curl http://localhost:3001/api/products
```

### 3. Check Logs

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f postgres
docker-compose logs -f redis
docker-compose logs -f api_backend

# Exit: Ctrl+C
```

---

## 💾 Database Setup

### 1. Initialize Database

```bash
# Database is auto-initialized with docker-compose
# Verify tables created:
psql -h localhost -U pintuotuo -d pintuotuo_db

# List tables
\dt

# Exit
\q
```

### 2. Load Seed Data (Optional)

```bash
# Use script in repo
./scripts/db/seed-development.sh

# Or manually
psql -h localhost -U pintuotuo -d pintuotuo_db < scripts/db/seed.sql

# Verify data loaded
psql -h localhost -U pintuotuo -d pintuotuo_db -c "SELECT COUNT(*) FROM users;"
```

### 3. Database Credentials (Development Only)

```
Host: localhost
Port: 5432
Database: pintuotuo_db
User: pintuotuo
Password: dev_password_123
```

**⚠️ Never use these credentials in production!**

---

## 🚀 Backend Development Setup

### Go Backend

```bash
# Navigate to backend directory
cd backend

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Run tests (optional)
go test ./...

# Start server locally (if not using Docker)
go run main.go

# Expected output:
# Server running on http://localhost:8080
```

### Node.js Backend (If applicable)

```bash
# Navigate to directory
cd services/service-name

# Install dependencies
npm install

# Start server
npm start

# Expected output:
# Server running on http://localhost:8081
```

---

## ⚛️ Frontend Development Setup

```bash
# Navigate to frontend
cd frontend

# Install dependencies
npm install

# Start development server
npm start

# Expected output:
# Compiled successfully!
# You can now view pintuotuo in the browser.
# http://localhost:3000

# Open browser automatically or manually go to:
# http://localhost:3000
```

### Useful Frontend Commands

```bash
# Run tests
npm test

# Run linter (ESLint)
npm run lint

# Format code (Prettier)
npm run format

# Build for production
npm run build

# Analyze bundle size
npm run analyze
```

---

## 🔄 Environment Variables

### Frontend (.env.development)

```bash
REACT_APP_API_URL=http://localhost:8080
REACT_APP_MOCK_API_URL=http://localhost:3001
REACT_APP_ENVIRONMENT=development
REACT_APP_LOG_LEVEL=debug
```

### Backend (.env.development)

```bash
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=pintuotuo
DB_PASSWORD=dev_password_123
DB_NAME=pintuotuo_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka
KAFKA_BROKERS=localhost:9092

# JWT
JWT_SECRET=dev_secret_key_change_in_production

# API Keys (Mock for development)
ALIPAY_API_KEY=mock_key
WECHAT_API_KEY=mock_key
```

Create these files in project root:
```bash
# Frontend
cp .env.example .env.development

# Backend
cp .env.example .env.development
```

---

## ✅ Verification Checklist

Run through this to verify everything is working:

```bash
# 1. Git
git --version
git status

# 2. Docker
docker --version
docker-compose ps

# 3. Frontend
cd frontend
npm --version
npm list react

# 4. Backend
cd ../backend
go version
go list ./...

# 5. Database
psql -h localhost -U pintuotuo -d pintuotuo_db -c "SELECT version();"

# 6. Redis
redis-cli ping

# 7. APIs
curl http://localhost:8080/health
curl http://localhost:3001/api/products
```

**All commands should complete without errors ✅**

---

## 🐛 Troubleshooting

### Docker Issues

**Problem**: "Docker daemon is not running"
```bash
# macOS/Windows: Start Docker Desktop
# Linux: sudo systemctl start docker
```

**Problem**: "Port already in use"
```bash
# Find process using port (e.g., 5432)
lsof -i :5432

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
```

**Problem**: "Cannot connect to database"
```bash
# Verify container is running
docker ps

# Check logs
docker logs postgres

# Restart services
docker-compose restart postgres
```

### Node/npm Issues

**Problem**: "npm ERR! ERESOLVE unable to resolve dependency tree"
```bash
# Clear cache
npm cache clean --force

# Install with legacy peer deps
npm install --legacy-peer-deps

# Or update Node to latest version
nvm install --lts
```

**Problem**: "Module not found"
```bash
# Delete node_modules and package-lock.json
rm -rf node_modules package-lock.json

# Reinstall
npm install
```

### Go Issues

**Problem**: "go: command not found"
```bash
# Add Go to PATH
export PATH=$PATH:/usr/local/go/bin

# Add to .bashrc or .zshrc for permanent
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**Problem**: "go mod tidy fails"
```bash
# Clear Go cache
go clean -modcache

# Tidy modules
go mod tidy
```

---

## 📚 Useful Commands Reference

### Git Commands
```bash
# Check status
git status

# Pull latest
git pull origin develop

# Create branch
git checkout -b feature/name

# Stage changes
git add .

# Commit
git commit -m "feat: description"

# Push
git push origin feature/name

# Create Pull Request
# (Go to GitHub/GitLab web interface)
```

### Docker Commands
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Rebuild containers
docker-compose build

# Clean up (remove unused containers/images)
docker system prune
```

### Frontend Commands
```bash
npm start        # Start dev server
npm test         # Run tests
npm run build    # Build for production
npm run lint     # Check code style
npm run format   # Format code
```

### Backend Commands (Go)
```bash
go run main.go           # Run application
go test ./...            # Run tests
go mod tidy              # Clean up modules
go fmt ./...             # Format code
golangci-lint run ./...  # Lint code
```

---

## 🔐 SSH Key Setup (Optional but Recommended)

```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your.email@example.com"

# Add to SSH agent
ssh-add ~/.ssh/id_ed25519

# Copy public key
cat ~/.ssh/id_ed25519.pub

# Add to GitHub:
# Settings → SSH and GPG keys → New SSH key → Paste
```

---

## 🎯 Next Steps After Setup

1. ✅ Complete all verification checklist items
2. ✅ Create your feature branch
3. ✅ Read CONTRIBUTING.md for code standards
4. ✅ Check Jira for assigned Week 2 tasks
5. ✅ Attend daily standup (09:15)
6. ✅ Join Slack channels (#general, #your-team)

---

## 💬 Getting Help

**Setup Issues?**
- Check this guide's Troubleshooting section
- Ask in #development Slack channel
- Tag @devops-lead for urgent issues

**Questions?**
- DM project manager for quick questions
- Schedule sync for complex discussions

**Documentation?**
- Code standards: CONTRIBUTING.md
- Architecture: 05_Technical_Architecture_and_Tech_Stack.md
- API spec: 04_API_Specification.md

---

## 📋 Setup Completion Checklist

- [ ] Git installed and configured
- [ ] Docker & docker-compose running
- [ ] Node.js v18+ installed
- [ ] Go 1.21+ installed
- [ ] IDE/Editor configured
- [ ] Repository cloned
- [ ] All Docker containers running
- [ ] Database initialized
- [ ] Frontend npm dependencies installed
- [ ] Backend Go modules downloaded
- [ ] Frontend dev server starts on :3000
- [ ] Backend API responds on :8080
- [ ] Mock API responds on :3001
- [ ] Environment variables configured
- [ ] Git configured with name/email
- [ ] Feature branch created
- [ ] Added to team Slack channels
- [ ] Assigned Week 2 tasks in Jira
- [ ] Read CONTRIBUTING.md

**When all items ✅**: Ready for Week 2 development!

---

**Version**: 1.0
**Last Updated**: 2026-03-14
**Maintained By**: DevOps Team
**Status**: Active & Ready for Use
