-- Test fixtures for statuspage_branding database
-- Used for integration and E2E testing

-- Clear existing test data
TRUNCATE TABLE custom_css, custom_domains, logos, themes, brands CASCADE;

-- Insert brands for Tenant 1
INSERT INTO brands (id, tenant_id, brand_name, is_default, created_at, updated_at) VALUES
('brand-1111-0001-0001-000000000001', 'tenant-1111-1111-1111-111111111111', 'Test1 Corporate Brand', true, NOW() - INTERVAL '6 months', NOW()),
('brand-1111-0002-0002-000000000002', 'tenant-1111-1111-1111-111111111111', 'Test1 Developer Portal', false, NOW() - INTERVAL '3 months', NOW());

-- Insert brands for Tenant 2
INSERT INTO brands (id, tenant_id, brand_name, is_default, created_at, updated_at) VALUES
('brand-2222-0001-0001-000000000001', 'tenant-2222-2222-2222-222222222222', 'Test2 Main Brand', true, NOW() - INTERVAL '3 months', NOW());

-- Insert brands for Tenant 3
INSERT INTO brands (id, tenant_id, brand_name, is_default, created_at, updated_at) VALUES
('brand-3333-0001-0001-000000000001', 'tenant-3333-3333-3333-333333333333', 'Test3 Default', true, NOW(), NOW());

-- Insert themes
INSERT INTO themes (id, brand_id, theme_name, primary_color, secondary_color, accent_color, background_color, text_color, link_color, header_background, header_text, footer_background, footer_text, font_family, created_at, updated_at) VALUES
-- Corporate theme for Tenant 1
('theme-1111-0001-0001-000000000001', 'brand-1111-0001-0001-000000000001', 'Corporate Blue', '#0066CC', '#0052A3', '#FF9900', '#FFFFFF', '#333333', '#0066CC', '#0066CC', '#FFFFFF', '#F5F5F5', '#666666', 'Inter, sans-serif', NOW() - INTERVAL '6 months', NOW()),

-- Developer portal theme for Tenant 1
('theme-1111-0002-0002-000000000002', 'brand-1111-0002-0002-000000000002', 'Developer Dark', '#1E1E1E', '#2D2D2D', '#00D9FF', '#121212', '#E0E0E0', '#00D9FF', '#1E1E1E', '#FFFFFF', '#0D0D0D', '#CCCCCC', 'JetBrains Mono, monospace', NOW() - INTERVAL '3 months', NOW()),

-- Main theme for Tenant 2
('theme-2222-0001-0001-000000000001', 'brand-2222-0001-0001-000000000001', 'Modern Green', '#10B981', '#059669', '#F59E0B', '#FFFFFF', '#1F2937', '#10B981', '#10B981', '#FFFFFF', '#F3F4F6', '#6B7280', 'Roboto, sans-serif', NOW() - INTERVAL '3 months', NOW()),

-- Default theme for Tenant 3
('theme-3333-0001-0001-000000000001', 'brand-3333-0001-0001-000000000001', 'Default Theme', '#3B82F6', '#2563EB', '#F97316', '#FFFFFF', '#374151', '#3B82F6', '#3B82F6', '#FFFFFF', '#F9FAFB', '#6B7280', 'system-ui, sans-serif', NOW(), NOW());

-- Insert logos
INSERT INTO logos (id, brand_id, logo_type, file_name, file_path, file_size_bytes, mime_type, width_px, height_px, created_at, updated_at) VALUES
-- Logos for Tenant 1 Corporate Brand
('logo-1111-0001-0001-000000000001', 'brand-1111-0001-0001-000000000001', 'primary', 'test1-logo.svg', '/uploads/brands/brand-1111-0001/logo-primary.svg', 12580, 'image/svg+xml', 200, 60, NOW() - INTERVAL '6 months', NOW()),
('logo-1111-0001-0002-000000000002', 'brand-1111-0001-0001-000000000001', 'favicon', 'test1-favicon.png', '/uploads/brands/brand-1111-0001/favicon.png', 3240, 'image/png', 32, 32, NOW() - INTERVAL '6 months', NOW()),
('logo-1111-0001-0003-000000000003', 'brand-1111-0001-0001-000000000001', 'icon', 'test1-icon.png', '/uploads/brands/brand-1111-0001/icon.png', 15680, 'image/png', 512, 512, NOW() - INTERVAL '6 months', NOW()),

-- Logos for Tenant 1 Developer Portal
('logo-1111-0002-0001-000000000001', 'brand-1111-0002-0002-000000000002', 'primary', 'test1-dev-logo.svg', '/uploads/brands/brand-1111-0002/logo-primary.svg', 8920, 'image/svg+xml', 180, 50, NOW() - INTERVAL '3 months', NOW()),
('logo-1111-0002-0002-000000000002', 'brand-1111-0002-0002-000000000002', 'favicon', 'test1-dev-favicon.png', '/uploads/brands/brand-1111-0002/favicon.png', 2840, 'image/png', 32, 32, NOW() - INTERVAL '3 months', NOW()),

-- Logos for Tenant 2
('logo-2222-0001-0001-000000000001', 'brand-2222-0001-0001-000000000001', 'primary', 'test2-logo.png', '/uploads/brands/brand-2222-0001/logo-primary.png', 45230, 'image/png', 250, 80, NOW() - INTERVAL '3 months', NOW()),
('logo-2222-0001-0002-000000000002', 'brand-2222-0001-0001-000000000001', 'favicon', 'test2-favicon.ico', '/uploads/brands/brand-2222-0001/favicon.ico', 5680, 'image/x-icon', 32, 32, NOW() - INTERVAL '3 months', NOW()),

-- Logo for Tenant 3 (minimal setup)
('logo-3333-0001-0001-000000000001', 'brand-3333-0001-0001-000000000001', 'primary', 'test3-logo.svg', '/uploads/brands/brand-3333-0001/logo-primary.svg', 9850, 'image/svg+xml', 180, 50, NOW(), NOW());

-- Insert custom domains
INSERT INTO custom_domains (id, brand_id, domain, is_verified, verification_token, verification_method, ssl_enabled, ssl_certificate_expires_at, dns_configured, created_at, updated_at, verified_at) VALUES
-- Verified domain for Tenant 1 Corporate
('domain-1111-0001-0001-000000000001', 'brand-1111-0001-0001-000000000001', 'status.test1.com', true, NULL, 'dns_txt', true, NOW() + INTERVAL '3 months', true, NOW() - INTERVAL '6 months', NOW(), NOW() - INTERVAL '6 months'),

-- Verified domain for Tenant 1 Developer
('domain-1111-0002-0001-000000000001', 'brand-1111-0002-0002-000000000002', 'devstatus.test1.com', true, NULL, 'dns_txt', true, NOW() + INTERVAL '4 months', true, NOW() - INTERVAL '3 months', NOW(), NOW() - INTERVAL '3 months'),

-- Verified domain for Tenant 2
('domain-2222-0001-0001-000000000001', 'brand-2222-0001-0001-000000000001', 'status.test2.com', true, NULL, 'dns_cname', true, NOW() + INTERVAL '5 months', true, NOW() - INTERVAL '3 months', NOW(), NOW() - INTERVAL '3 months'),

-- Unverified domain for Tenant 3 (trial tenant)
('domain-3333-0001-0001-000000000001', 'brand-3333-0001-0001-000000000001', 'status.test3.com', false, 'verify-token-3333-abcd1234', 'dns_txt', false, NULL, false, NOW(), NOW(), NULL);

-- Insert custom CSS
INSERT INTO custom_css (id, brand_id, css_name, css_content, is_active, created_at, updated_at) VALUES
-- Custom CSS for Tenant 1 Corporate
('css-1111-0001-0001-000000000001', 'brand-1111-0001-0001-000000000001', 'Corporate Styling',
'/* Custom styles for Test1 Corporate */
.status-page {
  font-family: Inter, sans-serif;
  max-width: 1200px;
  margin: 0 auto;
}

.incident-card {
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  padding: 20px;
  margin-bottom: 16px;
}

.component-status.operational {
  background-color: #10B981;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
}

.component-status.degraded {
  background-color: #F59E0B;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
}

.component-status.outage {
  background-color: #EF4444;
  color: white;
  padding: 4px 12px;
  border-radius: 4px;
}

.subscribe-button {
  background-color: #0066CC;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.subscribe-button:hover {
  background-color: #0052A3;
}',
true, NOW() - INTERVAL '6 months', NOW() - INTERVAL '1 month'),

-- Custom CSS for Tenant 1 Developer Portal
('css-1111-0002-0001-000000000001', 'brand-1111-0002-0002-000000000002', 'Developer Dark Mode',
'/* Dark mode for Developer Portal */
body {
  background-color: #121212;
  color: #E0E0E0;
}

.status-page {
  font-family: "JetBrains Mono", monospace;
  background: linear-gradient(135deg, #1E1E1E 0%, #2D2D2D 100%);
  border-radius: 12px;
  padding: 32px;
}

.component-card {
  background-color: #1E1E1E;
  border: 1px solid #404040;
  border-radius: 8px;
  padding: 16px;
  margin: 12px 0;
}

.incident-badge {
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

code {
  background-color: #2D2D2D;
  padding: 2px 6px;
  border-radius: 4px;
  color: #00D9FF;
}',
true, NOW() - INTERVAL '3 months', NOW() - INTERVAL '2 weeks'),

-- Custom CSS for Tenant 2
('css-2222-0001-0001-000000000001', 'brand-2222-0001-0001-000000000001', 'Modern Clean',
'/* Clean modern design */
.status-page {
  font-family: Roboto, sans-serif;
}

.header {
  background: linear-gradient(90deg, #10B981 0%, #059669 100%);
  padding: 48px 24px;
  text-align: center;
  border-radius: 12px 12px 0 0;
}

.component-group {
  margin: 24px 0;
  padding: 20px;
  background-color: #F9FAFB;
  border-radius: 8px;
}

.metric-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background-color: white;
  border-left: 4px solid #10B981;
  margin: 8px 0;
}',
true, NOW() - INTERVAL '3 months', NOW());

-- Print summary
SELECT 'Fixtures loaded:' as message;
SELECT COUNT(*) as brands FROM brands;
SELECT COUNT(*) as themes FROM themes;
SELECT COUNT(*) as logos FROM logos;
SELECT COUNT(*) as custom_domains FROM custom_domains;
SELECT COUNT(*) as custom_css FROM custom_css;
SELECT logo_type, COUNT(*) as count FROM logos GROUP BY logo_type ORDER BY logo_type;
SELECT is_verified, COUNT(*) as count FROM custom_domains GROUP BY is_verified ORDER BY is_verified;
SELECT is_default, COUNT(*) as count FROM brands GROUP BY is_default ORDER BY is_default;
