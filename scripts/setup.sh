#!/bin/bash

# Pintuotuo Project Setup Script
# This script sets up the development environment

set -e

echo "🚀 Starting Pintuotuo development environment setup..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for required tools
echo -e "${BLUE}Checking required tools...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${YELLOW}Docker Compose is not installed. Please install Docker Compose first.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker and Docker Compose found${NC}"

# Start Docker containers
echo -e "${BLUE}Starting Docker containers...${NC}"
docker-compose up -d

# Wait for PostgreSQL to be ready
echo -e "${BLUE}Waiting for PostgreSQL to be ready...${NC}"
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U pintuotuo &> /dev/null; then
        echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
        break
    fi
    echo "Waiting... ($i/30)"
    sleep 1
done

# Install backend dependencies（与 Makefile 一致：修正沙箱下异常的 GOMODCACHE）
echo -e "${BLUE}Installing backend dependencies...${NC}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export GOMODCACHE="$("$SCRIPT_DIR/ensure-go-modcache.sh")"
cd backend
go mod download
echo -e "${GREEN}✓ Backend dependencies installed${NC}"
cd ..

# Install frontend dependencies
echo -e "${BLUE}Installing frontend dependencies...${NC}"
cd frontend
npm install
echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
cd ..

echo ""
echo -e "${GREEN}✅ Pintuotuo development environment is ready!${NC}"
echo ""
echo "📚 Quick start commands:"
echo "  Backend:  cd backend && go run main.go"
echo "  Frontend: cd frontend && npm run dev"
echo "  Database: docker-compose exec postgres psql -U pintuotuo -d pintuotuo_db"
echo ""
echo "🔗 Service URLs:"
echo "  Backend:    http://localhost:8080"
echo "  Frontend:   http://localhost:5173"
echo "  PostgreSQL: localhost:5432"
echo "  Redis:      localhost:6379"
echo ""
