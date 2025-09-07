// Landing page JavaScript
document.addEventListener('DOMContentLoaded', function() {
    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Header scroll effect
    const header = document.querySelector('.header');
    let lastScrollY = window.scrollY;

    window.addEventListener('scroll', () => {
        const currentScrollY = window.scrollY;
        
        if (currentScrollY > 100) {
            header.style.background = 'rgba(255, 255, 255, 0.95)';
            header.style.backdropFilter = 'blur(10px)';
        } else {
            header.style.background = 'white';
            header.style.backdropFilter = 'none';
        }
        
        lastScrollY = currentScrollY;
    });

    // Animate elements on scroll
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    // Observe feature cards and pricing cards
    document.querySelectorAll('.feature-card, .pricing-card').forEach(card => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(20px)';
        card.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(card);
    });

    // Demo status page animation
    const statusDemo = document.querySelector('.status-demo');
    if (statusDemo) {
        // Simulate status changes
        setInterval(() => {
            const services = statusDemo.querySelectorAll('.service');
            services.forEach((service, index) => {
                setTimeout(() => {
                    const status = service.querySelector('.service-status');
                    const indicator = statusDemo.querySelector('.status-indicator');
                    
                    // Randomly change status
                    if (Math.random() < 0.1) { // 10% chance
                        const statuses = ['operational', 'degraded', 'outage'];
                        const currentStatus = statuses[Math.floor(Math.random() * statuses.length)];
                        
                        service.className = `service ${currentStatus}`;
                        status.textContent = currentStatus.charAt(0).toUpperCase() + currentStatus.slice(1);
                        
                        // Update overall status
                        const hasOutage = Array.from(services).some(s => s.classList.contains('outage'));
                        const hasDegraded = Array.from(services).some(s => s.classList.contains('degraded'));
                        
                        if (hasOutage) {
                            indicator.className = 'status-indicator outage';
                            statusDemo.querySelector('.status-header h3').textContent = 'Service Outage';
                        } else if (hasDegraded) {
                            indicator.className = 'status-indicator degraded';
                            statusDemo.querySelector('.status-header h3').textContent = 'Degraded Performance';
                        } else {
                            indicator.className = 'status-indicator operational';
                            statusDemo.querySelector('.status-header h3').textContent = 'All Systems Operational';
                        }
                    }
                }, index * 200);
            });
        }, 5000);
    }

    // Pricing card hover effects
    document.querySelectorAll('.pricing-card').forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-8px) scale(1.02)';
        });
        
        card.addEventListener('mouseleave', function() {
            if (this.classList.contains('featured')) {
                this.style.transform = 'scale(1.05)';
            } else {
                this.style.transform = 'translateY(0) scale(1)';
            }
        });
    });

    // Feature card animations
    document.querySelectorAll('.feature-card').forEach((card, index) => {
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-8px)';
            this.querySelector('.feature-icon').style.transform = 'scale(1.1)';
        });
        
        card.addEventListener('mouseleave', function() {
            this.style.transform = 'translateY(0)';
            this.querySelector('.feature-icon').style.transform = 'scale(1)';
        });
    });

    // Add loading animation to buttons
    document.querySelectorAll('.btn').forEach(btn => {
        btn.addEventListener('click', function(e) {
            if (this.href && this.href.includes('#')) {
                return; // Don't add loading for anchor links
            }
            
            const originalText = this.textContent;
            this.textContent = 'Loading...';
            this.style.pointerEvents = 'none';
            
            // Reset after 2 seconds (for demo purposes)
            setTimeout(() => {
                this.textContent = originalText;
                this.style.pointerEvents = 'auto';
            }, 2000);
        });
    });

    // Add parallax effect to hero section
    window.addEventListener('scroll', () => {
        const scrolled = window.pageYOffset;
        const hero = document.querySelector('.hero');
        if (hero) {
            hero.style.transform = `translateY(${scrolled * 0.5}px)`;
        }
    });

    // Add typing effect to hero title
    const heroTitle = document.querySelector('.hero-content h1');
    if (heroTitle) {
        const text = heroTitle.textContent;
        heroTitle.textContent = '';
        let i = 0;
        
        const typeWriter = () => {
            if (i < text.length) {
                heroTitle.textContent += text.charAt(i);
                i++;
                setTimeout(typeWriter, 50);
            }
        };
        
        // Start typing effect after a short delay
        setTimeout(typeWriter, 500);
    }
});
