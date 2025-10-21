# API Integration Guide - Frontend to Go Backend

**Go Backend**: Port 8098 (Gin framework)
**Next.js Frontend**: Port 3000 (development)
**API Proxy**: Configured in `next.config.mjs`

---

## API Endpoints Used

### Authentication

#### POST `/api/v1/auth/login`
```typescript
Request:
{
  username: string;
  password: string;
}

Response:
{
  status: string;
  message: string;
  token?: string;
  user?: {
    id: number;
    username: string;
    email: string;
    role: string;
  };
}
```

**Frontend Usage**:
```typescript
import { authApi } from '@/lib/api/auth';

await authApi.login({ username: 'admin', password: 'password' });
// Automatically stores token in localStorage
// Redirects to /admin/dashboard
```

#### POST `/api/v1/auth/logout`
```typescript
Response:
{
  status: string;
  message: string;
}
```

#### GET `/api/v1/auth/check`
```typescript
Headers: Authorization: Bearer <token>

Response:
{
  status: string;
  authenticated: boolean;
}
```

---

### Tenants (TO BE IMPLEMENTED)

#### GET `/api/v1/tenants`
```typescript
Response:
{
  status: string;
  data: Tenant[];
  count: number;
}
```

#### POST `/api/v1/tenants`
#### GET `/api/v1/tenants/:id`
#### PUT `/api/v1/tenants/:id`
#### DELETE `/api/v1/tenants/:id`

---

## How API Client Works

### Axios Interceptors

**Request Interceptor**:
```typescript
// Automatically adds JWT token to all requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

**Response Interceptor**:
```typescript
// Handles 401 errors by redirecting to login
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('admin_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

---

## Type Safety

All API responses are fully typed:

```typescript
// types/api.ts
export interface Tenant {
  id: string;
  name: string;
  contact_email: string;
  max_users?: number;
  // ... 15+ other fields
}

// Usage in component:
const response = await get<TenantsResponse>('/api/v1/tenants');
// response.data is typed as Tenant[]
```

---

## React Query Integration

```typescript
import { useQuery } from '@tanstack/react-query';
import { get } from '@/lib/api/client';

function TenantsPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['tenants'],
    queryFn: () => get<TenantsResponse>('/api/v1/tenants'),
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return <div>{data.data.length} tenants</div>;
}
```

---

## CORS Configuration

API proxy in `next.config.mjs` handles CORS:

```javascript
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8098/api/:path*',
    },
  ];
}
```

Frontend calls `/api/v1/tenants` → Next.js proxies to `http://localhost:8098/api/v1/tenants`

---

## Error Handling

```typescript
import { getErrorMessage } from '@/lib/api/client';

try {
  await authApi.login(credentials);
} catch (error) {
  const message = getErrorMessage(error);
  // "Invalid credentials" or "Network error" etc.
  setError(message);
}
```

---

## Testing API Integration

```bash
# 1. Start Go backend
cd microservices/saas-admin-service
./saas-admin-service

# 2. Test health endpoint
curl http://localhost:8098/api/v1/health

# 3. Start Next.js frontend
cd frontend
npm run dev

# 4. Visit http://localhost:3000/login
# 5. Login and check Network tab in browser DevTools
```

---

## Security

- JWT tokens stored in localStorage (consider httpOnly cookies for production)
- Authorization header automatically added to all requests
- 401 responses trigger automatic logout
- HTTPS required in production

---

For more details, see:
- [lib/api/client.ts](frontend/lib/api/client.ts) - HTTP client
- [lib/api/auth.ts](frontend/lib/api/auth.ts) - Auth endpoints
- [types/api.ts](frontend/types/api.ts) - Type definitions
