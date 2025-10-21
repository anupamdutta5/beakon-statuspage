// Enhanced Keyboard Navigation and Accessibility
class KeyboardNavigationManager {
    constructor() {
        this.focusableElements = [
            'a[href]',
            'button:not([disabled])',
            'input:not([disabled])',
            'select:not([disabled])',
            'textarea:not([disabled])',
            '[tabindex]:not([tabindex="-1"])'
        ];
        this.init();
    }

    init() {
        this.setupKeyboardNavigation();
        this.setupFocusManagement();
        this.setupSkipLinks();
        this.setupModalKeyboardTrap();
    }

    setupKeyboardNavigation() {
        document.addEventListener('keydown', (e) => {
            // Global keyboard shortcuts
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'k':
                        e.preventDefault();
                        this.openSearch();
                        break;
                    case '/':
                        e.preventDefault();
                        this.focusSearchInput();
                        break;
                }
            }

            // Navigation shortcuts
            switch (e.key) {
                case 'Escape':
                    this.closeAllModals();
                    this.closeMobileMenu();
                    break;
                case 'Tab':
                    this.handleTabNavigation(e);
                    break;
                case 'Enter':
                case ' ':
                    this.handleEnterSpace(e);
                    break;
                case 'ArrowDown':
                case 'ArrowUp':
                    this.handleArrowNavigation(e);
                    break;
            }
        });
    }

    setupFocusManagement() {
        // Track focus for better visibility
        document.addEventListener('focusin', (e) => {
            const element = e.target;
            if (this.isKeyboardUser()) {
                element.classList.add('focus-visible');
            }
        });

        document.addEventListener('focusout', (e) => {
            e.target.classList.remove('focus-visible');
        });

        // Detect keyboard vs mouse usage
        let hadKeyboardEvent = true;
        document.addEventListener('mousedown', () => {
            hadKeyboardEvent = false;
        });

        document.addEventListener('keydown', () => {
            hadKeyboardEvent = true;
        });

        this.isKeyboardUser = () => hadKeyboardEvent;
    }

    setupSkipLinks() {
        // Create skip to main content link
        const skipLink = document.createElement('a');
        skipLink.href = '#main-content';
        skipLink.className = 'skip-link';
        skipLink.textContent = 'Skip to main content';
        skipLink.addEventListener('click', (e) => {
            e.preventDefault();
            const mainContent = document.getElementById('main-content') || document.querySelector('main');
            if (mainContent) {
                mainContent.focus();
                mainContent.scrollIntoView();
            }
        });

        document.body.insertBefore(skipLink, document.body.firstChild);
    }

    setupModalKeyboardTrap() {
        document.addEventListener('keydown', (e) => {
            const activeModal = document.querySelector('.modal.active');
            if (activeModal && e.key === 'Tab') {
                this.trapFocus(e, activeModal);
            }
        });
    }

    trapFocus(e, container) {
        const focusableElements = container.querySelectorAll(this.focusableElements.join(', '));
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        if (e.shiftKey) {
            if (document.activeElement === firstElement) {
                e.preventDefault();
                lastElement.focus();
            }
        } else {
            if (document.activeElement === lastElement) {
                e.preventDefault();
                firstElement.focus();
            }
        }
    }

    handleTabNavigation(e) {
        // Enhanced tab navigation for complex components
        const activeElement = document.activeElement;

        // Special handling for dropdown menus
        if (activeElement.closest('.dropdown')) {
            this.handleDropdownNavigation(e);
        }

        // Special handling for card grids
        if (activeElement.closest('.features-grid') || activeElement.closest('.pricing-grid')) {
            this.handleGridNavigation(e);
        }
    }

    handleEnterSpace(e) {
        const element = e.target;

        // Handle custom interactive elements
        if (element.classList.contains('faq-question')) {
            e.preventDefault();
            element.click();
        }

        if (element.classList.contains('pricing-toggle')) {
            e.preventDefault();
            element.click();
        }
    }

    handleArrowNavigation(e) {
        const activeElement = document.activeElement;

        // Arrow navigation for FAQ items
        if (activeElement.closest('.faq-item')) {
            this.handleFAQArrowNavigation(e);
        }

        // Arrow navigation for pricing cards
        if (activeElement.closest('.pricing-grid')) {
            this.handlePricingArrowNavigation(e);
        }
    }

    handleFAQArrowNavigation(e) {
        const currentFAQ = e.target.closest('.faq-item');
        const allFAQs = Array.from(document.querySelectorAll('.faq-item'));
        const currentIndex = allFAQs.indexOf(currentFAQ);

        let targetIndex;
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            targetIndex = currentIndex + 1;
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            targetIndex = currentIndex - 1;
        }

        if (targetIndex >= 0 && targetIndex < allFAQs.length) {
            const targetQuestion = allFAQs[targetIndex].querySelector('.faq-question');
            if (targetQuestion) {
                targetQuestion.focus();
            }
        }
    }

    handlePricingArrowNavigation(e) {
        const currentCard = e.target.closest('.pricing-card');
        const allCards = Array.from(document.querySelectorAll('.pricing-card'));
        const currentIndex = allCards.indexOf(currentCard);

        let targetIndex;
        if (e.key === 'ArrowRight') {
            e.preventDefault();
            targetIndex = currentIndex + 1;
        } else if (e.key === 'ArrowLeft') {
            e.preventDefault();
            targetIndex = currentIndex - 1;
        }

        if (targetIndex >= 0 && targetIndex < allCards.length) {
            const targetButton = allCards[targetIndex].querySelector('.btn');
            if (targetButton) {
                targetButton.focus();
            }
        }
    }

    openSearch() {
        // Implementation for search modal
        console.log('Search functionality would open here');
    }

    focusSearchInput() {
        const searchInput = document.querySelector('input[type="search"]');
        if (searchInput) {
            searchInput.focus();
        }
    }

    closeAllModals() {
        const modals = document.querySelectorAll('.modal.active');
        modals.forEach(modal => {
            modal.classList.remove('active');
        });
    }

    closeMobileMenu() {
        const navToggle = document.getElementById('nav-toggle');
        const navMenu = document.getElementById('nav-menu');

        if (navToggle && navMenu) {
            navToggle.classList.remove('active');
            navMenu.classList.remove('active');
        }
    }

    // Announce content changes to screen readers
    announceToScreenReader(message) {
        const announcement = document.createElement('div');
        announcement.setAttribute('aria-live', 'polite');
        announcement.setAttribute('aria-atomic', 'true');
        announcement.className = 'sr-only';
        announcement.textContent = message;

        document.body.appendChild(announcement);

        setTimeout(() => {
            document.body.removeChild(announcement);
        }, 1000);
    }
}

// Accessibility styles
const accessibilityStyles = `
    .skip-link {
        position: absolute;
        top: -40px;
        left: 6px;
        background: var(--primary-color);
        color: white;
        padding: 8px 16px;
        border-radius: 4px;
        text-decoration: none;
        font-weight: 600;
        z-index: 1000;
        transition: top 0.3s ease;
    }

    .skip-link:focus {
        top: 6px;
    }

    .focus-visible {
        outline: 2px solid var(--primary-color) !important;
        outline-offset: 2px !important;
        border-radius: 4px;
    }

    .sr-only {
        position: absolute;
        width: 1px;
        height: 1px;
        padding: 0;
        margin: -1px;
        overflow: hidden;
        clip: rect(0, 0, 0, 0);
        white-space: nowrap;
        border: 0;
    }

    /* High contrast mode support */
    @media (prefers-contrast: high) {
        .btn {
            border: 2px solid currentColor;
        }

        .card {
            border: 1px solid currentColor;
        }
    }

    /* Reduced motion preferences */
    @media (prefers-reduced-motion: reduce) {
        * {
            animation-duration: 0.01ms !important;
            animation-iteration-count: 1 !important;
            transition-duration: 0.01ms !important;
        }
    }

    /* Focus management for interactive elements */
    .faq-question {
        cursor: pointer;
        border-radius: 4px;
        transition: background-color 0.2s ease;
    }

    .faq-question:hover,
    .faq-question:focus {
        background-color: rgba(0, 0, 0, 0.05);
    }

    .pricing-card:focus-within {
        transform: translateY(-4px);
        box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
    }

    /* Ensure interactive elements are large enough for touch */
    .btn,
    .nav-link,
    .faq-question,
    .social-link {
        min-height: 44px;
        min-width: 44px;
        display: flex;
        align-items: center;
        justify-content: center;
    }
`;

// Initialize keyboard navigation
document.addEventListener('DOMContentLoaded', () => {
    window.keyboardNavManager = new KeyboardNavigationManager();

    // Inject accessibility styles
    const styleSheet = document.createElement('style');
    styleSheet.textContent = accessibilityStyles;
    document.head.appendChild(styleSheet);
});