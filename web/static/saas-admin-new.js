// SaaS Admin Dashboard JavaScript - Simplified and Reliable Version

class SaaSAdminDashboard {
    constructor() {
        this.currentSection = 'overview';
        this.init();
    }

    init() {
        console.log('SaaS Admin Dashboard initializing...');
        this.setupEventListeners();
        this.loadInitialSection();
    }

    setupEventListeners() {
        // Navigation click handlers
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const href = link.getAttribute('href');
                if (href && href.startsWith('#')) {
                    const section = href.substring(1);
                    console.log('Navigation clicked:', section);
                    this.showSection(section);
                }
            });
        });

        // Hash change handler
        window.addEventListener('hashchange', () => {
            const hash = window.location.hash.substring(1);
            if (hash) {
                this.showSection(hash);
            }
        });
    }

    loadInitialSection() {
        const hash = window.location.hash.substring(1);
        if (hash) {
            this.showSection(hash);
        } else {
            this.showSection('overview');
        }
    }

    showSection(section) {
        console.log('showSection called with:', section);
        
        // Update current section
        this.currentSection = section;
        
        // Update URL hash
        if (window.location.hash !== `#${section}`) {
            window.location.hash = `#${section}`;
        }

        // Hide all content sections
        document.querySelectorAll('.content-section').forEach(el => {
            el.style.display = 'none';
        });

        // Show the selected section
        const sectionElement = document.getElementById(`${section}-content`);
        if (sectionElement) {
            console.log('Found section element:', sectionElement.id);
            sectionElement.style.display = 'block';
            sectionElement.style.visibility = 'visible';
            sectionElement.style.opacity = '1';
            
            // Load section data
            this.loadSectionData(section);
        } else {
            console.error('Section element not found:', `${section}-content`);
        }

        // Update active navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        const activeLink = document.querySelector(`[href="#${section}"]`);
        if (activeLink) {
            activeLink.closest('.nav-item')?.classList.add('active');
        }
    }

    async loadSectionData(section) {
        console.log('Loading data for section:', section);
        
        const loadingElement = document.getElementById(`${section}-loading`);
        const contentElement = document.getElementById(`${section}-content`);
        
        // Show loading
        if (loadingElement) {
            loadingElement.style.display = 'block';
        }
        if (contentElement) {
            contentElement.style.display = 'none';
        }

        try {
            switch (section) {
                case 'overview':
                    await this.loadOverviewData();
                    break;
                case 'tenants':
                    await this.loadTenantsData();
                    break;
                case 'subscriptions':
                    await this.loadSubscriptionsData();
                    break;
                case 'plans':
                    await this.loadPlansData();
                    break;
                case 'billing':
                    await this.loadBillingData();
                    break;
                case 'analytics':
                    await this.loadAnalyticsData();
                    break;
                case 'settings':
                    await this.loadSettingsData();
                    break;
                case 'feature-flags':
                    await this.loadFeatureFlagsData();
                    break;
                case 'integrations':
                    await this.loadIntegrationsData();
                    break;
                default:
                    console.log('No specific data loader for section:', section);
            }
        } catch (error) {
            console.error(`Error loading ${section} data:`, error);
            this.showError(section, error.message);
        } finally {
            // Hide loading and show content
            if (loadingElement) {
                loadingElement.style.display = 'none';
            }
            if (contentElement) {
                contentElement.style.display = 'block';
                contentElement.style.visibility = 'visible';
                contentElement.style.opacity = '1';
            }
        }
    }

    showError(section, message) {
        const contentElement = document.getElementById(`${section}-content`);
        if (contentElement) {
            contentElement.innerHTML = `
                <div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error: ${message}</p>
                    <button class="btn-primary" onclick="saasAdmin.loadSectionData('${section}')">Retry</button>
                </div>
            `;
        }
    }

    async loadOverviewData() {
        console.log('Loading overview data...');
        // Overview data loading logic here
    }

    async loadTenantsData() {
        console.log('Loading tenants data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/tenants');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Tenants data received:', data);
            this.renderTenantsTable(data.tenants || []);
        } catch (error) {
            console.error('Error loading tenants:', error);
            throw error;
        }
    }

    renderTenantsTable(tenants) {
        console.log('Rendering tenants table with', tenants.length, 'tenants');
        const tbody = document.getElementById('tenantsTableBody');
        if (!tbody) {
            console.error('tenantsTableBody not found');
            return;
        }

        tbody.innerHTML = '';

        if (tenants.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="text-center">
                        <div class="empty-state">
                            <i class="fas fa-users"></i>
                            <p>No tenants found</p>
                        </div>
                    </td>
                </tr>
            `;
            return;
        }

        tenants.forEach(tenant => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${this.escapeHtml(tenant.Name)}</td>
                <td>${this.escapeHtml(tenant.Slug)}</td>
                <td>${this.escapeHtml(tenant.Plan)}</td>
                <td><span class="status-badge status-${tenant.Status}">${this.escapeHtml(tenant.Status)}</span></td>
                <td>${new Date(tenant.CreatedAt).toLocaleDateString()}</td>
                <td>
                    <button class="btn-secondary" onclick="saasAdmin.editTenant(${tenant.ID})">Edit</button>
                    <button class="btn-danger" onclick="saasAdmin.deleteTenant(${tenant.ID})">Delete</button>
                </td>
            `;
            tbody.appendChild(row);
        });
        
        console.log('Tenants table rendered successfully');
    }

    async loadSubscriptionsData() {
        console.log('Loading subscriptions data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/subscriptions');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Subscriptions data received:', data);
            this.renderSubscriptionsTable(data.subscriptions || []);
        } catch (error) {
            console.error('Error loading subscriptions:', error);
            throw error;
        }
    }

    renderSubscriptionsTable(subscriptions) {
        console.log('Rendering subscriptions table with', subscriptions.length, 'subscriptions');
        const tbody = document.getElementById('subscriptionsTableBody');
        if (!tbody) {
            console.error('subscriptionsTableBody not found');
            return;
        }

        tbody.innerHTML = '';

        if (subscriptions.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="text-center">
                        <div class="empty-state">
                            <i class="fas fa-credit-card"></i>
                            <p>No subscriptions found</p>
                        </div>
                    </td>
                </tr>
            `;
            return;
        }

        subscriptions.forEach(subscription => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${this.escapeHtml(subscription.TenantName || 'N/A')}</td>
                <td>${this.escapeHtml(subscription.PlanName || 'N/A')}</td>
                <td><span class="status-badge status-${subscription.Status}">${this.escapeHtml(subscription.Status)}</span></td>
                <td>$${subscription.Amount || '0'}</td>
                <td>${new Date(subscription.CreatedAt).toLocaleDateString()}</td>
            `;
            tbody.appendChild(row);
        });
        
        console.log('Subscriptions table rendered successfully');
    }

    async loadPlansData() {
        console.log('Loading plans data...');
        // Plans data loading logic here
    }

    async loadBillingData() {
        console.log('Loading billing data...');
        // Billing data loading logic here
    }

    async loadAnalyticsData() {
        console.log('Loading analytics data...');
        // Analytics data loading logic here
    }

    async loadSettingsData() {
        console.log('Loading settings data...');
        // Settings data loading logic here
    }

    async loadFeatureFlagsData() {
        console.log('Loading feature flags data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/feature-availability');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Feature flags data received:', data);
            this.renderFeatureFlagsList(data.features || []);
        } catch (error) {
            console.error('Error loading feature flags:', error);
            throw error;
        }
    }

    renderFeatureFlagsList(features) {
        console.log('Rendering feature flags list with', features.length, 'features');
        const container = document.getElementById('feature-flags-content');
        if (!container) {
            console.error('feature-flags-content not found');
            return;
        }

        if (features.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-flag"></i>
                    <p>No feature flags found</p>
                </div>
            `;
            return;
        }

        const featuresHtml = features.map(flag => `
            <div class="feature-flag-card">
                <div class="feature-flag-header">
                    <h3>${this.escapeHtml(flag.Feature)}</h3>
                    <div class="feature-flag-toggle">
                        <label class="toggle-switch">
                            <input type="checkbox" ${flag.IsAvailable ? 'checked' : ''} 
                                   onchange="saasAdmin.toggleFeatureAvailability('${this.escapeHtml(flag.Feature)}', this.checked)">
                            <span class="toggle-slider"></span>
                        </label>
                    </div>
                </div>
                <div class="feature-flag-description">
                    <p>${this.escapeHtml(flag.Description || 'No description provided')}</p>
                </div>
            </div>
        `).join('');

        container.innerHTML = `
            <div class="feature-flags-grid">
                ${featuresHtml}
            </div>
        `;
        
        console.log('Feature flags list rendered successfully');
    }

    async loadIntegrationsData() {
        console.log('Loading integrations data...');
        // Integrations data loading logic here
    }

    // Utility functions
    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Action handlers
    editTenant(tenantId) {
        console.log('Edit tenant:', tenantId);
        // Edit tenant logic here
    }

    deleteTenant(tenantId) {
        console.log('Delete tenant:', tenantId);
        // Delete tenant logic here
    }

    async toggleFeatureAvailability(feature, isAvailable) {
        console.log('Toggle feature availability:', feature, isAvailable);
        try {
            const response = await fetch('/api/v1/admin/saas/admin/feature-availability', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    feature: feature,
                    isAvailable: isAvailable
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            console.log('Feature availability updated successfully');
        } catch (error) {
            console.error('Error updating feature availability:', error);
            // Revert the toggle
            const checkbox = document.querySelector(`input[onchange*="${feature}"]`);
            if (checkbox) {
                checkbox.checked = !isAvailable;
            }
        }
    }
}

// Initialize the dashboard when the page loads
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, initializing SaaS Admin Dashboard...');
    window.saasAdmin = new SaaSAdminDashboard();
    console.log('SaaS Admin Dashboard initialized');
});
