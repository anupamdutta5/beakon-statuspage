# Beakon Documentation Index

**Last Updated:** 2025-10-21
**Total Active Documents:** 25 files
**Archived Documents:** 50+ files (preserved for historical reference)

---

## 🚀 Quick Start (Read These First)

**For New Developers/AI:**
1. **[README.md](../README.md)** - Project overview and quickstart
2. **[CLAUDE.md](../CLAUDE.md)** - AI/Developer onboarding guide
3. **[AI_CONTEXT.md](../AI_CONTEXT.md)** - Quick reference for AI sessions
4. **[microservices/QUICK_START.md](../microservices/QUICK_START.md)** - 15-minute setup guide

**Time Investment:** 30-45 minutes to read these 4 docs

---

## 📚 Documentation Structure

### Level 1: Root Documentation (Essential)

**Core Architecture & Reference:**
- **[README.md](../README.md)** - Main project overview
- **[SERVICE_CATALOG.md](../SERVICE_CATALOG.md)** - Complete service reference (21 services)
- **[ARCHITECTURE.md](../ARCHITECTURE.md)** - System architecture overview
- **[DATABASE_ARCHITECTURE.md](../DATABASE_ARCHITECTURE.md)** - Database schemas (14 databases)

**Development Guides:**
- **[CLAUDE.md](../CLAUDE.md)** - Developer/AI onboarding (750 lines)
- **[AI_CONTEXT.md](../AI_CONTEXT.md)** - Quick context for AI (256 lines)
- **[AUTHENTICATION_GUIDE.md](../AUTHENTICATION_GUIDE.md)** - Auth & session management
- **[DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md)** - Production deployment (K8s/Docker)
- **[OPERATIONAL_RUNBOOK.md](../OPERATIONAL_RUNBOOK.md)** - Operations & troubleshooting

**Infrastructure:**
- **[REDIS_SETUP.md](../REDIS_SETUP.md)** - Redis setup (development)
- **[REDIS_PRODUCTION_DEPLOYMENT.md](../REDIS_PRODUCTION_DEPLOYMENT.md)** - Redis production deployment
- **[DOCUMENTATION_STRUCTURE.md](../DOCUMENTATION_STRUCTURE.md)** - How docs are organized

### Level 2: Microservices Documentation

**Quick Guides:**
- **[microservices/QUICK_START.md](../microservices/QUICK_START.md)** - 15-minute setup (NEW)
- **[microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md)** - Complete frontend guide (NEW)
- **[microservices/DATABASE_SETUP.md](../microservices/DATABASE_SETUP.md)** - Database initialization
- **[microservices/README.md](../microservices/README.md)** - Microservices overview

**Specialized Topics:**
- **[microservices/docs/architecture/](../microservices/docs/architecture/)**
  - `API_GATEWAY_COMMUNICATION_GUIDE.md` - Service communication patterns

- **[microservices/docs/testing/](../microservices/docs/testing/)**
  - `TESTING_GUIDE.md` - Complete testing guide (500+ lines)
  - `SECURITY_ROADMAP.md` - Security hardening plan (450+ lines)

### Level 3: Service-Specific Documentation

**Each service has:**
- `README.md` - Service overview, API endpoints, quickstart
- `ARCHITECTURE.md` - Service-specific architecture (if unique from root)

**Major Services:**
- **[saas-admin-service/README.md](../microservices/saas-admin-service/README.md)** - Platform admin
- **[saas-admin-service/ARCHITECTURE.md](../microservices/saas-admin-service/ARCHITECTURE.md)**
- **[saas-admin-service/API_INTEGRATION.md](../microservices/saas-admin-service/API_INTEGRATION.md)**

- **[tenant-admin-service/README.md](../microservices/tenant-admin-service/README.md)** - Tenant management
- **[tenant-admin-service/ARCHITECTURE.md](../microservices/tenant-admin-service/ARCHITECTURE.md)**

- **[saas-admin-frontend/README.md](../microservices/saas-admin-frontend/README.md)** - SaaS Admin UI
- **[tenant-admin-frontend/README.md](../microservices/tenant-admin-frontend/README.md)** - Tenant Admin UI

**Other Services:**
- [user-service/README.md](../microservices/user-service/README.md)
- [component-service/README.md](../microservices/component-service/README.md)
- [incident-service/README.md](../microservices/incident-service/README.md)
- [payment-service/README.md](../microservices/payment-service/README.md)
- [landing-page-service/README.md](../microservices/landing-page-service/README.md)
- ... (16 more services - see SERVICE_CATALOG.md)

---

## 🎯 Documentation by Use Case

### "I'm New - How Do I Start?"

**Read in order:**
1. [README.md](../README.md) - Project overview (5 min)
2. [CLAUDE.md](../CLAUDE.md) - Developer guide (15 min)
3. [microservices/QUICK_START.md](../microservices/QUICK_START.md) - Setup (30 min)
4. [SERVICE_CATALOG.md](../SERVICE_CATALOG.md) - Service reference (10 min)

**Total Time:** ~1 hour to get productive

### "I Need to Set Up Development Environment"

1. [microservices/QUICK_START.md](../microservices/QUICK_START.md) - Complete setup guide
2. [microservices/DATABASE_SETUP.md](../microservices/DATABASE_SETUP.md) - Database initialization
3. [REDIS_SETUP.md](../REDIS_SETUP.md) - Redis setup (optional)
4. [microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md) - Frontend setup

### "I'm Working on Authentication/Sessions"

1. [AUTHENTICATION_GUIDE.md](../AUTHENTICATION_GUIDE.md) - Complete auth guide
2. [microservices/docs/testing/SECURITY_ROADMAP.md](../microservices/docs/testing/SECURITY_ROADMAP.md) - Security hardening
3. [saas-admin-service/API_INTEGRATION.md](../microservices/saas-admin-service/API_INTEGRATION.md) - API auth patterns

### "I'm Working on Frontend"

1. [microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md) - Complete frontend guide
2. [saas-admin-frontend/README.md](../microservices/saas-admin-frontend/README.md) - SaaS Admin frontend
3. [tenant-admin-frontend/README.md](../microservices/tenant-admin-frontend/README.md) - Tenant Admin frontend
4. [microservices/docs/testing/TESTING_GUIDE.md](../microservices/docs/testing/TESTING_GUIDE.md) - Testing guide

### "I'm Working on Database"

1. [DATABASE_ARCHITECTURE.md](../DATABASE_ARCHITECTURE.md) - All database schemas
2. [microservices/DATABASE_SETUP.md](../microservices/DATABASE_SETUP.md) - Setup scripts
3. Individual service `README.md` for migrations

### "I'm Deploying to Production"

1. [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) - Kubernetes deployment
2. [OPERATIONAL_RUNBOOK.md](../OPERATIONAL_RUNBOOK.md) - Operations guide
3. [REDIS_PRODUCTION_DEPLOYMENT.md](../REDIS_PRODUCTION_DEPLOYMENT.md) - Redis production
4. [microservices/docs/testing/SECURITY_ROADMAP.md](../microservices/docs/testing/SECURITY_ROADMAP.md) - Security checklist

### "I'm Debugging/Troubleshooting"

1. [OPERATIONAL_RUNBOOK.md](../OPERATIONAL_RUNBOOK.md) - Operational procedures
2. [microservices/QUICK_START.md](../microservices/QUICK_START.md#troubleshooting) - Common issues
3. [microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md#troubleshooting) - Frontend issues
4. [CLAUDE.md](../CLAUDE.md#troubleshooting) - Development issues

### "I'm Testing the System"

1. [microservices/docs/testing/TESTING_GUIDE.md](../microservices/docs/testing/TESTING_GUIDE.md) - Complete testing guide
2. Individual service `README.md` for service-specific tests

---

## 📁 Archive Structure

**Historical documentation preserved in:**

```
docs/archive/
├── sessions/
│   ├── 2025-10-21/          # Recent session logs (8 files)
│   │   ├── URGENT_FIXES_NEEDED.md
│   │   ├── COMPLETE_TENANT_UPDATE_FIELDS.md
│   │   ├── FINAL_IMPLEMENTATION_STATUS.md
│   │   └── ... (5 more)
│   └── 2025-09-25/          # September sessions
│
microservices/docs/archive/   # Microservices session logs (7 files)
├── CLEANUP_PLAN.md
├── CLEANUP_SUMMARY.md
├── E2E_TEST_REPORT.md
├── FRONTEND_BACKEND_SPLIT_SUMMARY.md
├── IMPLEMENTATION_STATUS.md
├── PHASES_0-5_COMPLETION_SUMMARY.md
└── ... (more)

microservices/saas-admin-service/docs/archive/  # 16 session docs
microservices/tenant-admin-service/docs/archive/ # 13 session docs
microservices/landing-page-service/docs/archive/ # 4 session docs
```

**Total Archived:** 50+ documents (preserved for historical reference)

---

## 🗂️ Documentation by Category

### Architecture & Design
- [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture
- [DATABASE_ARCHITECTURE.md](../DATABASE_ARCHITECTURE.md) - Database design
- [SERVICE_CATALOG.md](../SERVICE_CATALOG.md) - Service catalog
- [saas-admin-service/ARCHITECTURE.md](../microservices/saas-admin-service/ARCHITECTURE.md)
- [tenant-admin-service/ARCHITECTURE.md](../microservices/tenant-admin-service/ARCHITECTURE.md)

### Development Guides
- [CLAUDE.md](../CLAUDE.md) - Developer onboarding
- [microservices/QUICK_START.md](../microservices/QUICK_START.md) - Quick setup
- [microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md) - Frontend development
- [AUTHENTICATION_GUIDE.md](../AUTHENTICATION_GUIDE.md) - Auth implementation
- [microservices/DATABASE_SETUP.md](../microservices/DATABASE_SETUP.md) - Database setup

### Operations & Deployment
- [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) - Production deployment
- [OPERATIONAL_RUNBOOK.md](../OPERATIONAL_RUNBOOK.md) - Operations guide
- [REDIS_SETUP.md](../REDIS_SETUP.md) - Redis development
- [REDIS_PRODUCTION_DEPLOYMENT.md](../REDIS_PRODUCTION_DEPLOYMENT.md) - Redis production

### Testing & Security
- [microservices/docs/testing/TESTING_GUIDE.md](../microservices/docs/testing/TESTING_GUIDE.md) - Testing guide
- [microservices/docs/testing/SECURITY_ROADMAP.md](../microservices/docs/testing/SECURITY_ROADMAP.md) - Security roadmap

### Infrastructure
- [microservices/postgres/README.md](../microservices/postgres/README.md) - PostgreSQL setup
- [microservices/redis/README.md](../microservices/redis/README.md) - Redis setup
- [microservices/rabbitmq/README.md](../microservices/rabbitmq/README.md) - RabbitMQ setup

### Service Communication
- [microservices/docs/architecture/API_GATEWAY_COMMUNICATION_GUIDE.md](../microservices/docs/architecture/API_GATEWAY_COMMUNICATION_GUIDE.md)

---

## 🔍 Finding Documentation

### By Topic

**Authentication:**
- Root: [AUTHENTICATION_GUIDE.md](../AUTHENTICATION_GUIDE.md)
- Frontend: [microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md#authentication--security)
- Service: [saas-admin-service/API_INTEGRATION.md](../microservices/saas-admin-service/API_INTEGRATION.md)

**Database:**
- Root: [DATABASE_ARCHITECTURE.md](../DATABASE_ARCHITECTURE.md)
- Setup: [microservices/DATABASE_SETUP.md](../microservices/DATABASE_SETUP.md)
- Service: Each service `README.md` has database section

**Frontend:**
- Guide: [microservices/FRONTEND_GUIDE.md](../microservices/FRONTEND_GUIDE.md)
- SaaS Admin: [saas-admin-frontend/README.md](../microservices/saas-admin-frontend/README.md)
- Tenant Admin: [tenant-admin-frontend/README.md](../microservices/tenant-admin-frontend/README.md)

**Testing:**
- Main: [microservices/docs/testing/TESTING_GUIDE.md](../microservices/docs/testing/TESTING_GUIDE.md)
- Security: [microservices/docs/testing/SECURITY_ROADMAP.md](../microservices/docs/testing/SECURITY_ROADMAP.md)

**Deployment:**
- Production: [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md)
- Operations: [OPERATIONAL_RUNBOOK.md](../OPERATIONAL_RUNBOOK.md)

### By File Location

```
Beakon/
├── README.md                    # Start here
├── CLAUDE.md                    # Developer guide
├── AI_CONTEXT.md                # AI quick reference
├── SERVICE_CATALOG.md           # Service reference
├── ARCHITECTURE.md              # System architecture
├── DATABASE_ARCHITECTURE.md     # Database schemas
├── AUTHENTICATION_GUIDE.md      # Auth guide
├── DEPLOYMENT_GUIDE.md          # Deployment
├── OPERATIONAL_RUNBOOK.md       # Operations
├── REDIS_SETUP.md               # Redis dev
├── REDIS_PRODUCTION_DEPLOYMENT.md
├── DOCUMENTATION_STRUCTURE.md
│
├── docs/
│   ├── INDEX.md (this file)
│   ├── archive/ (historical docs)
│   ├── current/ (legacy - being phased out)
│   └── historical/ (legacy - being phased out)
│
└── microservices/
    ├── README.md
    ├── QUICK_START.md           # Setup guide (NEW)
    ├── FRONTEND_GUIDE.md        # Frontend guide (NEW)
    ├── DATABASE_SETUP.md
    ├── docs/
    │   ├── architecture/
    │   │   └── API_GATEWAY_COMMUNICATION_GUIDE.md
    │   ├── testing/
    │   │   ├── TESTING_GUIDE.md
    │   │   └── SECURITY_ROADMAP.md
    │   └── archive/ (session logs)
    │
    └── {service-name}/
        ├── README.md
        ├── ARCHITECTURE.md (if unique)
        └── docs/archive/ (service session logs)
```

---

## 📊 Documentation Statistics

### Active Documentation
- **Root Level:** 12 files (~5,000 lines)
- **Microservices Root:** 4 files (~2,500 lines)
- **Microservices Docs:** 3 files (~1,500 lines)
- **Service-Specific:** ~40 README files (~10,000 lines)
- **Total Active:** ~25 essential files (~19,000 lines)

### Archived Documentation
- **Root Archive:** 8 files
- **Microservices Archive:** 7 files
- **Service Archives:** 35+ files
- **Total Archived:** 50+ files (preserved for history)

### Documentation Health
- ✅ **Organized:** Clear hierarchy (Root → Microservices → Service)
- ✅ **Up-to-Date:** Last consolidated October 21, 2025
- ✅ **Comprehensive:** All major topics covered
- ✅ **Accessible:** Easy to find with this index
- ✅ **Maintained:** Historical docs archived, not deleted

---

## 🛠️ Documentation Maintenance

### When to Update This Index

- New service added → Update SERVICE_CATALOG.md + service list here
- Major feature added → Update relevant guides + this index
- Documentation reorganization → Update structure + this index
- New consolidated doc created → Add to Quick Start or categories

### Documentation Standards

**Every README.md should have:**
1. Overview/Purpose
2. Prerequisites
3. Quick Start
4. API Reference (for services)
5. Configuration
6. Development Guide
7. Testing
8. Troubleshooting

**Documentation style:**
- Clear headings with anchors
- Code examples with syntax highlighting
- Step-by-step instructions
- Expected outputs shown
- Troubleshooting sections
- Cross-references to related docs

---

## 🔗 External References

**Repositories:**
- **Main:** GitHub (private)
- **SaaS Admin Frontend:** https://github.com/anupamdutta5/saas-admin-frontend
- **Tenant Admin Frontend:** https://github.com/anupamdutta5/tenant-admin-frontend
- **RabbitMQ Deployment:** https://github.com/anupamdutta5/rabbitmq
- **Shared Resilience Library:** https://github.com/anupamdutta5/shared-resilience

**Technologies:**
- **Go:** https://go.dev/doc/
- **Next.js:** https://nextjs.org/docs
- **PostgreSQL:** https://www.postgresql.org/docs/
- **React Query:** https://tanstack.com/query/latest/docs/react/overview
- **shadcn/ui:** https://ui.shadcn.com/

---

## 📝 Quick Reference Cards

### Port Allocation
```
3001  - SaaS Admin Frontend
3002  - Tenant Admin Frontend
8080  - API Gateway
8081  - User Service
8084  - Component Service
8085  - Notification Service
8086  - Incident Service
8088  - Payment Service
8090  - Analytics Service
8092  - Monitoring Service
8093  - Status UI Service
8096  - Event Store Service
8097  - Branding Service
8098  - SaaS Admin Backend
8099  - Tenant Admin Backend
8100  - Landing Page Service
5432  - PostgreSQL
6379  - Redis
5672  - RabbitMQ (AMQP)
15672 - RabbitMQ Management
```

### Database Names
```
saas_admin
tenant_admin_db
statuspage_user
statuspage_component
statuspage_incident
statuspage_payment
statuspage_landing
statuspage_notification
statuspage_branding
statuspage_event_store
statuspage_analytics
statuspage_monitoring
statuspage_audit
status_ui_db
```

### Key Commands
```bash
# Setup
cd microservices && ./init-all-databases.sh
cd microservices && ./start-all-services.sh

# Health checks
curl http://localhost:8098/api/v1/health
curl http://localhost:8099/health

# Database
psql -U postgres -d tenant_admin_db

# Logs
tail -f /tmp/saas-admin-backend.log
```

---

**Index Last Updated:** 2025-10-21
**Total Documents Indexed:** 70+ (25 active + 50+ archived)
**Documentation Coverage:** ✅ Complete

For questions about documentation structure, see [DOCUMENTATION_STRUCTURE.md](../DOCUMENTATION_STRUCTURE.md)
