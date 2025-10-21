# Complete Tenant Update Implementation - ALL Fields

## Summary of Missing Fields

You correctly identified that the following fields were missing from the tenant update functionality:

### Missing Fields:
1. **max_users** (int64) - CRITICAL - User limit per tenant
2. **settings** (text/JSON) - Configuration
3. **branding** (text/JSON) - Custom theming
4. **features** (text/JSON) - Feature flags
5. **metadata** (text/JSON) - Additional data

---

## Backend Implementation Status

### ✅ Service Layer - UPDATED
**File:** `internal/services/saas_admin_service.go:1456-1467`

Added support for:
- ✅ settings (line 1456-1458)
- ✅ branding (line 1459-1461)
- ✅ features (line 1462-1464)
- ✅ metadata (line 1465-1467)

### ⚠️ Still Need to Add: max_users

The `max_users` field is an `int64`, not a string, so it needs special handling.

**Add this code after line 1467:**

```go
// Handle max_users separately since it's int64, not string
// Only update if explicitly provided and > 0
if updates.MaxUsers > 0 {
    updateData["max_users"] = updates.MaxUsers
    s.logger.Info("Max users limit updated",
        zap.String("tenant_id", tenantID.String()),
        zap.Int64("max_users", updates.MaxUsers))
}
```

### ⚠️ Handler Layer - Needs Updates
**File:** `internal/handlers/saas_admin_handler.go:1358-1407`

Currently handles:
- ✅ name
- ✅ status
- ✅ contact_email
- ✅ billing_email
- ✅ is_active
- ✅ plan_id

**Add support for missing fields after line 1389:**

```go
// Handle max_users (int64)
if maxUsers, ok := updatesMap["max_users"].(float64); ok && maxUsers > 0 {
    updates.MaxUsers = int64(maxUsers)
}

// Handle JSON fields (stored as text)
if settings, ok := updatesMap["settings"].(string); ok && settings != "" {
    updates.Settings = settings
}
if branding, ok := updatesMap["branding"].(string); ok && branding != "" {
    updates.Branding = branding
}
if features, ok := updatesMap["features"].(string); ok && features != "" {
    updates.Features = features
}
if metadata, ok := updatesMap["metadata"].(string); ok && metadata != "" {
    updates.Metadata = metadata
}
```

---

## Frontend Implementation - Complete Modal

The frontend modal needs ALL these fields added. Here's the COMPLETE implementation:

### Field Categories for UI:

#### 1. **READ-ONLY Display Section** (Immutable)
- Domain
- Subdomain
- Slug (if needed)

#### 2. **Basic Info Section** (Safe to edit)
- Tenant Name
- Status dropdown

#### 3. **Contact Information Section** (With warnings)
- Contact Email (⚠️ affects login)
- Billing Email

#### 4. **Subscription Section**
- Plan dropdown (using plan_id)
- Max Users (number input with validation)

#### 5. **Advanced Settings Section** (Collapsible/Expandable)
- Settings (textarea for JSON)
- Branding (textarea for JSON)
- Features (textarea for JSON)
- Metadata (textarea for JSON)

### Complete Frontend Modal HTML:

```javascript
function showEditTenantModal(tenant) {
    // Build plan options
    const planOptions = availablePlans.map(plan =>
        `<option value="${plan.id}" ${tenant.plan_id === plan.id ? 'selected' : ''}>${plan.name || plan.slug} - $${plan.price}/mo</option>`
    ).join('');

    const modalHTML = `
        <div id="editTenantModal" style="
            position: fixed; top: 0; left: 0; width: 100%; height: 100%;
            background: rgba(0,0,0,0.8); z-index: 10001;
            display: flex; align-items: center; justify-content: center;">
            <div style="
                background: white; padding: 30px; border-radius: 10px;
                max-width: 700px; width: 90%; max-height: 90vh; overflow-y: auto;
                box-shadow: 0 10px 40px rgba(0,0,0,0.3);">

                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                    <h3 style="margin: 0; color: #333;">
                        <i class="fas fa-edit" style="color: #667eea; margin-right: 10px;"></i>Edit Tenant
                    </h3>
                    <button onclick="closeEditTenantModal()" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #666;">&times;</button>
                </div>

                <form id="editTenantForm" style="display: flex; flex-direction: column; gap: 15px;">

                    <!-- READ-ONLY SECTION -->
                    <div style="padding: 12px; background: #f8f9fa; border-radius: 5px; border-left: 4px solid #6c757d;">
                        <div style="font-weight: 600; margin-bottom: 8px; color: #495057;">
                            <i class="fas fa-lock" style="margin-right: 5px;"></i>Immutable Fields
                        </div>
                        <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                            <span style="color: #666; font-size: 13px;">Domain:</span>
                            <span style="font-weight: 600; color: #333;">${tenant.domain || 'N/A'}</span>
                        </div>
                        <div style="display: flex; justify-content: space-between;">
                            <span style="color: #666; font-size: 13px;">Subdomain:</span>
                            <span style="font-weight: 600; color: #333;">${tenant.subdomain || 'N/A'}</span>
                        </div>
                        <div style="color: #dc3545; font-size: 11px; margin-top: 8px; padding-top: 8px; border-top: 1px solid #dee2e6;">
                            Domain/Subdomain cannot be changed. Contact platform admin for URL changes.
                        </div>
                    </div>

                    <!-- BASIC INFO SECTION -->
                    <div style="border-top: 2px solid #e9ecef; padding-top: 15px;">
                        <div style="font-weight: 600; margin-bottom: 10px; color: #495057;">Basic Information</div>

                        <div style="margin-bottom: 15px;">
                            <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Tenant Name</label>
                            <input type="text" id="editTenantName" value="${tenant.name || ''}" style="
                                width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                font-size: 14px; box-sizing: border-box;" required>
                        </div>

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
                    </div>

                    <!-- CONTACT INFO SECTION -->
                    <div style="border-top: 2px solid #e9ecef; padding-top: 15px;">
                        <div style="font-weight: 600; margin-bottom: 10px; color: #495057;">Contact Information</div>

                        <div style="margin-bottom: 15px;">
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

                        <div>
                            <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Billing Email</label>
                            <input type="email" id="editTenantBillingEmail" value="${tenant.billing_email || ''}" style="
                                width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                font-size: 14px; box-sizing: border-box;">
                        </div>
                    </div>

                    <!-- SUBSCRIPTION SECTION -->
                    <div style="border-top: 2px solid #e9ecef; padding-top: 15px;">
                        <div style="font-weight: 600; margin-bottom: 10px; color: #495057;">Subscription & Limits</div>

                        <div style="margin-bottom: 15px;">
                            <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">Plan</label>
                            <select id="editTenantPlan" style="
                                width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                font-size: 14px; box-sizing: border-box;" required>
                                ${planOptions}
                            </select>
                        </div>

                        <div>
                            <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
                                Max Users Limit
                            </label>
                            <input type="number" id="editTenantMaxUsers" value="${tenant.max_users || ''}"
                                min="1" max="10000" style="
                                width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                font-size: 14px; box-sizing: border-box;">
                            <div style="color: #6c757d; font-size: 11px; margin-top: 3px;">
                                Maximum number of users allowed for this tenant (leave empty for unlimited)
                            </div>
                        </div>
                    </div>

                    <!-- ADVANCED SETTINGS (COLLAPSIBLE) -->
                    <div style="border-top: 2px solid #e9ecef; padding-top: 15px;">
                        <div style="display: flex; justify-content: space-between; align-items: center; cursor: pointer;"
                            onclick="document.getElementById('advancedSettings').style.display =
                            document.getElementById('advancedSettings').style.display === 'none' ? 'block' : 'none'">
                            <div style="font-weight: 600; color: #495057;">
                                <i class="fas fa-cog" style="margin-right: 5px;"></i>Advanced Settings (JSON)
                            </div>
                            <i class="fas fa-chevron-down" style="color: #6c757d;"></i>
                        </div>

                        <div id="advancedSettings" style="display: none; margin-top: 15px;">
                            <div style="margin-bottom: 15px;">
                                <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
                                    Settings (JSON)
                                </label>
                                <textarea id="editTenantSettings" rows="3" style="
                                    width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                    font-size: 12px; font-family: monospace; box-sizing: border-box;"
                                    placeholder='{"key": "value"}'>${tenant.settings || ''}</textarea>
                            </div>

                            <div style="margin-bottom: 15px;">
                                <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
                                    Branding (JSON)
                                </label>
                                <textarea id="editTenantBranding" rows="3" style="
                                    width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                    font-size: 12px; font-family: monospace; box-sizing: border-box;"
                                    placeholder='{"logo": "url", "colors": {}}'>${tenant.branding || ''}</textarea>
                            </div>

                            <div style="margin-bottom: 15px;">
                                <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
                                    Features (JSON)
                                </label>
                                <textarea id="editTenantFeatures" rows="3" style="
                                    width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                    font-size: 12px; font-family: monospace; box-sizing: border-box;"
                                    placeholder='{"feature1": true, "feature2": false}'>${tenant.features || ''}</textarea>
                            </div>

                            <div>
                                <label style="display: block; margin-bottom: 5px; color: #555; font-weight: 600;">
                                    Metadata (JSON)
                                </label>
                                <textarea id="editTenantMetadata" rows="3" style="
                                    width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 5px;
                                    font-size: 12px; font-family: monospace; box-sizing: border-box;"
                                    placeholder='{"custom": "data"}'>${tenant.metadata || ''}</textarea>
                            </div>
                        </div>
                    </div>

                    <!-- ACTION BUTTONS -->
                    <div style="display: flex; gap: 10px; margin-top: 20px; padding-top: 15px; border-top: 2px solid #e9ecef;">
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
            plan_id: document.getElementById('editTenantPlan').value,
            status: document.getElementById('editTenantStatus').value,
            is_active: true
        };

        // Add max_users if provided
        const maxUsers = document.getElementById('editTenantMaxUsers').value;
        if (maxUsers && parseInt(maxUsers) > 0) {
            updateData.max_users = parseInt(maxUsers);
        }

        // Add JSON fields if provided
        const settings = document.getElementById('editTenantSettings').value.trim();
        if (settings) updateData.settings = settings;

        const branding = document.getElementById('editTenantBranding').value.trim();
        if (branding) updateData.branding = branding;

        const features = document.getElementById('editTenantFeatures').value.trim();
        if (features) updateData.features = features;

        const metadata = document.getElementById('editTenantMetadata').value.trim();
        if (metadata) updateData.metadata = metadata;

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
```

---

## Complete Field Summary

### ✅ Currently Editable (After all updates):
1. name ✅
2. status ✅
3. contact_email ✅ (with warning)
4. billing_email ✅
5. plan_id ✅
6. is_active ✅
7. settings ✅ (NEW)
8. branding ✅ (NEW)
9. features ✅ (NEW)
10. metadata ✅ (NEW)
11. max_users ⚠️ (NEEDS HANDLER UPDATE)

### 🔒 Immutable (Blocked):
1. id
2. slug
3. domain
4. subdomain
5. created_at
6. updated_at
7. deleted_at

---

## Next Steps

1. Add max_users support to service layer (code provided above)
2. Add all missing fields to handler layer (code provided above)
3. Replace frontend modal completely (code provided above)
4. Test all fields update correctly
