// SaaS Admin Dashboard JavaScript - Based on Working Admin.js

class SaaSAdminDashboard {
    constructor() {
        console.log('SaaSAdminDashboard constructor called');
        this.currentSection = 'overview';
        this.init();
    }

    init() {
        console.log('SaaSAdminDashboard init called');
        this.setupEventListeners();
        
        // Check for initial hash and navigate accordingly
        setTimeout(() => {
            const initialHash = window.location.hash.substring(1);
            console.log('Initial hash:', initialHash);
            
            if (initialHash && initialHash !== 'overview') {
                // Navigate to the specified section
                this.navigateToSection(initialHash);
            } else {
                // Default to overview
                this.loadOverviewData();
            }
        }, 100);
    }

    setupEventListeners() {
        // Sidebar navigation
        this.setupSidebarNavigation();

        // Handle browser back/forward buttons
        window.addEventListener('popstate', (e) => {
            if (e.state && e.state.section) {
                this.navigateToSection(e.state.section);
            }
        });

        // Handle hash changes (when user manually changes URL)
        window.addEventListener('hashchange', (e) => {
            const newHash = window.location.hash.substring(1);
            console.log('Hash changed to:', newHash);
            if (newHash) {
                this.navigateToSection(newHash);
            } else {
                // If no hash, go to overview
                this.loadOverviewData();
            }
        });
    }

    setupSidebarNavigation() {
        const navLinks = document.querySelectorAll('.sidebar-nav a');
        console.log('Setting up sidebar navigation, found links:', navLinks.length);
        
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const href = link.getAttribute('href');
                console.log('Nav link clicked:', href);
                
                if (href.startsWith('#')) {
                    // Handle hash-based navigation
                    const section = href.substring(1);
                    console.log('Navigating to section:', section);
                    this.navigateToSection(section);
                } else if (href.startsWith('/')) {
                    // Handle direct URL navigation
                    console.log('Navigating to URL:', href);
                    window.location.href = href;
                }
            });
        });
    }

    async navigateToSection(section) {
        console.log('navigateToSection called with:', section);
        console.log('Current URL:', window.location.href);
        
        // Hide all content sections first - be very explicit
        const allSections = document.querySelectorAll('.content-section');
        console.log('Hiding all sections, found:', allSections.length);
        allSections.forEach(s => {
            s.style.display = 'none';
            s.style.visibility = 'hidden';
            console.log('Hiding section:', s.id);
        });
        
        // Show the target section
        const targetSection = document.getElementById(section);
        console.log('Target section found:', targetSection);
        if (targetSection) {
            targetSection.style.display = 'block';
            targetSection.style.visibility = 'visible';
            console.log('Showing section:', section);
        } else {
            console.error('Target section not found:', section);
            return;
        }

        // Update URL hash
        if (window.location.hash !== `#${section}`) {
            window.location.hash = `#${section}`;
            console.log('Updated URL hash to:', `#${section}`);
        }

        // Update current section
        this.currentSection = section;

        // Load section-specific data
        await this.loadSectionData(section);
    }

    async loadSectionData(section) {
        console.log('Loading data for section:', section);
        
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
                case 'analytics':
                    await this.loadAnalyticsData();
                    break;
                case 'billing':
                    await this.loadBillingData();
                    break;
                case 'feature-flags':
                    await this.loadFeatureFlagsData();
                    break;
                case 'settings':
                    await this.loadSettingsData();
                    break;
                case 'users':
                    await this.loadUsersData();
                    break;
                case 'integrations':
                    await this.loadIntegrationsData();
                    break;
                case 'notifications':
                    await this.loadNotificationsData();
                    break;
                default:
                    console.log('No specific data loading for section:', section);
            }
        } catch (error) {
            console.error('Error loading section data:', error);
        }
    }

    async loadOverviewData() {
        console.log('Loading overview data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Overview data received:', data);
            this.renderOverview(data);
        } catch (error) {
            console.error('Error loading overview:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('overview-loading');
            const content = document.getElementById('overview-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden overview loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed overview content');
            }
        }
    }

    renderOverview(data) {
        console.log('Rendering overview with data:', data);
        // Overview rendering logic here
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
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('tenants-loading');
            const content = document.getElementById('tenants-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden tenants loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed tenants content');
            }
        }
    }

    renderTenantsTable(tenants) {
        console.log('Rendering tenants table with:', tenants.length, 'tenants');
        const tbody = document.getElementById('tenantsTableBody');
        if (!tbody) {
            console.error('tenantsTableBody not found');
            return;
        }

        tbody.innerHTML = '';
        tenants.forEach(tenant => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${tenant.Name}</td>
                <td>${tenant.Domain}</td>
                <td>${tenant.Status}</td>
                <td>${tenant.CreatedAt ? new Date(tenant.CreatedAt).toLocaleDateString() : 'N/A'}</td>
                <td>
                    <button class="btn-secondary" onclick="saasAdmin.editTenant('${tenant.ID}')">Edit</button>
                    <button class="btn-danger" onclick="saasAdmin.deleteTenant('${tenant.ID}')">Delete</button>
                </td>
            `;
            tbody.appendChild(row);
        });
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
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('subscriptions-loading');
            const content = document.getElementById('subscriptions-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden subscriptions loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed subscriptions content');
            }
        }
    }

    renderSubscriptionsTable(subscriptions) {
        console.log('Rendering subscriptions table with:', subscriptions.length, 'subscriptions');
        const tbody = document.getElementById('subscriptionsTableBody');
        if (!tbody) {
            console.error('subscriptionsTableBody not found');
            return;
        }

        tbody.innerHTML = '';
        subscriptions.forEach(subscription => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${subscription.TenantName || 'N/A'}</td>
                <td>${subscription.PlanName || 'N/A'}</td>
                <td>${subscription.Status || 'N/A'}</td>
                <td>${subscription.CreatedAt ? new Date(subscription.CreatedAt).toLocaleDateString() : 'N/A'}</td>
                <td>
                    <button class="btn-secondary" onclick="saasAdmin.editSubscription('${subscription.ID}')">Edit</button>
                    <button class="btn-danger" onclick="saasAdmin.cancelSubscription('${subscription.ID}')">Cancel</button>
                </td>
            `;
            tbody.appendChild(row);
        });
    }

    async loadAnalyticsData() {
        console.log('Loading analytics data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Analytics data received:', data);
            this.renderAnalyticsCharts(data);
        } catch (error) {
            console.error('Error loading analytics:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('analytics-loading');
            const content = document.getElementById('analytics-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden analytics loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed analytics content');
            }
        }
    }

    renderAnalyticsCharts(data) {
        console.log('Rendering analytics charts with data:', data);
        // Analytics rendering logic here
    }

    async loadPlansData() {
        console.log('Loading plans data...');
        try {
            // Simulate plans data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Plans data loaded');
        } catch (error) {
            console.error('Error loading plans data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('plans-loading');
            const content = document.getElementById('plans-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden plans loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed plans content');
            }
        }
    }

    async loadBillingData() {
        console.log('Loading billing data...');
        try {
            // Simulate billing data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Billing data loaded');
        } catch (error) {
            console.error('Error loading billing data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('billing-loading');
            const content = document.getElementById('billing-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden billing loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed billing content');
            }
        }
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
            this.renderFeatureFlagsList(data.feature_availability || []);
        } catch (error) {
            console.error('Error loading feature flags:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('feature-flags-loading');
            const content = document.getElementById('feature-flags-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden feature flags loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed feature flags content');
            }
        }
    }

    renderFeatureFlagsList(features) {
        console.log('Rendering feature flags list with:', features.length, 'features');
        const container = document.getElementById('feature-flags-content');
        if (!container) {
            console.error('feature-flags-content not found');
            return;
        }

        if (features.length === 0) {
            container.innerHTML = '<p>No feature flags available.</p>';
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
                    <p>${flag.Description || 'No description provided'}</p>
                </div>
                <div class="feature-flag-actions">
                    <button class="btn-secondary" onclick="saasAdmin.editFeatureFlag('${this.escapeHtml(flag.Feature)}')">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <button class="btn-danger" onclick="saasAdmin.deleteFeatureFlag('${this.escapeHtml(flag.Feature)}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </div>
            </div>
        `).join('');

        container.innerHTML = featuresHtml;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    async loadSettingsData() {
        console.log('Loading settings data...');
        try {
            // Simulate settings data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Settings data loaded');
        } catch (error) {
            console.error('Error loading settings data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('settings-loading');
            const content = document.getElementById('settings-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden settings loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed settings content');
            }
        }
    }

    async loadUsersData() {
        console.log('Loading users data...');
        try {
            // Simulate users data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Users data loaded');
        } catch (error) {
            console.error('Error loading users data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('users-loading');
            const content = document.getElementById('users-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden users loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed users content');
            }
        }
    }

    async loadIntegrationsData() {
        console.log('Loading integrations data...');
        try {
            // Simulate integrations data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Integrations data loaded');
        } catch (error) {
            console.error('Error loading integrations data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('integrations-loading');
            const content = document.getElementById('integrations-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden integrations loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed integrations content');
            }
        }
    }

    async loadNotificationsData() {
        console.log('Loading notifications data...');
        try {
            // Simulate notifications data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Notifications data loaded');
        } catch (error) {
            console.error('Error loading notifications data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('notifications-loading');
            const content = document.getElementById('notifications-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden notifications loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed notifications content');
            }
        }
    }

    // Feature flag methods
    async toggleFeatureAvailability(feature, isAvailable) {
        console.log('Toggling feature availability:', feature, isAvailable);
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
        }
    }

    editFeatureFlag(feature) {
        console.log('Editing feature flag:', feature);
        // Edit feature flag logic here
    }

    deleteFeatureFlag(feature) {
        console.log('Deleting feature flag:', feature);
        // Delete feature flag logic here
    }

    // Tenant methods
    editTenant(tenantId) {
        console.log('Editing tenant:', tenantId);
        // Edit tenant logic here
    }

    deleteTenant(tenantId) {
        console.log('Deleting tenant:', tenantId);
        // Delete tenant logic here
    }

    // Subscription methods
    editSubscription(subscriptionId) {
        console.log('Editing subscription:', subscriptionId);
        // Edit subscription logic here
    }

    cancelSubscription(subscriptionId) {
        console.log('Cancelling subscription:', subscriptionId);
        // Cancel subscription logic here
    }
}

// Initialize the dashboard when the page loads
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, initializing SaaS Admin Dashboard...');
    window.saasAdmin = new SaaSAdminDashboard();
    console.log('SaaS Admin Dashboard initialized');
});