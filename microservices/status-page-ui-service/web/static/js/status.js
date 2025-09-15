// Status Page JavaScript
class StatusPage {
    constructor() {
        this.socket = null;
        this.lastUpdate = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.startRealTimeUpdates();
        this.updateLastUpdatedTime();
    }

    setupEventListeners() {
        // Auto-refresh every 5 minutes
        setInterval(() => {
            this.refreshData();
        }, 5 * 60 * 1000);

        // Update last updated time every minute
        setInterval(() => {
            this.updateLastUpdatedTime();
        }, 60 * 1000);
    }

    startRealTimeUpdates() {
        // Initialize Socket.IO connection
        if (typeof io !== 'undefined') {
            this.socket = io();
            
            this.socket.on('connect', () => {
                console.log('Connected to real-time updates');
            });

            this.socket.on('disconnect', () => {
                console.log('Disconnected from real-time updates');
            });

            this.socket.on('status_update', (data) => {
                this.handleStatusUpdate(data);
            });

            this.socket.on('incident_update', (data) => {
                this.handleIncidentUpdate(data);
            });

            this.socket.on('maintenance_update', (data) => {
                this.handleMaintenanceUpdate(data);
            });

            this.socket.on('component_update', (data) => {
                this.handleComponentUpdate(data);
            });
        }
    }

    handleStatusUpdate(data) {
        console.log('Status update received:', data);
        
        // Update overall status
        if (data.overallStatus) {
            this.updateOverallStatus(data.overallStatus);
        }

        // Update components
        if (data.components) {
            this.updateComponents(data.components);
        }

        this.showNotification('Status updated', 'success');
    }

    handleIncidentUpdate(data) {
        console.log('Incident update received:', data);
        
        if (data.type === 'new') {
            this.addNewIncident(data.incident);
        } else if (data.type === 'update') {
            this.updateIncident(data.incident);
        } else if (data.type === 'resolved') {
            this.resolveIncident(data.incident);
        }

        this.showNotification('Incident update received', 'info');
    }

    handleMaintenanceUpdate(data) {
        console.log('Maintenance update received:', data);
        
        if (data.type === 'scheduled') {
            this.addScheduledMaintenance(data.maintenance);
        } else if (data.type === 'started') {
            this.startMaintenance(data.maintenance);
        } else if (data.type === 'completed') {
            this.completeMaintenance(data.maintenance);
        }

        this.showNotification('Maintenance update received', 'info');
    }

    handleComponentUpdate(data) {
        console.log('Component update received:', data);
        
        const componentElement = document.querySelector(`[data-component-id="${data.component.id}"]`);
        if (componentElement) {
            this.updateComponentCard(componentElement, data.component);
        }
    }

    updateOverallStatus(statusData) {
        const statusIndicator = document.querySelector('.status-indicator');
        if (!statusIndicator) return;

        // Update status class
        statusIndicator.className = `status-indicator ${statusData.status}`;
        
        // Update status title
        const statusTitle = statusIndicator.querySelector('.status-title');
        if (statusTitle) {
            statusTitle.textContent = statusData.title;
        }

        // Update status description
        const statusDescription = statusIndicator.querySelector('.status-description');
        if (statusDescription) {
            statusDescription.textContent = statusData.description;
        }

        // Update status icon
        const statusIcon = statusIndicator.querySelector('.status-icon');
        if (statusIcon) {
            statusIcon.innerHTML = this.getStatusIcon(statusData.status);
        }

        // Add animation
        statusIndicator.classList.add('fade-in');
        setTimeout(() => {
            statusIndicator.classList.remove('fade-in');
        }, 300);
    }

    updateComponents(components) {
        const componentsGrid = document.querySelector('.components-grid');
        if (!componentsGrid) return;

        components.forEach(component => {
            const componentElement = document.querySelector(`[data-component-id="${component.id}"]`);
            if (componentElement) {
                this.updateComponentCard(componentElement, component);
            }
        });
    }

    updateComponentCard(element, component) {
        // Update status
        element.className = `component-card ${component.status}`;
        
        // Update status text
        const statusText = element.querySelector('.status-text');
        if (statusText) {
            statusText.textContent = component.status.replace('_', ' ');
        }

        // Update uptime if available
        if (component.uptime) {
            const uptimeValue = element.querySelector('.uptime-value');
            if (uptimeValue) {
                uptimeValue.textContent = `${component.uptime}%`;
            }
        }

        // Add animation
        element.classList.add('slide-up');
        setTimeout(() => {
            element.classList.remove('slide-up');
        }, 300);
    }

    addNewIncident(incident) {
        const incidentsList = document.querySelector('.incidents-list');
        if (!incidentsList) return;

        const incidentElement = this.createIncidentElement(incident);
        incidentsList.insertBefore(incidentElement, incidentsList.firstChild);
        
        // Show incidents section if hidden
        const incidentsSection = document.querySelector('.incidents-section');
        if (incidentsSection && incidentsSection.style.display === 'none') {
            incidentsSection.style.display = 'block';
        }
    }

    updateIncident(incident) {
        const incidentElement = document.querySelector(`[data-incident-id="${incident.id}"]`);
        if (!incidentElement) return;

        // Update incident status
        incidentElement.className = `incident-card ${incident.status}`;
        
        // Update status text
        const statusText = incidentElement.querySelector('.status-text');
        if (statusText) {
            statusText.textContent = incident.status.replace('_', ' ');
        }

        // Add new update if available
        if (incident.latestUpdate) {
            const updatesContainer = incidentElement.querySelector('.incident-updates');
            if (updatesContainer) {
                const updateElement = this.createUpdateElement(incident.latestUpdate);
                updatesContainer.insertBefore(updateElement, updatesContainer.firstChild);
            }
        }
    }

    resolveIncident(incident) {
        const incidentElement = document.querySelector(`[data-incident-id="${incident.id}"]`);
        if (!incidentElement) return;

        // Update to resolved status
        incidentElement.className = 'incident-card resolved';
        
        const statusText = incidentElement.querySelector('.status-text');
        if (statusText) {
            statusText.textContent = 'Resolved';
        }

        // Add resolution update
        if (incident.resolutionUpdate) {
            const updatesContainer = incidentElement.querySelector('.incident-updates');
            if (updatesContainer) {
                const updateElement = this.createUpdateElement(incident.resolutionUpdate);
                updatesContainer.insertBefore(updateElement, updatesContainer.firstChild);
            }
        }
    }

    addScheduledMaintenance(maintenance) {
        const maintenanceList = document.querySelector('.maintenance-list');
        if (!maintenanceList) return;

        const maintenanceElement = this.createMaintenanceElement(maintenance);
        maintenanceList.appendChild(maintenanceElement);
        
        // Show maintenance section if hidden
        const maintenanceSection = document.querySelector('.maintenance-section');
        if (maintenanceSection && maintenanceSection.style.display === 'none') {
            maintenanceSection.style.display = 'block';
        }
    }

    startMaintenance(maintenance) {
        const maintenanceElement = document.querySelector(`[data-maintenance-id="${maintenance.id}"]`);
        if (!maintenanceElement) return;

        // Update maintenance status
        maintenanceElement.classList.add('in-progress');
        
        const statusElement = maintenanceElement.querySelector('.maintenance-status');
        if (statusElement) {
            statusElement.textContent = 'In Progress';
            statusElement.className = 'maintenance-status in-progress';
        }
    }

    completeMaintenance(maintenance) {
        const maintenanceElement = document.querySelector(`[data-maintenance-id="${maintenance.id}"]`);
        if (!maintenanceElement) return;

        // Update maintenance status
        maintenanceElement.classList.add('completed');
        
        const statusElement = maintenanceElement.querySelector('.maintenance-status');
        if (statusElement) {
            statusElement.textContent = 'Completed';
            statusElement.className = 'maintenance-status completed';
        }
    }

    createIncidentElement(incident) {
        const div = document.createElement('div');
        div.className = `incident-card ${incident.status}`;
        div.setAttribute('data-incident-id', incident.id);
        
        div.innerHTML = `
            <div class="incident-header">
                <div class="incident-status">
                    <span class="status-dot ${incident.status}"></span>
                    <span class="status-text">${incident.status.replace('_', ' ')}</span>
                </div>
                <div class="incident-date">${this.formatDate(incident.createdAt)}</div>
            </div>
            <div class="incident-title">${incident.title}</div>
            ${incident.description ? `<div class="incident-description">${incident.description}</div>` : ''}
            ${incident.updates ? `
                <div class="incident-updates">
                    ${incident.updates.map(update => this.createUpdateHTML(update)).join('')}
                </div>
            ` : ''}
        `;
        
        return div;
    }

    createMaintenanceElement(maintenance) {
        const div = document.createElement('div');
        div.className = 'maintenance-card';
        div.setAttribute('data-maintenance-id', maintenance.id);
        
        div.innerHTML = `
            <div class="maintenance-header">
                <div class="maintenance-title">${maintenance.title}</div>
                <div class="maintenance-date">${this.formatDate(maintenance.scheduledTime)}</div>
            </div>
            ${maintenance.description ? `<div class="maintenance-description">${maintenance.description}</div>` : ''}
            <div class="maintenance-impact">
                <span class="impact-label">Impact:</span>
                <span class="impact-value ${maintenance.impact}">${maintenance.impact.replace('_', ' ')}</span>
            </div>
        `;
        
        return div;
    }

    createUpdateElement(update) {
        const div = document.createElement('div');
        div.className = 'update-item';
        div.innerHTML = this.createUpdateHTML(update);
        return div;
    }

    createUpdateHTML(update) {
        return `
            <div class="update-time">${this.formatDate(update.createdAt)}</div>
            <div class="update-message">${update.message}</div>
        `;
    }

    getStatusIcon(status) {
        const icons = {
            operational: `
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M9 12l2 2 4-4"/>
                    <circle cx="12" cy="12" r="10"/>
                </svg>
            `,
            degraded: `
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 9v3m0 0v3m0-3h3m-3 0H9"/>
                    <circle cx="12" cy="12" r="10"/>
                </svg>
            `,
            partial_outage: `
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
                </svg>
            `,
            major_outage: `
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M6 18L18 6M6 6l12 12"/>
                </svg>
            `
        };
        
        return icons[status] || icons.major_outage;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString();
    }

    updateLastUpdatedTime() {
        const lastUpdatedElement = document.getElementById('last-updated');
        if (lastUpdatedElement) {
            lastUpdatedElement.textContent = new Date().toLocaleString();
        }
    }

    refreshData() {
        // Reload the page to get fresh data
        window.location.reload();
    }

    showNotification(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        
        // Add styles
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 1rem 1.5rem;
            border-radius: 0.5rem;
            color: white;
            font-weight: 500;
            z-index: 1000;
            animation: slideInRight 0.3s ease-out;
        `;
        
        // Set background color based on type
        const colors = {
            success: '#10b981',
            error: '#ef4444',
            warning: '#f59e0b',
            info: '#3b82f6'
        };
        notification.style.backgroundColor = colors[type] || colors.info;
        
        // Add to page
        document.body.appendChild(notification);
        
        // Remove after 3 seconds
        setTimeout(() => {
            notification.style.animation = 'slideOutRight 0.3s ease-in';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 3000);
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new StatusPage();
});

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideInRight {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOutRight {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

