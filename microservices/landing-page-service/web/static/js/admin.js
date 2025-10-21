/**
 * Admin Dashboard JavaScript
 * Handles all admin panel interactions and API calls
 */

// API Base URL
const API_BASE = '/api/v1/admin';

// State Management
const state = {
    currentSection: 'dashboard',
    currentTab: 'hero',
    selectedMedia: [],
    abTests: [],
    content: {
        hero: {},
        features: [],
        testimonials: [],
        faqs: [],
        articles: []
    }
};

// Initialize Dashboard
document.addEventListener('DOMContentLoaded', () => {
    initializeNavigation();
    initializeTabs();
    initializeCharts();
    loadDashboardData();
    initializeMediaUpload();
    initializeContentForms();
});

// Navigation
function initializeNavigation() {
    const navItems = document.querySelectorAll('.sidebar-nav a');
    navItems.forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const section = item.dataset.section;
            switchSection(section);
        });
    });

    // Sidebar toggle for mobile
    const sidebarToggle = document.getElementById('sidebarToggle');
    sidebarToggle.addEventListener('click', () => {
        document.querySelector('.admin-sidebar').classList.toggle('active');
    });
}

function switchSection(section) {
    // Update nav active state
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-section="${section}"]`).parentElement.classList.add('active');

    // Show/hide sections
    document.querySelectorAll('.content-section').forEach(sec => {
        sec.classList.remove('active');
    });
    document.getElementById(`${section}-section`).classList.add('active');

    // Load section-specific data
    loadSectionData(section);
    state.currentSection = section;
}

// Tabs
function initializeTabs() {
    const tabBtns = document.querySelectorAll('.tab-btn');
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const tab = btn.dataset.tab;
            switchTab(tab);
        });
    });
}

function switchTab(tab) {
    // Update tab button states
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
        if (btn.dataset.tab === tab) {
            btn.classList.add('active');
        }
    });

    // Show/hide tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(`${tab}-tab`).classList.add('active');

    // Load tab-specific data
    loadTabData(tab);
    state.currentTab = tab;
}

// Dashboard Data Loading
async function loadDashboardData() {
    try {
        // Load stats
        const stats = await fetchAPI('/stats');
        updateDashboardStats(stats);

        // Load recent activity
        const activity = await fetchAPI('/activity');
        updateActivityFeed(activity);
    } catch (error) {
        showError('Failed to load dashboard data');
    }
}

function updateDashboardStats(stats) {
    // Update stat cards with actual data
    if (stats) {
        document.querySelectorAll('.stat-value').forEach((elem, index) => {
            const statKeys = ['pageViews', 'uniqueVisitors', 'activeTests', 'conversionRate'];
            if (stats[statKeys[index]]) {
                elem.textContent = stats[statKeys[index]];
            }
        });
    }
}

function updateActivityFeed(activity) {
    const activityList = document.querySelector('.activity-list');
    if (activity && activity.items) {
        activityList.innerHTML = activity.items.map(item => `
            <div class="activity-item">
                <div class="activity-icon">
                    <i class="fas fa-${item.icon}"></i>
                </div>
                <div class="activity-details">
                    <p><strong>${item.title}</strong></p>
                    <span class="activity-time">${item.time}</span>
                </div>
            </div>
        `).join('');
    }
}

// Load Section-specific Data
async function loadSectionData(section) {
    switch(section) {
        case 'content':
            await loadContentData();
            break;
        case 'media':
            await loadMediaLibrary();
            break;
        case 'ab-tests':
            await loadABTests();
            break;
        case 'seo':
            await loadSEOData();
            break;
        case 'analytics':
            await loadAnalytics();
            break;
    }
}

// Content Management
async function loadContentData() {
    try {
        const hero = await fetchAPI('/hero');
        if (hero) {
            state.content.hero = hero;
            updateHeroPreview(hero);
        }
    } catch (error) {
        showError('Failed to load content');
    }
}

async function loadTabData(tab) {
    try {
        switch(tab) {
            case 'features':
                const features = await fetchAPI('/features');
                updateFeaturesList(features);
                break;
            case 'testimonials':
                const testimonials = await fetchAPI('/testimonials');
                updateTestimonialsList(testimonials);
                break;
            case 'faqs':
                const faqs = await fetchAPI('/faqs');
                updateFAQsList(faqs);
                break;
            case 'blog':
                const articles = await fetchAPI('/articles');
                updateArticlesList(articles);
                break;
        }
    } catch (error) {
        showError(`Failed to load ${tab} data`);
    }
}

// Hero Section Management
function editHero() {
    const form = document.getElementById('hero-edit-form');
    const preview = document.querySelector('.hero-preview');

    // Populate form with current data
    document.getElementById('heroTitle').value = state.content.hero.title || '';
    document.getElementById('heroSubtitle').value = state.content.hero.subtitle || '';
    document.getElementById('heroCTA').value = state.content.hero.cta_text || '';
    document.getElementById('heroCTALink').value = state.content.hero.cta_link || '';

    // Show form, hide preview
    form.style.display = 'block';
    preview.style.display = 'none';
}

function cancelHeroEdit() {
    const form = document.getElementById('hero-edit-form');
    const preview = document.querySelector('.hero-preview');

    form.style.display = 'none';
    preview.style.display = 'block';
}

function updateHeroPreview(hero) {
    document.getElementById('hero-title').textContent = hero.title || 'Hero Title';
    document.getElementById('hero-subtitle').textContent = hero.subtitle || 'Hero Subtitle';
    document.querySelector('.preview-cta').textContent = hero.cta_text || 'Get Started';
}

// Features Management
function updateFeaturesList(features) {
    const container = document.getElementById('featuresList');
    if (!features || features.length === 0) {
        container.innerHTML = '<p>No features added yet.</p>';
        return;
    }

    container.innerHTML = features.map(feature => `
        <div class="feature-item">
            <div class="feature-icon">
                <i class="fas fa-${feature.icon}"></i>
            </div>
            <div class="feature-content">
                <h4>${feature.title}</h4>
                <p>${feature.description}</p>
            </div>
            <div class="feature-actions">
                <button onclick="editFeature(${feature.id})" class="btn btn-sm">Edit</button>
                <button onclick="deleteFeature(${feature.id})" class="btn btn-sm btn-danger">Delete</button>
            </div>
        </div>
    `).join('');
}

function addFeature() {
    // Open modal or inline form for adding new feature
    showFeatureModal();
}

// Media Library
async function loadMediaLibrary() {
    try {
        const media = await fetchAPI('/media');
        updateMediaGrid(media);
    } catch (error) {
        showError('Failed to load media library');
    }
}

function updateMediaGrid(media) {
    const grid = document.getElementById('mediaGrid');
    if (!media || media.length === 0) {
        grid.innerHTML = '<p>No media files uploaded yet.</p>';
        return;
    }

    grid.innerHTML = media.map(item => `
        <div class="media-item" data-id="${item.id}">
            ${item.type === 'image' ?
                `<img src="${item.url}" alt="${item.name}">` :
                `<div class="media-placeholder"><i class="fas fa-file"></i></div>`
            }
            <div class="media-info">
                <p>${item.name}</p>
                <span>${formatFileSize(item.size)}</span>
            </div>
        </div>
    `).join('');
}

// Media Upload
function initializeMediaUpload() {
    const uploadArea = document.getElementById('uploadArea');
    const fileInput = document.getElementById('fileInput');

    // Click to upload
    uploadArea.addEventListener('click', () => {
        fileInput.click();
    });

    // Drag and drop
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('dragging');
    });

    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('dragging');
    });

    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('dragging');
        handleFiles(e.dataTransfer.files);
    });

    // File input change
    fileInput.addEventListener('change', (e) => {
        handleFiles(e.target.files);
    });
}

async function handleFiles(files) {
    const uploadQueue = document.getElementById('uploadQueue');
    uploadQueue.innerHTML = '';

    for (let file of files) {
        const formData = new FormData();
        formData.append('file', file);

        try {
            const result = await uploadFile(formData, (progress) => {
                // Update progress UI
                updateUploadProgress(file.name, progress);
            });

            showSuccess(`${file.name} uploaded successfully`);
            loadMediaLibrary(); // Refresh media grid
        } catch (error) {
            showError(`Failed to upload ${file.name}`);
        }
    }
}

async function uploadFile(formData, onProgress) {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                const progress = (e.loaded / e.total) * 100;
                onProgress(progress);
            }
        });

        xhr.addEventListener('load', () => {
            if (xhr.status === 200) {
                resolve(JSON.parse(xhr.responseText));
            } else {
                reject(new Error('Upload failed'));
            }
        });

        xhr.addEventListener('error', () => {
            reject(new Error('Upload failed'));
        });

        xhr.open('POST', `${API_BASE}/media/upload`);
        xhr.send(formData);
    });
}

// A/B Testing
async function loadABTests() {
    try {
        const tests = await fetchAPI('/ab-tests');
        state.abTests = tests;
        updateTestsList(tests);
    } catch (error) {
        showError('Failed to load A/B tests');
    }
}

function updateTestsList(tests) {
    const container = document.getElementById('testsList');
    if (!tests || tests.length === 0) {
        container.innerHTML = '<p>No A/B tests created yet.</p>';
        return;
    }

    container.innerHTML = tests.map(test => `
        <div class="test-card">
            <div class="test-info">
                <h4>${test.name}</h4>
                <p>${test.element_type} - Traffic split: ${test.traffic_split}%</p>
                <span class="test-status ${test.status}">${test.status}</span>
            </div>
            <div class="test-results">
                <div class="variant-result">
                    <h5>Variant A</h5>
                    <p class="conversion-rate">${test.conversion_rate_a || 0}%</p>
                    <span class="sample-size">${test.views_a || 0} views</span>
                </div>
                <div class="variant-result">
                    <h5>Variant B</h5>
                    <p class="conversion-rate">${test.conversion_rate_b || 0}%</p>
                    <span class="sample-size">${test.views_b || 0} views</span>
                </div>
            </div>
            <div class="test-actions">
                ${test.status === 'running' ?
                    `<button onclick="stopABTest(${test.id})" class="btn btn-warning">Stop</button>` :
                    `<button onclick="startABTest(${test.id})" class="btn btn-success">Start</button>`
                }
                <button onclick="viewABTestDetails(${test.id})" class="btn">Details</button>
            </div>
        </div>
    `).join('');
}

function createABTest() {
    // Open modal for creating new A/B test
    showABTestModal();
}

async function startABTest(testId) {
    try {
        await postAPI(`/ab-tests/${testId}/start`);
        showSuccess('A/B test started');
        loadABTests();
    } catch (error) {
        showError('Failed to start A/B test');
    }
}

async function stopABTest(testId) {
    try {
        await postAPI(`/ab-tests/${testId}/stop`);
        showSuccess('A/B test stopped');
        loadABTests();
    } catch (error) {
        showError('Failed to stop A/B test');
    }
}

// SEO Management
async function loadSEOData() {
    try {
        const config = await fetchAPI('/seo/config');
        updateSEOConfig(config);

        const recommendations = await fetchAPI('/seo/recommendations');
        updateSEORecommendations(recommendations);
    } catch (error) {
        showError('Failed to load SEO data');
    }
}

function updateSEOConfig(config) {
    // Update SEO configuration forms
    // Populate forms with current config
}

function updateSEORecommendations(recommendations) {
    // Display SEO recommendations
}

// Analytics
async function loadAnalytics() {
    try {
        const analytics = await fetchAPI('/analytics');
        updateAnalyticsCharts(analytics);
    } catch (error) {
        showError('Failed to load analytics');
    }
}

// Charts
function initializeCharts() {
    // Traffic Chart
    const trafficCtx = document.getElementById('trafficChart');
    if (trafficCtx) {
        new Chart(trafficCtx, {
            type: 'line',
            data: {
                labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
                datasets: [{
                    label: 'Page Views',
                    data: [3200, 3500, 3100, 3800, 4200, 4500, 4100],
                    borderColor: '#3B82F6',
                    backgroundColor: 'rgba(59, 130, 246, 0.1)',
                    tension: 0.3
                }, {
                    label: 'Unique Visitors',
                    data: [1200, 1400, 1100, 1600, 1800, 2000, 1700],
                    borderColor: '#10B981',
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    tension: 0.3
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom'
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

    // Funnel Chart
    const funnelCtx = document.getElementById('funnelChart');
    if (funnelCtx) {
        new Chart(funnelCtx, {
            type: 'bar',
            data: {
                labels: ['Visitors', 'Engaged', 'Clicked CTA', 'Signed Up'],
                datasets: [{
                    label: 'Conversion Funnel',
                    data: [10000, 6500, 2000, 760],
                    backgroundColor: [
                        '#3B82F6',
                        '#60A5FA',
                        '#93C5FD',
                        '#DBEAFE'
                    ]
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                indexAxis: 'y',
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }
}

// Form Handling
function initializeContentForms() {
    // Hero form
    const heroForm = document.getElementById('heroForm');
    if (heroForm) {
        heroForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(heroForm);
            const data = Object.fromEntries(formData);

            try {
                await putAPI('/hero', data);
                showSuccess('Hero section updated');
                cancelHeroEdit();
                loadContentData();
            } catch (error) {
                showError('Failed to update hero section');
            }
        });
    }
}

// API Helper Functions
async function fetchAPI(endpoint) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        headers: {
            'Authorization': `Bearer ${getAuthToken()}`
        }
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
}

async function postAPI(endpoint, data = {}) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${getAuthToken()}`
        },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
}

async function putAPI(endpoint, data) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${getAuthToken()}`
        },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
}

async function deleteAPI(endpoint) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${getAuthToken()}`
        }
    });

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
}

// Auth Helper
function getAuthToken() {
    // Get token from localStorage or cookie
    return localStorage.getItem('admin_token') || '';
}

// UI Helper Functions
function showSuccess(message) {
    showNotification(message, 'success');
}

function showError(message) {
    showNotification(message, 'error');
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `alert alert-${type}`;
    notification.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
        <span>${message}</span>
    `;

    document.body.appendChild(notification);

    // Animation
    setTimeout(() => notification.classList.add('show'), 10);

    // Remove after 3 seconds
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// Modal Functions
function openUploadModal() {
    document.getElementById('uploadModal').classList.add('active');
}

function closeUploadModal() {
    document.getElementById('uploadModal').classList.remove('active');
}

function showFeatureModal(featureId = null) {
    // Create and show modal for adding/editing features
}

function showABTestModal() {
    // Create and show modal for creating A/B test
}

// Publish Changes
document.getElementById('publishChanges').addEventListener('click', async () => {
    if (confirm('Are you sure you want to publish all pending changes?')) {
        try {
            await postAPI('/publish');
            showSuccess('Changes published successfully');
        } catch (error) {
            showError('Failed to publish changes');
        }
    }
});

// Export functions for global access
window.editHero = editHero;
window.cancelHeroEdit = cancelHeroEdit;
window.addFeature = addFeature;
window.editFeature = (id) => console.log('Edit feature', id);
window.deleteFeature = (id) => console.log('Delete feature', id);
window.createABTest = createABTest;
window.startABTest = startABTest;
window.stopABTest = stopABTest;
window.viewABTestDetails = (id) => console.log('View test details', id);
window.openUploadModal = openUploadModal;
window.closeUploadModal = closeUploadModal;