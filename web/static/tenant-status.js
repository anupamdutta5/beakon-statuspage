class TenantStatusPage {
    constructor() {
        this.tenant = null;
        this.init();
    }

    async init() {
        await this.loadTenantInfo();
        await this.loadPageData();
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Subscribe form
        document.getElementById('subscribe-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleSubscribe(e);
        });

        // Incident modal
        document.getElementById('incident-modal').addEventListener('click', (e) => {
            if (e.target.id === 'incident-modal') {
                this.closeIncidentModal();
            }
        });
    }

    async loadTenantInfo() {
        try {
            const response = await fetch('/api/v1/tenant/info');
            if (response.ok) {
                const data = await response.json();
                this.tenant = data.tenant;
                this.updatePageInfo();
            }
        } catch (error) {
            console.error('Failed to load tenant info:', error);
        }
    }

    updatePageInfo() {
        if (!this.tenant) return;

        // Update page title
        document.title = `${this.tenant.name} - Status Page`;
        document.getElementById('page-title').textContent = `${this.tenant.name} - Status Page`;

        // Update company name
        document.getElementById('company-name').textContent = this.tenant.name;

        // Update logo if available
        if (this.tenant.logo_url) {
            const logo = document.getElementById('logo');
            logo.src = this.tenant.logo_url;
            logo.style.display = 'block';
        }

        // Update footer text
        if (this.tenant.footer_text) {
            document.getElementById('footer-text').textContent = this.tenant.footer_text;
        }

        // Apply custom colors
        if (this.tenant.primary_color) {
            document.documentElement.style.setProperty('--primary-color', this.tenant.primary_color);
        }
        if (this.tenant.secondary_color) {
            document.documentElement.style.setProperty('--secondary-color', this.tenant.secondary_color);
        }
    }

    async loadPageData() {
        await Promise.all([
            this.loadServices(),
            this.loadIncidents(),
            this.loadMaintenance(),
            this.loadUptimeStats()
        ]);
        this.updateOverallStatus();
    }

    async loadServices() {
        try {
            const response = await fetch('/api/v1/tenant/services');
            if (response.ok) {
                const data = await response.json();
                this.renderServices(data.services);
            }
        } catch (error) {
            console.error('Failed to load services:', error);
        }
    }

    renderServices(services) {
        const container = document.getElementById('services-grid');
        if (services.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-server"></i>
                    <h3>No Services</h3>
                    <p>No services configured</p>
                </div>
            `;
            return;
        }

        container.innerHTML = services.map(service => `
            <div class="service-card ${service.status}">
                <div class="service-icon">
                    <i class="fas fa-server"></i>
                </div>
                <div class="service-info">
                    <h3>${service.name}</h3>
                    <p>${service.description || 'No description available'}</p>
                </div>
                <div class="service-status ${service.status}">${service.status}</div>
            </div>
        `).join('');
    }

    async loadIncidents() {
        try {
            const response = await fetch('/api/v1/tenant/incidents');
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
            <div class="incident-item" onclick="showIncidentDetails(${incident.id})">
                <div class="incident-header">
                    <div class="incident-title">${incident.title}</div>
                    <div class="incident-date">${new Date(incident.created_at).toLocaleDateString()}</div>
                </div>
                <div class="incident-description">${incident.description}</div>
                <div class="incident-status ${incident.status}">${incident.status}</div>
            </div>
        `).join('');
    }

    async loadMaintenance() {
        try {
            const response = await fetch('/api/v1/tenant/maintenance');
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
            <div class="maintenance-item">
                <div class="maintenance-header">
                    <div class="maintenance-title">${item.title}</div>
                    <div class="maintenance-date">${new Date(item.scheduled_start).toLocaleDateString()}</div>
                </div>
                <div class="maintenance-description">${item.description}</div>
            </div>
        `).join('');
    }

    async loadUptimeStats() {
        try {
            const response = await fetch('/api/v1/tenant/uptime');
            if (response.ok) {
                const data = await response.json();
                document.getElementById('uptime-value').textContent = `${data.uptime_percentage}%`;
            }
        } catch (error) {
            console.error('Failed to load uptime stats:', error);
        }
    }

    updateOverallStatus() {
        const services = document.querySelectorAll('.service-card');
        const statusIndicator = document.getElementById('overall-status');
        const statusIcon = statusIndicator.querySelector('.status-icon i');
        const statusText = statusIndicator.querySelector('.status-text h2');
        const statusDescription = statusIndicator.querySelector('.status-text p');

        let hasOutage = false;
        let hasDegraded = false;

        services.forEach(service => {
            if (service.classList.contains('outage')) {
                hasOutage = true;
            } else if (service.classList.contains('degraded')) {
                hasDegraded = true;
            }
        });

        if (hasOutage) {
            statusIndicator.className = 'status-indicator outage';
            statusIcon.className = 'fas fa-exclamation-triangle';
            statusText.textContent = 'Service Outage';
            statusDescription.textContent = 'Some services are experiencing outages';
        } else if (hasDegraded) {
            statusIndicator.className = 'status-indicator degraded';
            statusIcon.className = 'fas fa-exclamation-circle';
            statusText.textContent = 'Degraded Performance';
            statusDescription.textContent = 'Some services are experiencing degraded performance';
        } else {
            statusIndicator.className = 'status-indicator operational';
            statusIcon.className = 'fas fa-check-circle';
            statusText.textContent = 'All Systems Operational';
            statusDescription.textContent = 'All services are running normally';
        }
    }

    async handleSubscribe(e) {
        const formData = new FormData(e.target);
        const email = formData.get('email');

        try {
            const response = await fetch('/api/v1/tenant/subscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email })
            });

            if (response.ok) {
                this.showNotification('Successfully subscribed to updates!', 'success');
                e.target.reset();
            } else {
                const error = await response.json();
                this.showNotification(error.error || 'Failed to subscribe', 'error');
            }
        } catch (error) {
            console.error('Failed to subscribe:', error);
            this.showNotification('Failed to subscribe', 'error');
        }
    }

    showNotification(message, type = 'success') {
        const notification = document.getElementById('notification');
        const notificationText = document.getElementById('notification-text');
        const icon = notification.querySelector('i');

        notificationText.textContent = message;
        notification.className = `notification ${type} show`;

        // Update icon based on type
        switch (type) {
            case 'success':
                icon.className = 'fas fa-check-circle';
                break;
            case 'error':
                icon.className = 'fas fa-exclamation-circle';
                break;
            case 'warning':
                icon.className = 'fas fa-exclamation-triangle';
                break;
            default:
                icon.className = 'fas fa-info-circle';
        }

        // Hide after 3 seconds
        setTimeout(() => {
            notification.classList.remove('show');
        }, 3000);
    }
}

// Global functions for incident modal
async function showIncidentDetails(incidentId) {
    try {
        const response = await fetch(`/api/v1/tenant/incidents/${incidentId}`);
        if (response.ok) {
            const data = await response.json();
            const incident = data.incident;
            
            document.getElementById('incident-modal-title').textContent = incident.title;
            document.getElementById('incident-modal-body').innerHTML = `
                <div class="incident-details">
                    <div class="incident-meta">
                        <div class="incident-status ${incident.status}">${incident.status}</div>
                        <div class="incident-date">${new Date(incident.created_at).toLocaleString()}</div>
                    </div>
                    <div class="incident-description">
                        <h4>Description</h4>
                        <p>${incident.description}</p>
                    </div>
                    ${incident.updates ? `
                        <div class="incident-updates">
                            <h4>Updates</h4>
                            ${incident.updates.map(update => `
                                <div class="update-item">
                                    <div class="update-date">${new Date(update.created_at).toLocaleString()}</div>
                                    <div class="update-content">${update.content}</div>
                                </div>
                            `).join('')}
                        </div>
                    ` : ''}
                </div>
            `;
            
            document.getElementById('incident-modal').classList.add('active');
        }
    } catch (error) {
        console.error('Failed to load incident details:', error);
    }
}

function closeIncidentModal() {
    document.getElementById('incident-modal').classList.remove('active');
}

// Initialize the application
let statusPage;
document.addEventListener('DOMContentLoaded', () => {
    statusPage = new TenantStatusPage();
});

// Auto-refresh every 30 seconds
setInterval(() => {
    if (statusPage) {
        statusPage.loadPageData();
    }
}, 30000);
