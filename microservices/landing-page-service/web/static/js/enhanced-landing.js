// Enhanced Landing Page JavaScript with Modern Features
class ModernLandingPage {
    constructor() {
        this.init();
        this.setupScrollEffects();
        this.setupIntersectionObserver();
        this.setupParallax();
        this.setupTypingEffect();
        this.setupScrollProgress();
        this.setupPerformanceOptimizations();
        this.setupWebVitalsTracking();
        this.setupAnalyticsTracking();
    }

    init() {
        // Enhanced page loading animation
        this.setupPageLoad();

        // Initialize existing functionality
        this.setupMobileNavigation();
        this.setupSmoothScrolling();
        this.setupFAQAccordion();
        this.setupPricingToggle();
        this.setupNavbarEffects();
        this.setupContactForm();
    }

    setupPageLoad() {
        document.body.style.opacity = '0';
        document.body.style.transform = 'translateY(20px)';

        window.addEventListener('load', () => {
            document.body.style.transition = 'opacity 0.8s cubic-bezier(0.4, 0, 0.2, 1), transform 0.8s cubic-bezier(0.4, 0, 0.2, 1)';
            document.body.style.opacity = '1';
            document.body.style.transform = 'translateY(0)';
        });
    }

    setupMobileNavigation() {
        const navToggle = document.getElementById('nav-toggle');
        const navMenu = document.getElementById('nav-menu');

        if (navToggle && navMenu) {
            navToggle.addEventListener('click', () => {
                const isActive = navToggle.classList.contains('active');

                navToggle.classList.toggle('active');
                navMenu.classList.toggle('active');

                // Prevent body scroll when menu is open
                document.body.style.overflow = isActive ? 'auto' : 'hidden';

                // Accessibility
                navToggle.setAttribute('aria-expanded', !isActive);
            });

            // Close mobile menu when clicking on a link
            const navLinks = navMenu.querySelectorAll('.nav-link');
            navLinks.forEach(link => {
                link.addEventListener('click', () => {
                    navToggle.classList.remove('active');
                    navMenu.classList.remove('active');
                    document.body.style.overflow = 'auto';
                    navToggle.setAttribute('aria-expanded', false);
                });
            });

            // Close menu when clicking outside
            document.addEventListener('click', (e) => {
                if (!navToggle.contains(e.target) && !navMenu.contains(e.target)) {
                    navToggle.classList.remove('active');
                    navMenu.classList.remove('active');
                    document.body.style.overflow = 'auto';
                }
            });
        }
    }

    setupSmoothScrolling() {
        const anchorLinks = document.querySelectorAll('a[href^="#"]');
        anchorLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = link.getAttribute('href');
                const targetElement = document.querySelector(targetId);

                if (targetElement) {
                    const offsetTop = targetElement.offsetTop - 80;

                    window.scrollTo({
                        top: offsetTop,
                        behavior: 'smooth'
                    });

                    // Update URL without triggering scroll
                    history.pushState(null, null, targetId);
                }
            });
        });
    }

    setupFAQAccordion() {
        const faqItems = document.querySelectorAll('.faq-item');
        faqItems.forEach((item, index) => {
            const question = item.querySelector('.faq-question');
            const answer = item.querySelector('.faq-answer');

            // Add ARIA attributes
            question.setAttribute('aria-expanded', false);
            question.setAttribute('aria-controls', `faq-answer-${index}`);
            answer.setAttribute('id', `faq-answer-${index}`);

            question.addEventListener('click', () => {
                const isActive = item.classList.contains('active');

                // Close other FAQ items with animation
                faqItems.forEach(otherItem => {
                    if (otherItem !== item && otherItem.classList.contains('active')) {
                        otherItem.classList.remove('active');
                        const otherQuestion = otherItem.querySelector('.faq-question');
                        const otherAnswer = otherItem.querySelector('.faq-answer');
                        otherQuestion.setAttribute('aria-expanded', false);
                        otherAnswer.style.maxHeight = '0';
                    }
                });

                // Toggle current FAQ item
                item.classList.toggle('active');
                question.setAttribute('aria-expanded', !isActive);

                if (!isActive) {
                    answer.style.maxHeight = answer.scrollHeight + 'px';
                } else {
                    answer.style.maxHeight = '0';
                }
            });
        });
    }

    setupPricingToggle() {
        const pricingToggle = document.getElementById('pricing-toggle');
        const pricingCards = document.querySelectorAll('.pricing-card');

        if (pricingToggle) {
            pricingToggle.addEventListener('change', () => {
                const isYearly = pricingToggle.checked;

                pricingCards.forEach(card => {
                    const priceAmount = card.querySelector('.price-amount');
                    const pricePeriod = card.querySelector('.price-period');

                    if (priceAmount && pricePeriod) {
                        // Add transition class
                        card.classList.add('pricing-transition');

                        setTimeout(() => {
                            const monthlyPrice = parseFloat(priceAmount.dataset.monthly || priceAmount.textContent);

                            if (isYearly) {
                                const yearlyPrice = Math.round(monthlyPrice * 12 * 0.8);
                                priceAmount.textContent = yearlyPrice;
                                pricePeriod.textContent = '/year';
                            } else {
                                priceAmount.textContent = monthlyPrice;
                                pricePeriod.textContent = '/month';
                            }

                            card.classList.remove('pricing-transition');
                        }, 150);
                    }
                });
            });
        }
    }

    setupNavbarEffects() {
        const navbar = document.querySelector('.navbar');
        let lastScrollTop = 0;
        let isScrolling = false;

        const handleScroll = () => {
            const scrollTop = window.pageYOffset || document.documentElement.scrollTop;

            // Glass morphism effect
            if (scrollTop > 50) {
                navbar.classList.add('glass-nav');
            } else {
                navbar.classList.remove('glass-nav');
            }

            // Hide/show navbar on scroll
            if (scrollTop > lastScrollTop && scrollTop > 200) {
                navbar.style.transform = 'translateY(-100%)';
            } else {
                navbar.style.transform = 'translateY(0)';
            }

            lastScrollTop = scrollTop;
            isScrolling = false;
        };

        window.addEventListener('scroll', () => {
            if (!isScrolling) {
                requestAnimationFrame(handleScroll);
                isScrolling = true;
            }
        }, { passive: true });
    }

    setupScrollEffects() {
        // Parallax effect for hero section
        const hero = document.querySelector('.hero');
        const heroContent = document.querySelector('.hero-content');

        if (hero && heroContent) {
            window.addEventListener('scroll', () => {
                const scrolled = window.pageYOffset;
                const rate = scrolled * -0.5;

                if (scrolled < hero.offsetHeight) {
                    heroContent.style.transform = `translateY(${rate}px)`;
                }
            }, { passive: true });
        }
    }

    setupIntersectionObserver() {
        const observerOptions = {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('animate-in');

                    // Stagger animations for grid items
                    if (entry.target.closest('.features-grid, .pricing-grid, .testimonials-grid')) {
                        const siblings = Array.from(entry.target.parentElement.children);
                        const index = siblings.indexOf(entry.target);
                        entry.target.style.animationDelay = `${index * 0.1}s`;
                    }
                }
            });
        }, observerOptions);

        // Observe elements for animation
        const animatedElements = document.querySelectorAll(
            '.feature-card, .pricing-card, .testimonial-card, .faq-item, .hero-stats .stat'
        );

        animatedElements.forEach(el => {
            el.classList.add('animate-on-scroll');
            observer.observe(el);
        });
    }

    setupParallax() {
        const parallaxElements = document.querySelectorAll('[data-parallax]');

        if (parallaxElements.length > 0) {
            window.addEventListener('scroll', () => {
                const scrollTop = window.pageYOffset;

                parallaxElements.forEach(el => {
                    const speed = el.dataset.parallax || 0.5;
                    const yPos = -(scrollTop * speed);
                    el.style.transform = `translateY(${yPos}px)`;
                });
            }, { passive: true });
        }
    }

    setupTypingEffect() {
        const typingElement = document.querySelector('[data-typing]');

        if (typingElement) {
            const text = typingElement.dataset.typing;
            const speed = parseInt(typingElement.dataset.speed) || 100;

            typingElement.textContent = '';

            let i = 0;
            const typeWriter = () => {
                if (i < text.length) {
                    typingElement.textContent += text.charAt(i);
                    i++;
                    setTimeout(typeWriter, speed);
                }
            };

            // Start typing when element comes into view
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        setTimeout(typeWriter, 500);
                        observer.unobserve(entry.target);
                    }
                });
            });

            observer.observe(typingElement);
        }
    }

    setupScrollProgress() {
        const progressBar = document.createElement('div');
        progressBar.className = 'scroll-progress';
        document.body.appendChild(progressBar);

        window.addEventListener('scroll', () => {
            const scrollable = document.documentElement.scrollHeight - window.innerHeight;
            const scrolled = window.scrollY;
            const progress = (scrolled / scrollable) * 100;

            progressBar.style.width = `${Math.min(progress, 100)}%`;
        }, { passive: true });
    }

    setupContactForm() {
        const contactForm = document.getElementById('contact-form');

        if (contactForm) {
            contactForm.addEventListener('submit', async (e) => {
                e.preventDefault();

                const formData = new FormData(contactForm);
                const submitButton = contactForm.querySelector('button[type="submit"]');
                const originalText = submitButton.textContent;

                // Show loading state
                submitButton.textContent = 'Sending...';
                submitButton.disabled = true;
                submitButton.classList.add('loading');

                try {
                    // Simulate API call (replace with actual endpoint)
                    await this.simulateFormSubmission(formData);

                    // Show success message
                    this.showNotification('Message sent successfully!', 'success');
                    contactForm.reset();

                } catch (error) {
                    this.showNotification('Failed to send message. Please try again.', 'error');
                } finally {
                    submitButton.textContent = originalText;
                    submitButton.disabled = false;
                    submitButton.classList.remove('loading');
                }
            });

            // Real-time validation
            const inputs = contactForm.querySelectorAll('input, textarea');
            inputs.forEach(input => {
                input.addEventListener('blur', () => {
                    this.validateInput(input);
                });
            });
        }
    }

    setupPerformanceOptimizations() {
        // Lazy load images
        const images = document.querySelectorAll('img[data-src]');
        const imageObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;
                    img.src = img.dataset.src;
                    img.classList.remove('lazy');
                    imageObserver.unobserve(img);
                }
            });
        });

        images.forEach(img => imageObserver.observe(img));

        // Preload critical resources
        this.preloadCriticalResources();
    }

    simulateFormSubmission(formData) {
        return new Promise((resolve, reject) => {
            setTimeout(() => {
                // Simulate random success/failure for demo
                Math.random() > 0.1 ? resolve() : reject();
            }, 2000);
        });
    }

    validateInput(input) {
        const value = input.value.trim();
        let isValid = true;
        let message = '';

        switch (input.type) {
            case 'email':
                isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
                message = 'Please enter a valid email address';
                break;
            case 'tel':
                isValid = /^[\+]?[1-9][\d]{0,15}$/.test(value);
                message = 'Please enter a valid phone number';
                break;
            default:
                isValid = value.length >= 2;
                message = 'This field is required';
        }

        this.updateInputValidation(input, isValid, message);
        return isValid;
    }

    updateInputValidation(input, isValid, message) {
        const errorElement = input.parentElement.querySelector('.error-message');

        if (isValid) {
            input.classList.remove('error');
            if (errorElement) errorElement.remove();
        } else {
            input.classList.add('error');
            if (!errorElement) {
                const error = document.createElement('span');
                error.className = 'error-message';
                error.textContent = message;
                input.parentElement.appendChild(error);
            }
        }
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;

        document.body.appendChild(notification);

        // Animate in
        requestAnimationFrame(() => {
            notification.classList.add('show');
        });

        // Auto remove
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.parentElement.removeChild(notification);
                }
            }, 300);
        }, 5000);
    }

    preloadCriticalResources() {
        const criticalResources = [
            '/static/css/animations.css',
            '/static/js/dark-mode.js',
            '/static/js/keyboard-navigation.js'
        ];

        criticalResources.forEach(resource => {
            const link = document.createElement('link');
            link.rel = 'preload';
            link.as = resource.endsWith('.css') ? 'style' : 'script';
            link.href = resource;
            document.head.appendChild(link);
        });
    }

    setupWebVitalsTracking() {
        // Track Core Web Vitals
        this.trackLCP();
        this.trackFID();
        this.trackCLS();
        this.trackFCP();
        this.trackTTI();
    }

    trackLCP() {
        // Largest Contentful Paint
        const observer = new PerformanceObserver((list) => {
            const entries = list.getEntries();
            const lastEntry = entries[entries.length - 1];

            const lcp = lastEntry.startTime;
            this.sendWebVital('lcp', lcp);
        });

        observer.observe({ type: 'largest-contentful-paint', buffered: true });
    }

    trackFID() {
        // First Input Delay
        const observer = new PerformanceObserver((list) => {
            const firstEntry = list.getEntries()[0];
            const fid = firstEntry.processingStart - firstEntry.startTime;
            this.sendWebVital('fid', fid);
        });

        observer.observe({ type: 'first-input', buffered: true });
    }

    trackCLS() {
        // Cumulative Layout Shift
        let clsValue = 0;
        let clsEntries = [];
        let sessionValue = 0;
        let sessionEntries = [];

        const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
                if (!entry.hadRecentInput) {
                    const firstSessionEntry = sessionEntries[0];
                    const lastSessionEntry = sessionEntries[sessionEntries.length - 1];

                    if (sessionValue &&
                        entry.startTime - lastSessionEntry.startTime < 1000 &&
                        entry.startTime - firstSessionEntry.startTime < 5000) {
                        sessionValue += entry.value;
                        sessionEntries.push(entry);
                    } else {
                        sessionValue = entry.value;
                        sessionEntries = [entry];
                    }

                    if (sessionValue > clsValue) {
                        clsValue = sessionValue;
                        clsEntries = [...sessionEntries];
                        this.sendWebVital('cls', clsValue);
                    }
                }
            }
        });

        observer.observe({ type: 'layout-shift', buffered: true });
    }

    trackFCP() {
        // First Contentful Paint
        const observer = new PerformanceObserver((list) => {
            const firstEntry = list.getEntries()[0];
            const fcp = firstEntry.startTime;
            this.sendWebVital('fcp', fcp);
        });

        observer.observe({ type: 'paint', buffered: true });
    }

    trackTTI() {
        // Time to Interactive (simplified implementation)
        window.addEventListener('load', () => {
            setTimeout(() => {
                const navigation = performance.getEntriesByType('navigation')[0];
                const tti = navigation.loadEventEnd;
                this.sendWebVital('tti', tti);
            }, 0);
        });
    }

    sendWebVital(name, value) {
        // Send Web Vital data to SEO endpoint
        fetch('/api/v1/seo/web-vitals', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                [name]: value,
                url: window.location.href,
                timestamp: new Date().toISOString()
            })
        }).catch(error => {
            console.warn('Failed to track Web Vital:', name, error);
        });
    }

    setupAnalyticsTracking() {
        // Track page views
        this.trackPageView();

        // Track scroll depth
        this.trackScrollDepth();

        // Track click events
        this.trackClicks();

        // Track form interactions
        this.trackFormInteractions();
    }

    trackPageView() {
        // Send page view analytics
        fetch('/api/v1/seo/analytics/pageview', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                url: window.location.href,
                title: document.title,
                timestamp: new Date().toISOString(),
                referrer: document.referrer,
                userAgent: navigator.userAgent
            })
        }).catch(error => {
            console.warn('Failed to track page view:', error);
        });
    }

    trackScrollDepth() {
        let maxScroll = 0;
        let scrollMilestones = [25, 50, 75, 100];
        let trackedMilestones = new Set();

        window.addEventListener('scroll', () => {
            const scrollTop = window.pageYOffset;
            const docHeight = document.documentElement.scrollHeight - window.innerHeight;
            const scrollPercent = Math.round((scrollTop / docHeight) * 100);

            maxScroll = Math.max(maxScroll, scrollPercent);

            scrollMilestones.forEach(milestone => {
                if (scrollPercent >= milestone && !trackedMilestones.has(milestone)) {
                    trackedMilestones.add(milestone);
                    this.trackEvent('scroll_depth', {
                        depth: milestone,
                        url: window.location.href
                    });
                }
            });
        }, { passive: true });
    }

    trackClicks() {
        // Track CTA clicks
        document.addEventListener('click', (e) => {
            const target = e.target.closest('a, button');
            if (!target) return;

            const isExternalLink = target.href && !target.href.startsWith(window.location.origin);
            const isCtaButton = target.classList.contains('btn-primary') ||
                              target.classList.contains('cta-button') ||
                              target.closest('.pricing-card');

            if (isCtaButton || isExternalLink) {
                this.trackEvent('click', {
                    element: target.tagName.toLowerCase(),
                    text: target.textContent.trim(),
                    href: target.href || '',
                    classes: target.className,
                    external: isExternalLink,
                    cta: isCtaButton
                });
            }
        });
    }

    trackFormInteractions() {
        // Track form submissions
        document.addEventListener('submit', (e) => {
            const form = e.target;
            if (form.tagName === 'FORM') {
                this.trackEvent('form_submit', {
                    form_id: form.id || 'unknown',
                    form_action: form.action || window.location.href,
                    fields_count: form.elements.length
                });
            }
        });

        // Track form field focus
        document.addEventListener('focus', (e) => {
            if (e.target.matches('input, textarea, select')) {
                this.trackEvent('form_field_focus', {
                    field_type: e.target.type || e.target.tagName.toLowerCase(),
                    field_name: e.target.name || e.target.id || 'unknown'
                });
            }
        }, true);
    }

    trackEvent(eventName, data) {
        fetch('/api/v1/seo/analytics/event', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                event: eventName,
                data: data,
                timestamp: new Date().toISOString(),
                url: window.location.href
            })
        }).catch(error => {
            console.warn('Failed to track event:', eventName, error);
        });
    }
}

// Initialize enhanced landing page
document.addEventListener('DOMContentLoaded', () => {
    window.modernLandingPage = new ModernLandingPage();
});

// Additional CSS for enhanced functionality
const enhancedStyles = `
    .pricing-transition {
        opacity: 0.7;
        transform: scale(0.98);
    }

    .glass-nav {
        background: rgba(255, 255, 255, 0.95) !important;
        backdrop-filter: blur(20px);
        border-bottom: 1px solid rgba(255, 255, 255, 0.2);
    }

    .navbar {
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }

    .notification {
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 16px 24px;
        border-radius: 8px;
        color: white;
        font-weight: 500;
        transform: translateX(400px);
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        z-index: 1000;
    }

    .notification.show {
        transform: translateX(0);
    }

    .notification-success {
        background: var(--success-color);
    }

    .notification-error {
        background: var(--danger-color);
    }

    .notification-info {
        background: var(--primary-color);
    }

    .error-message {
        display: block;
        color: var(--danger-color);
        font-size: 0.875rem;
        margin-top: 4px;
    }

    input.error,
    textarea.error {
        border-color: var(--danger-color);
        box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
    }

    .loading {
        position: relative;
        pointer-events: none;
    }

    .loading::after {
        content: '';
        position: absolute;
        width: 16px;
        height: 16px;
        margin: auto;
        border: 2px solid transparent;
        border-top-color: currentColor;
        border-radius: 50%;
        animation: rotate 1s linear infinite;
        top: 0;
        left: 0;
        bottom: 0;
        right: 0;
    }
`;

// Inject enhanced styles
const styleSheet = document.createElement('style');
styleSheet.textContent = enhancedStyles;
document.head.appendChild(styleSheet);