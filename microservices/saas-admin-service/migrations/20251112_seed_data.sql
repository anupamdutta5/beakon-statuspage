-- Seed data for SaaS Admin Service

-- Insert default admin user (password: 'password')
INSERT INTO saas_admin_users (email, username, password, first_name, last_name, role)
VALUES ('admin@beakon.io', 'admin', '$2a$10$K7L1OJ0TfgBAiPAQcmBUF.pr.3mLFnzlV1VlU4g5peR9Hg19IsqMy', 'Admin', 'User', 'super_admin')
ON CONFLICT (username) DO NOTHING;

-- Insert default plans with all required fields
INSERT INTO saas_plans (
    id, name, slug, description, price, currency, billing_interval,
    max_tenants, max_users, max_services, max_monitors, max_subscribers, max_incidents, max_maintenance,
    custom_domain, white_label, api, integrations, analytics, support,
    is_active, is_public, is_popular, button_text, button_url, display_order,
    features, limits
) VALUES
(
    gen_random_uuid(), 'Free', 'free', 'Perfect for getting started', 0, 'USD', 'monthly',
    1, 3, 5, 10, 50, 50, 10,
    false, false, false, false, false, 'email',
    true, true, false, 'Get Started Free', '/signup?plan=free', 1,
    '["Basic monitoring", "Email notifications", "Public status page", "5 components", "Community support"]',
    '{"api_calls": 1000, "data_retention_days": 7}'
),
(
    gen_random_uuid(), 'Starter', 'starter', 'Great for small teams', 29, 'USD', 'monthly',
    1, 10, 15, 50, 200, 200, 50,
    true, false, true, true, false, 'email',
    true, true, true, 'Start Free Trial', '/signup?plan=starter', 2,
    '["Everything in Free", "Custom domain", "API access", "Slack & Discord integrations", "15 components", "Priority email support"]',
    '{"api_calls": 10000, "data_retention_days": 30}'
),
(
    gen_random_uuid(), 'Professional', 'professional', 'For growing businesses', 99, 'USD', 'monthly',
    1, 50, 50, 200, 1000, 1000, 200,
    true, true, true, true, true, 'chat',
    true, true, false, 'Start Free Trial', '/signup?plan=professional', 3,
    '["Everything in Starter", "White labeling", "Advanced analytics", "All integrations", "Unlimited components", "24/7 chat support", "SLA monitoring"]',
    '{"api_calls": 100000, "data_retention_days": 90}'
),
(
    gen_random_uuid(), 'Enterprise', 'enterprise', 'Custom solutions for large organizations', 299, 'USD', 'monthly',
    -1, -1, -1, -1, -1, -1, -1,
    true, true, true, true, true, 'phone',
    true, true, false, 'Contact Sales', '/contact-sales', 4,
    '["Everything in Professional", "Unlimited everything", "Custom integrations", "Dedicated support", "Custom SLAs", "On-premise option", "Advanced security"]',
    '{"api_calls": -1, "data_retention_days": -1}'
)
ON CONFLICT (slug) DO NOTHING;

-- Insert pricing features
INSERT INTO pricing_features (name, description, category, icon, is_active, "order") VALUES
('Status Page', 'Public status page for your services', 'core', 'fa-globe', true, 1),
('Custom Domain', 'Use your own domain for status pages', 'core', 'fa-link', true, 2),
('SSL Certificate', 'Free SSL certificate for custom domains', 'core', 'fa-lock', true, 3),
('API Access', 'Full API access for automation', 'advanced', 'fa-code', true, 4),
('Webhooks', 'Real-time webhook notifications', 'advanced', 'fa-webhook', true, 5),
('Team Management', 'Advanced team and permission management', 'advanced', 'fa-users', true, 6),
('White Labeling', 'Remove Beakon branding', 'enterprise', 'fa-paint-brush', true, 7),
('Priority Support', '24/7 priority support', 'enterprise', 'fa-headset', true, 8),
('SLA Monitoring', 'Service Level Agreement monitoring', 'enterprise', 'fa-chart-line', true, 9),
('Advanced Analytics', 'Detailed analytics and reporting', 'enterprise', 'fa-chart-bar', true, 10)
ON CONFLICT DO NOTHING;

-- Insert platform features
INSERT INTO saas_features (name, slug, description, category, is_enabled, is_public) VALUES
('Status Pages', 'status-pages', 'Create and manage status pages', 'core', true, true),
('Incident Management', 'incident-management', 'Track and communicate incidents', 'core', true, true),
('Component Monitoring', 'component-monitoring', 'Monitor service components', 'core', true, true),
('Scheduled Maintenance', 'scheduled-maintenance', 'Plan and announce maintenance windows', 'core', true, true),
('Subscriber Management', 'subscriber-management', 'Manage status page subscribers', 'core', true, true),
('Custom Domains', 'custom-domains', 'Use custom domains for status pages', 'advanced', true, true),
('API Access', 'api-access', 'Programmatic access via API', 'advanced', true, true),
('Integrations', 'integrations', 'Third-party service integrations', 'advanced', true, true),
('Analytics', 'analytics', 'Detailed analytics and reports', 'enterprise', true, true),
('White Labeling', 'white-labeling', 'Remove platform branding', 'enterprise', true, true)
ON CONFLICT (slug) DO NOTHING;

-- Create initial stats record
INSERT INTO saas_stats (
    total_tenants, active_tenants, total_users, active_users,
    total_revenue, monthly_revenue, total_plans, active_plans,
    total_features, active_features
) VALUES (0, 0, 0, 0, 0, 0, 4, 4, 10, 10)
ON CONFLICT DO NOTHING;