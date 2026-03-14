# Week 2: Database & API Design Finalization - Detailed Execution Guide

**Document ID**: 14_Plan_Week2_Database_and_API_Design
**Version**: 1.0
**Status**: Active
**Timeline**: 5 working days (Week 2)
**Owner**: Backend Lead / Database Architect

---

## 📋 Week 2 Overview

**Objective**: Finalize database schema implementation, complete API interface design, establish mock API server for parallel frontend development

**Key Deliverables**:
- [ ] PostgreSQL database fully initialized with all 12 tables
- [ ] Database schema documentation and migration scripts
- [ ] Complete API specification with request/response examples
- [ ] Mock API server running and accessible
- [ ] API documentation and Postman collection
- [ ] Performance optimization plan (indexing, caching)

**Total Team Hours**: ~100 hours
**Key Milestone**: End of Friday = Ready for frontend integration

---

## 📅 Daily Breakdown

### Monday - Database Schema Finalization & Validation

#### Morning Session (09:00-12:30)

**1. Database Schema Review & Optimization** (09:00-10:30)
- Attendees: Database Architect, Backend Lead, Backend Team (4 people)
- Duration: 1.5 hours
- Tasks:
  - [ ] Review all 12 tables from 03_Data_Model_Design.md
  - [ ] Validate relationships and foreign key constraints
  - [ ] Review field types and default values
  - [ ] Discuss indexing strategy:
    - Primary key indexes (automatic)
    - Foreign key indexes for joins
    - Composite indexes for common queries (user_id, created_at)
    - Full-text search indexes for product descriptions
  - [ ] Plan for future scaling:
    - Table partitioning strategy (orders by date)
    - Archive strategy for old data
  - [ ] Verify data integrity constraints
- Output: Optimized schema ready for implementation
- Owner: Database Architect

**2. Migration Scripts & Versioning** (10:30-12:30)
- Attendees: Backend Lead + 1 Senior Backend Engineer
- Duration: 2 hours
- Tasks:
  - [ ] Create database migration tool setup (Flyway or Liquibase)
  - [ ] Write migration script: V001__initial_schema.sql
    - Create all 12 tables with proper DDL
    - Add all indexes
    - Add check constraints
  - [ ] Write seed data migration: V002__seed_initial_data.sql
    - Test merchants (at least 3)
    - Test product SKUs
    - Test users
  - [ ] Document migration process and rollback procedure
  - [ ] Test migrations locally (up and rollback)
- Output: Working migration system with version control
- Owner: Backend Lead

#### Afternoon Session (14:00-17:30)

**3. PostgreSQL Container & Local Database Setup** (14:00-15:30)
- Attendees: DevOps Lead + Backend Team
- Duration: 1.5 hours
- Tasks:
  - [ ] Create Dockerfile for PostgreSQL (or use official image)
  - [ ] Add to docker-compose.yml:
    - PostgreSQL service
    - Environment variables (DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD)
    - Volume mounting for data persistence
    - Port mapping (5432:5432)
  - [ ] Create initialization scripts directory
  - [ ] Test docker-compose up with database
  - [ ] Verify all containers running: `docker-compose ps`
- Output: PostgreSQL running in Docker locally
- Owner: DevOps Lead

**4. Database Connection & Testing** (15:30-17:30)
- Attendees: Backend Team + DevOps Lead
- Duration: 2 hours
- Tasks:
  - [ ] Connect to local PostgreSQL: `psql -h localhost -U pintuotuo -d pintuotuo_db`
  - [ ] Run migration scripts
  - [ ] Verify all 12 tables created: `\dt`
  - [ ] Verify seed data loaded:
    ```sql
    SELECT COUNT(*) FROM users;
    SELECT COUNT(*) FROM merchants;
    SELECT COUNT(*) FROM products;
    ```
  - [ ] Test sample queries:
    - User with orders: `SELECT u.*, COUNT(o.id) as order_count FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id;`
    - Product with group info: `SELECT p.*, COUNT(g.id) as active_groups FROM products p LEFT JOIN groups g ON p.id = g.product_id WHERE g.status = 'active' GROUP BY p.id;`
  - [ ] Set up backup strategy:
    - Nightly automatic backups
    - Backup location and retention
  - [ ] Document backup/restore procedures
- Output: Verified database with seed data + backup strategy
- Owner: Backend Lead

**EOD Monday Checklist**:
- [ ] All 12 tables created in PostgreSQL
- [ ] Seed data loaded and verified
- [ ] Migration system working
- [ ] Backup strategy documented
- [ ] All developers can connect to local database

---

### Tuesday - API Specification Finalization & Documentation

#### Morning Session (09:00-12:30)

**1. API Endpoint Specification Review** (09:00-11:00)
- Attendees: Backend Lead, Frontend Lead, PM
- Duration: 2 hours
- Review all 60+ endpoints from 04_API_Specification.md:
  - [ ] User endpoints (authentication, profile, settings)
  - [ ] Product endpoints (browse, search, filter)
  - [ ] Order endpoints (create, list, detail)
  - [ ] Group endpoints (create, join, list, status)
  - [ ] Token endpoints (consume, balance check)
  - [ ] Payment endpoints (create, webhook)
  - [ ] Merchant endpoints (if time permits)
- For each endpoint, verify:
  - [ ] Request parameters (query, body, path)
  - [ ] Response format and status codes
  - [ ] Error responses
  - [ ] Authentication requirements (JWT token needed?)
  - [ ] Rate limiting considerations
- Output: Approved API specification
- Owner: Backend Lead

**2. API Documentation Generation** (11:00-12:30)
- Attendees: 1 Senior Backend Engineer
- Duration: 1.5 hours
- Tasks:
  - [ ] Set up Swagger/OpenAPI documentation
  - [ ] Convert API spec to OpenAPI 3.0 format
  - [ ] Generate interactive API documentation
  - [ ] Add curl examples for common endpoints
  - [ ] Add request/response schemas
  - [ ] Generate Postman collection (for team testing)
  - [ ] Deploy documentation to internal wiki or GitHub Pages
- Output: Interactive API docs + Postman collection
- Owner: Senior Backend Engineer

#### Afternoon Session (14:00-17:30)

**3. API Request/Response Schemas** (14:00-15:30)
- Attendees: Backend Lead + 1 Backend Engineer
- Duration: 1.5 hours
- Tasks:
  - [ ] Define JSON schemas for all request bodies
  - [ ] Define JSON schemas for all response objects
  - [ ] Create error response format:
    ```json
    {
      "code": "ERROR_CODE",
      "message": "Human readable message",
      "details": {}
    }
    ```
  - [ ] Define pagination format for list endpoints:
    ```json
    {
      "data": [],
      "pagination": {
        "current_page": 1,
        "page_size": 20,
        "total_count": 100,
        "total_pages": 5
      }
    }
    ```
  - [ ] Create reusable schema components (User, Product, Order, Group, etc.)
  - [ ] Document enum values (order status, group status, etc.)
- Output: Comprehensive JSON schemas for all endpoints
- Owner: Backend Lead

**4. Authentication & Authorization Design** (15:30-17:30)
- Attendees: Backend Lead + Security-minded engineer
- Duration: 2 hours
- Tasks:
  - [ ] Design JWT token structure:
    - Payload: user_id, email, roles, iat, exp
    - Expiry: access token (1 hour), refresh token (30 days)
  - [ ] Design authorization levels:
    - Public endpoints (no auth required)
    - User endpoints (requires valid JWT)
    - Merchant endpoints (requires merchant role)
    - Admin endpoints (requires admin role)
  - [ ] Design API key authentication for merchant API integration:
    - API Key format
    - Signature generation
    - Rate limiting per API key
  - [ ] Plan token refresh flow
  - [ ] Plan logout strategy (token blacklist or just client-side discard)
  - [ ] Document security headers (CORS, CSP, etc.)
- Output: Security design document
- Owner: Backend Lead

**EOD Tuesday Checklist**:
- [ ] All 60+ API endpoints documented
- [ ] OpenAPI/Swagger setup complete
- [ ] Postman collection generated
- [ ] JSON schemas defined for all endpoints
- [ ] Authentication design finalized

---

### Wednesday - Mock API Server Implementation

#### Morning Session (09:00-12:30)

**1. Mock Server Setup** (09:00-10:00)
- Attendees: 1 Backend Engineer (junior or mid-level)
- Duration: 1 hour
- Technology choice: JSON Server or Prism (OpenAPI-based)
- Tasks:
  - [ ] Create mock server project directory: `/mock-api`
  - [ ] Install dependencies:
    - JSON Server: `npm install -g json-server`
    - Or: Prism CLI: `npm install -g @stoplight/prism-cli`
  - [ ] Create mock data file with realistic data
  - [ ] Configure CORS for localhost:3000 access
  - [ ] Set up port on 3001
  - [ ] Add to docker-compose.yml
- Output: Mock API server project structure
- Owner: Backend Engineer

**2. Mock Data Generation** (10:00-12:30)
- Attendees: Backend Engineer + 1 other
- Duration: 2.5 hours
- Tasks:
  - [ ] Create realistic mock data for all tables:
    - 10 test users with varied attributes
    - 5 test merchants with different products
    - 20 test products (various prices, token amounts, categories)
    - 5 test groups (different statuses: forming, active, completed, failed)
    - 10 test orders (various statuses)
  - [ ] Use seeded data generator for consistency
  - [ ] Ensure data relationships are valid:
    - Orders reference valid users and products
    - Groups reference valid products
    - All IDs are properly linked
  - [ ] Generate data at realistic volumes (not 10,000 records, but enough for testing)
  - [ ] Create separate datasets for different test scenarios:
    - Successful purchase flow
    - Failed group scenario
    - Token consumption scenario
- Output: Complete mock dataset
- Owner: Backend Engineer

#### Afternoon Session (14:00-17:30)

**3. Mock API Endpoint Implementation** (14:00-16:30)
- Attendees: Backend Engineer
- Duration: 2.5 hours
- Tasks:
  - [ ] Implement all C-end endpoints:
    - GET /products (list, filter, pagination)
    - GET /products/:id (detail)
    - POST /users/register
    - POST /users/login
    - GET /users/profile
    - POST /groups (create)
    - GET /groups/:id (detail)
    - POST /groups/:id/join (join group)
    - POST /orders (create order)
    - GET /orders (list user orders)
  - [ ] Implement stateful mock responses:
    - Group formation logic (when 3 members, mark as completed)
    - Order status progression (pending → paid → completed)
    - Token deduction on API call
  - [ ] Add realistic response delays (100-200ms) to simulate network
  - [ ] Add mock payment success/failure scenarios
- Output: Working mock API endpoints
- Owner: Backend Engineer

**4. Testing & Documentation** (16:30-17:30)
- Attendees: Backend Engineer + QA
- Duration: 1 hour
- Tasks:
  - [ ] Test all endpoints with curl or Postman:
    ```bash
    curl http://localhost:3001/api/products
    curl -X POST http://localhost:3001/api/users/register -d '...'
    ```
  - [ ] Verify CORS headers allow localhost:3000
  - [ ] Test with frontend developer (quick manual test)
  - [ ] Document mock server usage: "How to start, what endpoints available, sample requests"
  - [ ] Add README to mock-api directory
- Output: Tested and documented mock API
- Owner: Backend Engineer

**EOD Wednesday Checklist**:
- [ ] Mock API server running on localhost:3001
- [ ] All critical endpoints returning mock data
- [ ] CORS configured for frontend access
- [ ] Realistic mock data loaded
- [ ] Mock API documented and tested

---

### Thursday - API Gateway & Routing Setup

#### Morning Session (09:00-12:30)

**1. API Gateway Selection & Setup** (09:00-10:30)
- Attendees: Backend Lead, DevOps Lead
- Duration: 1.5 hours
- Decision: Kong vs nginx vs Spring Cloud Gateway
- Recommendation: Kong (API-first, great admin UI)
- Tasks:
  - [ ] Add Kong service to docker-compose.yml
  - [ ] Configure Kong container:
    - Port mapping (8000:8000 for API, 8001:8001 for admin)
    - Database (use PostgreSQL backend)
    - Environment setup
  - [ ] Create Kong admin UI Docker service (optional but helpful)
  - [ ] Test Kong startup: `docker-compose up kong`
  - [ ] Verify Kong admin API responds: `curl http://localhost:8001`
- Output: Kong API Gateway running locally
- Owner: DevOps Lead

**2. Route Configuration** (10:30-12:30)
- Attendees: Backend Lead + DevOps Lead
- Duration: 2 hours
- Tasks:
  - [ ] Create Kong routes configuration script
  - [ ] Define routes for key services:
    ```
    /api/users/* → http://user-service:8081
    /api/products/* → http://product-service:8082
    /api/orders/* → http://order-service:8083
    /api/groups/* → http://group-service:8084
    /api/tokens/* → http://token-service:8085
    /api/payments/* → http://payment-service:8086
    ```
  - [ ] Enable rate limiting plugin (100 req/min per user)
  - [ ] Enable authentication plugin (JWT validation)
  - [ ] Configure CORS headers
  - [ ] Add request/response logging
  - [ ] Test routing with curl:
    ```bash
    curl http://localhost:8000/api/products
    ```
- Output: Working API Gateway with configured routes
- Owner: DevOps Lead

#### Afternoon Session (14:00-17:30)

**3. Middleware & Interceptors** (14:00-15:30)
- Attendees: Backend Lead
- Duration: 1.5 hours
- Tasks (design phase, implementation in Week 3):
  - [ ] Design request/response interceptor pattern
  - [ ] Plan middleware layers:
    - Request logging (all requests)
    - Authentication (validate JWT)
    - Authorization (check permissions)
    - Request validation (validate schema)
    - Request ID tracking (for tracing)
  - [ ] Design error handling middleware
  - [ ] Design response formatting middleware (consistent response structure)
  - [ ] Plan for distributed tracing (Jaeger/Zipkin)
- Output: Middleware architecture document
- Owner: Backend Lead

**4. Documentation & Integration Testing** (15:30-17:30)
- Attendees: Backend Engineer + Frontend Lead
- Duration: 2 hours
- Tasks:
  - [ ] Document API Gateway setup in wiki/README
  - [ ] Create postman collection with gateway endpoints
  - [ ] Brief frontend team on API endpoint structure
  - [ ] Test mock API through gateway (if services set up):
    ```bash
    curl http://localhost:8000/api/products  # through Kong
    curl http://localhost:3001/api/products  # direct to mock API
    ```
  - [ ] Coordinate with frontend team on API base URL for dev environment
  - [ ] Plan integration testing schedule
- Output: Documented API Gateway + team alignment
- Owner: Backend Lead

**EOD Thursday Checklist**:
- [ ] API Gateway (Kong) running and accessible
- [ ] Routes configured for all microservices
- [ ] Rate limiting and authentication enabled
- [ ] API Gateway tested and documented
- [ ] Frontend team briefed on endpoints

---

### Friday - Integration Testing & Week 3 Planning

#### Morning Session (09:00-12:30)

**1. End-to-End Testing** (09:00-10:30)
- Attendees: Backend Team + QA Lead
- Duration: 1.5 hours
- Test scenarios:
  - [ ] User registration → login → get profile (user service flow)
  - [ ] Browse products → get product detail (product service flow)
  - [ ] Create group → join group → check members (group service flow)
  - [ ] Place order → check order status (order service flow)
  - [ ] All requests go through API Gateway
  - [ ] Error scenarios (invalid input, 404, auth failure)
  - [ ] Response format consistency (all responses have proper structure)
- Output: Test report with pass/fail status
- Owner: QA Lead

**2. Performance Baseline Testing** (10:30-12:30)
- Attendees: Backend Lead + 1 Engineer
- Duration: 2 hours
- Tasks:
  - [ ] Set up load testing tools (Apache JMeter or k6)
  - [ ] Create simple load test:
    - 10 concurrent users
    - 100 requests to each endpoint
    - Measure response time (target: < 200ms)
  - [ ] Test database connection pooling:
    - Verify no connection exhaustion
    - Monitor pool usage
  - [ ] Document baseline metrics:
    - Avg response time: __ms
    - 95th percentile: __ms
    - Requests/second: __
  - [ ] Identify bottlenecks for optimization
- Output: Performance baseline report
- Owner: Backend Lead

#### Afternoon Session (14:00-17:30)

**3. Database Optimization Review** (14:00-15:00)
- Attendees: Database Architect + Backend Lead
- Duration: 1 hour
- Tasks:
  - [ ] Review query execution plans for slow queries
  - [ ] Verify indexes being used:
    ```sql
    EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 1;
    ```
  - [ ] Check missing indexes and create if needed
  - [ ] Plan caching strategy:
    - Cache frequently queried data (product list, categories)
    - Cache TTL: 15 minutes for product data, 5 min for user data
  - [ ] Create indexing documentation
- Output: Optimized database with performance metrics
- Owner: Database Architect

**4. Week 2 Retrospective & Week 3 Planning** (15:00-17:30)
- Attendees: All Backend Team, DevOps Lead, QA Lead
- Duration: 2.5 hours
- Activities:
  - [ ] Review Week 2 deliverables:
    - Database: ✅ or ❌ ?
    - API specs: ✅ or ❌ ?
    - Mock API: ✅ or ❌ ?
    - Gateway: ✅ or ❌ ?
  - [ ] Identify blockers and issues
  - [ ] Plan Week 3 (Frontend setup phase):
    - Frontend team will build React components
    - Backend team will start API implementation
    - Coordinate on integration points
  - [ ] Assign Week 3 tasks:
    - API service implementation (User, Product, Order, Group, Token services)
    - Database connection pooling setup
    - Error handling implementation
    - Logging setup
  - [ ] Plan integration testing schedule for Week 3
- Output: Week 3 task list + calendar of integration testing
- Owner: Backend Lead + Project Manager

**EOD Friday Checklist**:
- [ ] All 12 database tables fully functional
- [ ] All 60+ API endpoints documented and mocked
- [ ] API Gateway running with working routes
- [ ] Mock API server accessible from frontend
- [ ] Database migration system verified
- [ ] Performance baseline established
- [ ] Week 3 tasks assigned to team members
- [ ] Frontend team ready to start integration

---

## 🎯 Week 2 Deliverables Checklist

### Database
- [ ] PostgreSQL running in Docker
- [ ] All 12 tables created with DDL
- [ ] Foreign key constraints implemented
- [ ] Indexes optimized for performance
- [ ] Seed data loaded
- [ ] Migration system (Flyway/Liquibase) functional
- [ ] Backup strategy documented
- [ ] Database documentation complete

### API Specification
- [ ] All 60+ endpoints documented
- [ ] OpenAPI 3.0 specification created
- [ ] Request/response schemas defined
- [ ] Error handling standardized
- [ ] Authentication design finalized
- [ ] Rate limiting strategy planned
- [ ] API documentation generated and accessible

### Mock API Server
- [ ] Running on localhost:3001
- [ ] All critical C-end endpoints implemented
- [ ] Realistic mock data generated
- [ ] CORS configured for frontend access
- [ ] Response delays simulated
- [ ] Documentation complete
- [ ] Tested and verified working

### API Gateway
- [ ] Kong or alternative running
- [ ] Routes configured for all services
- [ ] Rate limiting enabled
- [ ] Authentication middleware configured
- [ ] Logging and monitoring setup
- [ ] Documentation complete

### Tools & Documentation
- [ ] Postman collection generated
- [ ] Database documentation
- [ ] API gateway configuration documented
- [ ] Migration procedures documented
- [ ] Performance baseline established

---

## 📊 Week 2 Success Metrics

| Metric | Target | Verification |
|--------|--------|---------------|
| Database Tables | 12/12 | All created in PostgreSQL |
| API Endpoints Documented | 60+ | OpenAPI spec complete |
| Mock API Response Rate | 100% | All endpoints return mock data |
| API Documentation | Complete | Accessible to team |
| Average Response Time | <200ms | Load test baseline |
| Test Coverage | Happy path | E2E tests pass |

---

## 🚨 Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| DB schema issues | Medium | High | Review with architect, version control |
| API spec confusion | Medium | High | Frontend/Backend sync meetings |
| Mock API doesn't match real API | Medium | Medium | Keep sync'd with API spec updates |
| Performance issues in DB | Low | High | Index optimization, caching strategy |
| Migration script fails | Low | High | Test rollback procedures |

---

## 📞 Daily Communication

### Morning Standup (09:15-09:30)
- Database progress
- API specification status
- Any blockers from previous day

### Technical Discussions
- Slack: #backend-technical
- Daily sync between Backend Lead and DevOps

### Integration Points
- Mock API demos with Frontend Lead (Wednesday)
- Gateway documentation review (Thursday)

---

## ✅ Go/No-Go Criteria for Week 3

**READY FOR WEEK 3 IF ALL OF THESE ARE TRUE**:

1. ✅ PostgreSQL database fully initialized with all 12 tables
2. ✅ All 12 tables have seed data loaded
3. ✅ Migration system tested and working
4. ✅ All 60+ API endpoints documented in OpenAPI format
5. ✅ Mock API server running on localhost:3001
6. ✅ API Gateway (Kong) running with routes configured
7. ✅ Postman collection generated and working
8. ✅ Frontend team can call mock API endpoints
9. ✅ Database performance baseline established
10. ✅ No critical blockers outstanding

**IF BLOCKED**:
- Backend Lead escalates to CTO
- Allocate Friday 17:30-18:30 for unblocking
- Identify which service can start implementation without being blocked

---

## 🎯 Success Definition

**Week 2 is SUCCESSFUL when:**
- Database is production-ready and tested
- API contract is clear and agreed upon
- Mock API enables frontend development to proceed in parallel
- All backend engineers understand the API structure
- Foundation for API implementation is solid

---

**Owner**: Backend Lead / Database Architect
**Last Updated**: 2026-03-14
**Version**: 1.0
**Status**: Ready for Execution
