# 🚀 Week 1 Execution - Complete Team Runbook

**Status**: Ready to Execute
**Start Date**: Monday, March 17, 2026
**Team**: All 15.5 members
**Timeline**: 5 days (Mon-Fri)
**Success Criteria**: All 15 week-1 deliverables completed, team ready for Week 2

---

## 📋 Before Monday Morning

### Friday 2026-03-15 Preparation

**PM / Project Manager**:
- [ ] Import WEEK1_Jira_Task_Breakdown.md into Jira
- [ ] Create epic: "Week 1: Project Launch & Environment Setup"
- [ ] Assign tasks to team members (by team, by role)
- [ ] Create Slack channels: #general, #backend, #frontend, #design, #ops, #blockers, #random
- [ ] Block calendar for:
  - Daily 09:15 standups (Mon-Fri)
  - Friday 16:00 retrospective
  - All day-specific meeting blocks from WEEK1_Jira_Task_Breakdown
- [ ] Send team message: "Week 1 starts Monday! Review 17_Master_Execution_Plan"

**All Team Members**:
- [ ] Read: 17_Master_Execution_Plan_Complete_Overview.md (30 min)
- [ ] Read: EXECUTION_READY_Summary.md (15 min)
- [ ] Read: WEEK1_Jira_Task_Breakdown.md (skim, understand your assignments)
- [ ] Install Docker & Docker Compose (if not already done)
- [ ] Clone repository: `git clone https://github.com/pintuotuo/pintuotuo.git`
- [ ] Join Slack workspace and #general channel
- [ ] Confirm "Ready for Monday ✅" in Slack #general

**Team Leads** (Backend, Frontend, QA, DevOps, PM):
- [ ] Review your team's specific tasks for Week 1
- [ ] Review 11_Plan_Week1_Project_Launch_Execution.md (detailed breakdown)
- [ ] Schedule 1:1s with your team members (optional but recommended)
- [ ] Prepare breakout session agenda (what to cover, time allocation)
- [ ] Identify any resources or blockers that need pre-approval

**CTO**:
- [ ] Review architecture and execution plan one final time
- [ ] Prepare 15-min opening remarks for kickoff meeting
- [ ] Confirm all technical decisions are locked
- [ ] Be ready for Q&A during architecture review on Tuesday

---

## 📅 Monday 2026-03-17 - Day 1: Kickoff & Setup

### Morning (09:00-12:30)

#### 09:00-09:15: Team Assembling
- Everyone logging into Zoom/Teams
- Testing audio/video
- All 15.5 team members present

#### 09:15-10:00: Project Kickoff Meeting
- **Format**: All-hands meeting
- **Attendees**: All 15.5 team members
- **Agenda**:
  1. Welcome from CTO (5 min)
  2. Product vision & market opportunity (5 min)
  3. MVP scope and timeline overview (5 min)
  4. Success metrics (3 min)
  5. Q&A (2 min)
- **Output**: Team alignment, excitement, questions noted
- **Materials**: Slides from 05_Technical_Architecture_and_Tech_Stack.md + 01_PRD overview

#### 10:15-12:30: Role-Based Breakout Sessions
**Each team meets separately for 2.25 hours**:

**Backend Team Breakout** (4 people + Backend Lead):
- Review tech stack: Go, PostgreSQL, Redis, Kafka
- Discuss database design approach
- Plan API structure
- Assign Week 2-3 tasks
- **Deliverable**: All backend engineers understand their role and Week 2 tasks

**Frontend Team Breakout** (3 people + Frontend Lead):
- Review React + TypeScript setup
- Discuss component architecture
- Plan responsive design strategy
- Assign Week 2-3 tasks
- **Deliverable**: All frontend engineers understand their role and Week 2 tasks

**QA Team Breakout** (1.5 people + QA Lead):
- Review testing strategy from 08_MVP_Testing_Plan.md
- Discuss automation approach
- Plan testing tools setup
- **Deliverable**: QA team ready to support development

**DevOps Breakout** (1 person + DevOps Lead):
- Review infrastructure requirements
- Plan Docker & Kubernetes strategy
- Discuss CI/CD approach
- **Deliverable**: DevOps understands infrastructure plan

**PM/Design/Ops Breakout** (3+ people + PM):
- Review project management approach (Jira, sprints)
- Design tool setup (Figma)
- Communication channels (Slack)
- **Deliverable**: All support functions ready to operate

### Afternoon (14:00-17:30)

#### 14:00-15:30: Development Environment Setup Workshop
- **Type**: Hands-on workshop
- **Facilitators**: DevOps Lead + 2 Senior Engineers
- **Attendees**: All 9 engineers (backend + frontend + devops)
- **Tasks**:
  - Verify Docker installation on all machines
  - Run `docker-compose up -d` (from project directory)
  - Verify all containers running: `docker-compose ps`
  - Connect to PostgreSQL: `psql -h localhost -U pintuotuo -d pintuotuo_db`
  - Configure Git: `git config user.name "Your Name"` + email
  - Install IDE extensions (ESLint, Prettier, etc.)
  - Create feature branch: `git checkout -b feature/setup-[yourname]`
- **Reporting**: Each person confirms "✅ Environment ready" in Slack #general
- **Support**: DevOps + senior engineers available to help troubleshoot

#### 15:30-16:30: Git Repository & CI/CD Setup
- **Type**: Technical configuration
- **Facilitator**: DevOps Lead
- **Attendees**: All engineers (demo + hands-on)
- **Tasks**:
  - Clone repo with SSH or HTTPS
  - Verify branch protection rules (main/develop)
  - Create first feature branch
  - Make first commit with proper message format
  - Push to remote
  - Verify CI/CD pipeline triggers
- **Outcome**: Git workflow understood, first commit successful

#### 16:30-17:30: Project Tools Kickoff
- **Type**: Configuration + training
- **Facilitator**: Project Manager
- **Attendees**: All team members (can be async for some)
- **Tasks**:
  - Jira: Show board, explain sprint structure, assign tasks
  - Slack: Show channels, pin important docs, explain conventions
  - Figma: Share workspace link, explain access
  - GitHub/GitLab: Show repo structure, explain wikis/discussions
  - Status page: Share link if monitoring is set up
- **Outcome**: All team members can access and use tools

### EOD Monday Checklist
**PM to verify**:
- [ ] All 15.5 team members attended kickoff
- [ ] All 9 engineers have working Docker environment
- [ ] All team members joined Slack
- [ ] Git repo accessible to all
- [ ] All tools (Jira, Figma) accessible
- [ ] Excitement level high, no major concerns raised

**Action if blocked**: Call CTO + PM + relevant lead for 30-min resolution call

---

## 📅 Tuesday 2026-03-18 - Day 2: Architecture & Standards

### Morning (09:00-12:30)

#### 09:15-09:30: Daily Standup
- Each person: Yesterday (setup), Today (learn), Blocker (none expected)
- Duration: 15 min
- Format: By team (backend, frontend, ops)

#### 09:30-11:00: Architecture Deep Dive
- **Type**: Presentation + Q&A
- **Presenter**: CTO
- **Attendees**: All engineers + PM
- **Topics**:
  - 7-layer architecture overview
  - Tech stack rationale
  - Core modules and responsibilities
  - Data flows for key scenarios
  - High availability and scalability
- **Materials**: 05_Technical_Architecture_and_Tech_Stack.md (slides)
- **Duration**: 90 minutes
  - Slides & explanation: 60 min
  - Q&A: 30 min
- **Output**: All engineers understand the system design
- **Recording**: Save for people who miss it

#### 11:00-12:30: Technical Decision Documentation (ADR)
- **Type**: Technical meeting
- **Attendees**: CTO + Backend Lead + Frontend Lead + DevOps Lead
- **Tasks**:
  - Create ADR (Architecture Decision Record) document
  - Lock final decisions:
    - Database: PostgreSQL 15
    - Cache: Redis single-node
    - Message queue: Kafka (not RabbitMQ)
    - API Gateway: Kong (not nginx)
    - Frontend framework: React + TypeScript (not Vue)
    - Deployment: Docker + Kubernetes (not traditional VMs)
  - Document rationale for each decision
  - Save to: `docs/architecture/ADR.md`
- **Output**: No more architectural debates, decisions locked

### Afternoon (14:00-17:30)

#### 14:00-15:30: Code Standards & Quality Training
- **Type**: Training session
- **Facilitators**: Backend Lead + Frontend Lead
- **Attendees**: All engineers
- **Topics** (reference 13_Dev_Git_Workflow_Code_Standards.md):
  - Commit message format: `<type>(<scope>): <subject>`
  - PR process: min 1 approval, CI must pass
  - Code style: 2-space indent, max 100 chars, no console.log
  - File naming: kebab-case for files, camelCase for functions
  - Component patterns: functional components + hooks, TypeScript required
  - Testing: unit tests, integration tests, acceptance criteria
- **Materials**: document 13 (Git Workflow & Code Standards)
- **Outcome**: All engineers understand standards, ready to code

#### 15:30-17:30: Git Flow & CI/CD Hands-On
- **Type**: Hands-on workshop
- **Facilitator**: DevOps Lead
- **Attendees**: All 9 engineers
- **Hands-on Tasks** (each person does these):
  1. Create feature branch: `git checkout -b feature/test-[yourname]`
  2. Make a change to a file
  3. Stage and commit: `git add . && git commit -m "test(setup): verify git workflow"`
  4. Push to remote: `git push origin feature/test-[yourname]`
  5. Create pull request (in GitHub UI)
  6. Wait for CI/CD to run (should be automatic)
  7. Get approval from a colleague
  8. Merge PR
  9. Clean up local branch
- **Outcome**: Everyone has done a full PR workflow, CI/CD verified
- **Celebration**: First commit + PR for each person! 🎉

### EOD Tuesday Checklist
**CTO to verify**:
- [ ] All architects fully presented
- [ ] All 60+ engineers understand system design
- [ ] ADR document created and signed
- [ ] Code standards agreed and documented
- [ ] Everyone has created first PR successfully
- [ ] CI/CD pipeline working

---

## 📅 Wednesday 2026-03-19 - Day 3: Database & API

### Morning (09:00-12:30)

#### 09:15-09:30: Daily Standup
- Quick check-in on any overnight issues
- Confirm database setup schedule

#### 09:00-10:30: Database Schema Review & Finalization
- **Type**: Technical review
- **Attendees**: Database Architect + Backend Leads + PM
- **Duration**: 1.5 hours
- **Deliverable**: Final database schema locked
- **Review checklist**:
  - [ ] All 12 tables reviewed (from 03_Data_Model_Design.md)
  - [ ] Foreign key relationships confirmed
  - [ ] Primary key strategy confirmed (UUID vs auto-increment)
  - [ ] Indexing strategy finalized
  - [ ] Partition strategy for orders table (by date)
  - [ ] Null/not-null constraints reviewed
  - [ ] Default values defined
  - [ ] Schema ready for SQL generation

#### 10:30-12:30: Database Setup - Local Environment
- **Type**: Hands-on implementation
- **Facilitators**: DevOps Lead + Senior Backend Engineer
- **Attendees**: Backend team + DevOps
- **Tasks**:
  1. Run `docker-compose up -d` (if not already running)
  2. Verify PostgreSQL running: `docker ps | grep postgres`
  3. Connect to database: `psql -h localhost -U pintuotuo -d pintuotuo_db`
  4. Create initial schema (placeholder for now):
     ```sql
     CREATE TABLE IF NOT EXISTS schema_version (
       version_id SERIAL PRIMARY KEY,
       description VARCHAR(255),
       installed_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
     );
     INSERT INTO schema_version VALUES (1, 'Initial schema');
     ```
  5. Verify table created: `\dt`
  6. Load seed data (test users, products):
     ```bash
     psql -h localhost -U pintuotuo -d pintuotuo_db < scripts/db/init.sql
     ```
  7. Verify data: `SELECT COUNT(*) FROM schema_version;`
  8. Test backup procedure
  9. Document connection string
- **Outcome**: Database running, seed data loaded, backup tested

### Afternoon (14:00-17:30)

#### 14:00-15:30: API Specification Review
- **Type**: Technical review
- **Attendees**: Backend leads + Frontend leads + PM
- **Duration**: 1.5 hours
- **Review all 60+ endpoints** from 04_API_Specification.md:
  - C-end endpoints: /api/products, /api/users, /api/groups, /api/orders, /api/auth
  - B-end endpoints: /api/merchants, /api/skus, /api/dashboard
  - Admin endpoints: /api/admin/users, /api/admin/disputes
- **For each endpoint verify**:
  - [ ] Request format clear (body, query params, path params)
  - [ ] Response format consistent
  - [ ] Error responses defined
  - [ ] HTTP status codes correct
  - [ ] Authentication/authorization clear
  - [ ] Rate limiting defined
- **Outcome**: API specification approved, no changes later

#### 15:30-17:30: Mock API Server Creation
- **Type**: Implementation
- **Implementer**: Junior Backend Engineer (with support from senior)
- **Tools**: JSON Server or Prism
- **Tasks**:
  1. Create mock API directory structure
  2. Install JSON Server: `npm install -g json-server`
  3. Create `db.json` with mock data:
     ```json
     {
       "products": [
         {"id": 1, "name": "Token A", "price": 100, "groupPrice": 60},
         {"id": 2, "name": "Token B", "price": 200, "groupPrice": 120}
       ],
       "users": [
         {"id": 1, "email": "user1@test.com", "name": "Test User"}
       ]
     }
     ```
  4. Start mock server: `json-server --watch db.json --port 3001`
  5. Test endpoints:
     ```bash
     curl http://localhost:3001/products
     curl http://localhost:3001/users
     ```
  6. Verify CORS headers allow localhost:3000
  7. Add mock server to docker-compose.yml
  8. Create README for mock API usage
- **Outcome**: Mock API running, frontend can start building

### EOD Wednesday Checklist
**Backend Lead to verify**:
- [ ] PostgreSQL initialized with 12 tables (schema placeholder ok)
- [ ] Seed data loaded and queryable
- [ ] API specification approved and locked
- [ ] Mock API server running on localhost:3001
- [ ] Frontend team can access and use mock API

---

## 📅 Thursday 2026-03-20 - Day 4: Frontend & Design

### Morning (09:00-12:30)

#### 09:15-09:30: Daily Standup
- Check on any database or API issues
- Confirm frontend setup on schedule

#### 09:00-10:00: React Project Initialization
- **Type**: Hands-on implementation
- **Implementer**: Senior Frontend Engineer
- **Tasks**:
  1. Create React + TypeScript project: `npm create vite@latest pintuotuo-frontend -- --template react-ts`
  2. Navigate to project: `cd frontend`
  3. Install dependencies: `npm install`
  4. Install core packages: `npm install react-router zustand axios antd`
  5. Create tsconfig.json with path aliases
  6. Create .env.development with API URLs
  7. Verify project builds: `npm run build`
  8. Verify dev server runs: `npm run dev` (should open http://localhost:3000)
  9. Commit to Git with proper message
- **Outcome**: React project compiling and running locally

#### 10:00-12:30: Design System & Component Library Setup
- **Type**: Implementation + Design
- **Team**: Senior Frontend Engineer + Designer
- **Tasks**:
  1. Install Storybook: `npx storybook@latest init`
  2. Configure TypeScript support in Storybook
  3. Define design tokens (colors, typography, spacing):
     ```typescript
     // src/theme/colors.ts
     export const colors = {
       primary: '#007AFF',
       secondary: '#F2B900',
       danger: '#FF3B30',
       // ... more colors
     };
     ```
  4. Create base components (5-6 minimum):
     - Button.tsx + Button.stories.tsx
     - Input.tsx + Input.stories.tsx
     - Card.tsx + Card.stories.tsx
     - Modal.tsx + Modal.stories.tsx
     - Badge.tsx + Badge.stories.tsx
  5. Start Storybook: `npm run storybook`
  6. Verify components display: http://localhost:6006
  7. Commit components to Git
- **Outcome**: 5+ base components in Storybook, design system defined

### Afternoon (14:00-17:30)

#### 14:00-16:00: UI Design Mockups
- **Type**: Design work
- **Designer**: UI/UX Designer
- **Tools**: Figma
- **Deliverables**: High-fidelity mockups for:
  - [ ] Login page (email/password form)
  - [ ] Home page (product feed, filters)
  - [ ] Product detail page (product info, group options)
  - [ ] Group detail page (members, countdown, join action)
  - [ ] Order summary page (items, price, discount)
  - [ ] User profile page
- **Details**:
  - Mobile, tablet, desktop layouts
  - Color palette from design system
  - Proper spacing and typography
  - Annotations for developers
- **Outcome**: Figma workspace with 6+ page mockups ready for handoff

#### 16:00-17:30: Frontend Architecture Planning
- **Type**: Planning & documentation
- **Planner**: Senior Frontend Engineer
- **Team**: Frontend leads
- **Deliverable**: `docs/frontend-architecture.md` with:
  - [ ] Route structure (pages, paths)
  - [ ] Component hierarchy (pages > sections > components)
  - [ ] Zustand stores structure (auth, products, orders, ui)
  - [ ] API services organization
  - [ ] Folder structure and naming conventions
  - [ ] State management patterns
  - [ ] Data flow diagrams
- **Outcome**: Clear architecture document for Week 2 development

### EOD Thursday Checklist
**Frontend Lead to verify**:
- [ ] React project initialized and building
- [ ] 5+ base components created and in Storybook
- [ ] Design system colors, typography, spacing defined
- [ ] Storybook running at localhost:6006
- [ ] 6+ UI mockups created in Figma
- [ ] Frontend architecture documented

---

## 📅 Friday 2026-03-21 - Day 5: Finalization & Planning

### Morning (09:00-12:30)

#### 09:15-09:30: Daily Standup
- Final check on all deliverables
- Any last-minute items

#### 09:30-10:30: Documentation Creation
- **Type**: Writing + coordination
- **Team**: DevOps Lead + Backend Lead + Senior Engineers
- **Deliverables**:
  - [ ] Setup guide updated (with real commands, screenshots)
  - [ ] Architecture overview document
  - [ ] Git flow walkthrough
  - [ ] Troubleshooting guide
  - [ ] Team communication guide (Slack conventions, office hours)
- **Location**: `docs/` directory
- **Outcome**: New team members can onboard themselves

#### 10:30-12:30: Week 2-3 Task Planning
- **Type**: Planning meeting
- **Attendees**: All team leads + PM
- **Duration**: 2 hours
- **Agenda**:
  1. Review Week 1 achievements
  2. Assign Week 2 tasks (database + API implementation):
     - Database architect: Schema migration system
     - Backend engineer 1: User service API
     - Backend engineer 2: Product service API
     - Backend engineer 3: Group service API
     - Backend engineer 4: Order/Payment services
     - DevOps: API Gateway setup, monitoring
  3. Assign Week 3 tasks (frontend setup):
     - Frontend engineer 1: Auth pages (login, register)
     - Frontend engineer 2: Product pages (home, detail)
     - Frontend engineer 3: Order/Group pages
     - Designer: Final design specs
  4. Create Jira epics for Week 2-3
  5. Populate stories with acceptance criteria
  6. Set story points
  7. Assign to individuals
  8. Identify dependencies
- **Outcome**: Jira board populated, everyone knows Week 2 tasks

### Afternoon (14:00-17:30)

#### 14:00-15:00: Team Retrospective
- **Type**: Retrospective meeting
- **Attendees**: All 15.5 team members
- **Duration**: 1 hour
- **Format**:
  1. What went well this week? (5 min discussion)
  2. What could we improve? (5 min discussion)
  3. Shoutouts to team members who helped (5 min)
  4. Celebration of Week 1 completion! 🎉 (10 min)
  5. Preview of Week 2 (15 min)
  6. Q&A (15 min)
- **Outcome**: Team bonding, continuous improvement, excitement for Week 2

#### 15:00-17:30: Final Verification & Go/No-Go
- **Type**: Verification checklist
- **Attendees**: CTO + PM + All Team Leads
- **Checklist**:
  - [ ] All developers: "Environment ready ✅" in Slack
  - [ ] Git repo: Accessible, CI/CD working
  - [ ] Database: PostgreSQL running, tables created (schema ok, details later)
  - [ ] Mock API: Running on localhost:3001, returning data
  - [ ] React project: Compiling, running on localhost:3000
  - [ ] Design system: Storybook running with 5+ components
  - [ ] Architecture: ADR document created
  - [ ] Code standards: CONTRIBUTING.md completed
  - [ ] Week 2 Jira tasks: Created and assigned
  - [ ] Team alignment: Retrospective positive feedback
  - [ ] No critical blockers
- **Decision**: "✅ Ready for Week 2" or "⚠️ Address blockers before proceeding"
- **Outcome**: CTO sign-off on Week 1 completion

### Final Wrap-up (17:30)

**Celebrate!** 🎉
- Week 1 complete, team fully onboarded
- Foundation solid for rapid development
- Everyone understands their role
- Next Monday: Core implementation begins

**Send team message**:
```
✅ Week 1 Complete!

Achievements:
- All 15.5 team members onboarded ✅
- Development environment working for all ✅
- Git workflow & code standards locked ✅
- Architecture reviewed & documented ✅
- Database initialized ✅
- API specification finalized ✅
- Mock API running ✅
- React project initialized ✅
- Design system created ✅
- UI mockups completed ✅

Ready for Week 2: Database & API Implementation!

Have a great weekend. See you Monday 09:15 for Week 2 standups. 🚀
```

---

## 🎯 Success Criteria (Week 1 Completion)

**Technical**:
- ✅ All developers have working Docker environment
- ✅ PostgreSQL running with schema (placeholder ok)
- ✅ Mock API accessible on localhost:3001
- ✅ React + TypeScript building on localhost:3000
- ✅ Storybook running with 5+ components
- ✅ Git Flow working, CI/CD pipeline active
- ✅ Architecture & ADR documented

**Team**:
- ✅ All 15.5 people onboarded and engaged
- ✅ Roles and responsibilities clear
- ✅ Code standards agreed upon
- ✅ Communication channels established
- ✅ Team morale positive (retrospective feedback)

**Planning**:
- ✅ Week 2-3 tasks created and assigned in Jira
- ✅ Design system defined (colors, typography, spacing)
- ✅ UI mockups for all key pages
- ✅ API specification finalized and reviewed

**If ANY of above are not true**:
- CTO + PM assess severity
- Either address in Friday afternoon session (17:30-18:30)
- Or delay Week 2 start until resolved
- Document root cause and prevention for future weeks

---

## 📞 Daily Meeting Times (Week 1)

| Time | Meeting | Owner | Duration | Attendees |
|------|---------|-------|----------|-----------|
| 09:15 | Daily standup | Team lead | 15 min | By team |
| 09:15 | Project kickoff (Mon only) | PM | 45 min | All |
| 10:15 | Breakout sessions (Mon-Tue) | Team lead | 2+ hours | By team |
| 14:00 | Hands-on sessions | Lead + team | 1-2 hours | Engineers |
| 16:00 | Retrospective (Fri) | PM | 1 hour | All |
| 15:00 | Final verification (Fri) | CTO | 2.5 hours | Leads |

---

## 🆘 Support & Escalation

**If blocked**:
1. Post in #blockers Slack immediately
2. Tag your team lead
3. If cross-team: PM or CTO resolves
4. Max 30 min response time target

**Common issues & solutions**:
- Docker not running: "Docker Desktop" needs to start on Mac/Windows
- Port already in use: `lsof -i :5432` to find and kill process
- Git auth issues: Generate SSH key or use HTTPS
- npm dependencies: `npm cache clean --force && npm install`

**PM contact**: Available for blocking issues
**CTO contact**: Architecture/tech decisions
**Team lead**: Day-to-day questions

---

## ✨ What's Next After Week 1?

**Monday 2026-03-24 (Week 2 Begins)**:
- Backend: Database schema finalization, API implementation
- Frontend: Start building pages using components
- QA: Create test cases, set up automation
- All: Continue daily standups, maintain momentum

**Focus**: Build, build, build! 💪

---

**Good luck, team! Let's make Week 1 a success!** 🚀

*Questions? Check 17_Master_Execution_Plan_Complete_Overview.md or ask your team lead.*

---

**Created**: 2026-03-14
**Last Updated**: 2026-03-14
**Status**: ✅ Ready for Monday Execution
