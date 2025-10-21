# Next Steps - Testing & Production Hardening

**Date:** October 21, 2025
**Status:** Documentation Consolidated ✅ - Ready for Testing & Hardening

---

## 🎯 Current Status

### ✅ Completed
- [x] Documentation consolidation (120+ → 25 files)
- [x] Frontend-backend separation (Next.js + Go)
- [x] Middleware-based authentication
- [x] Comprehensive guides created (QUICK_START, FRONTEND_GUIDE, INDEX)
- [x] All services buildable and runnable

### 🔄 Ready for Next Phase
- [ ] Feature testing (follow TESTING_GUIDE.md)
- [ ] Production hardening (httpOnly cookies - Phase 1)
- [ ] New feature development on solid foundation

---

## 📋 Immediate Tasks (Priority Order)

### 1. Test Core Features (High Priority)

**Time:** 30-45 minutes
**Guide:** [microservices/docs/testing/TESTING_GUIDE.md](microservices/docs/testing/TESTING_GUIDE.md)

#### Quick Health Check (5 minutes)

```bash
# Check all services are running
lsof -i :3001 -i :3002 -i :8098 -i :8099 | grep LISTEN

# Check backend health
curl -s http://localhost:8098/api/v1/health | python3 -c "import sys, json; print(json.load(sys.stdin)['status'])"
curl -s http://localhost:8099/health | python3 -c "import sys, json; print(json.load(sys.stdin)['status'])"

# Check frontend redirects (middleware working)
curl -s -I http://localhost:3001 | grep HTTP
curl -s -I http://localhost:3002 | grep HTTP
```

**Expected Results:**
- ✅ All 4 services running
- ✅ Backends return "healthy"
- ✅ Frontends return "307 Temporary Redirect" (middleware working)

#### Authentication Flow Test (10 minutes)

**Test 1: Root URL Persistence**
```bash
# Manual test in browser:
1. Login at http://localhost:3001/login
   - Username: admin
   - Password: admin
2. Navigate to http://localhost:3001/admin/pricing
3. Change URL to http://localhost:3001/ (root)
4. Press Enter
```

**Expected:** ✅ Redirects to /admin/dashboard (stays logged in, NOT logged out)

**Test 2: Tenant Launch Button**
```bash
# Manual test in browser:
1. Login to SaaS Admin (http://localhost:3001/login)
2. Navigate to /admin/tenants
3. Click "Launch Tenant Admin" button on any tenant
```

**Expected:** ✅ Opens http://{subdomain}.localhost:3002 (frontend, NOT backend API)

#### Pricing Page Test (10 minutes)

```bash
# Manual test in browser:
1. Login to SaaS Admin
2. Navigate to /admin/pricing
3. Verify plans display correctly (features as bullet points)
4. Try creating a new plan:
   - Name: "Test Plan"
   - Slug: "test-plan"
   - Price: 19.99
   - Billing Interval: Monthly
   - Add 2-3 features
5. Click Save
```

**Expected:**
- ✅ Plans load without errors
- ✅ Features display as bullet points (not JSON string)
- ✅ Create operation succeeds
- ✅ New plan appears in list

#### API Integration Test (10 minutes)

```bash
# Get plans
curl -s http://localhost:8098/api/v1/plans | python3 -c "import sys, json; data=json.load(sys.stdin); print(f'Plans: {len(data.get(\"plans\", []))}')"

# Get tenants
curl -s http://localhost:8098/api/v1/tenants | python3 -c "import sys, json; data=json.load(sys.stdin); print(f'Tenants: {data.get(\"count\", 0)}')"

# Create tenant
curl -s -X POST http://localhost:8098/api/v1/tenants \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "E2E Test Company",
    "slug": "e2e-test",
    "contact_email": "e2e@example.com",
    "billing_email": "billing@e2e.com",
    "admin_email": "admin@e2e.com",
    "admin_password": "TestPassword123",
    "plan_id": "uuid-from-plans-list",
    "max_users": 50
  }' | python3 -c "import sys, json; print(json.dumps(json.load(sys.stdin), indent=2))"
```

**Expected:**
- ✅ API calls return valid JSON
- ✅ Data structure matches types
- ✅ Create operations succeed

**Full Testing Guide:** See [microservices/docs/testing/TESTING_GUIDE.md](microservices/docs/testing/TESTING_GUIDE.md) for complete test cases.

---

### 2. Fix RabbitMQ (Medium Priority)

**Issue:** RabbitMQ initialization failing (node not running)

**Quick Fix:**

```bash
cd microservices/rabbitmq

# Check status
./manage.sh health

# If not running, restart
./manage.sh down
./manage.sh up

# Verify
curl -u admin:SecureP@ssw0rd2024! http://localhost:15672/api/overview
```

**Expected:**
- ✅ RabbitMQ Management UI accessible at http://localhost:15672
- ✅ All queues created (9 queues)
- ✅ All exchanges created (5 exchanges)

**Alternative:** If RabbitMQ issues persist, services work fine without it (events just won't publish). Fix later if needed.

---

### 3. Production Hardening - Phase 1: httpOnly Cookies (High Priority)

**Time:** 2-3 hours
**Guide:** [microservices/docs/testing/SECURITY_ROADMAP.md](microservices/docs/testing/SECURITY_ROADMAP.md)

#### Why httpOnly Cookies?

**Current State:**
- Cookies: `document.cookie` (JavaScript accessible) ❌
- Security Risk: XSS attacks can steal tokens
- Browser Storage: Token in localStorage + cookie

**Target State:**
- Cookies: httpOnly (JavaScript inaccessible) ✅
- Security: XSS attacks cannot steal tokens
- Backend-Controlled: Server sets/deletes cookies

#### Implementation Steps

**Backend Changes (Go):**

1. **Update Login Handler** (saas-admin-service, tenant-admin-service)

```go
// File: internal/handlers/auth_handler.go

func (h *AuthHandler) Login(c *gin.Context) {
    // ... existing login logic ...

    // Generate JWT token
    token, err := generateJWT(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    // Set httpOnly cookie (NEW)
    c.SetCookie(
        "admin_token",           // name
        token,                   // value
        7*24*60*60,             // maxAge (7 days)
        "/",                     // path
        "",                      // domain (empty for current domain)
        false,                   // secure (false for dev, true for prod)
        true,                    // httpOnly (IMPORTANT!)
    )

    // Also return token in response for backward compatibility
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "token": token,
        "user": user,
    })
}
```

2. **Update Logout Handler**

```go
func (h *AuthHandler) Logout(c *gin.Context) {
    // Delete httpOnly cookie
    c.SetCookie(
        "admin_token",
        "",
        -1,                      // maxAge -1 deletes cookie
        "/",
        "",
        false,
        true,
    )

    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

**Frontend Changes (Next.js):**

1. **Update Login Function** (lib/api/auth.ts)

```typescript
login: async (credentials: LoginRequest): Promise<LoginResponse> => {
  const response = await post<LoginResponse, LoginRequest>('/api/v1/auth/login', credentials);

  if (response.token) {
    // Keep localStorage for backward compatibility with existing code
    localStorage.setItem('admin_token', response.token);

    // NOTE: Cookie is now set by backend (httpOnly)
    // Remove this code:
    // document.cookie = `admin_token=${response.token}; ...`
  }

  if (response.user) {
    localStorage.setItem('admin_user', JSON.stringify(response.user));
  }

  return response;
},
```

2. **Update Logout Function**

```typescript
logout: async (): Promise<void> => {
  try {
    // Call backend to delete httpOnly cookie
    await post('/api/v1/auth/logout');
  } finally {
    // Clean up localStorage
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_user');

    // NOTE: Cookie deleted by backend (httpOnly)
    // This line no longer needed:
    // document.cookie = 'admin_token=; path=/; max-age=0';
  }
},
```

3. **Middleware Already Works!**

The Next.js middleware reads cookies using `request.cookies.get()`, which works with both JavaScript-accessible and httpOnly cookies. No changes needed!

#### Testing httpOnly Cookies

```bash
# 1. Login via browser
# Go to http://localhost:3001/login
# Enter credentials

# 2. Check cookies in DevTools
# Open DevTools → Application → Cookies
# Look for admin_token cookie

# Expected:
# ✅ Name: admin_token
# ✅ HttpOnly: true (checkbox checked)
# ✅ Secure: false (dev mode)
# ✅ SameSite: Lax

# 3. Try to access cookie via JavaScript console
document.cookie
# Expected: admin_token should NOT appear in the output

# 4. Navigate around
# Change URL to http://localhost:3001/
# Expected: ✅ Still logged in (middleware reads httpOnly cookie)
```

**Files to Modify:**

Backend (2 services × 1 file each):
- `microservices/saas-admin-service/internal/handlers/saas_admin_handler.go`
- `microservices/tenant-admin-service/internal/handlers/user_handler.go`

Frontend (2 services × 1 file each):
- `microservices/saas-admin-frontend/lib/api/auth.ts`
- `microservices/tenant-admin-frontend/lib/api/auth.ts`

**Total:** 4 files to modify

**Estimated Time:**
- Backend changes: 30 minutes
- Frontend changes: 15 minutes
- Testing: 30 minutes
- Documentation: 15 minutes
- **Total: 1.5 hours**

---

### 4. Production Hardening - Phase 2: CSRF Protection (Medium Priority)

**Time:** 3-4 hours
**Prerequisites:** Phase 1 (httpOnly cookies) completed

**Guide:** See [microservices/docs/testing/SECURITY_ROADMAP.md](microservices/docs/testing/SECURITY_ROADMAP.md#phase-2-csrf-protection)

**Key Changes:**
- Backend generates CSRF tokens
- Frontend includes CSRF token in headers
- Backend validates CSRF on state-changing requests

**Defer until:** Phase 1 complete and tested

---

### 5. Continue Development - New Features

Once testing and Phase 1 hardening complete, you can build new features with confidence:

**Ideas for New Features:**

1. **Enhanced Analytics Dashboard**
   - Real-time metrics
   - Custom date ranges
   - Export capabilities

2. **Advanced RBAC**
   - Custom permissions
   - Role templates
   - Bulk user management

3. **Notification System**
   - Email templates
   - SMS integration
   - Webhook support

4. **Multi-Factor Authentication**
   - TOTP (Google Authenticator)
   - SMS verification
   - Backup codes

5. **API Rate Limiting UI**
   - Per-tenant limits
   - Real-time monitoring
   - Custom quotas

**Development Workflow:**

```bash
# 1. Create feature branch
git checkout -b feature/new-feature-name

# 2. Develop feature
# Reference CLAUDE.md for development patterns

# 3. Test feature
# Follow TESTING_GUIDE.md

# 4. Commit changes
git add .
git commit -m "feat: add new feature

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 5. Push to GitHub
git push origin feature/new-feature-name

# 6. Create PR (optional)
gh pr create --title "New Feature" --body "Description..."
```

---

## 📊 Progress Tracking

### Documentation ✅ (100% Complete)

- [x] Consolidation complete (25 files)
- [x] QUICK_START.md created
- [x] FRONTEND_GUIDE.md created
- [x] docs/INDEX.md created
- [x] Archives organized
- [x] Cross-references updated

### Testing 🔄 (0% Complete - Next Step)

- [ ] Health checks
- [ ] Authentication flow
- [ ] Pricing page functionality
- [ ] Tenant launch button
- [ ] API integration
- [ ] End-to-end scenarios

### Security Hardening 🔄 (40% Complete)

**Current (Development Mode):**
- [x] Middleware-based authentication
- [x] Cookie + LocalStorage dual storage
- [x] SameSite=Lax CSRF protection
- [x] 7-day token expiration
- [x] Environment-aware config

**Phase 1 (High Priority):**
- [ ] httpOnly cookies (backend-controlled)
- [ ] Remove client-side cookie manipulation

**Phase 2 (Medium Priority):**
- [ ] CSRF token protection
- [ ] CSRF validation middleware

**Phase 3 (Medium Priority):**
- [ ] Refresh token mechanism
- [ ] Token revocation on logout

**Phase 4 (High Priority):**
- [ ] Verify rate limiting active
- [ ] Per-endpoint limits

**Phase 5 (Production Only):**
- [ ] HTTPS enforcement
- [ ] Secure cookie flag
- [ ] HSTS headers

### Feature Development 🔄 (Ready)

Foundation ready for:
- [ ] New features
- [ ] Enhancements
- [ ] Integrations

---

## 🎯 Recommended Workflow

### This Week

**Day 1-2: Testing**
1. Follow [microservices/docs/testing/TESTING_GUIDE.md](microservices/docs/testing/TESTING_GUIDE.md)
2. Test all critical flows
3. Document any bugs found
4. Fix critical bugs

**Day 3-4: Phase 1 Security**
1. Implement httpOnly cookies (backend)
2. Update frontend to remove cookie manipulation
3. Test thoroughly
4. Commit changes

**Day 5: RabbitMQ + Planning**
1. Fix RabbitMQ if needed
2. Plan next sprint features
3. Update backlog

### Next Week

**Week 2: Phase 2 Security + New Features**
1. Implement CSRF protection
2. Start new feature development
3. Continue testing

---

## 📚 Documentation References

**Testing:**
- [microservices/docs/testing/TESTING_GUIDE.md](microservices/docs/testing/TESTING_GUIDE.md) - Complete testing guide

**Security:**
- [microservices/docs/testing/SECURITY_ROADMAP.md](microservices/docs/testing/SECURITY_ROADMAP.md) - 5-phase hardening plan

**Development:**
- [CLAUDE.md](CLAUDE.md) - Developer guide
- [microservices/QUICK_START.md](microservices/QUICK_START.md) - Setup guide
- [microservices/FRONTEND_GUIDE.md](microservices/FRONTEND_GUIDE.md) - Frontend development

**Reference:**
- [SERVICE_CATALOG.md](SERVICE_CATALOG.md) - All services
- [DATABASE_ARCHITECTURE.md](DATABASE_ARCHITECTURE.md) - Database schemas
- [docs/INDEX.md](docs/INDEX.md) - Documentation map

---

## ✅ Success Criteria

### Testing Complete When:
- [x] All services running and healthy
- [x] Authentication persists across navigation
- [x] Tenant launch button works correctly
- [x] Pricing page CRUD operations functional
- [x] API integration working

### Phase 1 Complete When:
- [x] Backend sets httpOnly cookies
- [x] Frontend doesn't manipulate cookies
- [x] DevTools shows HttpOnly flag checked
- [x] JavaScript cannot access cookie
- [x] Authentication still works correctly

### Ready for Production When:
- [x] All 5 security phases complete
- [x] Full test coverage
- [x] Performance benchmarks met
- [x] Documentation up-to-date
- [x] Monitoring configured

---

**Status:** Ready to begin testing! 🚀

Follow this guide in order for best results. Good luck!
