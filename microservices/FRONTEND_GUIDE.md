# Beakon Frontends - Complete Guide

**Last Updated:** 2025-10-21
**Frontend Framework:** Next.js 14 with React 18
**Architecture:** Separated Frontend/Backend Microservices

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Frontend Services](#frontend-services)
4. [Quick Start](#quick-start)
5. [Technology Stack](#technology-stack)
6. [Development Guide](#development-guide)
7. [Authentication & Security](#authentication--security)
8. [Subdomain Multi-Tenancy](#subdomain-multi-tenancy)
9. [API Integration](#api-integration)
10. [State Management](#state-management)
11. [Deployment](#deployment)
12. [Troubleshooting](#troubleshooting)

---

## Overview

Beakon uses **two separate Next.js 14 frontends** following the frontend-backend separation pattern:

1. **SaaS Admin Frontend** (Port 3001) - Platform administration
2. **Tenant Admin Frontend** (Port 3002) - Multi-tenant management

### Why Separate Frontends?

- **Independent Scaling:** Scale frontends independently from backends
- **Proper SSR:** Next.js Server-Side Rendering fixes authentication issues
- **Fast Refresh:** Better developer experience with hot module replacement
- **Production-Ready:** Docker builds, middleware authentication, environment-aware configs

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Browser                             │
└──────────────────┬─────────────────────┬────────────────────────┘
                   │                     │
                   │                     │
         ┌─────────▼─────────┐  ┌────────▼──────────┐
         │  SaaS Admin       │  │  Tenant Admin     │
         │  Frontend         │  │  Frontend         │
         │  (Next.js SSR)    │  │  (Next.js SSR)    │
         │  Port 3001        │  │  Port 3002        │
         │                   │  │  (subdomain)      │
         └─────────┬─────────┘  └────────┬──────────┘
                   │                     │
                   │ CORS                │ Same-Origin
                   │ (localhost:3001     │ (subdomain routing)
                   │  → localhost:8098)  │
                   │                     │
         ┌─────────▼─────────┐  ┌────────▼──────────┐
         │  SaaS Admin       │  │  Tenant Admin     │
         │  Backend (Go)     │  │  Backend (Go)     │
         │  Port 8098        │  │  Port 8099        │
         └───────────────────┘  └───────────────────┘
              ↓                        ↓
         PostgreSQL              PostgreSQL
         (saas_admin)            (tenant_admin_db)
```

---

## Frontend Services

### 1. SaaS Admin Frontend (Port 3001)

**Purpose:** Platform-wide administration and tenant management

**Repository:** https://github.com/anupamdutta5/saas-admin-frontend

**Key Features:**
- Dashboard with platform statistics
- Tenant management (CRUD operations)
- Subscription plan management
- Launch tenant admin panels
- User management

**Access:** http://localhost:3001

**Default Login:**
- Username: `admin`
- Password: `admin`

**Backend API:** http://localhost:8098
- CORS-enabled for cross-origin requests
- Token stored in localStorage + cookies

### 2. Tenant Admin Frontend (Port 3002)

**Purpose:** Multi-tenant management per tenant

**Repository:** https://github.com/anupamdutta5/tenant-admin-frontend

**Key Features:**
- Tenant-specific dashboard
- Component status management
- Incident management
- Subscriber management
- Team & RBAC management
- Multi-tenant subdomain routing

**Access:** http://{subdomain}.localhost:3002
- Example: http://anupam.localhost:3002
- Example: http://monday.localhost:3002

**Backend API:** http://localhost:8099
- Same-origin requests via subdomain routing
- Session cookie + JWT authentication

---

## Quick Start

### 1. Install Dependencies

```bash
# SaaS Admin Frontend
cd microservices/saas-admin-frontend
npm install

# Tenant Admin Frontend
cd microservices/tenant-admin-frontend
npm install
```

### 2. Configure Subdomain DNS (Tenant Admin Only)

Edit `/etc/hosts` (macOS/Linux) or `C:\Windows\System32\drivers\etc\hosts` (Windows):

```
127.0.0.1 anupam.localhost
127.0.0.1 monday.localhost
```

### 3. Start Services

**Option A: Start All Services**
```bash
cd microservices
./start-all-services.sh
```

**Option B: Start Frontends Only**
```bash
cd microservices
./start-all-frontends.sh
```

**Option C: Start Individually**
```bash
# SaaS Admin Frontend
cd saas-admin-frontend
npm run dev

# Tenant Admin Frontend
cd tenant-admin-frontend
npm run dev
```

### 4. Verify

```bash
# SaaS Admin Frontend
curl -I http://localhost:3001
# Expected: HTTP/1.1 307 Temporary Redirect (middleware working)

# Tenant Admin Frontend
curl -I http://anupam.localhost:3002
# Expected: HTTP/1.1 307 Temporary Redirect (middleware working)
```

---

## Technology Stack

### Core Framework
- **Next.js 14** - React framework with SSR, routing, and middleware
- **React 18** - UI library with hooks and suspense
- **TypeScript 5** - Type-safe development

### UI & Styling
- **Tailwind CSS 3** - Utility-first CSS framework
- **shadcn/ui** - Headless UI components
- **Radix UI** - Primitive components (Dialog, Select, Dropdown, etc.)
- **Lucide React** - Icon library

### Data & State Management
- **TanStack Query (React Query) v5** - Server state management
  - Data fetching, caching, synchronization
  - Automatic refetching and background updates
  - Optimistic updates
- **Zustand** - Client state management (SaaS Admin)
- **React Context API** - Authentication context (Tenant Admin)

### HTTP & API
- **Axios** - HTTP client with interceptors
- **Native Fetch** - For lightweight requests

### Development Tools
- **ESLint** - Code linting
- **PostCSS** - CSS processing
- **Autoprefixer** - CSS vendor prefixes

---

## Development Guide

### Project Structure

```
saas-admin-frontend/          tenant-admin-frontend/
├── app/                      ├── app/
│   ├── page.tsx             │   ├── page.tsx (root redirect)
│   ├── layout.tsx           │   ├── layout.tsx (global layout)
│   ├── providers.tsx        │   ├── providers.tsx (React Query)
│   ├── login/               │   ├── login/ (auth page)
│   │   └── page.tsx        │   │   └── page.tsx
│   └── admin/               │   └── admin/ (protected routes)
│       ├── dashboard/       │       ├── dashboard/
│       ├── tenants/         │       ├── components/
│       ├── pricing/         │       ├── incidents/
│       └── users/           │       ├── subscribers/
│                            │       └── teams/
├── components/              ├── components/
│   ├── ui/ (shadcn)        │   ├── ui/ (shadcn)
│   ├── auth/               │   ├── auth/
│   └── dashboard/          │   └── dashboard/
├── lib/                    ├── lib/
│   ├── api/                │   ├── api/ (API clients)
│   ├── hooks/              │   ├── hooks/ (custom hooks)
│   ├── contexts/           │   ├── contexts/ (React context)
│   └── utils.ts            │   └── utils.ts
├── types/                  ├── types/
│   └── api.ts              │   └── api.ts (TypeScript types)
├── middleware.ts (auth)    ├── middleware.ts (auth + subdomain)
├── next.config.mjs         ├── next.config.mjs
├── package.json            ├── package.json
├── tsconfig.json           ├── tsconfig.json
└── .env.local              └── .env.local
```

### Key Files

#### `middleware.ts` - Authentication & Routing

**Purpose:** Server-side authentication check before any page renders

**Features:**
- Checks cookie for JWT token
- Routes `/` based on auth state
- Protects `/admin/*` routes
- Redirects login page if already authenticated
- Tenant Admin: Logs subdomain for debugging

**Example:**
```typescript
export function middleware(request: NextRequest) {
  const token = request.cookies.get('admin_token')?.value;
  const { pathname } = request.nextUrl;

  // Root route: redirect based on auth
  if (pathname === '/') {
    return token
      ? NextResponse.redirect(new URL('/admin/dashboard', request.url))
      : NextResponse.redirect(new URL('/login', request.url));
  }

  // Protected routes: require auth
  if (pathname.startsWith('/admin') && !token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}
```

#### `app/providers.tsx` - Global Providers

**Purpose:** Wrap app with React Query and other providers

```typescript
'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000, // 1 minute
        refetchOnWindowFocus: false,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

#### `lib/api/client.ts` - HTTP Client

**Features:**
- Axios instance with base URL
- Request/response interceptors
- Automatic JWT token injection
- Error handling

```typescript
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: Add JWT token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: Handle 401 errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear auth and redirect to login
      localStorage.removeItem('admin_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

### Environment Configuration

**SaaS Admin Frontend (`.env.local`):**
```bash
NEXT_PUBLIC_API_URL=http://localhost:8098/api/v1
```

**Tenant Admin Frontend (`.env.local`):**
```bash
# Empty for same-origin requests (subdomain routing)
NEXT_PUBLIC_API_URL=
```

---

## Authentication & Security

### Authentication Flow

#### SaaS Admin Frontend

1. **Login:**
   - User submits credentials to `/login`
   - Frontend calls `POST /api/v1/auth/login` (backend port 8098)
   - Backend returns JWT token
   - Frontend stores token in:
     - `localStorage` (for React components)
     - `cookie` (for middleware)

2. **Protected Routes:**
   - Middleware checks cookie on every request
   - If no token: redirect to `/login`
   - If token exists: allow access

3. **API Requests:**
   - Axios interceptor adds `Authorization: Bearer {token}` header
   - Backend validates JWT on each request

#### Tenant Admin Frontend

**Same flow as SaaS Admin, plus:**
- Cookie name: `tenant_admin_token` (not `admin_token`)
- Subdomain detection in middleware
- Tenant context from JWT claims

### Middleware-Based Authentication

**Why Middleware?**
- **Server-side execution:** Runs before page render
- **Production-ready:** Industry standard for Next.js auth
- **Better UX:** No flash of unauthenticated content
- **SEO-friendly:** Proper redirects for search engines

**Middleware Configuration:**
```typescript
export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
```

### Cookie Management

**Development Mode:**
```typescript
const cookieAttributes = 'samesite=lax'; // HTTP-friendly
document.cookie = `admin_token=${token}; path=/; max-age=${7 * 24 * 60 * 60}; ${cookieAttributes}`;
```

**Production Mode:**
```typescript
const cookieAttributes = 'secure; samesite=strict'; // HTTPS required
document.cookie = `admin_token=${token}; path=/; max-age=${7 * 24 * 60 * 60}; ${cookieAttributes}`;
```

**Future:** httpOnly cookies (requires backend implementation)
- See [docs/testing/SECURITY_ROADMAP.md](docs/testing/SECURITY_ROADMAP.md)

---

## Subdomain Multi-Tenancy

### How It Works

**Tenant Admin Frontend** uses subdomain routing for multi-tenancy:

```
http://anupam.localhost:3002 → Tenant "anupam"
http://monday.localhost:3002 → Tenant "monday"
```

### DNS Configuration

**Development:**
Add to `/etc/hosts`:
```
127.0.0.1 anupam.localhost
127.0.0.1 monday.localhost
127.0.0.1 {any-tenant-slug}.localhost
```

**Production:**
Configure wildcard DNS:
```
*.yourdomain.com → Frontend Server IP
```

### Subdomain Detection

**Middleware logs subdomain:**
```typescript
const { pathname, hostname } = request.nextUrl;
console.log(`[Middleware] [${hostname}] ${pathname}`);
```

**API requests include subdomain:**
```typescript
// Tenant context is in JWT token (tenant_id claim)
// Backend extracts tenant from token, not hostname
```

### Launching Tenant Admin from SaaS Admin

**Tenant Launch Button:**
```typescript
const launchTenantAdmin = (tenant: Tenant) => {
  const subdomain = tenant.subdomain || tenant.name.toLowerCase().replace(/\s+/g, '-');
  const url = `http://${subdomain}.localhost:3002`;
  window.open(url, '_blank');
};
```

**Fixed in:** [saas-admin-frontend/app/admin/tenants/page.tsx:103-107](saas-admin-frontend/app/admin/tenants/page.tsx#L103-L107)

---

## API Integration

### React Query Patterns

**Custom Hooks:**
```typescript
// lib/hooks/use-tenants.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tenantsApi } from '@/lib/api/tenants';

export function useTenants() {
  return useQuery({
    queryKey: ['tenants'],
    queryFn: () => tenantsApi.getAll(),
  });
}

export function useCreateTenant() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: tenantsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] });
    },
  });
}
```

**Usage in Components:**
```typescript
'use client';

import { useTenants, useCreateTenant } from '@/lib/hooks/use-tenants';

export default function TenantsPage() {
  const { data, isLoading, error } = useTenants();
  const createTenant = useCreateTenant();

  const handleCreate = async (formData) => {
    await createTenant.mutateAsync(formData);
  };

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {data.data.map(tenant => (
        <div key={tenant.id}>{tenant.name}</div>
      ))}
    </div>
  );
}
```

### Data Transformation

**Backend returns JSON strings, frontend needs arrays:**

```typescript
// lib/api/plans.ts
function parsePlanFeatures(plan: any): Plan {
  if (plan.features && typeof plan.features === 'string') {
    try {
      plan.features = JSON.parse(plan.features);
    } catch (e) {
      console.error('Failed to parse plan features:', e);
      plan.features = [];
    }
  }
  return plan as Plan;
}

export const plansApi = {
  getAll: async () => {
    const response = await get<PlansResponse>('/api/v1/plans');
    if (response.plans) {
      response.plans = response.plans.map(parsePlanFeatures);
    }
    return response;
  },

  create: async (data: Partial<Plan>) => {
    const payload = { ...data };
    if (payload.features && Array.isArray(payload.features)) {
      payload.features = JSON.stringify(payload.features) as any;
    }
    return post('/api/v1/plans', payload);
  },
};
```

---

## State Management

### SaaS Admin Frontend

**React Query** for server state + **Zustand** for client state

```typescript
// lib/store/auth-store.ts
import { create } from 'zustand';

interface AuthState {
  user: User | null;
  setUser: (user: User | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  logout: () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_user');
    document.cookie = 'admin_token=; path=/; max-age=0';
    set({ user: null });
  },
}));
```

### Tenant Admin Frontend

**React Query** for server state + **Context API** for auth

```typescript
// lib/contexts/auth-context.tsx
'use client';

import { createContext, useContext, useState } from 'react';

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);

  const logout = () => {
    localStorage.removeItem('tenant_admin_token');
    document.cookie = 'tenant_admin_token=; path=/; max-age=0';
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, setUser, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
};
```

---

## Deployment

### Docker Build

**SaaS Admin Frontend:**
```bash
cd saas-admin-frontend
docker build -t saas-admin-frontend:latest .
docker run -p 3001:3001 saas-admin-frontend:latest
```

**Tenant Admin Frontend:**
```bash
cd tenant-admin-frontend
docker build -t tenant-admin-frontend:latest .
docker run -p 3002:3002 tenant-admin-frontend:latest
```

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm run start
```

### Environment Variables (Production)

**SaaS Admin:**
```bash
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
NODE_ENV=production
```

**Tenant Admin:**
```bash
NEXT_PUBLIC_API_URL=  # Empty for same-origin
NODE_ENV=production
```

### Kubernetes Deployment

See [DEPLOYMENT_GUIDE.md](../DEPLOYMENT_GUIDE.md) for complete Kubernetes configuration.

---

## Troubleshooting

### Port Already in Use

```bash
# Find process
lsof -i :3001

# Kill process
kill -9 <PID>
```

### Frontend Won't Start

```bash
# Clear cache and reinstall
rm -rf .next node_modules
npm install
npm run dev
```

### Authentication Not Persisting

**Symptoms:** User logged out when navigating to root URL

**Solution:**
1. Check `middleware.ts` exists in frontend root
2. Verify cookie is being set (DevTools → Application → Cookies)
3. Check cookie name matches:
   - SaaS Admin: `admin_token`
   - Tenant Admin: `tenant_admin_token`
4. Clear all cookies and re-login

### Subdomain Not Working

**Symptoms:** http://anupam.localhost:3002 doesn't load

**Solution:**
1. Check `/etc/hosts` has entry: `127.0.0.1 anupam.localhost`
2. Restart DNS: `sudo killall -HUP mDNSResponder` (macOS)
3. Try incognito mode (clear DNS cache)
4. Check tenant exists in database:
   ```bash
   psql -U postgres -d tenant_admin_db -c "SELECT slug, subdomain FROM tenants;"
   ```

### CORS Errors (SaaS Admin Only)

**Symptoms:** API calls blocked by CORS policy

**Solution:**
1. Check backend CORS middleware allows `http://localhost:3001`
2. Verify `NEXT_PUBLIC_API_URL` in `.env.local`
3. Restart both frontend and backend

### Features Not Displaying (Pricing Page)

**Symptoms:** `TypeError: plan.features.map is not a function`

**Solution:**
1. Check `lib/api/plans.ts` has `parsePlanFeatures()` function
2. Verify transformation applied in all API methods
3. See [docs/testing/TESTING_GUIDE.md](docs/testing/TESTING_GUIDE.md#pricing-page-testing)

---

## Testing

### Manual Testing

See complete testing guide: [docs/testing/TESTING_GUIDE.md](docs/testing/TESTING_GUIDE.md)

**Quick Tests:**
```bash
# Test root URL redirect (should stay logged in)
1. Login at http://localhost:3001/login
2. Navigate to http://localhost:3001/
3. Should redirect to /admin/dashboard (NOT logout)

# Test tenant launch button
1. Go to http://localhost:3001/admin/tenants
2. Click "Launch Tenant Admin" on any tenant
3. Should open http://{subdomain}.localhost:3002 (frontend, not backend)
```

### End-to-End Testing

See: [docs/testing/TESTING_GUIDE.md#end-to-end-flow-testing](docs/testing/TESTING_GUIDE.md#end-to-end-flow-testing)

---

## Related Documentation

- **[QUICK_START.md](./QUICK_START.md)** - Fast setup guide
- **[docs/testing/TESTING_GUIDE.md](./docs/testing/TESTING_GUIDE.md)** - Complete testing guide
- **[docs/testing/SECURITY_ROADMAP.md](./docs/testing/SECURITY_ROADMAP.md)** - Security hardening plan
- **[AUTHENTICATION_GUIDE.md](../AUTHENTICATION_GUIDE.md)** - Auth & session management
- **[SERVICE_CATALOG.md](../SERVICE_CATALOG.md)** - All services reference

---

## Frontend Repositories

- **SaaS Admin Frontend:** https://github.com/anupamdutta5/saas-admin-frontend
- **Tenant Admin Frontend:** https://github.com/anupamdutta5/tenant-admin-frontend

---

**Frontend Guide Complete!** 🎉

For questions or issues, refer to the troubleshooting section or check the related documentation above.
