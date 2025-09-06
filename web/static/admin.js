// Atlassian Statuspage Admin Dashboard JavaScript

class AdminDashboard {
    constructor() {
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadDashboardData();
        this.setupModals();
    }

    setupEventListeners() {
        // Sidebar toggle for mobile
        const sidebarToggle = document.getElementById('sidebarToggle');
        const sidebar = document.querySelector('.sidebar');
        
        if (sidebarToggle) {
            sidebarToggle.addEventListener('click', () => {
                sidebar.classList.toggle('open');
            });
        }

        // Quick action buttons
        document.getElementById('createIncidentBtn')?.addEventListener('click', () => {
            this.openModal('incidentModal');
            this.loadComponentsForModal('affectedComponents');
        });

        document.getElementById('scheduleMaintenanceBtn')?.addEventListener('click', () => {
            this.openModal('maintenanceModal');
            this.loadComponentsForModal('maintenanceComponents');
        });

        document.getElementById('addComponentBtn')?.addEventListener('click', () => {
            this.openModal('componentModal');
        });

        document.getElementById('addMonitorBtn')?.addEventListener('click', () => {
            this.openModal('monitorModal');
            this.loadServicesForMonitor();
        });

        // Form submissions
        document.getElementById('incidentForm')?.addEventListener('submit', (e) => {
            this.handleIncidentSubmit(e);
        });

        document.getElementById('componentForm')?.addEventListener('submit', (e) => {
            this.handleComponentSubmit(e);
        });

        document.getElementById('monitorForm')?.addEventListener('submit', (e) => {
            this.handleMonitorSubmit(e);
        });

        document.getElementById('maintenanceForm')?.addEventListener('submit', (e) => {
            this.handleMaintenanceSubmit(e);
        });

        // Cancel buttons
        document.getElementById('cancelIncident')?.addEventListener('click', () => {
            this.closeModal('incidentModal');
        });

        document.getElementById('cancelComponent')?.addEventListener('click', () => {
            this.closeModal('componentModal');
        });

        document.getElementById('cancelMonitor')?.addEventListener('click', () => {
            this.closeModal('monitorModal');
        });

        document.getElementById('cancelMaintenance')?.addEventListener('click', () => {
            this.closeModal('maintenanceModal');
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

    async loadDashboardData() {
        try {
            await Promise.all([
                this.loadStats(),
                this.loadComponents(),
                this.loadRecentIncidents(),
                this.loadRecentActivity()
            ]);
        } catch (error) {
            console.error('Error loading dashboard data:', error);
            this.showNotification('Failed to load dashboard data', 'error');
        }
    }

    async loadStats() {
        try {
            const [incidentsRes, maintenanceRes, subscribersRes] = await Promise.all([
                fetch('/api/v1/incidents/active', { credentials: 'include' }),
                fetch('/api/v1/maintenance/upcoming', { credentials: 'include' }),
                fetch('/api/v1/admin/subscribers', { credentials: 'include' })
            ]);

            const activeIncidents = await incidentsRes.json();
            const upcomingMaintenance = await maintenanceRes.json();
            const subscribers = await subscribersRes.json();

            document.getElementById('activeIncidents').textContent = activeIncidents.length || 0;
            document.getElementById('upcomingMaintenance').textContent = upcomingMaintenance.length || 0;
            document.getElementById('totalSubscribers').textContent = subscribers.length || 0;

        } catch (error) {
            console.error('Error loading stats:', error);
        }
    }

    async loadComponents() {
        try {
            const response = await fetch('/api/v1/status', { credentials: 'include' });
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
                        <p>Add your first component to get started</p>
                        <button class="btn-primary" onclick="adminDashboard.openModal('componentModal')">
                            Add Component
                        </button>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading components:', error);
        }
    }

    createComponentCard(service) {
        const card = document.createElement('div');
        card.className = 'component-card';
        
        const statusClass = service.status.replace('_', '-');
        const statusText = service.status.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
        
        card.innerHTML = `
            <div class="component-header">
                <div class="component-name">
                    <span class="status-indicator ${statusClass}"></span>
                    ${service.name}
                </div>
                <div class="component-status ${statusClass}">${statusText}</div>
            </div>
            <div class="component-description">${service.description || 'No description'}</div>
        `;
        
        return card;
    }

    async loadRecentIncidents() {
        try {
            const response = await fetch('/api/v1/incidents', { credentials: 'include' });
            const incidents = await response.json();
            
            const incidentsList = document.getElementById('recentIncidentsList');
            if (!incidentsList) return;

            incidentsList.innerHTML = '';

            if (incidents && incidents.length > 0) {
                const recentIncidents = incidents.slice(0, 5);
                recentIncidents.forEach(incident => {
                    const incidentCard = this.createIncidentCard(incident);
                    incidentsList.appendChild(incidentCard);
                });
            } else {
                incidentsList.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-exclamation-triangle"></i>
                        <h3>No Incidents</h3>
                        <p>All systems are running smoothly</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading recent incidents:', error);
        }
    }

    createIncidentCard(incident) {
        const card = document.createElement('div');
        card.className = 'incident-card';
        
        const createdAt = new Date(incident.CreatedAt).toLocaleDateString();
        const statusClass = incident.Status.toLowerCase();
        const statusText = incident.Status.charAt(0).toUpperCase() + incident.Status.slice(1);
        
        card.innerHTML = `
            <div class="incident-header">
                <div class="incident-title">${incident.Title}</div>
                <div class="incident-status ${statusClass}">${statusText}</div>
            </div>
            <div class="incident-description">${incident.Description || 'No description'}</div>
            <div class="incident-time">${createdAt}</div>
        `;
        
        return card;
    }

    async loadRecentActivity() {
        try {
            const response = await fetch('/api/v1/admin/audit-logs', { credentials: 'include' });
            const logs = await response.json();
            
            const activityList = document.getElementById('activityList');
            if (!activityList) return;

            // Keep the first activity item (system status)
            const systemStatusItem = activityList.querySelector('.activity-item');
            activityList.innerHTML = '';
            activityList.appendChild(systemStatusItem);

            if (logs && logs.length > 0) {
                const recentLogs = logs.slice(0, 5);
                recentLogs.forEach(log => {
                    const activityItem = this.createActivityItem(log);
                    activityList.appendChild(activityItem);
                });
            }
        } catch (error) {
            console.error('Error loading recent activity:', error);
        }
    }

    createActivityItem(log) {
        const item = document.createElement('div');
        item.className = 'activity-item';
        
        const timeAgo = this.getTimeAgo(new Date(log.CreatedAt));
        
        item.innerHTML = `
            <div class="activity-icon">
                <i class="fas fa-user"></i>
            </div>
            <div class="activity-content">
                <p><strong>${log.Action}</strong> ${log.Details}</p>
                <span class="activity-time">${timeAgo}</span>
            </div>
        `;
        
        return item;
    }

    async loadServicesForMonitor() {
        try {
            const response = await fetch('/api/v1/status', { credentials: 'include' });
            const data = await response.json();
            
            const serviceSelect = document.getElementById('monitorServiceId');
            if (!serviceSelect) return;

            // Clear existing options except the first one
            serviceSelect.innerHTML = '<option value="">Select a service...</option>';

            if (data.services && data.services.length > 0) {
                data.services.forEach(service => {
                    const option = document.createElement('option');
                    option.value = service.ID;
                    option.textContent = service.name;
                    serviceSelect.appendChild(option);
                });
            }
        } catch (error) {
            console.error('Error loading services for monitor:', error);
        }
    }

    async handleIncidentSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const serviceIds = Array.from(document.querySelectorAll('#affectedComponents input[name="service_ids"]:checked'))
            .map(cb => parseInt(cb.value));

        const incidentData = {
            title: formData.get('title'),
            description: formData.get('description'),
            status: formData.get('status'),
            impact: formData.get('impact'),
            service_ids: serviceIds
        };

        try {
            const response = await fetch('/api/v1/admin/incidents', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(incidentData)
            });

            if (response.ok) {
                this.showNotification('Incident created successfully', 'success');
                this.closeModal('incidentModal');
                this.loadDashboardData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to create incident');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async handleComponentSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const componentData = {
            name: formData.get('name'),
            description: formData.get('description'),
            group: formData.get('group'),
            status: formData.get('status'),
            show_uptime: formData.get('show_uptime') === 'on'
        };

        try {
            const response = await fetch('/api/v1/admin/status', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(componentData)
            });

            if (response.ok) {
                this.showNotification('Component created successfully', 'success');
                this.closeModal('componentModal');
                this.loadDashboardData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to create component');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async handleMonitorSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const monitorId = formData.get('id');
        
        const monitorData = {
            name: formData.get('name'),
            url: formData.get('url'),
            type: formData.get('type'),
            interval: parseInt(formData.get('interval')),
            expected_status: parseInt(formData.get('expected_status')),
            timeout: parseInt(formData.get('timeout')),
            service_id: parseInt(formData.get('service_id'))
        };

        try {
            const url = monitorId ? `/api/v1/admin/monitors/${monitorId}` : '/api/v1/admin/monitors';
            const method = monitorId ? 'PUT' : 'POST';
            
            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(monitorData)
            });

            if (response.ok) {
                this.showNotification(monitorId ? 'Monitor updated successfully' : 'Monitor created successfully', 'success');
                this.closeModal('monitorModal');
                this.loadDashboardData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to save monitor');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
    }

    async handleMaintenanceSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        const serviceIds = Array.from(document.querySelectorAll('#maintenanceComponents input[name="service_ids"]:checked'))
            .map(cb => parseInt(cb.value));

        const maintenanceData = {
            title: formData.get('title'),
            description: formData.get('description'),
            start_at: formData.get('start_at'),
            end_at: formData.get('end_at'),
            service_ids: serviceIds
        };

        try {
            const response = await fetch('/api/v1/admin/maintenance', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(maintenanceData)
            });

            if (response.ok) {
                this.showNotification('Maintenance scheduled successfully', 'success');
                this.closeModal('maintenanceModal');
                this.loadDashboardData();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to schedule maintenance');
            }
        } catch (error) {
            this.showNotification(error.message, 'error');
        }
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

    getTimeAgo(date) {
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} minutes ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
        return `${Math.floor(diffInSeconds / 86400)} days ago`;
    }
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.adminDashboard = new AdminDashboard();
});

// Export for global access
window.AdminDashboard = AdminDashboard;
