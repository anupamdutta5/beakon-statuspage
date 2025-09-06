// SaaS Admin Dashboard JavaScript

class SaaSAdminDashboard {
    constructor() {
        this.currentSection = 'overview';
        this.charts = {};
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadOverviewData();
        this.setupModals();
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
            const response = await fetch('/api/v1/saas/admin/metrics', { credentials: 'include' });
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
            const response = await fetch('/api/v1/saas/admin/activity', { credentials: 'include' });
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
            const response = await fetch('/api/v1/saas/admin/plan-distribution', { credentials: 'include' });
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
            const response = await fetch('/api/v1/saas/admin/charts/overview', { credentials: 'include' });
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
        // Mock tenants data
        const mockTenants = [
            { id: 1, name: "Acme Corp", domain: "acme.statuspage.com", status: "active", created_at: "2025-01-01T00:00:00Z" },
            { id: 2, name: "TechStart Inc", domain: "techstart.statuspage.com", status: "active", created_at: "2025-01-02T00:00:00Z" }
        ];

        this.renderTenantsTable(mockTenants);
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
        
        const createdAt = new Date(tenant.created_at).toLocaleDateString();
        
        row.innerHTML = `
            <td>
                <div class="tenant-info">
                    <strong>${tenant.name}</strong>
                    <small>${tenant.contact_email}</small>
                </div>
            </td>
            <td>
                <code>${tenant.slug}</code>
            </td>
            <td>
                <span class="plan-badge ${tenant.plan}">${tenant.plan}</span>
            </td>
            <td>
                <span class="status-badge ${tenant.status}">${tenant.status}</span>
            </td>
            <td>${createdAt}</td>
            <td>
                <div class="action-buttons">
                    <button class="action-btn edit" onclick="saasAdmin.editTenant(${tenant.id})">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn suspend" onclick="saasAdmin.suspendTenant(${tenant.id})">
                        <i class="fas fa-ban"></i>
                    </button>
                    <button class="action-btn delete" onclick="saasAdmin.deleteTenant(${tenant.id})">
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
        // Mock subscriptions data
        const mockSubscriptions = [
            { id: 1, tenant_name: "Acme Corp", plan_name: "Pro", status: "active", created_at: "2025-01-01T00:00:00Z" },
            { id: 2, tenant_name: "TechStart Inc", plan_name: "Basic", status: "active", created_at: "2025-01-02T00:00:00Z" }
        ];

        this.renderSubscriptionsTable(mockSubscriptions);
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
        
        const nextBilling = subscription.current_period_end ? 
            new Date(subscription.current_period_end).toLocaleDateString() : 'N/A';
        
        row.innerHTML = `
            <td>
                <div class="tenant-info">
                    <strong>${subscription.tenant_name}</strong>
                    <small>${subscription.tenant_slug}</small>
                </div>
            </td>
            <td>
                <span class="plan-badge ${subscription.plan_slug}">${subscription.plan_name}</span>
            </td>
            <td>
                <span class="status-badge ${subscription.status}">${subscription.status}</span>
            </td>
            <td>$${subscription.plan_price}</td>
            <td>${nextBilling}</td>
            <td>
                <div class="action-buttons">
                    <button class="action-btn edit" onclick="saasAdmin.editSubscription(${subscription.id})">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn suspend" onclick="saasAdmin.cancelSubscription(${subscription.id})">
                        <i class="fas fa-ban"></i>
                    </button>
                </div>
            </td>
        `;
        
        return row;
    }

    async loadPlansData() {
        // Mock plans data
        const mockPlans = [
            { id: 1, name: "Basic", price: 9.99, features: ["1 Status Page", "Email Notifications"], status: "active" },
            { id: 2, name: "Pro", price: 29.99, features: ["5 Status Pages", "Email & SMS", "Custom Domain"], status: "active" },
            { id: 3, name: "Enterprise", price: 99.99, features: ["Unlimited Pages", "All Notifications", "White Label"], status: "active" }
        ];

        this.renderPlansGrid(mockPlans);
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
        
        if (plan.slug === 'pro') {
            card.classList.add('popular');
        }
        
        const features = JSON.parse(plan.features || '[]');
        const featuresList = features.map(feature => 
            `<li><i class="fas fa-check"></i> ${feature.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}</li>`
        ).join('');
        
        card.innerHTML = `
            <div class="plan-header">
                <h3 class="plan-name">${plan.name}</h3>
                <div class="plan-price">
                    <span class="currency">$</span>${plan.price}
                    <span class="period">/${plan.billing_interval}</span>
                </div>
                <p class="plan-description">${plan.description}</p>
            </div>
            
            <div class="plan-limits">
                <h4>Limits</h4>
                <div class="limit">
                    <span>Services</span>
                    <span>${plan.max_services === -1 ? 'Unlimited' : plan.max_services}</span>
                </div>
                <div class="limit">
                    <span>Monitors</span>
                    <span>${plan.max_monitors === -1 ? 'Unlimited' : plan.max_monitors}</span>
                </div>
                <div class="limit">
                    <span>Subscribers</span>
                    <span>${plan.max_subscribers === -1 ? 'Unlimited' : plan.max_subscribers}</span>
                </div>
            </div>
            
            <ul class="plan-features">
                ${featuresList}
            </ul>
            
            <div class="plan-actions">
                <button class="btn-secondary" onclick="saasAdmin.editPlan(${plan.id})">
                    <i class="fas fa-edit"></i> Edit Plan
                </button>
                <button class="btn-primary" onclick="saasAdmin.duplicatePlan(${plan.id})">
                    <i class="fas fa-copy"></i> Duplicate
                </button>
            </div>
        `;
        
        return card;
    }

    async loadBillingData() {
        // Mock billing data
        const mockStats = {
            totalRevenue: 1250.00,
            monthlyRecurringRevenue: 1250.00,
            activeSubscriptions: 15,
            churnRate: 2.5
        };
        
        const mockEvents = [
            { id: 1, tenant_name: "Acme Corp", amount: 29.99, type: "subscription", created_at: "2025-01-07T10:00:00Z" },
            { id: 2, tenant_name: "TechStart Inc", amount: 9.99, type: "subscription", created_at: "2025-01-07T09:00:00Z" }
        ];

        this.renderBillingStats(mockStats);
        this.renderBillingEvents(mockEvents);
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
        // Mock analytics data
        const mockAnalytics = {
            usage_metrics: { total_tenants: 15, active_tenants: 12, total_revenue: 1250.00 },
            api_usage: { requests: 15000, errors: 25, avg_response_time: 120 },
            page_views: { total: 50000, unique: 2500, bounce_rate: 35.5 }
        };

        this.createUsageMetricsChart(mockAnalytics.usage_metrics);
        this.createApiUsageChart(mockAnalytics.api_usage);
        this.createPageViewsChart(mockAnalytics.page_views);
    }

    createUsageMetricsChart(data) {
        const ctx = document.getElementById('usageMetricsChart');
        if (!ctx) return;

        if (this.charts.usageMetrics) {
            this.charts.usageMetrics.destroy();
        }

        this.charts.usageMetrics = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels || [],
                datasets: [
                    {
                        label: 'Services',
                        data: data.services || [],
                        borderColor: '#667eea',
                        backgroundColor: 'rgba(102, 126, 234, 0.1)',
                        tension: 0.4
                    },
                    {
                        label: 'Monitors',
                        data: data.monitors || [],
                        borderColor: '#4facfe',
                        backgroundColor: 'rgba(79, 172, 254, 0.1)',
                        tension: 0.4
                    },
                    {
                        label: 'Subscribers',
                        data: data.subscribers || [],
                        borderColor: '#43e97b',
                        backgroundColor: 'rgba(67, 233, 123, 0.1)',
                        tension: 0.4
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'top'
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

    createApiUsageChart(data) {
        const ctx = document.getElementById('apiUsageChart');
        if (!ctx) return;

        if (this.charts.apiUsage) {
            this.charts.apiUsage.destroy();
        }

        this.charts.apiUsage = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: data.labels || [],
                datasets: [{
                    data: data.values || [],
                    backgroundColor: [
                        '#667eea',
                        '#4facfe',
                        '#43e97b',
                        '#f093fb',
                        '#f5576c'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
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
                labels: data.labels || [],
                datasets: [{
                    label: 'Page Views',
                    data: data.values || [],
                    backgroundColor: '#f093fb',
                    borderColor: '#f5576c',
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

    async loadSettingsData() {
        // Mock settings data
        const mockSettings = {
            general: {
                site_name: "StatusPage SaaS",
                contact_email: "admin@statuspage.com",
                max_tenants: 1000
            },
            feature_flags: {
                enable_white_label: true,
                enable_custom_domains: true,
                enable_api_access: true
            },
            billing: {
                stripe_public_key: "pk_test_...",
                stripe_secret_key: "sk_test_...",
                default_currency: "USD"
            }
        };

        this.renderGeneralSettings(mockSettings.general);
        this.renderFeatureFlags(mockSettings.feature_flags);
        this.renderBillingSettings(mockSettings.billing);
    }

    renderGeneralSettings(settings) {
        const container = document.getElementById('generalSettings');
        if (!container) return;

        container.innerHTML = '';

        settings.forEach(setting => {
            const settingElement = this.createSettingElement(setting);
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

        flags.forEach(flag => {
            const flagElement = this.createFeatureFlagElement(flag);
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

        settings.forEach(setting => {
            const settingElement = this.createSettingElement(setting);
            container.appendChild(settingElement);
        });
    }

    async loadNotificationsData() {
        // Mock notifications data
        const mockNotifications = [
            { id: 1, type: "email", status: "active", created_at: "2025-01-01T00:00:00Z" },
            { id: 2, type: "sms", status: "active", created_at: "2025-01-02T00:00:00Z" }
        ];

        this.renderNotificationsList(mockNotifications);
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
        div.className = `notification-item ${notification.type}`;

        const startDate = new Date(notification.start_at).toLocaleDateString();
        const endDate = notification.end_at ? new Date(notification.end_at).toLocaleDateString() : 'No end date';

        div.innerHTML = `
            <div class="notification-header">
                <h4 class="notification-title">${notification.title}</h4>
                <span class="notification-type ${notification.type}">${notification.type}</span>
            </div>
            <p class="notification-message">${notification.message}</p>
            <div class="notification-meta">
                <span>Start: ${startDate}</span>
                <span>End: ${endDate}</span>
                <span>Status: ${notification.is_active ? 'Active' : 'Inactive'}</span>
            </div>
        `;

        return div;
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
            const response = await fetch('/api/v1/saas/admin/tenants', {
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
            support: formData.get('support'),
            features: JSON.stringify(features)
        };

        try {
            const response = await fetch('/api/v1/saas/admin/plans', {
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
            start_at: formData.get('start_at'),
            end_at: formData.get('end_at') || null
        };

        try {
            const response = await fetch('/api/v1/saas/admin/notifications', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
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
    editTenant(id) {
        console.log('Edit tenant:', id);
        // Implementation for editing tenant
    }

    suspendTenant(id) {
        console.log('Suspend tenant:', id);
        // Implementation for suspending tenant
    }

    deleteTenant(id) {
        if (confirm('Are you sure you want to delete this tenant? This action cannot be undone.')) {
            console.log('Delete tenant:', id);
            // Implementation for deleting tenant
        }
    }

    editSubscription(id) {
        console.log('Edit subscription:', id);
        // Implementation for editing subscription
    }

    cancelSubscription(id) {
        if (confirm('Are you sure you want to cancel this subscription?')) {
            console.log('Cancel subscription:', id);
            // Implementation for cancelling subscription
        }
    }

    editPlan(id) {
        console.log('Edit plan:', id);
        // Implementation for editing plan
    }

    duplicatePlan(id) {
        console.log('Duplicate plan:', id);
        // Implementation for duplicating plan
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
