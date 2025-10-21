# URGENT: Features Lost Due to git reset --hard HEAD

**Date**: 2025-01-20
**Caused by**: Running `git reset --hard HEAD` which discarded ALL uncommitted changes
**Current State**: Code is at October 8th commit (cad2315)

---

## What Was Lost

All uncommitted work including:
1. Max users display and update functionality
2. Password reset functionality with bcrypt
3. Modern admin login modal UI
4. URL hash-based tab persistence

---

## Fix #1: Max Users Display and Update

### Database Schema (Already Exists - DO NOT MODIFY)
```sql
-- tenants table in tenant_admin_db
max_users BIGINT NULL
CHECK (max_users IS NULL OR max_users > 0)
```

### Fix Location 1: Service Layer
**File**: `microservices/saas-admin-service/internal/services/saas_admin_service.go`
**Line**: Around 1205 (UpdateTenant function)

**What to Add**: Handle max_users field properly in tenant updates

```go
func (s *SaaSAdminService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates *models.SaaSTenant) error {
    s.logger.Info("Updating tenant", zap.String("id", tenantID.String()))

    // Build updates map for GORM
    updatesMap := make(map[string]interface{})

    // Handle max_users specifically
    // If max_users is provided and > 0, update it
    // If max_users is 0 or not provided, don't update (keep existing value)
    if updates.MaxUsers != nil {
        if *updates.MaxUsers > 0 {
            updatesMap["max_users"] = *updates.MaxUsers
        }
        // If 0 is sent, it means "unlimited" - set to NULL
        if *updates.MaxUsers == 0 {
            updatesMap["max_users"] = nil
        }
    }

    // Add other fields to updatesMap as needed
    if updates.Name != "" {
        updatesMap["name"] = updates.Name
    }
    if updates.ContactEmail != "" {
        updatesMap["contact_email"] = updates.ContactEmail
    }
    // ... other fields

    // Perform update
    if err := s.db.Model(&models.SaaSTenant{}).Where("id = ?", tenantID).Updates(updatesMap).Error; err != nil {
        s.logger.Error("Failed to update tenant", zap.Error(err))
        return fmt.Errorf("failed to update tenant: %w", err)
    }

    return nil
}
```

### Fix Location 2: Models
**File**: `microservices/saas-admin-service/internal/models/saas_tenant.go` (or similar)

**What to Check**: Ensure SaaSTenant model has MaxUsers field defined correctly

```go
type SaaSTenant struct {
    ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Name          string     `gorm:"not null" json:"name"`
    ContactEmail  string     `gorm:"not null" json:"contact_email"`
    MaxUsers      *int64     `gorm:"column:max_users" json:"max_users,omitempty"` // POINTER for nullable
    // ... other fields
}
```

**CRITICAL**: Use `*int64` (pointer) NOT `int64` because the column is nullable

---

## Fix #2: Password Reset with bcrypt

### Fix Location: Service Layer
**File**: `microservices/saas-admin-service/internal/services/saas_admin_service.go`

**Function to Add/Fix**: ResetAdminPassword (or similar)

```go
// ResetAdminPassword resets an admin user's password
func (s *SaaSAdminService) ResetAdminPassword(ctx context.Context, adminID uuid.UUID, newPassword string) error {
    s.logger.Info("Resetting admin password", zap.String("admin_id", adminID.String()))

    // Validate password strength
    if len(newPassword) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }

    // Hash password with bcrypt (cost factor 10)
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
    if err != nil {
        s.logger.Error("Failed to hash password", zap.Error(err))
        return fmt.Errorf("failed to hash password: %w", err)
    }

    // Update password in database
    if err := s.db.Model(&models.SaaSAdmin{}).
        Where("id = ?", adminID).
        Update("password_hash", string(hashedPassword)).Error; err != nil {
        s.logger.Error("Failed to update password", zap.Error(err))
        return fmt.Errorf("failed to update password: %w", err)
    }

    s.logger.Info("Password reset successful", zap.String("admin_id", adminID.String()))
    return nil
}
```

### Handler to Add
**File**: `microservices/saas-admin-service/internal/handlers/saas_admin_handler.go`

```go
// ResetAdminPassword handles password reset requests
func (h *SaaSAdminHandler) ResetAdminPassword(c *gin.Context) {
    adminIDStr := c.Param("id")
    if adminIDStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Admin ID is required"})
        return
    }

    adminID, err := uuid.Parse(adminIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin ID"})
        return
    }

    var req struct {
        NewPassword string `json:"new_password" binding:"required,min=8"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    if err := h.service.ResetAdminPassword(c.Request.Context(), adminID, req.NewPassword); err != nil {
        h.logger.Error("Failed to reset password", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "Password reset successfully",
    })
}
```

### Route to Add
**File**: `microservices/saas-admin-service/cmd/main.go` (or routes file)

```go
adminRoutes.PUT("/admin-users/:id/password", handler.ResetAdminPassword)
```

---

## Fix #3: Modern Admin Login Modal UI

### File to Fix
**File**: `microservices/saas-admin-service/web/templates/partials/scripts.html`

**Issue**: The October 8th commit has an OLD basic login modal. The modern version had:
- Gradient background
- Modern styling with rounded corners
- Better UX
- Proper branding

**What to Do**:
1. Check if there's a backup of the modern login modal
2. If not, redesign the login modal in `scripts.html` with modern Bootstrap 5 components
3. Look for references to "old login page" in git history to see what it looked like before

### Current Problem (Line 112-117 in scripts.html)
```javascript
// This code shows login modal on EVERY page load (auto-logout bug)
document.addEventListener('DOMContentLoaded', async () => {
    const authenticated = await checkAuthentication();
    if (!authenticated) {
        showLoginModal();
    }
});
```

**This should be REMOVED or made conditional** - it forces login on every refresh

---

## Fix #4: URL Hash-Based Tab Persistence

### File to Fix
**File**: `microservices/saas-admin-service/web/templates/partials/scripts.html`

**What to Add**: URL hash-based tab switching that persists across page reloads

```javascript
// Tab persistence with URL hash
document.addEventListener('DOMContentLoaded', function() {
    // Handle initial hash on page load
    handleHashChange();

    // Listen for hash changes
    window.addEventListener('hashchange', handleHashChange);

    // Add click handlers to all nav links
    document.querySelectorAll('.nav-link[data-bs-toggle="tab"]').forEach(link => {
        link.addEventListener('click', function(e) {
            const targetId = this.getAttribute('href');
            if (targetId) {
                window.location.hash = targetId;
            }
        });
    });
});

function handleHashChange() {
    let hash = window.location.hash || '#overview'; // Default to overview

    // Remove 'active' from all tabs and panes
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.remove('show', 'active');
    });

    // Activate the correct tab
    const targetLink = document.querySelector(`.nav-link[href="${hash}"]`);
    const targetPane = document.querySelector(hash);

    if (targetLink && targetPane) {
        targetLink.classList.add('active');
        targetPane.classList.add('show', 'active');
    }
}

// Also update showTab function to use hash
function showTab(tabId) {
    window.location.hash = tabId;
}
```

### Files to Fix: Remove Hardcoded Active Classes

**File 1**: `microservices/saas-admin-service/web/templates/partials/sidebar.html`
- **Line ~14**: Remove `class="nav-link active"` from Overview tab
- Make it just `class="nav-link"`

**File 2**: `microservices/saas-admin-service/web/templates/partials/dashboard.html`
- **Line ~6**: Remove `class="tab-pane fade show active"` from first tab pane
- Make it just `class="tab-pane fade"`

---

## How to Test After Fixes

### 1. Test Max Users
```bash
# Get a tenant ID
curl http://localhost:8098/api/v1/tenants

# Update max_users
curl -X PUT http://localhost:8098/api/v1/tenants/TENANT_ID \
  -H 'Content-Type: application/json' \
  -d '{"max_users": 50}'

# Verify in database
docker exec statuspage-postgres psql -U postgres -d tenant_admin_db \
  -c "SELECT id, name, max_users FROM tenants WHERE id='TENANT_ID';"

# Check it shows in API response
curl http://localhost:8098/api/v1/tenants/TENANT_ID | python3 -m json.tool
```

### 2. Test Password Reset
```bash
# Reset password
curl -X PUT http://localhost:8098/api/v1/admin-users/ADMIN_ID/password \
  -H 'Content-Type: application/json' \
  -d '{"new_password": "NewSecurePass123!"}'

# Try logging in with new password
curl -X POST http://localhost:8098/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username": "admin", "password": "NewSecurePass123!"}'
```

### 3. Test Tab Persistence
1. Open http://localhost:8098/admin in browser
2. Click on "Tenants" tab
3. URL should change to `http://localhost:8098/admin#tenants`
4. Refresh the page (Cmd+R / Ctrl+R)
5. Should stay on Tenants tab, NOT go back to Overview

### 4. Test Login Modal
1. Open http://localhost:8098/admin
2. Login modal should appear if not logged in
3. After login, should NOT show modal on every page refresh
4. Modal should have modern styling (gradients, rounded corners)

---

## Critical Notes

1. **Do NOT run `git reset` or `git restore` commands** unless you fully understand them
2. **Always commit working code** before making major changes
3. **Test each fix individually** before moving to the next
4. **The tenant_admin_db database is SHARED** between saas-admin-service and tenant-admin-service
5. **bcrypt import already exists** in the service file - just use it
6. **MaxUsers must be a POINTER** (*int64) in the struct for nullable handling

---

## Estimated Implementation Time

- Max Users: 30-45 minutes
- Password Reset: 30-45 minutes
- Modern Login Modal: 1-2 hours (if redesigning from scratch)
- Tab Persistence: 30-45 minutes

**Total**: 3-4 hours for complete implementation and testing

---

## Additional Resources

- Database schema: See `DATABASE_ARCHITECTURE.md`
- Service architecture: See `ARCHITECTURE.md`
- SaaS Admin service docs: See `microservices/saas-admin-service/README.md`
- GORM docs: https://gorm.io/docs/
- bcrypt docs: https://pkg.go.dev/golang.org/x/crypto/bcrypt
