# SaaS Admin Session Management - Proper Architecture Fix

**Issue**: SaaS Admin service cannot login without Tenant Admin service running
**Root Cause**: Login handler makes blocking, synchronous call to Tenant Admin for session creation
**Current Behavior**: Login fails with "Session creation failed" if Tenant Admin is down

---

## Current (Broken) Architecture

```
SaaS Admin Login Flow:
1. Validate credentials ✓
2. Call Tenant Admin to create session ✗ (BLOCKS HERE if Tenant Admin is down)
3. Return session cookie
```

**Problem**: Tight coupling between SaaS Admin and Tenant Admin services

---

## Proper Solution (NOT Band-Aid)

### Option 1: Local Session Management (Recommended)

**Principle**: Each service manages its own sessions

```
SaaS Admin Database:
├── sessions table (ALREADY EXISTS!)
│   ├── id (varchar 128, PRIMARY KEY)
│   ├── user_id (bigint, FK to saas_admin_users)
│   ├── tenant_id (bigint)
│   ├── ip_address
│   ├── user_agent
│   ├── is_active
│   ├── expires_at
│   └── created_at/updated_at
└── user_sessions table (ALREADY EXISTS!)
    ├── id (bigserial)
    ├── user_id (bigint)
    ├── token (text, UNIQUE)
    ├── expires_at
    ├── ip_address
    ├── user_agent
    └── is_active
```

**Implementation Steps**:

1. **Create Session Model**
```go
// internal/models/session.go
package models

type Session struct {
    ID        string    `gorm:"primaryKey;size:128"`
    UserID    uint      `gorm:"not null;index"`
    TenantID  uint      `gorm:"not null;index"`
    IPAddress string    `gorm:"size:45"`
    UserAgent string    `gorm:"type:text"`
    IsActive  bool      `gorm:"default:true;index"`
    LastSeen  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
    ExpiresAt time.Time `gorm:"not null;index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserSession struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    UserID    uint      `gorm:"not null;index"`
    Token     string    `gorm:"not null;uniqueIndex"`
    ExpiresAt time.Time `gorm:"not null;index"`
    IPAddress string    `gorm:"type:text"`
    UserAgent string    `gorm:"type:text"`
    IsActive  bool      `gorm:"default:true;index"`
}
```

2. **Create Session Service**
```go
// internal/services/session_service.go
package services

type SessionService struct {
    db     *gorm.DB
    logger *zap.Logger
}

func (s *SessionService) CreateSession(ctx context.Context, userID uint, ipAddress, userAgent string) (*models.Session, error) {
    session := &models.Session{
        ID:        uuid.New().String(),
        UserID:    userID,
        TenantID:  0, // SaaS admin doesn't belong to a tenant
        IPAddress: ipAddress,
        UserAgent: userAgent,
        IsActive:  true,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }

    if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
        return nil, err
    }

    return session, nil
}

func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
    var session models.Session
    err := s.db.WithContext(ctx).
        Where("id = ? AND is_active = ? AND expires_at > ?", sessionID, true, time.Now()).
        First(&session).Error

    if err != nil {
        return nil, err
    }

    // Update last_seen
    s.db.WithContext(ctx).Model(&session).Update("last_seen", time.Now())

    return &session, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
    return s.db.WithContext(ctx).
        Model(&models.Session{}).
        Where("id = ?", sessionID).
        Update("is_active", false).Error
}
```

3. **Update Login Handler**
```go
// internal/handlers/saas_admin_handler.go

func (h *SaaSAdminHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
        return
    }

    // Validate credentials
    user, err := h.service.ValidateAdminCredentials(c.Request.Context(), req.Username, req.Password)
    if err != nil {
        h.logger.Warn("Login failed", zap.String("username", req.Username), zap.Error(err))
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Create LOCAL session in saas_admin database
    session, err := h.sessionService.CreateSession(
        c.Request.Context(),
        user.ID,
        c.ClientIP(),
        c.Request.UserAgent(),
    )
    if err != nil {
        h.logger.Error("Failed to create session", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Session creation failed"})
        return
    }

    // Set session cookie
    c.SetCookie("session_id", session.ID, 86400, "/", "", false, true)

    c.JSON(http.StatusOK, gin.H{
        "success":    true,
        "session_id": session.ID,
        "message":    "Login successful",
        "user": gin.H{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
        },
    })
}
```

4. **Add Session Middleware for Protected Routes**
```go
// internal/middleware/auth.go

func SessionAuth(sessionService *services.SessionService, logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionID, err := c.Cookie("session_id")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
            c.Abort()
            return
        }

        session, err := sessionService.ValidateSession(c.Request.Context(), sessionID)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
            c.Abort()
            return
        }

        // Set user context
        c.Set("user_id", session.UserID)
        c.Set("session_id", session.ID)
        c.Next()
    }
}
```

---

### Option 2: JWT-Based Authentication (Alternative)

**Principle**: Stateless authentication without session storage

```go
type JWTClaims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func (h *SaaSAdminHandler) Login(c *gin.Context) {
    // ... validate credentials ...

    // Create JWT token
    claims := &JWTClaims{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        Role:     "admin",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "saas-admin-service",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

    c.SetCookie("token", tokenString, 86400, "/", "", false, true)
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "token":   tokenString,
    })
}
```

**Pros**: No database lookups, stateless, scales horizontally
**Cons**: Cannot revoke tokens until expiry, larger cookie size

---

## Comparison

| Aspect | Local Sessions | JWT | Current (Tenant Admin) |
|--------|---------------|-----|------------------------|
| **Independence** | ✅ Fully independent | ✅ Fully independent | ❌ Depends on Tenant Admin |
| **Performance** | ⚠️ DB lookup per request | ✅ No DB lookup | ❌ Network call + DB |
| **Revocation** | ✅ Immediate | ❌ Wait for expiry | ✅ Immediate |
| **Scalability** | ⚠️ Requires DB | ✅ Stateless | ❌ Service dependency |
| **Session Management** | ✅ Full control | ⚠️ Limited | ❌ Delegated |
| **Multi-device** | ✅ Track all devices | ⚠️ Separate tokens | ✅ Track all devices |

---

## Recommended Approach

**Hybrid: Local Sessions + Optional JWT for API**

1. **Web Login**: Use local sessions (database-backed)
   - Full session management
   - Can revoke immediately
   - Track IP, user agent, last seen

2. **API Access**: Provide JWT tokens
   - Stateless API calls
   - Mobile/external integrations
   - No session overhead

3. **Tenant Admin Sync**: Make it optional and async
   - Don't block login
   - Sync in background for cross-service session awareness
   - Log failures but don't fail login

---

## Why NOT a Band-Aid Solution

❌ **Bad (Band-Aid)**:
```go
// Just generate UUID and return, no storage
sessionID := uuid.New().String()
c.SetCookie("session_id", sessionID, ...)
return // Session not stored anywhere!
```

**Problems**:
- Session validation impossible
- Cannot revoke sessions
- No session tracking
- Security risk

✅ **Good (Proper)**:
```go
// Create session in database
session, err := h.sessionService.CreateSession(...)
if err != nil { return error }

// Store in DB, then return
c.SetCookie("session_id", session.ID, ...)
return
```

**Benefits**:
- Session validation works
- Can revoke sessions
- Full audit trail
- Secure

---

## Implementation Checklist

- [ ] Create `internal/models/session.go` with Session and UserSession models
- [ ] Create `internal/services/session_service.go` with CRUD operations
- [ ] Create `internal/middleware/auth.go` with session validation middleware
- [ ] Update `internal/handlers/saas_admin_handler.go` to use SessionService
- [ ] Remove Tenant Admin dependency from login flow
- [ ] Add session cleanup job (delete expired sessions)
- [ ] Add session management endpoints (list, revoke)
- [ ] Update tests
- [ ] Update documentation

---

## Migration Path

1. **Phase 1**: Implement local session management (keeps existing behavior)
2. **Phase 2**: Make Tenant Admin sync optional/async
3. **Phase 3**: Remove Tenant Admin dependency completely
4. **Phase 4**: Add JWT support for API clients

---

## Conclusion

The current architecture is **fundamentally flawed** - SaaS Admin should manage its own sessions.

**DO NOT** use the band-aid solution I implemented earlier (generating UUID without storage).

**DO** implement proper local session management using the existing `sessions` table.

This follows microservices best practices where each service is autonomous and independent.
