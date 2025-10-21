# Authentication Security Roadmap

## Overview

This document outlines the current authentication security implementation and the roadmap for production hardening.

---

## ✅ Current Implementation (Development Mode)

### Frontend (Both SaaS Admin & Tenant Admin)

#### Middleware-Based Authentication
- **File:** `middleware.ts` (both frontends)
- **Pattern:** Next.js 14 middleware (industry standard)
- **Coverage:** Global auth protection for all routes
- **Execution:** Server-side, before page render

#### Cookie Strategy (Development)
```javascript
// Client-side cookie setting (development mode)
document.cookie = `admin_token=${token}; path=/; max-age=${7 * 24 * 60 * 60}; samesite=lax`;
```

**Current Cookie Attributes:**
- ✅ `path=/` - Available across entire app
- ✅ `max-age=7 days` - Auto-expires after 1 week
- ✅ `samesite=lax` - Basic CSRF protection
- ❌ `httpOnly` - NOT set (requires backend support)
- ❌ `secure` - NOT set (requires HTTPS)

#### Dual Storage Pattern
- **LocalStorage:** Client-side React components
- **Cookies:** Server-side middleware
- **Why Both:** Complete SSR + CSR coverage

#### Cookie Names
- **SaaS Admin:** `admin_token`
- **Tenant Admin:** `tenant_admin_token`

---

## 🔐 Production Hardening Roadmap

### Phase 1: Backend-Controlled Cookies (HIGH PRIORITY)

**Goal:** Move cookie management to backend for httpOnly support

#### Backend Changes Required

**File:** `saas-admin-service/internal/handlers/saas_admin_handler.go`

```go
func (h *SaaSAdminHandler) Login(c *gin.Context) {
    // ... existing login logic ...

    // Set httpOnly cookie (server-side)
    c.SetCookie(
        "admin_token",           // name
        token,                   // value
        7*24*60*60,             // maxAge (7 days in seconds)
        "/",                     // path
        "",                      // domain (empty for same-domain)
        false,                   // secure (set to true in production with HTTPS)
        true,                    // httpOnly (prevents JavaScript access)
    )

    // Also return token in response for backward compatibility
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "token":  token,
        "user":   user,
    })
}

func (h *SaaSAdminHandler) Logout(c *gin.Context) {
    // Clear cookie
    c.SetCookie(
        "admin_token",
        "",
        -1,    // maxAge -1 deletes the cookie
        "/",
        "",
        false,
        true,
    )

    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

**File:** `tenant-admin-service/internal/handlers/auth_handler.go`

```go
// Same pattern, use cookie name: "tenant_admin_token"
```

#### Frontend Changes (After Backend Implementation)

**File:** `lib/api/auth.ts` (both frontends)

```typescript
export const authApi = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    // Backend now sets httpOnly cookie automatically
    const response = await post<LoginResponse, LoginRequest>('/api/v1/auth/login', credentials);

    // Only store user data in localStorage (token is in httpOnly cookie)
    if (response.user) {
      localStorage.setItem('admin_user', JSON.stringify(response.user));
    }

    // NO NEED to set cookie manually - backend does it!
    return response;
  },

  logout: async (): Promise<void> => {
    // Backend clears httpOnly cookie automatically
    await post('/api/v1/auth/logout');

    // Clear localStorage
    localStorage.removeItem('admin_user');
  },
};
```

**Benefits:**
- ✅ JavaScript cannot access token (XSS protection)
- ✅ Token automatically included in API requests
- ✅ Centralized cookie management

---

### Phase 2: CSRF Protection (HIGH PRIORITY)

**Goal:** Protect against Cross-Site Request Forgery attacks

#### Backend Implementation

**File:** `saas-admin-service/internal/middleware/csrf.go` (NEW)

```go
package middleware

import (
    "crypto/rand"
    "encoding/base64"
    "github.com/gin-gonic/gin"
    "net/http"
)

// CSRF token storage (use Redis in production)
var csrfTokens = make(map[string]bool)

// GenerateCSRFToken generates a random CSRF token
func GenerateCSRFToken() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// CSRFMiddleware validates CSRF tokens on state-changing requests
func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only check CSRF for state-changing methods
        if c.Request.Method == "POST" || c.Request.Method == "PUT" ||
           c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {

            // Get CSRF token from header
            csrfToken := c.GetHeader("X-CSRF-Token")

            if csrfToken == "" || !csrfTokens[csrfToken] {
                c.JSON(http.StatusForbidden, gin.H{
                    "error": "Invalid CSRF token",
                })
                c.Abort()
                return
            }
        }

        c.Next()
    }
}

// GetCSRFToken endpoint to get a CSRF token
func GetCSRFToken(c *gin.Context) {
    token, err := GenerateCSRFToken()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    csrfTokens[token] = true
    c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}
```

#### Frontend Implementation

**File:** `lib/api/client.ts` (both frontends)

```typescript
// Get CSRF token on app initialization
let csrfToken: string | null = null;

async function getCSRFToken(): Promise<string> {
  if (csrfToken) return csrfToken;

  const response = await axios.get('/api/v1/csrf-token');
  csrfToken = response.data.csrf_token;
  return csrfToken;
}

// Axios request interceptor - add CSRF token to all requests
apiClient.interceptors.request.use(async (config) => {
  // Add CSRF token for state-changing requests
  if (['post', 'put', 'delete', 'patch'].includes(config.method?.toLowerCase() || '')) {
    const token = await getCSRFToken();
    config.headers['X-CSRF-Token'] = token;
  }

  return config;
});
```

---

### Phase 3: Refresh Token Mechanism (MEDIUM PRIORITY)

**Goal:** Extend sessions without re-login, improve security with short-lived access tokens

#### Token Strategy
- **Access Token:** Short-lived (15 minutes), used for API requests
- **Refresh Token:** Long-lived (7 days), used to get new access tokens

#### Backend Implementation

**Database Schema:**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id)
);
```

**Login Response:**
```go
type LoginResponse struct {
    AccessToken  string `json:"access_token"`  // 15 min expiry
    RefreshToken string `json:"refresh_token"` // 7 day expiry
    ExpiresIn    int    `json:"expires_in"`    // seconds until access token expires
    User         User   `json:"user"`
}
```

**Refresh Endpoint:**
```go
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    refreshToken := c.PostForm("refresh_token")

    // Validate refresh token
    // Generate new access token
    // Return new tokens
}
```

#### Frontend Implementation

**Auto-refresh before expiry:**
```typescript
// Automatically refresh token 1 minute before expiry
let refreshTimeout: NodeJS.Timeout;

function scheduleTokenRefresh(expiresIn: number) {
  const refreshTime = (expiresIn - 60) * 1000; // 1 minute before expiry

  refreshTimeout = setTimeout(async () => {
    await authApi.refreshAccessToken();
  }, refreshTime);
}

export const authApi = {
  async refreshAccessToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    const response = await post('/api/v1/auth/refresh', { refresh_token: refreshToken });

    // Update tokens
    localStorage.setItem('access_token', response.access_token);

    // Schedule next refresh
    scheduleTokenRefresh(response.expires_in);
  }
};
```

---

### Phase 4: Rate Limiting (HIGH PRIORITY)

**Goal:** Prevent brute-force attacks on login

#### Backend Implementation

Already implemented in `shared-resilience` library:

```go
// Apply rate limiting to auth endpoints
r.POST("/api/v1/auth/login",
    resilience.RateLimitMiddleware(5, time.Minute), // 5 attempts per minute
    authHandler.Login,
)
```

**Additional Lockout Logic:**
```go
// Lock account after 5 failed attempts for 15 minutes
type LoginAttempt struct {
    Username      string
    Attempts      int
    LockedUntil   time.Time
}
```

---

### Phase 5: HTTPS & Secure Cookies (PRODUCTION ONLY)

**Goal:** Encrypt traffic and set secure cookies

#### Environment-Based Cookie Configuration

**File:** `lib/api/auth.ts`

```typescript
const isProduction = process.env.NODE_ENV === 'production';

// Cookie attributes based on environment
const cookieAttributes = isProduction
  ? `secure; samesite=strict` // Production (HTTPS)
  : `samesite=lax`;           // Development (HTTP)

document.cookie = `admin_token=${token}; path=/; max-age=${7 * 24 * 60 * 60}; ${cookieAttributes}`;
```

#### Backend Cookie Configuration

```go
c.SetCookie(
    "admin_token",
    token,
    7*24*60*60,
    "/",
    "",
    os.Getenv("ENVIRONMENT") == "production", // secure flag (true in production)
    true, // httpOnly
)
```

---

## 📋 Security Checklist

### Development (Current)
- ✅ Middleware-based authentication
- ✅ Cookie + LocalStorage dual storage
- ✅ SameSite=lax CSRF protection
- ✅ 7-day token expiration
- ✅ Automatic token validation
- ✅ Protected routes
- ❌ httpOnly cookies (requires backend)
- ❌ Secure cookies (requires HTTPS)
- ❌ CSRF tokens (requires backend)
- ❌ Refresh tokens (requires backend)

### Production (Required Before Launch)
- ⬜ httpOnly cookies (Phase 1)
- ⬜ CSRF token protection (Phase 2)
- ⬜ Refresh token mechanism (Phase 3)
- ⬜ Rate limiting on login (Phase 4)
- ⬜ HTTPS enforcement (Phase 5)
- ⬜ Secure cookie flag (Phase 5)
- ⬜ Token revocation on logout
- ⬜ Session management in Redis
- ⬜ Audit logging for auth events

---

## 🚀 Implementation Priority

**For Next Sprint:**
1. **Phase 1:** httpOnly cookies (Backend changes)
2. **Phase 2:** CSRF protection (Backend + Frontend)
3. **Phase 4:** Verify rate limiting is active

**For Production:**
1. **Phase 3:** Refresh tokens
2. **Phase 5:** HTTPS + Secure cookies
3. Session management with Redis
4. Comprehensive audit logging

---

## 🔍 Testing Checklist

### Current Implementation Tests
- ✅ Root URL navigation while logged in → stays logged in
- ✅ Root URL navigation while logged out → redirects to login
- ✅ Protected route access without auth → redirects to login
- ✅ Login page access while authenticated → redirects to dashboard
- ✅ Token persists across page refreshes
- ✅ Logout clears both cookie and localStorage

### Future Tests (Post-Implementation)
- ⬜ httpOnly cookie prevents JavaScript access
- ⬜ CSRF token validation blocks unauthorized requests
- ⬜ Refresh token extends session without re-login
- ⬜ Rate limiting blocks brute-force attempts
- ⬜ Secure cookie only works over HTTPS

---

## 📚 References

- **OWASP Authentication Cheat Sheet:** https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- **Next.js Middleware:** https://nextjs.org/docs/app/building-your-application/routing/middleware
- **httpOnly Cookies:** https://owasp.org/www-community/HttpOnly
- **CSRF Protection:** https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html

---

## 📝 Notes

**Current Status:** Development-ready with client-side cookie management
**Next Step:** Implement Phase 1 (httpOnly cookies) with backend team
**Security Level:** Medium (suitable for development, NOT production)
**Production Readiness:** 40% complete (middleware done, backend hardening needed)
