// API Type Definitions for SaaS Admin Service
// Generated from Go backend models

// ============= Authentication =============

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  status: string;
  message: string;
  token?: string;
  user?: AdminUser;
}

export interface AdminUser {
  id: number;
  username: string;
  email: string;
  role: string;
  created_at: string;
  updated_at: string;
}

// ============= Tenants =============

export interface Tenant {
  id: string; // UUID
  name: string;
  domain?: string;
  subdomain?: string;
  contact_email: string;
  billing_email?: string;
  plan_id?: string;
  is_active: boolean;
  max_users?: number | null;
  settings?: Record<string, any>;
  branding?: Record<string, any>;
  features?: Record<string, any>;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
  // Metrics (from stats endpoint)
  user_count?: number;
  subscription_status?: string;
}

export interface TenantsResponse {
  status: string;
  data: Tenant[];
  count?: number;
  limit?: number;
  offset?: number;
}

export interface TenantResponse {
  status: string;
  data: Tenant;
}

export interface CreateTenantRequest {
  name: string;
  contact_email: string;
  billing_email?: string;
  plan_id?: string;
  max_users?: number | null;
  settings?: Record<string, any>;
  branding?: Record<string, any>;
  features?: Record<string, any>;
  metadata?: Record<string, any>;
}

export interface UpdateTenantRequest {
  name?: string;
  contact_email?: string;
  billing_email?: string;
  plan_id?: string;
  is_active?: boolean;
  max_users?: number | null;
  settings?: Record<string, any>;
  branding?: Record<string, any>;
  features?: Record<string, any>;
  metadata?: Record<string, any>;
}

// ============= Plans =============

export interface Plan {
  id: string; // UUID
  name: string;
  slug: string;
  description?: string;
  price: number;
  billing_period?: string; // monthly, yearly
  features?: string[];
  max_users?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PlansResponse {
  status?: string;
  plans: Plan[];
  count: number;
  limit?: number;
  offset?: number;
}

export interface PlanResponse {
  status: string;
  data: Plan;
}

// ============= Stats & Analytics =============

export interface DashboardStats {
  total_tenants: number;
  active_tenants: number;
  monthly_revenue: number;
  active_subscriptions: number;
  churn_rate: number;
  growth_rate: number;
}

export interface StatsResponse {
  status: string;
  data: DashboardStats;
}

export interface AnalyticsData {
  revenue_trend: Array<{ date: string; revenue: number }>;
  subscription_distribution: Array<{ plan: string; count: number }>;
  tenant_growth: Array<{ date: string; count: number }>;
}

export interface AnalyticsResponse {
  status: string;
  data: AnalyticsData;
}

// ============= Features =============

export interface Feature {
  id: string;
  name: string;
  slug: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface FeaturesResponse {
  status: string;
  data: Feature[];
  count?: number;
}

// ============= Notifications =============

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
  is_read: boolean;
  created_at: string;
}

export interface NotificationsResponse {
  status: string;
  data: Notification[];
  count?: number;
}

// ============= Activity Logs =============

export interface Activity {
  id: string;
  action: string;
  description: string;
  user_id?: number;
  user_email?: string;
  tenant_id?: string;
  created_at: string;
}

export interface ActivitiesResponse {
  status: string;
  data: Activity[];
  count?: number;
}

// ============= Health Check =============

export interface HealthResponse {
  service: string;
  status: string;
  timestamp: string;
  components?: Record<string, any>;
}

// ============= Error Response =============

export interface ErrorResponse {
  status: string;
  error: string;
  message?: string;
  details?: any;
}

// ============= Common =============

export interface ApiResponse<T = any> {
  status: string;
  data?: T;
  message?: string;
  error?: string;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface ListResponse<T> {
  status: string;
  data: T[];
  count: number;
  limit?: number;
  offset?: number;
}
