// Dark Mode Toggle Functionality
class DarkModeManager {
    constructor() {
        this.darkMode = this.getStoredTheme() || this.getSystemTheme();
        this.init();
    }

    init() {
        this.createToggleButton();
        this.applyTheme(this.darkMode);
        this.setupEventListeners();
        this.setupSystemThemeListener();
    }

    createToggleButton() {
        // Create dark mode toggle button
        const toggleButton = document.createElement('button');
        toggleButton.className = 'dark-mode-toggle';
        toggleButton.setAttribute('aria-label', 'Toggle dark mode');
        toggleButton.innerHTML = `
            <svg class="sun-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="5"/>
                <line x1="12" y1="1" x2="12" y2="3"/>
                <line x1="12" y1="21" x2="12" y2="23"/>
                <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
                <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
                <line x1="1" y1="12" x2="3" y2="12"/>
                <line x1="21" y1="12" x2="23" y2="12"/>
                <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
                <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
            </svg>
            <svg class="moon-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
            </svg>
        `;

        // Add to navigation
        const navbar = document.querySelector('.navbar .nav-container');
        if (navbar) {
            navbar.appendChild(toggleButton);
        }

        this.toggleButton = toggleButton;
    }

    setupEventListeners() {
        if (this.toggleButton) {
            this.toggleButton.addEventListener('click', () => {
                this.toggle();
            });
        }

        // Keyboard accessibility
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.shiftKey && e.key === 'D') {
                this.toggle();
            }
        });
    }

    setupSystemThemeListener() {
        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.addListener((e) => {
            if (!this.getStoredTheme()) {
                this.applyTheme(e.matches ? 'dark' : 'light');
            }
        });
    }

    toggle() {
        this.darkMode = this.darkMode === 'dark' ? 'light' : 'dark';
        this.applyTheme(this.darkMode);
        this.storeTheme(this.darkMode);
    }

    applyTheme(theme) {
        this.darkMode = theme;
        document.documentElement.setAttribute('data-theme', theme);

        // Update toggle button state
        if (this.toggleButton) {
            this.toggleButton.setAttribute('data-theme', theme);
        }

        // Dispatch custom event for other components
        document.dispatchEvent(new CustomEvent('themechange', {
            detail: { theme }
        }));
    }

    getStoredTheme() {
        return localStorage.getItem('theme');
    }

    storeTheme(theme) {
        localStorage.setItem('theme', theme);
    }

    getSystemTheme() {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }

    getCurrentTheme() {
        return this.darkMode;
    }
}

// Initialize dark mode when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.darkModeManager = new DarkModeManager();
});

// CSS for dark mode toggle
const darkModeStyles = `
    .dark-mode-toggle {
        position: relative;
        background: none;
        border: none;
        padding: 8px;
        border-radius: 50%;
        cursor: pointer;
        transition: all 0.3s ease;
        margin-left: 16px;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .dark-mode-toggle:hover {
        background: rgba(0, 0, 0, 0.1);
    }

    .dark-mode-toggle .sun-icon,
    .dark-mode-toggle .moon-icon {
        position: absolute;
        transition: all 0.3s ease;
    }

    .dark-mode-toggle[data-theme="light"] .sun-icon {
        opacity: 1;
        transform: rotate(0deg);
    }

    .dark-mode-toggle[data-theme="light"] .moon-icon {
        opacity: 0;
        transform: rotate(180deg);
    }

    .dark-mode-toggle[data-theme="dark"] .sun-icon {
        opacity: 0;
        transform: rotate(180deg);
    }

    .dark-mode-toggle[data-theme="dark"] .moon-icon {
        opacity: 1;
        transform: rotate(0deg);
    }

    /* Dark theme styles */
    [data-theme="dark"] {
        --primary-color: #818cf8;
        --primary-dark: #6366f1;
        --gray-50: #111827;
        --gray-100: #1f2937;
        --gray-200: #374151;
        --gray-300: #4b5563;
        --gray-400: #6b7280;
        --gray-500: #9ca3af;
        --gray-600: #d1d5db;
        --gray-700: #e5e7eb;
        --gray-800: #f3f4f6;
        --gray-900: #f9fafb;
    }

    [data-theme="dark"] body {
        background: #111827;
        color: #f9fafb;
    }

    [data-theme="dark"] .glass-nav {
        background: rgba(17, 24, 39, 0.95);
        backdrop-filter: blur(20px);
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }

    [data-theme="dark"] .glass-card {
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    [data-theme="dark"] .dark-mode-toggle:hover {
        background: rgba(255, 255, 255, 0.1);
    }
`;

// Inject dark mode styles
const styleSheet = document.createElement('style');
styleSheet.textContent = darkModeStyles;
document.head.appendChild(styleSheet);