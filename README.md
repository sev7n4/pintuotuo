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

### 🚀 Essential Developer Guide
- **[CLAUDE.md](./CLAUDE.md)** - **START HERE** - Complete development standards, Git workflow, code guidelines, and AI assistant guidance (1000+ lines)

### Getting Started
1. **Setup Guide**: `12_Dev_Setup_Environment_Configuration.md`
2. **Git Workflow & Code Standards**: `13_Dev_Git_Workflow_Code_Standards.md`
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

## 自动化部署测试
测试时间: 2026-03-20 07:13:43
