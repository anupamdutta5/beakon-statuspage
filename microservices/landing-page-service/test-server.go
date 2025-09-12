package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting test server on port 8100...")

	// Simple handler that serves the template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received for:", r.URL.Path)

		// Load template
		tmpl, err := template.ParseFiles("web/templates/index.html")
		if err != nil {
			fmt.Printf("Template error: %v\n", err)
			http.Error(w, "Template error", 500)
			return
		}

		// Simple data
		data := map[string]interface{}{
			"SiteName":        "StatusPage Pro",
			"SiteDescription": "Professional status page platform",
			"SiteURL":         "http://localhost:8100",
			"SiteKeywords":    []string{"status page", "monitoring"},
			"ContactEmail":    "hello@example.com",
			"SupportEmail":    "support@example.com",
			"SocialLinks":     map[string]string{},
			"AnalyticsID":     "",
			"OGImage":         "/static/images/og-image.png",
			"Favicon":         "/static/images/favicon.ico",
			"CustomCSS":       "",
			"CustomJS":        "",
			"CurrentYear":     2025,
			"Hero": map[string]interface{}{
				"Title":        "Keep Your Users Informed, Always",
				"Subtitle":     "Professional status page platform for modern teams",
				"PrimaryCTA":   "Get Started Free",
				"SecondaryCTA": "Learn More",
			},
			"Features": []map[string]interface{}{
				{
					"Title":       "Real-time Monitoring",
					"Description": "Monitor your services with live updates",
					"Icon":        "fas fa-chart-line",
				},
			},
			"PricingPlans": []map[string]interface{}{},
			"Testimonials": []map[string]interface{}{},
			"FAQs":         []map[string]interface{}{},
		}

		// Execute template
		if err := tmpl.Execute(w, data); err != nil {
			fmt.Printf("Template execution error: %v\n", err)
			http.Error(w, "Template execution error", 500)
			return
		}

		fmt.Println("Template served successfully")
	})

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Println("Test server ready at http://localhost:8100")
	log.Fatal(http.ListenAndServe(":8100", nil))
}
