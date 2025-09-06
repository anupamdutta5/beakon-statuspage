// Atlassian Statuspage Public JavaScript

class PublicStatusPage {
    constructor() {
        this.currentFilter = 'all';
        this.currentPeriod = '24h';
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadPageData();
        this.setupModals();
    }

    setupEventListeners() {
        // Subscribe button
        document.getElementById('subscribeBtn')?.addEventListener('click', () => {
            this.openModal('subscribeModal');
            this.loadServicesForSubscription();
        });

        // Incident filters
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.setIncidentFilter(e.target.dataset.filter);
            });
        });

        // Uptime period buttons
        document.querySelectorAll('.period-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.setUptimePeriod(e.target.dataset.period);
            });
        });

        // Form submissions
        document.getElementById('subscribeForm')?.addEventListener('submit', (e) => {
            this.handleSubscription(e);
        });

        // Cancel buttons
        document.getElementById('cancelSubscribe')?.addEventListener('click', () => {
            this.closeModal('subscribeModal');
        });

        // All services checkbox
        document.getElementById('allServices')?.addEventListener('change', (e) => {
            this.toggleAllServices(e.target.checked);
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

    async loadPageData() {
        try {
            await Promise.all([
                this.loadComponents(),
                this.loadIncidents(),
                this.loadMaintenance(),
                this.loadUptimeStats()
            ]);
        } catch (error) {
            console.error('Error loading page data:', error);
            this.showNotification('Failed to load status information', 'error');
        }
    }

    async loadComponents() {
        try {
            const response = await fetch('/api/v1/status');
            const data = await response.json();
            
            const componentsGrid = document.getElementById('componentsGrid');
            if (!componentsGrid) return;

            componentsGrid.innerHTML = '';

            if (data.services && data.services.length > 0) {
                data.services.forEach(service => {
                    const componentCard = this.createComponentCard(service);
                    componentsGrid.appendChild(componentCard);
                });
            } else {
                componentsGrid.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-cogs"></i>
                        <h3>No Components</h3>
                        <p>No services are currently being monitored</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading components:', error);
        }
    }

    createComponentCard(service) {
        const card = document.createElement('div');
        card.className = 'component-card fade-in';
        
        const statusClass = service.status.replace('_', '-');
        const statusText = service.status.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
        
        card.innerHTML = `
            <div class="component-header">
                <div class="component-name">${service.name}</div>
                <div class="component-status ${statusClass}">${statusText}</div>
            </div>
            <div class="component-description">${service.description || 'No description available'}</div>
            ${service.show_uptime ? `
                <div class="component-uptime">
                    <i class="fas fa-chart-line"></i>
                    <span>Uptime: <span class="uptime-value">99.9%</span></span>
                </div>
            ` : ''}
        `;
        
        return card;
    }

    async loadIncidents() {
        try {
            const response = await fetch('/api/v1/incidents');
            const incidents = await response.json();
            
            const incidentsList = document.getElementById('incidentsList');
            if (!incidentsList) return;

            this.allIncidents = incidents || [];
            this.filterIncidents();
        } catch (error) {
            console.error('Error loading incidents:', error);
        }
    }

    filterIncidents() {
        const incidentsList = document.getElementById('incidentsList');
        if (!incidentsList || !this.allIncidents) return;

        let filteredIncidents = this.allIncidents;

        if (this.currentFilter === 'active') {
            filteredIncidents = this.allIncidents.filter(incident => 
                ['investigating', 'identified', 'monitoring'].includes(incident.Status.toLowerCase())
            );
        } else if (this.currentFilter === 'resolved') {
            filteredIncidents = this.allIncidents.filter(incident => 
                incident.Status.toLowerCase() === 'resolved'
            );
        }

        incidentsList.innerHTML = '';

        if (filteredIncidents.length > 0) {
            filteredIncidents.forEach(incident => {
                const incidentCard = this.createIncidentCard(incident);
                incidentsList.appendChild(incidentCard);
            });
        } else {
            incidentsList.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h3>No ${this.currentFilter === 'all' ? '' : this.currentFilter} incidents</h3>
                    <p>All systems are running smoothly</p>
                </div>
            `;
        }
    }

    createIncidentCard(incident) {
        const card = document.createElement('div');
        card.className = 'incident-card fade-in';
        
        const createdAt = new Date(incident.CreatedAt);
        const timeAgo = this.getTimeAgo(createdAt);
        const statusClass = incident.Status.toLowerCase();
        const statusText = incident.Status.charAt(0).toUpperCase() + incident.Status.slice(1);
        
        card.innerHTML = `
            <div class="incident-header">
                <div class="incident-title">${incident.Title}</div>
                <div class="incident-status ${statusClass}">${statusText}</div>
            </div>
            <div class="incident-description">${incident.Description || 'No description available'}</div>
            <div class="incident-meta">
                <div class="incident-time">
                    <i class="fas fa-clock"></i>
                    <span>${timeAgo}</span>
                </div>
                ${incident.Impact ? `
                    <div class="incident-impact">
                        <span class="impact-indicator ${incident.Impact.toLowerCase()}"></span>
                        <span>${incident.Impact.charAt(0).toUpperCase() + incident.Impact.slice(1)} Impact</span>
                    </div>
                ` : ''}
            </div>
        `;
        
        // Add click handler for incident details
        card.addEventListener('click', () => {
            this.showIncidentDetails(incident);
        });
        
        return card;
    }

    async loadMaintenance() {
        try {
            const response = await fetch('/api/v1/maintenance');
            const maintenanceEvents = await response.json();
            
            const maintenanceList = document.getElementById('maintenanceList');
            if (!maintenanceList) return;

            maintenanceList.innerHTML = '';

            if (maintenanceEvents && maintenanceEvents.length > 0) {
                maintenanceEvents.forEach(event => {
                    const maintenanceCard = this.createMaintenanceCard(event);
                    maintenanceList.appendChild(maintenanceCard);
                });
            } else {
                maintenanceList.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-tools"></i>
                        <h3>No Scheduled Maintenance</h3>
                        <p>No maintenance events are currently scheduled</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading maintenance events:', error);
        }
    }

    createMaintenanceCard(event) {
        const card = document.createElement('div');
        card.className = 'maintenance-card fade-in';
        
        const startAt = new Date(event.StartAt);
        const endAt = new Date(event.EndAt);
        const startTime = startAt.toLocaleString();
        const endTime = endAt.toLocaleString();
        const duration = this.getDuration(startAt, endAt);
        
        card.innerHTML = `
            <div class="maintenance-header">
                <div class="maintenance-title">${event.Title}</div>
                <div class="maintenance-status">${event.Status}</div>
            </div>
            <div class="maintenance-description">${event.Description || 'No description available'}</div>
            <div class="maintenance-schedule">
                <div class="schedule-time">
                    <i class="fas fa-calendar"></i>
                    <span>Starts: ${startTime}</span>
                </div>
                <div class="schedule-time">
                    <i class="fas fa-clock"></i>
                    <span>Duration: ${duration}</span>
                </div>
            </div>
        `;
        
        return card;
    }

    async loadUptimeStats() {
        try {
            const response = await fetch(`/api/v1/uptime/history/1?period=${this.currentPeriod}`);
            const stats = await response.json();
            
            const uptimeStats = document.getElementById('uptimeStats');
            if (!uptimeStats) return;

            uptimeStats.innerHTML = '';

            if (stats && stats.length > 0) {
                const overallUptime = this.calculateOverallUptime(stats);
                
                uptimeStats.innerHTML = `
                    <div class="uptime-stat-card fade-in">
                        <div class="uptime-stat-value">${overallUptime}%</div>
                        <div class="uptime-stat-label">Overall Uptime</div>
                    </div>
                    <div class="uptime-stat-card fade-in">
                        <div class="uptime-stat-value">${stats.length}</div>
                        <div class="uptime-stat-label">Total Checks</div>
                    </div>
                    <div class="uptime-stat-card fade-in">
                        <div class="uptime-stat-value">${this.calculateMTTR(stats)}</div>
                        <div class="uptime-stat-label">Avg Response Time</div>
                    </div>
                `;
            } else {
                uptimeStats.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-chart-line"></i>
                        <h3>No Uptime Data</h3>
                        <p>Uptime statistics are not available for this period</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading uptime stats:', error);
        }
    }

    async loadServicesForSubscription() {
        try {
            const response = await fetch('/api/v1/status');
            const data = await response.json();
            
            const servicesList = document.getElementById('servicesList');
            if (!servicesList) return;

            servicesList.innerHTML = '';

            if (data.services && data.services.length > 0) {
                data.services.forEach(service => {
                    const checkbox = document.createElement('label');
                    checkbox.innerHTML = `
                        <input type="checkbox" name="services" value="${service.ID}">
                        ${service.name}
                    `;
                    servicesList.appendChild(checkbox);
                });
            } else {
                servicesList.innerHTML = '<p>No services available for subscription.</p>';
            }
        } catch (error) {
            console.error('Error loading services for subscription:', error);
        }
    }

    async handleSubscription(e) {
        e.preventDefault();
        
        const email = document.getElementById('email').value;
        const selectedServices = Array.from(document.querySelectorAll('input[name="services"]:checked'))
            .map(cb => parseInt(cb.value));
        const submitButton = e.target.querySelector('button[type="submit"]');
        const messageElement = document.getElementById('subscriptionMessage');

        submitButton.disabled = true;
        messageElement.textContent = '';

        try {
            const response = await fetch('/api/v1/subscribers', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    email: email, 
                    services: selectedServices.length > 0 ? selectedServices : null 
                }),
            });

            if (response.ok) {
                messageElement.textContent = 'Successfully subscribed! You will receive email notifications for updates.';
                messageElement.className = 'message success';
                setTimeout(() => {
                    this.closeModal('subscribeModal');
                    messageElement.textContent = '';
                    messageElement.className = '';
                }, 3000);
            } else {
                const errorData = await response.json();
                messageElement.textContent = `Subscription failed: ${errorData.error}`;
                messageElement.className = 'message error';
            }
        } catch (error) {
            console.error('Error:', error);
            messageElement.textContent = 'An error occurred. Please try again.';
            messageElement.className = 'message error';
        } finally {
            submitButton.disabled = false;
        }
    }

    toggleAllServices(checked) {
        const checkboxes = document.querySelectorAll('input[name="services"]');
        checkboxes.forEach(checkbox => {
            checkbox.checked = checked;
        });
    }

    setIncidentFilter(filter) {
        this.currentFilter = filter;
        
        // Update active button
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-filter="${filter}"]`).classList.add('active');
        
        this.filterIncidents();
    }

    setUptimePeriod(period) {
        this.currentPeriod = period;
        
        // Update active button
        document.querySelectorAll('.period-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-period="${period}"]`).classList.add('active');
        
        this.loadUptimeStats();
    }

    showIncidentDetails(incident) {
        const modal = document.getElementById('incidentModal');
        const title = document.getElementById('incidentModalTitle');
        const body = document.getElementById('incidentModalBody');
        
        if (!modal || !title || !body) return;
        
        title.textContent = incident.Title;
        
        const createdAt = new Date(incident.CreatedAt);
        const statusClass = incident.Status.toLowerCase();
        const statusText = incident.Status.charAt(0).toUpperCase() + incident.Status.slice(1);
        
        body.innerHTML = `
            <div class="incident-details">
                <div class="incident-meta">
                    <div class="incident-status ${statusClass}">${statusText}</div>
                    <div class="incident-time">${createdAt.toLocaleString()}</div>
                </div>
                <div class="incident-description">
                    <h3>Description</h3>
                    <p>${incident.Description || 'No description available'}</p>
                </div>
                ${incident.Impact ? `
                    <div class="incident-impact">
                        <h3>Impact</h3>
                        <p>${incident.Impact.charAt(0).toUpperCase() + incident.Impact.slice(1)} Impact</p>
                    </div>
                ` : ''}
            </div>
        `;
        
        this.openModal('incidentModal');
    }

    calculateOverallUptime(stats) {
        if (!stats || stats.length === 0) return '99.9';
        
        const successfulChecks = stats.filter(stat => stat.status === 'success').length;
        const uptime = (successfulChecks / stats.length) * 100;
        return uptime.toFixed(1);
    }

    calculateMTTR(stats) {
        if (!stats || stats.length === 0) return '0ms';
        
        const avgResponseTime = stats.reduce((sum, stat) => sum + (stat.response_time || 0), 0) / stats.length;
        return `${Math.round(avgResponseTime)}ms`;
    }

    getTimeAgo(date) {
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} minutes ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
        return `${Math.floor(diffInSeconds / 86400)} days ago`;
    }

    getDuration(start, end) {
        const diffInMinutes = Math.floor((end - start) / (1000 * 60));
        
        if (diffInMinutes < 60) return `${diffInMinutes} minutes`;
        if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)} hours`;
        return `${Math.floor(diffInMinutes / 1440)} days`;
    }

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
}

// Initialize the page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.publicStatusPage = new PublicStatusPage();
});

// Export for global access
window.PublicStatusPage = PublicStatusPage;
