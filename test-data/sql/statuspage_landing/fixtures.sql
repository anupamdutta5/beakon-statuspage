-- Test fixtures for statuspage_landing database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE landing_page_sections, landing_page_assets, landing_page_forms, landing_pages CASCADE;

-- Insert landing pages
INSERT INTO landing_pages (id, tenant_id, page_name, page_slug, title, subtitle, meta_title, meta_description, meta_keywords, is_published, template_name, custom_css, custom_js, created_at, updated_at, published_at) VALUES
-- Tenant 1 main landing page
('land-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Test1 Homepage', 'home', 'Welcome to Test1 - Reliable Status Page Platform', 'Real-time status monitoring and incident management for your services', 'Test1 - Status Page & Monitoring Platform', 'Professional status page and uptime monitoring solution for modern businesses', 'status page, uptime monitoring, incident management, service health', true, 'modern',
'/* Custom homepage styles */
.hero-section {
  background: linear-gradient(135deg, #0066CC 0%, #0052A3 100%);
  padding: 80px 20px;
  text-align: center;
  color: white;
}

.feature-card {
  background: white;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  margin: 16px;
}

.cta-button {
  background-color: #FF9900;
  color: white;
  padding: 16px 40px;
  border-radius: 8px;
  font-size: 18px;
  font-weight: bold;
  border: none;
  cursor: pointer;
  transition: transform 0.2s;
}

.cta-button:hover {
  transform: scale(1.05);
}',
'// Custom analytics tracking
(function() {
  console.log("Test1 landing page loaded");
  // Track page view
  if (window.analytics) {
    analytics.page("Home", { title: "Test1 Homepage" });
  }
})();',
NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week', NOW() - INTERVAL '6 months'),

-- Tenant 1 pricing page
('land-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'Pricing', 'pricing', 'Test1 Pricing - Plans That Scale With You', 'Choose the perfect plan for your team', 'Test1 Pricing Plans - Affordable Status Page Solutions', 'Flexible pricing plans for teams of all sizes', 'pricing, plans, subscription, costs', true, 'pricing',
'/* Pricing page styles */
.pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
  padding: 40px 20px;
}

.pricing-card {
  background: white;
  border: 2px solid #E5E7EB;
  border-radius: 16px;
  padding: 40px;
  text-align: center;
}

.pricing-card.featured {
  border-color: #0066CC;
  box-shadow: 0 8px 24px rgba(0, 102, 204, 0.2);
  transform: scale(1.05);
}

.price {
  font-size: 48px;
  font-weight: bold;
  color: #0066CC;
}',
NULL,
NOW() - INTERVAL '6 months', NOW() - INTERVAL '2 weeks', NOW() - INTERVAL '6 months'),

-- Tenant 1 features page
('land-1111-0003-0003-000000000003', 'tenant-1111-1111-1111-111111111111', 'Features', 'features', 'Powerful Features for Modern Teams', 'Everything you need to communicate service health', 'Test1 Features - Complete Status Page Solution', 'Explore all features of Test1 status page platform', 'features, capabilities, functionality', true, 'features', NULL, NULL, NOW() - INTERVAL '5 months', NOW() - INTERVAL '1 month', NOW() - INTERVAL '5 months'),

-- Tenant 2 landing page
('land-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'Test2 Home', 'home', 'Test2 - Enterprise Status Communication', 'Professional status pages for enterprise teams', 'Test2 Status Pages - Enterprise Solution', 'Enterprise-grade status page and incident communication platform', 'enterprise status page, incident communication', true, 'enterprise', NULL, NULL, NOW() - INTERVAL '3 months', NOW(), NOW() - INTERVAL '3 months'),

-- Tenant 3 landing page (unpublished draft)
('land-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'Test3 Homepage', 'home', 'Test3 Services', 'Coming soon', 'Test3', 'Test3 status page', 'status', false, 'minimal', NULL, NULL, NOW(), NOW(), NULL);

-- Insert landing page sections for Tenant 1 homepage
INSERT INTO landing_page_sections (id, page_id, section_type, section_name, content, display_order, is_visible, background_color, text_color, created_at, updated_at) VALUES
-- Hero section
('sect-1111-0001-0001-000000000001', 'land-1111-0001-0001-000000000001', 'hero', 'Hero Banner',
'{"heading": "Communicate Service Status with Confidence", "subheading": "Professional status pages trusted by thousands of teams worldwide", "cta_primary": {"text": "Start Free Trial", "url": "/signup"}, "cta_secondary": {"text": "View Demo", "url": "/demo"}, "background_image": "/assets/hero-bg.jpg"}',
1, true, '#0066CC', '#FFFFFF', NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week'),

-- Features section
('sect-1111-0001-0002-000000000002', 'land-1111-0001-0001-000000000001', 'features', 'Key Features',
'{"heading": "Everything You Need", "subheading": "Powerful features for modern teams", "features": [{"icon": "chart-line", "title": "Real-Time Monitoring", "description": "Monitor uptime and performance 24/7 with automated checks"}, {"icon": "bell", "title": "Multi-Channel Alerts", "description": "Notify users via email, SMS, Slack, and webhooks"}, {"icon": "palette", "title": "Custom Branding", "description": "Fully customize your status page to match your brand"}, {"icon": "shield", "title": "99.99% Uptime SLA", "description": "Enterprise-grade reliability you can count on"}]}',
2, true, '#FFFFFF', '#333333', NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week'),

-- Social proof section
('sect-1111-0001-0003-000000000003', 'land-1111-0001-0001-000000000001', 'testimonials', 'Customer Testimonials',
'{"heading": "Trusted by Leading Teams", "testimonials": [{"name": "Sarah Johnson", "title": "CTO", "company": "TechCorp", "quote": "Test1 has transformed how we communicate with our customers during incidents. Essential tool for any SaaS business.", "avatar": "/assets/testimonials/sarah.jpg"}, {"name": "Michael Chen", "title": "VP Engineering", "company": "DataFlow", "quote": "The best status page solution we''ve used. Easy to set up and incredibly reliable.", "avatar": "/assets/testimonials/michael.jpg"}, {"name": "Emily Rodriguez", "title": "DevOps Lead", "company": "CloudScale", "quote": "Fantastic product with excellent support. The multi-channel notifications are a game changer.", "avatar": "/assets/testimonials/emily.jpg"}]}',
3, true, '#F9FAFB', '#1F2937', NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week'),

-- Pricing preview section
('sect-1111-0001-0004-000000000004', 'land-1111-0001-0001-000000000001', 'pricing_preview', 'Pricing',
'{"heading": "Simple, Transparent Pricing", "subheading": "Plans that grow with your business", "show_annual_toggle": true, "cta": {"text": "View All Plans", "url": "/pricing"}, "featured_plans": [{"name": "Starter", "price": "$29", "period": "month", "features": ["5 components", "Email notifications", "90-day history"]}, {"name": "Professional", "price": "$99", "period": "month", "featured": true, "features": ["Unlimited components", "All notification channels", "Custom branding", "1-year history"]}, {"name": "Enterprise", "price": "Custom", "features": ["Everything in Pro", "SSO/SAML", "99.99% SLA", "Dedicated support"]}]}',
4, true, '#FFFFFF', '#333333', NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week'),

-- CTA section
('sect-1111-0001-0005-000000000005', 'land-1111-0001-0001-000000000001', 'cta', 'Final Call to Action',
'{"heading": "Ready to Get Started?", "subheading": "Join thousands of teams using Test1 to communicate service health", "cta_primary": {"text": "Start Your Free Trial", "url": "/signup"}, "cta_secondary": {"text": "Schedule a Demo", "url": "/demo"}}',
5, true, '#0066CC', '#FFFFFF', NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week'),

-- Footer section
('sect-1111-0001-0006-000000000006', 'land-1111-0001-0001-000000000001', 'footer', 'Footer',
'{"columns": [{"heading": "Product", "links": [{"text": "Features", "url": "/features"}, {"text": "Pricing", "url": "/pricing"}, {"text": "Integrations", "url": "/integrations"}]}, {"heading": "Company", "links": [{"text": "About", "url": "/about"}, {"text": "Blog", "url": "/blog"}, {"text": "Careers", "url": "/careers"}]}, {"heading": "Support", "links": [{"text": "Documentation", "url": "/docs"}, {"text": "Contact", "url": "/contact"}, {"text": "Status", "url": "https://status.test1.com"}]}], "social_links": [{"platform": "twitter", "url": "https://twitter.com/test1"}, {"platform": "linkedin", "url": "https://linkedin.com/company/test1"}], "copyright": "© 2024 Test1. All rights reserved."}',
6, true, '#1F2937', '#E5E7EB', NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 week');

-- Insert landing page sections for Tenant 1 pricing page
INSERT INTO landing_page_sections (id, page_id, section_type, section_name, content, display_order, is_visible, background_color, text_color, created_at, updated_at) VALUES
('sect-1111-0002-0001-000000000001', 'land-1111-0002-0002-000000000002', 'hero', 'Pricing Header',
'{"heading": "Choose Your Plan", "subheading": "Flexible pricing for teams of all sizes"}',
1, true, '#F9FAFB', '#1F2937', NOW() - INTERVAL '6 months', NOW() - INTERVAL '2 weeks'),

('sect-1111-0002-0002-000000000002', 'land-1111-0002-0002-000000000002', 'pricing_table', 'Pricing Tiers',
'{"show_annual_discount": true, "annual_discount_percentage": 20, "plans": [{"id": "free", "name": "Free", "price": 0, "features": ["1 status page", "3 components", "Email notifications", "30-day history"]}, {"id": "starter", "name": "Starter", "price": 29, "features": ["1 status page", "10 components", "Email + SMS notifications", "90-day history"]}, {"id": "professional", "name": "Professional", "price": 99, "featured": true, "features": ["5 status pages", "Unlimited components", "All notification channels", "Custom branding", "1-year history", "API access"]}, {"id": "enterprise", "name": "Enterprise", "price": null, "contact_sales": true, "features": ["Everything in Pro", "Unlimited status pages", "SSO/SAML", "Custom integrations", "99.99% SLA", "Dedicated support"]}]}',
2, true, '#FFFFFF', '#333333', NOW() - INTERVAL '6 months', NOW() - INTERVAL '2 weeks'),

('sect-1111-0002-0003-000000000003', 'land-1111-0002-0002-000000000002', 'faq', 'Pricing FAQ',
'{"heading": "Frequently Asked Questions", "questions": [{"question": "Can I change plans anytime?", "answer": "Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately."}, {"question": "What payment methods do you accept?", "answer": "We accept all major credit cards (Visa, MasterCard, Amex) and PayPal for monthly plans."}, {"question": "Is there a free trial?", "answer": "Yes! All paid plans include a 14-day free trial. No credit card required."}, {"question": "What happens if I exceed my limits?", "answer": "We''ll notify you when you approach your limits. You can upgrade to a higher plan or contact us for custom pricing."}]}',
3, true, '#F9FAFB', '#1F2937', NOW() - INTERVAL '6 months', NOW() - INTERVAL '2 weeks');

-- Insert landing page assets
INSERT INTO landing_page_assets (id, page_id, asset_type, asset_name, file_name, file_path, file_size_bytes, mime_type, alt_text, created_at, updated_at) VALUES
-- Assets for Tenant 1 homepage
('asset-1111-0001-0001-000000000001', 'land-1111-0001-0001-000000000001', 'image', 'Hero Background', 'hero-bg.jpg', '/uploads/landing/land-1111-0001/hero-bg.jpg', 245680, 'image/jpeg', 'Professional team collaborating', NOW() - INTERVAL '6 months', NOW()),

('asset-1111-0001-0002-000000000002', 'land-1111-0001-0001-000000000001', 'image', 'Feature Screenshot 1', 'feature-dashboard.png', '/uploads/landing/land-1111-0001/feature-dashboard.png', 182340, 'image/png', 'Status page dashboard interface', NOW() - INTERVAL '6 months', NOW()),

('asset-1111-0001-0003-000000000003', 'land-1111-0001-0001-000000000001', 'image', 'Feature Screenshot 2', 'feature-notifications.png', '/uploads/landing/land-1111-0001/feature-notifications.png', 156780, 'image/png', 'Multi-channel notification settings', NOW() - INTERVAL '6 months', NOW()),

('asset-1111-0001-0004-000000000004', 'land-1111-0001-0001-000000000001', 'image', 'Testimonial Avatar 1', 'sarah.jpg', '/uploads/landing/land-1111-0001/testimonials/sarah.jpg', 12450, 'image/jpeg', 'Sarah Johnson headshot', NOW() - INTERVAL '6 months', NOW()),

('asset-1111-0001-0005-000000000005', 'land-1111-0001-0001-000000000001', 'image', 'Testimonial Avatar 2', 'michael.jpg', '/uploads/landing/land-1111-0001/testimonials/michael.jpg', 14230, 'image/jpeg', 'Michael Chen headshot', NOW() - INTERVAL '6 months', NOW()),

('asset-1111-0001-0006-000000000006', 'land-1111-0001-0001-000000000001', 'image', 'Testimonial Avatar 3', 'emily.jpg', '/uploads/landing/land-1111-0001/testimonials/emily.jpg', 13890, 'image/jpeg', 'Emily Rodriguez headshot', NOW() - INTERVAL '6 months', NOW()),

('asset-1111-0001-0007-000000000007', 'land-1111-0001-0001-000000000001', 'video', 'Product Demo', 'product-demo.mp4', '/uploads/landing/land-1111-0001/product-demo.mp4', 8450600, 'video/mp4', 'Test1 product demonstration video', NOW() - INTERVAL '5 months', NOW()),

-- Assets for Tenant 2 homepage
('asset-2222-0001-0001-000000000001', 'land-2222-0001-0001-000000000001', 'image', 'Enterprise Hero', 'enterprise-hero.jpg', '/uploads/landing/land-2222-0001/enterprise-hero.jpg', 312450, 'image/jpeg', 'Enterprise team meeting', NOW() - INTERVAL '3 months', NOW());

-- Insert landing page forms
INSERT INTO landing_page_forms (id, page_id, form_name, form_type, fields, submit_url, success_message, notification_emails, is_active, created_at, updated_at) VALUES
-- Contact form for Tenant 1 homepage
('form-1111-0001-0001-000000000001', 'land-1111-0001-0001-000000000001', 'Contact Us', 'contact',
'[{"name": "name", "label": "Full Name", "type": "text", "required": true, "placeholder": "John Doe"}, {"name": "email", "label": "Email Address", "type": "email", "required": true, "placeholder": "john@example.com"}, {"name": "company", "label": "Company", "type": "text", "required": false, "placeholder": "Acme Inc."}, {"name": "message", "label": "Message", "type": "textarea", "required": true, "placeholder": "Tell us how we can help..."}]',
'/api/v1/forms/contact',
'Thank you for reaching out! We''ll get back to you within 24 hours.',
'["sales@test1.com", "support@test1.com"]',
true, NOW() - INTERVAL '6 months', NOW()),

-- Demo request form for Tenant 1 homepage
('form-1111-0001-0002-000000000002', 'land-1111-0001-0001-000000000001', 'Request Demo', 'demo',
'[{"name": "name", "label": "Full Name", "type": "text", "required": true}, {"name": "email", "label": "Work Email", "type": "email", "required": true}, {"name": "company", "label": "Company Name", "type": "text", "required": true}, {"name": "company_size", "label": "Company Size", "type": "select", "required": true, "options": ["1-10", "11-50", "51-200", "201-1000", "1000+"]}, {"name": "phone", "label": "Phone Number", "type": "tel", "required": false}]',
'/api/v1/forms/demo',
'Demo request received! Our team will contact you shortly to schedule a personalized demo.',
'["sales@test1.com"]',
true, NOW() - INTERVAL '6 months', NOW()),

-- Newsletter signup for Tenant 1
('form-1111-0001-0003-000000000003', 'land-1111-0001-0001-000000000001', 'Newsletter Signup', 'newsletter',
'[{"name": "email", "label": "Email Address", "type": "email", "required": true, "placeholder": "your@email.com"}]',
'/api/v1/forms/newsletter',
'You''re subscribed! Check your email for confirmation.',
'["marketing@test1.com"]',
true, NOW() - INTERVAL '6 months', NOW()),

-- Enterprise contact form for Tenant 2
('form-2222-0001-0001-000000000001', 'land-2222-0001-0001-000000000001', 'Enterprise Inquiry', 'contact',
'[{"name": "name", "label": "Name", "type": "text", "required": true}, {"name": "email", "label": "Email", "type": "email", "required": true}, {"name": "company", "label": "Company", "type": "text", "required": true}, {"name": "employees", "label": "Number of Employees", "type": "select", "required": true, "options": ["50-100", "101-500", "501-1000", "1000+"]}, {"name": "message", "label": "Requirements", "type": "textarea", "required": true}]',
'/api/v1/forms/enterprise',
'Thank you for your interest. An enterprise specialist will contact you within 1 business day.',
'["enterprise@test2.com"]',
true, NOW() - INTERVAL '3 months', NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as landing_pages FROM landing_pages;
SELECT COUNT(*) as sections FROM landing_page_sections;
SELECT COUNT(*) as assets FROM landing_page_assets;
SELECT COUNT(*) as forms FROM landing_page_forms;
SELECT is_published, COUNT(*) as count FROM landing_pages GROUP BY is_published ORDER BY is_published;
SELECT section_type, COUNT(*) as count FROM landing_page_sections GROUP BY section_type ORDER BY section_type;
SELECT asset_type, COUNT(*) as count FROM landing_page_assets GROUP BY asset_type ORDER BY asset_type;
SELECT form_type, COUNT(*) as count FROM landing_page_forms GROUP BY form_type ORDER BY form_type;
