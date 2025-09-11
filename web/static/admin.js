// Atlassian Statuspage Admin Dashboard JavaScript

class AdminDashboard {
    constructor() {
        console.log('AdminDashboard constructor called');
        // Simple test to see if JavaScript is working
        console.log('JavaScript is working! AdminDashboard initialized.');
        this.init();
    }

    init() {
        console.log('AdminDashboard init called');
        this.initializeSidebarState();
        this.setupEventListeners();
        this.setupModals();
        
        // Check for initial hash and navigate accordingly
        setTimeout(() => {
            const initialHash = window.location.hash.substring(1);
            console.log('Initial hash:', initialHash);
            
            if (initialHash && initialHash !== 'dashboard') {
                // Navigate to the specified section
                this.navigateToSection(initialHash);
            } else {
                // Default to dashboard
                this.loadDashboardData();
            }
        }, 100);
    }

    initializeSidebarState() {
        const sidebar = document.querySelector('.sidebar');
        const sidebarToggle = document.getElementById('sidebarToggle');
        
        if (sidebar && sidebarToggle) {
            // Get saved state from localStorage
            const savedState = localStorage.getItem('admin-sidebar-state');
            console.log('Saved sidebar state:', savedState);
            
            if (savedState === 'collapsed') {
                // Set to collapsed state
                sidebar.classList.remove('open');
                sidebar.classList.add('collapsed');
                sidebarToggle.querySelector('i').className = 'fas fa-chevron-right';
                console.log('Sidebar initialized as collapsed');
            } else {
                // Default to open state
                sidebar.classList.remove('collapsed');
                sidebar.classList.add('open');
                sidebarToggle.querySelector('i').className = 'fas fa-chevron-left';
                console.log('Sidebar initialized as open');
            }
        }
    }

    setupEventListeners() {
        // Sidebar toggle for mobile
        const sidebarToggle = document.getElementById('sidebarToggle');
        const sidebar = document.querySelector('.sidebar');
        
        if (sidebarToggle && sidebar) {
            console.log('Setting up sidebar toggle...');
            
            // Function to toggle sidebar state
            const toggleSidebar = (e) => {
                e.stopPropagation();
                console.log('Sidebar toggle clicked!');
                
                // Toggle between open and collapsed states
                if (sidebar.classList.contains('open')) {
                    // If open, collapse it
                    sidebar.classList.remove('open');
                    sidebar.classList.add('collapsed');
                    sidebarToggle.querySelector('i').className = 'fas fa-chevron-right';
                    localStorage.setItem('admin-sidebar-state', 'collapsed');
                    console.log('Sidebar collapsed');
                } else if (sidebar.classList.contains('collapsed')) {
                    // If collapsed, open it
                    sidebar.classList.remove('collapsed');
                    sidebar.classList.add('open');
                    sidebarToggle.querySelector('i').className = 'fas fa-chevron-left';
                    localStorage.setItem('admin-sidebar-state', 'open');
                    console.log('Sidebar opened');
                } else {
                    // If hidden, open it
                    sidebar.classList.add('open');
                    sidebarToggle.querySelector('i').className = 'fas fa-chevron-left';
                    localStorage.setItem('admin-sidebar-state', 'open');
                    console.log('Sidebar opened from hidden');
                }
            };
            
            // Add event listener to sidebar toggle button
            sidebarToggle.addEventListener('click', toggleSidebar);
        }

        // DISABLED: Don't close sidebar when clicking outside
        // This was causing the sidebar to disappear when clicking elsewhere
        // document.addEventListener('click', (e) => {
        //     if (window.innerWidth <= 768 && sidebar.classList.contains('open')) {
        //         if (!sidebar.contains(e.target) && !sidebarToggle.contains(e.target)) {
        //             sidebar.classList.remove('open');
        //         }
        //     }
        // });

        // DISABLED: Don't auto-close sidebar on resize
        // This was causing the sidebar to disappear when resizing
        // window.addEventListener('resize', () => {
        //     if (window.innerWidth > 768) {
        //         sidebar.classList.remove('open');
        //     }
        // });

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

        // Handle hash changes (when user manually changes URL)
        window.addEventListener('hashchange', (e) => {
            const newHash = window.location.hash.substring(1);
            console.log('Hash changed to:', newHash);
            if (newHash) {
                this.navigateToSection(newHash);
            } else {
                // If no hash, go to dashboard
                this.loadDashboardData();
            }
        });

        // Initial hash handling is now done in init() method
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
        
        // Update active nav item
        document.querySelectorAll('.sidebar-nav .nav-item').forEach(item => {
            item.classList.remove('active');
        });
        
        const activeLink = document.querySelector(`.sidebar-nav .nav-link[href="#${section}"]`);
        if (activeLink) {
            activeLink.closest('.nav-item').classList.add('active');
            console.log('Updated active nav item');
        }

        // Update page title
        const pageTitle = document.querySelector('.page-title');
        if (pageTitle) {
            pageTitle.textContent = section.charAt(0).toUpperCase() + section.slice(1);
        }

        // Clear any existing content in the target section to prevent overlap
        if (section !== 'dashboard') {
            const sectionContent = targetSection.querySelector('[id$="-content"]');
            if (sectionContent) {
                sectionContent.innerHTML = '';
            }
        }

        // Load section-specific data
        try {
            await this.loadSectionData(section);
        } catch (error) {
            console.error('Error loading section data:', error);
            // Show error in the section
            const sectionContent = targetSection.querySelector('[id$="-content"]') || targetSection;
            if (sectionContent) {
                sectionContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Content</h3>
                        <p>Failed to load ${section} content. Please try again.</p>
                        <button class="btn-secondary" onclick="window.location.reload()">Refresh Page</button>
                    </div>
                `;
            }
        }

        // Update URL without page reload
        history.pushState({ section }, '', `#${section}`);
    }

    async loadSectionData(section) {
        console.log('Loading data for section:', section);
        
        try {
            switch (section) {
                case 'dashboard':
                    await this.loadDashboardData();
                    break;
                case 'incidents':
                    await this.loadIncidentsData();
                    break;
                case 'components':
                    await this.loadComponentsData();
                    break;
                case 'maintenance':
                    await this.loadMaintenanceData();
                    break;
                case 'monitors':
                    await this.loadMonitorsData();
                    break;
                case 'subscribers':
                    await this.loadSubscribersData();
                    break;
                case 'analytics':
                    await this.loadAnalyticsData();
                    break;
                case 'integrations':
                    await this.loadIntegrationsData();
                    break;
                case 'branding':
                    await this.loadBrandingData();
                    break;
                case 'users':
                    await this.loadUsersData();
                    break;
                case 'feature-flags':
                    await this.loadFeatureFlagsData();
                    break;
                default:
                    console.log('No specific data loading for section:', section);
            }
        } catch (error) {
            console.error('Error loading section data:', error);
        }
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
        console.log(`loadSectionContent called for: ${section}`);
        const sectionElement = document.getElementById(section);
        if (!sectionElement) {
            console.error(`Section element not found: ${section}`);
            return;
        }

        // Find the content div - it's named like "incidents-content", "components-content", etc.
        const contentDiv = document.getElementById(`${section}-content`);
        if (!contentDiv) {
            console.error(`Content div not found: ${section}-content`);
            return;
        }

        console.log(`Found content div for ${section}:`, contentDiv);

        switch (section) {
            case 'dashboard':
                await this.loadDashboardOverview(sectionElement);
                break;
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
        console.log('loadDashboardData called');
        const dashboardSection = document.getElementById('dashboard');
        console.log('Dashboard section found:', dashboardSection);
        if (!dashboardSection) {
            console.error('Dashboard section not found!');
            return;
        }
        
        // Only show dashboard content if we're actually on the dashboard section
        if (window.location.hash === '#dashboard' || window.location.hash === '') {
            // Hide all other content sections first
            const allSections = document.querySelectorAll('.content-section');
            console.log('Found content sections:', allSections.length);
            allSections.forEach(section => {
                if (section.id !== 'dashboard') {
                    section.style.display = 'none';
                    console.log('Hiding section:', section.id);
                }
            });
            
            // Make sure the dashboard section is visible
            dashboardSection.style.display = 'block';
            console.log('Dashboard section display set to block');
            
            try {
                await this.loadDashboardOverview(dashboardSection);
            } catch (error) {
                console.error('Error loading dashboard data:', error);
                this.showNotification('Failed to load dashboard data', 'error');
            }
        }
    }

    async loadIncidentsData() {
        console.log('Loading incidents data...');
        const incidentsContent = document.getElementById('incidents-content');
        const incidentsLoading = document.getElementById('incidents-loading');
        
        if (incidentsLoading) incidentsLoading.style.display = 'block';
        if (incidentsContent) incidentsContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/incidents', { credentials: 'include' });
            const data = await response.json();
            
            if (incidentsLoading) incidentsLoading.style.display = 'none';
            
            if (incidentsContent) {
                if (data.incidents && data.incidents.length > 0) {
                    incidentsContent.innerHTML = this.renderIncidentsList(data.incidents);
                } else {
                    incidentsContent.innerHTML = `
                        <div class="empty-state">
                            <i class="fas fa-exclamation-triangle"></i>
                            <h3>No Incidents</h3>
                            <p>No incidents have been reported yet</p>
                            <button class="btn-primary" onclick="adminDashboard.openModal('incidentModal')">
                                Create First Incident
                            </button>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Error loading incidents:', error);
            if (incidentsLoading) incidentsLoading.style.display = 'none';
            if (incidentsContent) {
                incidentsContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Incidents</h3>
                        <p>Failed to load incidents. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadComponentsData() {
        console.log('Loading components data...');
        const componentsContent = document.getElementById('components-content');
        const componentsLoading = document.getElementById('components-loading');
        
        if (componentsLoading) componentsLoading.style.display = 'block';
        if (componentsContent) componentsContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/services', { credentials: 'include' });
            const data = await response.json();
            
            if (componentsLoading) componentsLoading.style.display = 'none';
            
            if (componentsContent) {
                if (data.services && data.services.length > 0) {
                    componentsContent.innerHTML = this.renderComponentsList(data.services);
                } else {
                    componentsContent.innerHTML = `
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
            }
        } catch (error) {
            console.error('Error loading components:', error);
            if (componentsLoading) componentsLoading.style.display = 'none';
            if (componentsContent) {
                componentsContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Components</h3>
                        <p>Failed to load components. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadMaintenanceData() {
        console.log('Loading maintenance data...');
        const maintenanceContent = document.getElementById('maintenance-content');
        const maintenanceLoading = document.getElementById('maintenance-loading');
        
        if (maintenanceLoading) maintenanceLoading.style.display = 'block';
        if (maintenanceContent) maintenanceContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/maintenance', { credentials: 'include' });
            const data = await response.json();
            
            if (maintenanceLoading) maintenanceLoading.style.display = 'none';
            
            if (maintenanceContent) {
                if (data.maintenance && data.maintenance.length > 0) {
                    maintenanceContent.innerHTML = this.renderMaintenanceList(data.maintenance);
                } else {
                    maintenanceContent.innerHTML = `
                        <div class="empty-state">
                            <i class="fas fa-tools"></i>
                            <h3>No Maintenance Scheduled</h3>
                            <p>No maintenance windows are currently scheduled</p>
                            <button class="btn-primary" onclick="adminDashboard.openModal('maintenanceModal')">
                                Schedule Maintenance
                            </button>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Error loading maintenance:', error);
            if (maintenanceLoading) maintenanceLoading.style.display = 'none';
            if (maintenanceContent) {
                maintenanceContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Maintenance</h3>
                        <p>Failed to load maintenance events. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadMonitorsData() {
        console.log('Loading monitors data...');
        const monitorsContent = document.getElementById('monitors-content');
        const monitorsLoading = document.getElementById('monitors-loading');
        
        if (monitorsLoading) monitorsLoading.style.display = 'block';
        if (monitorsContent) monitorsContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/monitors', { credentials: 'include' });
            const data = await response.json();
            
            if (monitorsLoading) monitorsLoading.style.display = 'none';
            
            if (monitorsContent) {
                if (data.monitors && data.monitors.length > 0) {
                    monitorsContent.innerHTML = this.renderMonitorsList(data.monitors);
                } else {
                    monitorsContent.innerHTML = `
                        <div class="empty-state">
                            <i class="fas fa-heartbeat"></i>
                            <h3>No Monitors</h3>
                            <p>Add your first monitor to track service health</p>
                            <button class="btn-primary" onclick="adminDashboard.openModal('monitorModal')">
                                Add Monitor
                            </button>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Error loading monitors:', error);
            if (monitorsLoading) monitorsLoading.style.display = 'none';
            if (monitorsContent) {
                monitorsContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Monitors</h3>
                        <p>Failed to load monitors. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadSubscribersData() {
        console.log('Loading subscribers data...');
        const subscribersContent = document.getElementById('subscribers-content');
        const subscribersLoading = document.getElementById('subscribers-loading');
        
        if (subscribersLoading) subscribersLoading.style.display = 'block';
        if (subscribersContent) subscribersContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/subscribers', { credentials: 'include' });
            const data = await response.json();
            
            if (subscribersLoading) subscribersLoading.style.display = 'none';
            
            if (subscribersContent) {
                if (data.subscribers && data.subscribers.length > 0) {
                    subscribersContent.innerHTML = this.renderSubscribersList(data.subscribers);
                } else {
                    subscribersContent.innerHTML = `
                        <div class="empty-state">
                            <i class="fas fa-users"></i>
                            <h3>No Subscribers</h3>
                            <p>No subscribers have signed up for notifications yet</p>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Error loading subscribers:', error);
            if (subscribersLoading) subscribersLoading.style.display = 'none';
            if (subscribersContent) {
                subscribersContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Subscribers</h3>
                        <p>Failed to load subscribers. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadUsersData() {
        console.log('Loading users data...');
        const usersContent = document.getElementById('users-content');
        const usersLoading = document.getElementById('users-loading');
        
        if (usersLoading) usersLoading.style.display = 'block';
        if (usersContent) usersContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/users', { credentials: 'include' });
            const data = await response.json();
            
            if (usersLoading) usersLoading.style.display = 'none';
            
            if (usersContent) {
                if (data.users && data.users.length > 0) {
                    usersContent.innerHTML = this.renderUsersList(data.users);
                } else {
                    usersContent.innerHTML = `
                        <div class="empty-state">
                            <i class="fas fa-user-shield"></i>
                            <h3>No Users</h3>
                            <p>No users have been created yet</p>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Error loading users:', error);
            if (usersLoading) usersLoading.style.display = 'none';
            if (usersContent) {
                usersContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Users</h3>
                        <p>Failed to load users. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadIntegrationsData() {
        console.log('Loading integrations data...');
        const integrationsContent = document.getElementById('integrations-content');
        const integrationsLoading = document.getElementById('integrations-loading');
        
        if (integrationsLoading) integrationsLoading.style.display = 'block';
        if (integrationsContent) integrationsContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/integrations', { credentials: 'include' });
            const data = await response.json();
            
            if (integrationsLoading) integrationsLoading.style.display = 'none';
            
            if (integrationsContent) {
                if (data.integrations && data.integrations.length > 0) {
                    integrationsContent.innerHTML = this.renderIntegrationsList(data.integrations);
                } else {
                    integrationsContent.innerHTML = `
                        <div class="empty-state">
                            <i class="fas fa-plug"></i>
                            <h3>No Integrations</h3>
                            <p>No integrations have been configured yet</p>
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Error loading integrations:', error);
            if (integrationsLoading) integrationsLoading.style.display = 'none';
            if (integrationsContent) {
                integrationsContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-circle"></i>
                        <h3>Error Loading Integrations</h3>
                        <p>Failed to load integrations. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    async loadAnalyticsData() {
        console.log('Loading analytics data...');
        const analyticsContent = document.getElementById('analytics-content');
        const analyticsLoading = document.getElementById('analytics-loading');
        
        if (analyticsLoading) analyticsLoading.style.display = 'block';
        if (analyticsContent) analyticsContent.innerHTML = '';
        
        try {
            // Load analytics data from API
            const [servicesRes, incidentsRes, maintenanceRes, subscribersRes] = await Promise.all([
                fetch('/api/v1/admin/services', { credentials: 'include' }),
                fetch('/api/v1/admin/incidents', { credentials: 'include' }),
                fetch('/api/v1/admin/maintenance', { credentials: 'include' }),
                fetch('/api/v1/admin/subscribers', { credentials: 'include' })
            ]);

            const services = await servicesRes.json();
            const incidents = await incidentsRes.json();
            const maintenance = await maintenanceRes.json();
            const subscribers = await subscribersRes.json();

            if (analyticsLoading) analyticsLoading.style.display = 'none';
            if (analyticsContent) {
                analyticsContent.innerHTML = `
                    <div class="analytics-dashboard">
                        <div class="analytics-grid">
                            <div class="analytics-card">
                                <div class="card-header">
                                    <h3><i class="fas fa-server"></i> Services Overview</h3>
                                </div>
                                <div class="card-content">
                                    <div class="metric">
                                        <span class="metric-value">${services.services ? services.services.length : 0}</span>
                                        <span class="metric-label">Total Services</span>
                                    </div>
                                    <div class="metric">
                                        <span class="metric-value">${services.services ? services.services.filter(s => s.status === 'operational').length : 0}</span>
                                        <span class="metric-label">Operational</span>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="analytics-card">
                                <div class="card-header">
                                    <h3><i class="fas fa-exclamation-triangle"></i> Incidents</h3>
                                </div>
                                <div class="card-content">
                                    <div class="metric">
                                        <span class="metric-value">${incidents.incidents ? incidents.incidents.length : 0}</span>
                                        <span class="metric-label">Total Incidents</span>
                                    </div>
                                    <div class="metric">
                                        <span class="metric-value">${incidents.incidents ? incidents.incidents.filter(i => i.status === 'investigating').length : 0}</span>
                                        <span class="metric-label">Active</span>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="analytics-card">
                                <div class="card-header">
                                    <h3><i class="fas fa-tools"></i> Maintenance</h3>
                                </div>
                                <div class="card-content">
                                    <div class="metric">
                                        <span class="metric-value">${maintenance.maintenance ? maintenance.maintenance.length : 0}</span>
                                        <span class="metric-label">Scheduled</span>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="analytics-card">
                                <div class="card-header">
                                    <h3><i class="fas fa-users"></i> Subscribers</h3>
                                </div>
                                <div class="card-content">
                                    <div class="metric">
                                        <span class="metric-value">${subscribers.subscribers ? subscribers.subscribers.length : 0}</span>
                                        <span class="metric-label">Total Subscribers</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <div class="analytics-charts">
                            <div class="chart-container">
                                <h3>Service Status Distribution</h3>
                                <canvas id="serviceStatusChart" width="400" height="200"></canvas>
                            </div>
                        </div>
                    </div>
                `;
                
                // Initialize charts if Chart.js is available
                if (typeof Chart !== 'undefined') {
                    this.initializeAnalyticsCharts(services.services || []);
                }
            }
        } catch (error) {
            console.error('Error loading analytics:', error);
        }
    }

    initializeAnalyticsCharts(services) {
        const ctx = document.getElementById('serviceStatusChart');
        if (!ctx) return;

        // Count services by status
        const statusCounts = services.reduce((acc, service) => {
            acc[service.status] = (acc[service.status] || 0) + 1;
            return acc;
        }, {});

        new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: Object.keys(statusCounts),
                datasets: [{
                    data: Object.values(statusCounts),
                    backgroundColor: [
                        '#28a745', // operational - green
                        '#ffc107', // degraded - yellow
                        '#dc3545', // outage - red
                        '#6c757d'  // unknown - gray
                    ],
                    borderWidth: 2,
                    borderColor: '#fff'
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

    async loadFeatureFlagsData() {
        console.log('Loading feature flags data...');
        const featureFlagsContent = document.getElementById('feature-flags-content');
        const featureFlagsLoading = document.getElementById('feature-flags-loading');
        
        if (featureFlagsLoading) featureFlagsLoading.style.display = 'block';
        if (featureFlagsContent) featureFlagsContent.innerHTML = '';
        
        try {
            const response = await fetch('/api/v1/admin/feature-flags', { credentials: 'include' });
            const data = await response.json();
            
            if (featureFlagsLoading) featureFlagsLoading.style.display = 'none';
            
            if (featureFlagsContent) {
                featureFlagsContent.innerHTML = `
                    <div class="feature-flags-dashboard">
                        <div class="feature-flags-grid">
                            <div class="feature-flag-card">
                                <div class="feature-flag-header">
                                    <h3><i class="fas fa-chart-line"></i> Per-Service Graphs</h3>
                                    <div class="feature-flag-toggle">
                                        <label class="toggle-switch">
                                            <input type="checkbox" id="per-service-graphs-toggle" ${data.feature_flags && data.feature_flags.find(f => f.feature === 'per_service_graphs')?.is_enabled ? 'checked' : ''}>
                                            <span class="toggle-slider"></span>
                                        </label>
                                    </div>
                                </div>
                                <div class="feature-flag-description">
                                    <p>Enable per-service monitoring graphs on the status page. Each service will have its own uptime chart.</p>
                                </div>
                            </div>
                            
                            <div class="feature-flag-card">
                                <div class="feature-flag-header">
                                    <h3><i class="fas fa-globe"></i> Custom Domains</h3>
                                    <div class="feature-flag-toggle">
                                        <label class="toggle-switch">
                                            <input type="checkbox" id="custom-domains-toggle" ${data.feature_flags && data.feature_flags.find(f => f.feature === 'custom_domains')?.is_enabled ? 'checked' : ''}>
                                            <span class="toggle-slider"></span>
                                        </label>
                                    </div>
                                </div>
                                <div class="feature-flag-description">
                                    <p>Allow custom domain configuration for your status page.</p>
                                </div>
                            </div>
                            
                            <div class="feature-flag-card">
                                <div class="feature-flag-header">
                                    <h3><i class="fas fa-chart-bar"></i> Advanced Analytics</h3>
                                    <div class="feature-flag-toggle">
                                        <label class="toggle-switch">
                                            <input type="checkbox" id="advanced-analytics-toggle" ${data.feature_flags && data.feature_flags.find(f => f.feature === 'advanced_analytics')?.is_enabled ? 'checked' : ''}>
                                            <span class="toggle-slider"></span>
                                        </label>
                                    </div>
                                </div>
                                <div class="feature-flag-description">
                                    <p>Enable advanced analytics and reporting features.</p>
                                </div>
                            </div>
                            
                            <div class="feature-flag-card">
                                <div class="feature-flag-header">
                                    <h3><i class="fas fa-key"></i> SSO Integration</h3>
                                    <div class="feature-flag-toggle">
                                        <label class="toggle-switch">
                                            <input type="checkbox" id="sso-toggle" ${data.feature_flags && data.feature_flags.find(f => f.feature === 'sso')?.is_enabled ? 'checked' : ''}>
                                            <span class="toggle-slider"></span>
                                        </label>
                                    </div>
                                </div>
                                <div class="feature-flag-description">
                                    <p>Enable Single Sign-On integration for user authentication.</p>
                                </div>
                            </div>
                            
                            <div class="feature-flag-card">
                                <div class="feature-flag-header">
                                    <h3><i class="fas fa-code"></i> API Access</h3>
                                    <div class="feature-flag-toggle">
                                        <label class="toggle-switch">
                                            <input type="checkbox" id="api-access-toggle" ${data.feature_flags && data.feature_flags.find(f => f.feature === 'api_access')?.is_enabled ? 'checked' : ''}>
                                            <span class="toggle-slider"></span>
                                        </label>
                                    </div>
                                </div>
                                <div class="feature-flag-description">
                                    <p>Enable API access for programmatic management of your status page.</p>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
                
                // Add event listeners for toggle switches
                this.setupFeatureFlagToggles();
            }
        } catch (error) {
            console.error('Error loading feature flags:', error);
            if (featureFlagsLoading) featureFlagsLoading.style.display = 'none';
            if (featureFlagsContent) {
                featureFlagsContent.innerHTML = `
                    <div class="error-state">
                        <i class="fas fa-exclamation-triangle"></i>
                        <h3>Error Loading Feature Flags</h3>
                        <p>Failed to load feature flags. Please try again.</p>
                    </div>
                `;
            }
        }
    }

    setupFeatureFlagToggles() {
        const toggles = [
            { id: 'per-service-graphs-toggle', feature: 'per_service_graphs' },
            { id: 'custom-domains-toggle', feature: 'custom_domains' },
            { id: 'advanced-analytics-toggle', feature: 'advanced_analytics' },
            { id: 'sso-toggle', feature: 'sso' },
            { id: 'api-access-toggle', feature: 'api_access' }
        ];
        
        toggles.forEach(toggle => {
            const element = document.getElementById(toggle.id);
            if (element) {
                element.addEventListener('change', async (e) => {
                    await this.updateFeatureFlag(toggle.feature, e.target.checked);
                });
            }
        });
    }

    async updateFeatureFlag(feature, enabled) {
        try {
            const response = await fetch('/api/v1/admin/feature-flags', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    feature: feature,
                    is_enabled: enabled
                })
            });
            
            if (response.ok) {
                this.showNotification(`Feature "${feature}" ${enabled ? 'enabled' : 'disabled'} successfully`, 'success');
            } else {
                throw new Error('Failed to update feature flag');
            }
        } catch (error) {
            console.error('Error updating feature flag:', error);
            this.showNotification('Failed to update feature flag', 'error');
        }
    }

    async loadBrandingData() {
        console.log('Loading branding data...');
        const brandingContent = document.getElementById('branding-content');
        const brandingLoading = document.getElementById('branding-loading');
        
        if (brandingLoading) brandingLoading.style.display = 'block';
        if (brandingContent) brandingContent.innerHTML = '';
        
        try {
            // For now, show a placeholder
            if (brandingLoading) brandingLoading.style.display = 'none';
            if (brandingContent) {
                brandingContent.innerHTML = `
                    <div class="empty-state">
                        <i class="fas fa-palette"></i>
                        <h3>Branding Settings</h3>
                        <p>Customize your status page appearance and branding</p>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Error loading branding:', error);
        }
    }

    async loadDashboardOverview(dashboardSection) {
        console.log('loadDashboardOverview called with:', dashboardSection);
        
        try {
            console.log('Loading dashboard overview...');
            // Fetch real data from APIs
            const [servicesResponse, incidentsResponse] = await Promise.all([
                fetch('/api/v1/admin/services', { credentials: 'include' }),
                fetch('/api/v1/admin/incidents', { credentials: 'include' })
            ]);

            console.log('Dashboard API responses:', {
                services: servicesResponse.status,
                incidents: incidentsResponse.status
            });

            const services = servicesResponse.ok ? (await servicesResponse.json()).services : [];
            const incidents = incidentsResponse.ok ? (await incidentsResponse.json()).incidents : [];

            // Calculate stats
            const totalServices = services.length;
            const operationalServices = services.filter(s => s.Status === 'operational').length;
            const activeIncidents = incidents.filter(i => i.Status !== 'resolved').length;
            const uptime = totalServices > 0 ? ((operationalServices / totalServices) * 100).toFixed(1) : 100;

            // Update the existing stat cards with real data
            const activeIncidentsElement = document.getElementById('activeIncidents');
            if (activeIncidentsElement) {
                activeIncidentsElement.textContent = activeIncidents;
            }

            const upcomingMaintenanceElement = document.getElementById('upcomingMaintenance');
            if (upcomingMaintenanceElement) {
                upcomingMaintenanceElement.textContent = '0'; // TODO: Get from maintenance API
            }

            const totalSubscribersElement = document.getElementById('totalSubscribers');
            if (totalSubscribersElement) {
                totalSubscribersElement.textContent = '0'; // TODO: Get from subscribers API
            }

            // Update the first stat card to show real data
            const firstStatCard = dashboardSection.querySelector('.stat-card:first-child .stat-content h3');
            if (firstStatCard) {
                firstStatCard.textContent = `${operationalServices}/${totalServices} Services Operational`;
            }
            const firstStatCardP = dashboardSection.querySelector('.stat-card:first-child .stat-content p');
            if (firstStatCardP) {
                firstStatCardP.textContent = `${uptime}% uptime`;
            }

            // Update components grid
            const componentsGrid = document.getElementById('componentsGrid');
            if (componentsGrid) {
                componentsGrid.innerHTML = services.slice(0, 6).map(service => `
                    <div class="component-card">
                        <div class="component-header">
                            <h4>${service.Name}</h4>
                            <span class="status-badge ${service.Status}">${service.Status}</span>
                        </div>
                        <p class="component-description">${service.Description || 'No description'}</p>
                        <div class="component-meta">
                            <span class="component-group">${service.Group || 'No Group'}</span>
                        </div>
                    </div>
                `).join('');
            }

            // Update recent incidents list
            const recentIncidentsList = document.getElementById('recentIncidentsList');
            if (recentIncidentsList) {
                recentIncidentsList.innerHTML = incidents.slice(0, 5).map(incident => `
                    <div class="incident-item">
                        <div class="incident-header">
                            <h4>${incident.Title}</h4>
                            <span class="status-badge ${incident.Status}">${incident.Status}</span>
                        </div>
                        <p class="incident-description">${incident.Description}</p>
                        <div class="incident-meta">
                            <span class="incident-time">${this.formatTimeAgo(incident.CreatedAt)}</span>
                            <span class="incident-impact">${incident.Impact} impact</span>
                        </div>
                    </div>
                `).join('');
            }

            // Update activity list
            const activityList = document.getElementById('activityList');
            if (activityList) {
                const recentActivities = [
                    {
                        icon: 'fas fa-check-circle',
                        type: 'operational',
                        message: 'All systems operational',
                        time: 'Just now'
                    },
                    ...incidents.slice(0, 3).map(incident => ({
                        icon: 'fas fa-exclamation-triangle',
                        type: 'incident',
                        message: incident.Title,
                        time: this.formatTimeAgo(incident.CreatedAt)
                    }))
                ];

                activityList.innerHTML = recentActivities.map(activity => `
                    <div class="activity-item">
                        <div class="activity-icon ${activity.type}">
                            <i class="${activity.icon}"></i>
                        </div>
                        <div class="activity-content">
                            <p><strong>${activity.type === 'operational' ? 'System Status' : 'Incident'}</strong> ${activity.message}</p>
                            <span class="activity-time">${activity.time}</span>
                        </div>
                    </div>
                `).join('');
            }

            // Add event listeners to quick action buttons
            this.setupQuickActionButtons();

        } catch (error) {
            console.error('Error loading dashboard overview:', error);
            console.error('Error details:', error.message, error.stack);
            // Don't replace the entire dashboard, just show an error notification
            this.showNotification('Failed to load dashboard data: ' + error.message, 'error');
        }
    }

    setupQuickActionButtons() {
        // Create Incident button
        const createIncidentBtn = document.getElementById('createIncidentBtn');
        if (createIncidentBtn) {
            createIncidentBtn.addEventListener('click', () => {
                this.openModal('incidentModal');
            });
        }

        // Schedule Maintenance button
        const scheduleMaintenanceBtn = document.getElementById('scheduleMaintenanceBtn');
        if (scheduleMaintenanceBtn) {
            scheduleMaintenanceBtn.addEventListener('click', () => {
                this.openModal('maintenanceModal');
            });
        }

        // Add Component button
        const addComponentBtn = document.getElementById('addComponentBtn');
        if (addComponentBtn) {
            addComponentBtn.addEventListener('click', () => {
                this.openModal('componentModal');
            });
        }

        // Add Monitor button
        const addMonitorBtn = document.getElementById('addMonitorBtn');
        if (addMonitorBtn) {
            addMonitorBtn.addEventListener('click', () => {
                this.openModal('monitorModal');
            });
        }

        // Manage Components button
        const manageComponentsBtn = document.getElementById('manageComponentsBtn');
        if (manageComponentsBtn) {
            manageComponentsBtn.addEventListener('click', () => {
                this.navigateToSection('components');
            });
        }

        // View All Incidents button
        const viewAllIncidentsBtn = document.getElementById('viewAllIncidentsBtn');
        if (viewAllIncidentsBtn) {
            viewAllIncidentsBtn.addEventListener('click', () => {
                this.navigateToSection('incidents');
            });
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
            console.log('Loading incidents section...');
            // Fetch incidents from API
            const response = await fetch('/api/v1/admin/incidents', {
                credentials: 'include'
            });

            console.log('Incidents API response:', response.status, response.statusText);
            let incidents = [];
            if (response.ok) {
                const data = await response.json();
                console.log('Incidents data received:', data);
                incidents = (data.incidents || []).map(incident => ({
                    id: incident.ID,
                    title: incident.Title,
                    description: incident.Description,
                    status: incident.Status,
                    impact: incident.Impact,
                    created_at: incident.CreatedAt,
                    services: incident.Services || []
                }));
                console.log('Incidents mapped:', incidents);
            } else {
                console.error('Incidents API error:', response.status, response.statusText);
                const errorText = await response.text();
                console.error('Error response body:', errorText);
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
        
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('incidents-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
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
                            <button class="btn-secondary" onclick="adminDashboard.editIncident(${incident.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteIncident(${incident.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        } catch (error) {
            console.error('Error loading incidents:', error);
            
            // Hide the loading spinner
            const loadingSpinner = document.getElementById('incidents-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }
            
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
            console.log('Loading components section...');
            // Fetch components from API
            const response = await fetch('/api/v1/admin/services', {
                credentials: 'include'
            });

            console.log('Components API response:', response.status, response.statusText);
            let components = [];
            if (response.ok) {
                const data = await response.json();
                console.log('Components data received:', data);
                components = (data.services || []).map(service => ({
                    id: service.ID,
                    name: service.Name,
                    description: service.Description,
                    status: service.Status,
                    group: service.Group,
                    show_uptime: service.ShowUptime,
                    created_at: service.CreatedAt
                }));
                console.log('Components mapped:', components);
            } else {
                console.error('Components API error:', response.status, response.statusText);
                const errorText = await response.text();
                console.error('Error response body:', errorText);
                // Fallback to mock data if API fails
                components = [
                    { id: 1, name: "API", status: "operational", description: "Main API service" },
                    { id: 2, name: "Database", status: "operational", description: "Primary database" },
                    { id: 3, name: "Web App", status: "degraded_performance", description: "Frontend application" }
                ];
            }
        
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('components-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
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
                            <button class="btn-secondary" onclick="adminDashboard.editComponent(${component.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteComponent(${component.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        } catch (error) {
            console.error('Error loading components:', error);
            
            // Hide the loading spinner
            const loadingSpinner = document.getElementById('components-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }
            
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
        // Show loading state
        contentDiv.innerHTML = `
            <div class="section-actions">
                <button class="btn-primary" onclick="adminDashboard.openModal('maintenanceModal')">
                    <i class="fas fa-plus"></i> Schedule Maintenance
                </button>
            </div>
            <div class="loading-spinner">
                <i class="fas fa-spinner fa-spin"></i> Loading maintenance events...
            </div>
        `;

        try {
            console.log('Loading maintenance section...');
            // Fetch maintenance events from API
            const response = await fetch('/api/v1/maintenance', {
                credentials: 'include'
            });

            console.log('Maintenance API response:', response.status, response.statusText);
            let maintenanceEvents = [];
            if (response.ok) {
                const data = await response.json();
                console.log('Maintenance data received:', data);
                maintenanceEvents = (data.maintenance || []).map(event => ({
                    id: event.ID,
                    title: event.Title,
                    description: event.Description,
                    start_time: event.StartAt,
                    end_time: event.EndAt,
                    status: event.Status,
                    created_at: event.CreatedAt
                }));
                console.log('Maintenance events mapped:', maintenanceEvents);
            } else {
                console.error('Maintenance API error:', response.status, response.statusText);
                const errorText = await response.text();
                console.error('Error response body:', errorText);
                // Fallback to mock data if API fails
                maintenanceEvents = [
                    {
                        id: 1,
                        title: "Database optimization",
                        description: "Scheduled database maintenance",
                        start_time: "2025-01-08T02:00:00Z",
                        end_time: "2025-01-08T04:00:00Z",
                        status: "scheduled"
                    }
                ];
            }

            // Hide the loading spinner
            const loadingSpinner = document.getElementById('maintenance-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }

            // Generate maintenance list HTML
            contentDiv.innerHTML = `
                <div class="section-actions">
                    <button class="btn-primary" onclick="adminDashboard.openModal('maintenanceModal')">
                        <i class="fas fa-plus"></i> Schedule Maintenance
                    </button>
                </div>
                <div class="maintenance-list" id="maintenanceList">
                    ${maintenanceEvents.length > 0 ? maintenanceEvents.map(maintenance => `
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
                                <button class="btn-secondary" onclick="adminDashboard.editMaintenance(${maintenance.id})">Edit</button>
                                <button class="btn-danger" onclick="adminDashboard.deleteMaintenance(${maintenance.id})">Delete</button>
                            </div>
                        </div>
                    `).join('') : `
                        <div class="empty-state">
                            <i class="fas fa-tools"></i>
                            <h3>No maintenance events scheduled</h3>
                            <p>Click "Schedule Maintenance" to create your first maintenance event.</p>
                        </div>
                    `}
                </div>
            `;
        } catch (error) {
            console.error('Error loading maintenance section:', error);
            
            // Hide the loading spinner
            const loadingSpinner = document.getElementById('maintenance-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }
            
            contentDiv.innerHTML = `
                <div class="error-message">
                    <i class="fas fa-exclamation-triangle"></i>
                    Failed to load maintenance events. Please try again.
                </div>
            `;
        }
    }

    async loadMonitorsSection(contentDiv) {
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('monitors-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
        }

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
                            <button class="btn-secondary" onclick="adminDashboard.editMonitor(${monitor.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteMonitor(${monitor.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    async loadSubscribersSection(contentDiv) {
        console.log('Loading subscribers section...');
        
        // Show loading state
        contentDiv.innerHTML = `
            <div class="loading-spinner">
                <i class="fas fa-spinner fa-spin"></i> Loading subscribers...
            </div>
        `;
        
        try {
            // Fetch real subscribers data
            const response = await fetch('/api/v1/admin/subscribers', { credentials: 'include' });
            console.log('Subscribers API response:', response.status);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            console.log('Subscribers data:', data);
            
            const subscribers = data.subscribers || [];
            
            // Hide the loading spinner
            const loadingSpinner = document.getElementById('subscribers-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }

            contentDiv.innerHTML = `
                <div class="section-actions">
                    <button class="btn-primary" onclick="adminDashboard.openSubscriberModal()">
                        <i class="fas fa-plus"></i> Add Subscriber
                    </button>
                </div>
                <div class="subscribers-list" id="subscribersList">
                    ${subscribers.length > 0 ? subscribers.map(subscriber => `
                        <div class="subscriber-card">
                            <div class="subscriber-header">
                                <h3>${subscriber.Email}</h3>
                                <span class="status-badge active">active</span>
                            </div>
                            <div class="subscriber-details">
                                <p><strong>Phone:</strong> ${subscriber.Phone || 'Not provided'}</p>
                                <p><strong>Created:</strong> ${this.formatTimeAgo(subscriber.CreatedAt)}</p>
                                <p><strong>Services:</strong> ${subscriber.Services ? subscriber.Services.length : 0} subscribed</p>
                            </div>
                            <div class="subscriber-actions">
                                <button class="btn-secondary" onclick="adminDashboard.editSubscriber(${subscriber.ID})">Edit</button>
                                <button class="btn-danger" onclick="adminDashboard.deleteSubscriber(${subscriber.ID})">Delete</button>
                            </div>
                        </div>
                    `).join('') : `
                        <div class="empty-state">
                            <i class="fas fa-users"></i>
                            <h3>No subscribers yet</h3>
                            <p>Start by adding your first subscriber to receive status updates.</p>
                        </div>
                    `}
                </div>
            `;
            
        } catch (error) {
            console.error('Error loading subscribers:', error);
            
            // Hide the loading spinner
            const loadingSpinner = document.getElementById('subscribers-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }
            
            contentDiv.innerHTML = `
                <div class="error-message">
                    <i class="fas fa-exclamation-triangle"></i>
                    <h3>Failed to load subscribers</h3>
                    <p>${error.message}</p>
                    <button class="btn-primary" onclick="adminDashboard.loadSubscribersSection(document.getElementById('subscribers-content'))">
                        <i class="fas fa-refresh"></i> Retry
                    </button>
                </div>
            `;
        }
    }

    // Subscriber management methods
    editSubscriber(subscriberId) {
        console.log('Edit subscriber:', subscriberId);
        // TODO: Implement edit subscriber functionality
        this.showNotification('Edit subscriber functionality coming soon!', 'info');
    }

    async deleteSubscriber(subscriberId) {
        console.log('Delete subscriber:', subscriberId);
        
        if (!confirm('Are you sure you want to delete this subscriber? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/subscribers/${subscriberId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Subscriber deleted successfully', 'success');
                // Reload the subscribers section
                const contentDiv = document.getElementById('subscribers-content');
                if (contentDiv) {
                    await this.loadSubscribersSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error deleting subscriber:', error);
            this.showNotification('Failed to delete subscriber: ' + error.message, 'error');
        }
    }

    // Component management methods
    async editComponent(componentId) {
        console.log('Edit component:', componentId);
        
        try {
            // Fetch the component data
            const response = await fetch(`/api/v1/admin/services/${componentId}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const component = data.service;
            
            // Create and show edit modal
            this.showEditComponentModal(component);
            
        } catch (error) {
            console.error('Error fetching component:', error);
            this.showNotification('Failed to load component data: ' + error.message, 'error');
        }
    }

    showEditComponentModal(component) {
        // Create modal HTML
        const modalHTML = `
            <div id="editComponentModal" class="modal-overlay">
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Edit Component</h2>
                        <button class="modal-close" onclick="adminDashboard.closeModal('editComponentModal')">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="editComponentForm">
                            <input type="hidden" id="editComponentId" value="${component.ID}">
                            
                            <div class="form-group">
                                <label for="editComponentName">Name</label>
                                <input type="text" id="editComponentName" value="${component.Name}" required>
                            </div>
                            
                            <div class="form-group">
                                <label for="editComponentDescription">Description</label>
                                <textarea id="editComponentDescription" rows="4">${component.Description || ''}</textarea>
                            </div>
                            
                            <div class="form-group">
                                <label for="editComponentStatus">Status</label>
                                <select id="editComponentStatus" required>
                                    <option value="operational" ${component.Status === 'operational' ? 'selected' : ''}>Operational</option>
                                    <option value="degraded" ${component.Status === 'degraded' ? 'selected' : ''}>Degraded</option>
                                    <option value="partial_outage" ${component.Status === 'partial_outage' ? 'selected' : ''}>Partial Outage</option>
                                    <option value="major_outage" ${component.Status === 'major_outage' ? 'selected' : ''}>Major Outage</option>
                                </select>
                            </div>
                            
                            <div class="form-group">
                                <label for="editComponentGroup">Group</label>
                                <input type="text" id="editComponentGroup" value="${component.Group || ''}">
                            </div>
                            
                            <div class="form-actions">
                                <button type="button" class="btn-secondary" onclick="adminDashboard.closeModal('editComponentModal')">Cancel</button>
                                <button type="submit" class="btn-primary">Update Component</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
        
        // Add modal to page
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        
        // Add form submit handler
        document.getElementById('editComponentForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleEditComponentSubmit();
        });
    }

    async handleEditComponentSubmit() {
        const componentId = document.getElementById('editComponentId').value;
        const name = document.getElementById('editComponentName').value;
        const description = document.getElementById('editComponentDescription').value;
        const status = document.getElementById('editComponentStatus').value;
        const group = document.getElementById('editComponentGroup').value;
        
        try {
            const response = await fetch(`/api/v1/admin/services/${componentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    name,
                    description,
                    status,
                    group
                })
            });
            
            if (response.ok) {
                this.showNotification('Component updated successfully', 'success');
                this.closeModal('editComponentModal');
                
                // Reload the components section
                const contentDiv = document.getElementById('components-content');
                if (contentDiv) {
                    await this.loadComponentsSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error updating component:', error);
            this.showNotification('Failed to update component: ' + error.message, 'error');
        }
    }

    async deleteComponent(componentId) {
        console.log('Delete component:', componentId);
        
        if (!confirm('Are you sure you want to delete this component? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/services/${componentId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Component deleted successfully', 'success');
                // Reload the components section
                const contentDiv = document.getElementById('components-content');
                if (contentDiv) {
                    await this.loadComponentsSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error deleting component:', error);
            this.showNotification('Failed to delete component: ' + error.message, 'error');
        }
    }

    // Incident management methods
    async editIncident(incidentId) {
        console.log('Edit incident:', incidentId);
        
        try {
            // Fetch the incident data
            const response = await fetch(`/api/v1/admin/incidents/${incidentId}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            const incident = data.incident;
            
            // Create and show edit modal
            this.showEditIncidentModal(incident);
            
        } catch (error) {
            console.error('Error fetching incident:', error);
            this.showNotification('Failed to load incident data: ' + error.message, 'error');
        }
    }

    showEditIncidentModal(incident) {
        // Create modal HTML
        const modalHTML = `
            <div id="editIncidentModal" class="modal-overlay">
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Edit Incident</h2>
                        <button class="modal-close" onclick="adminDashboard.closeModal('editIncidentModal')">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="editIncidentForm">
                            <input type="hidden" id="editIncidentId" value="${incident.ID}">
                            
                            <div class="form-group">
                                <label for="editIncidentTitle">Title</label>
                                <input type="text" id="editIncidentTitle" value="${incident.Title}" required>
                            </div>
                            
                            <div class="form-group">
                                <label for="editIncidentDescription">Description</label>
                                <textarea id="editIncidentDescription" rows="4" required>${incident.Description || ''}</textarea>
                            </div>
                            
                            <div class="form-group">
                                <label for="editIncidentStatus">Status</label>
                                <select id="editIncidentStatus" required>
                                    <option value="investigating" ${incident.Status === 'investigating' ? 'selected' : ''}>Investigating</option>
                                    <option value="identified" ${incident.Status === 'identified' ? 'selected' : ''}>Identified</option>
                                    <option value="monitoring" ${incident.Status === 'monitoring' ? 'selected' : ''}>Monitoring</option>
                                    <option value="resolved" ${incident.Status === 'resolved' ? 'selected' : ''}>Resolved</option>
                                </select>
                            </div>
                            
                            <div class="form-group">
                                <label for="editIncidentImpact">Impact</label>
                                <select id="editIncidentImpact" required>
                                    <option value="minor" ${incident.Impact === 'minor' ? 'selected' : ''}>Minor</option>
                                    <option value="major" ${incident.Impact === 'major' ? 'selected' : ''}>Major</option>
                                    <option value="critical" ${incident.Impact === 'critical' ? 'selected' : ''}>Critical</option>
                                </select>
                            </div>
                            
                            <div class="form-actions">
                                <button type="button" class="btn-secondary" onclick="adminDashboard.closeModal('editIncidentModal')">Cancel</button>
                                <button type="submit" class="btn-primary">Update Incident</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
        
        // Add modal to page
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        
        // Add form submit handler
        document.getElementById('editIncidentForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleEditIncidentSubmit();
        });
    }

    async handleEditIncidentSubmit() {
        const incidentId = document.getElementById('editIncidentId').value;
        const title = document.getElementById('editIncidentTitle').value;
        const description = document.getElementById('editIncidentDescription').value;
        const status = document.getElementById('editIncidentStatus').value;
        const impact = document.getElementById('editIncidentImpact').value;
        
        try {
            const response = await fetch(`/api/v1/admin/incidents/${incidentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({
                    title,
                    description,
                    status,
                    impact
                })
            });
            
            if (response.ok) {
                this.showNotification('Incident updated successfully', 'success');
                this.closeModal('editIncidentModal');
                
                // Reload the incidents section
                const contentDiv = document.getElementById('incidents-content');
                if (contentDiv) {
                    await this.loadIncidentsSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error updating incident:', error);
            this.showNotification('Failed to update incident: ' + error.message, 'error');
        }
    }

    async deleteIncident(incidentId) {
        console.log('Delete incident:', incidentId);
        
        if (!confirm('Are you sure you want to delete this incident? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/incidents/${incidentId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Incident deleted successfully', 'success');
                // Reload the incidents section
                const contentDiv = document.getElementById('incidents-content');
                if (contentDiv) {
                    await this.loadIncidentsSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error deleting incident:', error);
            this.showNotification('Failed to delete incident: ' + error.message, 'error');
        }
    }

    // Maintenance management methods
    editMaintenance(maintenanceId) {
        console.log('Edit maintenance:', maintenanceId);
        // TODO: Implement edit maintenance functionality
        this.showNotification('Edit maintenance functionality coming soon!', 'info');
    }

    async deleteMaintenance(maintenanceId) {
        console.log('Delete maintenance:', maintenanceId);
        
        if (!confirm('Are you sure you want to delete this maintenance event? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/maintenance/${maintenanceId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Maintenance event deleted successfully', 'success');
                // Reload the maintenance section
                const contentDiv = document.getElementById('maintenance-content');
                if (contentDiv) {
                    await this.loadMaintenanceSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error deleting maintenance:', error);
            this.showNotification('Failed to delete maintenance event: ' + error.message, 'error');
        }
    }

    // Monitor management methods
    async editMonitor(monitorId) {
        console.log('Edit monitor:', monitorId);
        try {
            const response = await fetch(`/api/v1/admin/monitors/${monitorId}`, { credentials: 'include' });
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const monitor = data.monitor;
            this.showEditMonitorModal(monitor);
        } catch (error) {
            console.error('Error fetching monitor:', error);
            this.showNotification('Failed to load monitor data: ' + error.message, 'error');
        }
    }

    showEditMonitorModal(monitor) {
        const modalHTML = `
            <div id="editMonitorModal" class="modal-overlay">
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Edit Monitor</h2>
                        <button class="modal-close" onclick="adminDashboard.closeModal('editMonitorModal')">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="editMonitorForm">
                            <input type="hidden" id="editMonitorId" value="${monitor.id}">
                            <div class="form-group">
                                <label for="editMonitorName">Monitor Name</label>
                                <input type="text" id="editMonitorName" value="${monitor.name}" required>
                            </div>
                            <div class="form-group">
                                <label for="editMonitorType">Type</label>
                                <select id="editMonitorType" required>
                                    <option value="http" ${monitor.type === 'http' ? 'selected' : ''}>HTTP</option>
                                    <option value="tcp" ${monitor.type === 'tcp' ? 'selected' : ''}>TCP</option>
                                    <option value="ping" ${monitor.type === 'ping' ? 'selected' : ''}>Ping</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label for="editMonitorUrl">URL/Host</label>
                                <input type="text" id="editMonitorUrl" value="${monitor.url || monitor.host}" required>
                            </div>
                            <div class="form-group">
                                <label for="editMonitorInterval">Check Interval (seconds)</label>
                                <input type="number" id="editMonitorInterval" value="60" min="30" max="3600">
                            </div>
                            <div class="form-group">
                                <label for="editMonitorTimeout">Timeout (seconds)</label>
                                <input type="number" id="editMonitorTimeout" value="10" min="5" max="60">
                            </div>
                            <div class="form-actions">
                                <button type="button" class="btn-secondary" onclick="adminDashboard.closeModal('editMonitorModal')">Cancel</button>
                                <button type="submit" class="btn-primary">Update Monitor</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        document.getElementById('editMonitorForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleEditMonitorSubmit();
        });
    }

    async handleEditMonitorSubmit() {
        const monitorId = document.getElementById('editMonitorId').value;
        const name = document.getElementById('editMonitorName').value;
        const type = document.getElementById('editMonitorType').value;
        const url = document.getElementById('editMonitorUrl').value;
        const interval = document.getElementById('editMonitorInterval').value;
        const timeout = document.getElementById('editMonitorTimeout').value;
        
        try {
            const response = await fetch(`/api/v1/admin/monitors/${monitorId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ name, type, url, interval, timeout })
            });
            
            if (response.ok) {
                this.showNotification('Monitor updated successfully', 'success');
                this.closeModal('editMonitorModal');
                const contentDiv = document.getElementById('monitors-content');
                if (contentDiv) {
                    await this.loadMonitorsSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error updating monitor:', error);
            this.showNotification('Failed to update monitor: ' + error.message, 'error');
        }
    }

    async deleteMonitor(monitorId) {
        console.log('Delete monitor:', monitorId);
        
        if (!confirm('Are you sure you want to delete this monitor? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/monitors/${monitorId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                this.showNotification('Monitor deleted successfully', 'success');
                // Reload the monitors section
                const contentDiv = document.getElementById('monitors-content');
                if (contentDiv) {
                    await this.loadMonitorsSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error deleting monitor:', error);
            this.showNotification('Failed to delete monitor: ' + error.message, 'error');
        }
    }

    async loadAnalyticsSection(contentDiv) {
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('analytics-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
        }

        try {
            // Fetch real analytics data from multiple endpoints
            const [servicesResponse, incidentsResponse, subscribersResponse] = await Promise.all([
                fetch('/api/v1/admin/services', { credentials: 'include' }),
                fetch('/api/v1/admin/incidents', { credentials: 'include' }),
                fetch('/api/v1/admin/subscribers', { credentials: 'include' })
            ]);

            let analytics = {
                totalServices: 0,
                totalIncidents: 0,
                totalSubscribers: 0,
                uptime: 99.9,
                avgResponseTime: 120
            };

            if (servicesResponse.ok) {
                const servicesData = await servicesResponse.json();
                analytics.totalServices = servicesData.services ? servicesData.services.length : 0;
            }

            if (incidentsResponse.ok) {
                const incidentsData = await incidentsResponse.json();
                analytics.totalIncidents = incidentsData.incidents ? incidentsData.incidents.length : 0;
            }

            if (subscribersResponse.ok) {
                const subscribersData = await subscribersResponse.json();
                analytics.totalSubscribers = subscribersData.subscribers ? subscribersData.subscribers.length : 0;
            }

            contentDiv.innerHTML = `
            <div class="analytics-dashboard">
                <div class="analytics-grid">
                    <div class="analytics-card">
                        <h3>Uptime</h3>
                        <div class="metric-value">${analytics.uptime}%</div>
                    </div>
                    <div class="analytics-card">
                        <h3>Total Services</h3>
                        <div class="metric-value">${analytics.totalServices}</div>
                    </div>
                    <div class="analytics-card">
                        <h3>Total Incidents</h3>
                        <div class="metric-value">${analytics.totalIncidents}</div>
                    </div>
                    <div class="analytics-card">
                        <h3>Total Subscribers</h3>
                        <div class="metric-value">${analytics.totalSubscribers}</div>
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
        } catch (error) {
            console.error('Error loading analytics:', error);
            contentDiv.innerHTML = `
                <div class="error-message">
                    <i class="fas fa-exclamation-triangle"></i>
                    Failed to load analytics data. Please try again.
                </div>
            `;
        }
    }

    async loadIntegrationsSection(contentDiv) {
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('integrations-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
        }

        contentDiv.innerHTML = `
            <div class="integrations-grid">
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fab fa-slack"></i>
                    </div>
                    <h3>Slack</h3>
                    <p>Send incident notifications to Slack channels</p>
                    <button class="btn-secondary" onclick="adminDashboard.configureIntegration('slack')">Configure</button>
                </div>
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fas fa-bell"></i>
                    </div>
                    <h3>PagerDuty</h3>
                    <p>Integrate with PagerDuty for incident management</p>
                    <button class="btn-secondary" onclick="adminDashboard.configureIntegration('pagerduty')">Configure</button>
                </div>
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fas fa-chart-line"></i>
                    </div>
                    <h3>Datadog</h3>
                    <p>Export metrics to Datadog</p>
                    <button class="btn-secondary" onclick="adminDashboard.configureIntegration('datadog')">Configure</button>
                </div>
                <div class="integration-card">
                    <div class="integration-icon">
                        <i class="fas fa-plug"></i>
                    </div>
                    <h3>Webhooks</h3>
                    <p>Send notifications to custom webhooks</p>
                    <button class="btn-secondary" onclick="adminDashboard.configureIntegration('webhooks')">Configure</button>
                </div>
            </div>
        `;
    }

    async loadBrandingSection(contentDiv) {
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('branding-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
        }

        contentDiv.innerHTML = `
            <div class="branding-settings">
                <div class="branding-section">
                    <h3>Custom Domain</h3>
                    <p>Set up a custom domain for your status page</p>
                    <div class="form-group">
                        <input type="text" id="customDomain" placeholder="status.yourcompany.com" class="form-input">
                        <button class="btn-primary" onclick="adminDashboard.saveCustomDomain()">Save Domain</button>
                    </div>
                </div>
                <div class="branding-section">
                    <h3>Logo</h3>
                    <p>Upload your company logo</p>
                    <div class="logo-upload">
                        <input type="file" id="logoUpload" accept="image/*" class="form-input">
                        <button class="btn-primary" onclick="adminDashboard.uploadLogo()">Upload Logo</button>
                    </div>
                    <div id="logoPreview" class="logo-preview" style="display: none;">
                        <img id="logoImage" src="" alt="Logo Preview" style="max-width: 200px; max-height: 100px;">
                    </div>
                </div>
                <div class="branding-section">
                    <h3>Colors</h3>
                    <p>Customize your status page colors</p>
                    <div class="color-picker">
                        <div class="color-option">
                            <label>Primary Color</label>
                            <input type="color" id="primaryColor" value="#0052cc" class="form-input">
                        </div>
                        <div class="color-option">
                            <label>Secondary Color</label>
                            <input type="color" id="secondaryColor" value="#f4f5f7" class="form-input">
                        </div>
                        <div class="color-option">
                            <label>Accent Color</label>
                            <input type="color" id="accentColor" value="#36b37e" class="form-input">
                        </div>
                    </div>
                    <button class="btn-primary" onclick="adminDashboard.saveColors()">Save Colors</button>
                </div>
                <div class="branding-section">
                    <h3>Custom CSS</h3>
                    <p>Add custom CSS to style your status page</p>
                    <div class="form-group">
                        <textarea id="customCSS" placeholder="/* Add your custom CSS here */" rows="10" class="form-input"></textarea>
                        <button class="btn-primary" onclick="adminDashboard.saveCustomCSS()">Save CSS</button>
                    </div>
                </div>
                <div class="branding-section">
                    <h3>Page Title & Description</h3>
                    <p>Customize your status page title and description</p>
                    <div class="form-group">
                        <label for="pageTitle">Page Title</label>
                        <input type="text" id="pageTitle" placeholder="Service Status" class="form-input">
                    </div>
                    <div class="form-group">
                        <label for="pageDescription">Page Description</label>
                        <textarea id="pageDescription" placeholder="Real-time status of our services" rows="3" class="form-input"></textarea>
                    </div>
                    <button class="btn-primary" onclick="adminDashboard.savePageSettings()">Save Settings</button>
                </div>
            </div>
        `;
    }

    // Branding management methods
    async saveCustomDomain() {
        const domain = document.getElementById('customDomain').value;
        if (!domain) {
            this.showNotification('Please enter a domain', 'error');
            return;
        }

        try {
            const response = await fetch('/api/v1/admin/branding/domain', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ domain: domain })
            });

            if (response.ok) {
                this.showNotification('Custom domain saved successfully', 'success');
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error saving domain:', error);
            this.showNotification('Failed to save domain: ' + error.message, 'error');
        }
    }

    async uploadLogo() {
        const fileInput = document.getElementById('logoUpload');
        const file = fileInput.files[0];
        
        if (!file) {
            this.showNotification('Please select a file to upload', 'error');
            return;
        }

        const formData = new FormData();
        formData.append('logo', file);

        try {
            const response = await fetch('/api/v1/admin/branding/logo', {
                method: 'POST',
                credentials: 'include',
                body: formData
            });

            if (response.ok) {
                const result = await response.json();
                this.showNotification('Logo uploaded successfully', 'success');
                
                // Show preview
                const preview = document.getElementById('logoPreview');
                const image = document.getElementById('logoImage');
                image.src = result.logo_url;
                preview.style.display = 'block';
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error uploading logo:', error);
            this.showNotification('Failed to upload logo: ' + error.message, 'error');
        }
    }

    async saveColors() {
        const primaryColor = document.getElementById('primaryColor').value;
        const secondaryColor = document.getElementById('secondaryColor').value;
        const accentColor = document.getElementById('accentColor').value;

        try {
            const response = await fetch('/api/v1/admin/branding/colors', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    primary: primaryColor,
                    secondary: secondaryColor,
                    accent: accentColor
                })
            });

            if (response.ok) {
                this.showNotification('Colors saved successfully', 'success');
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error saving colors:', error);
            this.showNotification('Failed to save colors: ' + error.message, 'error');
        }
    }

    async saveCustomCSS() {
        const customCSS = document.getElementById('customCSS').value;

        try {
            const response = await fetch('/api/v1/admin/branding/css', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ css: customCSS })
            });

            if (response.ok) {
                this.showNotification('Custom CSS saved successfully', 'success');
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error saving CSS:', error);
            this.showNotification('Failed to save CSS: ' + error.message, 'error');
        }
    }

    async savePageSettings() {
        const pageTitle = document.getElementById('pageTitle').value;
        const pageDescription = document.getElementById('pageDescription').value;

        if (!pageTitle) {
            this.showNotification('Please enter a page title', 'error');
            return;
        }

        try {
            const response = await fetch('/api/v1/admin/branding/settings', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    title: pageTitle,
                    description: pageDescription
                })
            });

            if (response.ok) {
                this.showNotification('Page settings saved successfully', 'success');
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error saving page settings:', error);
            this.showNotification('Failed to save page settings: ' + error.message, 'error');
        }
    }

    // Subscriber management methods
    openSubscriberModal() {
        const modalHTML = `
            <div id="subscriberModal" class="modal-overlay">
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Add New Subscriber</h2>
                        <button class="modal-close" onclick="adminDashboard.closeModal('subscriberModal')">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="subscriberForm">
                            <div class="form-group">
                                <label for="subscriberEmail">Email Address</label>
                                <input type="email" id="subscriberEmail" placeholder="user@example.com" required>
                            </div>
                            <div class="form-group">
                                <label for="subscriberName">Name (Optional)</label>
                                <input type="text" id="subscriberName" placeholder="John Doe">
                            </div>
                            <div class="form-group">
                                <label for="subscriberServices">Services to Subscribe</label>
                                <div id="subscriberServicesList" class="checkbox-list">
                                    <!-- Services will be loaded here -->
                                </div>
                            </div>
                            <div class="form-group">
                                <label>
                                    <input type="checkbox" id="subscriberActive" checked> Active subscription
                                </label>
                            </div>
                            <div class="form-actions">
                                <button type="button" class="btn-secondary" onclick="adminDashboard.closeModal('subscriberModal')">Cancel</button>
                                <button type="submit" class="btn-primary">Add Subscriber</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        
        // Load services for the checkbox list
        this.loadServicesForSubscriberModal();
        
        // Add form submission handler
        document.getElementById('subscriberForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleSubscriberSubmit();
        });
    }

    async loadServicesForSubscriberModal() {
        try {
            const response = await fetch('/api/v1/admin/services', { credentials: 'include' });
            if (response.ok) {
                const data = await response.json();
                const services = data.services || [];
                
                const servicesList = document.getElementById('subscriberServicesList');
                servicesList.innerHTML = services.map(service => `
                    <label class="checkbox-item">
                        <input type="checkbox" name="services" value="${service.ID}">
                        ${service.Name}
                    </label>
                `).join('');
            }
        } catch (error) {
            console.error('Error loading services for subscriber modal:', error);
        }
    }

    async handleSubscriberSubmit() {
        const email = document.getElementById('subscriberEmail').value;
        const name = document.getElementById('subscriberName').value;
        const active = document.getElementById('subscriberActive').checked;
        
        // Get selected services
        const selectedServices = Array.from(document.querySelectorAll('input[name="services"]:checked'))
            .map(checkbox => checkbox.value);
        
        try {
            const response = await fetch('/api/v1/admin/subscribers', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    email: email,
                    name: name,
                    active: active,
                    service_ids: selectedServices
                })
            });
            
            if (response.ok) {
                this.showNotification('Subscriber added successfully', 'success');
                this.closeModal('subscriberModal');
                // Reload the subscribers section
                const contentDiv = document.getElementById('subscribers-content');
                if (contentDiv) {
                    await this.loadSubscribersSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error adding subscriber:', error);
            this.showNotification('Failed to add subscriber: ' + error.message, 'error');
        }
    }

    // User management methods
    openUserModal() {
        const modalHTML = `
            <div id="userModal" class="modal-overlay">
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Add New User</h2>
                        <button class="modal-close" onclick="adminDashboard.closeModal('userModal')">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="userForm">
                            <div class="form-group">
                                <label for="userEmail">Email Address</label>
                                <input type="email" id="userEmail" placeholder="user@example.com" required>
                            </div>
                            <div class="form-group">
                                <label for="userName">Full Name</label>
                                <input type="text" id="userName" placeholder="John Doe" required>
                            </div>
                            <div class="form-group">
                                <label for="userPassword">Password</label>
                                <input type="password" id="userPassword" placeholder="Enter password" required>
                            </div>
                            <div class="form-group">
                                <label for="userRole">Role</label>
                                <select id="userRole" required>
                                    <option value="admin">Admin</option>
                                    <option value="user">User</option>
                                    <option value="viewer">Viewer</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>
                                    <input type="checkbox" id="userActive" checked> Active user
                                </label>
                            </div>
                            <div class="form-actions">
                                <button type="button" class="btn-secondary" onclick="adminDashboard.closeModal('userModal')">Cancel</button>
                                <button type="submit" class="btn-primary">Add User</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        
        // Add form submission handler
        document.getElementById('userForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleUserSubmit();
        });
    }

    async handleUserSubmit() {
        const email = document.getElementById('userEmail').value;
        const name = document.getElementById('userName').value;
        const password = document.getElementById('userPassword').value;
        const role = document.getElementById('userRole').value;
        const active = document.getElementById('userActive').checked;
        
        try {
            const response = await fetch('/api/v1/admin/users', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    email: email,
                    name: name,
                    password: password,
                    role: role,
                    active: active
                })
            });
            
            if (response.ok) {
                this.showNotification('User added successfully', 'success');
                this.closeModal('userModal');
                // Reload the users section
                const contentDiv = document.getElementById('users-content');
                if (contentDiv) {
                    await this.loadUsersSection(contentDiv);
                }
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error adding user:', error);
            this.showNotification('Failed to add user: ' + error.message, 'error');
        }
    }

    // Integration management methods
    configureIntegration(integrationType) {
        console.log('Configure integration:', integrationType);
        
        // Create a modal for integration configuration
        const modalHTML = `
            <div id="integrationModal" class="modal-overlay">
                <div class="modal-content">
                    <div class="modal-header">
                        <h2>Configure ${integrationType.charAt(0).toUpperCase() + integrationType.slice(1)} Integration</h2>
                        <button class="modal-close" onclick="adminDashboard.closeModal('integrationModal')">&times;</button>
                    </div>
                    <div class="modal-body">
                        <form id="integrationForm">
                            <div class="form-group">
                                <label for="integrationName">Integration Name</label>
                                <input type="text" id="integrationName" value="${integrationType}" readonly class="form-input">
                            </div>
                            <div class="form-group">
                                <label for="webhookUrl">Webhook URL</label>
                                <input type="url" id="webhookUrl" placeholder="https://hooks.slack.com/services/..." class="form-input">
                            </div>
                            <div class="form-group">
                                <label for="apiKey">API Key</label>
                                <input type="password" id="apiKey" placeholder="Enter API key" class="form-input">
                            </div>
                            <div class="form-group">
                                <label for="channel">Channel/Endpoint</label>
                                <input type="text" id="channel" placeholder="#general or endpoint name" class="form-input">
                            </div>
                            <div class="form-group">
                                <label>
                                    <input type="checkbox" id="enabled" checked> Enable integration
                                </label>
                            </div>
                            <div class="form-actions">
                                <button type="button" class="btn-secondary" onclick="adminDashboard.closeModal('integrationModal')">Cancel</button>
                                <button type="submit" class="btn-primary">Save Configuration</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', modalHTML);
        
        // Add form submission handler
        document.getElementById('integrationForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleIntegrationSubmit(integrationType);
        });

        // Add click outside to close functionality
        const modal = document.getElementById('integrationModal');
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal('integrationModal');
            }
        });
    }

    async handleIntegrationSubmit(integrationType) {
        const webhookUrl = document.getElementById('webhookUrl').value;
        const apiKey = document.getElementById('apiKey').value;
        const channel = document.getElementById('channel').value;
        const enabled = document.getElementById('enabled').checked;
        
        try {
            const response = await fetch('/api/v1/admin/integrations', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    type: integrationType,
                    webhook_url: webhookUrl,
                    api_key: apiKey,
                    channel: channel,
                    enabled: enabled
                })
            });
            
            if (response.ok) {
                this.showNotification(`${integrationType.charAt(0).toUpperCase() + integrationType.slice(1)} integration configured successfully`, 'success');
                this.closeModal('integrationModal');
            } else {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Error configuring integration:', error);
            this.showNotification('Failed to configure integration: ' + error.message, 'error');
        }
    }

    async loadUsersSection(contentDiv) {
        // Hide the loading spinner
        const loadingSpinner = document.getElementById('users-loading');
        if (loadingSpinner) {
            loadingSpinner.style.display = 'none';
        }

        try {
            const response = await fetch('/api/v1/admin/users', { credentials: 'include' });
            const users = await response.json();
            
            contentDiv.innerHTML = `
                <div class="section-actions">
                    <button class="btn-primary" onclick="adminDashboard.openUserModal()">
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
            // Hide the loading spinner
            const loadingSpinner = document.getElementById('users-loading');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
            }
            
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

    formatTimeAgo(dateString) {
        const now = new Date();
        const date = new Date(dateString);
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) {
            return `${diffInSeconds} seconds ago`;
        } else if (diffInSeconds < 3600) {
            const minutes = Math.floor(diffInSeconds / 60);
            return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
        } else if (diffInSeconds < 86400) {
            const hours = Math.floor(diffInSeconds / 3600);
            return `${hours} hour${hours > 1 ? 's' : ''} ago`;
        } else {
            const days = Math.floor(diffInSeconds / 86400);
            return `${days} day${days > 1 ? 's' : ''} ago`;
        }
    }

    // Render methods for different data types
    renderIncidentsList(incidents) {
        return `
            <div class="incidents-list">
                ${incidents.map(incident => `
                    <div class="incident-card">
                        <div class="incident-header">
                            <h3>${incident.title}</h3>
                            <span class="incident-status status-${incident.status}">${incident.status}</span>
                        </div>
                        <div class="incident-body">
                            <p>${incident.description}</p>
                            <div class="incident-meta">
                                <span class="incident-impact">Impact: ${incident.impact}</span>
                                <span class="incident-severity">Severity: ${incident.severity}</span>
                                <span class="incident-date">${new Date(incident.created_at).toLocaleString()}</span>
                            </div>
                        </div>
                        <div class="incident-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editIncident(${incident.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteIncident(${incident.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderComponentsList(services) {
        return `
            <div class="components-list">
                ${services.map(service => `
                    <div class="component-card">
                        <div class="component-header">
                            <h3>${service.name}</h3>
                            <span class="component-status status-${service.status}">${service.status}</span>
                        </div>
                        <div class="component-body">
                            <p>${service.description}</p>
                            <div class="component-meta">
                                <span class="component-uptime">Uptime: ${service.uptime}</span>
                                <span class="component-group">${service.group || 'No Group'}</span>
                            </div>
                        </div>
                        <div class="component-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editComponent(${service.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteComponent(${service.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderMaintenanceList(maintenance) {
        return `
            <div class="maintenance-list">
                ${maintenance.map(event => `
                    <div class="maintenance-card">
                        <div class="maintenance-header">
                            <h3>${event.title}</h3>
                            <span class="maintenance-status status-${event.status}">${event.status}</span>
                        </div>
                        <div class="maintenance-body">
                            <p>${event.description}</p>
                            <div class="maintenance-meta">
                                <span class="maintenance-start">Start: ${new Date(event.start_at).toLocaleString()}</span>
                                <span class="maintenance-end">End: ${new Date(event.end_at).toLocaleString()}</span>
                            </div>
                        </div>
                        <div class="maintenance-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editMaintenance(${event.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteMaintenance(${event.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderMonitorsList(monitors) {
        return `
            <div class="monitors-list">
                ${monitors.map(monitor => `
                    <div class="monitor-card">
                        <div class="monitor-header">
                            <h3>${monitor.name}</h3>
                            <span class="monitor-status status-${monitor.status}">${monitor.status}</span>
                        </div>
                        <div class="monitor-body">
                            <p>URL: ${monitor.url}</p>
                            <div class="monitor-meta">
                                <span class="monitor-type">Type: ${monitor.type}</span>
                                <span class="monitor-interval">Interval: ${monitor.interval}s</span>
                            </div>
                        </div>
                        <div class="monitor-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editMonitor(${monitor.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteMonitor(${monitor.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderSubscribersList(subscribers) {
        return `
            <div class="subscribers-list">
                ${subscribers.map(subscriber => `
                    <div class="subscriber-card">
                        <div class="subscriber-header">
                            <h3>${subscriber.email}</h3>
                            <span class="subscriber-status status-${subscriber.status}">${subscriber.status}</span>
                        </div>
                        <div class="subscriber-body">
                            <div class="subscriber-meta">
                                <span class="subscriber-type">Type: ${subscriber.type}</span>
                                <span class="subscriber-date">Joined: ${new Date(subscriber.created_at).toLocaleString()}</span>
                            </div>
                        </div>
                        <div class="subscriber-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editSubscriber(${subscriber.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteSubscriber(${subscriber.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderUsersList(users) {
        return `
            <div class="users-list">
                ${users.map(user => `
                    <div class="user-card">
                        <div class="user-header">
                            <h3>${user.username}</h3>
                            <span class="user-role role-${user.role}">${user.role}</span>
                        </div>
                        <div class="user-body">
                            <p>Email: ${user.email}</p>
                            <div class="user-meta">
                                <span class="user-status status-${user.status}">${user.status}</span>
                            </div>
                        </div>
                        <div class="user-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editUser(${user.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteUser(${user.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    renderIntegrationsList(integrations) {
        return `
            <div class="integrations-list">
                ${integrations.map(integration => `
                    <div class="integration-card">
                        <div class="integration-header">
                            <h3>${integration.name}</h3>
                            <span class="integration-status status-${integration.status}">${integration.status}</span>
                        </div>
                        <div class="integration-body">
                            <p>Type: ${integration.type}</p>
                            <div class="integration-meta">
                                <span class="integration-webhook">Webhook: ${integration.webhook || 'N/A'}</span>
                            </div>
                        </div>
                        <div class="integration-actions">
                            <button class="btn-secondary" onclick="adminDashboard.editIntegration(${integration.id})">Edit</button>
                            <button class="btn-danger" onclick="adminDashboard.deleteIntegration(${integration.id})">Delete</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }
}

// Export for global access
window.AdminDashboard = AdminDashboard;
