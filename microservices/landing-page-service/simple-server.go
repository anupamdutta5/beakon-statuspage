package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	// Simple HTTP server that serves the landing page without database dependency
	http.HandleFunc("/", landingPageHandler)
	http.HandleFunc("/health", healthHandler)
	
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))
	
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8100"
	}
	
	log.Printf("Starting simple landing page server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	// Load the HTML template
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Prepare template data
	data := map[string]interface{}{
		"SiteName":        "StatusPage Pro",
		"SiteDescription": "Professional status page platform for modern teams",
		"SiteURL":         "https://statuspage.pro",
		"SiteKeywords":    []string{"status page", "uptime monitoring", "incident management", "team communication"},
		"ContactEmail":    "hello@statuspage.pro",
		"SupportEmail":    "support@statuspage.pro",
		"SocialLinks": map[string]string{
			"twitter":  "https://twitter.com/statuspagepro",
			"linkedin": "https://linkedin.com/company/statuspagepro",
			"github":   "https://github.com/statuspagepro",
		},
		"AnalyticsID": "",
		"OGImage":     "/static/images/og-image.png",
		"Favicon":     "/static/images/favicon.ico",
		"CustomCSS":   "",
		"CustomJS":    "",
	}
	
	// Execute template
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"landing-page-service"}`))
}
