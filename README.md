# Beakon Status Page Platform

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Status](https://img.shields.io/badge/status-production%20ready-green.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)
![PostgreSQL](https://img.shields.io/badge/postgresql-16-336791.svg)

**Enterprise-grade, multi-tenant status page platform** similar to Statuspage.io

[Features](#-features) • [Quick Start](#-quick-start) • [Architecture](#-architecture) • [Documentation](#-documentation)

</div>

---

## 📖 Table of Contents

- [Overview](#-overview)
- [Quick Start](#-quick-start)
- [Features](#-features)
- [Architecture](#-architecture)
- [Services](#-services)
- [Documentation](#-documentation)
- [Development](#-development)
- [Support](#-support)

---

## 🌟 Overview

Beakon is a comprehensive status page platform that enables businesses to communicate service status to their customers. Built with microservices architecture, it provides enterprise-level features including multi-tenancy, real-time monitoring, incident management, and customizable status pages.

### Platform Metrics

| Metric | Value |
|--------|-------|
| **Total Services** | 21 microservices (19 backend + 2 frontend) |
| **Backend** | Go 1.21+ with Gin framework |
| **Frontend** | Next.js 14 with TypeScript |
| **Databases** | 14 PostgreSQL (database-per-service) |
| **Build Status** | ✅ All 19 services building |
| **Production Status** | ✅ Ready |

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+, PostgreSQL 16, Node.js 18+
- Optional: Docker, Redis, RabbitMQ

### Installation

```bash
# Initialize databases
cd microservices && ./init-all-databases.sh

# Start backends
./start-all-backends.sh

# Start frontends (new terminal)
./start-all-frontends.sh
```

### Verify

```bash
curl http://localhost:8080/health  # API Gateway
open http://localhost:3001          # SaaS Admin UI
open http://anupam.localhost:3002   # Tenant Admin UI
```

See **[CLAUDE.md](CLAUDE.md)** for detailed setup guide.

---

## ✨ Features

### Core Capabilities

**Multi-Tenancy**
- Complete tenant isolation with UUID-based partitioning
- Subdomain-based routing
- Tenant-specific branding

**Status Pages**
- Component hierarchy & real-time status
- Public branded status pages
- Historical uptime (90 days)
- Embeddable status badges

**Monitoring** (8 types)
- HTTP/HTTPS, TCP, Ping, Keyword
- Heartbeat, Container, Kubernetes, External
- Multi-location monitoring
- Auto-incident creation
- Anomaly detection

**Incident Management**
- Full lifecycle (Investigating → Resolved)
- Incident templates
- Real-time updates
- Post-mortem reports

**Notifications** (9 channels)
- Email, SMS, Webhooks
- Slack, Teams, Discord, Telegram
- PagerDuty, Push Notifications

**Analytics**
- Uptime metrics & SLA tracking
- Performance analytics
- Custom reports (PDF/CSV)

**Security**
- JWT authentication
- RBAC with granular permissions
- Team-based access
- Three-tier session system

See **[FEATURES.md](FEATURES.md)** for complete feature documentation.

---

## 🏗 Architecture

### Microservices Overview

```
Clients → API Gateway (8080) → Backend Services
                               ├── User Service (8081)
                               ├── Component Service (8084)
                               ├── Notification Service (8085)
                               ├── Incident Service (8086)
                               ├── Monitoring Service (8092)
                               └── ... (10 more)

PostgreSQL (14 databases) ← All Services
RabbitMQ ← Event-Driven Communication
Redis ← Optional Caching
```

### Key Design Patterns
- **Database-per-service**: Complete data isolation
- **Event-driven**: RabbitMQ for async communication
- **Shared library**: `shared-resilience` for common functionality
- **API Gateway**: Centralized routing & auth

See **[ARCHITECTURE.md](ARCHITECTURE.md)** for detailed architecture.

---

## 🔧 Services

### Backend HTTP Services (15)

| Service | Port | Purpose |
|---------|------|---------|
| api-gateway | 8080 | Request routing, JWT validation |
| user-service | 8081 | Authentication |
| component-service | 8084 | Component management |
| notification-service | 8085 | Multi-channel notifications |
| incident-service | 8086 | Incident lifecycle |
| payment-service | 8088 | Billing & payments |
| analytics-service | 8090 | Metrics & SLA |
| monitoring-service | 8092 | Health checks & alerts |
| status-ui-service | 8093 | Public status pages |
| event-store-service | 8096 | Event sourcing |
| branding-service | 8097 | Customization |
| saas-admin-service | 8098 | Platform admin |
| tenant-admin-service | 8099 | Tenant management |
| landing-page-service | 8100 | Marketing site |

### Background Consumers (4)
- analytics-consumer, audit-consumer, billing-consumer, notification-consumer

### Frontend Services (2)
- saas-admin-frontend (3001), tenant-admin-frontend (3002)

See **[SERVICE_CATALOG.md](SERVICE_CATALOG.md)** for complete service reference.

---

## 📚 Documentation

### Essential Docs (Read in Order)

1. **[AI_CONTEXT.md](AI_CONTEXT.md)** - Quick reference (5 min)
2. **[CLAUDE.md](CLAUDE.md)** - Developer guide (15 min)
3. **[FEATURES.md](FEATURES.md)** - ⭐ All features explained
4. **[SERVICE_CATALOG.md](SERVICE_CATALOG.md)** - Service details

### Technical Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design
- **[DATABASE_ARCHITECTURE.md](DATABASE_ARCHITECTURE.md)** - 14 database schemas
- **[AUTHENTICATION_GUIDE.md](AUTHENTICATION_GUIDE.md)** - Auth & sessions
- **[DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)** - Production deployment
- **[OPERATIONAL_RUNBOOK.md](OPERATIONAL_RUNBOOK.md)** - Troubleshooting

### Development Guides

- **[microservices/FRONTEND_GUIDE.md](microservices/FRONTEND_GUIDE.md)** - Frontend dev
- **[microservices/docs/testing/TESTING_GUIDE.md](microservices/docs/testing/TESTING_GUIDE.md)** - Testing

---

## 💻 Development

### Project Structure

```
Beakon/
├── microservices/              # 21 services (each is a git repo)
│   ├── api-gateway/           # Port 8080
│   ├── user-service/          # Port 8081
│   ├── ...                    # 15 more backend services
│   ├── analytics-consumer/    # Background workers (4)
│   ├── saas-admin-frontend/   # Port 3001
│   ├── tenant-admin-frontend/ # Port 3002
│   ├── shared-resilience/     # Shared Go library
│   └── postgres/              # Infrastructure configs
├── docs/                      # Documentation
├── FEATURES.md                # ⭐ Complete feature docs
├── CLAUDE.md                  # Developer guide
└── README.md                  # This file
```

### Common Commands

```bash
# Run a service
cd microservices/monitoring-service
./start-dev.sh

# Run tests
go test ./...

# Build service
go build -o monitoring-service cmd/main.go

# Update dependencies
go mod tidy
go get github.com/anupamdutta5/shared-resilience@latest

# Check health
curl http://localhost:8092/health
```

### Database Operations

```bash
# Initialize all databases
cd microservices && ./init-all-databases.sh

# Initialize specific service
cd microservices/tenant-admin-service && ./init-db.sh

# Connect to database
psql -U postgres -d tenant_admin_db
```

---

## 🆘 Support

### Getting Help

1. Check **[CLAUDE.md](CLAUDE.md)** for common issues
2. Review **[OPERATIONAL_RUNBOOK.md](OPERATIONAL_RUNBOOK.md)**
3. Check service README: `microservices/<service>/README.md`

### Common Issues

**Port already in use:**
```bash
lsof -i :8099 && kill -9 <PID>
```

**Database connection failed:**
```bash
pg_isready -h localhost -p 5432
psql -U postgres -l
```

**Service won't start:**
```bash
tail -f logs/<service>.log
env | grep DB_
go mod verify
```

---

## 📊 Status & Roadmap

### Current Status
✅ All 19 backend services building
✅ Multi-tenancy implemented
✅ Core features complete
⚠️ Security hardening in progress
⚠️ Performance optimization in progress

### Roadmap 2026
- Q1: GraphQL API, WebSocket updates, Mobile apps
- Q2: Multi-region, Integrations marketplace
- Q3: Custom reporting, API analytics, SOC 2

---

<div align="center">

**Built with Go, Next.js, and PostgreSQL**

[Documentation](docs/) • [Features](FEATURES.md) • [Architecture](ARCHITECTURE.md) • [Guide](CLAUDE.md)

</div>
