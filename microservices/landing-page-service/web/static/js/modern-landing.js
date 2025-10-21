/**
 * Modern Landing Page JavaScript
 * Performance-optimized with lazy loading, intersection observer, and smooth interactions
 */

// ==========================================
// Performance Monitoring
// ==========================================
class PerformanceMonitor {
    constructor() {
        this.metrics = {};
        this.init();
    }

    init() {
        // Track Core Web Vitals
        if ('PerformanceObserver' in window) {
            // Largest Contentful Paint
            try {
                const lcpObserver = new PerformanceObserver((list) => {
                    const entries = list.getEntries();
                    const lastEntry = entries[entries.length - 1];
                    this.metrics.lcp = lastEntry.renderTime || lastEntry.loadTime;
                    this.reportMetric('LCP', this.metrics.lcp);
                });
                lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });
            } catch (e) {
                console.log('LCP observer not supported');
            }

            // First Input Delay
            try {
                const fidObserver = new PerformanceObserver((list) => {
                    const entries = list.getEntries();
                    entries.forEach((entry) => {
                        this.metrics.fid = entry.processingStart - entry.startTime;
                        this.reportMetric('FID', this.metrics.fid);
                    });
                });
                fidObserver.observe({ type: 'first-input', buffered: true });
            } catch (e) {
                console.log('FID observer not supported');
            }

            // Cumulative Layout Shift
            try {
                let clsValue = 0;
                let clsEntries = [];

                const clsObserver = new PerformanceObserver((list) => {
                    for (const entry of list.getEntries()) {
                        if (!entry.hadRecentInput) {
                            const firstSessionEntry = clsEntries[0];
                            const lastSessionEntry = clsEntries[clsEntries.length - 1];

                            if (entry.startTime - lastSessionEntry.startTime < 1000 &&
                                entry.startTime - firstSessionEntry.startTime < 5000) {
                                clsEntries.push(entry);
                                clsValue += entry.value;
                            } else {
                                clsEntries = [entry];
                                clsValue = entry.value;
                            }
                        }
                    }
                    this.metrics.cls = clsValue;
                    this.reportMetric('CLS', this.metrics.cls);
                });
                clsObserver.observe({ type: 'layout-shift', buffered: true });
            } catch (e) {
                console.log('CLS observer not supported');
            }
        }

        // Track page load time
        window.addEventListener('load', () => {
            const loadTime = performance.timing.loadEventEnd - performance.timing.navigationStart;
            this.metrics.loadTime = loadTime;
            this.reportMetric('Page Load Time', loadTime);
        });
    }

    reportMetric(name, value) {
        // Send to analytics endpoint
        if (window.location.hostname !== 'localhost') {
            fetch('/api/v1/analytics/metrics', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    metric: name,
                    value: value,
                    url: window.location.href,
                    timestamp: new Date().toISOString()
                })
            }).catch(() => {
                // Silently fail
            });
        }
        console.log(`Performance: ${name} = ${value}`);
    }
}

// ==========================================
// Lazy Loading with Intersection Observer
// ==========================================
class LazyLoader {
    constructor() {
        this.imageObserver = null;
        this.sectionObserver = null;
        this.init();
    }

    init() {
        // Lazy load images
        this.setupImageLazyLoading();

        // Animate sections on scroll
        this.setupSectionAnimation();

        // Lazy load iframes
        this.setupIframeLazyLoading();
    }

    setupImageLazyLoading() {
        const imageOptions = {
            root: null,
            rootMargin: '50px',
            threshold: 0.01
        };

        this.imageObserver = new IntersectionObserver((entries, observer) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const img = entry.target;

                    // Load image
                    if (img.dataset.src) {
                        // Create new image to preload
                        const newImg = new Image();
                        newImg.onload = () => {
                            img.src = newImg.src;
                            img.classList.add('loaded');
                        };
                        newImg.src = img.dataset.src;

                        // Load srcset if available
                        if (img.dataset.srcset) {
                            img.srcset = img.dataset.srcset;
                        }
                    }

                    // Load background image
                    if (img.dataset.bg) {
                        img.style.backgroundImage = `url(${img.dataset.bg})`;
                        img.classList.add('loaded');
                    }

                    observer.unobserve(img);
                }
            });
        }, imageOptions);

        // Observe all lazy images
        document.querySelectorAll('[data-src], [data-bg]').forEach(img => {
            this.imageObserver.observe(img);
        });
    }

    setupSectionAnimation() {
        const sectionOptions = {
            root: null,
            rootMargin: '-100px',
            threshold: 0.1
        };

        this.sectionObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('aos-animate');
                }
            });
        }, sectionOptions);

        // Observe all animated sections
        document.querySelectorAll('[data-aos]').forEach(section => {
            this.sectionObserver.observe(section);
        });
    }

    setupIframeLazyLoading() {
        const iframes = document.querySelectorAll('iframe[data-src]');

        if ('IntersectionObserver' in window) {
            const iframeObserver = new IntersectionObserver((entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const iframe = entry.target;
                        iframe.src = iframe.dataset.src;
                        observer.unobserve(iframe);
                    }
                });
            });

            iframes.forEach(iframe => {
                iframeObserver.observe(iframe);
            });
        } else {
            // Fallback for browsers without IntersectionObserver
            iframes.forEach(iframe => {
                iframe.src = iframe.dataset.src;
            });
        }
    }
}

// ==========================================
// Smooth Scroll and Navigation
// ==========================================
class SmoothNavigation {
    constructor() {
        this.navbar = document.getElementById('navbar');
        this.navToggle = document.getElementById('navToggle');
        this.navMenu = document.getElementById('navMenu');
        this.lastScrollY = 0;
        this.init();
    }

    init() {
        // Smooth scroll for anchor links
        this.setupSmoothScroll();

        // Sticky nav with hide/show on scroll
        this.setupStickyNav();

        // Mobile menu toggle
        this.setupMobileMenu();

        // Active nav highlighting
        this.setupActiveNav();
    }

    setupSmoothScroll() {
        document.querySelectorAll('a[href^="#"]').forEach(anchor => {
            anchor.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = anchor.getAttribute('href');
                if (targetId === '#') return;

                const target = document.querySelector(targetId);
                if (target) {
                    const navHeight = this.navbar.offsetHeight;
                    const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - navHeight;

                    window.scrollTo({
                        top: targetPosition,
                        behavior: 'smooth'
                    });

                    // Close mobile menu if open
                    if (this.navMenu.classList.contains('active')) {
                        this.toggleMobileMenu();
                    }
                }
            });
        });
    }

    setupStickyNav() {
        let ticking = false;

        const updateNav = () => {
            const scrollY = window.scrollY;

            if (scrollY > 100) {
                this.navbar.classList.add('scrolled');
            } else {
                this.navbar.classList.remove('scrolled');
            }

            // Hide/show on scroll
            if (scrollY > this.lastScrollY && scrollY > 500) {
                this.navbar.classList.add('nav-hidden');
            } else {
                this.navbar.classList.remove('nav-hidden');
            }

            this.lastScrollY = scrollY;
            ticking = false;
        };

        window.addEventListener('scroll', () => {
            if (!ticking) {
                requestAnimationFrame(updateNav);
                ticking = true;
            }
        });
    }

    setupMobileMenu() {
        this.navToggle.addEventListener('click', () => {
            this.toggleMobileMenu();
        });

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.navbar.contains(e.target) && this.navMenu.classList.contains('active')) {
                this.toggleMobileMenu();
            }
        });
    }

    toggleMobileMenu() {
        this.navMenu.classList.toggle('active');
        this.navToggle.classList.toggle('active');
        document.body.classList.toggle('menu-open');
    }

    setupActiveNav() {
        const sections = document.querySelectorAll('section[id]');

        const observerOptions = {
            root: null,
            rootMargin: '-50% 0px -50% 0px',
            threshold: 0
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const id = entry.target.getAttribute('id');
                    this.updateActiveNavLink(id);
                }
            });
        }, observerOptions);

        sections.forEach(section => {
            observer.observe(section);
        });
    }

    updateActiveNavLink(sectionId) {
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('href') === `#${sectionId}`) {
                link.classList.add('active');
            }
        });
    }
}

// ==========================================
// Pricing Toggle
// ==========================================
class PricingToggle {
    constructor() {
        this.toggle = document.getElementById('pricingToggle');
        this.init();
    }

    init() {
        if (!this.toggle) return;

        this.toggle.addEventListener('change', () => {
            this.updatePrices();
        });
    }

    updatePrices() {
        const isYearly = this.toggle.checked;
        const priceElements = document.querySelectorAll('[data-monthly][data-yearly]');

        priceElements.forEach(element => {
            const monthlyPrice = element.dataset.monthly;
            const yearlyPrice = element.dataset.yearly;

            // Animate price change
            element.style.opacity = '0';
            setTimeout(() => {
                element.textContent = isYearly ? yearlyPrice : monthlyPrice;
                element.style.opacity = '1';
            }, 150);
        });

        // Update period text
        document.querySelectorAll('.period').forEach(period => {
            period.textContent = isYearly ? '/year' : '/month';
        });
    }
}

// ==========================================
// FAQ Accordion
// ==========================================
class FAQAccordion {
    constructor() {
        this.faqItems = document.querySelectorAll('.faq-item');
        this.init();
    }

    init() {
        this.faqItems.forEach(item => {
            const question = item.querySelector('.faq-question');

            question.addEventListener('click', () => {
                this.toggleFAQ(item);
            });
        });
    }

    toggleFAQ(item) {
        const isActive = item.classList.contains('active');

        // Close all other FAQs
        this.faqItems.forEach(faq => {
            faq.classList.remove('active');
        });

        // Toggle current FAQ
        if (!isActive) {
            item.classList.add('active');
        }
    }
}

// ==========================================
// Testimonial Carousel
// ==========================================
class TestimonialCarousel {
    constructor() {
        this.carousel = document.querySelector('.testimonials-carousel');
        this.prevBtn = document.querySelector('.carousel-prev');
        this.nextBtn = document.querySelector('.carousel-next');
        this.currentIndex = 0;
        this.autoplayInterval = null;
        this.init();
    }

    init() {
        if (!this.carousel) return;

        this.cards = this.carousel.querySelectorAll('.testimonial-card');
        this.totalCards = this.cards.length;

        // Setup controls
        this.prevBtn?.addEventListener('click', () => this.prev());
        this.nextBtn?.addEventListener('click', () => this.next());

        // Setup touch/swipe support
        this.setupTouchSupport();

        // Start autoplay
        this.startAutoplay();

        // Pause on hover
        this.carousel.addEventListener('mouseenter', () => this.stopAutoplay());
        this.carousel.addEventListener('mouseleave', () => this.startAutoplay());
    }

    prev() {
        this.currentIndex = (this.currentIndex - 1 + this.totalCards) % this.totalCards;
        this.updateCarousel();
    }

    next() {
        this.currentIndex = (this.currentIndex + 1) % this.totalCards;
        this.updateCarousel();
    }

    updateCarousel() {
        const cardWidth = this.cards[0].offsetWidth + 24; // Including gap
        const offset = -this.currentIndex * cardWidth;
        this.carousel.style.transform = `translateX(${offset}px)`;
    }

    startAutoplay() {
        this.autoplayInterval = setInterval(() => {
            this.next();
        }, 5000);
    }

    stopAutoplay() {
        clearInterval(this.autoplayInterval);
    }

    setupTouchSupport() {
        let startX = 0;
        let currentX = 0;
        let isDragging = false;

        this.carousel.addEventListener('touchstart', (e) => {
            startX = e.touches[0].clientX;
            isDragging = true;
        });

        this.carousel.addEventListener('touchmove', (e) => {
            if (!isDragging) return;
            e.preventDefault();
            currentX = e.touches[0].clientX;
        });

        this.carousel.addEventListener('touchend', () => {
            if (!isDragging) return;
            isDragging = false;

            const diff = startX - currentX;
            if (Math.abs(diff) > 50) {
                if (diff > 0) {
                    this.next();
                } else {
                    this.prev();
                }
            }
        });
    }
}

// ==========================================
// Announcement Bar
// ==========================================
class AnnouncementBar {
    constructor() {
        this.bar = document.getElementById('announcementBar');
        this.init();
    }

    init() {
        // Check if already closed in this session
        if (sessionStorage.getItem('announcementClosed')) {
            this.bar?.remove();
        }
    }
}

function closeAnnouncement() {
    const bar = document.getElementById('announcementBar');
    if (bar) {
        bar.style.animation = 'slideUp 0.3s ease';
        setTimeout(() => {
            bar.remove();
            sessionStorage.setItem('announcementClosed', 'true');
        }, 300);
    }
}

// ==========================================
// Form Handling
// ==========================================
class FormHandler {
    constructor() {
        this.forms = document.querySelectorAll('form[data-ajax]');
        this.init();
    }

    init() {
        this.forms.forEach(form => {
            form.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleSubmit(form);
            });
        });
    }

    async handleSubmit(form) {
        const formData = new FormData(form);
        const data = Object.fromEntries(formData);
        const action = form.action;
        const method = form.method || 'POST';

        // Show loading state
        const submitBtn = form.querySelector('[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.textContent = 'Sending...';
        submitBtn.disabled = true;

        try {
            const response = await fetch(action, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                this.showSuccess(form, 'Thank you! We\'ll be in touch soon.');
                form.reset();
            } else {
                this.showError(form, 'Something went wrong. Please try again.');
            }
        } catch (error) {
            this.showError(form, 'Network error. Please check your connection.');
        } finally {
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        }
    }

    showSuccess(form, message) {
        this.showMessage(form, message, 'success');
    }

    showError(form, message) {
        this.showMessage(form, message, 'error');
    }

    showMessage(form, message, type) {
        // Remove existing messages
        form.querySelectorAll('.form-message').forEach(msg => msg.remove());

        const messageEl = document.createElement('div');
        messageEl.className = `form-message form-message-${type}`;
        messageEl.textContent = message;

        form.appendChild(messageEl);

        setTimeout(() => {
            messageEl.remove();
        }, 5000);
    }
}

// ==========================================
// Resource Hints and Preloading
// ==========================================
class ResourceOptimizer {
    constructor() {
        this.init();
    }

    init() {
        // Preconnect to external domains
        this.addPreconnect('https://fonts.googleapis.com');
        this.addPreconnect('https://fonts.gstatic.com');
        this.addPreconnect('https://cdnjs.cloudflare.com');

        // Prefetch next likely pages
        this.setupPrefetch();

        // Preload critical resources
        this.preloadCriticalResources();
    }

    addPreconnect(url) {
        const link = document.createElement('link');
        link.rel = 'preconnect';
        link.href = url;
        link.crossOrigin = 'anonymous';
        document.head.appendChild(link);
    }

    setupPrefetch() {
        // Prefetch on link hover
        document.querySelectorAll('a[href^="/"]').forEach(link => {
            link.addEventListener('mouseenter', () => {
                this.prefetchPage(link.href);
            });
        });
    }

    prefetchPage(url) {
        // Check if already prefetched
        if (document.querySelector(`link[rel="prefetch"][href="${url}"]`)) {
            return;
        }

        const link = document.createElement('link');
        link.rel = 'prefetch';
        link.href = url;
        document.head.appendChild(link);
    }

    preloadCriticalResources() {
        // Preload critical CSS
        const criticalCSS = [
            '/static/css/modern-landing.css'
        ];

        criticalCSS.forEach(css => {
            const link = document.createElement('link');
            link.rel = 'preload';
            link.href = css;
            link.as = 'style';
            document.head.appendChild(link);
        });

        // Preload hero image
        const heroImage = document.querySelector('.hero-image');
        if (heroImage && heroImage.dataset.src) {
            const link = document.createElement('link');
            link.rel = 'preload';
            link.href = heroImage.dataset.src;
            link.as = 'image';
            document.head.appendChild(link);
        }
    }
}

// ==========================================
// Analytics Integration
// ==========================================
class Analytics {
    constructor() {
        this.init();
    }

    init() {
        // Track page views
        this.trackPageView();

        // Track CTA clicks
        this.trackCTAClicks();

        // Track scroll depth
        this.trackScrollDepth();
    }

    trackPageView() {
        if (typeof gtag !== 'undefined') {
            gtag('event', 'page_view', {
                page_path: window.location.pathname,
                page_title: document.title
            });
        }
    }

    trackCTAClicks() {
        document.querySelectorAll('[data-track-cta]').forEach(cta => {
            cta.addEventListener('click', () => {
                const action = cta.dataset.trackCta;
                this.trackEvent('CTA Click', action);
            });
        });
    }

    trackScrollDepth() {
        let scrollDepths = [25, 50, 75, 100];
        let achievedDepths = [];

        window.addEventListener('scroll', () => {
            const scrollPercent = (window.scrollY + window.innerHeight) / document.body.scrollHeight * 100;

            scrollDepths.forEach(depth => {
                if (scrollPercent >= depth && !achievedDepths.includes(depth)) {
                    achievedDepths.push(depth);
                    this.trackEvent('Scroll Depth', `${depth}%`);
                }
            });
        });
    }

    trackEvent(category, action, label = null, value = null) {
        if (typeof gtag !== 'undefined') {
            gtag('event', action, {
                event_category: category,
                event_label: label,
                value: value
            });
        }

        // Also send to internal analytics
        fetch('/api/v1/analytics/events', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                category,
                action,
                label,
                value,
                timestamp: new Date().toISOString()
            })
        }).catch(() => {
            // Silently fail
        });
    }
}

// ==========================================
// Service Worker Registration
// ==========================================
class ServiceWorkerManager {
    constructor() {
        this.init();
    }

    init() {
        if ('serviceWorker' in navigator) {
            window.addEventListener('load', () => {
                this.registerServiceWorker();
            });
        }
    }

    async registerServiceWorker() {
        try {
            const registration = await navigator.serviceWorker.register('/sw.js');
            console.log('ServiceWorker registered:', registration);

            // Check for updates
            registration.addEventListener('updatefound', () => {
                const newWorker = registration.installing;
                newWorker.addEventListener('statechange', () => {
                    if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                        // New service worker available
                        this.showUpdateNotification();
                    }
                });
            });
        } catch (error) {
            console.log('ServiceWorker registration failed:', error);
        }
    }

    showUpdateNotification() {
        const notification = document.createElement('div');
        notification.className = 'update-notification';
        notification.innerHTML = `
            <p>A new version is available!</p>
            <button onclick="location.reload()">Update</button>
        `;
        document.body.appendChild(notification);
    }
}

// ==========================================
// Initialize Everything
// ==========================================
class LandingPage {
    constructor() {
        this.initializeComponents();
        this.setupEventListeners();
    }

    initializeComponents() {
        // Performance monitoring
        this.performanceMonitor = new PerformanceMonitor();

        // Core functionality
        this.lazyLoader = new LazyLoader();
        this.navigation = new SmoothNavigation();
        this.pricingToggle = new PricingToggle();
        this.faqAccordion = new FAQAccordion();
        this.testimonialCarousel = new TestimonialCarousel();
        this.announcementBar = new AnnouncementBar();
        this.formHandler = new FormHandler();

        // Optimization
        this.resourceOptimizer = new ResourceOptimizer();
        this.analytics = new Analytics();

        // Progressive enhancement
        if ('serviceWorker' in navigator) {
            this.serviceWorker = new ServiceWorkerManager();
        }
    }

    setupEventListeners() {
        // Debounced resize handler
        let resizeTimeout;
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                this.handleResize();
            }, 250);
        });

        // Page visibility change
        document.addEventListener('visibilitychange', () => {
            this.handleVisibilityChange();
        });

        // Print styles
        window.addEventListener('beforeprint', () => {
            this.optimizeForPrint();
        });
    }

    handleResize() {
        // Recalculate any size-dependent features
        console.log('Window resized');
    }

    handleVisibilityChange() {
        if (document.hidden) {
            // Pause animations, stop autoplay
            this.testimonialCarousel?.stopAutoplay();
        } else {
            // Resume animations, restart autoplay
            this.testimonialCarousel?.startAutoplay();
        }
    }

    optimizeForPrint() {
        // Add print-specific optimizations
        document.body.classList.add('print-mode');
    }
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.landingPage = new LandingPage();
    });
} else {
    window.landingPage = new LandingPage();
}

// Export functions for global use
window.closeAnnouncement = closeAnnouncement;