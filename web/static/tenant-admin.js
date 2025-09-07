class TenantAdmin {
    constructor() {
        this.currentSection = 'overview';
        this.tenant = null;
        this.init();
    }

    async init() {
        this.setupEventListeners();
        await this.loadTenantInfo();
        await this.loadOverviewData();
    }

    setupEventListeners() {
        // Sidebar navigation
        document.querySelectorAll('.sidebar-menu a').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = e.target.getAttribute('href').substring(1);
                this.showSection(section);
            });
        });

        // Dropdown menu
        const dropdownBtn = document.querySelector('.dropdown-btn');
        const dropdownContent = document.querySelector('.dropdown-content');
        
        dropdownBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            dropdownContent.style.display = dropdownContent.style.display === 'block' ? 'none' : 'block';
        });

        document.addEventListener('click', () => {
            dropdownContent.style.display = 'none';
        });

        // Form submissions
        document.getElementById('branding-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleBrandingSubmit(e);
        });

        document.getElementById('settings-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleSettingsSubmit(e);
        });
    }

    async loadTenantInfo() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/info');
            if (response.ok) {
                const data = await response.json();
                this.tenant = data.tenant;
                document.getElementById('tenant-name').textContent = this.tenant.name;
            }
        } catch (error) {
            console.error('Failed to load tenant info:', error);
        }
    }

    async loadOverviewData() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/overview');
            if (response.ok) {
                const data = await response.json();
                this.updateOverviewStats(data);
                this.createOverviewCharts(data);
            }
        } catch (error) {
            console.error('Failed to load overview data:', error);
        }
    }

    updateOverviewStats(data) {
        document.getElementById('total-services').textContent = data.stats.total_services || 0;
        document.getElementById('active-incidents').textContent = data.stats.active_incidents || 0;
        document.getElementById('total-subscribers').textContent = data.stats.total_subscribers || 0;
        document.getElementById('uptime-percentage').textContent = `${data.stats.uptime_percentage || 0}%`;
    }

    createOverviewCharts(data) {
        // Service Status Chart
        const serviceStatusCtx = document.getElementById('service-status-chart').getContext('2d');
        new Chart(serviceStatusCtx, {
            type: 'doughnut',
            data: {
                labels: ['Operational', 'Degraded', 'Outage'],
                datasets: [{
                    data: [
                        data.stats.operational_services || 0,
                        data.stats.degraded_services || 0,
                        data.stats.outage_services || 0
                    ],
                    backgroundColor: ['#2ecc71', '#f39c12', '#e74c3c'],
                    borderWidth: 0
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

        // Incidents Timeline Chart
        const incidentsCtx = document.getElementById('incidents-chart').getContext('2d');
        new Chart(incidentsCtx, {
            type: 'line',
            data: {
                labels: data.incidents_timeline?.labels || [],
                datasets: [{
                    label: 'Incidents',
                    data: data.incidents_timeline?.data || [],
                    borderColor: '#e74c3c',
                    backgroundColor: 'rgba(231, 76, 60, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    showSection(sectionName) {
        // Update active menu item
        document.querySelectorAll('.sidebar-menu a').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`a[href="#${sectionName}"]`).classList.add('active');

        // Update page title
        const titles = {
            overview: 'Overview',
            services: 'Services',
            incidents: 'Incidents',
            maintenance: 'Maintenance',
            subscribers: 'Subscribers',
            monitors: 'Monitors',
            analytics: 'Analytics',
            branding: 'Branding',
            settings: 'Settings',
            domains: 'Domains',
            security: 'Security',
            billing: 'Billing'
        };
        document.getElementById('page-title').textContent = titles[sectionName] || 'Dashboard';

        // Show/hide sections
        document.querySelectorAll('.section').forEach(section => {
            section.classList.remove('active');
        });
        document.getElementById(`${sectionName}-section`).classList.add('active');

        this.currentSection = sectionName;

        // Load section-specific data
        this.loadSectionData(sectionName);
    }

    async loadSectionData(sectionName) {
        switch (sectionName) {
            case 'services':
                await this.loadServices();
                break;
            case 'incidents':
                await this.loadIncidents();
                break;
            case 'maintenance':
                await this.loadMaintenance();
                break;
            case 'subscribers':
                await this.loadSubscribers();
                break;
            case 'monitors':
                await this.loadMonitors();
                break;
            case 'analytics':
                await this.loadAnalytics();
                break;
            case 'branding':
                await this.loadBranding();
                break;
            case 'settings':
                await this.loadSettings();
                break;
            case 'domains':
                await this.loadDomains();
                break;
            case 'security':
                await this.loadSecurity();
                break;
            case 'billing':
                await this.loadBilling();
                break;
        }
    }

    async loadServices() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/services');
            if (response.ok) {
                const data = await response.json();
                this.renderServices(data.services);
            }
        } catch (error) {
            console.error('Failed to load services:', error);
        }
    }

    renderServices(services) {
        const container = document.getElementById('services-list');
        if (services.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-server"></i>
                    <h3>No Services</h3>
                    <p>Add your first service to start monitoring</p>
                    <button class="btn btn-primary" onclick="showCreateServiceModal()">
                        <i class="fas fa-plus"></i> Add Service
                    </button>
                </div>
            `;
            return;
        }

        container.innerHTML = services.map(service => `
            <div class="list-item">
                <div class="list-item-content">
                    <h4>${service.name}</h4>
                    <p>${service.description || 'No description'}</p>
                    <span class="status-badge ${service.status}">${service.status}</span>
                </div>
                <div class="list-item-actions">
                    <button class="action-btn edit" onclick="editService(${service.id})" title="Edit">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn delete" onclick="deleteService(${service.id})" title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async loadIncidents() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/incidents');
            if (response.ok) {
                const data = await response.json();
                this.renderIncidents(data.incidents);
            }
        } catch (error) {
            console.error('Failed to load incidents:', error);
        }
    }

    renderIncidents(incidents) {
        const container = document.getElementById('incidents-list');
        if (incidents.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h3>No Incidents</h3>
                    <p>No incidents reported</p>
                </div>
            `;
            return;
        }

        container.innerHTML = incidents.map(incident => `
            <div class="list-item">
                <div class="list-item-content">
                    <h4>${incident.title}</h4>
                    <p>${incident.description}</p>
                    <span class="status-badge ${incident.status}">${incident.status}</span>
                </div>
                <div class="list-item-actions">
                    <button class="action-btn edit" onclick="editIncident(${incident.id})" title="Edit">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn delete" onclick="deleteIncident(${incident.id})" title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async loadMaintenance() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/maintenance');
            if (response.ok) {
                const data = await response.json();
                this.renderMaintenance(data.maintenance);
            }
        } catch (error) {
            console.error('Failed to load maintenance:', error);
        }
    }

    renderMaintenance(maintenance) {
        const container = document.getElementById('maintenance-list');
        if (maintenance.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-tools"></i>
                    <h3>No Maintenance</h3>
                    <p>No scheduled maintenance</p>
                </div>
            `;
            return;
        }

        container.innerHTML = maintenance.map(item => `
            <div class="list-item">
                <div class="list-item-content">
                    <h4>${item.title}</h4>
                    <p>${item.description}</p>
                    <p><strong>Scheduled:</strong> ${new Date(item.scheduled_start).toLocaleString()}</p>
                </div>
                <div class="list-item-actions">
                    <button class="action-btn edit" onclick="editMaintenance(${item.id})" title="Edit">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn delete" onclick="deleteMaintenance(${item.id})" title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async loadSubscribers() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/subscribers');
            if (response.ok) {
                const data = await response.json();
                this.renderSubscribers(data.subscribers);
            }
        } catch (error) {
            console.error('Failed to load subscribers:', error);
        }
    }

    renderSubscribers(subscribers) {
        const container = document.getElementById('subscribers-list');
        if (subscribers.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-users"></i>
                    <h3>No Subscribers</h3>
                    <p>No email subscribers yet</p>
                </div>
            `;
            return;
        }

        container.innerHTML = subscribers.map(subscriber => `
            <div class="list-item">
                <div class="list-item-content">
                    <h4>${subscriber.email}</h4>
                    <p>Subscribed: ${new Date(subscriber.created_at).toLocaleDateString()}</p>
                </div>
                <div class="list-item-actions">
                    <button class="action-btn delete" onclick="deleteSubscriber(${subscriber.id})" title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async loadMonitors() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/monitors');
            if (response.ok) {
                const data = await response.json();
                this.renderMonitors(data.monitors);
            }
        } catch (error) {
            console.error('Failed to load monitors:', error);
        }
    }

    renderMonitors(monitors) {
        const container = document.getElementById('monitors-list');
        if (monitors.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-heartbeat"></i>
                    <h3>No Monitors</h3>
                    <p>No monitors configured</p>
                </div>
            `;
            return;
        }

        container.innerHTML = monitors.map(monitor => `
            <div class="list-item">
                <div class="list-item-content">
                    <h4>${monitor.name}</h4>
                    <p>${monitor.url}</p>
                    <span class="status-badge ${monitor.status}">${monitor.status}</span>
                </div>
                <div class="list-item-actions">
                    <button class="action-btn edit" onclick="editMonitor(${monitor.id})" title="Edit">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="action-btn delete" onclick="deleteMonitor(${monitor.id})" title="Delete">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async loadBranding() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/branding');
            if (response.ok) {
                const data = await response.json();
                this.populateBrandingForm(data.branding);
            }
        } catch (error) {
            console.error('Failed to load branding:', error);
        }
    }

    populateBrandingForm(branding) {
        document.getElementById('company-name').value = branding.company_name || '';
        document.getElementById('primary-color').value = branding.primary_color || '#0052cc';
        document.getElementById('secondary-color').value = branding.secondary_color || '#f4f5f7';
        document.getElementById('logo-url').value = branding.logo_url || '';
        document.getElementById('footer-text').value = branding.footer_text || '';
    }

    async loadSettings() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/settings');
            if (response.ok) {
                const data = await response.json();
                this.populateSettingsForm(data.settings);
            }
        } catch (error) {
            console.error('Failed to load settings:', error);
        }
    }

    populateSettingsForm(settings) {
        document.getElementById('contact-email').value = settings.contact_email || '';
        document.getElementById('timezone').value = settings.timezone || 'UTC';
        document.getElementById('custom-domain').value = settings.custom_domain || '';
    }

    async loadBilling() {
        try {
            // Load billing overview data
            const [billingResponse, metricsResponse] = await Promise.all([
                this.authenticatedFetch('/api/v1/tenant/billing'),
                this.authenticatedFetch('/api/v1/tenant/billing/metrics?period=30d')
            ]);

            if (billingResponse.ok && metricsResponse.ok) {
                const billingData = await billingResponse.json();
                const metricsData = await metricsResponse.json();
                this.renderBillingOverview(billingData, metricsData);
            }
        } catch (error) {
            console.error('Failed to load billing:', error);
        }
    }

    async loadBillingInvoices() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/billing/invoices');
            if (response.ok) {
                const data = await response.json();
                this.renderBillingInvoices(data.invoices);
            }
        } catch (error) {
            console.error('Failed to load invoices:', error);
        }
    }

    async loadBillingPayments() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/billing/payments');
            if (response.ok) {
                const data = await response.json();
                this.renderBillingPayments(data.payments);
            }
        } catch (error) {
            console.error('Failed to load payments:', error);
        }
    }

    async loadDomains() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/domains/status');
            if (response.ok) {
                const data = await response.json();
                this.renderDomainStatus(data.status, data.instructions);
            } else if (response.status === 404) {
                // No custom domain configured
                this.renderDomainStatus(null, null);
            }
        } catch (error) {
            console.error('Failed to load domain status:', error);
            this.renderDomainStatus(null, null);
        }
    }

    renderBillingOverview(billingData, metricsData) {
        // Update metrics
        document.getElementById('total-invoices').textContent = metricsData.total_invoices || 0;
        document.getElementById('total-paid').textContent = `$${(metricsData.total_paid || 0).toFixed(2)}`;
        document.getElementById('total-outstanding').textContent = `$${(metricsData.total_outstanding || 0).toFixed(2)}`;
        document.getElementById('net-revenue').textContent = `$${(metricsData.net_revenue || 0).toFixed(2)}`;

        // Render current plan
        this.renderCurrentPlan(billingData.current_plan);

        // Initialize charts
        this.initializeBillingCharts(metricsData);
    }

    renderCurrentPlan(currentPlan) {
        const currentPlanContainer = document.getElementById('current-plan');
        if (currentPlan) {
            currentPlanContainer.innerHTML = `
                <h3>Current Plan</h3>
                <div class="plan-card current">
                    <div class="plan-header">
                        <span class="plan-name">${currentPlan.name}</span>
                        <span class="plan-price">$${currentPlan.price}/${currentPlan.billing_interval}</span>
                    </div>
                    <p>${currentPlan.description}</p>
                    <ul class="plan-features">
                        <li><i class="fas fa-check"></i> ${currentPlan.max_services} Services</li>
                        <li><i class="fas fa-check"></i> ${currentPlan.max_monitors} Monitors</li>
                        <li><i class="fas fa-check"></i> ${currentPlan.max_subscribers} Subscribers</li>
                        <li><i class="fas fa-check"></i> ${currentPlan.max_incidents} Incidents</li>
                    </ul>
                    <button class="btn btn-danger" onclick="cancelSubscription()">Cancel Subscription</button>
                </div>
            `;
        } else {
            currentPlanContainer.innerHTML = `
                <h3>Current Plan</h3>
                <div class="plan-card current">
                    <div class="plan-header">
                        <span class="plan-name">Free Plan</span>
                        <span class="plan-price">$0/month</span>
                    </div>
                    <p>Basic monitoring and notifications</p>
                    <ul class="plan-features">
                        <li><i class="fas fa-check"></i> 3 Services</li>
                        <li><i class="fas fa-check"></i> 5 Monitors</li>
                        <li><i class="fas fa-check"></i> 100 Subscribers</li>
                        <li><i class="fas fa-check"></i> Basic Support</li>
                    </ul>
                </div>
            `;
        }
    }

    renderBillingInvoices(invoices) {
        const invoicesList = document.getElementById('invoices-list');
        if (invoices.length === 0) {
            invoicesList.innerHTML = '<div class="empty-state"><p>No invoices found</p></div>';
            return;
        }

        invoicesList.innerHTML = `
            <div class="invoices-table">
                <table>
                    <thead>
                        <tr>
                            <th>Invoice #</th>
                            <th>Date</th>
                            <th>Amount</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${invoices.map(invoice => `
                            <tr>
                                <td>#${invoice.invoice_number}</td>
                                <td>${new Date(invoice.created_at).toLocaleDateString()}</td>
                                <td>$${invoice.total_amount.toFixed(2)}</td>
                                <td><span class="status-badge status-${invoice.status}">${invoice.status}</span></td>
                                <td>
                                    <button class="btn btn-sm btn-primary" onclick="viewInvoice('${invoice.id}')">View</button>
                                    ${invoice.status === 'pending' ? `<button class="btn btn-sm btn-success" onclick="payInvoice('${invoice.id}')">Pay</button>` : ''}
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    renderBillingPayments(payments) {
        const paymentsList = document.getElementById('payments-list');
        if (payments.length === 0) {
            paymentsList.innerHTML = '<div class="empty-state"><p>No payments found</p></div>';
            return;
        }

        paymentsList.innerHTML = `
            <div class="payments-table">
                <table>
                    <thead>
                        <tr>
                            <th>Payment ID</th>
                            <th>Date</th>
                            <th>Amount</th>
                            <th>Method</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${payments.map(payment => `
                            <tr>
                                <td>${payment.gateway_id}</td>
                                <td>${new Date(payment.created_at).toLocaleDateString()}</td>
                                <td>$${payment.amount.toFixed(2)}</td>
                                <td>${payment.method}</td>
                                <td><span class="status-badge status-${payment.status}">${payment.status}</span></td>
                                <td>
                                    <button class="btn btn-sm btn-primary" onclick="viewPayment('${payment.id}')">View</button>
                                    ${payment.status === 'completed' ? `<button class="btn btn-sm btn-warning" onclick="refundPayment('${payment.id}')">Refund</button>` : ''}
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    initializeBillingCharts(metricsData) {
        // Revenue trend chart
        const revenueCtx = document.getElementById('revenue-chart');
        if (revenueCtx) {
            new Chart(revenueCtx, {
                type: 'line',
                data: {
                    labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
                    datasets: [{
                        label: 'Revenue',
                        data: [metricsData.total_paid * 0.2, metricsData.total_paid * 0.4, metricsData.total_paid * 0.6, metricsData.total_paid],
                        borderColor: '#4CAF50',
                        backgroundColor: 'rgba(76, 175, 80, 0.1)',
                        tension: 0.4
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

        // Payment methods chart
        const paymentMethodsCtx = document.getElementById('payment-methods-chart');
        if (paymentMethodsCtx) {
            new Chart(paymentMethodsCtx, {
                type: 'doughnut',
                data: {
                    labels: ['Card', 'UPI', 'Net Banking', 'Wallet'],
                    datasets: [{
                        data: [40, 30, 20, 10],
                        backgroundColor: ['#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0']
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
    }

    async handleBrandingSubmit(e) {
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());

        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/branding', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                this.showNotification('Branding updated successfully', 'success');
            } else {
                this.showNotification('Failed to update branding', 'error');
            }
        } catch (error) {
            console.error('Failed to update branding:', error);
            this.showNotification('Failed to update branding', 'error');
        }
    }

    async handleSettingsSubmit(e) {
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());

        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/settings', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                this.showNotification('Settings updated successfully', 'success');
            } else {
                this.showNotification('Failed to update settings', 'error');
            }
        } catch (error) {
            console.error('Failed to update settings:', error);
            this.showNotification('Failed to update settings', 'error');
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
            window.location.href = '/login';
            return response;
        }

        return response;
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.innerHTML = `
            <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
            <span>${message}</span>
        `;

        // Add to page
        document.body.appendChild(notification);

        // Remove after 3 seconds
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }

    renderDomainStatus(status, instructions) {
        const statusContent = document.getElementById('domain-status-content');
        const setupInstructions = document.getElementById('setup-instructions');
        const sslManagement = document.getElementById('ssl-management');

        if (!status) {
            // No custom domain configured
            statusContent.innerHTML = `
                <div class="no-domain">
                    <div class="no-domain-icon">
                        <i class="fas fa-globe"></i>
                    </div>
                    <h4>No Custom Domain Configured</h4>
                    <p>Set up a custom domain to use your own domain for your status page and admin dashboard.</p>
                </div>
            `;
            setupInstructions.style.display = 'none';
            sslManagement.style.display = 'none';
            return;
        }

        // Render domain status
        const statusHTML = `
            <div class="domain-info">
                <div class="domain-name">
                    <i class="fas fa-globe"></i>
                    <span>${status.domain}</span>
                </div>
                <div class="domain-status-badges">
                    <span class="status-badge ${status.is_verified ? 'status-verified' : 'status-pending'}">
                        <i class="fas fa-${status.is_verified ? 'check-circle' : 'clock'}"></i>
                        ${status.is_verified ? 'Verified' : 'Pending Verification'}
                    </span>
                    <span class="status-badge ${status.ssl_status.is_valid ? 'status-verified' : 'status-pending'}">
                        <i class="fas fa-${status.ssl_status.is_valid ? 'shield-alt' : 'exclamation-triangle'}"></i>
                        ${status.ssl_status.is_valid ? 'SSL Active' : 'SSL Pending'}
                    </span>
                </div>
                <div class="domain-details">
                    <p><strong>Last Checked:</strong> ${new Date(status.last_checked).toLocaleString()}</p>
                    ${status.error ? `<p class="error"><strong>Error:</strong> ${status.error}</p>` : ''}
                    ${status.ssl_status.is_valid ? `
                        <p><strong>SSL Expires:</strong> ${new Date(status.ssl_status.expires_at).toLocaleDateString()}</p>
                        <p><strong>SSL Issuer:</strong> ${status.ssl_status.issuer}</p>
                    ` : ''}
                </div>
            </div>
        `;
        statusContent.innerHTML = statusHTML;

        // Show/hide setup instructions and SSL management based on status
        if (!status.is_verified || !status.ssl_status.is_valid) {
            setupInstructions.style.display = 'block';
            this.renderSetupInstructions(instructions);
        } else {
            setupInstructions.style.display = 'none';
        }

        if (status.is_verified && !status.ssl_status.is_valid) {
            sslManagement.style.display = 'block';
        } else {
            sslManagement.style.display = 'none';
        }
    }

    renderSetupInstructions(instructions) {
        const instructionsContent = document.getElementById('instructions-content');
        
        if (!instructions) {
            instructionsContent.innerHTML = '<p>Loading instructions...</p>';
            return;
        }

        let instructionsHTML = '';

        if (instructions.dns_instructions) {
            instructionsHTML += `
                <div class="instruction-section">
                    <h4><i class="fas fa-server"></i> DNS Configuration</h4>
                    <p>Add the following DNS records to your domain:</p>
                    <div class="dns-records">
                        ${instructions.dns_instructions.instructions.map(record => `
                            <div class="dns-record">
                                <div class="record-type">${record.type}</div>
                                <div class="record-name">${record.name}</div>
                                <div class="record-value">${record.value}</div>
                                <div class="record-ttl">TTL: ${record.ttl}</div>
                            </div>
                        `).join('')}
                    </div>
                    ${instructions.dns_instructions.verification ? `
                        <div class="verification-record">
                            <h5>Verification Record</h5>
                            <div class="dns-record">
                                <div class="record-type">${instructions.dns_instructions.verification.type}</div>
                                <div class="record-name">${instructions.dns_instructions.verification.name}</div>
                                <div class="record-value">${instructions.dns_instructions.verification.value}</div>
                            </div>
                        </div>
                    ` : ''}
                </div>
            `;
        }

        if (instructions.ssl_instructions) {
            instructionsHTML += `
                <div class="instruction-section">
                    <h4><i class="fas fa-shield-alt"></i> SSL Certificate Setup</h4>
                    <p>Choose one of the following SSL options:</p>
                    <div class="ssl-options">
                        ${instructions.ssl_instructions.options.map(option => `
                            <div class="ssl-option">
                                <h5>${option.name}</h5>
                                <p>${option.description}</p>
                                <p><strong>Setup:</strong> ${option.setup}</p>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        instructionsContent.innerHTML = instructionsHTML;
    }

    async loadSecurity() {
        try {
            // Load security configuration and metrics
            const [configResponse, metricsResponse] = await Promise.all([
                this.authenticatedFetch('/api/v1/tenant/security/config'),
                this.authenticatedFetch('/api/v1/tenant/security/metrics')
            ]);

            if (configResponse.ok) {
                const configData = await configResponse.json();
                this.renderSecurityConfig(configData.config);
            }

            if (metricsResponse.ok) {
                const metricsData = await metricsResponse.json();
                this.renderSecurityMetrics(metricsData.metrics);
            }

            // Load audit logs
            await this.loadAuditLogs();
        } catch (error) {
            console.error('Failed to load security data:', error);
        }
    }

    renderSecurityConfig(config) {
        // Update form fields with current configuration
        document.getElementById('data-encryption').checked = config.data_encryption;
        document.getElementById('audit-logging').checked = config.audit_logging;
        document.getElementById('two-factor-auth').checked = config.two_factor_auth;
        document.getElementById('session-timeout').value = config.session_timeout;
        document.getElementById('password-policy').value = config.password_policy;
        document.getElementById('api-key-rotation').value = config.api_key_rotation;
    }

    renderSecurityMetrics(metrics) {
        const metricsContainer = document.getElementById('security-metrics');
        
        const metricsHTML = `
            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-clipboard-list"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${metrics.audit_events.total}</h4>
                        <p>Audit Events</p>
                        <small>${metrics.audit_events.successful} successful, ${metrics.audit_events.failed} failed</small>
                    </div>
                </div>
                
                <div class="metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-exclamation-triangle"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${metrics.security_violations.total}</h4>
                        <p>Security Violations</p>
                        <small>${metrics.security_violations.by_severity.critical} critical, ${metrics.security_violations.by_severity.high} high</small>
                    </div>
                </div>
                
                <div class="metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-database"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${metrics.data_access.total_requests}</h4>
                        <p>Data Access Requests</p>
                        <small>${metrics.data_access.unique_users} unique users</small>
                    </div>
                </div>
                
                <div class="metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-shield-alt"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${metrics.compliance.data_encryption_enabled ? 'Enabled' : 'Disabled'}</h4>
                        <p>Data Encryption</p>
                        <small>${metrics.compliance.audit_logging_enabled ? 'Audit logging active' : 'Audit logging disabled'}</small>
                    </div>
                </div>
            </div>
        `;
        
        metricsContainer.innerHTML = metricsHTML;
    }

    async loadAuditLogs() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/security/audit-logs');
            if (response.ok) {
                const data = await response.json();
                this.renderAuditLogs(data.logs);
            }
        } catch (error) {
            console.error('Failed to load audit logs:', error);
        }
    }

    renderAuditLogs(logs) {
        const logsContainer = document.getElementById('audit-logs-list');
        
        if (logs.length === 0) {
            logsContainer.innerHTML = '<p class="no-data">No audit logs found</p>';
            return;
        }

        const logsHTML = `
            <div class="audit-logs-table">
                <div class="table-header">
                    <div class="header-cell">Timestamp</div>
                    <div class="header-cell">User</div>
                    <div class="header-cell">Action</div>
                    <div class="header-cell">Resource</div>
                    <div class="header-cell">IP Address</div>
                    <div class="header-cell">Status</div>
                </div>
                ${logs.map(log => `
                    <div class="table-row">
                        <div class="table-cell">${new Date(log.timestamp).toLocaleString()}</div>
                        <div class="table-cell">User ${log.user_id || 'N/A'}</div>
                        <div class="table-cell">${log.action}</div>
                        <div class="table-cell">${log.resource}</div>
                        <div class="table-cell">${log.ip_address}</div>
                        <div class="table-cell">
                            <span class="status-badge ${log.success ? 'status-success' : 'status-error'}">
                                ${log.success ? 'Success' : 'Failed'}
                            </span>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        
        logsContainer.innerHTML = logsHTML;
    }

    async loadSecurityViolations() {
        try {
            const response = await this.authenticatedFetch('/api/v1/tenant/security/violations');
            if (response.ok) {
                const data = await response.json();
                this.renderSecurityViolations(data.violations);
            }
        } catch (error) {
            console.error('Failed to load security violations:', error);
        }
    }

    renderSecurityViolations(violations) {
        const violationsContainer = document.getElementById('violations-list');
        
        if (violations.length === 0) {
            violationsContainer.innerHTML = '<p class="no-data">No security violations found</p>';
            return;
        }

        const violationsHTML = `
            <div class="violations-table">
                <div class="table-header">
                    <div class="header-cell">Timestamp</div>
                    <div class="header-cell">Type</div>
                    <div class="header-cell">Severity</div>
                    <div class="header-cell">Description</div>
                    <div class="header-cell">IP Address</div>
                    <div class="header-cell">Status</div>
                    <div class="header-cell">Actions</div>
                </div>
                ${violations.map(violation => `
                    <div class="table-row">
                        <div class="table-cell">${new Date(violation.timestamp).toLocaleString()}</div>
                        <div class="table-cell">${violation.type}</div>
                        <div class="table-cell">
                            <span class="severity-badge severity-${violation.severity}">
                                ${violation.severity.toUpperCase()}
                            </span>
                        </div>
                        <div class="table-cell">${violation.description}</div>
                        <div class="table-cell">${violation.ip_address}</div>
                        <div class="table-cell">
                            <span class="status-badge ${violation.resolved ? 'status-resolved' : 'status-pending'}">
                                ${violation.resolved ? 'Resolved' : 'Unresolved'}
                            </span>
                        </div>
                        <div class="table-cell">
                            ${!violation.resolved ? `
                                <button class="btn btn-sm btn-success" onclick="resolveViolation(${violation.id})">
                                    <i class="fas fa-check"></i> Resolve
                                </button>
                            ` : ''}
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        
        violationsContainer.innerHTML = violationsHTML;
    }

    async loadAnalytics() {
        try {
            const period = document.getElementById('analytics-period')?.value || '24h';
            
            // Load analytics data
            const [summaryResponse, metricsResponse, healthResponse] = await Promise.all([
                this.authenticatedFetch('/api/v1/tenant/monitoring/uptime-summary'),
                this.authenticatedFetch(`/api/v1/tenant/monitoring/metrics?period=${period}`),
                this.authenticatedFetch('/api/v1/tenant/monitoring/services/health')
            ]);

            if (summaryResponse.ok) {
                const summaryData = await summaryResponse.json();
                this.renderAnalyticsOverview(summaryData.summary);
            }

            if (metricsResponse.ok) {
                const metricsData = await metricsResponse.json();
                this.renderAnalyticsMetrics(metricsData.metrics);
            }

            if (healthResponse.ok) {
                const healthData = await healthResponse.json();
                this.renderServiceHealthData(healthData.services);
            }

            // Initialize charts
            this.initializeAnalyticsCharts();
        } catch (error) {
            console.error('Failed to load analytics data:', error);
        }
    }

    renderAnalyticsOverview(summary) {
        const overviewContainer = document.getElementById('overview-metrics');
        
        const overviewHTML = `
            <div class="overview-metrics-grid">
                <div class="overview-metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-server"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${summary.total_services}</h4>
                        <p>Total Services</p>
                        <small>${summary.up_services} up, ${summary.degraded_services} degraded, ${summary.down_services} down</small>
                    </div>
                </div>
                
                <div class="overview-metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-chart-line"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${summary.average_uptime.toFixed(2)}%</h4>
                        <p>Average Uptime</p>
                        <small>Overall system uptime</small>
                    </div>
                </div>
                
                <div class="overview-metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-shield-alt ${summary.overall_status === 'up' ? 'status-up' : summary.overall_status === 'degraded' ? 'status-degraded' : 'status-down'}"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${summary.overall_status.toUpperCase()}</h4>
                        <p>Overall Status</p>
                        <small>System health status</small>
                    </div>
                </div>
                
                <div class="overview-metric-card">
                    <div class="metric-icon">
                        <i class="fas fa-clock"></i>
                    </div>
                    <div class="metric-info">
                        <h4>${new Date(summary.last_updated).toLocaleTimeString()}</h4>
                        <p>Last Updated</p>
                        <small>Data refresh time</small>
                    </div>
                </div>
            </div>
        `;
        
        overviewContainer.innerHTML = overviewHTML;
    }

    renderAnalyticsMetrics(metrics) {
        // Store metrics for chart rendering
        this.analyticsMetrics = metrics;
        
        // Render uptime table
        this.renderUptimeTable(metrics.services);
        
        // Render response time table
        this.renderResponseTimeTable(metrics.services);
        
        // Render error table
        this.renderErrorTable(metrics.services);
        
        // Render incident table
        this.renderIncidentTable(metrics.incidents);
    }

    renderServiceHealthData(services) {
        // Store services for chart rendering
        this.serviceHealthData = services;
    }

    renderUptimeTable(services) {
        const tableBody = document.getElementById('uptime-table-body');
        
        if (!services || services.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="6" class="no-data">No service data available</td></tr>';
            return;
        }

        const tableHTML = services.map(service => `
            <tr>
                <td>${service.service_name}</td>
                <td>
                    <span class="status-badge status-${service.status}">
                        ${service.status.toUpperCase()}
                    </span>
                </td>
                <td>${service.uptime.toFixed(2)}%</td>
                <td>${new Date(service.last_checked).toLocaleString()}</td>
                <td>${service.incident_count}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewServiceDetails(${service.service_id})">
                        <i class="fas fa-chart-line"></i> View Details
                    </button>
                </td>
            </tr>
        `).join('');
        
        tableBody.innerHTML = tableHTML;
    }

    renderResponseTimeTable(services) {
        const tableBody = document.getElementById('response-time-table-body');
        
        if (!services || services.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="7" class="no-data">No service data available</td></tr>';
            return;
        }

        const tableHTML = services.map(service => `
            <tr>
                <td>${service.service_name}</td>
                <td>${service.response_time.toFixed(0)}ms</td>
                <td>${(service.response_time * 0.8).toFixed(0)}ms</td>
                <td>${(service.response_time * 1.5).toFixed(0)}ms</td>
                <td>${(service.response_time * 1.2).toFixed(0)}ms</td>
                <td>${(service.response_time * 1.4).toFixed(0)}ms</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewServiceDetails(${service.service_id})">
                        <i class="fas fa-chart-line"></i> View Details
                    </button>
                </td>
            </tr>
        `).join('');
        
        tableBody.innerHTML = tableHTML;
    }

    renderErrorTable(services) {
        const tableBody = document.getElementById('error-table-body');
        
        if (!services || services.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="6" class="no-data">No service data available</td></tr>';
            return;
        }

        const tableHTML = services.map(service => `
            <tr>
                <td>${service.service_name}</td>
                <td>${service.error_rate.toFixed(2)}%</td>
                <td>${Math.floor(service.error_rate * 10)}</td>
                <td>${new Date(service.last_checked).toLocaleString()}</td>
                <td>
                    <span class="status-badge status-${service.status}">
                        ${service.status.toUpperCase()}
                    </span>
                </td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewServiceDetails(${service.service_id})">
                        <i class="fas fa-chart-line"></i> View Details
                    </button>
                </td>
            </tr>
        `).join('');
        
        tableBody.innerHTML = tableHTML;
    }

    renderIncidentTable(incidents) {
        const tableBody = document.getElementById('incident-table-body');
        
        if (!incidents || incidents.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="7" class="no-data">No incidents found</td></tr>';
            return;
        }

        const tableHTML = incidents.map(incident => `
            <tr>
                <td>${incident.Title}</td>
                <td>Service ${incident.ServiceID || 'N/A'}</td>
                <td>
                    <span class="status-badge status-${incident.Status}">
                        ${incident.Status.toUpperCase()}
                    </span>
                </td>
                <td>${incident.Impact || 'Unknown'}</td>
                <td>${this.calculateDuration(incident.CreatedAt, incident.UpdatedAt)}</td>
                <td>${new Date(incident.CreatedAt).toLocaleString()}</td>
                <td>
                    <button class="btn btn-sm btn-primary" onclick="viewIncidentDetails(${incident.ID})">
                        <i class="fas fa-eye"></i> View
                    </button>
                </td>
            </tr>
        `).join('');
        
        tableBody.innerHTML = tableHTML;
    }

    initializeAnalyticsCharts() {
        // Initialize Chart.js charts
        if (typeof Chart === 'undefined') {
            console.warn('Chart.js not loaded, skipping chart initialization');
            return;
        }

        // Overall Uptime Chart
        this.initializeOverallUptimeChart();
        
        // Service Uptime Comparison Chart
        this.initializeServiceUptimeChart();
        
        // Response Time Chart
        this.initializeResponseTimeChart();
        
        // Error Rate Chart
        this.initializeErrorRateChart();
        
        // Incidents Timeline Chart
        this.initializeIncidentsTimelineChart();
    }

    initializeOverallUptimeChart() {
        const ctx = document.getElementById('overall-uptime-chart');
        if (!ctx) return;

        const metrics = this.analyticsMetrics;
        if (!metrics) return;

        // Generate mock uptime data
        const labels = [];
        const uptimeData = [];
        const downtimeData = [];
        
        const startTime = new Date(metrics.start_time);
        const endTime = new Date(metrics.end_time);
        const interval = (endTime - startTime) / 20; // 20 data points
        
        for (let i = 0; i < 20; i++) {
            const time = new Date(startTime.getTime() + i * interval);
            labels.push(time.toLocaleTimeString());
            
            const baseUptime = metrics.overall_uptime;
            const variation = (Math.random() - 0.5) * 2; // ±1%
            const uptime = Math.max(95, Math.min(100, baseUptime + variation));
            
            uptimeData.push(uptime);
            downtimeData.push(100 - uptime);
        }

        new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Uptime %',
                    data: uptimeData,
                    borderColor: '#4CAF50',
                    backgroundColor: 'rgba(76, 175, 80, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: false,
                        min: 95,
                        max: 100
                    }
                }
            }
        });
    }

    initializeServiceUptimeChart() {
        const ctx = document.getElementById('service-uptime-chart');
        if (!ctx) return;

        const services = this.serviceHealthData;
        if (!services) return;

        const labels = services.map(s => s.service_name);
        const uptimeData = services.map(s => s.uptime);

        new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Uptime %',
                    data: uptimeData,
                    backgroundColor: services.map(s => 
                        s.status === 'up' ? '#4CAF50' : 
                        s.status === 'degraded' ? '#FF9800' : '#F44336'
                    )
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: false,
                        min: 95,
                        max: 100
                    }
                }
            }
        });
    }

    initializeResponseTimeChart() {
        const ctx = document.getElementById('response-time-chart');
        if (!ctx) return;

        // Generate mock response time data
        const labels = [];
        const responseTimeData = [];
        
        for (let i = 0; i < 24; i++) {
            const hour = i;
            labels.push(`${hour}:00`);
            
            const baseTime = 150;
            const variation = (Math.random() - 0.5) * 100;
            const responseTime = Math.max(50, baseTime + variation);
            
            responseTimeData.push(responseTime);
        }

        new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Response Time (ms)',
                    data: responseTimeData,
                    borderColor: '#2196F3',
                    backgroundColor: 'rgba(33, 150, 243, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    initializeErrorRateChart() {
        const ctx = document.getElementById('error-rate-chart');
        if (!ctx) return;

        // Generate mock error rate data
        const labels = [];
        const errorRateData = [];
        
        for (let i = 0; i < 24; i++) {
            const hour = i;
            labels.push(`${hour}:00`);
            
            const baseRate = 0.1;
            const variation = (Math.random() - 0.5) * 0.2;
            const errorRate = Math.max(0, baseRate + variation);
            
            errorRateData.push(errorRate);
        }

        new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Error Rate %',
                    data: errorRateData,
                    borderColor: '#F44336',
                    backgroundColor: 'rgba(244, 67, 54, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    initializeIncidentsTimelineChart() {
        const ctx = document.getElementById('incidents-timeline-chart');
        if (!ctx) return;

        const incidents = this.analyticsMetrics?.incidents || [];
        
        // Group incidents by day
        const incidentCounts = {};
        incidents.forEach(incident => {
            const date = new Date(incident.CreatedAt).toDateString();
            incidentCounts[date] = (incidentCounts[date] || 0) + 1;
        });

        const labels = Object.keys(incidentCounts);
        const data = Object.values(incidentCounts);

        new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    label: 'Incidents',
                    data: data,
                    backgroundColor: '#FF9800'
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });
    }

    calculateDuration(startTime, endTime) {
        const start = new Date(startTime);
        const end = endTime ? new Date(endTime) : new Date();
        const diff = end - start;
        
        const hours = Math.floor(diff / (1000 * 60 * 60));
        const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        
        if (hours > 0) {
            return `${hours}h ${minutes}m`;
        }
        return `${minutes}m`;
    }
}

// Billing tab switching functionality
function switchBillingTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.billing-tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Remove active class from all tab buttons
    document.querySelectorAll('.billing-tabs .tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab content
    document.getElementById(`billing-${tabName}`).classList.add('active');
    
    // Add active class to clicked button
    event.target.classList.add('active');
    
    // Load data for the selected tab
    switch(tabName) {
        case 'overview':
            adminDashboard.loadBilling();
            break;
        case 'invoices':
            adminDashboard.loadBillingInvoices();
            break;
        case 'payments':
            adminDashboard.loadBillingPayments();
            break;
        case 'plans':
            // Plans are loaded with the main billing data
            break;
    }
}

// Billing action functions
function viewInvoice(invoiceId) {
    // Open invoice in new window or modal
    window.open(`/api/v1/tenant/billing/invoices/${invoiceId}/pdf`, '_blank');
}

function payInvoice(invoiceId) {
    // Create payment session for invoice
    fetch('/api/v1/tenant/billing/invoices/' + invoiceId + '/pay', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.payment_url) {
            window.location.href = data.payment_url;
        } else {
            adminDashboard.showNotification('Payment initiated', 'success');
        }
    })
    .catch(error => {
        console.error('Payment error:', error);
        adminDashboard.showNotification('Payment failed', 'error');
    });
}

function viewPayment(paymentId) {
    // Show payment details modal
    adminDashboard.showNotification('Payment details feature coming soon', 'info');
}

function refundPayment(paymentId) {
    if (confirm('Are you sure you want to refund this payment?')) {
        fetch('/api/v1/tenant/billing/payments/' + paymentId + '/refund', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
            }
        })
        .then(response => response.json())
        .then(data => {
            adminDashboard.showNotification('Refund initiated', 'success');
            adminDashboard.loadBillingPayments(); // Refresh payments list
        })
        .catch(error => {
            console.error('Refund error:', error);
            adminDashboard.showNotification('Refund failed', 'error');
        });
    }
}

function downloadInvoice() {
    adminDashboard.showNotification('Invoice download feature coming soon', 'info');
}

function upgradeToPlan(planSlug) {
    // Create checkout session for plan upgrade
    fetch('/api/v1/tenant/checkout', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: JSON.stringify({
            plan_slug: planSlug
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.checkout_url) {
            window.location.href = data.checkout_url;
        } else {
            adminDashboard.showNotification('Checkout session created', 'success');
        }
    })
    .catch(error => {
        console.error('Checkout error:', error);
        adminDashboard.showNotification('Failed to create checkout session', 'error');
    });
}

function cancelSubscription() {
    if (confirm('Are you sure you want to cancel your subscription? This action cannot be undone.')) {
        fetch('/api/v1/tenant/subscription/cancel', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
            }
        })
        .then(response => response.json())
        .then(data => {
            adminDashboard.showNotification('Subscription cancelled', 'success');
            adminDashboard.loadBilling(); // Refresh billing data
        })
        .catch(error => {
            console.error('Cancellation error:', error);
            adminDashboard.showNotification('Failed to cancel subscription', 'error');
        });
    }
}

// Domain management functions
function handleDomainSetup(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData.entries());
    
    fetch('/api/v1/tenant/domains', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if (data.message) {
            adminDashboard.showNotification(data.message, 'success');
            adminDashboard.loadDomains(); // Refresh domain status
        } else {
            adminDashboard.showNotification('Domain setup initiated', 'success');
        }
    })
    .catch(error => {
        console.error('Domain setup error:', error);
        adminDashboard.showNotification('Failed to set up domain', 'error');
    });
}

function refreshDomainStatus() {
    adminDashboard.loadDomains();
    adminDashboard.showNotification('Domain status refreshed', 'info');
}

function enableAutomaticSSL() {
    fetch('/api/v1/tenant/domains/ssl/auto', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        }
    })
    .then(response => response.json())
    .then(data => {
        adminDashboard.showNotification(data.message || 'Automatic SSL enabled', 'success');
        adminDashboard.loadDomains(); // Refresh domain status
    })
    .catch(error => {
        console.error('SSL enable error:', error);
        adminDashboard.showNotification('Failed to enable automatic SSL', 'error');
    });
}

function handleSSLUpload(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    
    fetch('/api/v1/tenant/domains/ssl/upload', {
        method: 'POST',
        headers: {
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        adminDashboard.showNotification(data.message || 'SSL certificate uploaded', 'success');
        adminDashboard.loadDomains(); // Refresh domain status
    })
    .catch(error => {
        console.error('SSL upload error:', error);
        adminDashboard.showNotification('Failed to upload SSL certificate', 'error');
    });
}

// Security management functions
function handleSecurityConfigSubmit(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData.entries());
    
    // Convert checkboxes to boolean
    data.data_encryption = document.getElementById('data-encryption').checked;
    data.audit_logging = document.getElementById('audit-logging').checked;
    data.two_factor_auth = document.getElementById('two-factor-auth').checked;
    
    fetch('/api/v1/tenant/security/config', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        adminDashboard.showNotification(data.message || 'Security configuration updated', 'success');
        adminDashboard.loadSecurity(); // Refresh security data
    })
    .catch(error => {
        console.error('Security config error:', error);
        adminDashboard.showNotification('Failed to update security configuration', 'error');
    });
}

function switchSecurityTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Remove active class from all tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab content
    document.getElementById(tabName + '-tab').classList.add('active');
    
    // Add active class to clicked button
    event.target.classList.add('active');
    
    // Load data for the active tab
    switch(tabName) {
        case 'audit-logs':
            adminDashboard.loadAuditLogs();
            break;
        case 'violations':
            adminDashboard.loadSecurityViolations();
            break;
        case 'api-keys':
            // Load API keys (placeholder)
            break;
        case 'data-encryption':
            // Data encryption tools are already loaded
            break;
    }
}

function refreshAuditLogs() {
    adminDashboard.loadAuditLogs();
    adminDashboard.showNotification('Audit logs refreshed', 'info');
}

function refreshViolations() {
    adminDashboard.loadSecurityViolations();
    adminDashboard.showNotification('Security violations refreshed', 'info');
}

function resolveViolation(violationId) {
    if (confirm('Are you sure you want to resolve this security violation?')) {
        fetch(`/api/v1/tenant/security/violations/${violationId}/resolve`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
            }
        })
        .then(response => response.json())
        .then(data => {
            adminDashboard.showNotification(data.message || 'Security violation resolved', 'success');
            adminDashboard.loadSecurityViolations(); // Refresh violations list
        })
        .catch(error => {
            console.error('Resolve violation error:', error);
            adminDashboard.showNotification('Failed to resolve security violation', 'error');
        });
    }
}

function showGenerateAPIKeyModal() {
    const name = prompt('Enter a name for the API key:');
    if (name) {
        generateAPIKey(name);
    }
}

function generateAPIKey(name) {
    fetch('/api/v1/tenant/security/api-keys', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: JSON.stringify({ name: name })
    })
    .then(response => response.json())
    .then(data => {
        if (data.api_key) {
            // Show API key in a modal or alert
            alert(`API Key Generated!\n\nName: ${data.name}\nKey: ${data.api_key}\n\n${data.warning}`);
            adminDashboard.showNotification('API key generated successfully', 'success');
        } else {
            adminDashboard.showNotification(data.message || 'API key generated', 'success');
        }
    })
    .catch(error => {
        console.error('Generate API key error:', error);
        adminDashboard.showNotification('Failed to generate API key', 'error');
    });
}

function handleEncryptData(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData.entries());
    
    fetch('/api/v1/tenant/security/encrypt', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if (data.encrypted_data) {
            document.getElementById('encrypted-data').value = data.encrypted_data;
            document.getElementById('encrypted-result').style.display = 'block';
            adminDashboard.showNotification('Data encrypted successfully', 'success');
        } else {
            adminDashboard.showNotification('Failed to encrypt data', 'error');
        }
    })
    .catch(error => {
        console.error('Encrypt data error:', error);
        adminDashboard.showNotification('Failed to encrypt data', 'error');
    });
}

function handleDecryptData(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData.entries());
    
    fetch('/api/v1/tenant/security/decrypt', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if (data.data) {
            document.getElementById('decrypted-data').value = data.data;
            document.getElementById('decrypted-result').style.display = 'block';
            adminDashboard.showNotification('Data decrypted successfully', 'success');
        } else {
            adminDashboard.showNotification('Failed to decrypt data', 'error');
        }
    })
    .catch(error => {
        console.error('Decrypt data error:', error);
        adminDashboard.showNotification('Failed to decrypt data', 'error');
    });
}

function copyToClipboard(elementId) {
    const element = document.getElementById(elementId);
    element.select();
    element.setSelectionRange(0, 99999); // For mobile devices
    
    try {
        document.execCommand('copy');
        adminDashboard.showNotification('Copied to clipboard', 'success');
    } catch (err) {
        adminDashboard.showNotification('Failed to copy to clipboard', 'error');
    }
}

// Analytics management functions
function switchAnalyticsTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Remove active class from all tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab content
    document.getElementById(tabName + '-tab').classList.add('active');
    
    // Add active class to clicked button
    event.target.classList.add('active');
    
    // Refresh charts for the active tab
    setTimeout(() => {
        if (adminDashboard && adminDashboard.initializeAnalyticsCharts) {
            adminDashboard.initializeAnalyticsCharts();
        }
    }, 100);
}

function changeAnalyticsPeriod() {
    const period = document.getElementById('analytics-period').value;
    
    // Reload analytics data with new period
    if (adminDashboard && adminDashboard.loadAnalytics) {
        adminDashboard.loadAnalytics();
    }
}

function viewServiceDetails(serviceId) {
    // Show service detail modal
    const modal = document.getElementById('service-detail-modal');
    const title = document.getElementById('service-detail-title');
    
    title.textContent = `Service ${serviceId} Details`;
    modal.style.display = 'block';
    
    // Load service details
    loadServiceDetails(serviceId);
}

function closeServiceDetailModal() {
    const modal = document.getElementById('service-detail-modal');
    modal.style.display = 'none';
}

function switchServiceDetailTab(tabName) {
    // Hide all detail tab contents
    document.querySelectorAll('.detail-tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Remove active class from all detail tab buttons
    document.querySelectorAll('.detail-tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected detail tab content
    document.getElementById('service-' + tabName + '-tab').classList.add('active');
    
    // Add active class to clicked button
    event.target.classList.add('active');
    
    // Load data for the active tab
    loadServiceDetailData(tabName);
}

function loadServiceDetails(serviceId) {
    // Load service overview data
    loadServiceDetailData('overview');
}

function loadServiceDetailData(tabName) {
    // In a real implementation, you'd load specific service data
    // For now, we'll just show a placeholder
    const contentDiv = document.getElementById('service-' + tabName + '-tab');
    
    if (tabName === 'overview') {
        contentDiv.innerHTML = `
            <div class="service-overview">
                <h4>Service Overview</h4>
                <p>Detailed service information and metrics will be displayed here.</p>
                <div class="service-metrics">
                    <div class="metric">
                        <span class="metric-label">Uptime:</span>
                        <span class="metric-value">99.5%</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Response Time:</span>
                        <span class="metric-value">150ms</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Error Rate:</span>
                        <span class="metric-value">0.1%</span>
                    </div>
                </div>
            </div>
        `;
    } else {
        // Initialize chart for the specific tab
        setTimeout(() => {
            initializeServiceDetailChart(tabName);
        }, 100);
    }
}

function initializeServiceDetailChart(tabName) {
    if (typeof Chart === 'undefined') {
        console.warn('Chart.js not loaded, skipping chart initialization');
        return;
    }

    const canvasId = 'service-' + tabName + '-detail-chart';
    const ctx = document.getElementById(canvasId);
    if (!ctx) return;

    // Generate mock data based on tab type
    let chartConfig;
    
    switch (tabName) {
        case 'uptime':
            chartConfig = {
                type: 'line',
                data: {
                    labels: Array.from({length: 24}, (_, i) => `${i}:00`),
                    datasets: [{
                        label: 'Uptime %',
                        data: Array.from({length: 24}, () => 99.5 + (Math.random() - 0.5) * 1),
                        borderColor: '#4CAF50',
                        backgroundColor: 'rgba(76, 175, 80, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: false,
                            min: 95,
                            max: 100
                        }
                    }
                }
            };
            break;
            
        case 'response-time':
            chartConfig = {
                type: 'line',
                data: {
                    labels: Array.from({length: 24}, (_, i) => `${i}:00`),
                    datasets: [{
                        label: 'Response Time (ms)',
                        data: Array.from({length: 24}, () => 150 + (Math.random() - 0.5) * 100),
                        borderColor: '#2196F3',
                        backgroundColor: 'rgba(33, 150, 243, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            };
            break;
            
        case 'errors':
            chartConfig = {
                type: 'line',
                data: {
                    labels: Array.from({length: 24}, (_, i) => `${i}:00`),
                    datasets: [{
                        label: 'Error Rate %',
                        data: Array.from({length: 24}, () => Math.max(0, 0.1 + (Math.random() - 0.5) * 0.2)),
                        borderColor: '#F44336',
                        backgroundColor: 'rgba(244, 67, 54, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            };
            break;
    }

    if (chartConfig) {
        new Chart(ctx, chartConfig);
    }
}

function viewIncidentDetails(incidentId) {
    // Show incident details
    adminDashboard.showNotification(`Viewing incident ${incidentId} details`, 'info');
}

// Modal functions
function showModal(title, content) {
    document.getElementById('modal-title').textContent = title;
    document.getElementById('modal-body').innerHTML = content;
    document.getElementById('modal-overlay').classList.add('active');
}

function closeModal() {
    document.getElementById('modal-overlay').classList.remove('active');
}

// Modal content generators
function showCreateServiceModal() {
    const content = `
        <form id="create-service-form">
            <div class="form-group">
                <label for="service-name">Service Name</label>
                <input type="text" id="service-name" name="name" required>
            </div>
            <div class="form-group">
                <label for="service-description">Description</label>
                <textarea id="service-description" name="description" rows="3"></textarea>
            </div>
            <div class="form-group">
                <label for="service-status">Status</label>
                <select id="service-status" name="status">
                    <option value="operational">Operational</option>
                    <option value="degraded">Degraded</option>
                    <option value="outage">Outage</option>
                </select>
            </div>
            <button type="submit" class="btn btn-primary">Create Service</button>
        </form>
    `;
    showModal('Create Service', content);
    
    document.getElementById('create-service-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        
        try {
            const response = await tenantAdmin.authenticatedFetch('/api/v1/tenant/services', {
                method: 'POST',
                body: JSON.stringify(data)
            });
            
            if (response.ok) {
                closeModal();
                tenantAdmin.loadServices();
                tenantAdmin.showNotification('Service created successfully', 'success');
            } else {
                tenantAdmin.showNotification('Failed to create service', 'error');
            }
        } catch (error) {
            console.error('Failed to create service:', error);
            tenantAdmin.showNotification('Failed to create service', 'error');
        }
    });
}

function showCreateIncidentModal() {
    const content = `
        <form id="create-incident-form">
            <div class="form-group">
                <label for="incident-title">Title</label>
                <input type="text" id="incident-title" name="title" required>
            </div>
            <div class="form-group">
                <label for="incident-description">Description</label>
                <textarea id="incident-description" name="description" rows="4" required></textarea>
            </div>
            <div class="form-group">
                <label for="incident-status">Status</label>
                <select id="incident-status" name="status">
                    <option value="investigating">Investigating</option>
                    <option value="identified">Identified</option>
                    <option value="monitoring">Monitoring</option>
                    <option value="resolved">Resolved</option>
                </select>
            </div>
            <div class="form-group">
                <label for="incident-severity">Severity</label>
                <select id="incident-severity" name="severity">
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="critical">Critical</option>
                </select>
            </div>
            <button type="submit" class="btn btn-primary">Create Incident</button>
        </form>
    `;
    showModal('Create Incident', content);
    
    document.getElementById('create-incident-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData.entries());
        
        try {
            const response = await tenantAdmin.authenticatedFetch('/api/v1/tenant/incidents', {
                method: 'POST',
                body: JSON.stringify(data)
            });
            
            if (response.ok) {
                closeModal();
                tenantAdmin.loadIncidents();
                tenantAdmin.showNotification('Incident created successfully', 'success');
            } else {
                tenantAdmin.showNotification('Failed to create incident', 'error');
            }
        } catch (error) {
            console.error('Failed to create incident:', error);
            tenantAdmin.showNotification('Failed to create incident', 'error');
        }
    });
}

// Initialize the application
let tenantAdmin;
document.addEventListener('DOMContentLoaded', () => {
    tenantAdmin = new TenantAdmin();
});
