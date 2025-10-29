# Beakon Documentation

This directory contains all project documentation organized by category.

## Directory Structure

```
docs/
├── README.md                    # This file
├── guides/                      # Configuration and how-to guides
│   ├── AUTHENTICATION_GUIDE.md
│   ├── DEPLOYMENT_GUIDE.md
│   ├── OPERATIONAL_RUNBOOK.md
│   └── SAML_CONFIGURATION_GUIDE.md
│
├── architecture/                # Architecture and design docs
│   ├── ARCHITECTURE.md
│   ├── DATABASE_ARCHITECTURE.md
│   ├── SERVICE_CATALOG.md
│   └── FEATURES.md
│
├── planning/                    # Planning and migration docs
│   ├── SAML_MIGRATION_PLAN.md
│   ├── IMPROVEMENTS_TRACKER.md
│   └── GROUP2_USER_SERVICE_AUDIT.md
│
└── summaries/                   # Session summaries and reports
    ├── SESSION_SUMMARY_2025-10-29_SAML_MIGRATION.md
    └── archived/                # Historical summaries
        └── (20 historical files)
```

## Quick Links

### For New Developers
1. Start with [README.md](../README.md) in root (project overview)
2. Read [AI_CONTEXT.md](../AI_CONTEXT.md) (quick reference)
3. Review [FEATURES.md](architecture/FEATURES.md) (all features explained)
4. Check [SERVICE_CATALOG.md](architecture/SERVICE_CATALOG.md) (service details)

### For Configuration
- [AUTHENTICATION_GUIDE.md](guides/AUTHENTICATION_GUIDE.md) - Auth & sessions
- [SAML_CONFIGURATION_GUIDE.md](guides/SAML_CONFIGURATION_GUIDE.md) - SSO setup
- [DEPLOYMENT_GUIDE.md](guides/DEPLOYMENT_GUIDE.md) - Production deployment
- [OPERATIONAL_RUNBOOK.md](guides/OPERATIONAL_RUNBOOK.md) - Operations & troubleshooting

### For Architecture Understanding
- [ARCHITECTURE.md](architecture/ARCHITECTURE.md) - System architecture
- [DATABASE_ARCHITECTURE.md](architecture/DATABASE_ARCHITECTURE.md) - Database schemas
- [SERVICE_CATALOG.md](architecture/SERVICE_CATALOG.md) - Complete service reference
- [FEATURES.md](architecture/FEATURES.md) - Feature documentation

### For Development
- [CLAUDE.md](../CLAUDE.md) - Developer workflow guide
- [IMPROVEMENTS_TRACKER.md](planning/IMPROVEMENTS_TRACKER.md) - Current work tracking

## Root Directory Files

Essential files kept in root for visibility:
- **README.md** - Project overview (START HERE)
- **CLAUDE.md** - Developer guide
- **AI_CONTEXT.md** - Quick reference
- **FEATURES.md** - Feature documentation (may move to docs/architecture/)

## Documentation Standards

- Use markdown for all documentation
- Include table of contents for docs > 100 lines
- Update "Last Updated" dates when making changes
- Link related documents for easy navigation
