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

        // Sidebar navigation
        this.setupSidebarNavigation();

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

    setupSidebarNavigation() {
        console.log('Setting up sidebar navigation...');
        const navLinks = document.querySelectorAll('.sidebar-nav .nav-link');
        console.log('Found navigation links:', navLinks.length);
        
        // Handle sidebar navigation clicks
        navLinks.forEach(link => {
            console.log('Setting up listener for:', link.getAttribute('href'));
            link.addEventListener('click', (e) => {
                console.log('Navigation link clicked:', link.getAttribute('href'));
                e.preventDefault();
                const href = link.getAttribute('href');
                
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

        // Handle browser back/forward buttons
        window.addEventListener('popstate', (e) => {
            if (e.state && e.state.section) {
                this.navigateToSection(e.state.section);
            }
        });

        // Handle initial hash on page load
        const initialHash = window.location.hash.substring(1);
        if (initialHash) {
            this.navigateToSection(initialHash);
        }
    }

    navigateToSection(section) {
        console.log('navigateToSection called with:', section);
        
        // Update active nav item
        document.querySelectorAll('.sidebar-nav .nav-item').forEach(item => {
            item.classList.remove('active');
        });
        
        const activeLink = document.querySelector(`.sidebar-nav .nav-link[href="#${section}"]`);
        if (activeLink) {
            activeLink.closest('.nav-item').classList.add('active');
            console.log('Updated active nav item');
        }

        // Hide all content sections
        const allSections = document.querySelectorAll('.content-section');
        console.log('Found content sections:', allSections.length);
        allSections.forEach(section => {
            console.log('Hiding section:', section.id);
            section.style.display = 'none';
        });

        // Show the selected section
        const targetSection = document.getElementById(section);
        if (targetSection) {
            targetSection.style.display = 'block';
            targetSection.scrollIntoView({ behavior: 'smooth' });
            console.log('Showing section:', section);
        } else {
            console.log('Section not found, loading dynamically:', section);
            // If section doesn't exist, load it dynamically
            this.loadSection(section);
        }

        // Update URL without page reload
        history.pushState({ section }, '', `#${section}`);
    }

    async loadSection(section) {
        try {
            // Create a placeholder section if it doesn't exist
            let sectionElement = document.getElementById(section);
            if (!sectionElement) {
                sectionElement = document.createElement('div');
                sectionElement.id = section;
                sectionElement.className = 'content-section';
                sectionElement.innerHTML = `
                    <div class="section-header">
                        <h2>${this.getSectionTitle(section)}</h2>
                    </div>
                    <div class="section-content">
                        <div class="loading-spinner">
                            <i class="fas fa-spinner fa-spin"></i>
                            <p>Loading ${this.getSectionTitle(section)}...</p>
                        </div>
                    </div>
                `;
                
                const mainContent = document.querySelector('.main-content');
                if (mainContent) {
                    mainContent.appendChild(sectionElement);
                }
            }

            // Load section-specific content
            await this.loadSectionContent(section);
            
        } catch (error) {
            console.error(`Error loading section ${section}:`, error);
            this.showNotification(`Failed to load ${this.getSectionTitle(section)}`, 'error');
        }
    }

    getSectionTitle(section) {
        const titles = {
            'dashboard': 'Dashboard',
            'incidents': 'Incidents',
            'components': 'Components',
            'maintenance': 'Maintenance',
            'monitors': 'Monitors',
            'subscribers': 'Subscribers',
            'analytics': 'Analytics',
            'integrations': 'Integrations',
            'branding': 'Branding',
            'users': 'Users'
        };
        return titles[section] || section.charAt(0).toUpperCase() + section.slice(1);
    }

    async loadSectionContent(section) {
        const sectionElement = document.getElementById(section);
        if (!sectionElement) return;

        const contentDiv = sectionElement.querySelector('.section-content');
        if (!contentDiv) return;

        switch (section) {
            case 'incidents':
                await this.loadIncidentsSection(contentDiv);
                break;
            case 'components':
                await this.loadComponentsSection(contentDiv);
                break;
            case 'maintenance':
                await this.loadMaintenanceSection(contentDiv);
                break;
            case 'monitors':
                await this.loadMonitorsSection(contentDiv);
                break;
            case 'subscribers':
                await this.loadSubscribersSection(contentDiv);
                break;
            case 'analytics':
                await this.loadAnalyticsSection(contentDiv);
                break;
            case 'integrations':
                await this.loadIntegrationsSection(contentDiv);
                break;
            case 'branding':
                await this.loadBrandingSection(contentDiv);
                break;
            case 'users':
                await this.loadUsersSection(contentDiv);
                break;
            default:
                contentDiv.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-info-circle"></i>
                        <h3>Section Coming Soon</h3>
                        <p>The ${this.getSectionTitle(section)} section is under development.</p>
                    </div>
                `;
        }
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
                const result = await response.json();
                this.showNotification('Incident created successfully', 'success');
                this.closeModal('incidentModal');
                this.loadIncidentsContent(); // Reload the incidents section
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
                const result = await response.json();
                this.showNotification('Component created successfully', 'success');
                this.closeModal('componentModal');
                this.loadComponentsContent(); // Reload the components section
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

    // Section loading methods
    async loadIncidentsSection(contentDiv) {
        // Show loading state
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('incidentModal')">
                    <i class="fas fa-plus"></i> Create Incident
                </button>
            </div>
            <div class="loading-spinner">
                <i class="fas fa-spinner fa-spin"></i> Loading incidents...
            </div>
        `;

        try {
            // Fetch incidents from API
            const response = await fetch('/api/v1/admin/incidents', {
                credentials: 'include'
            });

            let incidents = [];
            if (response.ok) {
                const data = await response.json();
                incidents = data.incidents || [];
            } else {
                // Fallback to mock data if API fails
                incidents = [
                    {
                        id: 1,
                        title: "Database connectivity issues",
                        status: "investigating",
                        impact: "major",
                        created_at: "2025-01-07T10:30:00Z",
                        services: ["Database", "API"]
                    },
                    {
                        id: 2,
                        title: "Scheduled maintenance",
                        status: "monitoring",
                        impact: "minor",
                        created_at: "2025-01-07T08:00:00Z",
                        services: ["Web App"]
                    }
                ];
            }
        
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('incidentModal')">
                    <i class="fas fa-plus"></i> Create Incident
                </button>
            </div>
            <div class="incidents-list" id="incidentsList">
                ${incidents.map(incident => `
                    <div class="incident-card">
                        <div class="incident-header">
                            <h3>${incident.title}</h3>
                            <span class="status-badge ${incident.status}">${incident.status}</span>
                        </div>
                        <div class="incident-details">
                            <p><strong>Impact:</strong> ${incident.impact}</p>
                            <p><strong>Services:</strong> ${incident.services.join(', ')}</p>
                            <p><strong>Created:</strong> ${new Date(incident.created_at).toLocaleString()}</p>
                        </div>
                        <div class="incident-actions">
                            <button class="btn-secondary" onclick="dashboard.editIncident(${incident.id})">Edit</button>
                            <button class="btn-danger" onclick="dashboard.deleteIncident(${incident.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        } catch (error) {
            console.error('Error loading incidents:', error);
            contentDiv.innerHTML = `
                <div class="section-actions">
                    <button class="btn-primary" onclick="adminDashboard.openModal('incidentModal')">
                        <i class="fas fa-plus"></i> Create Incident
                    </button>
                </div>
                <div class="error-message">
                    <i class="fas fa-exclamation-triangle"></i>
                    Failed to load incidents. Please try again.
                </div>
            `;
        }
    }

    async loadComponentsSection(contentDiv) {
        // Show loading state
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('componentModal')">
                    <i class="fas fa-plus"></i> Add Component
                </button>
            </div>
            <div class="loading-spinner">
                <i class="fas fa-spinner fa-spin"></i> Loading components...
            </div>
        `;

        try {
            // Fetch components from API
            const response = await fetch('/api/v1/admin/services', {
                credentials: 'include'
            });

            let components = [];
            if (response.ok) {
                const data = await response.json();
                components = data.services || [];
            } else {
                // Fallback to mock data if API fails
                components = [
                    { id: 1, name: "API", status: "operational", description: "Main API service" },
                    { id: 2, name: "Database", status: "operational", description: "Primary database" },
                    { id: 3, name: "Web App", status: "degraded_performance", description: "Frontend application" }
                ];
            }
        
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('componentModal')">
                    <i class="fas fa-plus"></i> Add Component
                </button>
            </div>
            <div class="components-grid" id="componentsGrid">
                ${components.map(component => `
                    <div class="component-card">
                        <div class="component-header">
                            <h3>${component.name}</h3>
                            <span class="status-indicator ${component.status}"></span>
                        </div>
                        <p>${component.description}</p>
                        <div class="component-actions">
                            <button class="btn-secondary" onclick="dashboard.editComponent(${component.id})">Edit</button>
                            <button class="btn-danger" onclick="dashboard.deleteComponent(${component.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        } catch (error) {
            console.error('Error loading components:', error);
            contentDiv.innerHTML = `
                <div class="section-actions">
                    <button class="btn-primary" onclick="adminDashboard.openModal('componentModal')">
                        <i class="fas fa-plus"></i> Add Component
                    </button>
                </div>
                <div class="error-message">
                    <i class="fas fa-exclamation-triangle"></i>
                    Failed to load components. Please try again.
                </div>
            `;
        }
    }

    async loadMaintenanceSection(contentDiv) {
        // Mock maintenance data
        const mockMaintenance = [
            {
                id: 1,
                title: "Database optimization",
                description: "Scheduled database maintenance",
                start_time: "2025-01-08T02:00:00Z",
                end_time: "2025-01-08T04:00:00Z",
                status: "scheduled"
            }
        ];
        
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('maintenanceModal')">
                    <i class="fas fa-plus"></i> Schedule Maintenance
                </button>
            </div>
            <div class="maintenance-list" id="maintenanceList">
                ${mockMaintenance.map(maintenance => `
                    <div class="maintenance-card">
                        <div class="maintenance-header">
                            <h3>${maintenance.title}</h3>
                            <span class="status-badge ${maintenance.status}">${maintenance.status}</span>
                        </div>
                        <p>${maintenance.description}</p>
                        <div class="maintenance-details">
                            <p><strong>Start:</strong> ${new Date(maintenance.start_time).toLocaleString()}</p>
                            <p><strong>End:</strong> ${new Date(maintenance.end_time).toLocaleString()}</p>
                        </div>
                        <div class="maintenance-actions">
                            <button class="btn-secondary" onclick="dashboard.editMaintenance(${maintenance.id})">Edit</button>
                            <button class="btn-danger" onclick="dashboard.deleteMaintenance(${maintenance.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    async loadMonitorsSection(contentDiv) {
        // Mock monitors data
        const mockMonitors = [
            { id: 1, name: "API Health Check", type: "http", url: "https://api.example.com/health", status: "up" },
            { id: 2, name: "Database Ping", type: "tcp", host: "db.example.com:5432", status: "up" }
        ];
        
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('monitorModal')">
                    <i class="fas fa-plus"></i> Add Monitor
                </button>
            </div>
            <div class="monitors-list" id="monitorsList">
                ${mockMonitors.map(monitor => `
                    <div class="monitor-card">
                        <div class="monitor-header">
                            <h3>${monitor.name}</h3>
                            <span class="status-indicator ${monitor.status}"></span>
                        </div>
                        <div class="monitor-details">
                            <p><strong>Type:</strong> ${monitor.type.toUpperCase()}</p>
                            <p><strong>Target:</strong> ${monitor.url || monitor.host}</p>
                        </div>
                        <div class="monitor-actions">
                            <button class="btn-secondary" onclick="dashboard.editMonitor(${monitor.id})">Edit</button>
                            <button class="btn-danger" onclick="dashboard.deleteMonitor(${monitor.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    async loadSubscribersSection(contentDiv) {
        // Mock subscribers data
        const mockSubscribers = [
            { id: 1, email: "admin@example.com", phone: "+1234567890", status: "active", created_at: "2025-01-01T00:00:00Z" },
            { id: 2, email: "user@example.com", phone: "+0987654321", status: "active", created_at: "2025-01-02T00:00:00Z" }
        ];
        
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('subscriberModal')">
                    <i class="fas fa-plus"></i> Add Subscriber
                </button>
            </div>
            <div class="subscribers-list" id="subscribersList">
                ${mockSubscribers.map(subscriber => `
                    <div class="subscriber-card">
                        <div class="subscriber-header">
                            <h3>${subscriber.email}</h3>
                            <span class="status-badge ${subscriber.status}">${subscriber.status}</span>
                        </div>
                        <div class="subscriber-details">
                            <p><strong>Phone:</strong> ${subscriber.phone}</p>
                            <p><strong>Created:</strong> ${new Date(subscriber.created_at).toLocaleString()}</p>
                        </div>
                        <div class="subscriber-actions">
                            <button class="btn-secondary" onclick="dashboard.editSubscriber(${subscriber.id})">Edit</button>
                            <button class="btn-danger" onclick="dashboard.deleteSubscriber(${subscriber.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    async loadAnalyticsSection(contentDiv) {
        // Mock analytics data
        const mockAnalytics = {
            uptime: 99.9,
            incidents: 3,
            subscribers: 150,
            avgResponseTime: 120
        };
        
        contentDiv.innerHTML = `
            <div class="analytics-dashboard">
                <div class="analytics-grid">
                    <div class="analytics-card">
                        <h3>Uptime</h3>
                        <div class="metric-value">${mockAnalytics.uptime}%</div>
                    </div>
                    <div class="analytics-card">
                        <h3>Incidents</h3>
                        <div class="metric-value">${mockAnalytics.incidents}</div>
                    </div>
                    <div class="analytics-card">
                        <h3>Subscribers</h3>
                        <div class="metric-value">${mockAnalytics.subscribers}</div>
                    </div>
                    <div class="analytics-card">
                        <h3>Avg Response Time</h3>
                        <div class="metric-value">${mockAnalytics.avgResponseTime}ms</div>
                    </div>
                </div>
                <div class="analytics-charts">
                    <div class="chart-placeholder">
                        <i class="fas fa-chart-line"></i>
                        <p>Analytics charts will be displayed here</p>
                    </div>
                </div>
            </div>
        `;
    }

    async loadIntegrationsSection(contentDiv) {
        contentDiv.innerHTML = `
            <div class="integrations-grid">
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fab fa-slack"></i>
                    </div>
                    <h3>Slack</h3>
                    <p>Send incident notifications to Slack channels</p>
                    <button class="btn-secondary">Configure</button>
                </div>
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fas fa-bell"></i>
                    </div>
                    <h3>PagerDuty</h3>
                    <p>Integrate with PagerDuty for incident management</p>
                    <button class="btn-secondary">Configure</button>
                </div>
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fas fa-chart-line"></i>
                    </div>
                    <h3>Datadog</h3>
                    <p>Export metrics to Datadog</p>
                    <button class="btn-secondary">Configure</button>
                </div>
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fas fa-plug"></i>
                    </div>
                    <h3>Webhooks</h3>
                    <p>Send notifications to custom webhooks</p>
                    <button class="btn-secondary">Configure</button>
                </div>
            </div>
        `;
    }

    async loadBrandingSection(contentDiv) {
        contentDiv.innerHTML = `
            <div class="branding-settings">
                <div class="branding-section">
                    <h3>Custom Domain</h3>
                    <p>Set up a custom domain for your status page</p>
                    <div class="form-group">
                        <input type="text" placeholder="status.yourcompany.com" class="form-input">
                        <button class="btn-primary">Save</button>
                    </div>
                </div>
                <div class="branding-section">
                    <h3>Logo</h3>
                    <p>Upload your company logo</p>
                    <div class="logo-upload">
                        <input type="file" accept="image/*" class="form-input">
                        <button class="btn-primary">Upload</button>
                    </div>
                </div>
                <div class="branding-section">
                    <h3>Colors</h3>
                    <p>Customize your status page colors</p>
                    <div class="color-picker">
                        <div class="color-option">
                            <label>Primary Color</label>
                            <input type="color" value="#0052cc" class="form-input">
                        </div>
                        <div class="color-option">
                            <label>Secondary Color</label>
                            <input type="color" value="#f4f5f7" class="form-input">
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    async loadUsersSection(contentDiv) {
        try {
            const response = await fetch('/api/v1/admin/users', { credentials: 'include' });
            const users = await response.json();
            
            contentDiv.innerHTML = `
                <div class="section-actions">
                    <button class="btn-primary" onclick="adminDashboard.openModal('userModal')">
                        <i class="fas fa-plus"></i> Add User
                    </button>
                </div>
                <div class="users-list" id="usersList">
                    ${users.length > 0 ? 
                        users.map(user => this.createUserCard(user)).join('') :
                        '<div class="empty-state"><i class="fas fa-user-shield"></i><h3>No Users</h3><p>Add users to manage your status page</p></div>'
                    }
                </div>
            `;
        } catch (error) {
            contentDiv.innerHTML = `<div class="error-state">Failed to load users: ${error.message}</div>`;
        }
    }

    // Helper methods for creating cards
    createMaintenanceCard(maintenance) {
        const startDate = new Date(maintenance.StartAt).toLocaleDateString();
        const endDate = new Date(maintenance.EndAt).toLocaleDateString();
        const statusClass = maintenance.Status.toLowerCase();
        
        return `
            <div class="maintenance-card">
                <div class="maintenance-header">
                    <div class="maintenance-title">${maintenance.Title}</div>
                    <div class="maintenance-status ${statusClass}">${maintenance.Status}</div>
                </div>
                <div class="maintenance-description">${maintenance.Description || 'No description'}</div>
                <div class="maintenance-time">${startDate} - ${endDate}</div>
            </div>
        `;
    }

    createMonitorCard(monitor) {
        const statusClass = monitor.Status ? monitor.Status.toLowerCase() : 'unknown';
        const lastCheck = monitor.LastCheckAt ? new Date(monitor.LastCheckAt).toLocaleDateString() : 'Never';
        
        return `
            <div class="monitor-card">
                <div class="monitor-header">
                    <div class="monitor-name">${monitor.Name}</div>
                    <div class="monitor-status ${statusClass}">${monitor.Status || 'Unknown'}</div>
                </div>
                <div class="monitor-url">${monitor.URL}</div>
                <div class="monitor-details">
                    <span>Type: ${monitor.Type}</span>
                    <span>Interval: ${monitor.Interval}s</span>
                    <span>Last Check: ${lastCheck}</span>
                </div>
            </div>
        `;
    }

    createSubscriberCard(subscriber) {
        return `
            <div class="subscriber-card">
                <div class="subscriber-email">${subscriber.Email}</div>
                <div class="subscriber-services">
                    ${subscriber.Services ? subscriber.Services.length : 0} services subscribed
                </div>
                <div class="subscriber-actions">
                    <button class="btn-sm btn-danger">Remove</button>
                </div>
            </div>
        `;
    }

    createUserCard(user) {
        const roleClass = user.Role ? user.Role.toLowerCase() : 'unknown';
        const lastLogin = user.LastLoginAt ? new Date(user.LastLoginAt).toLocaleDateString() : 'Never';
        
        return `
            <div class="user-card">
                <div class="user-header">
                    <div class="user-email">${user.Email}</div>
                    <div class="user-role ${roleClass}">${user.Role || 'Unknown'}</div>
                </div>
                <div class="user-details">
                    <span>Username: ${user.Username}</span>
                    <span>Last Login: ${lastLogin}</span>
                </div>
                <div class="user-actions">
                    <button class="btn-sm btn-secondary">Edit</button>
                    <button class="btn-sm btn-danger">Delete</button>
                </div>
            </div>
        `;
    }

    // Helper functions to reload specific sections
    loadIncidentsContent() {
        const contentDiv = document.getElementById('incidents');
        if (contentDiv) {
            this.loadIncidentsSection(contentDiv);
        }
    }

    loadComponentsContent() {
        const contentDiv = document.getElementById('components');
        if (contentDiv) {
            this.loadComponentsSection(contentDiv);
        }
    }
}

// Initialize the dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, initializing AdminDashboard...');
    window.adminDashboard = new AdminDashboard();
    console.log('AdminDashboard initialized');
});

// Export for global access
window.AdminDashboard = AdminDashboard;
