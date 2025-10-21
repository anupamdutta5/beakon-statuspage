# Tenant Update Functionality - Complete Implementation Summary

## Overview
This document summarizes all changes made to fix the tenant update functionality with proper security controls.

---

## Backend Changes ✅ COMPLETED

### 1. Service Layer: `internal/services/saas_admin_service.go:1418-1484`

**Security Policy Implemented:**
- ✅ **BLOCKS domain/subdomain changes completely** - Returns error if attempted
- ✅ Supports plan_id updates (UUID format)
- ✅ Supports contact_email updates (with security warning logged)
- ✅ Supports billing_email updates
- ✅ Supports status updates
- ✅ Supports name updates
- ✅ Supports is_active boolean updates

**Code snippet:**
```go
// IMMUTABLE FIELDS - Domain/Subdomain changes are blocked
if updates.Domain != "" || updates.Subdomain != "" {
    s.logger.Warn("Attempted to update immutable fields", ...)
    return fmt.Errorf("domain and subdomain cannot be changed after tenant creation")
}
```

### 2. Handler Layer: `internal/handlers/saas_admin_handler.go:1348-1407`

**Field Handling:**
- ✅ Accepts plan_id as string, converts to UUID
- ✅ Validates UUID format
- ✅ Accepts all safe-to-edit fields
- ✅ Returns proper error messages (400 status with error text)

---

## Frontend Changes NEEDED

### Current Location: `web/templates/partials/scripts.html`

The `showEditTenantModal()` function needs complete replacement.

### Required Changes:

#### 1. **Remove Editable Domain/Subdomain Fields** ❌
Currently lines 1288-1298 allow editing domain. This MUST be removed.

Replace with **read-only display**:
```html
<div style="padding: 10px; background: #f8f9fa; border-radius: 5px; border-left: 3px solid #6c757d;">
    <div style="display: flex; justify-content: space-between;">
        <span style="color: #666; font-size: 12px;">Domain (Read-only):</span>
        <span style="font-weight: 600;">${tenant.domain || 'N/A'}</span>
    </div>
    <div style="display: flex; justify-content: space-between; margin-top: 5px;">
        <span style="color: #666; font-size: 12px;">Subdomain (Read-only):</span>
        <span style="font-weight: 600;">${tenant.subdomain || 'N/A'}</span>
    </div>
    <div style="color: #dc3545; font-size: 11px; margin-top: 8px;">
        <i class="fas fa-lock" style="margin-right: 5px;"></i>
        Domain/Subdomain cannot be changed. Contact platform admin for URL changes.
    </div>
</div>
```

#### 2. **Fix Plan Selection** - Change from Plan Names to Plan IDs ❌

Currently (lines 1299-1313) uses plan names. Must change to plan_id UUIDs.

**Add global plans array:**
```javascript
let availablePlans = [];

// Fetch plans on page load
function loadPlans() {
    fetch('/api/v1/plans')
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                availablePlans = data.data;
            }
        })
        .catch(error => console.error('Error loading plans:', error));
}

// Call on page load
document.addEventListener('DOMContentLoaded', function() {
    loadPlans();
});
```

**Update plan dropdown in modal:**
```javascript
// Find tenant's current plan_id from the API
let tenantPlanId = ''; // Will be fetched from API
const planOptions = availablePlans.map(plan =>
    `<option value="${plan.id}" ${plan.id === tenantPlanId ? 'selected' : ''}>${plan.name} - $${plan.price}/mo</option>`
).join('');

// In modal HTML:
<select id="editTenantPlan">
    ${planOptions}
</select>
```

#### 3. **Add Contact Email Field with Warning** ❌

```html
<div>
    <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
        <i class="fas fa-exclamation-triangle" style="color: #ff9800;"></i>
        Contact Email (Used for Login)
    </label>
    <input type="email" id="editTenantContactEmail" value="${tenant.contact_email || ''}" style="
        width: 100%;
        padding: 10px;
        border: 1px solid #ff9800;
        border-radius: 5px;
        font-size: 14px;
        box-sizing: border-box;
    " required>
    <div style="color: #ff9800; font-size: 11px; margin-top: 3px;">
        ⚠️ This email is used for tenant admin login. Changing it affects authentication.
    </div>
</div>
```

#### 4. **Add Billing Email Field** ❌

```html
<div>
    <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Billing Email</label>
    <input type="email" id="editTenantBillingEmail" value="${tenant.billing_email || ''}" style="
        width: 100%;
        padding: 10px;
        border: 1px solid #ddd;
        border-radius: 5px;
        font-size: 14px;
        box-sizing: border-box;
    ">
</div>
```

#### 5. **Update Fetch Body** (lines 1375-1381) ❌

**Current (WRONG):**
```javascript
body: JSON.stringify({
    name: newName,
    domain: newDomain,        // REMOVE - immutable
    subdomain: newDomain,     // REMOVE - immutable
    plan: newPlan,            // WRONG - should be plan_id
    status: newStatus
})
```

**New (CORRECT):**
```javascript
body: JSON.stringify({
    name: document.getElementById('editTenantName').value,
    contact_email: document.getElementById('editTenantContactEmail').value,
    billing_email: document.getElementById('editTenantBillingEmail').value,
    plan_id: document.getElementById('editTenantPlan').value,  // UUID
    status: document.getElementById('editTenantStatus').value,
    is_active: true  // Or add a checkbox for this
})
```

---

## Complete Replacement Code for showEditTenantModal()

Save this to a file and use it to replace lines 1241-1400 in scripts.html:

```javascript
function showEditTenantModal(tenant) {
    // Build plan options dynamically
    const planOptions = availablePlans.map(plan =>
        `<option value="${plan.id}" ${plan.plan_id === plan.id ? 'selected' : ''}>${plan.name || plan.slug} - $${plan.price}/mo</option>`
    ).join('');

    const modalHTML = `
        <div id="editTenantModal" style="
            position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.8); z-index: 10001;
            display: flex; align-items: center; justify-content: center;">
            <div style="
                background: white; padding: 30px; border-radius: 10px;
                max-width: 600px; width: 90%; max-height: 90vh; overflow-y: auto;
                box-shadow: 0 10px 40px rgba(0,0,0,0.3);">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                    <h3 style="margin: 0; color: #333;">
                        <i class="fas fa-edit" style="color: #667eea; margin-right: 10px;"></i>Edit Tenant
                    </h3>
                    <button onclick="closeEditTenantModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>

                <form id="editTenantForm" style="display: flex; flex-direction: column; gap: 15px;">

                    <!-- READ-ONLY DOMAIN/SUBDOMAIN INFO -->
                    <div style="padding: 12px; background: #f8f9fa; border-radius: 5px; border-left: 4px solid #6c757d;">
                        <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                            <span style="color: #666; font-size: 13px; font-weight: 600;">Domain:</span>
                            <span style="font-weight: 600; color: #333;">${tenant.domain || 'N/A'}</span>
                        </div>
                        <div style="display: flex; justify-content: space-between;">
                            <span style="color: #666; font-size: 13px; font-weight: 600;">Subdomain:</span>
                            <span style="font-weight: 600; color: #333;">${tenant.subdomain || 'N/A'}</span>
                        </div>
                        <div style="color: #dc3545; font-size: 11px; margin-top: 10px; padding-top: 10px; border-top: 1px solid #dee2e6;">
                            <i class="fas fa-lock" style="margin-right: 5px;"></i>
                            Domain/Subdomain are locked and cannot be changed via UI. Contact platform administrator for URL changes.
                        </div>
                    </div>

                    <!-- TENANT NAME -->
                    <div>
                        <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Tenant Name</label>
                        <input type="text" id="editTenantName" value="${tenant.name || ''}" style="
                            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                            font-size: 14px; box-sizing: border-box;" required>
                    </div>

                    <!-- CONTACT EMAIL (WITH WARNING) -->
                    <div>
                        <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
                            <i class="fas fa-exclamation-triangle" style="color: #ff9800; margin-right: 5px;"></i>
                            Contact Email (Login Email)
                        </label>
                        <input type="email" id="editTenantContactEmail" value="${tenant.contact_email || ''}" style="
                            width: 100%; padding: 10px; border: 2px solid #ff9800; border-radius: 5px;
                            font-size: 14px; box-sizing: border-box;" required>
                        <div style="color: #ff9800; font-size: 11px; margin-top: 5px; font-weight: 500;">
                            ⚠️ This email is used for tenant admin login. Changing it affects authentication.
                        </div>
                    </div>

                    <!-- BILLING EMAIL -->
                    <div>
                        <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Billing Email</label>
                        <input type="email" id="editTenantBillingEmail" value="${tenant.billing_email || ''}" style="
                            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                            font-size: 14px; box-sizing: border-box;">
                    </div>

                    <!-- PLAN SELECTION (WITH PLAN_ID) -->
                    <div>
                        <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Subscription Plan</label>
                        <select id="editTenantPlan" style="
                            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                            font-size: 14px; box-sizing: border-box;" required>
                            ${planOptions}
                        </select>
                    </div>

                    <!-- STATUS -->
                    <div>
                        <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Status</label>
                        <select id="editTenantStatus" style="
                            width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                            font-size: 14px; box-sizing: border-box;" required>
                            <option value="active" ${tenant.status === 'active' ? 'selected' : ''}>Active</option>
                            <option value="trial" ${tenant.status === 'trial' ? 'selected' : ''}>Trial</option>
                            <option value="suspended" ${tenant.status === 'suspended' ? 'selected' : ''}>Suspended</option>
                            <option value="inactive" ${tenant.status === 'inactive' ? 'selected' : ''}>Inactive</option>
                        </select>
                    </div>

                    <div style="display: flex; gap: 10px; margin-top: 10px;">
                        <button type="button" onclick="closeEditTenantModal()" style="
                            flex: 1; padding: 12px; border: 1px solid #ddd; background: white;
                            color: #666; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 600;">
                            Cancel
                        </button>
                        <button type="submit" style="
                            flex: 1; padding: 12px; border: none;
                            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                            color: white; border-radius: 6px; cursor: pointer; font-size: 14px; font-weight: 600;">
                            Update Tenant
                        </button>
                    </div>
                </form>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHTML);

    // Form submit handler
    document.getElementById('editTenantForm').addEventListener('submit', function(e) {
        e.preventDefault();

        const updateData = {
            name: document.getElementById('editTenantName').value,
            contact_email: document.getElementById('editTenantContactEmail').value,
            billing_email: document.getElementById('editTenantBillingEmail').value,
            plan_id: document.getElementById('editTenantPlan').value,  // UUID!
            status: document.getElementById('editTenantStatus').value,
            is_active: true  // Default to true, or add checkbox
        };

        // Update tenant via API
        fetch(`/api/v1/tenants/${tenant.id}`, {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(updateData)
        })
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                showNotification('Tenant updated successfully!', 'success', 4000);
                closeEditTenantModal();
                closeTenantModal();
                loadDashboardData();
                loadTenantsPage();
            } else {
                showNotification('Error: ' + (data.error || data.message), 'error', 6000);
            }
        })
        .catch(error => {
            console.error('Error updating tenant:', error);
            showNotification('Error updating tenant: ' + error.message, 'error', 6000);
        });
    });
}

// Also add this function at page load
let availablePlans = [];

function loadPlans() {
    fetch('/api/v1/plans')
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                availablePlans = data.data;
            }
        })
        .catch(error => console.error('Error loading plans:', error));
}

// Initialize plans when page loads
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', loadPlans);
} else {
    loadPlans();
}
```

---

## Testing Checklist

After implementing the frontend changes:

1. ✅ **Test domain/subdomain block:**
   - Try to edit tenant
   - Domain/subdomain should be READ-ONLY gray boxes
   - Should show warning message about contacting admin

2. ✅ **Test plan update:**
   - Select different plan from dropdown
   - Save
   - Verify plan changes in database
   - Check logs show plan_id UUID

3. ✅ **Test contact_email update:**
   - Change contact email
   - Should see orange warning
   - Verify database updated
   - Check backend logs show warning

4. ✅ **Test billing_email update:**
   - Change billing email
   - Should save without warnings

5. ✅ **Test name/status update:**
   - Change name and status
   - Should update immediately in UI after save

---

## Database Verification Commands

```bash
# Check plan IDs
docker exec statuspage-postgres psql -U postgres -d saas_admin -c "SELECT id, name, slug, price FROM saas_plans;"

# Verify tenant update
docker exec statuspage-postgres psql -U postgres -d tenant_admin_db -c "SELECT id, name, contact_email, billing_email, plan_id, status FROM tenants WHERE id='<tenant-id>';"
```

---

## Summary

**Backend:** ✅ Complete - Domain/subdomain blocked, all safe fields supported
**Frontend:** ❌ Needs implementation - Use code above to replace modal function
**Testing:** ⏳ Pending frontend implementation

The system is now secure with Option B implemented: **Domain/subdomain changes completely blocked via UI, requiring manual database updates by platform administrators.**
