package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// MockTenantAdminHandler handles tenant admin requests without database
type MockTenantAdminHandler struct {
	templates *template.Template
}

// NewMockTenantAdminHandler creates a new mock handler
func NewMockTenantAdminHandler() *MockTenantAdminHandler {
	return &MockTenantAdminHandler{}
}

// loadTemplates attempts to load templates from various paths
func (h *MockTenantAdminHandler) loadTemplates() error {
	templatePaths := []string{
		"web/templates/*.html",
		"./web/templates/*.html",
		"../web/templates/*.html",
		"./microservices/tenant-admin-service/web/templates/*.html",
	}

	for _, path := range templatePaths {
		if templates, err := template.ParseGlob(path); err == nil {
			if len(templates.Templates()) > 0 {
				h.templates = templates
				log.Printf("✅ Templates loaded successfully from: %s", path)
				return nil
			}
		}
	}

	// Create a simple in-memory template if file loading fails
	tmpl := template.New("admin_dashboard")
	h.templates = tmpl

	// Read the admin dashboard template file directly
	adminDashboardPath := "./web/templates/admin_dashboard.html"
	if _, err := os.Stat(adminDashboardPath); err == nil {
		content, err := os.ReadFile(adminDashboardPath)
		if err == nil {
			h.templates, _ = template.New("admin_dashboard.html").Parse(string(content))
			log.Printf("✅ Admin dashboard template loaded from: %s", adminDashboardPath)
			return nil
		}
	}

	log.Println("⚠️  No templates found, using embedded template")
	return nil
}

// HealthCheck returns service health status
func (h *MockTenantAdminHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "tenant-admin-service",
		"version": "1.0.0-mock",
	})
}

// GetAdminDashboard serves the admin dashboard page
func (h *MockTenantAdminHandler) GetAdminDashboard(c *gin.Context) {
	// Try to render with template
	if h.templates != nil {
		if tmpl := h.templates.Lookup("admin_dashboard.html"); tmpl != nil {
			data := gin.H{
				"Title":       "Tenant Admin Dashboard",
				"Description": "Manage your status page and services",
				"User": gin.H{
					"Name":  c.Query("user_name"),
					"Email": c.Query("user_email"),
					"Plan":  c.Query("plan"),
				},
				"Company":   c.Query("company"),
				"TenantURL": c.Query("tenant_url"),
			}
			tmpl.Execute(c.Writer, data)
			return
		}
	}

	// Fallback: render basic HTML
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tenant Admin Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        .sidebar {
            min-height: 100vh;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
        }
        .sidebar .nav-link {
            color: rgba(255, 255, 255, 0.8);
            border-radius: 0.5rem;
            margin: 0.25rem 0;
        }
        .sidebar .nav-link:hover, .sidebar .nav-link.active {
            color: white;
            background-color: rgba(255, 255, 255, 0.1);
        }
        .main-content {
            background-color: #f8f9fa;
            min-height: 100vh;
        }
        .card {
            border: none;
            box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
            border-radius: 0.75rem;
        }
    </style>
</head>
<body>
    <div class="container-fluid">
        <div class="row">
            <!-- Sidebar -->
            <nav class="col-md-3 col-lg-2 d-md-block sidebar collapse">
                <div class="position-sticky pt-3">
                    <div class="text-center mb-4">
                        <h4 class="text-white">Status Admin</h4>
                        <small class="text-white-50">Welcome %s!</small>
                    </div>
                    <ul class="nav flex-column">
                        <li class="nav-item">
                            <a class="nav-link active" href="#dashboard">
                                <i class="fas fa-tachometer-alt me-2"></i>
                                Dashboard
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#status-pages">
                                <i class="fas fa-globe me-2"></i>
                                Status Pages
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#components">
                                <i class="fas fa-cogs me-2"></i>
                                Components
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#incidents">
                                <i class="fas fa-exclamation-triangle me-2"></i>
                                Incidents
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#maintenance">
                                <i class="fas fa-tools me-2"></i>
                                Maintenance
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#subscribers">
                                <i class="fas fa-users me-2"></i>
                                Subscribers
                            </a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link" href="#settings">
                                <i class="fas fa-cog me-2"></i>
                                Settings
                            </a>
                        </li>
                    </ul>
                </div>
            </nav>

            <!-- Main content -->
            <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4 main-content">
                <div class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom">
                    <h1 class="h2">Status Page Admin Dashboard</h1>
                    <div class="btn-toolbar mb-2 mb-md-0">
                        <button type="button" class="btn btn-sm btn-primary">
                            <i class="fas fa-plus me-1"></i>
                            New Status Page
                        </button>
                    </div>
                </div>

                <!-- Overall Status -->
                <div class="row mb-4">
                    <div class="col-12">
                        <div class="card">
                            <div class="card-body">
                                <div class="d-flex align-items-center">
                                    <div class="flex-shrink-0">
                                        <i class="fas fa-check-circle fa-3x text-success"></i>
                                    </div>
                                    <div class="flex-grow-1 ms-3">
                                        <h3 class="card-title mb-1">All Systems Operational</h3>
                                        <p class="card-text text-muted">Welcome to your status page admin panel!</p>
                                        <div class="mt-2">
                                            <strong>User:</strong> %s<br>
                                            <strong>Email:</strong> %s<br>
                                            <strong>Plan:</strong> %s<br>
                                            <strong>Company:</strong> %s<br>
                                            <strong>Status Page:</strong> <a href="%s" target="_blank">%s</a>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Quick Stats -->
                <div class="row mb-4">
                    <div class="col-md-3">
                        <div class="card text-center">
                            <div class="card-body">
                                <i class="fas fa-globe fa-2x text-primary mb-2"></i>
                                <h5 class="card-title">Status Pages</h5>
                                <h3 class="text-primary">1</h3>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="card text-center">
                            <div class="card-body">
                                <i class="fas fa-cogs fa-2x text-info mb-2"></i>
                                <h5 class="card-title">Components</h5>
                                <h3 class="text-info">5</h3>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="card text-center">
                            <div class="card-body">
                                <i class="fas fa-exclamation-triangle fa-2x text-warning mb-2"></i>
                                <h5 class="card-title">Active Incidents</h5>
                                <h3 class="text-warning">0</h3>
                            </div>
                        </div>
                    </div>
                    <div class="col-md-3">
                        <div class="card text-center">
                            <div class="card-body">
                                <i class="fas fa-users fa-2x text-success mb-2"></i>
                                <h5 class="card-title">Subscribers</h5>
                                <h3 class="text-success">0</h3>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Feature Message -->
                <div class="row">
                    <div class="col-12">
                        <div class="card">
                            <div class="card-body text-center py-5">
                                <i class="fas fa-rocket fa-3x text-info mb-3"></i>
                                <h4>🎉 Welcome to Your Status Page Admin!</h4>
                                <p class="lead">You've successfully set up your status page. This is where you can:</p>
                                <div class="row mt-4">
                                    <div class="col-md-4">
                                        <i class="fas fa-plus-circle fa-2x text-primary mb-2"></i>
                                        <h6>Create Incidents</h6>
                                        <p class="small text-muted">Report and manage service incidents</p>
                                    </div>
                                    <div class="col-md-4">
                                        <i class="fas fa-cogs fa-2x text-info mb-2"></i>
                                        <h6>Manage Components</h6>
                                        <p class="small text-muted">Add and monitor your services</p>
                                    </div>
                                    <div class="col-md-4">
                                        <i class="fas fa-paint-brush fa-2x text-success mb-2"></i>
                                        <h6>Customize Branding</h6>
                                        <p class="small text-muted">Make it match your brand</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>`,
		c.Query("user_name"),
		c.Query("user_name"),
		c.Query("user_email"),
		c.Query("plan"),
		c.Query("company"),
		c.Query("tenant_url"),
		c.Query("tenant_url"))
}

func main() {
	// Set the port from environment variable or default to 8099
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8099"
	}

	// Create Gin router
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	// Initialize handler
	handler := NewMockTenantAdminHandler()
	handler.loadTemplates()

	// Setup routes
	router.GET("/health", handler.HealthCheck)
	router.GET("/", handler.GetAdminDashboard)
	router.GET("/admin", handler.GetAdminDashboard)

	// Serve static files if they exist
	if _, err := os.Stat("./web/static"); err == nil {
		router.Static("/static", "./web/static")
	}

	log.Printf("🚀 Starting Mock Tenant Admin Service on port %s...", port)
	log.Printf("📍 Admin dashboard available at: http://localhost:%s/admin", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}