# Beakon Documentation Structure Guide

**Last Updated**: October 14, 2025
**Purpose**: Guide for understanding and maintaining the Beakon platform documentation

---

## Documentation Philosophy

This documentation structure follows a **single source of truth** principle:
- **One comprehensive reference** at platform root
- **One primary doc** per microservice (README.md)
- **One architecture doc** for complex services (ARCHITECTURE.md)
- **No duplicates, no session-specific reports**

---

## Root Level Documentation (`/Beakon`)

### Essential Platform Documents (6 files)

**1. `README.md`**
- **Purpose**: Main entry point for the entire platform
- **Audience**: All developers, new team members
- **Content**: Quick start, project overview, directory structure
- **When to update**: Major architectural changes, new services

**2. `SERVICE_CATALOG.md`** ⭐ **PRIMARY REFERENCE**
- **Purpose**: Complete catalog of all 20 microservices
- **Audience**: All developers, AI assistants, operations
- **Content**:
  - Service purposes, ports, databases
  - API endpoints for each service
  - Dependencies and relationships
  - Technology stack per service
- **When to update**: New service, port changes, API changes
- **Why this is primary**: Most comprehensive and up-to-date service reference

**3. `DATABASE_ARCHITECTURE.md`**
- **Purpose**: Database schemas and architecture patterns
- **Audience**: Backend developers, database admins
- **Content**: Schema definitions, relationships, multi-tenancy patterns
- **When to update**: Schema changes, new databases

**4. `DEPLOYMENT_GUIDE.md`**
- **Purpose**: Production deployment procedures
- **Audience**: DevOps, SRE, deployment engineers
- **Content**: Deployment steps, configuration, troubleshooting
- **When to update**: Deployment process changes

**5. `OPERATIONAL_RUNBOOK.md`**
- **Purpose**: Incident response and operations guide
- **Audience**: On-call engineers, operations team
- **Content**: Troubleshooting, monitoring, incident response
- **When to update**: New monitoring, common issues discovered

**6. `ARCHITECTURE.md`**
- **Purpose**: High-level system architecture overview
- **Audience**: Architects, senior engineers, new team members
- **Content**: Architecture patterns, design decisions, system flow
- **When to update**: Major architectural decisions

---

## Microservices Directory Level (`/microservices`)

### Coordination Documents (2 files)

**1. `README.md`**
- **Purpose**: Development workflow guide for all services
- **Content**:
  - How to start/stop all services
  - Development scripts (start-dev.sh, stop-dev.sh)
  - Common development tasks
  - Port listing for all services
- **When to update**: New services, script changes

**2. `API_GATEWAY_COMMUNICATION_GUIDE.md`**
- **Purpose**: Inter-service communication patterns
- **Content**:
  - How services communicate through API Gateway
  - Service routing configuration
  - Authentication flow between services
- **When to update**: API Gateway changes, new routing patterns

---

## Service-Level Documentation (`/microservices/<service-name>`)

### Standard Service Structure

Each microservice has **1-2 documents**:

#### All Services Have:
**`README.md`** (Required for all services)
- **Purpose**: Service overview and API reference
- **Content**:
  - Service purpose and key features
  - API endpoints with examples
  - Database schema (if applicable)
  - Configuration and environment variables
  - Port number and dependencies
  - Quick start guide
- **When to update**: API changes, feature additions

#### Complex Services Also Have:
**`ARCHITECTURE.md`** (Only for complex services)
- **Purpose**: Detailed architecture and implementation guide
- **Services that have this**:
  - `tenant-admin-service` (complex multi-tenant logic)
  - `saas-admin-service` (platform administration)
- **Content**:
  - Architectural patterns used
  - Code organization and structure
  - Business logic details
  - Database relationships
  - Integration points
- **When to update**: Major refactoring, architectural changes

---

## Service Documentation Status

### Services with README only (18 services):
- `analytics-consumer`
- `analytics-service`
- `api-gateway`
- `audit-consumer`
- `billing-consumer`
- `branding-service`
- `component-service`
- `database-service`
- `event-store-service`
- `incident-service`
- `landing-page-service`
- `monitoring-service`
- `notification-consumer`
- `notification-service`
- `payment-service`
- `shared-resilience`
- `status-ui-service`
- `user-service`

### Services with README + ARCHITECTURE (2 services):
- `tenant-admin-service` - Complex multi-tenant management
- `saas-admin-service` - Platform-wide administration

---

## What We Removed (36 files deleted)

### Root Level (7 files):
- ❌ `IMPLEMENTATION_SUMMARY.md` - Outdated audit report
- ❌ `ARCHITECTURE_AUDIT_REPORT.md` - Old audit (Oct 1)
- ❌ `ARCHITECTURE_FINDINGS_OCT_14_2025.md` - Session-specific findings
- ❌ `TESTING_COMPLETE_SUMMARY.md` - Session test results
- ❌ `TEST_RESULTS_OCTOBER_14_2025.md` - Session test results
- ❌ `SAAS_ADMIN_TESTING_STRATEGY.md` - Old test strategy
- ❌ `STATUS_PAGE_FEATURE_ANALYSIS.md` - Outdated feature analysis

### Microservices Directory (8 files):
- ❌ `BEAKON_MICROSERVICES_REFERENCE.md` - Duplicate (less complete)
- ❌ `COMPLETE_MICROSERVICES_REFERENCE.md` - Duplicate (outdated ports)
- ❌ `MICROSERVICES_PORT_REFERENCE.md` - Info in SERVICE_CATALOG.md
- ❌ `IMPLEMENTATION_SUMMARY.md` - Session-specific report
- ❌ `COMPLETE_SYSTEM_IMPROVEMENTS.md` - Session-specific report
- ❌ `FINAL_VALIDATION_REPORT.md` - Session-specific test results
- ❌ `SERVICE_STATUS_REPORT.md` - Outdated status report
- ❌ `calude-reads.md` - Temporary file (typo in name)

### Tenant Admin Service (17 files):
- ❌ All session-specific test results and reports
- ❌ All temporary implementation guides
- ❌ All feature-specific docs (info consolidated into ARCHITECTURE.md)

### Other Services (4 files):
- ❌ notification-consumer: STATUSPAGE_CHECKLIST.md, DOCKER.md, issue_tracker.md
- ❌ landing-page-service: MODERNIZATION_SUMMARY.md

---

## Documentation Maintenance Guidelines

### When to Create New Documentation:
✅ **DO create documentation for**:
- New microservices (README.md required)
- Complex architectural patterns (ARCHITECTURE.md if needed)
- New deployment procedures (update DEPLOYMENT_GUIDE.md)
- New inter-service communication patterns (update API_GATEWAY_COMMUNICATION_GUIDE.md)

### When NOT to Create Documentation:
❌ **DO NOT create**:
- Session-specific test results (put in PR description instead)
- Temporary implementation notes (put in code comments)
- Duplicate service catalogs (update SERVICE_CATALOG.md)
- Feature-specific guides (put in service ARCHITECTURE.md)
- Completion reports (use Git commit messages)

### Update Process:
1. **Service Changes**: Update service's README.md
2. **Port Changes**: Update SERVICE_CATALOG.md
3. **API Changes**: Update both service README.md and SERVICE_CATALOG.md
4. **Architecture Changes**: Update service ARCHITECTURE.md (if exists) or root ARCHITECTURE.md
5. **New Services**: Create README.md in service dir, add to SERVICE_CATALOG.md

---

## Quick Reference for AI Assistants

### To Understand a Service:
1. **Start here**: `/Beakon/SERVICE_CATALOG.md` (most comprehensive)
2. **Then read**: `/microservices/<service-name>/README.md`
3. **For complex services**: `/microservices/<service-name>/ARCHITECTURE.md`

### To Understand the Platform:
1. **Start here**: `/Beakon/README.md`
2. **Architecture**: `/Beakon/ARCHITECTURE.md`
3. **All services**: `/Beakon/SERVICE_CATALOG.md`
4. **Databases**: `/Beakon/DATABASE_ARCHITECTURE.md`

### To Understand Service Communication:
1. **Inter-service**: `/microservices/API_GATEWAY_COMMUNICATION_GUIDE.md`
2. **Specific service**: `/Beakon/SERVICE_CATALOG.md` (dependencies section)

---

## Documentation Health Metrics

| Metric | Status | Count |
|--------|--------|-------|
| **Root documents** | ✅ Optimal | 6 essential docs |
| **Microservices docs** | ✅ Optimal | 2 coordination docs |
| **Services with README** | ✅ Complete | 20/20 services |
| **Services with ARCHITECTURE** | ✅ Appropriate | 2 complex services |
| **Duplicate catalogs** | ✅ Eliminated | 0 duplicates |
| **Session reports** | ✅ Cleaned | 0 temp files |
| **Total documentation files** | ✅ Streamlined | 28 files (was 64) |

---

## Benefits of This Structure

1. **Single Source of Truth**: No confusion about which doc is current
2. **Easy Navigation**: Clear hierarchy (platform → coordination → service)
3. **AI-Friendly**: Clear structure helps AI assistants understand codebase
4. **Maintainable**: Fewer docs to keep updated
5. **Developer-Friendly**: Easy to find what you need
6. **No Staleness**: Removed all temporary and session-specific docs

---

**Last Cleanup**: October 14, 2025
**Files Removed**: 36 files
**Files Kept**: 28 files
**Status**: ✅ Optimized and current
