image.png# 📋 API Contracts Documentation

## 🎯 Overview

Each microservice now includes its own API contract definition using OpenAPI 3.0.3 specification. This approach ensures:

- **✅ True Independence**: Each service owns its own API contract
- **✅ No Shared Dependencies**: No centralized contract service
- **✅ Co-located Documentation**: Contracts live with the implementation

## 🏗️ Architecture

```
microservices/
├── user-service/
│   ├── api/
│   │   └── user-service-api.yaml    # User Service API Contract
│   ├── cmd/
│   ├── internal/
│   └── ...
├── tenant-service/
│   ├── api/
│   │   └── tenant-service-api.yaml  # Tenant Service API Contract
│   ├── cmd/
│   ├── internal/
│   └── ...
├── payment-service/
│   ├── api/
│   │   └── payment-service-api.yaml # Payment Service API Contract
│   ├── cmd/
│   ├── internal/
│   └── ...
└── ...
```

## 📚 Available API Contracts

### **Core Services**
- **User Service** (`/microservices/user-service/api/user-service-api.yaml`)
  - User management (CRUD operations)
  - Authentication endpoints
  - User role management

- **Tenant Service** (`/microservices/tenant-service/api/tenant-service-api.yaml`)
  - Tenant management (CRUD operations)
  - Tenant settings management
  - Multi-tenancy support

- **Payment Service** (`/microservices/payment-service/api/payment-service-api.yaml`)
  - Payment processing
  - Subscription management
  - Refund handling

### **Status Page Services**
- **Component Service** - Component management
- **Incident Service** - Incident tracking and management
- **Monitoring Service** - System monitoring and alerts

### **Admin Services**
- **SaaS Admin Service** - Platform administration
- **Tenant Admin Service** - Tenant-specific administration

### **Supporting Services**
- **Analytics Service** - Data analytics and reporting
- **Notification Service** - Multi-channel notifications
- **Branding Service** - Custom branding and theming
- **Database Service** - Data persistence and caching
- **Event Store Service** - Event sourcing and CQRS

## 🚀 Usage

### **1. Viewing API Documentation**

Each service's API contract can be viewed using any OpenAPI-compatible tool:

```bash
# Using Swagger UI
npx swagger-ui-serve microservices/user-service/api/user-service-api.yaml

# Using Redoc
npx redoc-cli serve microservices/user-service/api/user-service-api.yaml

# Using OpenAPI Generator
npx @openapitools/openapi-generator-cli generate \
  -i microservices/user-service/api/user-service-api.yaml \
  -g go \
  -o ./generated/user-service-client
```

### **2. Generating Client SDKs**

Generate client SDKs for any language:

```bash
# Generate Go client
openapi-generator generate -i microservices/user-service/api/user-service-api.yaml -g go -o ./clients/go/user-service

# Generate JavaScript client
openapi-generator generate -i microservices/user-service/api/user-service-api.yaml -g javascript -o ./clients/js/user-service

# Generate Python client
openapi-generator generate -i microservices/user-service/api/user-service-api.yaml -g python -o ./clients/python/user-service
```

### **3. API Testing**

Use the contracts for automated testing:

```bash
# Using Dredd for contract testing
dredd microservices/user-service/api/user-service-api.yaml http://localhost:8001

# Using Postman
# Import the YAML file into Postman for testing
```

## 🔧 Service Communication

### **Synchronous Communication (REST APIs)**
Services communicate via well-defined REST APIs:

```go
// Example: User Service calling Tenant Service
type TenantClient struct {
    baseURL string
    client  *http.Client
}

func (c *TenantClient) GetTenant(tenantID string) (*Tenant, error) {
    resp, err := c.client.Get(fmt.Sprintf("%s/tenants/%s", c.baseURL, tenantID))
    // Handle response according to API contract
}
```

### **Asynchronous Communication (Events)**
Services communicate via events for loose coupling:

```go
// Example: Event publishing
type EventPublisher struct {
    // Event publishing implementation
}

func (p *EventPublisher) PublishUserCreated(user *User) error {
    event := Event{
        Type: "user.created",
        Data: user,
    }
    return p.Publish(event)
}
```

## 📋 Contract Standards

### **1. OpenAPI 3.0.3 Specification**
- All contracts use OpenAPI 3.0.3
- Consistent structure across all services
- Full request/response schemas defined

### **2. Error Handling**
Standardized error responses:

```yaml
Error:
  type: object
  required:
    - error
    - message
  properties:
    error:
      type: string
      description: Error code
    message:
      type: string
      description: Error message
    details:
      type: object
      description: Additional error details
```

### **3. Authentication**
- JWT-based authentication
- Consistent auth headers across services
- Role-based access control

### **4. Pagination**
Standardized pagination for list endpoints:

```yaml
parameters:
  - name: limit
    in: query
    schema:
      type: integer
      default: 20
      maximum: 100
  - name: offset
    in: query
    schema:
      type: integer
      default: 0
```

## 🔄 Contract Evolution

### **1. Versioning Strategy**
- Use semantic versioning in API contracts
- Maintain backward compatibility
- Deprecate old versions gracefully

### **2. Breaking Changes**
- Document all breaking changes
- Provide migration guides
- Use deprecation warnings

### **3. Contract Testing**
- Validate contracts against implementations
- Automated contract testing in CI/CD
- Consumer-driven contract testing

## 🛠️ Development Workflow

### **1. API-First Development**
1. Define API contract first
2. Generate server stubs
3. Implement business logic
4. Validate against contract

### **2. Contract Validation**
```bash
# Validate contract syntax
swagger-codegen validate -i microservices/user-service/api/user-service-api.yaml

# Test contract against implementation
dredd microservices/user-service/api/user-service-api.yaml http://localhost:8001
```

### **3. Documentation Generation**
```bash
# Generate HTML documentation
redoc-cli build microservices/user-service/api/user-service-api.yaml

# Generate Markdown documentation
swagger-codegen generate -i microservices/user-service/api/user-service-api.yaml -l markdown
```

## 🎯 Benefits

### **✅ Independence**
- Each service owns its API contract
- No centralized dependencies
- True microservices architecture

### **✅ Maintainability**
- Contracts co-located with code
- Easy to keep in sync
- Clear ownership

### **✅ Developer Experience**
- Auto-generated client SDKs
- Interactive API documentation
- Contract testing tools

### **✅ Quality Assurance**
- Contract validation
- Automated testing
- Clear API boundaries

## 📖 Next Steps

1. **Complete API Contracts**: Add contracts for all remaining services
2. **Client SDK Generation**: Generate SDKs for all services
3. **Contract Testing**: Implement automated contract testing
4. **Documentation**: Generate comprehensive API documentation
5. **Monitoring**: Add API contract monitoring and validation

---

**This approach ensures true microservices independence while maintaining clear, well-defined service boundaries through API contracts!** 🚀

