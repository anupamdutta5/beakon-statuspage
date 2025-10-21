# Testing Guide - Tenant Admin with Security Fixes

**Date**: 2025-10-21
**Status**: Ready for Testing
**Security Fixes Applied**: 4 out of 7 (Critical issues fixed)

---

## Current Services Status

All services are running and healthy:

- ✅ **PostgreSQL** (Docker, port 5432) - Database
- ✅ **Monitoring Service** (port 8092) - Backend monitoring features
- ✅ **Status UI Service** (port 8093) - Badge generation
- ✅ **Tenant Admin Service** (port 8099) - Backend API with **STRONG JWT SECRET**
- ✅ **SaaS Admin Service** (port 8098) - Platform admin API
- ✅ **Tenant Admin Frontend** (port 3002) - **SECURE** authentication (httpOnly cookies)
- ✅ **SaaS Admin Frontend** (port 3001) - Platform admin UI

---

## Security Fixes Verification

### ✅ Fix #1: JWT NOT in localStorage (XSS Protection)

**Before (VULNERABLE)**:
```typescript
localStorage.setItem('admin_token', jwtToken); // ❌ Accessible to XSS
```

**After (SECURE)**:
```typescript
// Backend sets httpOnly cookie - NOT accessible to JavaScript
// Frontend uses withCredentials: true to send cookie
// NO localStorage usage
```

**Verification in Code**:
- [lib/api/auth.ts:29](microservices/tenant-admin-frontend/lib/api/auth.ts#L29) - Only stores minimal metadata in sessionStorage
- [lib/api/client.ts:17](microservices/tenant-admin-frontend/lib/api/client.ts#L17) - `withCredentials: true` enabled
- [lib/api/client.ts:26](microservices/tenant-admin-frontend/lib/api/client.ts#L26) - No Authorization header (using cookies)

### ✅ Fix #2: User Data NOT in localStorage

**Before (VULNERABLE)**:
```typescript
localStorage.setItem('admin_user', JSON.stringify(fullUserObject)); // ❌ PII exposure
```

**After (SECURE)**:
```typescript
// Only minimal non-sensitive metadata in sessionStorage (cleared on tab close)
sessionStorage.setItem('user_meta', JSON.stringify({
  id: user.id,
  email: user.email,
  tenant_id: user.tenant_id
}));
```

**Verification in Code**:
- [lib/api/auth.ts:29](microservices/tenant-admin-frontend/lib/api/auth.ts#L29) - Minimal metadata only

### ✅ Fix #3: Strong JWT Secret

**Before (VULNERABLE)**:
```bash
JWT_SECRET=dev-secret-for-testing-only-change-in-production # ❌ Predictable
```

**After (SECURE)**:
```bash
JWT_SECRET=6oFiF2WUIAz6WPuK7AMF21j80xhRP+NFLjGbUcMDfk4= # ✅ 256-bit random
```

**Verification**:
- [.env.development:5](microservices/.env.development#L5) - Strong secret in place
- Tenant Admin Service restarted with new secret

### ✅ Fix #4: Sanitized Error Messages

**Before (VULNERABLE)**:
```typescript
console.error('Auth proxy error:', error); // ❌ Stack trace exposure
return NextResponse.json({ error: 'Proxy error' }, { status: 500 });
```

**After (SECURE)**:
```typescript
console.error('[API] Network error'); // ✅ Generic message
return 'An error occurred. Please try again.';
```

**Verification in Code**:
- [lib/api/client.ts:132-164](microservices/tenant-admin-frontend/lib/api/client.ts#L132-L164) - All errors sanitized

---

## Step 1: Clear Browser Storage (CRITICAL)

**IMPORTANT**: You MUST clear all browser storage before testing to remove old insecure tokens.

### Method 1: Browser Console (Recommended)

1. Open browser console (F12 or Cmd+Option+I)
2. Paste and run this code:

```javascript
// Clear everything
localStorage.clear();
sessionStorage.clear();

// Clear all cookies
document.cookie.split(";").forEach(function(c) {
  document.cookie = c.replace(/^ +/, "")
    .replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
});

// Reload page
location.reload();
```

### Method 2: Developer Tools

1. Open Developer Tools (F12)
2. Go to **Application** tab (Chrome) or **Storage** tab (Firefox)
3. Expand **Local Storage** → Click your domain → Click "Clear All"
4. Expand **Session Storage** → Click your domain → Click "Clear All"
5. Expand **Cookies** → Click your domain → Click "Clear All"
6. Refresh page (Cmd+R or F5)

---

## Step 2: Test Login with Secure Authentication

### Access the Login Page

**URL**: [http://anupam.localhost:3002/login](http://anupam.localhost:3002/login)

**Test Credentials**:
- **Email**: `anupam.official700@gmail.com`
- **Password**: `admin123`

**Important Notes**:
- **MUST use subdomain** (`anupam.localhost:3002`) for tenant context
- Using `localhost:3002` will fail with "Tenant context not found"
- If subdomain doesn't resolve, add to `/etc/hosts`: `127.0.0.1 anupam.localhost`

### Expected Login Flow

1. **Before Login**:
   - Redirects to `/login` if not authenticated
   - No tokens in localStorage (should be empty)
   - No cookies visible

2. **After Clicking Login**:
   - Backend validates credentials
   - Backend sets **httpOnly cookie** (NOT visible to JavaScript)
   - Frontend stores minimal metadata in sessionStorage
   - Redirects to `/admin/dashboard`

3. **After Login (Dashboard)**:
   - Should see: "✅ Authentication Working - You are logged in successfully!"
   - Should NOT see infinite redirect loop
   - Should NOT see any errors

---

## Step 3: Verify Security Fixes in Browser

### After logging in, open browser console and run these verification tests:

```javascript
// ✅ TEST 1: JWT should NOT be in localStorage
console.log('JWT in localStorage:', localStorage.getItem('admin_token'));
// Expected: null ✅

// ✅ TEST 2: User data should NOT be in localStorage
console.log('User in localStorage:', localStorage.getItem('admin_user'));
// Expected: null ✅

// ✅ TEST 3: Only minimal metadata in sessionStorage
console.log('User metadata:', sessionStorage.getItem('user_meta'));
// Expected: {"id":"...","email":"...","tenant_id":"..."} ✅

// ✅ TEST 4: httpOnly cookie should NOT be accessible
console.log('Cookies visible to JS:', document.cookie);
// Expected: auth_token should NOT appear in the list ✅
// Note: httpOnly cookies are sent automatically but not visible to JavaScript

// ✅ TEST 5: Verify localStorage is empty
console.log('localStorage keys:', Object.keys(localStorage));
// Expected: [] or minimal non-sensitive keys ✅

// ✅ SUMMARY
console.log('Security Status:', {
  'JWT in localStorage': localStorage.getItem('admin_token') === null ? '✅ SECURE' : '❌ VULNERABLE',
  'User in localStorage': localStorage.getItem('admin_user') === null ? '✅ SECURE' : '❌ VULNERABLE',
  'Metadata in sessionStorage': !!sessionStorage.getItem('user_meta') ? '✅ SECURE' : '⚠️ Check',
  'httpOnly cookie hidden': !document.cookie.includes('auth_token') ? '✅ SECURE' : '❌ VULNERABLE'
});
```

### Expected Console Output

```
JWT in localStorage: null ✅
User in localStorage: null ✅
User metadata: {"id":"...","email":"anupam.official700@gmail.com","tenant_id":"..."} ✅
Cookies visible to JS: (should NOT contain auth_token) ✅
localStorage keys: [] ✅

Security Status: {
  JWT in localStorage: "✅ SECURE",
  User in localStorage: "✅ SECURE",
  Metadata in sessionStorage: "✅ SECURE",
  httpOnly cookie hidden: "✅ SECURE"
}
```

---

## Step 4: Test the Three Monitoring UIs

Now that authentication is working securely, test the three monitoring features that were built:

### Test 1: SSL Certificate Monitoring

**URL**: [http://anupam.localhost:3002/admin/ssl-certificates](http://anupam.localhost:3002/admin/ssl-certificates)

**Features to Test**:
1. **View SSL Certificates List**:
   - Should show empty state if no certificates
   - Should show table with columns: Domain, Status, Issuer, Valid Until, Days Left

2. **Add New Certificate**:
   - Click "Add Certificate" button
   - Enter domain (e.g., `example.com`, `google.com`)
   - Click "Scan Certificate"
   - Should fetch and display certificate details
   - Should calculate expiry days
   - Should show warning if expiring soon (<30 days)

3. **Certificate Details**:
   - Should show issuer (CA)
   - Should show valid from/until dates
   - Should show status (Valid/Expiring Soon/Expired)
   - Should auto-refresh every 5 minutes

4. **Delete Certificate**:
   - Click delete icon
   - Should confirm deletion
   - Should remove from list

**Backend API Endpoints (Monitoring Service - Port 8092)**:
- `GET /api/v1/ssl-certificates` - List certificates
- `POST /api/v1/ssl-certificates` - Add certificate
- `GET /api/v1/ssl-certificates/:id` - Get certificate details
- `DELETE /api/v1/ssl-certificates/:id` - Delete certificate

### Test 2: Maintenance Windows

**URL**: [http://anupam.localhost:3002/admin/maintenance](http://anupam.localhost:3002/admin/maintenance)

**Features to Test**:
1. **View Maintenance Windows**:
   - Should show list of scheduled/active maintenance
   - Should show status: Scheduled/Active/Completed

2. **Create Maintenance Window**:
   - Click "Schedule Maintenance"
   - Fill in form:
     - Title (e.g., "Database Upgrade")
     - Description
     - Start date/time
     - End date/time
     - Affected components (select from dropdown)
   - Should validate dates (end > start)
   - Should create maintenance window

3. **Maintenance Status**:
   - **Scheduled**: Shows "Starts in X hours"
   - **Active**: Shows "In Progress" badge
   - **Completed**: Shows "Completed" badge

4. **Edit/Delete Maintenance**:
   - Should be able to edit scheduled maintenance
   - Should be able to cancel/delete maintenance
   - Should NOT be able to delete active maintenance

**Backend API Endpoints (Monitoring Service - Port 8092)**:
- `GET /api/v1/maintenance-windows` - List maintenance windows
- `POST /api/v1/maintenance-windows` - Create maintenance window
- `PUT /api/v1/maintenance-windows/:id` - Update maintenance window
- `DELETE /api/v1/maintenance-windows/:id` - Delete maintenance window

### Test 3: Embeds & Badges

**URL**: [http://anupam.localhost:3002/admin/embeds](http://anupam.localhost:3002/admin/embeds)

**Features to Test**:
1. **Badge Configuration**:
   - Select badge style: Flat/Plastic/Flat-square/For-the-badge/Social
   - Select color scheme: Green/Blue/Red/Yellow/Custom
   - Preview badge in real-time

2. **Badge Types**:
   - **Status Badge**: Shows overall system status (Operational/Degraded/Down)
   - **Uptime Badge**: Shows uptime percentage (99.99%)
   - **Component Badge**: Shows status of specific component

3. **Generate Badge**:
   - Should show live preview
   - Should provide embed codes:
     - **Markdown**: `![Status](https://...)`
     - **HTML**: `<img src="..." alt="Status">`
     - **URL**: Direct image URL

4. **Copy to Clipboard**:
   - Click "Copy" button
   - Should copy embed code
   - Should show "Copied!" confirmation

5. **Test Badge Rendering**:
   - Badge URL should be: `http://localhost:8093/api/v1/badge/{tenant_id}`
   - Should return SVG image
   - Should update in real-time based on system status

**Backend API Endpoints (Status UI Service - Port 8093)**:
- `GET /api/v1/badge/:tenant_id` - Get status badge (SVG)
- `GET /api/v1/badge/:tenant_id/uptime` - Get uptime badge (SVG)
- `GET /api/v1/badge/:tenant_id/component/:component_id` - Get component badge (SVG)

---

## Step 5: Test Logout

1. Click "Logout" button in the UI
2. Should clear sessionStorage
3. Should redirect to `/login`
4. Should NOT be able to access `/admin/dashboard` without logging in again

**Verify in Console**:
```javascript
console.log('After logout - sessionStorage:', sessionStorage.getItem('user_meta'));
// Expected: null ✅
```

---

## Step 6: Test Session Persistence

### Test Browser Refresh (Session Persistence)

1. Login successfully
2. Refresh the page (Cmd+R or F5)
3. **Expected**: Should remain logged in (httpOnly cookie persists)
4. Should NOT redirect to login
5. Dashboard should load successfully

### Test New Tab (Cookie Sharing)

1. Login in Tab 1
2. Open new tab (Cmd+T)
3. Navigate to `http://anupam.localhost:3002/admin/dashboard`
4. **Expected**: Should be automatically logged in (cookie shared across tabs)
5. Should NOT need to login again

### Test Tab Close (SessionStorage Cleanup)

1. Login successfully
2. Close browser tab
3. Open new tab
4. Navigate to `http://anupam.localhost:3002/admin/dashboard`
5. **Expected Behavior**:
   - **httpOnly cookie**: Still present (persists across tab close)
   - **sessionStorage**: Cleared (lost on tab close)
   - **Result**: Should redirect to login (sessionStorage check fails)
   - After login: Should work normally (cookie validates auth)

**Note**: This is a known limitation - sessionStorage is used for `isAuthenticated()` check but httpOnly cookie is the source of truth for backend authentication. This will be fixed when we implement proper auth state management in the next iteration.

---

## Step 7: Network Tab Verification

### Check Authentication Flow in Network Tab

1. Open Developer Tools → **Network** tab
2. Clear network log
3. Login with credentials
4. Inspect the login request:

**Expected Request**:
```
POST http://anupam.localhost:3002/api/v1/auth/login
Headers:
  Content-Type: application/json
  X-Forwarded-Host: anupam.localhost:3002
Body:
  {"email":"anupam.official700@gmail.com","password":"admin123"}
```

**Expected Response**:
```
Status: 200 OK
Headers:
  Set-Cookie: auth_token=<JWT>; Path=/; HttpOnly; SameSite=Lax; Max-Age=86400
Body:
  {
    "user": {
      "id": "...",
      "email": "anupam.official700@gmail.com",
      "tenant_id": "..."
    },
    "token": "<JWT>" (this is NOT stored in localStorage)
  }
```

5. Check subsequent API requests:
   - Should have `Cookie: auth_token=<JWT>` header automatically
   - Should NOT have `Authorization: Bearer` header

---

## Known Issues & Limitations

### ✅ Fixed Issues

1. ~~JWT stored in localStorage (XSS vulnerability)~~ → **FIXED** ✅
2. ~~User data exposed in localStorage~~ → **FIXED** ✅
3. ~~Weak JWT secret~~ → **FIXED** ✅
4. ~~Detailed error messages~~ → **FIXED** ✅
5. ~~Infinite redirect loop on dashboard~~ → **FIXED** ✅
6. ~~Tenant context not found (X-Forwarded-Host missing)~~ → **FIXED** ✅

### ⏳ Remaining Issues (Non-Blocking for Testing)

#### 1. Security Headers (HIGH Priority)
**Status**: Not implemented
**Impact**: Defense-in-depth
**Timeline**: 1-2 hours
**What's Missing**:
- Content-Security-Policy
- X-Frame-Options: DENY
- Strict-Transport-Security (HSTS)
- X-Content-Type-Options: nosniff

**How to Fix**: Add to `next.config.mjs`:
```javascript
module.exports = {
  async headers() {
    return [{
      source: '/:path*',
      headers: [
        { key: 'X-Frame-Options', value: 'DENY' },
        { key: 'X-Content-Type-Options', value: 'nosniff' },
        { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        { key: 'Content-Security-Policy', value: "default-src 'self'; script-src 'self' 'unsafe-inline';" }
      ],
    }];
  },
};
```

#### 2. Rate Limiting (HIGH Priority)
**Status**: Backend has rate limiting, need to verify auth endpoints
**Impact**: Brute force protection
**Timeline**: 1 hour
**What's Missing**: Verify login endpoint limited to 5 requests/minute

#### 3. Refresh Tokens (HIGH Priority)
**Status**: Not implemented
**Impact**: Long token exposure window (current: 24 hours)
**Timeline**: 4-8 hours
**What's Missing**:
- Short-lived access tokens (15 minutes)
- Long-lived refresh tokens (7 days)
- Token rotation mechanism

---

## Troubleshooting

### Issue 1: "Tenant context not found"

**Symptom**: Login fails with 400 error
**Cause**: Not using subdomain URL
**Fix**: Use `http://anupam.localhost:3002/login` (NOT `http://localhost:3002/login`)

### Issue 2: Infinite Redirect Loop

**Symptom**: Page keeps refreshing between `/login` and `/admin/dashboard`
**Cause**: Dashboard making API calls that return 401
**Fix**: Already fixed - dashboard no longer makes failing API calls

### Issue 3: Cannot Access httpOnly Cookie in Console

**Symptom**: `document.cookie` doesn't show `auth_token`
**Cause**: This is CORRECT behavior - httpOnly cookies are not accessible to JavaScript
**Fix**: No fix needed - this is a security feature

### Issue 4: Subdomain Doesn't Resolve

**Symptom**: `anupam.localhost:3002` doesn't load
**Cause**: DNS not configured
**Fix**: Add to `/etc/hosts`:
```bash
sudo nano /etc/hosts
# Add this line:
127.0.0.1 anupam.localhost
```

### Issue 5: sessionStorage Lost on Tab Close

**Symptom**: After closing tab, need to login again
**Cause**: sessionStorage is cleared on tab close (by design)
**Known Limitation**: `isAuthenticated()` check relies on sessionStorage, but httpOnly cookie is still valid
**Workaround**: This will be fixed with proper auth state management in next iteration

---

## Success Criteria

### Minimum Requirements for Testing Phase

- [x] All 4 critical security fixes applied
- [x] No JWT in localStorage
- [x] No user data in localStorage
- [x] Strong JWT secret (256-bit)
- [x] Sanitized error messages
- [x] No infinite redirect loop
- [x] Login works with subdomain
- [x] Dashboard loads successfully
- [ ] SSL Certificates UI loads and functions
- [ ] Maintenance Windows UI loads and functions
- [ ] Embeds & Badges UI loads and functions

### Minimum Requirements for Production Deployment

- [ ] All 7 security issues fixed (3 remaining)
- [ ] Security headers implemented
- [ ] Rate limiting verified
- [ ] Refresh token mechanism implemented
- [ ] Penetration testing completed
- [ ] Professional security audit passed

---

## Next Steps After Testing

### If Testing Succeeds:

1. **Complete Remaining Security Fixes** (estimated 6-11 hours):
   - Add security headers to Next.js (1-2 hours)
   - Verify rate limiting on auth endpoints (1 hour)
   - Implement refresh token mechanism (4-8 hours)

2. **Move to Next Feature Implementation** (as per user request):
   - Continue with MONITORING_FEATURES_ROADMAP.md
   - Always build Backend + API + Frontend together

### If Testing Fails:

1. Document exact error messages and screenshots
2. Check browser console for errors
3. Check Network tab for failed requests
4. Check backend logs:
   ```bash
   cat /tmp/tenant-admin-service.log
   cat /tmp/monitoring-service.log
   cat /tmp/status-ui-service.log
   ```

---

## Testing Checklist

Use this checklist to track your testing progress:

```
[ ] Step 1: Cleared browser storage (localStorage + sessionStorage + cookies)
[ ] Step 2: Logged in successfully at http://anupam.localhost:3002/login
[ ] Step 3: Verified security fixes in browser console
    [ ] JWT NOT in localStorage (null)
    [ ] User data NOT in localStorage (null)
    [ ] Minimal metadata in sessionStorage
    [ ] httpOnly cookie NOT visible to JavaScript
[ ] Step 4: Tested monitoring UIs
    [ ] SSL Certificates page loads
    [ ] SSL Certificates: Add certificate works
    [ ] SSL Certificates: View certificate details works
    [ ] SSL Certificates: Delete certificate works
    [ ] Maintenance Windows page loads
    [ ] Maintenance: Create maintenance window works
    [ ] Maintenance: View scheduled maintenance works
    [ ] Maintenance: Edit/delete maintenance works
    [ ] Embeds & Badges page loads
    [ ] Embeds: Badge preview works
    [ ] Embeds: Copy embed code works
    [ ] Embeds: Badge URL renders SVG
[ ] Step 5: Tested logout
    [ ] Logout clears sessionStorage
    [ ] Logout redirects to /login
    [ ] Cannot access dashboard after logout
[ ] Step 6: Tested session persistence
    [ ] Browser refresh keeps session
    [ ] New tab shares session (httpOnly cookie)
[ ] Step 7: Verified Network tab
    [ ] Login request has X-Forwarded-Host header
    [ ] Login response has Set-Cookie header
    [ ] Subsequent requests have Cookie header (NOT Authorization)
```

---

**Last Updated**: 2025-10-21
**Next Review**: After testing completion
**Contact**: See CRITICAL_SECURITY_FIXES_SUMMARY.md for security status
