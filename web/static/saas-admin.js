// SaaS Admin Dashboard JavaScript

class SaaSAdminDashboard {
    constructor() {
        this.currentSection = 'overview';
        this.charts = {};
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupModals();
        
        // Check authentication first
        this.checkAuthentication().then(() => {
            // Check for initial hash and navigate accordingly
            setTimeout(() => {
                const initialHash = window.location.hash.substring(1);
                console.log('Initial hash:', initialHash);
                
                if (initialHash && initialHash !== 'overview') {
                    // Navigate to the specified section
                    this.showSection(initialHash);
                } else {
                    // Default to overview
                    this.loadOverviewData();
                }
            }, 100);
        }).catch(() => {
            // Redirect to login if not authenticated
            window.location.href = '/admin/login';
        });
    }

    async checkAuthentication() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/settings', { 
                credentials: 'include',
                method: 'GET'
            });
            
            if (response.status === 401 || response.status === 403) {
                throw new Error('Not authenticated');
            }
            
            return true;
        } catch (error) {
            console.error('Authentication check failed:', error);
            throw error;
        }
    }

    async authenticatedFetch(url, options = {}) {
        const defaultOptions = {
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            }
        };

        const response = await fetch(url, { ...defaultOptions, ...options });
        
        if (response.status === 401 || response.status === 403) {
            // Redirect to login if not authenticated
            window.location.href = '/admin/login';
            throw new Error('Authentication required');
        }
        
        return response;
    }

    setupEventListeners() {
        // Sidebar navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                const href = link.getAttribute('href');
                
                // Allow external navigation (non-hash links)
                if (!href.startsWith('#')) {
                    return; // Let the browser handle the navigation
                }
                
                e.preventDefault();
                const section = href.substring(1);
                this.showSection(section);
            });
        });

        // Sidebar toggle for mobile
        const sidebarToggle = document.getElementById('sidebarToggle');
        const sidebar = document.querySelector('.sidebar');
        
        if (sidebarToggle) {
            sidebarToggle.addEventListener('click', () => {
                sidebar.classList.toggle('open');
            });
        }

        // Close sidebar when clicking outside on mobile
        document.addEventListener('click', (e) => {
            if (window.innerWidth <= 768 && sidebar.classList.contains('open')) {
                if (!sidebar.contains(e.target) && !sidebarToggle.contains(e.target)) {
                    sidebar.classList.remove('open');
                }
            }
        });

        // Close sidebar when window is resized to desktop size
        window.addEventListener('resize', () => {
            if (window.innerWidth > 768) {
                sidebar.classList.remove('open');
            }
        });

        // Quick action buttons
        document.getElementById('createTenantBtn')?.addEventListener('click', () => {
            this.openModal('tenantModal');
        });

        document.getElementById('createPlanBtn')?.addEventListener('click', () => {
            this.openModal('planModal');
        });

        document.getElementById('createNotificationBtn')?.addEventListener('click', () => {
            this.openModal('notificationModal');
        });

        // Form submissions
        document.getElementById('tenantForm')?.addEventListener('submit', (e) => {
            this.handleTenantSubmit(e);
        });

        document.getElementById('planForm')?.addEventListener('submit', (e) => {
            this.handlePlanSubmit(e);
        });

        document.getElementById('notificationForm')?.addEventListener('submit', (e) => {
            this.handleNotificationSubmit(e);
        });

        // Cancel buttons
        document.getElementById('cancelTenant')?.addEventListener('click', () => {
            this.closeModal('tenantModal');
        });

        document.getElementById('cancelPlan')?.addEventListener('click', () => {
            this.closeModal('planModal');
        });

        document.getElementById('cancelNotification')?.addEventListener('click', () => {
            this.closeModal('notificationModal');
        });

        // Filters
        document.getElementById('tenantSearch')?.addEventListener('input', (e) => {
            this.filterTenants();
        });

        document.getElementById('tenantStatusFilter')?.addEventListener('change', (e) => {
            this.filterTenants();
        });

        document.getElementById('tenantPlanFilter')?.addEventListener('change', (e) => {
            this.filterTenants();
        });

        // Analytics date range
        document.getElementById('analyticsDateRange')?.addEventListener('change', (e) => {
            this.loadAnalyticsData();
        });

        // Handle hash changes (for refresh navigation)
        window.addEventListener('hashchange', (e) => {
            const newHash = window.location.hash.substring(1);
            console.log('Hash changed to:', newHash);
            if (newHash) {
                this.showSection(newHash);
            } else {
                // If no hash, go to overview
                this.loadOverviewData();
            }
        });
    }

    setupModals() {
        // Close modals when clicking outside
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.closeModal(modal.id);
                }
            });
        });

        // Close modals with close buttons
        document.querySelectorAll('.close-button').forEach(button => {
            button.addEventListener('click', (e) => {
                const modal = e.target.closest('.modal');
                if (modal) {
                    this.closeModal(modal.id);
                }
            });
        });
    }

    showSection(section) {
        // Update URL hash
        if (window.location.hash !== `#${section}`) {
            window.location.hash = `#${section}`;
            console.log('Updated URL hash to:', `#${section}`);
        }

        // Hide all sections (both dashboard-content and content-section)
        document.querySelectorAll('.dashboard-content, .content-section').forEach(content => {
            content.style.display = 'none';
        });

        // Show selected section
        let sectionElement = document.getElementById(`${section}-content`);
        if (!sectionElement) {
            // Try with just the section name
            sectionElement = document.getElementById(section);
        }
        
        if (sectionElement) {
            sectionElement.style.display = 'block';
            // If it's a content-section, load data dynamically
            if (sectionElement.classList.contains('content-section')) {
                this.loadSectionData(section);
            }
        }

        // Update active nav item
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[href="#${section}"]`)?.closest('.nav-item')?.classList.add('active');

        // Update page title
        const pageTitle = document.querySelector('.page-title');
        if (pageTitle) {
            const titles = {
                'overview': 'SaaS Platform Overview',
                'tenants': 'Tenant Management',
                'subscriptions': 'Subscription Management',
                'plans': 'Subscription Plans',
                'billing': 'Billing & Payments',
                'analytics': 'Platform Analytics',
                'settings': 'Platform Settings',
                'notifications': 'System Notifications'
            };
            pageTitle.textContent = titles[section] || 'SaaS Admin Dashboard';
        }

        // Load section-specific data
        this.currentSection = section;
        this.loadSectionData(section);
    }

    async loadSectionData(section) {
        const loadingElement = document.getElementById(`${section}-loading`);
        const contentElement = document.getElementById(`${section}-content`);
        
        // Show loading spinner for new content sections
        if (loadingElement) loadingElement.style.display = 'block';
        if (contentElement) contentElement.style.display = 'none';

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
                case 'saas-settings':
                    await this.loadSettingsData();
                    break;
                case 'notifications':
                    await this.loadNotificationsData();
                    break;
                default:
                    console.log(`No specific loader for section: ${section}`);
            }
        } catch (error) {
            console.error(`Error loading ${section} data:`, error);
            if (contentElement) {
                contentElement.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading ${section} data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadSectionData('${section}')">Retry</button>
                </div>`;
            }
        } finally {
            // Hide loading spinner and show content
            if (loadingElement) loadingElement.style.display = 'none';
            if (contentElement) contentElement.style.display = 'block';
        }
    }

    async loadOverviewData() {
        try {
            await Promise.all([
                this.loadOverviewMetrics(),
                this.loadRecentActivity(),
                this.loadPlanDistribution(),
                this.loadOverviewCharts()
            ]);
        } catch (error) {
            console.error('Error loading overview data:', error);
            this.showNotification('Failed to load overview data', 'error');
        }
    }

    async loadOverviewMetrics() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/metrics', { credentials: 'include' });
            const metrics = await response.json();

            document.getElementById('totalTenants').textContent = metrics.total_tenants || 0;
            document.getElementById('activeTenants').textContent = metrics.active_tenants || 0;
            document.getElementById('monthlyRevenue').textContent = `$${metrics.monthly_revenue || 0}`;
            document.getElementById('churnRate').textContent = `${metrics.churn_rate || 0}%`;
        } catch (error) {
            console.error('Error loading overview metrics:', error);
        }
    }

    async loadRecentActivity() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/activity', { credentials: 'include' });
            const activities = await response.json();

            const activityList = document.getElementById('recentActivityList');
            if (!activityList) return;

            activityList.innerHTML = '';

            if (activities && activities.length > 0) {
                activities.forEach(activity => {
                    const activityItem = this.createActivityItem(activity);
                    activityList.appendChild(activityItem);
                });
            } else {
                activityList.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-history"></i>
                        <h3>No Recent Activity</h3>
                        <p>Activity will appear here as users interact with the platform</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading recent activity:', error);
        }
    }

    createActivityItem(activity) {
        const item = document.createElement('div');
        item.className = 'activity-item';
        
        const timeAgo = this.getTimeAgo(new Date(activity.created_at));
        
        item.innerHTML = `
            <div class="activity-icon">
                <i class="fas fa-${this.getActivityIcon(activity.type)}"></i>
            </div>
            <div class="activity-content">
                <p><strong>${activity.title}</strong> ${activity.description}</p>
                <span class="activity-time">${timeAgo}</span>
            </div>
        `;
        
        return item;
    }

    getActivityIcon(type) {
        const icons = {
            'tenant_created': 'building',
            'subscription_created': 'credit-card',
            'payment_received': 'dollar-sign',
            'tenant_suspended': 'ban',
            'plan_updated': 'edit'
        };
        return icons[type] || 'info-circle';
    }

    async loadPlanDistribution() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/plan-distribution', { credentials: 'include' });
            const distribution = await response.json();

            const planStats = document.getElementById('planStats');
            if (!planStats) return;

            planStats.innerHTML = '';

            if (distribution && distribution.length > 0) {
                distribution.forEach(plan => {
                    const planStat = this.createPlanStat(plan);
                    planStats.appendChild(planStat);
                });
            } else {
                planStats.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-chart-pie"></i>
                        <h3>No Plan Data</h3>
                        <p>Plan distribution data will appear here</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading plan distribution:', error);
        }
    }

    createPlanStat(plan) {
        const stat = document.createElement('div');
        stat.className = 'plan-stat';
        
        stat.innerHTML = `
            <h4>${plan.name}</h4>
            <div class="count">${plan.count}</div>
            <div class="percentage">${plan.percentage}%</div>
        `;
        
        return stat;
    }

    async loadOverviewCharts() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/charts/overview', { credentials: 'include' });
            const chartData = await response.json();

            this.createRevenueChart(chartData.revenue);
            this.createTenantGrowthChart(chartData.tenant_growth);
        } catch (error) {
            console.error('Error loading overview charts:', error);
        }
    }

    createRevenueChart(data) {
        const ctx = document.getElementById('revenueChart');
        if (!ctx) return;

        if (this.charts.revenue) {
            this.charts.revenue.destroy();
        }

        this.charts.revenue = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels || [],
                datasets: [{
                    label: 'Revenue',
                    data: data.values || [],
                    borderColor: '#667eea',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            callback: function(value) {
                                return '$' + value;
                            }
                        }
                    }
                }
            }
        });
    }

    createTenantGrowthChart(data) {
        const ctx = document.getElementById('tenantGrowthChart');
        if (!ctx) return;

        if (this.charts.tenantGrowth) {
            this.charts.tenantGrowth.destroy();
        }

        this.charts.tenantGrowth = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: data.labels || [],
                datasets: [{
                    label: 'New Tenants',
                    data: data.values || [],
                    backgroundColor: '#4facfe',
                    borderColor: '#00f2fe',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    async loadTenantsData() {
        try {
            const response = await this.authenticatedFetch('/api/v1/admin/saas/admin/tenants');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const tenants = data.tenants || [];
            this.renderTenantsTable(tenants);
        } catch (error) {
            console.error('Error loading tenants:', error);
            const table = document.getElementById('tenantsTable');
            if (table) {
                table.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading tenants data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadTenantsData()">Retry</button>
                </div>`;
            }
        }
    }

    renderTenantsTable(tenants) {
        const tbody = document.getElementById('tenantsTableBody');
        if (!tbody) return;

        tbody.innerHTML = '';

        if (tenants && tenants.length > 0) {
            tenants.forEach(tenant => {
                const row = this.createTenantRow(tenant);
                tbody.appendChild(row);
            });
        } else {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="empty-state">
                        <i class="fas fa-building"></i>
                        <h3>No Tenants</h3>
                        <p>Create your first tenant to get started</p>
                    </td>
                </tr>
            `;
        }
    }

    createTenantRow(tenant) {
        const row = document.createElement('tr');
        
        const createdAt = new Date(tenant.CreatedAt).toLocaleDateString();
        
        row.innerHTML = `
            <td>
                <div class="tenant-info">
                    <strong>${tenant.Name}</strong>
                    <small>${tenant.ContactEmail || tenant.BillingEmail || 'No email'}</small>
                </div>
            </td>
            <td>
                <code>${tenant.Slug}</code>
            </td>
            <td>
                <span class="plan-badge ${tenant.Plan}">${tenant.Plan}</span>
            </td>
            <td>
                <span class="status-badge ${tenant.Status}">${tenant.Status}</span>
            </td>
            <td>${createdAt}</td>
            <td>
                <div class="action-buttons">
                    <button class="action-btn edit" onclick="saasAdmin.editTenant(${tenant.ID})">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn suspend" onclick="saasAdmin.suspendTenant(${tenant.ID})">
                        <i class="fas fa-ban"></i>
                    </button>
                    <button class="action-btn delete" onclick="saasAdmin.deleteTenant(${tenant.ID})">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </td>
        `;
        
        return row;
    }

    filterTenants() {
        const search = document.getElementById('tenantSearch').value.toLowerCase();
        const statusFilter = document.getElementById('tenantStatusFilter').value;
        const planFilter = document.getElementById('tenantPlanFilter').value;

        const rows = document.querySelectorAll('#tenantsTableBody tr');
        rows.forEach(row => {
            const name = row.querySelector('strong')?.textContent.toLowerCase() || '';
            const email = row.querySelector('small')?.textContent.toLowerCase() || '';
            const status = row.querySelector('.status-badge')?.textContent.toLowerCase() || '';
            const plan = row.querySelector('.plan-badge')?.textContent.toLowerCase() || '';

            const matchesSearch = name.includes(search) || email.includes(search);
            const matchesStatus = !statusFilter || status === statusFilter;
            const matchesPlan = !planFilter || plan === planFilter;

            row.style.display = matchesSearch && matchesStatus && matchesPlan ? '' : 'none';
        });
    }

    async loadSubscriptionsData() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/subscriptions', { credentials: 'include' });
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const subscriptions = data.subscriptions || [];
            this.renderSubscriptionsTable(subscriptions);
        } catch (error) {
            console.error('Error loading subscriptions:', error);
            const table = document.getElementById('subscriptionsTable');
            if (table) {
                table.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading subscriptions data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadSubscriptionsData()">Retry</button>
                </div>`;
            }
        }
    }

    renderSubscriptionsTable(subscriptions) {
        const tbody = document.getElementById('subscriptionsTableBody');
        if (!tbody) return;

        tbody.innerHTML = '';

        if (subscriptions && subscriptions.length > 0) {
            subscriptions.forEach(subscription => {
                const row = this.createSubscriptionRow(subscription);
                tbody.appendChild(row);
            });
        } else {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="empty-state">
                        <i class="fas fa-credit-card"></i>
                        <h3>No Subscriptions</h3>
                        <p>Subscriptions will appear here when tenants subscribe to plans</p>
                    </td>
                </tr>
            `;
        }
    }

    createSubscriptionRow(subscription) {
        const row = document.createElement('tr');
        
        const nextBilling = subscription.CurrentPeriodEnd ? 
            new Date(subscription.CurrentPeriodEnd).toLocaleDateString() : 'N/A';
        
        row.innerHTML = `
            <td>
                <div class="tenant-info">
                    <strong>${subscription.Tenant?.Name || 'Unknown Tenant'}</strong>
                    <small>${subscription.Tenant?.Slug || 'No slug'}</small>
                </div>
            </td>
            <td>
                <span class="plan-badge ${subscription.Plan?.Slug || 'unknown'}">${subscription.Plan?.Name || 'Unknown Plan'}</span>
            </td>
            <td>
                <span class="status-badge ${subscription.Status}">${subscription.Status}</span>
            </td>
            <td>$${subscription.Plan?.Price || 0}</td>
            <td>${nextBilling}</td>
            <td>
                <div class="action-buttons">
                    <button class="action-btn edit" onclick="saasAdmin.editSubscription(${subscription.ID})">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn suspend" onclick="saasAdmin.cancelSubscription(${subscription.ID})">
                        <i class="fas fa-ban"></i>
                    </button>
                </div>
            </td>
        `;
        
        return row;
    }

    async loadPlansData() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/plans', { credentials: 'include' });
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const plans = data.plans || [];
            
            this.renderPlansGrid(plans);
        } catch (error) {
            console.error('Error loading plans:', error);
            
            // Show error state
            const grid = document.getElementById('plansGrid');
            if (grid) {
                grid.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-triangle"></i>
                        <h3>Error loading plans data</h3>
                        <p>Please try again</p>
                        <button class="btn-primary" onclick="saasAdmin.loadPlansData()">Retry</button>
                    </div>
                `;
            }
        }
    }

    renderPlansGrid(plans) {
        const grid = document.getElementById('plansGrid');
        if (!grid) return;

        grid.innerHTML = '';

        if (plans && plans.length > 0) {
            plans.forEach(plan => {
                const planCard = this.createPlanCard(plan);
                grid.appendChild(planCard);
            });
        } else {
            grid.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-list-alt"></i>
                    <h3>No Subscription Plans</h3>
                    <p>Create your first subscription plan to get started</p>
                </div>
            `;
        }
    }

    createPlanCard(plan) {
        const card = document.createElement('div');
        card.className = 'plan-card';
        
        if (plan.Slug === 'pro') {
            card.classList.add('popular');
        }
        
        // Build features list from boolean fields
        const features = [];
        if (plan.CustomDomain) features.push('Custom Domain');
        if (plan.WhiteLabel) features.push('White Label');
        if (plan.API) features.push('API Access');
        if (plan.Integrations) features.push('Integrations');
        if (plan.Analytics) features.push('Analytics');
        
        const featuresList = features.map(feature => 
            `<li><i class="fas fa-check"></i> ${feature}</li>`
        ).join('');
        
        card.innerHTML = `
            <div class="plan-header">
                <h3 class="plan-name">${plan.Name}</h3>
                <div class="plan-price">
                    <span class="currency">$</span>${plan.Price}
                    <span class="period">/${plan.BillingInterval}</span>
                </div>
                <p class="plan-description">${plan.Description || 'No description available'}</p>
            </div>
            
            <div class="plan-limits">
                <h4>Limits</h4>
                <div class="limit">
                    <span>Services</span>
                    <span>${plan.MaxServices === -1 ? 'Unlimited' : plan.MaxServices}</span>
                </div>
                <div class="limit">
                    <span>Monitors</span>
                    <span>${plan.MaxMonitors === -1 ? 'Unlimited' : plan.MaxMonitors}</span>
                </div>
                <div class="limit">
                    <span>Subscribers</span>
                    <span>${plan.MaxSubscribers === -1 ? 'Unlimited' : plan.MaxSubscribers}</span>
                </div>
                <div class="limit">
                    <span>Incidents</span>
                    <span>${plan.MaxIncidents === -1 ? 'Unlimited' : plan.MaxIncidents}</span>
                </div>
                <div class="limit">
                    <span>Maintenance</span>
                    <span>${plan.MaxMaintenance === -1 ? 'Unlimited' : plan.MaxMaintenance}</span>
                </div>
            </div>
            
            <ul class="plan-features">
                ${featuresList}
            </ul>
            
            <div class="plan-actions">
                <button class="btn-secondary" onclick="saasAdmin.editPlan(${plan.ID})">
                    <i class="fas fa-edit"></i> Edit Plan
                </button>
                <button class="btn-primary" onclick="saasAdmin.duplicatePlan(${plan.ID})">
                    <i class="fas fa-copy"></i> Duplicate
                </button>
                <button class="btn-danger" onclick="saasAdmin.deletePlan(${plan.ID})">
                    <i class="fas fa-trash"></i> Delete
                </button>
            </div>
        `;
        
        return card;
    }

    async loadBillingData() {
        try {
            // Load billing stats
            const statsResponse = await fetch('/api/v1/admin/saas/admin/billing/stats', { credentials: 'include' });
            if (!statsResponse.ok) {
                throw new Error(`HTTP ${statsResponse.status}: ${statsResponse.statusText}`);
            }
            const statsData = await statsResponse.json();
            this.renderBillingStats(statsData.stats);

            // Load billing events
            const eventsResponse = await fetch('/api/v1/admin/saas/admin/billing/events', { credentials: 'include' });
            if (!eventsResponse.ok) {
                throw new Error(`HTTP ${eventsResponse.status}: ${eventsResponse.statusText}`);
            }
            const eventsData = await eventsResponse.json();
            this.renderBillingEvents(eventsData.events);

        } catch (error) {
            console.error('Error loading billing data:', error);
            const content = document.getElementById('billing-content');
            if (content) {
                content.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading billing data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadBillingData()">Retry</button>
                </div>`;
            }
        }
    }

    renderBillingStats(stats) {
        document.getElementById('totalRevenue').textContent = `$${stats.total_revenue || 0}`;
        document.getElementById('pendingPayments').textContent = `$${stats.pending_payments || 0}`;
        document.getElementById('failedPayments').textContent = `$${stats.failed_payments || 0}`;
    }

    renderBillingEvents(events) {
        const tbody = document.getElementById('billingEventsTableBody');
        if (!tbody) return;

        tbody.innerHTML = '';

        if (events && events.length > 0) {
            events.forEach(event => {
                const row = this.createBillingEventRow(event);
                tbody.appendChild(row);
            });
        } else {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-receipt"></i>
                        <h3>No Billing Events</h3>
                        <p>Billing events will appear here as payments are processed</p>
                    </td>
                </tr>
            `;
        }
    }

    createBillingEventRow(event) {
        const row = document.createElement('tr');
        
        const date = new Date(event.created_at).toLocaleDateString();
        
        row.innerHTML = `
            <td>${date}</td>
            <td>
                <div class="tenant-info">
                    <strong>${event.tenant_name}</strong>
                </div>
            </td>
            <td>
                <span class="event-type">${event.event_type.replace('_', ' ')}</span>
            </td>
            <td>$${event.amount || 0}</td>
            <td>
                <span class="status-badge ${event.status || 'success'}">${event.status || 'success'}</span>
            </td>
        `;
        
        return row;
    }

    async loadAnalyticsData() {
        try {
            const response = await this.authenticatedFetch('/api/v1/admin/saas/admin/analytics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const analytics = data.analytics;
            
            this.createUsageMetricsChart(analytics.usage_metrics);
            this.createApiUsageChart(analytics.api_usage);
            this.createPageViewsChart(analytics.page_views);
        } catch (error) {
            console.error('Error loading analytics:', error);
            const content = document.getElementById('analytics-content');
            if (content) {
                content.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading analytics data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadAnalyticsData()">Retry</button>
                </div>`;
            }
        }
    }

    createUsageMetricsChart(data) {
        const ctx = document.getElementById('usageMetricsChart');
        if (!ctx) return;

        if (this.charts.usageMetrics) {
            this.charts.usageMetrics.destroy();
        }

        // Create simple bar chart with the actual data
        this.charts.usageMetrics = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['Total Tenants', 'Active Tenants', 'Total Services', 'Total Incidents', 'Total Subscribers'],
                datasets: [{
                    label: 'Count',
                    data: [
                        data.total_tenants || 0,
                        data.active_tenants || 0,
                        data.total_services || 0,
                        data.total_incidents || 0,
                        data.total_subscribers || 0
                    ],
                    backgroundColor: [
                        '#667eea',
                        '#4facfe',
                        '#43e97b',
                        '#f093fb',
                        '#f5576c'
                    ],
                    borderColor: [
                        '#667eea',
                        '#4facfe',
                        '#43e97b',
                        '#f093fb',
                        '#f5576c'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                aspectRatio: 2,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            display: true
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                },
                layout: {
                    padding: {
                        top: 10,
                        bottom: 10,
                        left: 10,
                        right: 10
                    }
                }
            }
        });
    }

    createApiUsageChart(data) {
        const ctx = document.getElementById('apiUsageChart');
        if (!ctx) return;

        if (this.charts.apiUsage) {
            this.charts.apiUsage.destroy();
        }

        this.charts.apiUsage = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Successful Requests', 'Failed Requests'],
                datasets: [{
                    data: [
                        (data.requests || 0) - (data.errors || 0),
                        data.errors || 0
                    ],
                    backgroundColor: [
                        '#43e97b',
                        '#f5576c'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                aspectRatio: 1,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                },
                layout: {
                    padding: {
                        top: 10,
                        bottom: 10,
                        left: 10,
                        right: 10
                    }
                }
            }
        });
    }

    createPageViewsChart(data) {
        const ctx = document.getElementById('pageViewsChart');
        if (!ctx) return;

        if (this.charts.pageViews) {
            this.charts.pageViews.destroy();
        }

        this.charts.pageViews = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: ['Total Page Views', 'Unique Visitors', 'Bounce Rate %'],
                datasets: [{
                    label: 'Count',
                    data: [
                        data.total || 0,
                        data.unique || 0,
                        data.bounce_rate || 0
                    ],
                    backgroundColor: [
                        '#f093fb',
                        '#f5576c',
                        '#667eea'
                    ],
                    borderColor: [
                        '#f093fb',
                        '#f5576c',
                        '#667eea'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                aspectRatio: 1.5,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            display: true
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                },
                layout: {
                    padding: {
                        top: 10,
                        bottom: 10,
                        left: 10,
                        right: 10
                    }
                }
            }
        });
    }

    async loadSettingsData() {
        try {
            const response = await this.authenticatedFetch('/api/v1/admin/saas/admin/settings');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const settings = data.settings;
            
            this.renderGeneralSettings(settings.general);
            this.renderFeatureFlags(settings.feature_flags);
            this.renderBillingSettings(settings.billing);
            this.renderSecuritySettings(settings.security);
        } catch (error) {
            console.error('Error loading settings:', error);
            const content = document.getElementById('settings-content');
            if (content) {
                content.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading settings data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadSettingsData()">Retry</button>
                </div>`;
            }
        }
    }

    renderGeneralSettings(settings) {
        const container = document.getElementById('generalSettings');
        if (!container) return;

        container.innerHTML = '';

        Object.entries(settings).forEach(([key, value]) => {
            const settingElement = this.createSettingElement({
                key: key,
                value: value,
                type: typeof value === 'boolean' ? 'boolean' : 'text'
            });
            container.appendChild(settingElement);
        });
    }

    createSettingElement(setting) {
        const div = document.createElement('div');
        div.className = 'form-group';

        let input;
        if (setting.type === 'boolean') {
            input = `<input type="checkbox" ${setting.value === 'true' ? 'checked' : ''} onchange="saasAdmin.updateSetting('${setting.key}', this.checked)">`;
        } else if (setting.type === 'number') {
            input = `<input type="number" value="${setting.value}" onchange="saasAdmin.updateSetting('${setting.key}', this.value)">`;
        } else {
            input = `<input type="text" value="${setting.value}" onchange="saasAdmin.updateSetting('${setting.key}', this.value)">`;
        }

        div.innerHTML = `
            <label>${setting.description || setting.key}</label>
            ${input}
        `;

        return div;
    }

    renderFeatureFlags(flags) {
        const container = document.getElementById('featureFlags');
        if (!container) return;

        container.innerHTML = '';

        Object.entries(flags).forEach(([key, value]) => {
            const flagElement = this.createFeatureFlagElement({
                key: key,
                value: value,
                type: 'boolean'
            });
            container.appendChild(flagElement);
        });
    }

    createFeatureFlagElement(flag) {
        const div = document.createElement('div');
        div.className = 'feature-flag';

        div.innerHTML = `
            <div class="feature-flag-info">
                <h4>${flag.feature}</h4>
                <p>${flag.description || 'No description available'}</p>
            </div>
            <label class="feature-flag-toggle">
                <input type="checkbox" ${flag.is_enabled ? 'checked' : ''} onchange="saasAdmin.toggleFeatureFlag('${flag.feature}', this.checked)">
                <span class="feature-flag-slider"></span>
            </label>
        `;

        return div;
    }

    renderBillingSettings(settings) {
        const container = document.getElementById('billingSettings');
        if (!container) return;

        container.innerHTML = '';

        Object.entries(settings).forEach(([key, value]) => {
            const settingElement = this.createSettingElement({
                key: key,
                value: value,
                type: typeof value === 'boolean' ? 'boolean' : 'text'
            });
            container.appendChild(settingElement);
        });
    }

    renderSecuritySettings(settings) {
        const container = document.getElementById('securitySettings');
        if (!container) return;

        container.innerHTML = '';

        Object.entries(settings).forEach(([key, value]) => {
            const settingElement = this.createSettingElement({
                key: key,
                value: value,
                type: typeof value === 'boolean' ? 'boolean' : 'text'
            });
            container.appendChild(settingElement);
        });
    }

    async loadNotificationsData() {
        try {
            const response = await this.authenticatedFetch('/api/v1/admin/saas/admin/notifications');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const notifications = data.notifications || [];
            this.renderNotificationsList(notifications);
        } catch (error) {
            console.error('Error loading notifications:', error);
            const container = document.getElementById('notificationsList');
            if (container) {
                container.innerHTML = `<div class="error-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <p>Error loading notifications data. Please try again.</p>
                    <button class="btn-primary" onclick="saasAdmin.loadNotificationsData()">Retry</button>
                </div>`;
            }
        }
    }

    renderNotificationsList(notifications) {
        const container = document.getElementById('notificationsList');
        if (!container) return;

        container.innerHTML = '';

        if (notifications && notifications.length > 0) {
            notifications.forEach(notification => {
                const notificationElement = this.createNotificationElement(notification);
                container.appendChild(notificationElement);
            });
        } else {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-bell"></i>
                    <h3>No System Notifications</h3>
                    <p>Create notifications to communicate with all tenants</p>
                </div>
            `;
        }
    }

    createNotificationElement(notification) {
        const div = document.createElement('div');
        div.className = `notification-item ${notification.Type} ${notification.IsActive ? 'active' : 'inactive'}`;

        const startDate = new Date(notification.StartAt).toLocaleDateString();
        const endDate = notification.EndAt ? new Date(notification.EndAt).toLocaleDateString() : 'No end date';
        const createdDate = new Date(notification.CreatedAt).toLocaleDateString();

        div.innerHTML = `
            <div class="notification-header">
                <h4 class="notification-title">${notification.Title}</h4>
                <div class="notification-badges">
                    <span class="notification-type ${notification.Type}">${notification.Type}</span>
                    <span class="notification-status ${notification.IsActive ? 'active' : 'inactive'}">${notification.IsActive ? 'Active' : 'Inactive'}</span>
                </div>
            </div>
            <p class="notification-message">${notification.Message}</p>
            <div class="notification-meta">
                <span><i class="fas fa-calendar"></i> Start: ${startDate}</span>
                <span><i class="fas fa-calendar"></i> End: ${endDate}</span>
                <span><i class="fas fa-clock"></i> Created: ${createdDate}</span>
            </div>
            <div class="notification-actions">
                <button class="action-btn edit" onclick="saasAdmin.editNotification(${notification.ID})" title="Edit">
                    <i class="fas fa-edit"></i>
                </button>
                <button class="action-btn toggle" onclick="saasAdmin.toggleNotification(${notification.ID}, ${!notification.IsActive})" title="${notification.IsActive ? 'Deactivate' : 'Activate'}">
                    <i class="fas fa-${notification.IsActive ? 'pause' : 'play'}"></i>
                </button>
                <button class="action-btn delete" onclick="saasAdmin.deleteNotification(${notification.ID})" title="Delete">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        `;

        return div;
    }

    // Notification Management Functions
    showCreateNotificationForm() {
        this.openModal('notificationModal');
    }

    async editNotification(id) {
        try {
            const response = await this.authenticatedFetch(`/api/v1/admin/saas/admin/notifications/${id}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            this.showEditNotificationModal(data.notification);
        } catch (error) {
            console.error('Error fetching notification:', error);
            this.showNotification('Failed to fetch notification details', 'error');
        }
    }

    showEditNotificationModal(notification) {
        // Create edit modal if it doesn't exist
        this.createEditNotificationModal();
        
        // Populate the form with notification data
        document.getElementById('editNotificationTitle').value = notification.Title;
        document.getElementById('editNotificationMessage').value = notification.Message;
        document.getElementById('editNotificationType').value = notification.Type;
        document.getElementById('editNotificationStartAt').value = new Date(notification.StartAt).toISOString().slice(0, 16);
        document.getElementById('editNotificationEndAt').value = notification.EndAt ? new Date(notification.EndAt).toISOString().slice(0, 16) : '';
        document.getElementById('editNotificationModal').dataset.notificationId = notification.ID;
        
        this.openModal('editNotificationModal');
    }

    createEditNotificationModal() {
        if (document.getElementById('editNotificationModal')) return;

        const modal = document.createElement('div');
        modal.id = 'editNotificationModal';
        modal.className = 'modal';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h2>Edit System Notification</h2>
                    <span class="close-button">&times;</span>
                </div>
                <form id="editNotificationForm">
                    <div class="form-group">
                        <label for="editNotificationTitle">Title</label>
                        <input type="text" id="editNotificationTitle" name="title" required>
                    </div>
                    <div class="form-group">
                        <label for="editNotificationMessage">Message</label>
                        <textarea id="editNotificationMessage" name="message" rows="4" required></textarea>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editNotificationType">Type</label>
                            <select id="editNotificationType" name="type" required>
                                <option value="info">Info</option>
                                <option value="warning">Warning</option>
                                <option value="error">Error</option>
                                <option value="success">Success</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label for="editNotificationStartAt">Start Date</label>
                            <input type="datetime-local" id="editNotificationStartAt" name="start_at" required>
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="editNotificationEndAt">End Date (optional)</label>
                        <input type="datetime-local" id="editNotificationEndAt" name="end_at">
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn-secondary" onclick="saasAdmin.closeModal('editNotificationModal')">Cancel</button>
                        <button type="submit" class="btn-primary">Update Notification</button>
                    </div>
                </form>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Add form submit handler
        document.getElementById('editNotificationForm').addEventListener('submit', (e) => this.handleEditNotificationSubmit(e));
    }

    async handleEditNotificationSubmit(e) {
        e.preventDefault();
        const formData = new FormData(e.target);
        const notificationId = document.getElementById('editNotificationModal').dataset.notificationId;
        
        const notificationData = {
            title: formData.get('title'),
            message: formData.get('message'),
            type: formData.get('type'),
            start_at: new Date(formData.get('start_at')).toISOString(),
            end_at: formData.get('end_at') ? new Date(formData.get('end_at')).toISOString() : ''
        };

        try {
            const response = await this.authenticatedFetch(`/api/v1/admin/saas/admin/notifications/${notificationId}`, {
                method: 'PUT',
                body: JSON.stringify(notificationData)
            });

            if (response.ok) {
                this.showNotification('Notification updated successfully', 'success');
                this.closeModal('editNotificationModal');
                this.loadNotificationsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to update notification');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async toggleNotification(id, isActive) {
        try {
            const response = await this.authenticatedFetch(`/api/v1/admin/saas/admin/notifications/${id}`, {
                method: 'PUT',
                body: JSON.stringify({ is_active: isActive })
            });

            if (response.ok) {
                this.showNotification(`Notification ${isActive ? 'activated' : 'deactivated'} successfully`, 'success');
                this.loadNotificationsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to toggle notification');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async deleteNotification(id) {
        if (!confirm('Are you sure you want to delete this notification? This action cannot be undone.')) {
            return;
        }

        try {
            const response = await this.authenticatedFetch(`/api/v1/admin/saas/admin/notifications/${id}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.showNotification('Notification deleted successfully', 'success');
                this.loadNotificationsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to delete notification');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    // Modal Management
    openModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'block';
            document.body.style.overflow = 'hidden';
        }
    }

    closeModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'none';
            document.body.style.overflow = 'auto';
            
            // Reset forms
            const form = modal.querySelector('form');
            if (form) {
                form.reset();
            }
        }
    }

    // Form Handlers
    async handleTenantSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const tenantData = {
            name: formData.get('name'),
            slug: formData.get('slug'),
            contact_email: formData.get('contact_email'),
            billing_email: formData.get('billing_email'),
            subdomain: formData.get('subdomain'),
            domain: formData.get('domain'),
            plan: formData.get('plan')
        };

        try {
            const response = await fetch('/api/v1/admin/saas/admin/tenants', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(tenantData)
            });

            if (response.ok) {
                this.showNotification('Tenant created successfully', 'success');
                this.closeModal('tenantModal');
                this.loadTenantsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to create tenant');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async handlePlanSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const features = Array.from(document.querySelectorAll('input[name="features"]:checked'))
            .map(cb => cb.value);

        const planData = {
            name: formData.get('name'),
            slug: formData.get('slug'),
            description: formData.get('description'),
            price: parseFloat(formData.get('price')),
            billing_interval: formData.get('billing_interval'),
            max_services: parseInt(formData.get('max_services')),
            max_monitors: parseInt(formData.get('max_monitors')),
            max_subscribers: parseInt(formData.get('max_subscribers')),
            max_incidents: 50, // Default value
            max_maintenance: 20, // Default value
            support: formData.get('support'),
            is_active: true,
            // Convert features array to boolean fields
            custom_domain: features.includes('custom_domain'),
            white_label: features.includes('white_label'),
            api: features.includes('api_access'),
            integrations: features.includes('integrations'),
            analytics: features.includes('analytics')
        };

        try {
            const response = await fetch('/api/v1/admin/saas/admin/plans', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(planData)
            });

            if (response.ok) {
                this.showNotification('Plan created successfully', 'success');
                this.closeModal('planModal');
                this.loadPlansData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to create plan');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async handleNotificationSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const notificationData = {
            title: formData.get('title'),
            message: formData.get('message'),
            type: formData.get('type'),
            start_at: new Date(formData.get('start_at')).toISOString(),
            end_at: formData.get('end_at') ? new Date(formData.get('end_at')).toISOString() : ''
        };

        try {
            const response = await this.authenticatedFetch('/api/v1/admin/saas/admin/notifications', {
                method: 'POST',
                body: JSON.stringify(notificationData)
            });

            if (response.ok) {
                this.showNotification('Notification created successfully', 'success');
                this.closeModal('notificationModal');
                this.loadNotificationsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to create notification');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    // Utility Methods
    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        
        // Style the notification
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 6px;
            color: white;
            font-weight: 500;
            z-index: 3000;
            transform: translateX(100%);
            transition: transform 0.3s ease;
        `;
        
        // Set background color based on type
        const colors = {
            success: '#36b37e',
            error: '#ff5630',
            warning: '#ffab00',
            info: '#0052cc'
        };
        notification.style.backgroundColor = colors[type] || colors.info;
        
        document.body.appendChild(notification);
        
        // Animate in
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
        }, 100);
        
        // Remove after 3 seconds
        setTimeout(() => {
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    getTimeAgo(date) {
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} minutes ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
        return `${Math.floor(diffInSeconds / 86400)} days ago`;
    }

    // Action Methods (to be implemented)
    async editTenant(id) {
        console.log('Edit tenant:', id);
        
        try {
            // Fetch the tenant data
            const response = await fetch(`/api/v1/admin/saas/admin/tenants/${id}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const tenant = data.tenant;
            
            // Show the edit modal with pre-populated data
            this.showEditTenantModal(tenant);
            
        } catch (error) {
            console.error('Error fetching tenant:', error);
            this.showNotification('Failed to load tenant data: ' + error.message, 'error');
        }
    }

    showEditTenantModal(tenant) {
        // Create edit modal if it doesn't exist
        let editModal = document.getElementById('editTenantModal');
        if (!editModal) {
            editModal = this.createEditTenantModal();
            document.body.appendChild(editModal);
        }

        // Populate the form with tenant data
        document.getElementById('editTenantName').value = tenant.Name || '';
        document.getElementById('editTenantSlug').value = tenant.Slug || '';
        document.getElementById('editTenantDomain').value = tenant.Domain || '';
        document.getElementById('editTenantSubdomain').value = tenant.Subdomain || '';
        document.getElementById('editTenantBillingEmail').value = tenant.BillingEmail || '';
        document.getElementById('editTenantContactEmail').value = tenant.ContactEmail || '';
        document.getElementById('editTenantPlan').value = tenant.Plan || 'free';
        document.getElementById('editTenantStatus').value = tenant.Status || 'active';

        // Store the tenant ID for the form submission
        editModal.dataset.tenantId = tenant.ID;

        // Show the modal
        this.openModal('editTenantModal');
    }

    createEditTenantModal() {
        const modal = document.createElement('div');
        modal.id = 'editTenantModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal-content large">
                <div class="modal-header">
                    <h2>Edit Tenant</h2>
                    <span class="close-button">&times;</span>
                </div>
                <form id="editTenantForm">
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editTenantName">Tenant Name</label>
                            <input type="text" id="editTenantName" name="name" required>
                        </div>
                        <div class="form-group">
                            <label for="editTenantSlug">Slug</label>
                            <input type="text" id="editTenantSlug" name="slug" required>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editTenantDomain">Domain</label>
                            <input type="text" id="editTenantDomain" name="domain" placeholder="example.com">
                        </div>
                        <div class="form-group">
                            <label for="editTenantSubdomain">Subdomain</label>
                            <input type="text" id="editTenantSubdomain" name="subdomain" required>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editTenantBillingEmail">Billing Email</label>
                            <input type="email" id="editTenantBillingEmail" name="billing_email">
                        </div>
                        <div class="form-group">
                            <label for="editTenantContactEmail">Contact Email</label>
                            <input type="email" id="editTenantContactEmail" name="contact_email">
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editTenantPlan">Plan</label>
                            <select id="editTenantPlan" name="plan" required>
                                <option value="free">Free</option>
                                <option value="pro">Pro</option>
                                <option value="enterprise">Enterprise</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label for="editTenantStatus">Status</label>
                            <select id="editTenantStatus" name="status" required>
                                <option value="active">Active</option>
                                <option value="suspended">Suspended</option>
                                <option value="cancelled">Cancelled</option>
                            </select>
                        </div>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn-secondary" id="cancelEditTenant">Cancel</button>
                        <button type="submit" class="btn-primary">Update Tenant</button>
                    </div>
                </form>
            </div>
        `;

        // Add event listeners
        modal.querySelector('#cancelEditTenant').addEventListener('click', () => {
            this.closeModal('editTenantModal');
        });

        modal.querySelector('#editTenantForm').addEventListener('submit', (e) => {
            this.handleEditTenantSubmit(e);
        });

        // Add modal close functionality
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal('editTenantModal');
            }
        });

        modal.querySelector('.close-button').addEventListener('click', () => {
            this.closeModal('editTenantModal');
        });

        return modal;
    }

    async handleEditTenantSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const tenantId = document.getElementById('editTenantModal').dataset.tenantId;
        
        const tenantData = {
            name: formData.get('name'),
            slug: formData.get('slug'),
            domain: formData.get('domain'),
            subdomain: formData.get('subdomain'),
            billing_email: formData.get('billing_email'),
            contact_email: formData.get('contact_email'),
            plan: formData.get('plan'),
            status: formData.get('status')
        };

        try {
            const response = await fetch(`/api/v1/admin/saas/admin/tenants/${tenantId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(tenantData)
            });

            if (response.ok) {
                this.showNotification('Tenant updated successfully', 'success');
                this.closeModal('editTenantModal');
                this.loadTenantsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to update tenant');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    suspendTenant(id) {
        console.log('Suspend tenant:', id);
        // Implementation for suspending tenant
    }

    async deleteTenant(id) {
        if (!confirm('Are you sure you want to delete this tenant? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/saas/admin/tenants/${id}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Tenant deleted successfully', 'success');
                this.loadTenantsData(); // Reload the tenants list
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to delete tenant');
            }
        } catch (error) {
            console.error('Error deleting tenant:', error);
            this.showNotification('Failed to delete tenant: ' + error.message, 'error');
        }
    }

    async editSubscription(id) {
        console.log('Edit subscription:', id);
        
        try {
            // Fetch the subscription data
            const response = await fetch(`/api/v1/admin/saas/admin/subscriptions/${id}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const subscription = data.subscription;
            
            // Show the edit modal with pre-populated data
            this.showEditSubscriptionModal(subscription);
            
        } catch (error) {
            console.error('Error fetching subscription:', error);
            this.showNotification('Failed to load subscription data: ' + error.message, 'error');
        }
    }

    showEditSubscriptionModal(subscription) {
        // Create edit modal if it doesn't exist
        let editModal = document.getElementById('editSubscriptionModal');
        if (!editModal) {
            editModal = this.createEditSubscriptionModal();
            document.body.appendChild(editModal);
        }

        // Populate the form with subscription data
        document.getElementById('editSubscriptionStatus').value = subscription.Status || 'active';
        document.getElementById('editSubscriptionCancelAtPeriodEnd').checked = subscription.CancelAtPeriodEnd || false;

        // Store the subscription ID for the form submission
        editModal.dataset.subscriptionId = subscription.ID;

        // Show the modal
        this.openModal('editSubscriptionModal');
    }

    createEditSubscriptionModal() {
        const modal = document.createElement('div');
        modal.id = 'editSubscriptionModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h2>Edit Subscription</h2>
                    <span class="close-button">&times;</span>
                </div>
                <form id="editSubscriptionForm">
                    <div class="form-group">
                        <label for="editSubscriptionStatus">Status</label>
                        <select id="editSubscriptionStatus" name="status" required>
                            <option value="active">Active</option>
                            <option value="cancelled">Cancelled</option>
                            <option value="past_due">Past Due</option>
                            <option value="trialing">Trialing</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="editSubscriptionCancelAtPeriodEnd" name="cancel_at_period_end">
                            Cancel at period end
                        </label>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn-secondary" id="cancelEditSubscription">Cancel</button>
                        <button type="submit" class="btn-primary">Update Subscription</button>
                    </div>
                </form>
            </div>
        `;

        // Add event listeners
        modal.querySelector('#cancelEditSubscription').addEventListener('click', () => {
            this.closeModal('editSubscriptionModal');
        });

        modal.querySelector('#editSubscriptionForm').addEventListener('submit', (e) => {
            this.handleEditSubscriptionSubmit(e);
        });

        // Add modal close functionality
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal('editSubscriptionModal');
            }
        });

        modal.querySelector('.close-button').addEventListener('click', () => {
            this.closeModal('editSubscriptionModal');
        });

        return modal;
    }

    async handleEditSubscriptionSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const subscriptionId = document.getElementById('editSubscriptionModal').dataset.subscriptionId;
        
        const subscriptionData = {
            status: formData.get('status'),
            cancel_at_period_end: formData.get('cancel_at_period_end') === 'on'
        };

        try {
            const response = await fetch(`/api/v1/admin/saas/admin/subscriptions/${subscriptionId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(subscriptionData)
            });

            if (response.ok) {
                this.showNotification('Subscription updated successfully', 'success');
                this.closeModal('editSubscriptionModal');
                this.loadSubscriptionsData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to update subscription');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async cancelSubscription(id) {
        if (!confirm('Are you sure you want to cancel this subscription? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/saas/admin/subscriptions/${id}/cancel`, {
                method: 'POST',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Subscription cancelled successfully', 'success');
                this.loadSubscriptionsData(); // Reload the subscriptions list
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to cancel subscription');
            }
        } catch (error) {
            console.error('Error cancelling subscription:', error);
            this.showNotification('Failed to cancel subscription: ' + error.message, 'error');
        }
    }

    async editPlan(id) {
        console.log('Edit plan:', id);
        
        try {
            // Fetch the plan data
            const response = await fetch(`/api/v1/admin/saas/admin/plans/${id}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const plan = data.plan;
            
            // Show the edit modal with pre-populated data
            this.showEditPlanModal(plan);
            
        } catch (error) {
            console.error('Error fetching plan:', error);
            this.showNotification('Failed to load plan data: ' + error.message, 'error');
        }
    }

    showEditPlanModal(plan) {
        // Create edit modal if it doesn't exist
        let editModal = document.getElementById('editPlanModal');
        if (!editModal) {
            editModal = this.createEditPlanModal();
            document.body.appendChild(editModal);
        }

        // Populate the form with plan data
        document.getElementById('editPlanName').value = plan.Name || '';
        document.getElementById('editPlanSlug').value = plan.Slug || '';
        document.getElementById('editPlanDescription').value = plan.Description || '';
        document.getElementById('editPlanPrice').value = plan.Price || 0;
        document.getElementById('editPlanBillingInterval').value = plan.BillingInterval || 'monthly';
        document.getElementById('editPlanMaxServices').value = plan.MaxServices || 5;
        document.getElementById('editPlanMaxMonitors').value = plan.MaxMonitors || 10;
        document.getElementById('editPlanMaxSubscribers').value = plan.MaxSubscribers || 100;
        document.getElementById('editPlanSupport').value = plan.Support || 'email';
        
        // Set feature checkboxes
        document.getElementById('editPlanCustomDomain').checked = plan.CustomDomain || false;
        document.getElementById('editPlanWhiteLabel').checked = plan.WhiteLabel || false;
        document.getElementById('editPlanAPI').checked = plan.API || false;
        document.getElementById('editPlanIntegrations').checked = plan.Integrations || false;
        document.getElementById('editPlanAnalytics').checked = plan.Analytics || false;

        // Store the plan ID for the form submission
        editModal.dataset.planId = plan.ID;

        // Show the modal
        this.openModal('editPlanModal');
    }

    createEditPlanModal() {
        const modal = document.createElement('div');
        modal.id = 'editPlanModal';
        modal.className = 'modal';
        
        modal.innerHTML = `
            <div class="modal-content large">
                <div class="modal-header">
                    <h2>Edit Subscription Plan</h2>
                    <span class="close-button">&times;</span>
                </div>
                <form id="editPlanForm">
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editPlanName">Plan Name</label>
                            <input type="text" id="editPlanName" name="name" required>
                        </div>
                        <div class="form-group">
                            <label for="editPlanSlug">Slug</label>
                            <input type="text" id="editPlanSlug" name="slug" required>
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="editPlanDescription">Description</label>
                        <textarea id="editPlanDescription" name="description" rows="3"></textarea>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editPlanPrice">Price (USD)</label>
                            <input type="number" id="editPlanPrice" name="price" step="0.01" min="0" required>
                        </div>
                        <div class="form-group">
                            <label for="editPlanBillingInterval">Billing Interval</label>
                            <select id="editPlanBillingInterval" name="billing_interval" required>
                                <option value="monthly">Monthly</option>
                                <option value="yearly">Yearly</option>
                            </select>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editPlanMaxServices">Max Services</label>
                            <input type="number" id="editPlanMaxServices" name="max_services" min="-1" value="5">
                        </div>
                        <div class="form-group">
                            <label for="editPlanMaxMonitors">Max Monitors</label>
                            <input type="number" id="editPlanMaxMonitors" name="max_monitors" min="-1" value="10">
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label for="editPlanMaxSubscribers">Max Subscribers</label>
                            <input type="number" id="editPlanMaxSubscribers" name="max_subscribers" min="-1" value="100">
                        </div>
                        <div class="form-group">
                            <label for="editPlanSupport">Support Level</label>
                            <select id="editPlanSupport" name="support" required>
                                <option value="email">Email</option>
                                <option value="chat">Chat</option>
                                <option value="phone">Phone</option>
                            </select>
                        </div>
                    </div>
                    <div class="form-group">
                        <label>Features</label>
                        <div class="checkbox-group">
                            <label><input type="checkbox" id="editPlanCustomDomain" name="custom_domain"> Custom Domain</label>
                            <label><input type="checkbox" id="editPlanWhiteLabel" name="white_label"> White Label</label>
                            <label><input type="checkbox" id="editPlanAPI" name="api"> API Access</label>
                            <label><input type="checkbox" id="editPlanIntegrations" name="integrations"> Integrations</label>
                            <label><input type="checkbox" id="editPlanAnalytics" name="analytics"> Analytics</label>
                        </div>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn-secondary" id="cancelEditPlan">Cancel</button>
                        <button type="submit" class="btn-primary">Update Plan</button>
                    </div>
                </form>
            </div>
        `;

        // Add event listeners
        modal.querySelector('#cancelEditPlan').addEventListener('click', () => {
            this.closeModal('editPlanModal');
        });

        modal.querySelector('#editPlanForm').addEventListener('submit', (e) => {
            this.handleEditPlanSubmit(e);
        });

        // Add modal close functionality
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal('editPlanModal');
            }
        });

        modal.querySelector('.close-button').addEventListener('click', () => {
            this.closeModal('editPlanModal');
        });

        return modal;
    }

    async handleEditPlanSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const planId = document.getElementById('editPlanModal').dataset.planId;
        
        const planData = {
            name: formData.get('name'),
            slug: formData.get('slug'),
            description: formData.get('description'),
            price: parseFloat(formData.get('price')),
            billing_interval: formData.get('billing_interval'),
            max_services: parseInt(formData.get('max_services')),
            max_monitors: parseInt(formData.get('max_monitors')),
            max_subscribers: parseInt(formData.get('max_subscribers')),
            max_incidents: 50, // Default value
            max_maintenance: 20, // Default value
            support: formData.get('support'),
            is_active: true,
            // Convert features to boolean fields
            custom_domain: formData.get('custom_domain') === 'on',
            white_label: formData.get('white_label') === 'on',
            api: formData.get('api') === 'on',
            integrations: formData.get('integrations') === 'on',
            analytics: formData.get('analytics') === 'on'
        };

        try {
            const response = await fetch(`/api/v1/admin/saas/admin/plans/${planId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(planData)
            });

            if (response.ok) {
                this.showNotification('Plan updated successfully', 'success');
                this.closeModal('editPlanModal');
                this.loadPlansData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to update plan');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async duplicatePlan(id) {
        console.log('Duplicate plan:', id);
        
        try {
            // Fetch the plan data
            const response = await fetch(`/api/v1/admin/saas/admin/plans/${id}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const plan = data.plan;
            
            // Create a duplicate with modified name and slug
            const duplicateData = {
                name: `${plan.Name} (Copy)`,
                slug: `${plan.Slug}-copy-${Date.now()}`,
                description: plan.Description,
                price: parseFloat(plan.Price) || 0, // Ensure price is a valid number
                currency: plan.Currency || 'USD',
                billing_interval: plan.BillingInterval,
                max_services: plan.MaxServices,
                max_monitors: plan.MaxMonitors,
                max_subscribers: plan.MaxSubscribers,
                max_incidents: plan.MaxIncidents,
                max_maintenance: plan.MaxMaintenance,
                custom_domain: plan.CustomDomain,
                white_label: plan.WhiteLabel,
                api: plan.API,
                integrations: plan.Integrations,
                analytics: plan.Analytics,
                support: plan.Support,
                is_active: plan.IsActive
            };

            // Create the duplicate plan
            const createResponse = await fetch('/api/v1/admin/saas/admin/plans', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(duplicateData)
            });

            if (createResponse.ok) {
                this.showNotification('Plan duplicated successfully', 'success');
                this.loadPlansData(); // Reload the plans list
            } else {
                const error = await createResponse.json();
                throw new Error(error.error || 'Failed to duplicate plan');
            }
            
        } catch (error) {
            console.error('Error duplicating plan:', error);
            this.showNotification('Failed to duplicate plan: ' + error.message, 'error');
        }
    }

    async deletePlan(id) {
        if (!confirm('Are you sure you want to delete this plan? This action cannot be undone.')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/admin/saas/admin/plans/${id}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            if (response.ok) {
                this.showNotification('Plan deleted successfully', 'success');
                this.loadPlansData(); // Reload the plans
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to delete plan');
            }
        } catch (error) {
            console.error('Error deleting plan:', error);
            this.showNotification(error.message, 'error');
        }
    }

    updateSetting(key, value) {
        console.log('Update setting:', key, value);
        // Implementation for updating setting
    }

    toggleFeatureFlag(feature, enabled) {
        console.log('Toggle feature flag:', feature, enabled);
        // Implementation for toggling feature flag
    }
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.saasAdmin = new SaaSAdminDashboard();
});

// Export for global access
window.SaaSAdminDashboard = SaaSAdminDashboard;
