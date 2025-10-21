/**
 * A/B Testing Client Library
 * Handles variant assignment, tracking, and application on the frontend
 */

class ABTestClient {
    constructor(config = {}) {
        this.apiBase = config.apiBase || '/api/v1';
        this.debug = config.debug || false;
        this.autoTrack = config.autoTrack !== false; // Default true
        this.cookieName = 'ab_session_id';
        this.activeTests = new Map();
        this.conversionEvents = new Set();

        this.init();
    }

    /**
     * Initialize A/B testing
     */
    async init() {
        // Load active tests from data attributes
        this.loadTestsFromDOM();

        // Apply variants for all active tests
        for (const [testId, testConfig] of this.activeTests) {
            await this.applyTest(testId, testConfig);
        }

        // Set up automatic conversion tracking if enabled
        if (this.autoTrack) {
            this.setupAutoTracking();
        }
    }

    /**
     * Load test configurations from DOM data attributes
     */
    loadTestsFromDOM() {
        const testElements = document.querySelectorAll('[data-ab-test]');

        testElements.forEach(element => {
            const testId = element.dataset.abTest;
            const elementType = element.dataset.abElement || 'text';
            const conversionEvent = element.dataset.abConversion || 'click';

            this.activeTests.set(testId, {
                element,
                elementType,
                conversionEvent
            });
        });

        if (this.debug) {
            console.log('Loaded A/B tests:', Array.from(this.activeTests.keys()));
        }
    }

    /**
     * Apply A/B test variant to an element
     */
    async applyTest(testId, testConfig) {
        try {
            // Get variant assignment from server
            const variant = await this.getVariant(testId);

            if (this.debug) {
                console.log(`Test ${testId}: Assigned to variant ${variant.variant}`, variant.config);
            }

            // Apply variant configuration to element
            this.applyVariantToElement(testConfig.element, variant, testConfig.elementType);

            // Store variant for tracking
            testConfig.variant = variant.variant;

            // Set up conversion tracking for this element
            if (testConfig.conversionEvent) {
                this.trackElementConversion(testId, testConfig);
            }

        } catch (error) {
            console.error(`Failed to apply A/B test ${testId}:`, error);
        }
    }

    /**
     * Get variant assignment from server
     */
    async getVariant(testId) {
        const response = await fetch(`${this.apiBase}/public/ab-test/${testId}/variant`, {
            method: 'GET',
            credentials: 'include' // Include cookies
        });

        if (!response.ok) {
            throw new Error(`Failed to get variant: ${response.statusText}`);
        }

        return await response.json();
    }

    /**
     * Apply variant configuration to an element
     */
    applyVariantToElement(element, variant, elementType) {
        const config = variant.config || {};

        switch (elementType) {
            case 'text':
                if (config.text) {
                    element.textContent = config.text;
                }
                break;

            case 'html':
                if (config.html) {
                    element.innerHTML = config.html;
                }
                break;

            case 'button':
                if (config.text) {
                    element.textContent = config.text;
                }
                if (config.color) {
                    element.style.backgroundColor = config.color;
                }
                if (config.class) {
                    element.className = config.class;
                }
                break;

            case 'image':
                if (config.src && element.tagName === 'IMG') {
                    element.src = config.src;
                }
                if (config.alt) {
                    element.alt = config.alt;
                }
                break;

            case 'style':
                if (config.styles) {
                    Object.assign(element.style, config.styles);
                }
                break;

            case 'custom':
                // Allow custom handler via event
                const event = new CustomEvent('ab-variant-apply', {
                    detail: { variant: variant.variant, config }
                });
                element.dispatchEvent(event);
                break;

            default:
                console.warn(`Unknown element type: ${elementType}`);
        }

        // Add data attribute to indicate variant
        element.dataset.abVariant = variant.variant;

        // Add variant-specific class
        element.classList.add(`ab-variant-${variant.variant.toLowerCase()}`);
    }

    /**
     * Track conversion for a specific test
     */
    async trackConversion(testId) {
        // Prevent duplicate tracking
        const conversionKey = `${testId}-${Date.now()}`;
        if (this.conversionEvents.has(conversionKey)) {
            return;
        }
        this.conversionEvents.add(conversionKey);

        try {
            const response = await fetch(`${this.apiBase}/public/ab-test/${testId}/conversion`, {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error(`Failed to track conversion: ${response.statusText}`);
            }

            if (this.debug) {
                console.log(`Conversion tracked for test ${testId}`);
            }

            // Fire custom event for analytics
            window.dispatchEvent(new CustomEvent('ab-conversion', {
                detail: { testId }
            }));

        } catch (error) {
            console.error(`Failed to track conversion for test ${testId}:`, error);
        }
    }

    /**
     * Set up automatic conversion tracking for an element
     */
    trackElementConversion(testId, testConfig) {
        const { element, conversionEvent } = testConfig;

        const handler = () => {
            this.trackConversion(testId);
        };

        // Remove any existing handler to prevent duplicates
        element.removeEventListener(conversionEvent, handler);

        // Add new handler
        element.addEventListener(conversionEvent, handler);

        // Store handler for cleanup
        testConfig.conversionHandler = handler;
    }

    /**
     * Set up automatic tracking for common conversion events
     */
    setupAutoTracking() {
        // Track CTA button clicks
        document.addEventListener('click', (e) => {
            const target = e.target.closest('[data-ab-track="cta"]');
            if (target) {
                const testId = target.dataset.abTest;
                if (testId) {
                    this.trackConversion(testId);
                }
            }
        });

        // Track form submissions
        document.addEventListener('submit', (e) => {
            const form = e.target;
            if (form.dataset.abTrack === 'form') {
                const testId = form.dataset.abTest;
                if (testId) {
                    this.trackConversion(testId);
                }
            }
        });

        // Track signup/purchase completion via custom events
        window.addEventListener('purchase-complete', (e) => {
            // Track all active tests that have purchase as conversion
            for (const [testId, config] of this.activeTests) {
                if (config.conversionEvent === 'purchase') {
                    this.trackConversion(testId);
                }
            }
        });
    }

    /**
     * Manually apply a test to an element
     */
    async applyManualTest(testId, element, config = {}) {
        const testConfig = {
            element,
            elementType: config.elementType || 'text',
            conversionEvent: config.conversionEvent || 'click'
        };

        this.activeTests.set(testId, testConfig);
        await this.applyTest(testId, testConfig);
    }

    /**
     * Get all active tests
     */
    getActiveTests() {
        return Array.from(this.activeTests.entries()).map(([id, config]) => ({
            id,
            variant: config.variant,
            element: config.element
        }));
    }

    /**
     * Clean up event listeners
     */
    destroy() {
        // Remove conversion handlers
        for (const [testId, config] of this.activeTests) {
            if (config.conversionHandler && config.element) {
                config.element.removeEventListener(
                    config.conversionEvent,
                    config.conversionHandler
                );
            }
        }

        this.activeTests.clear();
        this.conversionEvents.clear();
    }
}

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    // Auto-initialize if configured
    if (window.ABTestConfig) {
        window.abTest = new ABTestClient(window.ABTestConfig);
    } else {
        // Initialize with defaults
        window.abTest = new ABTestClient({
            debug: window.location.hostname === 'localhost'
        });
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ABTestClient;
}

/**
 * Usage Examples:
 *
 * HTML:
 * <h1 data-ab-test="1" data-ab-element="text" data-ab-conversion="view">
 *   Original Headline
 * </h1>
 *
 * <button data-ab-test="2" data-ab-element="button" data-ab-track="cta">
 *   Get Started
 * </button>
 *
 * JavaScript:
 * // Manual test application
 * abTest.applyManualTest('3', document.getElementById('pricing'), {
 *   elementType: 'custom',
 *   conversionEvent: 'purchase'
 * });
 *
 * // Manual conversion tracking
 * document.getElementById('checkout-btn').addEventListener('click', () => {
 *   abTest.trackConversion('3');
 * });
 *
 * // Listen for variant application
 * document.getElementById('hero').addEventListener('ab-variant-apply', (e) => {
 *   console.log('Variant applied:', e.detail);
 *   // Custom variant handling
 * });
 */