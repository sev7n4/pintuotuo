# Documentation Naming Convention & Index

**Last Updated**: 2026-03-14
**Version**: 1.0

---

## 📋 Naming Convention Rules (All New Documents)

### Format: `[Priority][Document_Type]_[Category]_[Description].md`

#### Naming Structure:
```
[Numeric Priority (00-99)]_[Type]_[Category]_[Brief_Title].md

Example:
  ✅ 06_Design_Frontend_Login_Page_Wireframes.md
  ✅ 12_Design_Mobile_Responsive_Layout.md
  ✅ 15_Backend_API_Authentication_Implementation.md
```

### Priority Levels:
- `00-09`: Core Project Documents (PRD, Architecture, etc.)
- `10-19`: Original Requirements & References
- `20-29`: Design Documents (UI/UX, Wireframes)
- `30-39`: Backend Development (API, Database, Services)
- `40-49`: Frontend Development (React, Mobile, etc.)
- `50-59`: DevOps & Infrastructure (Deployment, Monitoring)
- `60-69`: Testing & QA (Test Plans, Test Cases)
- `70-79`: Project Management (Planning, Schedules, Risks)
- `80-89`: Financial & Resource Planning
- `90-99`: Operations & Growth (Marketing, User Growth, etc.)

### Document Types:
- `PRD` - Product Requirements Document
- `Design` - Design & UX Documentation
- `Backend` - Backend Development
- `Frontend` - Frontend Development
- `DevOps` - Infrastructure & Deployment
- `Testing` - QA & Testing
- `API` - API Documentation
- `Database` - Database Schema & Design
- `Plan` - Project Planning
- `Analysis` - Data Analysis & Metrics

### Naming Best Practices:
- Use **snake_case** for multi-word titles
- Keep filenames **short but descriptive** (< 60 characters)
- Use **English only** (no Chinese/special characters)
- Use **hyphens** between words (NOT underscores for readability)
- Avoid **generic names** like "Document" or "Notes"

---

## 📂 Current Document Index

### Core Project Documents (00-09)

| # | Filename | Description | Owner |
|---|----------|-------------|-------|
| 00 | **00_Project_Delivery_Summary.md** | Project delivery checklist & document index | Product Team |
| 01 | **01_PRD_Complete_Product_Specification.md** | Complete PRD with 13 chapters | Product Manager |
| 02 | **02_User_Flow_and_Journey.md** | C-end & B-end user journeys & interaction flows | Product Designer |
| 03 | **03_Data_Model_Design.md** | 12 core data tables & DB schema | Database Architect |
| 04 | **04_API_Specification.md** | 60+ API endpoints with request/response | Backend Lead |
| 05 | **05_Technical_Architecture_and_Tech_Stack.md** | System architecture & technology selection | CTO |

### Project Execution Documents (06-09)

| # | Filename | Description | Owner |
|---|----------|-------------|-------|
| 06 | **06_Project_Launch_and_Milestone_Planning.md** | 8-week MVP execution plan (week-by-week) | Project Manager |
| 07 | **07_UI_UX_Design_Guidelines.md** | Complete design system & interface reference | UI/UX Designer |
| 08 | **08_MVP_Testing_Plan.md** | QA strategy, 30+ test cases & acceptance criteria | QA Lead |
| 09 | **09_Cost_Estimation_and_Resource_Planning.md** | Financial model, budget & resource planning | CFO |

### Original Requirements (10-19)

| # | Filename | Description | Owner |
|---|----------|-------------|-------|
| 10 | **10_Original_Business_Requirements.md** | Original business requirements document | Business Analyst |

---

## 🎯 Document Usage Map

### By Role:

**👨‍💼 Product Manager**
- Read: 01, 02, 08, 06
- Time: 4-5 hours

**👨‍💻 CTO / Tech Lead**
- Read: 05, 03, 04, 06
- Time: 4-5 hours

**🎨 UI/UX Designer**
- Read: 07, 02
- Time: 2-3 hours

**👨‍💻 Backend Engineer**
- Read: 03, 04, 05
- Time: 3-4 hours

**🧪 QA / Test Engineer**
- Read: 08, 03, 01 (functional requirements)
- Time: 3-4 hours

**💰 CFO / Finance**
- Read: 09, 01 (business model)
- Time: 2-3 hours

**📊 Project Manager**
- Read: 06, 00, 09
- Time: 3-4 hours

---

## 📝 New Document Creation Checklist

When creating a new document, follow these steps:

### 1. **Determine Priority & Type**
   - [ ] What category does this document belong to? (00-99)
   - [ ] What type is it? (PRD, Design, API, Plan, etc.)
   - [ ] Is there a similar document already?

### 2. **Create Meaningful Filename**
   - [ ] Format: `[Priority]_[Type]_[Category]_[Description].md`
   - [ ] Example: `21_Design_Mobile_Home_Page_Wireframe.md`
   - [ ] Keep it under 60 characters
   - [ ] Use snake_case for multi-word descriptions

### 3. **Add Document Header**
```markdown
# [Document Title]

**Document ID**: XX_[Type]_[Category]
**Version**: 1.0
**Last Updated**: YYYY-MM-DD
**Owner**: [Name/Team]
**Status**: Draft / In Review / Finalized

---

[Content starts here...]
```

### 4. **Link Back to Index**
   - [ ] Add reference in this index file
   - [ ] Link from related documents

### 5. **Review & Publish**
   - [ ] Spelling & grammar check (English only)
   - [ ] Technical accuracy verified
   - [ ] Links to other documents verified
   - [ ] Share with relevant team members

---

## 🗂️ Directory Structure (Recommended)

```
pintuotuo/
├── README.md                           # Project overview
├── 00_Project_Delivery_Summary.md      # (Core Documents)
├── 01_PRD_Complete_Product_Specification.md
├── 02_User_Flow_and_Journey.md
├── 03_Data_Model_Design.md
├── 04_API_Specification.md
├── 05_Technical_Architecture_and_Tech_Stack.md
├── 06_Project_Launch_and_Milestone_Planning.md
├── 07_UI_UX_Design_Guidelines.md
├── 08_MVP_Testing_Plan.md
├── 09_Cost_Estimation_and_Resource_Planning.md
├── 10_Original_Business_Requirements.md
│
├── /docs/                              # Additional documentation
│   ├── /design/                        # UI/UX wireframes, mockups
│   ├── /architecture/                  # System architecture diagrams
│   ├── /database/                      # Database schemas, migrations
│   ├── /api/                           # API documentation, examples
│   ├── /testing/                       # Test cases, test reports
│   └── /operations/                    # Runbooks, deployment guides
│
├── /code/                              # Source code
│   ├── /backend/                       # Go, Node.js services
│   ├── /frontend/                      # React web app
│   └── /mobile/                        # React Native / Flutter apps
│
└── /assets/                            # Logos, images, templates
```

---

## ✅ Document Status Tracking

### Status Definitions:
- **Draft** - In progress, not ready for review
- **In Review** - Ready for team feedback
- **Approved** - Reviewed and approved by stakeholders
- **Finalized** - Ready for implementation
- **Archived** - No longer in use (keep for reference)

### How to Update Status:
Add to document header:
```markdown
**Status**: Finalized
**Last Review**: 2026-03-14
**Approved By**: [Name], [Name]
**Next Review**: [Date if applicable]
```

---

## 📞 Document Maintenance

### Weekly:
- [ ] Check for broken links between documents
- [ ] Update version numbers if content changed significantly

### Monthly:
- [ ] Review document index for accuracy
- [ ] Archive outdated documents (move to /archive/)
- [ ] Update status of documents

### Quarterly:
- [ ] Full review of all documentation
- [ ] Identify gaps or redundancies
- [ ] Plan documentation improvements

---

## 🔗 Quick Links

**Core Documents**:
- [PRD](01_PRD_Complete_Product_Specification.md)
- [Architecture](05_Technical_Architecture_and_Tech_Stack.md)
- [API Spec](04_API_Specification.md)
- [Project Plan](06_Project_Launch_and_Milestone_Planning.md)

**Design**:
- [UI/UX Guidelines](07_UI_UX_Design_Guidelines.md)
- [User Flows](02_User_Flow_and_Journey.md)

**Development**:
- [Data Model](03_Data_Model_Design.md)
- [Testing Plan](08_MVP_Testing_Plan.md)

**Management**:
- [Cost & Resource](09_Cost_Estimation_and_Resource_Planning.md)
- [Project Summary](00_Project_Delivery_Summary.md)

---

**Last Updated**: 2026-03-14
**Maintained By**: Product Team
**Version**: 1.0
