// SaaS Admin Dashboard JavaScript - Based on Working Admin.js

class SaaSAdminDashboard {
    constructor() {
        console.log('SaaSAdminDashboard constructor called');
        this.currentSection = 'overview';
        this.init();
    }

    init() {
        console.log('SaaSAdminDashboard init called');
        this.initializeSidebarState();
        console.log('Setting up event listeners...');
        this.setupEventListeners();
        console.log('Event listeners setup complete');
        
            // Check for initial hash and navigate accordingly
            setTimeout(() => {
                const initialHash = window.location.hash.substring(1);
                console.log('Initial hash:', initialHash);
                
                if (initialHash && initialHash !== 'overview') {
                    // Navigate to the specified section
                this.navigateToSection(initialHash);
                } else {
                    // Default to overview
                this.currentSection = 'overview';
                this.updateActiveNavigation('overview');
                this.updatePageTitle('overview');
                    this.loadOverviewData();
                }
            }, 100);
    }

    initializeSidebarState() {
        const sidebar = document.querySelector('.sidebar');
        const sidebarToggle = document.getElementById('sidebarToggle');
        
        if (sidebar && sidebarToggle) {
            // Get saved state from localStorage
            const savedState = localStorage.getItem('saas-sidebar-state');
            console.log('Saved SaaS sidebar state:', savedState);
            
            if (savedState === 'collapsed') {
                // Set to collapsed state
                sidebar.classList.remove('open');
                sidebar.classList.add('collapsed');
                sidebarToggle.querySelector('i').className = 'fas fa-chevron-right';
                console.log('SaaS Sidebar initialized as collapsed');
            } else {
                // Default to open state
                sidebar.classList.remove('collapsed');
                sidebar.classList.add('open');
                sidebarToggle.querySelector('i').className = 'fas fa-chevron-left';
                console.log('SaaS Sidebar initialized as open');
            }
        }
    }


    setupEventListeners() {
        // Sidebar navigation
        this.setupSidebarNavigation();

        // Sidebar toggle
        this.setupSidebarToggle();

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
                // If no hash, go to overview
                this.currentSection = 'overview';
                this.updateActiveNavigation('overview');
                this.updatePageTitle('overview');
                this.loadOverviewData();
            }
        });
    }

    setupSidebarNavigation() {
        const navLinks = document.querySelectorAll('.sidebar-nav a');
        console.log('Setting up sidebar navigation, found links:', navLinks.length);
        
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const href = link.getAttribute('href');
                console.log('Nav link clicked:', href);
                
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
    }

    setupSidebarToggle() {
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
                    localStorage.setItem('saas-sidebar-state', 'collapsed');
                    console.log('Sidebar collapsed');
                } else if (sidebar.classList.contains('collapsed')) {
                    // If collapsed, open it
                    sidebar.classList.remove('collapsed');
                    sidebar.classList.add('open');
                    sidebarToggle.querySelector('i').className = 'fas fa-chevron-left';
                    localStorage.setItem('saas-sidebar-state', 'open');
                    console.log('Sidebar opened');
                } else {
                    // If hidden, open it
                    sidebar.classList.add('open');
                    sidebarToggle.querySelector('i').className = 'fas fa-chevron-left';
                    localStorage.setItem('saas-sidebar-state', 'open');
                    console.log('Sidebar opened from hidden');
                }
            };
            
            // Add event listener to sidebar toggle button
            sidebarToggle.addEventListener('click', toggleSidebar);

            // DISABLED: Don't close sidebar when clicking outside
            // This was causing the sidebar to disappear when clicking elsewhere
            // document.addEventListener('click', (e) => {
            //     if (sidebar.classList.contains('open') && 
            //         !sidebar.contains(e.target) && 
            //         !sidebarToggle.contains(e.target)) {
            //         sidebar.classList.remove('open');
            //         sidebar.classList.add('collapsed');
            //         sidebarToggle.querySelector('i').className = 'fas fa-chevron-right';
            //         console.log('Sidebar collapsed by outside click');
            //     }
            // });

            // Don't auto-close on navigation links - let user decide
        }
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

        // Update current section
        this.currentSection = section;

        // Update active navigation state
        this.updateActiveNavigation(section);

        // Update page title
        this.updatePageTitle(section);

        // Load section-specific data
        await this.loadSectionData(section);
    }

    updateActiveNavigation(section) {
        console.log('Updating active navigation for section:', section);
        
        // Remove active class from all navigation items (both li and a elements)
        const navItems = document.querySelectorAll('.sidebar-nav .nav-item');
        navItems.forEach(item => {
            item.classList.remove('active');
        });
        
        const navLinks = document.querySelectorAll('.sidebar-nav a');
        navLinks.forEach(link => {
            link.classList.remove('active');
        });
        
        // Add active class to the current section's navigation item
        const activeLink = document.querySelector(`.sidebar-nav a[href="#${section}"]`);
        if (activeLink) {
            // Add active class to both the link and its parent li
            activeLink.classList.add('active');
            activeLink.parentElement.classList.add('active');
            console.log('Set active navigation for:', section);
        } else {
            console.warn('Could not find navigation link for section:', section);
        }
    }

    updatePageTitle(section) {
        console.log('Updating page title for section:', section);
        
        const titleMap = {
                'overview': 'SaaS Platform Overview',
                'tenants': 'Tenant Management',
                'subscriptions': 'Subscription Management',
            'plans': 'Plan Management',
            'analytics': 'Analytics Dashboard',
            'billing': 'Billing Management',
            'feature-flags': 'Feature Flags',
            'tenant-features': 'Tenant Features',
            'settings': 'System Settings',
            'users': 'User Management',
            'integrations': 'Integrations',
            'notifications': 'Notification Settings'
        };
        
        const newTitle = titleMap[section] || 'SaaS Platform Overview';
        
        // Update the main page title
        const pageTitle = document.querySelector('.page-title');
        if (pageTitle) {
            pageTitle.textContent = newTitle;
            console.log('Updated page title to:', newTitle);
        } else {
            console.warn('Could not find page title element');
        }
    }

    async loadSectionData(section) {
        console.log('Loading data for section:', section);

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
                case 'analytics':
                    await this.loadAnalyticsData();
                    break;
                case 'billing':
                    await this.loadBillingData();
                    break;
                case 'pricing':
                    // Pricing section doesn't need data loading, it's static
                    break;
                case 'feature-flags':
                    await this.loadFeatureFlagsData();
                    break;
                case 'tenant-features':
                    await this.loadTenantFeaturesData();
                    break;
                case 'settings':
                    await this.loadSettingsData();
                    break;
                case 'users':
                    await this.loadUsersData();
                    break;
                case 'integrations':
                    await this.loadIntegrationsData();
                    break;
                case 'notifications':
                    await this.loadNotificationsData();
                    break;
                default:
                    console.log('No specific data loading for section:', section);
            }
        } catch (error) {
            console.error('Error loading section data:', error);
        }
    }

    async loadOverviewData() {
        console.log('Loading overview data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Overview data received:', data);
            this.renderOverview(data);
        } catch (error) {
            console.error('Error loading overview:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('overview-loading');
            const content = document.getElementById('overview-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden overview loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed overview content');
            }
        }
    }

    renderOverview(data) {
        console.log('Rendering overview with data:', data);
        // Overview rendering logic here
    }

    async loadTenantsData() {
        console.log('Loading tenants data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/tenants');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Tenants data received:', data);
            this.renderTenantsTable(data.tenants || []);
        } catch (error) {
            console.error('Error loading tenants:', error);
            this.showError('Failed to load tenants: ' + error.message);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('tenants-loading');
            const content = document.getElementById('tenants-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden tenants loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed tenants content');
            }
        }
    }

    renderTenantsTable(tenants) {
        console.log('Rendering tenants table with:', tenants.length, 'tenants');
        const tbody = document.getElementById('tenantsTableBody');
        if (!tbody) {
            console.error('tenantsTableBody not found');
            return;
        }

        tbody.innerHTML = '';
            tenants.forEach(tenant => {
        const row = document.createElement('tr');
        row.innerHTML = `
                <td>${tenant.Name || 'N/A'}</td>
                <td>${tenant.Slug || 'N/A'}</td>
                <td>${tenant.Plan || 'N/A'}</td>
                <td>${tenant.Status || 'N/A'}</td>
                <td>${tenant.CreatedAt ? new Date(tenant.CreatedAt).toLocaleDateString() : 'N/A'}</td>
                <td>
                    <button class="btn-secondary" onclick="window.saasAdminInstance.editTenant('${tenant.ID}')">Edit</button>
                    <button class="btn-danger" onclick="window.saasAdminInstance.deleteTenant('${tenant.ID}')">Delete</button>
                </td>
        `;
            tbody.appendChild(row);
        });
    }

    async loadSubscriptionsData() {
        console.log('Loading subscriptions data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/subscriptions');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Subscriptions data received:', data);
            this.renderSubscriptionsTable(data.subscriptions || []);
        } catch (error) {
            console.error('Error loading subscriptions:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('subscriptions-loading');
            const content = document.getElementById('subscriptions-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden subscriptions loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed subscriptions content');
            }
        }
    }

    renderSubscriptionsTable(subscriptions) {
        console.log('Rendering subscriptions table with:', subscriptions.length, 'subscriptions');
        const tbody = document.getElementById('subscriptionsTableBody');
        if (!tbody) {
            console.error('subscriptionsTableBody not found');
            return;
        }

        tbody.innerHTML = '';
            subscriptions.forEach(subscription => {
        const row = document.createElement('tr');
        row.innerHTML = `
                <td>${subscription.Tenant?.Name || 'N/A'}</td>
                <td>${subscription.Plan?.Name || 'N/A'}</td>
                <td>${subscription.Status || 'N/A'}</td>
                <td>$${subscription.Plan?.Price || '0'}/${subscription.Plan?.BillingInterval || 'month'}</td>
                <td>${subscription.CurrentPeriodEnd ? new Date(subscription.CurrentPeriodEnd).toLocaleDateString() : 'N/A'}</td>
                <td>
                    <button class="btn-secondary" onclick="window.saasAdminInstance.editSubscription('${subscription.ID}')">Edit</button>
                    <button class="btn-danger" onclick="window.saasAdminInstance.cancelSubscription('${subscription.ID}')">Cancel</button>
                </td>
        `;
                tbody.appendChild(row);
            });
    }

    async loadAnalyticsData() {
        console.log('Loading analytics data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/metrics');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Analytics data received:', data);
            this.renderAnalyticsCharts(data);
        } catch (error) {
            console.error('Error loading analytics:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('analytics-loading');
            const content = document.getElementById('analytics-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden analytics loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed analytics content');
            }
        }
    }

    renderAnalyticsCharts(data) {
        console.log('Rendering analytics charts with data:', data);
        
        const content = document.getElementById('analytics-content');
        if (!content) {
            console.error('Analytics content element not found');
            this.showError('Unable to render analytics - analytics section not found');
            return;
        }
        content.innerHTML = `
            <div class="analytics-dashboard">
                <div class="dashboard-header">
                    <h2>Analytics Dashboard</h2>
                    <div class="analytics-controls">
                        <select id="analytics-period" class="form-control">
                            <option value="7d">Last 7 days</option>
                            <option value="30d" selected>Last 30 days</option>
                            <option value="90d">Last 90 days</option>
                            <option value="1y">Last year</option>
                        </select>
                        <button class="btn-primary" onclick="window.saasAdminInstance.refreshAnalytics()">
                            <i class="fas fa-sync-alt"></i> Refresh
                        </button>
                    </div>
                </div>
                
                <!-- Key Metrics Cards -->
                <div class="metrics-grid">
                    <div class="metric-card">
                        <div class="metric-icon"><i class="fas fa-users"></i></div>
                        <div class="metric-content">
                            <h3>Total Users</h3>
                            <div class="metric-value">1,247</div>
                            <div class="metric-trend positive">+12% from last month</div>
            </div>
            </div>
                    
                    <div class="metric-card">
                        <div class="metric-icon"><i class="fas fa-dollar-sign"></i></div>
                        <div class="metric-content">
                            <h3>Monthly Revenue</h3>
                            <div class="metric-value">$24,580</div>
                            <div class="metric-trend positive">+8.3% from last month</div>
                        </div>
                    </div>
                    
                    <div class="metric-card">
                        <div class="metric-icon"><i class="fas fa-chart-line"></i></div>
                        <div class="metric-content">
                            <h3>Conversion Rate</h3>
                            <div class="metric-value">3.4%</div>
                            <div class="metric-trend negative">-0.2% from last month</div>
                        </div>
                    </div>
                    
                    <div class="metric-card">
                        <div class="metric-icon"><i class="fas fa-heart"></i></div>
                        <div class="metric-content">
                            <h3>Churn Rate</h3>
                            <div class="metric-value">2.1%</div>
                            <div class="metric-trend positive">-0.5% from last month</div>
                        </div>
                    </div>
                </div>
                
                <!-- Charts Section -->
                <div class="charts-grid">
                    <div class="chart-container">
                        <h3>User Growth</h3>
                        <canvas id="userGrowthChart" width="400" height="200"></canvas>
                    </div>
                    
                    <div class="chart-container">
                        <h3>Revenue Trends</h3>
                        <canvas id="revenueChart" width="400" height="200"></canvas>
                    </div>
                    
                    <div class="chart-container">
                        <h3>Plan Distribution</h3>
                        <canvas id="planDistributionChart" width="400" height="200"></canvas>
                    </div>
                    
                    <div class="chart-container">
                        <h3>System Performance</h3>
                        <canvas id="performanceChart" width="400" height="200"></canvas>
                    </div>
                </div>
                
                <!-- Data Tables -->
                <div class="tables-section">
                    <div class="table-container">
                        <h3>Top Performing Plans</h3>
                        <table class="data-table">
                            <thead>
                                <tr>
                                    <th>Plan Name</th>
                                    <th>Subscribers</th>
                                    <th>Revenue</th>
                                    <th>Growth</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>Professional</td>
                                    <td>456</td>
                                    <td>$13,680</td>
                                    <td class="positive">+15%</td>
                                </tr>
                                <tr>
                                    <td>Enterprise</td>
                                    <td>89</td>
                                    <td>$8,900</td>
                                    <td class="positive">+22%</td>
                                </tr>
                                <tr>
                                    <td>Starter</td>
                                    <td>702</td>
                                    <td>$2,000</td>
                                    <td class="positive">+8%</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    
                    <div class="table-container">
                        <h3>Recent Activity</h3>
                        <table class="data-table">
                            <thead>
                                <tr>
                                    <th>Event</th>
                                    <th>User</th>
                                    <th>Time</th>
                                    <th>Impact</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>New Subscription</td>
                                    <td>john@example.com</td>
                                    <td>2 mins ago</td>
                                    <td class="positive">+$99</td>
                                </tr>
                                <tr>
                                    <td>Plan Upgrade</td>
                                    <td>sarah@company.com</td>
                                    <td>15 mins ago</td>
                                    <td class="positive">+$50</td>
                                </tr>
                                <tr>
                                    <td>Subscription Cancelled</td>
                                    <td>mike@startup.io</td>
                                    <td>1 hour ago</td>
                                    <td class="negative">-$29</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        `;
        
        // Initialize charts
        this.initializeAnalyticsCharts();
    }
    
    initializeAnalyticsCharts() {
        // Destroy existing charts if they exist
        if (window.analyticsCharts) {
            window.analyticsCharts.forEach(chart => {
                if (chart) {
                    chart.destroy();
                }
            });
        }
        window.analyticsCharts = [];
        
        // User Growth Chart
        const userGrowthCtx = document.getElementById('userGrowthChart').getContext('2d');
        const userGrowthChart = new Chart(userGrowthCtx, {
            type: 'line',
            data: {
                labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
                datasets: [{
                    label: 'New Users',
                    data: [45, 62, 78, 91],
                    borderColor: '#007bff',
                    backgroundColor: 'rgba(0, 123, 255, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
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
        window.analyticsCharts.push(userGrowthChart);
        
        // Revenue Chart
        const revenueCtx = document.getElementById('revenueChart').getContext('2d');
        const revenueChart = new Chart(revenueCtx, {
            type: 'bar',
            data: {
                labels: ['Week 1', 'Week 2', 'Week 3', 'Week 4'],
                datasets: [{
                    label: 'Revenue',
                    data: [5200, 6800, 7200, 5380],
                    backgroundColor: '#28a745'
                }]
            },
            options: {
                responsive: true,
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
        window.analyticsCharts.push(revenueChart);
        
        // Plan Distribution Chart
        const planCtx = document.getElementById('planDistributionChart').getContext('2d');
        const planChart = new Chart(planCtx, {
            type: 'doughnut',
            data: {
                labels: ['Starter', 'Professional', 'Enterprise'],
                datasets: [{
                    data: [702, 456, 89],
                    backgroundColor: ['#ffc107', '#007bff', '#6f42c1']
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
        window.analyticsCharts.push(planChart);
        
        // Performance Chart
        const performanceCtx = document.getElementById('performanceChart').getContext('2d');
        const performanceChart = new Chart(performanceCtx, {
            type: 'line',
            data: {
                labels: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
                datasets: [{
                    label: 'Response Time (ms)',
                    data: [45, 52, 48, 61, 55, 47],
                    borderColor: '#dc3545',
                    backgroundColor: 'rgba(220, 53, 69, 0.1)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
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
        window.analyticsCharts.push(performanceChart);
    }

    async loadPlansData() {
        console.log('Loading plans data...');
        try {
            // Simulate plans data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Plans data loaded');
        } catch (error) {
            console.error('Error loading plans data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('plans-loading');
            const content = document.getElementById('plans-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden plans loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed plans content');
            }
        }
    }

    async loadBillingData() {
        console.log('Loading billing data...');
        try {
            // Simulate billing data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Billing data loaded');
        } catch (error) {
            console.error('Error loading billing data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('billing-loading');
            const content = document.getElementById('billing-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden billing loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed billing content');
            }
        }
    }

    async loadFeatureFlagsData() {
        console.log('Loading feature flags data...');
        try {
            const response = await fetch('/api/v1/admin/saas/admin/feature-availability');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            console.log('Feature flags data received:', data);
            this.renderFeatureFlagsList(data.feature_availability || []);
        } catch (error) {
            console.error('Error loading feature flags:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('feature-flags-loading');
            const content = document.getElementById('feature-flags-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden feature flags loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed feature flags content');
            }
        }
    }

    renderFeatureFlagsList(features) {
        console.log('Rendering feature flags list with:', features.length, 'features');
        const container = document.getElementById('feature-flags-content');
        if (!container) {
            console.error('feature-flags-content not found');
            return;
        }

        if (features.length === 0) {
            container.innerHTML = '<p>No feature flags available.</p>';
            return;
        }

        const featuresHtml = features.map(flag => `
            <div class="feature-flag-card">
                <div class="feature-flag-header">
                    <h3>${this.escapeHtml(flag.Feature)}</h3>
                    <div class="feature-flag-toggle">
                        <label class="toggle-switch">
                            <input type="checkbox" ${flag.IsAvailable ? 'checked' : ''} 
                                   onchange="window.saasAdminInstance.toggleFeatureAvailability('${this.escapeHtml(flag.Feature)}', this.checked)">
                            <span class="toggle-slider"></span>
            </label>
                </div>
                </div>
                <div class="feature-flag-description">
                    <p>${flag.Description || 'No description provided'}</p>
            </div>
                <div class="feature-flag-actions">
                    <button class="btn-secondary" onclick="window.saasAdminInstance.editFeatureFlag('${this.escapeHtml(flag.Feature)}', '${this.escapeHtml(flag.Description || '')}', ${flag.IsAvailable})">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <button class="btn-danger" onclick="window.saasAdminInstance.deleteFeatureFlag('${this.escapeHtml(flag.Feature)}')">
                        <i class="fas fa-trash"></i> Delete
                    </button>
                </div>
                </div>
        `).join('');

        container.innerHTML = featuresHtml;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    async loadSettingsData() {
        console.log('Loading settings data...');
        const loadingSpinner = document.getElementById('settings-loading');
        const content = document.getElementById('settings-content');
        
        try {
            // Simulate API call for settings data
            await new Promise(resolve => setTimeout(resolve, 500));
            
            // Render settings content
            this.renderSettingsContent();
            console.log('Settings data loaded');
        } catch (error) {
            console.error('Error loading settings data:', error);
            this.showError('Failed to load settings: ' + error.message);
        } finally {
            // Hide loading spinner and show content
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden settings loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed settings content');
            }
        }
    }

    renderSettingsContent() {
        const generalSettings = document.getElementById('generalSettings');
        const billingSettings = document.getElementById('billingSettings');
        
        if (generalSettings) {
            generalSettings.innerHTML = `
                <div class="form-group">
                    <label for="siteName">Site Name</label>
                    <input type="text" id="siteName" value="StatusPage Pro" class="form-control">
                </div>
                <div class="form-group">
                    <label for="siteDescription">Site Description</label>
                    <textarea id="siteDescription" class="form-control" rows="3">Professional status page monitoring for your services</textarea>
                </div>
                <div class="form-group">
                    <label for="contactEmail">Contact Email</label>
                    <input type="email" id="contactEmail" value="admin@statuspage.com" class="form-control">
                </div>
                <div class="form-group">
                    <label for="timezone">Timezone</label>
                    <select id="timezone" class="form-control">
                        <option value="UTC">UTC</option>
                        <option value="America/New_York">Eastern Time</option>
                        <option value="America/Chicago">Central Time</option>
                        <option value="America/Denver">Mountain Time</option>
                        <option value="America/Los_Angeles">Pacific Time</option>
                    </select>
                </div>
                <div class="form-group">
                    <label for="maintenanceMode">
                        <input type="checkbox" id="maintenanceMode"> Enable Maintenance Mode
                    </label>
                </div>
                <div class="form-actions">
                    <button class="btn-primary" onclick="window.saasAdminInstance.saveGeneralSettings()">Save Settings</button>
                </div>
            `;
        }
        
        if (billingSettings) {
            billingSettings.innerHTML = `
                <div class="form-group">
                    <label for="billingEmail">Billing Email</label>
                    <input type="email" id="billingEmail" value="billing@statuspage.com" class="form-control">
                </div>
                <div class="form-group">
                    <label for="currency">Currency</label>
                    <select id="currency" class="form-control">
                        <option value="USD">USD - US Dollar</option>
                        <option value="EUR">EUR - Euro</option>
                        <option value="GBP">GBP - British Pound</option>
                        <option value="CAD">CAD - Canadian Dollar</option>
                    </select>
            </div>
                <div class="form-group">
                    <label for="taxRate">Tax Rate (%)</label>
                    <input type="number" id="taxRate" value="8.5" step="0.1" class="form-control">
                </div>
                <div class="form-group">
                    <label for="invoicePrefix">Invoice Prefix</label>
                    <input type="text" id="invoicePrefix" value="INV-" class="form-control">
                </div>
                <div class="form-group">
                    <label for="autoRenewal">
                        <input type="checkbox" id="autoRenewal" checked> Enable Auto-Renewal
                    </label>
                </div>
                <div class="form-actions">
                    <button class="btn-primary" onclick="window.saasAdminInstance.saveBillingSettings()">Save Billing Settings</button>
                </div>
            `;
        }
    }

    saveGeneralSettings() {
        console.log('saveGeneralSettings called');
        const siteName = document.getElementById('siteName')?.value;
        const siteDescription = document.getElementById('siteDescription')?.value;
        const contactEmail = document.getElementById('contactEmail')?.value;
        const timezone = document.getElementById('timezone')?.value;
        const maintenanceMode = document.getElementById('maintenanceMode')?.checked;
        
        console.log('Saving general settings:', {
            siteName, siteDescription, contactEmail, timezone, maintenanceMode
        });
        
        // Simulate save
        alert('General settings saved successfully!');
    }

    saveBillingSettings() {
        console.log('saveBillingSettings called');
        const billingEmail = document.getElementById('billingEmail')?.value;
        const currency = document.getElementById('currency')?.value;
        const taxRate = document.getElementById('taxRate')?.value;
        const invoicePrefix = document.getElementById('invoicePrefix')?.value;
        const autoRenewal = document.getElementById('autoRenewal')?.checked;
        
        console.log('Saving billing settings:', {
            billingEmail, currency, taxRate, invoicePrefix, autoRenewal
        });
        
        // Simulate save
        alert('Billing settings saved successfully!');
    }

    async loadUsersData() {
        console.log('Loading users data...');
        try {
            // Simulate users data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Users data loaded');
        } catch (error) {
            console.error('Error loading users data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('users-loading');
            const content = document.getElementById('users-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden users loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed users content');
            }
        }
    }

    async loadIntegrationsData() {
        console.log('Loading integrations data...');
        try {
            // Simulate integrations data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Integrations data loaded');
        } catch (error) {
            console.error('Error loading integrations data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('integrations-loading');
            const content = document.getElementById('integrations-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden integrations loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed integrations content');
            }
        }
    }

    async loadNotificationsData() {
        console.log('Loading notifications data...');
        try {
            // Simulate notifications data loading
            await new Promise(resolve => setTimeout(resolve, 500));
            console.log('Notifications data loaded');
        } catch (error) {
            console.error('Error loading notifications data:', error);
        } finally {
            // Hide loading spinner and show content
            const loadingSpinner = document.getElementById('notifications-loading');
            const content = document.getElementById('notifications-content');
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden notifications loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed notifications content');
            }
        }
    }

    // Feature flag methods
    async toggleFeatureAvailability(feature, isAvailable) {
        console.log('Toggling feature availability:', feature, isAvailable);
        try {
            const response = await fetch(`/api/v1/admin/saas/admin/feature-availability/${encodeURIComponent(feature)}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    is_available: isAvailable
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            console.log('Feature availability updated successfully');
            // Reload the feature flags to show updated state
            await this.loadFeatureFlagsData();
        } catch (error) {
            console.error('Error updating feature availability:', error);
            alert('Error updating feature availability: ' + error.message);
        }
    }

    showCreateFeatureFlagForm() {
        console.log('Showing create feature flag form');
        const modal = document.getElementById('featureFlagModal');
        if (modal) {
            modal.style.display = 'block';
            // Reset form
            document.getElementById('featureFlagForm').reset();
            document.getElementById('featureFlagModalTitle').textContent = 'Add Feature Flag';
            // Clear any edit state
            delete modal.dataset.originalFeature;
            // Enable feature name field for new features
            document.getElementById('featureName').disabled = false;
        } else {
            console.error('Feature flag modal not found');
        }
    }

    closeModal(modalId) {
        console.log('Closing modal:', modalId);
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'none';
            // Reset modal state when closing
            if (modalId === 'tenantModal') {
                modal.dataset.mode = 'create';
                modal.dataset.tenantId = '';
                // Reset form
                const form = document.getElementById('tenantForm');
                if (form) {
                    form.reset();
                }
                // Reset modal title and button
                const modalTitle = modal.querySelector('h2');
                if (modalTitle) {
                    modalTitle.textContent = 'Create Tenant';
                }
                const submitButton = modal.querySelector('button[onclick*="submitTenant"]');
                if (submitButton) {
                    submitButton.textContent = 'Create Tenant';
                }
            }
        }
    }

    async createFeatureFlag(event) {
        event.preventDefault();
        console.log('Creating/updating feature flag...');
        
        const formData = new FormData(event.target);
        const modal = document.getElementById('featureFlagModal');
        const isEdit = modal.dataset.originalFeature;
        
        const featureData = {
            feature: formData.get('feature'),
            description: formData.get('description'),
            is_available: formData.get('is_available') === 'on'
        };

        try {
            let response;
            if (isEdit) {
                // Check if feature name has changed
                if (featureData.feature !== isEdit) {
                    // Feature name changed - delete old and create new
                    console.log('Feature name changed from', isEdit, 'to', featureData.feature);
                    
                    // Delete the old feature
                    const deleteResponse = await fetch(`/api/v1/admin/saas/admin/feature-availability/${encodeURIComponent(isEdit)}`, {
                        method: 'DELETE',
                        headers: {
                            'Content-Type': 'application/json',
                        }
                    });
                    
                    if (!deleteResponse.ok) {
                        throw new Error(`Failed to delete old feature: HTTP ${deleteResponse.status}`);
                    }
                    
                    // Create new feature with new name
                    response = await fetch('/api/v1/admin/saas/admin/feature-availability', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify(featureData)
                    });
                } else {
                    // Feature name unchanged - just update
                    const updateData = {
                        is_available: featureData.is_available,
                        description: featureData.description
                    };
                    console.log('Updating feature:', isEdit, 'with data:', updateData);
                    response = await fetch(`/api/v1/admin/saas/admin/feature-availability/${encodeURIComponent(isEdit)}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify(updateData)
                    });
                }
            } else {
                // Create new feature flag
                response = await fetch('/api/v1/admin/saas/admin/feature-availability', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(featureData)
                });
            }

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            console.log(`Feature flag ${isEdit ? 'updated' : 'created'} successfully`);
            this.closeModal('featureFlagModal');
            // Clear the original feature name
            if (modal) {
                delete modal.dataset.originalFeature;
            }
            // Reload the feature flags to show the updated list
            await this.loadFeatureFlagsData();
        } catch (error) {
            console.error(`Error ${isEdit ? 'updating' : 'creating'} feature flag:`, error);
            alert(`Error ${isEdit ? 'updating' : 'creating'} feature flag: ` + error.message);
        }
    }

    editFeatureFlag(feature, description, isAvailable) {
        console.log('Editing feature flag:', feature, description, isAvailable);
        const modal = document.getElementById('featureFlagModal');
        if (modal) {
            modal.style.display = 'block';
            
            // Populate form with existing data
            document.getElementById('featureName').value = feature;
            document.getElementById('featureDescription').value = description || '';
            document.getElementById('featureAvailable').checked = isAvailable;
            document.getElementById('featureFlagModalTitle').textContent = 'Edit Feature Flag';
            
            // Store original feature name for update
            modal.dataset.originalFeature = feature;
            
            // Enable feature name field for editing
            document.getElementById('featureName').disabled = false;
        } else {
            console.error('Feature flag modal not found');
        }
    }

    async deleteFeatureFlag(feature) {
        console.log('Deleting feature flag:', feature);
        if (confirm(`Are you sure you want to delete the feature flag "${feature}"?`)) {
            try {
                const response = await fetch(`/api/v1/admin/saas/admin/feature-availability/${encodeURIComponent(feature)}`, {
                    method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    }
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
                console.log('Feature flag deleted successfully');
                // Reload the feature flags to show updated list
                await this.loadFeatureFlagsData();
        } catch (error) {
                console.error('Error deleting feature flag:', error);
                alert('Error deleting feature flag: ' + error.message);
            }
        }
    }

    // Tenant methods
    editTenant(tenantId) {
        console.log('Editing tenant:', tenantId);
        // Edit tenant logic here
    }

    // Tenant Features methods
    async loadTenantFeaturesData() {
        console.log('Loading tenant features data...');
        console.log('Method called successfully');
        const loadingSpinner = document.getElementById('tenant-features-loading');
        const content = document.getElementById('tenant-features-content');
        console.log('Loading spinner found:', !!loadingSpinner);
        console.log('Content container found:', !!content);
        
        if (loadingSpinner) {
            loadingSpinner.style.display = 'block';
            console.log('Showed tenant features loading spinner');
        }
        if (content) {
            content.style.display = 'none';
            content.style.visibility = 'hidden';
            console.log('Hid tenant features content');
        }

        try {
            // Load tenants first
            const tenantsResponse = await fetch('/api/v1/admin/saas/admin/tenants');
            if (!tenantsResponse.ok) {
                throw new Error(`HTTP ${tenantsResponse.status}: ${tenantsResponse.statusText}`);
            }
            const tenantsData = await tenantsResponse.json();
            console.log('Tenants data received:', tenantsData);

            // Load available features
            const featuresResponse = await fetch('/api/v1/admin/saas/admin/feature-availability');
            if (!featuresResponse.ok) {
                throw new Error(`HTTP ${featuresResponse.status}: ${featuresResponse.statusText}`);
            }
            const featuresData = await featuresResponse.json();
            console.log('Available features data received:', featuresData);

            this.renderTenantFeaturesList(tenantsData.tenants, featuresData.feature_availability || []);

        } catch (error) {
            console.error('Error loading tenant features data:', error);
            this.showError('Failed to load tenant features data: ' + error.message);
        } finally {
            if (loadingSpinner) {
                loadingSpinner.style.display = 'none';
                console.log('Hidden tenant features loading spinner');
            }
            if (content) {
                content.style.display = 'block';
                content.style.visibility = 'visible';
                console.log('Showed tenant features content');
            }
        }
    }

    showError(message) {
        console.error('Error:', message);
        // For now, just log the error. In a real app, you'd show a user-friendly error message
        alert('Error: ' + message);
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    renderTenantFeaturesList(tenants, availableFeatures) {
        console.log('Rendering tenant features list with:', tenants.length, 'tenants and', availableFeatures.length, 'features');
        const container = document.getElementById('tenant-features-content');
        if (!container) {
            console.error('tenant-features-content not found');
            return;
        }

        if (tenants.length === 0) {
            container.innerHTML = '<p>No tenants found.</p>';
            return;
        }

        const tenantsHtml = tenants.map(tenant => `
            <div class="tenant-feature-card">
                <div class="tenant-feature-header">
                    <h3>${this.escapeHtml(tenant.Name)}</h3>
                    <span class="tenant-slug">${this.escapeHtml(tenant.Slug)}</span>
                </div>
                <div class="tenant-feature-list">
                    ${availableFeatures.map(feature => `
                        <div class="tenant-feature-item">
                            <div class="feature-info">
                                <span class="feature-name">${this.escapeHtml(feature.Feature)}</span>
                                <span class="feature-description">${this.escapeHtml(feature.Description || 'No description')}</span>
            </div>
                            <div class="feature-toggle">
                                <label class="toggle-switch">
                                    <input type="checkbox" 
                                           data-tenant-id="${tenant.ID}" 
                                           data-feature="${this.escapeHtml(feature.Feature)}"
                                           onchange="window.saasAdminInstance.toggleTenantFeature(${tenant.ID}, '${this.escapeHtml(feature.Feature)}', this.checked)">
                                    <span class="toggle-slider"></span>
                                </label>
            </div>
            </div>
                    `).join('')}
                </div>
            </div>
        `).join('');

        container.innerHTML = tenantsHtml;
    }

    async toggleTenantFeature(tenantId, feature, isEnabled) {
        console.log('Toggling tenant feature:', tenantId, feature, isEnabled);
        try {
            const response = await fetch(`/api/v1/admin/saas/admin/tenants/${tenantId}/feature-flags`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    feature: feature,
                    is_enabled: isEnabled
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            console.log('Tenant feature toggled successfully');
        } catch (error) {
            console.error('Error toggling tenant feature:', error);
            alert('Error updating tenant feature: ' + error.message);
            // Revert the toggle
            const checkbox = document.querySelector(`input[data-tenant-id="${tenantId}"][data-feature="${feature}"]`);
            if (checkbox) {
                checkbox.checked = !isEnabled;
            }
        }
    }

    showAddFeatureForm() {
        console.log('Show add feature form');
        
        const content = document.getElementById('tenant-features-content');
        if (!content) {
            console.error('Tenant features content element not found');
            this.showError('Unable to show add feature form - tenant features section not found');
            return;
        }
        
        content.innerHTML = `
            <div class="add-feature-form">
                <h3>Add New Feature</h3>
                <form id="addFeatureForm" class="form-container">
                    <div class="form-group">
                        <label for="featureName">Feature Name</label>
                        <input type="text" id="featureName" name="feature" required class="form-control" placeholder="e.g., advanced_analytics">
                    </div>
                    <div class="form-group">
                        <label for="featureDescription">Description</label>
                        <textarea id="featureDescription" name="description" class="form-control" rows="3" placeholder="Describe what this feature does"></textarea>
                    </div>
                        <div class="form-group">
                        <label for="featureCategory">Category</label>
                        <select id="featureCategory" name="category" class="form-control">
                            <option value="analytics">Analytics</option>
                            <option value="billing">Billing</option>
                            <option value="monitoring">Monitoring</option>
                            <option value="customization">Customization</option>
                            <option value="integration">Integration</option>
                            <option value="other">Other</option>
                            </select>
                        </div>
                        <div class="form-group">
                        <label for="featureDefaultEnabled">Default State</label>
                        <select id="featureDefaultEnabled" name="default_enabled" class="form-control">
                            <option value="false">Disabled by default</option>
                            <option value="true">Enabled by default</option>
                        </select>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn-primary">
                            <i class="fas fa-plus"></i> Add Feature
                        </button>
                        <button type="button" class="btn-secondary" onclick="window.saasAdminInstance.loadTenantFeaturesData()">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        `;
        
        // Add form handler
        document.getElementById('addFeatureForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.submitAddFeature();
        });
    }
    
    async submitAddFeature() {
        const form = document.getElementById('addFeatureForm');
        const formData = new FormData(form);
        
        const featureData = {
            feature: formData.get('feature'),
            description: formData.get('description'),
            category: formData.get('category'),
            default_enabled: formData.get('default_enabled') === 'true'
        };
        
        console.log('Adding new feature:', featureData);
        
        // Simulate API call
        console.log('Simulating feature addition...');
        
        setTimeout(() => {
            alert('Feature added successfully!');
            this.loadTenantFeaturesData(); // Refresh the list
        }, 1000);
    }

    async deleteTenant(tenantId) {
        console.log('Deleting tenant:', tenantId);
        
        // Confirm deletion
        if (!confirm('Are you sure you want to delete this tenant? This action cannot be undone.')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/v1/admin/saas/admin/tenants/${tenantId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.showSuccess('Tenant deleted successfully!');
                this.loadTenantsData(); // Refresh the tenants list
            } else {
                const error = await response.json();
                this.showError('Failed to delete tenant: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error deleting tenant:', error);
            this.showError('Error deleting tenant: ' + error.message);
        }
    }

    // Subscription methods
    async editSubscription(subscriptionId) {
        console.log('Editing subscription:', subscriptionId);
        try {
            // Fetch subscription details
            const response = await fetch(`/api/v1/admin/saas/admin/subscriptions/${subscriptionId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const subscription = data.subscription || data;
            
            // Show edit form with populated data
            this.showEditSubscriptionForm(subscription);
        } catch (error) {
            console.error('Error fetching subscription:', error);
            this.showError('Failed to load subscription details: ' + error.message);
        }
    }

    async cancelSubscription(subscriptionId) {
        console.log('Cancelling subscription:', subscriptionId);
        
        if (!confirm('Are you sure you want to cancel this subscription? This action cannot be undone.')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/admin/saas/admin/subscriptions/${subscriptionId}/cancel`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            if (response.ok) {
                this.showSuccess('Subscription cancelled successfully.');
                this.loadSectionData('subscriptions'); // Refresh the subscriptions list
            } else {
                const error = await response.json();
                this.showError('Failed to cancel subscription: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error cancelling subscription:', error);
            this.showError('Error cancelling subscription: ' + error.message);
        }
    }

    showEditSubscriptionForm(subscription) {
        console.log('Show edit subscription form for:', subscription);
        
        const content = document.getElementById('subscriptions-content');
        if (!content) {
            console.error('Subscriptions content element not found');
            this.showError('Unable to show subscription form - subscriptions section not found');
            return;
        }
        
        // Use the same form as create, but populate with existing data
        this.showCreateSubscriptionForm();
        
        // Update form title and button text
        const formTitle = content.querySelector('h2');
        if (formTitle) {
            formTitle.textContent = 'Edit Subscription';
        }
        
        const submitButton = content.querySelector('button[type="submit"]');
        if (submitButton) {
            submitButton.innerHTML = '<i class="fas fa-save"></i> Update Subscription';
        }
        
        // Populate form fields with existing data
        if (subscription.Tenant) {
            const tenantSelect = document.getElementById('tenantSelect');
            if (tenantSelect) {
                tenantSelect.value = subscription.Tenant.ID;
            }
        }
        
        if (subscription.Plan) {
            const planSelect = document.getElementById('subscriptionPlan');
            if (planSelect) {
                planSelect.value = subscription.Plan.Slug;
            }
        }
        
        // Set other fields
        const startDate = document.getElementById('startDate');
        if (startDate && subscription.CurrentPeriodStart) {
            startDate.value = new Date(subscription.CurrentPeriodStart).toISOString().split('T')[0];
        }
        
        // Update form to handle edit mode
        const form = document.getElementById('subscriptionForm');
        if (form) {
            form.dataset.mode = 'edit';
            form.dataset.subscriptionId = subscription.ID;
        }
    }

    // Tenant Management Methods
    showCreateTenantForm() {
        console.log('Show create tenant form');
        const modal = document.getElementById('tenantModal');
        if (modal) {
            modal.style.display = 'block';
            // Clear form
            document.getElementById('tenantForm').reset();
            // Set modal title to Create
            const modalTitle = modal.querySelector('h2');
            if (modalTitle) {
                modalTitle.textContent = 'Create Tenant';
            }
            // Set form mode to create
            modal.dataset.mode = 'create';
            modal.dataset.tenantId = '';
            // Update submit button text
            const submitButton = modal.querySelector('button[onclick*="submitTenant"]');
            if (submitButton) {
                submitButton.textContent = 'Create Tenant';
            }
        }
    }

    showEditTenantForm(tenantData) {
        console.log('Show edit tenant form with data:', tenantData);
        console.log('Tenant ID from data:', tenantData.ID || tenantData.id);
        const modal = document.getElementById('tenantModal');
        if (modal) {
            modal.style.display = 'block';
            
            // Set modal title to Edit
            const modalTitle = modal.querySelector('h2');
            if (modalTitle) {
                modalTitle.textContent = 'Edit Tenant';
            }
            
            // Set form mode to edit
            modal.dataset.mode = 'edit';
            const tenantId = tenantData.ID || tenantData.id;
            modal.dataset.tenantId = tenantId;
            console.log('Set modal tenantId to:', tenantId);
            
            // Update submit button text
            const submitButton = modal.querySelector('button[onclick*="submitTenant"]');
            if (submitButton) {
                submitButton.textContent = 'Update Tenant';
            }
            
            // Populate form with existing data
            const form = document.getElementById('tenantForm');
            if (form) {
                form.reset();
                
                // Map the tenant data to form fields
                const fieldMappings = {
                    'tenantName': tenantData.Name || tenantData.name,
                    'tenantSlug': tenantData.Slug || tenantData.slug,
                    'tenantEmail': tenantData.ContactEmail || tenantData.contact_email,
                    'tenantBillingEmail': tenantData.BillingEmail || tenantData.billing_email,
                    'tenantSubdomain': tenantData.Subdomain || tenantData.subdomain,
                    'tenantDomain': tenantData.Domain || tenantData.domain,
                    'tenantPlan': tenantData.Plan || tenantData.plan
                };
                console.log('Field mappings:', fieldMappings);
                
                // Populate each field
                Object.entries(fieldMappings).forEach(([fieldId, value]) => {
                    const field = document.getElementById(fieldId);
                    console.log(`Setting field ${fieldId} to value:`, value);
                    if (field && value) {
                        field.value = value;
                        console.log(`Successfully set ${fieldId} = ${value}`);
                    } else {
                        console.log(`Field ${fieldId} not found or value is empty:`, value);
                    }
                });
            }
        }
    }

    async submitTenant() {
        const form = document.getElementById('tenantForm');
        const modal = document.getElementById('tenantModal');
        const formData = new FormData(form);
        
        const tenantData = {
            name: formData.get('name'),
            slug: formData.get('slug'),
            contact_email: formData.get('contact_email'),
            billing_email: formData.get('billing_email') || formData.get('contact_email'),
            subdomain: formData.get('subdomain'),
            domain: formData.get('domain'),
            plan: formData.get('plan') || 'free',
            status: 'active'
        };

        // Determine if this is create or edit mode
        const mode = modal.dataset.mode || 'create';
        const tenantId = modal.dataset.tenantId;
        console.log('Submit tenant - mode:', mode, 'tenantId:', tenantId);

        try {
            let response;
            if (mode === 'edit' && tenantId) {
                // Update existing tenant
                response = await fetch(`/api/v1/admin/saas/admin/tenants/${tenantId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(tenantData)
                });
            } else {
                // Create new tenant
                response = await fetch('/api/v1/admin/saas/admin/tenants', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(tenantData)
            });
            }

            if (response.ok) {
                const result = await response.json();
                const action = mode === 'edit' ? 'updated' : 'created';
                this.showSuccess(`Tenant ${action} successfully!`);
                this.closeModal('tenantModal');
                this.loadTenantsData(); // Refresh the tenants list
            } else {
                const error = await response.json();
                const action = mode === 'edit' ? 'updating' : 'creating';
                this.showError(`Failed to ${action} tenant: ` + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error(`Error ${mode === 'edit' ? 'updating' : 'creating'} tenant:`, error);
            const action = mode === 'edit' ? 'updating' : 'creating';
            this.showError(`Error ${action} tenant: ` + error.message);
        }
    }


    displayTenants(tenants) {
        const tbody = document.getElementById('tenantsTableBody');
        if (!tbody) return;

        if (tenants.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="text-center">No tenants found</td></tr>';
            return;
        }

        tbody.innerHTML = tenants.map(tenant => `
            <tr>
                <td>${tenant.name || 'N/A'}</td>
                <td>${tenant.slug || 'N/A'}</td>
                <td><span class="badge badge-${tenant.plan || 'free'}">${tenant.plan || 'free'}</span></td>
                <td><span class="badge badge-${tenant.status || 'active'}">${tenant.status || 'active'}</span></td>
                <td>${tenant.created_at ? new Date(tenant.created_at).toLocaleDateString() : 'N/A'}</td>
                <td>
                    <button class="btn-sm btn-primary" onclick="window.saasAdminInstance.viewTenant(${tenant.id})">
                        <i class="fas fa-eye"></i> View
                    </button>
                    <button class="btn-sm btn-secondary" onclick="window.saasAdminInstance.editTenant(${tenant.id})">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                </td>
            </tr>
        `).join('');
    }

    viewTenant(tenantId) {
        console.log('View tenant:', tenantId);
        // TODO: Implement tenant view functionality
        this.showInfo('Tenant view functionality coming soon!');
    }

    async editTenant(tenantId) {
        console.log('Edit tenant:', tenantId);
        try {
            // Fetch tenant data
            const response = await fetch(`/api/v1/admin/saas/admin/tenants/${tenantId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const responseData = await response.json();
            console.log('API response received:', responseData);
            
            // Extract tenant data from the response
            const tenantData = responseData.tenant || responseData;
            console.log('Tenant data extracted:', tenantData);
            
            // Show the modal and populate with existing data
            this.showEditTenantForm(tenantData);
        } catch (error) {
            console.error('Error loading tenant data:', error);
            this.showError('Failed to load tenant data: ' + error.message);
        }
    }

    async createTenantForPlan(planType) {
        const planNames = {
            'free': 'Free',
            'pro': 'Pro',
            'enterprise': 'Enterprise'
        };

        const tenantData = {
            name: `${planNames[planType]} Plan Customer`,
            slug: `${planType}-customer-${Date.now()}`,
            contact_email: 'customer@example.com',
            billing_email: 'customer@example.com',
            subdomain: `${planType}-customer-${Date.now()}`,
            plan: planType,
            status: 'active'
        };

        try {
            const response = await fetch('/api/v1/admin/saas/admin/tenants', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(tenantData)
            });

            if (response.ok) {
                const result = await response.json();
                this.showSuccess(`${planNames[planType]} plan tenant created successfully!`);
                this.loadTenantsData(); // Refresh tenants list
                
                // Navigate to tenants section to show the new tenant
                this.navigateToSection('tenants');
            } else {
                const error = await response.json();
                this.showError('Failed to create tenant: ' + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Error creating tenant:', error);
            this.showError('Failed to create tenant: ' + error.message);
        }
    }

    showCreateSubscriptionForm() {
        console.log('Show create subscription form');
        
        const content = document.getElementById('subscriptions-content');
        if (!content) {
            console.error('Subscriptions content element not found');
            this.showError('Unable to show subscription form - subscriptions section not found');
            return;
        }
        content.innerHTML = `
            <div class="create-subscription-form">
                <h2>Create New Subscription</h2>
                <form id="subscriptionForm" class="form-container">
                    <div class="form-section">
                        <h3>Customer Information</h3>
                        <div class="form-row">
                            <div class="form-group">
                                <label for="customerEmail">Customer Email</label>
                                <input type="email" id="customerEmail" name="email" required class="form-control" placeholder="customer@example.com">
                            </div>
                            
                            <div class="form-group">
                                <label for="customerName">Customer Name</label>
                                <input type="text" id="customerName" name="name" required class="form-control" placeholder="John Doe">
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label for="companyName">Company Name (Optional)</label>
                            <input type="text" id="companyName" name="company" class="form-control" placeholder="Acme Corporation">
                        </div>
                    </div>
                    
                    <div class="form-section">
                        <h3>Subscription Details</h3>
                        <div class="form-group">
                            <label for="tenantSelect">Tenant</label>
                            <select id="tenantSelect" name="tenantId" required class="form-control">
                                <option value="">Select a tenant</option>
                            </select>
                        </div>
                        
                        <div class="form-group">
                            <label for="subscriptionPlan">Plan</label>
                            <select id="subscriptionPlan" name="plan" required class="form-control">
                                <option value="">Select a plan</option>
                                <option value="free">Free - $0/month</option>
                                <option value="pro">Pro - $29/month</option>
                                <option value="enterprise">Enterprise - $99/month</option>
                            </select>
                        </div>
                        
                        <div class="form-group">
                            <label for="billingCycle">Billing Cycle</label>
                            <select id="billingCycle" name="cycle" required class="form-control">
                                <option value="monthly">Monthly</option>
                                <option value="yearly">Yearly (20% discount)</option>
                            </select>
                        </div>
                        
                        <div class="form-group">
                            <label for="startDate">Start Date</label>
                            <input type="date" id="startDate" name="startDate" required class="form-control">
                        </div>
                        
                        <div class="form-group">
                            <label for="trialDays">Trial Days (Optional)</label>
                            <input type="number" id="trialDays" name="trialDays" min="0" max="365" class="form-control" placeholder="14">
                        </div>
                    </div>
                    
                    <div class="form-section">
                        <h3>Payment Information</h3>
                        <div class="form-group">
                            <label for="paymentMethod">Payment Method</label>
                            <select id="paymentMethod" name="paymentMethod" required class="form-control">
                                <option value="">Select payment method</option>
                                <option value="credit_card">Credit Card</option>
                                <option value="paypal">PayPal</option>
                                <option value="bank_transfer">Bank Transfer</option>
                                <option value="invoice">Invoice (Enterprise only)</option>
                            </select>
                </div>
                        
                        <div id="creditCardFields" style="display: none;">
                    <div class="form-row">
                        <div class="form-group">
                                    <label for="cardNumber">Card Number</label>
                                    <input type="text" id="cardNumber" name="cardNumber" class="form-control" placeholder="1234 5678 9012 3456">
                        </div>
                                
                        <div class="form-group">
                                    <label for="expiryDate">Expiry Date</label>
                                    <input type="text" id="expiryDate" name="expiryDate" class="form-control" placeholder="MM/YY">
                        </div>
                                
                        <div class="form-group">
                                    <label for="cvv">CVV</label>
                                    <input type="text" id="cvv" name="cvv" class="form-control" placeholder="123">
                        </div>
                        </div>
                    </div>
                        </div>
                    
                    <div class="form-section">
                        <h3>Additional Settings</h3>
                        <div class="form-group">
                            <label>
                                <input type="checkbox" id="autoRenewal" name="autoRenewal" checked> Enable Auto-Renewal
                            </label>
                        </div>
                        
                        <div class="form-group">
                            <label>
                                <input type="checkbox" id="sendWelcomeEmail" name="sendWelcomeEmail" checked> Send Welcome Email
                            </label>
                        </div>
                        
                        <div class="form-group">
                            <label for="notes">Notes (Internal)</label>
                            <textarea id="notes" name="notes" class="form-control" rows="3" placeholder="Internal notes about this subscription"></textarea>
                        </div>
                    </div>
                    
                    <div class="subscription-summary">
                        <h3>Subscription Summary</h3>
                        <div class="summary-details">
                            <div class="summary-item">
                                <span class="label">Plan:</span>
                                <span class="value" id="summaryPlan">-</span>
                            </div>
                            <div class="summary-item">
                                <span class="label">Billing:</span>
                                <span class="value" id="summaryBilling">-</span>
                            </div>
                            <div class="summary-item">
                                <span class="label">Total:</span>
                                <span class="value" id="summaryTotal">$0.00</span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="form-actions">
                        <button type="submit" class="btn-primary">
                            <i class="fas fa-credit-card"></i> Create Subscription
                        </button>
                        <button type="button" class="btn-secondary" onclick="window.saasAdminInstance.loadSectionData('subscriptions')">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        `;

        // Set default start date to today
        document.getElementById('startDate').value = new Date().toISOString().split('T')[0];
        
        // Load tenants for the dropdown
        this.loadTenantsForSubscription();
        
        // Reset form mode to create
        const form = document.getElementById('subscriptionForm');
        if (form) {
            form.dataset.mode = 'create';
            delete form.dataset.subscriptionId;
        }
        
        // Add form handlers
        document.getElementById('subscriptionForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.submitSubscription();
        });
        
        // Show/hide payment fields based on method
        document.getElementById('paymentMethod').addEventListener('change', (e) => {
            const creditCardFields = document.getElementById('creditCardFields');
            if (e.target.value === 'credit_card') {
                creditCardFields.style.display = 'block';
            } else {
                creditCardFields.style.display = 'none';
            }
        });
        
        // Update summary on plan/billing change
        document.getElementById('subscriptionPlan').addEventListener('change', this.updateSubscriptionSummary.bind(this));
        document.getElementById('billingCycle').addEventListener('change', this.updateSubscriptionSummary.bind(this));
    }
    
    async loadTenantsForSubscription() {
        try {
            const response = await fetch('/api/v1/admin/saas/admin/tenants');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            const tenants = data.tenants || [];
            
            const tenantSelect = document.getElementById('tenantSelect');
            if (tenantSelect) {
                tenantSelect.innerHTML = '<option value="">Select a tenant</option>';
                tenants.forEach(tenant => {
                    const option = document.createElement('option');
                    option.value = tenant.ID;
                    option.textContent = `${tenant.Name} (${tenant.Slug})`;
                    tenantSelect.appendChild(option);
                });
            }
        } catch (error) {
            console.error('Error loading tenants for subscription:', error);
            this.showError('Failed to load tenants. Please try again.');
        }
    }
    
    updateSubscriptionSummary() {
        const plan = document.getElementById('subscriptionPlan').value;
        const cycle = document.getElementById('billingCycle').value;
        
        const planPrices = {
            free: { monthly: 0, yearly: 0 },
            pro: { monthly: 29, yearly: 290 },
            enterprise: { monthly: 99, yearly: 990 }
        };
        
        const planNames = {
            free: 'Free',
            pro: 'Pro',
            enterprise: 'Enterprise'
        };
        
        if (plan && cycle) {
            const price = planPrices[plan][cycle];
            const discount = cycle === 'yearly' ? ' (20% discount)' : '';
            
            document.getElementById('summaryPlan').textContent = planNames[plan];
            document.getElementById('summaryBilling').textContent = cycle.charAt(0).toUpperCase() + cycle.slice(1) + discount;
            document.getElementById('summaryTotal').textContent = `$${price.toFixed(2)}${cycle === 'yearly' ? '/year' : '/month'}`;
        }
    }
    
    async submitSubscription() {
        console.log('submitSubscription called');
        const form = document.getElementById('subscriptionForm');
        if (!form) {
            console.error('Subscription form not found');
            this.showError('Subscription form not found');
            return;
        }
        const formData = new FormData(form);
        
        const tenantId = formData.get('tenantId');
        if (!tenantId) {
            this.showError('Please select a tenant for the subscription.');
            return;
        }
        
        const subscriptionData = {
            customer: {
                email: formData.get('email'),
            name: formData.get('name'),
                company: formData.get('company')
            },
            plan: formData.get('plan'),
            billingCycle: formData.get('cycle'),
            startDate: formData.get('startDate'),
            trialDays: parseInt(formData.get('trialDays')) || 0,
            paymentMethod: formData.get('paymentMethod'),
            autoRenewal: formData.get('autoRenewal') === 'on',
            sendWelcomeEmail: formData.get('sendWelcomeEmail') === 'on',
            notes: formData.get('notes')
        };
        
        const mode = form.dataset.mode || 'create';
        const subscriptionId = form.dataset.subscriptionId;
        
        console.log(`${mode} subscription for tenant:`, tenantId, subscriptionData);
        
        try {
            let response;
            if (mode === 'edit' && subscriptionId) {
                // Update existing subscription
                response = await fetch(`/api/v1/admin/saas/admin/subscriptions/${subscriptionId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(subscriptionData)
                });
            } else {
                // Create new subscription
                response = await fetch(`/api/v1/admin/saas/admin/tenants/${tenantId}/subscription`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(subscriptionData)
                });
            }

            if (response.ok) {
                const result = await response.json();
                const message = mode === 'edit' ? 'Subscription updated successfully!' : 'Subscription created successfully! Welcome email sent to customer.';
                this.showSuccess(message);
                this.loadSectionData('subscriptions'); // Refresh the subscriptions list
            } else {
                const error = await response.json();
                const action = mode === 'edit' ? 'update' : 'create';
                this.showError(`Failed to ${action} subscription: ` + (error.error || 'Unknown error'));
            }
        } catch (error) {
            console.error(`Error ${mode}ing subscription:`, error);
            const action = mode === 'edit' ? 'updating' : 'creating';
            this.showError(`Error ${action} subscription: ` + error.message);
        }
    }

    showCreatePlanForm() {
        console.log('Show create plan form');
        
        const content = document.getElementById('plans-content');
        if (!content) {
            console.error('Plans content element not found');
            this.showError('Unable to show plan form - plans section not found');
            return;
        }
        content.innerHTML = `
            <div class="create-plan-form">
                <h2>Create New Plan</h2>
                <form id="planForm" class="form-container">
                    <div class="form-section">
                        <h3>Basic Information</h3>
                        <div class="form-group">
                            <label for="planName">Plan Name</label>
                            <input type="text" id="planName" name="name" required class="form-control" placeholder="e.g., Professional">
                        </div>
                        
                        <div class="form-group">
                            <label for="planDescription">Description</label>
                            <textarea id="planDescription" name="description" required class="form-control" rows="3" placeholder="Brief description of the plan"></textarea>
                        </div>
                        
                        <div class="form-group">
                            <label for="planType">Plan Type</label>
                            <select id="planType" name="type" required class="form-control">
                                <option value="">Select plan type</option>
                                <option value="freemium">Freemium</option>
                                <option value="basic">Basic</option>
                                <option value="professional">Professional</option>
                                <option value="enterprise">Enterprise</option>
                            </select>
                        </div>
                    </div>
                    
                    <div class="form-section">
                        <h3>Pricing</h3>
                        <div class="form-row">
                            <div class="form-group">
                                <label for="monthlyPrice">Monthly Price ($)</label>
                                <input type="number" id="monthlyPrice" name="monthlyPrice" step="0.01" min="0" class="form-control" placeholder="29.99">
                            </div>
                            
                            <div class="form-group">
                                <label for="yearlyPrice">Yearly Price ($)</label>
                                <input type="number" id="yearlyPrice" name="yearlyPrice" step="0.01" min="0" class="form-control" placeholder="299.99">
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label for="trialDays">Trial Days</label>
                            <input type="number" id="trialDays" name="trialDays" min="0" max="365" class="form-control" placeholder="14">
                        </div>
                    </div>
                    
                    <div class="form-section">
                        <h3>Features & Limits</h3>
                        <div class="form-row">
                            <div class="form-group">
                                <label for="maxServices">Max Services</label>
                                <input type="number" id="maxServices" name="maxServices" min="1" class="form-control" placeholder="10">
                            </div>
                            
                            <div class="form-group">
                                <label for="maxUsers">Max Users</label>
                                <input type="number" id="maxUsers" name="maxUsers" min="1" class="form-control" placeholder="5">
                            </div>
                        </div>
                        
                        <div class="form-row">
                            <div class="form-group">
                                <label for="storageLimit">Storage Limit (GB)</label>
                                <input type="number" id="storageLimit" name="storageLimit" min="0" class="form-control" placeholder="100">
                            </div>
                            
                            <div class="form-group">
                                <label for="apiCalls">Monthly API Calls</label>
                                <input type="number" id="apiCalls" name="apiCalls" min="0" class="form-control" placeholder="10000">
                            </div>
                        </div>
                        
                        <div class="form-group">
                            <label>Features Included</label>
                            <div class="checkbox-group">
                                <label><input type="checkbox" name="features" value="custom_domain"> Custom Domain</label>
                                <label><input type="checkbox" name="features" value="ssl"> SSL Certificate</label>
                                <label><input type="checkbox" name="features" value="analytics"> Advanced Analytics</label>
                                <label><input type="checkbox" name="features" value="api_access"> API Access</label>
                                <label><input type="checkbox" name="features" value="white_label"> White Label</label>
                                <label><input type="checkbox" name="features" value="priority_support"> Priority Support</label>
                                <label><input type="checkbox" name="features" value="sla"> SLA Guarantee</label>
                                <label><input type="checkbox" name="features" value="backup"> Daily Backup</label>
                            </div>
                        </div>
                    </div>
                    
                    <div class="form-section">
                        <h3>Display Settings</h3>
                        <div class="form-group">
                            <label for="planColor">Plan Color</label>
                            <input type="color" id="planColor" name="color" value="#007bff" class="form-control">
                        </div>
                        
                        <div class="form-group">
                            <label for="planPosition">Display Position</label>
                            <input type="number" id="planPosition" name="position" min="1" class="form-control" placeholder="1">
                </div>
                        
                    <div class="form-group">
                            <label>
                                <input type="checkbox" id="featured" name="featured"> Mark as Featured Plan
                            </label>
                    </div>
                        
                    <div class="form-group">
                        <label>
                                <input type="checkbox" id="active" name="active" checked> Plan Active
                        </label>
                    </div>
                    </div>
                    
                    <div class="form-actions">
                        <button type="submit" class="btn-primary">
                            <i class="fas fa-plus"></i> Create Plan
                        </button>
                        <button type="button" class="btn-secondary" onclick="window.saasAdminInstance.loadSectionData('plans')">
                            Cancel
                        </button>
                    </div>
                </form>
            </div>
        `;

        // Add form submit handler
        document.getElementById('planForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.submitPlan();
        });
        
        // Auto-calculate yearly price with 20% discount
        document.getElementById('monthlyPrice').addEventListener('input', (e) => {
            const monthly = parseFloat(e.target.value);
            if (monthly > 0) {
                const yearly = (monthly * 12 * 0.8).toFixed(2); // 20% discount
                document.getElementById('yearlyPrice').value = yearly;
            }
        });
    }
    
    submitPlan() {
        console.log('submitPlan called');
        const form = document.getElementById('planForm');
        if (!form) {
            console.error('Plan form not found');
            this.showError('Plan form not found');
            return;
        }
        const formData = new FormData(form);
        
        // Get selected features
        const features = Array.from(document.querySelectorAll('input[name="features"]:checked'))
            .map(cb => cb.value);
        
        const planData = {
            name: formData.get('name'),
            description: formData.get('description'),
            type: formData.get('type'),
            monthlyPrice: parseFloat(formData.get('monthlyPrice')) || 0,
            yearlyPrice: parseFloat(formData.get('yearlyPrice')) || 0,
            trialDays: parseInt(formData.get('trialDays')) || 0,
            maxServices: parseInt(formData.get('maxServices')) || 0,
            maxUsers: parseInt(formData.get('maxUsers')) || 0,
            storageLimit: parseInt(formData.get('storageLimit')) || 0,
            apiCalls: parseInt(formData.get('apiCalls')) || 0,
            features: features,
            color: formData.get('color'),
            position: parseInt(formData.get('position')) || 1,
            featured: formData.get('featured') === 'on',
            active: formData.get('active') === 'on'
        };
        
        console.log('Creating plan:', planData);
        
        // Simulate API call
        console.log('Simulating plan creation...');
        
        setTimeout(async () => {
            alert('Plan created successfully!');
            this.loadSectionData('plans');
            
            // Auto-create a demo tenant for the new plan
            await this.createDemoTenantForPlan(planData);
        }, 1500);
    }

    async createDemoTenantForPlan(planData) {
        const demoTenantData = {
            name: `${planData.name} Demo Tenant`,
            slug: `${planData.name.toLowerCase().replace(/\s+/g, '-')}-demo`,
            contact_email: 'demo@example.com',
            billing_email: 'demo@example.com',
            subdomain: `${planData.name.toLowerCase().replace(/\s+/g, '-')}-demo`,
            plan: planData.name.toLowerCase(),
            status: 'active'
        };

        try {
            const response = await fetch('/api/v1/admin/saas/admin/tenants', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(demoTenantData)
            });

            if (response.ok) {
                this.showInfo('Demo tenant created automatically for the new plan!');
                this.loadTenantsData(); // Refresh tenants list
            } else {
                console.log('Demo tenant creation failed, but plan was created successfully');
            }
        } catch (error) {
            console.log('Demo tenant creation failed, but plan was created successfully');
        }
    }

    showCreateNotificationForm() {
        console.log('Create notification clicked');
        const content = document.getElementById('notifications-content');
        if (!content) {
            console.error('Notifications content element not found');
            this.showError('Unable to show notification form - notifications section not found');
            return;
        }
        content.innerHTML = `
            <div class="create-notification-form">
                <h2>Create Notification</h2>
                <form id="notificationForm" class="form-container">
                    <div class="form-group">
                        <label for="notificationTitle">Title</label>
                        <input type="text" id="notificationTitle" name="title" required class="form-control">
                    </div>
                    
                    <div class="form-group">
                        <label for="notificationMessage">Message</label>
                        <textarea id="notificationMessage" name="message" required class="form-control" rows="4"></textarea>
                    </div>
                    
                    <div class="form-group">
                        <label for="notificationType">Type</label>
                        <select id="notificationType" name="type" required class="form-control">
                            <option value="">Select notification type</option>
                            <option value="info">Information</option>
                            <option value="warning">Warning</option>
                            <option value="critical">Critical</option>
                            <option value="maintenance">Maintenance</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="notificationAudience">Audience</label>
                        <select id="notificationAudience" name="audience" required class="form-control">
                            <option value="">Select audience</option>
                            <option value="all">All Users</option>
                            <option value="subscribers">Subscribers Only</option>
                            <option value="admins">Admins Only</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="notificationChannels">Notification Channels</label>
                        <div class="checkbox-group">
                            <label><input type="checkbox" name="channels" value="email" checked> Email</label>
                            <label><input type="checkbox" name="channels" value="sms"> SMS</label>
                            <label><input type="checkbox" name="channels" value="push"> Push Notification</label>
                            <label><input type="checkbox" name="channels" value="slack"> Slack</label>
                        </div>
                    </div>
                    
                    <div class="form-group">
                        <label for="scheduleTime">Schedule (Optional)</label>
                        <input type="datetime-local" id="scheduleTime" name="scheduleTime" class="form-control">
                        <small>Leave empty to send immediately</small>
                    </div>
                    
                    <div class="form-actions">
                        <button type="submit" class="btn-primary">Create Notification</button>
                        <button type="button" class="btn-secondary" onclick="window.saasAdminInstance.loadSectionData('notifications')">Cancel</button>
                    </div>
                </form>
            </div>
        `;
        
        // Add form submit handler
        document.getElementById('notificationForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.submitNotification();
        });
    }
    
    submitNotification() {
        console.log('submitNotification called');
        const form = document.getElementById('notificationForm');
        if (!form) {
            console.error('Notification form not found');
            this.showError('Notification form not found');
            return;
        }
        const formData = new FormData(form);
        
        // Get selected channels
        const channels = Array.from(document.querySelectorAll('input[name="channels"]:checked'))
            .map(cb => cb.value);
        
        const notificationData = {
            title: formData.get('title'),
            message: formData.get('message'),
            type: formData.get('type'),
            audience: formData.get('audience'),
            channels: channels,
            scheduleTime: formData.get('scheduleTime') || null
        };
        
        console.log('Submitting notification:', notificationData);
        
        // Simulate API call
        console.log('Simulating notification creation...');
        
        setTimeout(() => {
            alert('Notification created successfully!');
            this.loadSectionData('notifications');
        }, 1500);
    }

    generateInvoiceReport() {
        console.log('Generate invoice report');
        
        const content = document.getElementById('billing-content');
        if (!content) {
            console.error('Billing content element not found');
            this.showError('Unable to generate report - billing section not found');
            return;
        }
        content.innerHTML = `
            <div class="report-generator">
                <h2>Generate Invoice Report</h2>
                <form id="reportForm" class="form-container">
                        <div class="form-group">
                        <label for="reportType">Report Type</label>
                        <select id="reportType" name="type" required class="form-control">
                            <option value="">Select report type</option>
                            <option value="monthly">Monthly Summary</option>
                            <option value="quarterly">Quarterly Report</option>
                            <option value="yearly">Yearly Report</option>
                            <option value="custom">Custom Date Range</option>
                        </select>
                        </div>
                    
                    <div class="form-group" id="dateRangeGroup" style="display: none;">
                        <label for="startDate">Start Date</label>
                        <input type="date" id="startDate" name="startDate" class="form-control">
                        <label for="endDate">End Date</label>
                        <input type="date" id="endDate" name="endDate" class="form-control">
                        </div>
                    
                    <div class="form-group">
                        <label for="tenantFilter">Tenant Filter</label>
                        <select id="tenantFilter" name="tenant" class="form-control">
                            <option value="all">All Tenants</option>
                            <option value="acme-corp">Acme Corporation</option>
                            <option value="tech-startup">Tech Startup Inc</option>
                            <option value="enterprise-ltd">Enterprise Ltd</option>
                        </select>
                    </div>
                    
                        <div class="form-group">
                        <label for="reportFormat">Format</label>
                        <select id="reportFormat" name="format" required class="form-control">
                            <option value="pdf">PDF</option>
                            <option value="excel">Excel</option>
                            <option value="csv">CSV</option>
                            </select>
                        </div>
                    
                        <div class="form-group">
                        <label>Include Sections</label>
                        <div class="checkbox-group">
                            <label><input type="checkbox" name="sections" value="summary" checked> Executive Summary</label>
                            <label><input type="checkbox" name="sections" value="details" checked> Detailed Transactions</label>
                            <label><input type="checkbox" name="sections" value="charts" checked> Charts & Graphs</label>
                            <label><input type="checkbox" name="sections" value="comparisons"> Period Comparisons</label>
                        </div>
                        </div>
                    
                    <div class="form-actions">
                        <button type="submit" class="btn-primary">
                            <i class="fas fa-file-download"></i> Generate Report
                        </button>
                        <button type="button" class="btn-secondary" onclick="window.saasAdminInstance.loadSectionData('tenants')">
                            Cancel
                        </button>
                    </div>
                </form>
                
                <div id="reportPreview" class="report-preview" style="display: none;">
                    <h3>Report Preview</h3>
                    <div class="preview-content">
                        <div class="report-summary">
                            <h4>Summary</h4>
                            <div class="summary-stats">
                                <div class="stat-item">
                                    <span class="label">Total Revenue:</span>
                                    <span class="value">$24,580</span>
                        </div>
                                <div class="stat-item">
                                    <span class="label">Total Invoices:</span>
                                    <span class="value">156</span>
                        </div>
                                <div class="stat-item">
                                    <span class="label">Average Invoice:</span>
                                    <span class="value">$157.56</span>
                    </div>
                        </div>
                    </div>
                    </div>
                </div>
            </div>
        `;

        // Add form handlers
        document.getElementById('reportForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.processReportGeneration();
        });
        
        document.getElementById('reportType').addEventListener('change', (e) => {
            const dateRangeGroup = document.getElementById('dateRangeGroup');
            if (e.target.value === 'custom') {
                dateRangeGroup.style.display = 'block';
            } else {
                dateRangeGroup.style.display = 'none';
            }
        });
    }
    
    processReportGeneration() {
        console.log('processReportGeneration called');
        const form = document.getElementById('reportForm');
        if (!form) {
            console.error('Report form not found');
            this.showError('Report form not found');
            return;
        }
        const formData = new FormData(form);
        
        // Get selected sections
        const sections = Array.from(document.querySelectorAll('input[name="sections"]:checked'))
            .map(cb => cb.value);
        
        const reportData = {
            type: formData.get('type'),
            tenant: formData.get('tenant'),
            format: formData.get('format'),
            sections: sections,
            startDate: formData.get('startDate'),
            endDate: formData.get('endDate')
        };
        
        console.log('Generating report with:', reportData);
        
        // Show loading and preview
        console.log('Simulating report generation...');
        
        setTimeout(() => {
            document.getElementById('reportPreview').style.display = 'block';
            
            // Simulate download
            const blob = new Blob(['Mock report content'], { type: 'text/plain' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `invoice-report-${reportData.type}.${reportData.format}`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
            
            alert('Report generated and downloaded successfully!');
        }, 2000);
    }

    refreshAnalytics() {
        console.log('Refresh analytics');
        this.loadSectionData('analytics');
    }

    // Notification functions
    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showInfo(message) {
        this.showNotification(message, 'info');
    }

    showNotification(message, type = 'info') {
        // Remove existing notifications
        const existingNotifications = document.querySelectorAll('.notification');
        existingNotifications.forEach(notification => notification.remove());

        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.innerHTML = `
            <div class="notification-content">
                <span class="notification-message">${message}</span>
                <button class="notification-close" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;

        // Add styles if not already present
        if (!document.getElementById('notification-styles')) {
            const styles = document.createElement('style');
            styles.id = 'notification-styles';
            styles.textContent = `
                .notification {
                    position: fixed;
                    top: 20px;
                    right: 20px;
                    z-index: 10000;
                    padding: 12px 16px;
                    border-radius: 6px;
                    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
                    max-width: 400px;
                    animation: slideIn 0.3s ease-out;
                }
                .notification-success {
                    background-color: #d4edda;
                    color: #155724;
                    border: 1px solid #c3e6cb;
                }
                .notification-error {
                    background-color: #f8d7da;
                    color: #721c24;
                    border: 1px solid #f5c6cb;
                }
                .notification-info {
                    background-color: #d1ecf1;
                    color: #0c5460;
                    border: 1px solid #bee5eb;
                }
                .notification-content {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                }
                .notification-close {
                    background: none;
                    border: none;
                    font-size: 18px;
                    cursor: pointer;
                    margin-left: 10px;
                    opacity: 0.7;
                }
                .notification-close:hover {
                    opacity: 1;
                }
                @keyframes slideIn {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
            `;
            document.head.appendChild(styles);
        }

        // Add to page
        document.body.appendChild(notification);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 5000);
    }
}
