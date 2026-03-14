#!/bin/bash

# Project Initialization Script for Pintuotuo MVP
# This script sets up the initial project structure for Week 1

set -e

echo "🚀 Initializing Pintuotuo Project Structure..."

# Create main directories
echo "📁 Creating directories..."

# Backend directories
mkdir -p backend/services/{user,product,group,order,payment,token}
mkdir -p backend/middleware
mkdir -p backend/models
mkdir -p backend/utils
mkdir -p backend/config
mkdir -p backend/tests

# Frontend directories
mkdir -p frontend/src/{components,pages,hooks,services,stores,types,utils,styles}
mkdir -p frontend/public

# Services directories
mkdir -p services/mock-api/{routes,data,utils}
mkdir -p services/api-gateway
mkdir -p services/notification

# Database directories
mkdir -p scripts/db/{migrations,seeds}

# Documentation directories
mkdir -p docs/architecture
mkdir -p docs/api

# Testing directories
mkdir -p tests/{unit,integration,e2e}

# CI/CD directories
mkdir -p .github/workflows
mkdir -p .gitlab-ci

echo "✅ Directory structure created"

# Create backend placeholder files
echo "📝 Creating backend placeholders..."
cat > backend/go.mod << 'EOF'
module github.com/pintuotuo/backend

go 1.21

require (
    github.com/gin-gonic/gin v1.9.0
    github.com/lib/pq v1.10.9
    github.com/golang-jwt/jwt/v5 v5.0.0
    github.com/redis/go-redis/v9 v9.0.0
    github.com/segmentio/kafka-go v0.4.40
)
EOF

cat > backend/main.go << 'EOF'
package main

import (
    "fmt"
)

func main() {
    fmt.Println("🚀 Pintuotuo Backend Server - Week 1 Initialization")
    fmt.Println("Backend services will be implemented in Week 2-4")
    fmt.Println("Documentation: ../05_Technical_Architecture_and_Tech_Stack.md")
}
EOF

echo "✅ Backend placeholders created"

# Create frontend placeholder files
echo "📝 Creating frontend placeholders..."
cat > frontend/package.json << 'EOF'
{
  "name": "pintuotuo-frontend",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint src --ext ts,tsx",
    "type-check": "tsc --noEmit",
    "format": "prettier --write src/"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.15.0",
    "zustand": "^4.4.0",
    "axios": "^1.6.0",
    "antd": "^5.10.0",
    "react-icons": "^4.12.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.1.0",
    "vite": "^4.5.0",
    "typescript": "^5.2.0",
    "prettier": "^3.1.0",
    "eslint": "^8.50.0"
  }
}
EOF

cat > frontend/tsconfig.json << 'EOF'
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,

    /* Bundler mode */
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "moduleResolution": "node",
    "noEmit": true,
    "jsx": "react-jsx",

    /* Linting */
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,

    /* Path mapping */
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"],
      "@pages/*": ["src/pages/*"],
      "@services/*": ["src/services/*"],
      "@stores/*": ["src/stores/*"],
      "@hooks/*": ["src/hooks/*"],
      "@types/*": ["src/types/*"],
      "@utils/*": ["src/utils/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
EOF

echo "✅ Frontend placeholders created"

# Create mock API placeholder
echo "📝 Creating mock API placeholder..."
cat > services/mock-api/package.json << 'EOF'
{
  "name": "mock-api",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "start": "node server.js"
  },
  "dependencies": {
    "json-server": "^0.17.3",
    "cors": "^2.8.5"
  }
}
EOF

echo "✅ Mock API placeholders created"

# Create database initialization script
echo "📝 Creating database initialization script..."
cat > scripts/db/init.sql << 'EOF'
-- Initial database schema will be created in Week 2
-- This file is a placeholder for the full schema

-- Database: pintuotuo_db (created automatically by docker-compose)

-- TODO: Add schema creation in Week 2 after architecture review
-- Reference: 03_Data_Model_Design.md for complete schema

CREATE SCHEMA IF NOT EXISTS public;

-- Placeholder tables - to be replaced with real schema
CREATE TABLE IF NOT EXISTS schema_version (
    version_id SERIAL PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    installed_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_version (description) VALUES
('Initial schema - Week 2 planning');
EOF

echo "✅ Database script created"

# Create .env.development file
echo "📝 Creating environment configuration..."
cat > .env.development << 'EOF'
# Backend Configuration
BACKEND_PORT=8080
DATABASE_URL=postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db
REDIS_URL=redis://localhost:6379
KAFKA_BROKERS=localhost:9092

# Frontend Configuration
VITE_API_URL=http://localhost:8000
VITE_MOCK_API_URL=http://localhost:3001

# JWT Configuration
JWT_SECRET=dev_secret_key_change_in_production
JWT_EXPIRE_HOURS=24
JWT_REFRESH_EXPIRE_DAYS=30

# Application
APP_ENV=development
APP_LOG_LEVEL=debug

# Third-party APIs (Mock for development)
ALIPAY_API_KEY=mock_key_change_in_production
WECHAT_API_KEY=mock_key_change_in_production
EOF

echo "✅ Environment configuration created"

# Create .gitignore
echo "📝 Creating .gitignore..."
cat > .gitignore << 'EOF'
# Dependencies
node_modules/
vendor/
go.sum

# Build outputs
dist/
build/
*.o
*.so
*.dylib

# Environment
.env
.env.local
.env.*.local
.DS_Store

# IDE
.vscode/
.idea/
*.swp
*.swo
*~
.vim/

# Logs
*.log
logs/

# Database
*.db
*.sqlite

# Testing
coverage/
.nyc_output/

# Cache
.cache/
tmp/
EOF

echo "✅ .gitignore created"

# Create README
echo "📝 Creating project README..."
cat > README.md << 'EOF'
# Pintuotuo (拼脱脱) - AI Token Secondary Market Platform

## Quick Start (5 minutes)

### Prerequisites
- Docker & Docker Compose
- Git
- Node.js v18+ (for frontend development)
- Go 1.21+ (for backend development)

### Setup

```bash
# Clone the repository
git clone https://github.com/pintuotuo/pintuotuo.git
cd pintuotuo

# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps

# Expected services:
# postgres    Up
# redis       Up
# kafka       Up
# zookeeper   Up
# mock_api    Up
```

### Verify Installation

```bash
# Check PostgreSQL
psql -h localhost -U pintuotuo -d pintuotuo_db -c "SELECT version();"

# Check Redis
redis-cli ping

# Check Mock API
curl http://localhost:3001/api/products

# Expected output: JSON array of products
```

## Project Structure

```
pintuotuo/
├── backend/              # Go backend services
├── frontend/             # React frontend
├── services/
│   ├── mock-api/        # Mock API for development
│   ├── api-gateway/     # API Gateway (Kong)
│   └── notification/    # Notification service
├── scripts/
│   └── db/              # Database scripts
├── docs/                # Documentation
├── docker-compose.yml   # Docker services
└── README.md
```

## Documentation

### Getting Started
1. **Setup Guide**: `12_Dev_Setup_Environment_Configuration.md`
2. **Git Workflow**: `13_Dev_Git_Workflow_Code_Standards.md`
3. **Architecture**: `05_Technical_Architecture_and_Tech_Stack.md`

### Product Documentation
1. **PRD**: `01_PRD_Complete_Product_Specification.md`
2. **User Flows**: `02_User_Flow_and_Journey.md`
3. **Data Model**: `03_Data_Model_Design.md`
4. **API Spec**: `04_API_Specification.md`

### Execution Plans
1. **Master Plan**: `17_Master_Execution_Plan_Complete_Overview.md`
2. **Week 1**: `11_Plan_Week1_Project_Launch_Execution.md`
3. **Week 2**: `14_Plan_Week2_Database_and_API_Design.md`
4. **Week 3**: `15_Plan_Week3_Frontend_Setup_Design_System.md`

## Development

### Backend Development
```bash
cd backend
go mod download
go run main.go
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

### Stop Services
```bash
docker-compose down
```

## Environment Variables

See `.env.development` for all configuration options.

**⚠️ Never commit `.env` or `.env.*.local` files!**

## Timeline

- **Week 1** (2026-03-17): Setup & Alignment ✅
- **Week 2** (2026-03-24): Database & API Design
- **Week 3** (2026-03-31): Frontend Setup & Design System
- **Week 4-5** (2026-04-07): Core Feature Implementation
- **Week 6** (2026-04-14): Optimization & Advanced Features
- **Week 7** (2026-04-21): QA & Testing
- **Week 8** (2026-04-28): Gray Release & Launch

## Team

- **Backend**: 4 engineers
- **Frontend**: 3 engineers
- **QA/Testing**: 1.5 engineers
- **DevOps**: 1 engineer
- **Product/Design/Ops**: 3+ people

## Communication

- **Daily Standups**: 09:15 AM (15 minutes)
- **Slack**: #general, #backend, #frontend, #design, #ops, #blockers
- **Jira**: Sprint planning and task tracking

## Support

For issues or questions:
1. Check the documentation in this repository
2. Ask in the relevant Slack channel
3. Create a Jira ticket if it's a bug or feature request
4. Contact your team lead

## License

(To be determined)

---

**Status**: Week 1 Planning Complete ✅
**Next Milestone**: Week 2 Database & API Design
**Last Updated**: 2026-03-14
EOF

echo "✅ README created"

echo ""
echo "✨ Project initialization complete!"
echo ""
echo "📋 Next Steps:"
echo "1. Run: docker-compose up -d"
echo "2. Verify: docker-compose ps"
echo "3. Check database: psql -h localhost -U pintuotuo -d pintuotuo_db -c \"SELECT version();\""
echo "4. Initialize Git: git init && git add . && git commit -m \"init: project structure\""
echo "5. Review: 17_Master_Execution_Plan_Complete_Overview.md"
echo ""
echo "🚀 Ready for Week 1 execution!"
