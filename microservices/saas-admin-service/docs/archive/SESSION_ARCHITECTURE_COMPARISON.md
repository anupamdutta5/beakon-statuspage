# Session Management Architecture - Industry Best Practices

## TL;DR - Recommendation

**For SaaS Admin Service: Hybrid Approach (JWT + Redis Sessions)**

- **JWT for API authentication** (stateless, scalable)
- **Redis-backed sessions for web UI** (fast, revocable)
- **Local PostgreSQL as fallback** (when Redis unavailable)

---

## Industry Standard Approaches

### 1. JWT (JSON Web Tokens) ⭐⭐⭐⭐⭐

**Used by**: Auth0, Okta, Firebase, AWS Cognito, Stripe API, GitHub API

**Architecture**:
```
┌─────────────┐                    ┌──────────────┐
│   Client    │──── JWT Token ────▶│  API Server  │
│             │                    │ (Stateless)  │
└─────────────┘                    └──────────────┘
                                          │
                                          ▼
                                   Verify signature
                                   (No DB lookup!)
```

**Pros**:
- ✅ **Stateless** - No session storage needed
- ✅ **Horizontally scalable** - Any server can validate
- ✅ **Performance** - No database/Redis lookup per request
- ✅ **Cross-service auth** - Same token works across microservices
- ✅ **Industry standard** - Well-understood, proven at scale
- ✅ **Mobile-friendly** - Easy to implement in apps

**Cons**:
- ❌ **Cannot revoke** until expiry (mitigate with short TTL + refresh tokens)
- ❌ **Token size** - Larger than session ID (mitigate with claims optimization)
- ❌ **Secret rotation** - Complex to rotate signing keys

**Best For**:
- API authentication
- Microservices architectures
- High-scale systems (millions of users)
- Mobile/SPA applications

**Example (Go)**:
```go
// Create token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": user.ID,
    "email":   user.Email,
    "exp":     time.Now().Add(15 * time.Minute).Unix(),
})
tokenString, _ := token.SignedString([]byte(secret))

// Validate token
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil
})
```

---

### 2. Redis Sessions ⭐⭐⭐⭐⭐

**Used by**: Netflix, Uber, Twitter, Instagram, Airbnb

**Architecture**:
```
┌─────────────┐                    ┌──────────────┐
│   Client    │──── Session ID ───▶│  API Server  │
│             │                    │              │
└─────────────┘                    └──────┬───────┘
                                          │
                                          ▼
                                   ┌──────────────┐
                                   │    Redis     │
                                   │ (In-Memory)  │
                                   └──────────────┘
```

**Pros**:
- ✅ **Fast** - Sub-millisecond lookups (in-memory)
- ✅ **Revocable** - Delete session immediately
- ✅ **TTL built-in** - Automatic expiration
- ✅ **Distributed** - Redis cluster for HA
- ✅ **Session data** - Can store complex objects
- ✅ **Atomic operations** - INCR, SETEX, etc.

**Cons**:
- ⚠️ **Memory cost** - All sessions in RAM
- ⚠️ **Additional service** - Redis must be available
- ⚠️ **Persistence** - Need Redis persistence (RDB/AOF)

**Best For**:
- Web applications
- Real-time applications
- Systems requiring immediate revocation
- High-traffic applications

**Example (Go)**:
```go
// Create session
sessionID := uuid.New().String()
sessionData, _ := json.Marshal(user)
rdb.Set(ctx, "session:"+sessionID, sessionData, 24*time.Hour)

// Validate session
val, err := rdb.Get(ctx, "session:"+sessionID).Result()
if err == redis.Nil {
    return ErrSessionExpired
}
```

---

### 3. Database Sessions ⭐⭐⭐

**Used by**: Traditional web apps, Django, Rails (default)

**Architecture**:
```
┌─────────────┐                    ┌──────────────┐
│   Client    │──── Session ID ───▶│  API Server  │
│             │                    │              │
└─────────────┘                    └──────┬───────┘
                                          │
                                          ▼
                                   ┌──────────────┐
                                   │  PostgreSQL  │
                                   │   Sessions   │
                                   └──────────────┘
```

**Pros**:
- ✅ **Persistent** - Survives restarts
- ✅ **No extra service** - Uses existing DB
- ✅ **Audit trail** - Full session history
- ✅ **Revocable** - Update is_active = false

**Cons**:
- ❌ **Slower** - DB query per request (10-50ms)
- ❌ **Database load** - Every request hits DB
- ❌ **Scaling** - DB becomes bottleneck

**Best For**:
- Small to medium apps
- When Redis not available
- Compliance/audit requirements
- Financial applications

---

### 4. Hybrid JWT + Redis ⭐⭐⭐⭐⭐ (RECOMMENDED)

**Used by**: Shopify, Slack, GitHub, GitLab

**Architecture**:
```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ├─── Access Token (JWT, 15min) ────▶ API Validation (Stateless)
       │
       └─── Refresh Token (ID) ───────────▶ Redis Lookup (Stateful)
```

**How it Works**:

1. **Login**:
   - Generate short-lived JWT (15 min)
   - Generate refresh token, store in Redis (7 days)
   - Return both to client

2. **API Requests**:
   - Client sends JWT
   - Server validates JWT (no DB/Redis lookup)
   - Fast and stateless

3. **Token Refresh**:
   - JWT expires after 15 min
   - Client sends refresh token
   - Server validates refresh token (Redis lookup)
   - Issue new JWT

4. **Logout/Revoke**:
   - Delete refresh token from Redis
   - Current JWT expires in max 15 min

**Pros**:
- ✅ **Best of both worlds** - Fast + revocable
- ✅ **Scalable** - Most requests are stateless
- ✅ **Secure** - Can revoke within 15 min
- ✅ **Efficient** - Minimal Redis lookups
- ✅ **Industry standard** - OAuth 2.0 pattern

**Cons**:
- ⚠️ **Complexity** - Two token types to manage
- ⚠️ **Implementation** - More code than pure JWT

**Best For**:
- **Production SaaS applications** ← YOUR USE CASE
- High-scale systems
- APIs + Web UIs
- Security-sensitive applications

**Example (Go)**:
```go
func (h *Handler) Login(c *gin.Context) {
    // Validate credentials
    user := validateCredentials(...)

    // 1. Generate short-lived access token (JWT)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "exp":     time.Now().Add(15 * time.Minute).Unix(), // Short TTL
    })
    accessTokenString, _ := accessToken.SignedString([]byte(secret))

    // 2. Generate refresh token (stored in Redis)
    refreshToken := uuid.New().String()
    sessionData := map[string]interface{}{
        "user_id":    user.ID,
        "ip_address": c.ClientIP(),
        "user_agent": c.Request.UserAgent(),
    }
    sessionJSON, _ := json.Marshal(sessionData)
    rdb.Set(ctx, "refresh:"+refreshToken, sessionJSON, 7*24*time.Hour)

    c.JSON(200, gin.H{
        "access_token":  accessTokenString,  // Use for API calls
        "refresh_token": refreshToken,        // Use to get new access token
        "token_type":    "Bearer",
        "expires_in":    900, // 15 minutes
    })
}

func (h *Handler) Refresh(c *gin.Context) {
    refreshToken := c.PostForm("refresh_token")

    // Validate refresh token in Redis
    sessionData, err := rdb.Get(ctx, "refresh:"+refreshToken).Result()
    if err == redis.Nil {
        c.JSON(401, gin.H{"error": "Invalid refresh token"})
        return
    }

    // Generate new access token
    accessToken := jwt.NewWithClaims(...)
    c.JSON(200, gin.H{"access_token": accessToken})
}

func (h *Handler) Logout(c *gin.Context) {
    refreshToken := c.PostForm("refresh_token")

    // Delete refresh token from Redis
    rdb.Del(ctx, "refresh:"+refreshToken)

    c.JSON(200, gin.H{"message": "Logged out"})
}
```

---

## Real-World Examples

### GitHub
- **Access Token**: JWT (1 hour)
- **Refresh Token**: Opaque ID in database
- **Revocation**: Delete refresh token, access token expires in 1 hour

### Shopify
- **Access Token**: JWT (15 min)
- **Session**: Redis-backed
- **API**: JWT only
- **Admin Panel**: Session cookies

### Stripe
- **API Keys**: Long-lived, stored in database
- **JWT**: For temporary access
- **Webhooks**: Signature verification (no session)

### AWS Cognito
- **ID Token**: JWT (1 hour) - User identity
- **Access Token**: JWT (1 hour) - API access
- **Refresh Token**: Opaque (30 days) - Get new tokens

---

## Recommendation for SaaS Admin Service

### **Hybrid: JWT + Redis (with PostgreSQL fallback)**

```go
// Architecture
┌──────────────────────────────────────────────────────┐
│                 SaaS Admin Service                   │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Web UI Login                    API Authentication  │
│  ↓                               ↓                   │
│  Access Token (JWT, 15min)       API Key (DB)       │
│  Refresh Token (Redis, 7d)       or JWT             │
│                                                      │
│  Fallback: PostgreSQL sessions if Redis down        │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Implementation**:

1. **Phase 1: JWT + PostgreSQL** (No Redis dependency)
   - Access Token: JWT (15 min)
   - Refresh Token: PostgreSQL `user_sessions` table
   - Works immediately, no new infrastructure

2. **Phase 2: Add Redis** (Production optimization)
   - Move refresh tokens to Redis
   - Keep PostgreSQL as fallback
   - Performance improvement

3. **Phase 3: Multi-device support**
   - Track all sessions per user
   - Allow "logout all devices"
   - Session management UI

**Code Structure**:
```
internal/
├── auth/
│   ├── jwt.go           # JWT generation & validation
│   ├── refresh.go       # Refresh token management
│   └── middleware.go    # Auth middleware
├── models/
│   └── session.go       # Session model
└── services/
    └── session_service.go # Session CRUD with Redis + DB fallback
```

---

## Performance Comparison

| Approach | Latency | Throughput | Memory | Revocation Time |
|----------|---------|------------|--------|-----------------|
| **JWT Only** | 0.1ms | 50k req/s | Low | Minutes-Hours |
| **Redis Sessions** | 1-2ms | 10k req/s | High | Immediate |
| **DB Sessions** | 10-50ms | 1k req/s | Low | Immediate |
| **Hybrid JWT+Redis** | 0.1ms (access)<br>2ms (refresh) | 45k req/s | Medium | 15 min |

---

## Security Considerations

### JWT
- ✅ Verify signature algorithm (`alg` header)
- ✅ Validate `exp`, `iss`, `aud` claims
- ✅ Use short TTL (15 min recommended)
- ✅ Rotate secrets regularly
- ❌ Don't store sensitive data in payload
- ❌ Don't use `none` algorithm

### Redis
- ✅ Use authentication (`requirepass`)
- ✅ Enable TLS for production
- ✅ Set maxmemory-policy (eviction)
- ✅ Use Redis Cluster for HA
- ✅ Enable persistence (AOF)

### Database
- ✅ Index session_id column
- ✅ Clean up expired sessions (cron job)
- ✅ Use prepared statements
- ✅ Connection pooling

---

## Final Recommendation

**For Your SaaS Admin Service:**

```
Phase 1 (Immediate): JWT + PostgreSQL
- Implement JWT for access tokens
- Use existing user_sessions table for refresh tokens
- No new infrastructure needed
- Time: 4-6 hours

Phase 2 (Later): Add Redis
- Move refresh tokens to Redis
- Keep PostgreSQL as fallback
- Add Redis to docker-compose
- Time: 2-3 hours

Phase 3 (Future): Session Management UI
- Show active sessions
- Revoke individual sessions
- "Logout all devices"
- Time: 4-6 hours
```

**Why This Approach?**
1. ✅ Industry standard (used by GitHub, Shopify, Stripe)
2. ✅ Scalable (handles millions of users)
3. ✅ Secure (revocable within 15 min)
4. ✅ Fast (stateless validation)
5. ✅ Flexible (works with/without Redis)
6. ✅ Battle-tested (OAuth 2.0 pattern)

---

## Answer: Which is Best?

**Industry Best Practice**: **Hybrid JWT + Redis** (or JWT + PostgreSQL as starting point)

**Most Robust**: **Hybrid JWT + Redis** with database fallback

**Your Situation**: Start with **JWT + PostgreSQL**, add Redis later when needed

This gives you:
- Immediate solution (no new dependencies)
- Production-grade security
- Path to scale
- Industry standard approach
