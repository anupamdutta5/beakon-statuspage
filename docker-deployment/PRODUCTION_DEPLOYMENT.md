# Production Deployment Guide

## ⚠️ CRITICAL: Docker Image Strategy

The `docker-deployment/` structure is designed for **production deployment** and must be **completely self-contained**.

### Current Issue

The generated docker-compose.yml files currently reference:
```yaml
build:
  context: ../../microservices/user-service  # ❌ WRONG - references git repo
  dockerfile: Dockerfile
```

**This will NOT work in production** because:
- You copy only `docker-deployment/` to production servers
- `../../microservices/` doesn't exist on production servers
- Build would fail

### Production Solution: Use Pre-built Images

## Recommended Approach: CI/CD Pipeline

### 1. Build Phase (CI/CD - GitHub Actions, GitLab CI, Jenkins, etc.)

```bash
# Build all images
docker-compose build

# Tag images for registry
docker tag beakon-user-service:latest myregistry.com/beakon-user-service:v1.0.0
docker tag beakon-user-service:latest myregistry.com/beakon-user-service:latest

# Push to registry
docker push myregistry.com/beakon-user-service:v1.0.0
docker push myregistry.com/beakon-user-service:latest
```

### 2. Update docker-compose.yml Files

**Remove build section**, use only image:

```yaml
services:
  user-service:
    image: myregistry.com/beakon-user-service:latest  # ✅ CORRECT
    # NO build section
    container_name: beakon-user-service
    ports:
      - "8081:8081"
    volumes:
      - ./configs:/app/configs:ro
    # ... rest of config
```

### 3. Deploy to Production

```bash
# On production server
scp -r docker-deployment/ user@prod-server:/opt/beakon/

# On production server
cd /opt/beakon/docker-deployment/user-service
docker-compose pull  # Pull latest images
docker-compose up -d
```

## Alternative Approach: Self-Contained Deployment Package

If you can't use a Docker registry, create a self-contained package:

### 1. Export Images to TAR Files

```bash
# Export all images
docker save beakon-user-service:latest -o user-service.tar
docker save beakon-tenant-admin-service:latest -o tenant-admin-service.tar
# ... repeat for all services

# Create deployment package
mkdir deployment-package
cp -r docker-deployment/ deployment-package/
mkdir deployment-package/images
mv *.tar deployment-package/images/

# Create deployment archive
tar -czf beakon-deployment-v1.0.0.tar.gz deployment-package/
```

### 2. Deploy to Production

```bash
# Copy to production server
scp beakon-deployment-v1.0.0.tar.gz user@prod-server:/opt/

# On production server
cd /opt
tar -xzf beakon-deployment-v1.0.0.tar.gz
cd deployment-package

# Load images
for img in images/*.tar; do
    docker load -i "$img"
done

# Deploy services
cd docker-deployment/postgres && docker-compose up -d && cd ../..
cd docker-deployment/redis && docker-compose up -d && cd ../..
cd docker-deployment/rabbitmq && docker-compose up -d && cd ../..
cd docker-deployment/user-service && docker-compose up -d && cd ../..
```

## Current Development Workflow

For **development only**, the current setup works because:
- Git repository is available
- Build context `../../microservices/` exists
- You're building locally

## Deployment Comparison

### ❌ Current (Development Only)
```
Production Server
└── docker-deployment/
    └── user-service/
        ├── docker-compose.yml (references ../../microservices/ ❌)
        ├── .env
        └── configs/
```
**Result**: Build fails - microservices/ doesn't exist!

### ✅ Production (Pre-built Images)
```
Production Server
└── docker-deployment/
    └── user-service/
        ├── docker-compose.yml (uses registry image ✅)
        ├── .env
        └── configs/

Docker Registry
└── myregistry.com/beakon-user-service:latest ✅
```
**Result**: Works perfectly!

### ✅ Production (Self-Contained TAR)
```
Production Server
└── deployment-package/
    ├── docker-deployment/
    │   └── user-service/
    │       ├── docker-compose.yml
    │       ├── .env
    │       └── configs/
    └── images/
        └── user-service.tar ✅
```
**Result**: Load TAR, then deploy!

## How to Fix for Production

### Option 1: Use Docker Registry (RECOMMENDED)

1. **Set up Docker Registry**:
   - DockerHub (public or private)
   - AWS ECR
   - Google GCR
   - Azure ACR
   - Self-hosted registry

2. **Update CI/CD**:
   ```yaml
   # .github/workflows/build-and-push.yml
   name: Build and Push Images
   on:
     push:
       branches: [main]
   jobs:
     build:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v2
         - name: Build images
           run: docker-compose build
         - name: Push to registry
           run: |
             docker login -u ${{ secrets.DOCKER_USERNAME }} -p ${{ secrets.DOCKER_PASSWORD }}
             docker tag beakon-user-service:latest myregistry/beakon-user-service:${{ github.sha }}
             docker push myregistry/beakon-user-service:${{ github.sha }}
   ```

3. **Update docker-compose.yml files** (remove build sections):
   ```yaml
   services:
     user-service:
       image: myregistry.com/beakon-user-service:latest
       # Remove build section entirely
   ```

### Option 2: Create Deployment Script

Create `scripts/create-deployment-package.sh`:
```bash
#!/bin/bash

VERSION=${1:-latest}
OUTPUT_DIR="beakon-deployment-$VERSION"

# Build all images
docker-compose build

# Export images
mkdir -p "$OUTPUT_DIR/images"
docker save beakon-user-service:latest -o "$OUTPUT_DIR/images/user-service.tar"
# ... export all services

# Copy deployment configs
cp -r docker-deployment/ "$OUTPUT_DIR/"

# Create archive
tar -czf "$OUTPUT_DIR.tar.gz" "$OUTPUT_DIR"

echo "Deployment package created: $OUTPUT_DIR.tar.gz"
```

## Infrastructure Services (postgres, redis, rabbitmq)

These are fine because they use public Docker images:
```yaml
services:
  postgres:
    image: postgres:14-alpine  # ✅ Public image - works everywhere
```

## Summary

| Deployment Type | Works in Dev | Works in Prod | Recommendation |
|-----------------|--------------|---------------|----------------|
| Current (references ../../microservices/) | ✅ | ❌ | Development only |
| Registry images | ✅ | ✅ | **RECOMMENDED** |
| TAR export/import | ✅ | ✅ | Offline deployments |
| Self-contained build | ✅ | ✅ | Complex setup |

## Action Required

For production deployment, you MUST:

1. ✅ **Set up Docker registry** OR create deployment TAR packages
2. ✅ **Update docker-compose.yml** files to remove build sections
3. ✅ **Use only `image:` parameter** pointing to registry or loaded images
4. ✅ **Test deployment** on a clean server without git repository

## Questions?

- **Q**: Can I use the current docker-deployment/ in production?
  **A**: NO - build context references git repository that won't exist

- **Q**: What's the easiest production deployment method?
  **A**: Use a Docker registry + CI/CD pipeline

- **Q**: Can I deploy without a registry?
  **A**: Yes - export images to TAR files and load on prod server

- **Q**: Do infrastructure services (postgres, redis) work?
  **A**: YES - they use public images from Docker Hub

---

**Status**: Documentation complete - Production deployment strategy clarified
**Created**: October 26, 2025
**Critical**: Current docker-deployment structure requires build images before production use
