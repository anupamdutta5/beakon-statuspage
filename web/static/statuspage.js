// StatusPage.io Style JavaScript

class StatusPage {
    constructor() {
        this.currentTab = 'status';
        this.charts = {};
        this.init();
    }

    init() {
        console.log('StatusPage initialized');
        this.setupEventListeners();
        this.loadInitialData();
        this.setupCharts();
    }

    setupEventListeners() {
        // Tab navigation
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.addEventListener('click', (e) => {
                const tabName = e.target.dataset.tab;
                this.switchTab(tabName);
            });
        });

        // Subscribe modal
        const subscribeBtn = document.getElementById('subscribe-btn');
        const closeModal = document.getElementById('close-modal');
        const modal = document.getElementById('subscribe-modal');

        subscribeBtn.addEventListener('click', () => {
            modal.style.display = 'block';
            this.loadServicesForSubscription();
        });

        closeModal.addEventListener('click', () => {
            modal.style.display = 'none';
        });

        window.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });

        // SMS notifications toggle
        const smsCheckbox = document.getElementById('sms-notifications');
        const phoneGroup = document.getElementById('phone-group');
        
        smsCheckbox.addEventListener('change', (e) => {
            phoneGroup.style.display = e.target.checked ? 'block' : 'none';
        });

        // Subscribe form
        const subscribeForm = document.getElementById('subscribe-form');
        subscribeForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleSubscription(e);
        });

        // Filter changes
        const incidentFilter = document.getElementById('incident-status-filter');
        if (incidentFilter) {
            incidentFilter.addEventListener('change', () => {
                this.filterIncidents();
            });
        }

        const monitoringFilter = document.getElementById('monitoring-timeframe');
        if (monitoringFilter) {
            monitoringFilter.addEventListener('change', () => {
                this.updateMonitoringCharts();
            });
        }

        const historyFilter = document.getElementById('history-timeframe');
        if (historyFilter) {
            historyFilter.addEventListener('change', () => {
                this.loadStatusHistory();
            });
        }
    }

    switchTab(tabName) {
        // Update active tab
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update active content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabName}-tab`).classList.add('active');

        this.currentTab = tabName;

        // Load tab-specific data
        switch (tabName) {
            case 'status':
                this.loadServices();
                break;
            case 'incidents':
                this.loadIncidents();
                break;
            case 'maintenance':
                this.loadMaintenance();
                break;
            case 'monitoring':
                this.updateMonitoringCharts();
                break;
            case 'history':
                this.loadStatusHistory();
                break;
        }
    }

    async loadInitialData() {
        await Promise.all([
            this.loadServices(),
            this.loadIncidents(),
            this.loadMaintenance(),
            this.loadOverallStatus(),
            this.loadFeatureFlags()
        ]);
    }

    async loadFeatureFlags() {
        try {
            const response = await fetch('/api/v1/admin/feature-flags');
            const data = await response.json();
            this.featureFlags = data.feature_flags || [];
            console.log('Feature flags loaded:', this.featureFlags);
        } catch (error) {
            console.error('Error loading feature flags:', error);
            this.featureFlags = [];
        }
    }

    isFeatureEnabled(featureName) {
        const flag = this.featureFlags.find(f => f.feature === featureName);
        return flag ? flag.is_enabled : false;
    }

    async loadOverallStatus() {
        try {
            const response = await fetch('/api/v1/status');
            const data = await response.json();
            
            this.updateOverallStatus(data);
        } catch (error) {
            console.error('Error loading overall status:', error);
        }
    }

    updateOverallStatus(data) {
        const statusBadge = document.getElementById('overall-status');
        const statusIndicator = statusBadge.querySelector('.status-indicator');
        const statusText = statusBadge.querySelector('.status-text');

        if (data.status === 'operational') {
            statusIndicator.className = 'status-indicator operational';
            statusText.textContent = 'All Systems Operational';
            statusText.style.color = '#16a34a';
        } else if (data.status === 'degraded') {
            statusIndicator.className = 'status-indicator degraded';
            statusText.textContent = 'Degraded Performance';
            statusText.style.color = '#d97706';
        } else {
            statusIndicator.className = 'status-indicator outage';
            statusText.textContent = 'Service Outage';
            statusText.style.color = '#dc2626';
        }
    }

    async loadServices() {
        try {
            const response = await fetch('/api/v1/services');
            const services = await response.json();
            
            this.renderServices(services);
        } catch (error) {
            console.error('Error loading services:', error);
        }
    }

    renderServices(services) {
        const servicesList = document.getElementById('services-list');
        servicesList.innerHTML = '';

        services.forEach(service => {
            const serviceItem = document.createElement('div');
            serviceItem.className = 'service-item';
            
            const statusClass = service.status === 'operational' ? 'operational' : 
                               service.status === 'degraded' ? 'degraded' : 'outage';
            
            serviceItem.innerHTML = `
                <div class="service-info">
                    <span class="service-status ${statusClass}"></span>
                    <span class="service-name">${service.name}</span>
                </div>
                <div class="service-uptime">${service.uptime || '99.9%'} uptime</div>
            `;
            
            servicesList.appendChild(serviceItem);
        });
    }

    async loadIncidents() {
        try {
            const response = await fetch('/api/v1/incidents');
            const incidents = await response.json();
            
            this.renderIncidents(incidents);
        } catch (error) {
            console.error('Error loading incidents:', error);
        }
    }

    renderIncidents(incidents) {
        const incidentsList = document.getElementById('incidents-list');
        incidentsList.innerHTML = '';

        if (incidents.length === 0) {
            incidentsList.innerHTML = '<p style="text-align: center; color: #64748b; padding: 40px;">No incidents reported.</p>';
            return;
        }

        incidents.forEach(incident => {
            const incidentItem = document.createElement('div');
            incidentItem.className = 'incident-item';
            
            const createdDate = new Date(incident.created_at).toLocaleDateString();
            const updatedDate = new Date(incident.updated_at).toLocaleDateString();
            
            incidentItem.innerHTML = `
                <div class="incident-header">
                    <div class="incident-title">${incident.title}</div>
                    <span class="incident-status ${incident.status}">${incident.status}</span>
                </div>
                <div class="incident-content">
                    <div class="incident-description">${incident.description}</div>
                    <div class="incident-meta">
                        <span>Created: ${createdDate}</span>
                        <span>Updated: ${updatedDate}</span>
                        <span>Impact: ${incident.impact}</span>
                    </div>
                </div>
            `;
            
            incidentsList.appendChild(incidentItem);
        });
    }

    async loadMaintenance() {
        try {
            const response = await fetch('/api/v1/maintenance');
            const maintenance = await response.json();
            
            this.renderMaintenance(maintenance);
        } catch (error) {
            console.error('Error loading maintenance:', error);
        }
    }

    renderMaintenance(maintenance) {
        const maintenanceList = document.getElementById('maintenance-list');
        maintenanceList.innerHTML = '';

        if (maintenance.length === 0) {
            maintenanceList.innerHTML = '<p style="text-align: center; color: #64748b; padding: 40px;">No scheduled maintenance.</p>';
            return;
        }

        maintenance.forEach(item => {
            const maintenanceItem = document.createElement('div');
            maintenanceItem.className = 'incident-item';
            
            const startDate = new Date(item.start_at).toLocaleString();
            const endDate = new Date(item.end_at).toLocaleString();
            
            maintenanceItem.innerHTML = `
                <div class="incident-header">
                    <div class="incident-title">${item.title}</div>
                    <span class="incident-status monitoring">Scheduled</span>
                </div>
                <div class="incident-content">
                    <div class="incident-description">${item.description}</div>
                    <div class="incident-meta">
                        <span>Start: ${startDate}</span>
                        <span>End: ${endDate}</span>
                    </div>
                </div>
            `;
            
            maintenanceList.appendChild(maintenanceItem);
        });
    }

    setupCharts() {
        // Uptime Chart
        const uptimeCtx = document.getElementById('uptime-chart');
        if (uptimeCtx) {
            this.charts.uptime = new Chart(uptimeCtx, {
                type: 'line',
                data: {
                    labels: this.generateTimeLabels(24),
                    datasets: [{
                        label: 'Uptime %',
                        data: this.generateUptimeData(24),
                        borderColor: '#22c55e',
                        backgroundColor: 'rgba(34, 197, 94, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100,
                            ticks: {
                                callback: function(value) {
                                    return value + '%';
                                }
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }

        // Response Time Chart
        const responseCtx = document.getElementById('response-time-chart');
        if (responseCtx) {
            this.charts.responseTime = new Chart(responseCtx, {
                type: 'line',
                data: {
                    labels: this.generateTimeLabels(24),
                    datasets: [{
                        label: 'Response Time (ms)',
                        data: this.generateResponseTimeData(24),
                        borderColor: '#3b82f6',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            ticks: {
                                callback: function(value) {
                                    return value + 'ms';
                                }
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }

        // Error Rate Chart
        const errorCtx = document.getElementById('error-rate-chart');
        if (errorCtx) {
            this.charts.errorRate = new Chart(errorCtx, {
                type: 'line',
                data: {
                    labels: this.generateTimeLabels(24),
                    datasets: [{
                        label: 'Error Rate %',
                        data: this.generateErrorRateData(24),
                        borderColor: '#ef4444',
                        backgroundColor: 'rgba(239, 68, 68, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 5,
                            ticks: {
                                callback: function(value) {
                                    return value + '%';
                                }
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    }
                }
            });
        }
    }

    generateTimeLabels(hours) {
        const labels = [];
        const now = new Date();
        for (let i = hours - 1; i >= 0; i--) {
            const time = new Date(now.getTime() - (i * 60 * 60 * 1000));
            labels.push(time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' }));
        }
        return labels;
    }

    generateUptimeData(hours) {
        const data = [];
        for (let i = 0; i < hours; i++) {
            // Generate realistic uptime data (99.5% - 100%)
            data.push(99.5 + Math.random() * 0.5);
        }
        return data;
    }

    generateResponseTimeData(hours) {
        const data = [];
        for (let i = 0; i < hours; i++) {
            // Generate realistic response time data (20ms - 100ms)
            data.push(20 + Math.random() * 80);
        }
        return data;
    }

    generateErrorRateData(hours) {
        const data = [];
        for (let i = 0; i < hours; i++) {
            // Generate realistic error rate data (0% - 2%)
            data.push(Math.random() * 2);
        }
        return data;
    }

    updateMonitoringCharts() {
        const timeframe = document.getElementById('monitoring-timeframe')?.value || '24h';
        const hours = timeframe === '24h' ? 24 : timeframe === '7d' ? 168 : 720;
        
        // Check if per-service graphs feature is enabled
        if (this.isFeatureEnabled('per_service_graphs')) {
            this.loadPerServiceCharts(hours);
        } else {
            // Update general chart data
            Object.values(this.charts).forEach(chart => {
                chart.data.labels = this.generateTimeLabels(hours);
                chart.data.datasets[0].data = this.generateUptimeData(hours);
                chart.update();
            });
        }
    }

    async loadPerServiceCharts(hours) {
        const monitoringContainer = document.getElementById('monitoring-charts');
        if (!monitoringContainer) return;

        try {
            // Load services data
            const servicesResponse = await fetch('/api/v1/services');
            const servicesData = await servicesResponse.json();
            const services = servicesData.services || [];

            // Clear existing charts
            monitoringContainer.innerHTML = '';

            // Create charts for each service
            for (const service of services) {
                const serviceChartContainer = document.createElement('div');
                serviceChartContainer.className = 'service-chart-container';
                serviceChartContainer.innerHTML = `
                    <div class="service-chart-header">
                        <h3>${service.name}</h3>
                        <div class="service-status ${service.status}">
                            <span class="status-indicator"></span>
                            ${service.status.charAt(0).toUpperCase() + service.status.slice(1)}
                        </div>
                    </div>
                    <div class="chart-wrapper">
                        <canvas id="chart-${service.id}" width="400" height="200"></canvas>
                    </div>
                `;
                monitoringContainer.appendChild(serviceChartContainer);

                // Create chart for this service
                const ctx = document.getElementById(`chart-${service.id}`);
                if (ctx) {
                    const chart = new Chart(ctx, {
                        type: 'line',
                        data: {
                            labels: this.generateTimeLabels(hours),
                            datasets: [{
                                label: 'Uptime %',
                                data: this.generateServiceUptimeData(service.id, hours),
                                borderColor: this.getStatusColor(service.status),
                                backgroundColor: this.getStatusColor(service.status, 0.1),
                                tension: 0.4,
                                fill: true
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            scales: {
                                y: {
                                    beginAtZero: true,
                                    max: 100,
                                    ticks: {
                                        callback: function(value) {
                                            return value + '%';
                                        }
                                    }
                                }
                            },
                            plugins: {
                                legend: {
                                    display: false
                                }
                            }
                        }
                    });
                    this.charts[`service-${service.id}`] = chart;
                }
            }
        } catch (error) {
            console.error('Error loading per-service charts:', error);
        }
    }

    generateServiceUptimeData(serviceId, hours) {
        // Generate mock uptime data for a specific service
        const data = [];
        const now = new Date();
        
        for (let i = hours; i >= 0; i--) {
            const timestamp = new Date(now.getTime() - (i * 60 * 60 * 1000));
            // Simulate some variation in uptime
            const baseUptime = 95 + Math.random() * 5;
            const variation = (Math.sin(timestamp.getHours() / 24 * Math.PI * 2) * 2);
            data.push(Math.max(0, Math.min(100, baseUptime + variation)));
        }
        
        return data;
    }

    getStatusColor(status, alpha = 1) {
        const colors = {
            'operational': `rgba(40, 167, 69, ${alpha})`,
            'degraded_performance': `rgba(255, 193, 7, ${alpha})`,
            'partial_outage': `rgba(220, 53, 69, ${alpha})`,
            'major_outage': `rgba(220, 53, 69, ${alpha})`,
            'under_maintenance': `rgba(108, 117, 125, ${alpha})`
        };
        return colors[status] || `rgba(108, 117, 125, ${alpha})`;
    }

    async loadServicesForSubscription() {
        try {
            const response = await fetch('/api/v1/services');
            const services = await response.json();
            
            const checkboxesContainer = document.getElementById('services-checkboxes');
            checkboxesContainer.innerHTML = '';
            
            services.forEach(service => {
                const checkbox = document.createElement('div');
                checkbox.className = 'service-checkbox';
                checkbox.innerHTML = `
                    <input type="checkbox" id="service-${service.id}" name="services" value="${service.id}" checked>
                    <label for="service-${service.id}">${service.name}</label>
                `;
                checkboxesContainer.appendChild(checkbox);
            });
        } catch (error) {
            console.error('Error loading services for subscription:', error);
        }
    }

    async handleSubscription(e) {
        const formData = new FormData(e.target);
        const subscriptionData = {
            email: formData.get('email'),
            phone: formData.get('phone'),
            services: formData.getAll('services'),
            sms: formData.get('sms') === 'on'
        };

        try {
            const response = await fetch('/api/v1/subscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(subscriptionData)
            });

            if (response.ok) {
                alert('Successfully subscribed to updates!');
                document.getElementById('subscribe-modal').style.display = 'none';
                e.target.reset();
            } else {
                throw new Error('Subscription failed');
            }
        } catch (error) {
            console.error('Error subscribing:', error);
            alert('Failed to subscribe. Please try again.');
        }
    }

    filterIncidents() {
        const filter = document.getElementById('incident-status-filter').value;
        const incidents = document.querySelectorAll('.incident-item');
        
        incidents.forEach(incident => {
            const status = incident.querySelector('.incident-status').textContent.toLowerCase();
            if (filter === 'all' || status === filter) {
                incident.style.display = 'block';
            } else {
                incident.style.display = 'none';
            }
        });
    }

    async loadStatusHistory() {
        // This would load historical status data
        // For now, we'll show a placeholder
        const timeline = document.getElementById('history-timeline');
        timeline.innerHTML = '<p style="text-align: center; color: #64748b; padding: 40px;">Status history will be displayed here.</p>';
    }
}

// Initialize the status page when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    window.statusPage = new StatusPage();
});
