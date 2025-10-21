// Error Boundary and Error Handling for Landing Page
class ErrorBoundary {
    constructor() {
        this.init();
        this.setupGlobalErrorHandling();
        this.setupUnhandledRejectionHandling();
        this.setupNetworkErrorHandling();
    }

    init() {
        // Create error UI container
        this.createErrorContainer();

        // Set up error reporting
        this.setupErrorReporting();

        // Handle browser compatibility issues
        this.handleBrowserCompatibility();
    }

    createErrorContainer() {
        const errorContainer = document.createElement('div');
        errorContainer.id = 'error-boundary';
        errorContainer.className = 'error-boundary hidden';
        errorContainer.innerHTML = `
            <div class="error-content">
                <div class="error-icon">
                    <i class="fas fa-exclamation-triangle"></i>
                </div>
                <h2 class="error-title">Oops! Something went wrong</h2>
                <p class="error-message">We're sorry, but an unexpected error occurred. Please try refreshing the page.</p>
                <div class="error-actions">
                    <button class="btn btn-primary" onclick="window.location.reload()">Refresh Page</button>
                    <button class="btn btn-outline" onclick="this.reportError()">Report Issue</button>
                </div>
                <details class="error-details">
                    <summary>Technical Details</summary>
                    <pre class="error-stack"></pre>
                </details>
            </div>
        `;

        document.body.appendChild(errorContainer);
        this.errorContainer = errorContainer;
    }

    setupGlobalErrorHandling() {
        window.addEventListener('error', (event) => {
            this.handleError({
                message: event.message,
                filename: event.filename,
                lineno: event.lineno,
                colno: event.colno,
                error: event.error,
                type: 'javascript'
            });
        });
    }

    setupUnhandledRejectionHandling() {
        window.addEventListener('unhandledrejection', (event) => {
            this.handleError({
                message: 'Unhandled Promise Rejection',
                error: event.reason,
                type: 'promise'
            });
        });
    }

    setupNetworkErrorHandling() {
        // Monitor failed resource loads
        window.addEventListener('error', (event) => {
            if (event.target !== window) {
                this.handleResourceError(event.target);
            }
        }, true);

        // Monitor fetch failures
        this.monitorFetchErrors();
    }

    handleError(errorInfo) {
        console.error('Error caught by boundary:', errorInfo);

        // Don't show error UI for minor issues
        if (this.isMinorError(errorInfo)) {
            return;
        }

        // Log error for analytics
        this.logError(errorInfo);

        // Show graceful error UI
        this.showErrorUI(errorInfo);

        // Attempt recovery
        this.attemptRecovery(errorInfo);
    }

    handleResourceError(element) {
        const resourceType = element.tagName.toLowerCase();
        const resourceSrc = element.src || element.href;

        console.warn(`Failed to load ${resourceType}:`, resourceSrc);

        // Attempt to recover from resource failures
        switch (resourceType) {
            case 'img':
                this.handleImageError(element);
                break;
            case 'script':
                this.handleScriptError(element);
                break;
            case 'link':
                this.handleStylesheetError(element);
                break;
        }
    }

    handleImageError(img) {
        // Replace with fallback image
        if (!img.dataset.fallbackAttempted) {
            img.dataset.fallbackAttempted = 'true';
            img.src = '/static/images/fallback-image.svg';
            img.alt = 'Image not available';
        }
    }

    handleScriptError(script) {
        // Log script loading failure
        console.error('Script failed to load:', script.src);

        // Disable features that depend on the failed script
        this.disableDependentFeatures(script);
    }

    handleStylesheetError(link) {
        // Log stylesheet loading failure
        console.error('Stylesheet failed to load:', link.href);

        // Apply fallback styles
        this.applyFallbackStyles();
    }

    monitorFetchErrors() {
        const originalFetch = window.fetch;

        window.fetch = async (...args) => {
            try {
                const response = await originalFetch(...args);

                if (!response.ok) {
                    this.handleFetchError(response, args[0]);
                }

                return response;
            } catch (error) {
                this.handleFetchError(error, args[0]);
                throw error;
            }
        };
    }

    handleFetchError(error, url) {
        console.error('Fetch error:', error, 'URL:', url);

        // Show user-friendly message for API failures
        if (typeof url === 'string' && url.includes('/api/')) {
            this.showNetworkErrorMessage();
        }
    }

    isMinorError(errorInfo) {
        const minorErrorPatterns = [
            /non-passive event listener/i,
            /script error/i,
            /network error/i
        ];

        return minorErrorPatterns.some(pattern =>
            pattern.test(errorInfo.message || '')
        );
    }

    showErrorUI(errorInfo) {
        const errorMessage = this.errorContainer.querySelector('.error-message');
        const errorStack = this.errorContainer.querySelector('.error-stack');

        // Customize error message based on error type
        switch (errorInfo.type) {
            case 'javascript':
                errorMessage.textContent = 'A JavaScript error occurred. The page may not function correctly.';
                break;
            case 'promise':
                errorMessage.textContent = 'An asynchronous operation failed. Some features may be unavailable.';
                break;
            default:
                errorMessage.textContent = 'An unexpected error occurred. Please try refreshing the page.';
        }

        // Show technical details in development
        if (this.isDevelopment()) {
            errorStack.textContent = errorInfo.error ? errorInfo.error.stack : JSON.stringify(errorInfo, null, 2);
        }

        // Show error container
        this.errorContainer.classList.remove('hidden');

        // Auto-hide after 10 seconds for non-critical errors
        if (!this.isCriticalError(errorInfo)) {
            setTimeout(() => {
                this.hideErrorUI();
            }, 10000);
        }
    }

    hideErrorUI() {
        this.errorContainer.classList.add('hidden');
    }

    showNetworkErrorMessage() {
        // Show a toast notification for network errors
        const toast = document.createElement('div');
        toast.className = 'error-toast';
        toast.innerHTML = `
            <div class="toast-content">
                <i class="fas fa-wifi" style="color: #ef4444;"></i>
                <span>Network error. Please check your connection.</span>
                <button class="toast-close" onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
        `;

        document.body.appendChild(toast);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (toast.parentElement) {
                toast.parentElement.removeChild(toast);
            }
        }, 5000);
    }

    attemptRecovery(errorInfo) {
        // Try to recover from common errors
        if (errorInfo.type === 'javascript') {
            // Restart failed components
            this.restartFailedComponents();
        }

        if (errorInfo.type === 'promise') {
            // Retry failed async operations
            this.retryFailedOperations();
        }
    }

    restartFailedComponents() {
        // Reinitialize components that might have failed
        try {
            if (window.modernLandingPage) {
                console.log('Attempting to restart landing page components...');
                // Reinitialize critical components
                window.modernLandingPage.init();
            }
        } catch (error) {
            console.error('Failed to restart components:', error);
        }
    }

    retryFailedOperations() {
        // Retry failed fetch operations or other async tasks
        console.log('Attempting to retry failed operations...');
    }

    disableDependentFeatures(failedScript) {
        const scriptName = failedScript.src.split('/').pop();

        // Disable features based on failed script
        switch (scriptName) {
            case 'enhanced-landing.js':
                console.warn('Enhanced features disabled due to script failure');
                document.body.classList.add('fallback-mode');
                break;
            case 'dark-mode.js':
                console.warn('Dark mode disabled due to script failure');
                break;
            case 'keyboard-navigation.js':
                console.warn('Enhanced keyboard navigation disabled');
                break;
        }
    }

    applyFallbackStyles() {
        // Apply basic fallback styles if stylesheets fail to load
        const fallbackStyles = `
            .fallback-mode {
                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            }
            .fallback-mode .btn {
                background: #007bff;
                color: white;
                padding: 10px 20px;
                border: none;
                border-radius: 4px;
                cursor: pointer;
            }
            .fallback-mode .btn:hover {
                background: #0056b3;
            }
        `;

        const styleSheet = document.createElement('style');
        styleSheet.textContent = fallbackStyles;
        document.head.appendChild(styleSheet);
    }

    logError(errorInfo) {
        // Log error to analytics service (implement based on your analytics setup)
        if (window.gtag) {
            window.gtag('event', 'exception', {
                description: errorInfo.message,
                fatal: this.isCriticalError(errorInfo)
            });
        }

        // Could also send to external error reporting service
        this.sendToErrorReporting(errorInfo);
    }

    sendToErrorReporting(errorInfo) {
        // Implement integration with error reporting service (e.g., Sentry, Bugsnag)
        if (this.isDevelopment()) {
            console.log('Error would be sent to reporting service:', errorInfo);
            return;
        }

        // Example implementation for external service
        /*
        fetch('/api/error-report', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                ...errorInfo,
                userAgent: navigator.userAgent,
                url: window.location.href,
                timestamp: new Date().toISOString()
            })
        }).catch(err => console.error('Failed to report error:', err));
        */
    }

    setupErrorReporting() {
        // Add report error functionality
        window.reportError = (errorInfo) => {
            const subject = encodeURIComponent('Website Error Report');
            const body = encodeURIComponent(`
Error Details:
${JSON.stringify(errorInfo, null, 2)}

Browser: ${navigator.userAgent}
URL: ${window.location.href}
Timestamp: ${new Date().toISOString()}
            `);

            window.open(`mailto:support@example.com?subject=${subject}&body=${body}`);
        };
    }

    handleBrowserCompatibility() {
        // Check for required browser features
        const requiredFeatures = [
            'IntersectionObserver',
            'fetch',
            'Promise',
            'addEventListener'
        ];

        const unsupportedFeatures = requiredFeatures.filter(feature =>
            !window[feature]
        );

        if (unsupportedFeatures.length > 0) {
            this.showBrowserCompatibilityWarning(unsupportedFeatures);
        }
    }

    showBrowserCompatibilityWarning(unsupportedFeatures) {
        const warning = document.createElement('div');
        warning.className = 'browser-warning';
        warning.innerHTML = `
            <div class="warning-content">
                <h3>Browser Compatibility Notice</h3>
                <p>Your browser doesn't support some modern features used by this website. For the best experience, please update your browser.</p>
                <button onclick="this.parentElement.parentElement.remove()">Dismiss</button>
            </div>
        `;

        document.body.insertBefore(warning, document.body.firstChild);
    }

    isCriticalError(errorInfo) {
        const criticalPatterns = [
            /cannot read property/i,
            /undefined is not a function/i,
            /script error/i
        ];

        return criticalPatterns.some(pattern =>
            pattern.test(errorInfo.message || '')
        );
    }

    isDevelopment() {
        return window.location.hostname === 'localhost' ||
               window.location.hostname === '127.0.0.1' ||
               window.location.hostname.includes('.local');
    }
}

// Error boundary styles
const errorBoundaryStyles = `
    .error-boundary {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.8);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10000;
        transition: opacity 0.3s ease;
    }

    .error-boundary.hidden {
        display: none;
    }

    .error-content {
        background: white;
        padding: 2rem;
        border-radius: 8px;
        max-width: 500px;
        text-align: center;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
    }

    .error-icon {
        font-size: 3rem;
        color: #ef4444;
        margin-bottom: 1rem;
    }

    .error-title {
        font-size: 1.5rem;
        font-weight: 600;
        margin-bottom: 1rem;
        color: #1f2937;
    }

    .error-message {
        color: #6b7280;
        margin-bottom: 2rem;
        line-height: 1.6;
    }

    .error-actions {
        display: flex;
        gap: 1rem;
        justify-content: center;
        margin-bottom: 1rem;
    }

    .error-details {
        text-align: left;
        margin-top: 1rem;
    }

    .error-details summary {
        cursor: pointer;
        color: #6b7280;
        font-size: 0.875rem;
    }

    .error-stack {
        background: #f3f4f6;
        padding: 1rem;
        border-radius: 4px;
        font-size: 0.75rem;
        max-height: 200px;
        overflow-y: auto;
        margin-top: 0.5rem;
    }

    .error-toast {
        position: fixed;
        bottom: 20px;
        right: 20px;
        background: white;
        border: 1px solid #e5e7eb;
        border-radius: 8px;
        padding: 1rem;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        z-index: 1000;
        max-width: 300px;
    }

    .toast-content {
        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .toast-close {
        background: none;
        border: none;
        font-size: 1.2rem;
        cursor: pointer;
        margin-left: auto;
        color: #6b7280;
    }

    .browser-warning {
        background: #fef3c7;
        border-bottom: 1px solid #f59e0b;
        padding: 1rem;
        text-align: center;
    }

    .warning-content h3 {
        margin: 0 0 0.5rem 0;
        color: #92400e;
    }

    .warning-content p {
        margin: 0 0 1rem 0;
        color: #78350f;
    }

    .warning-content button {
        background: #f59e0b;
        color: white;
        border: none;
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
    }

    [data-theme="dark"] .error-content {
        background: #1f2937;
        color: #f9fafb;
    }

    [data-theme="dark"] .error-title {
        color: #f9fafb;
    }

    [data-theme="dark"] .error-stack {
        background: #374151;
        color: #e5e7eb;
    }
`;

// Initialize error boundary
document.addEventListener('DOMContentLoaded', () => {
    window.errorBoundary = new ErrorBoundary();

    // Inject error boundary styles
    const styleSheet = document.createElement('style');
    styleSheet.textContent = errorBoundaryStyles;
    document.head.appendChild(styleSheet);
});