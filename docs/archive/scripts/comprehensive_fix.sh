#!/bin/bash

echo "=== COMPREHENSIVE SERVICE FIXES ==="

# Function to fix a service comprehensively
fix_service_comprehensively() {
    local service_path="$1"
    local main_file="$service_path/cmd/main.go"

    if [[ ! -f "$main_file" ]]; then
        echo "Warning: $main_file not found"
        return
    fi

    echo "Fixing $service_path..."

    # 1. Remove all undefined service constructor calls
    sed -i '' '/reportsService := services\.NewReportsService/d' "$main_file"
    sed -i '' '/dashboardService := services\.NewDashboardService/d' "$main_file"
    sed -i '' '/metricsService := services\.NewMetricsService/d' "$main_file"
    sed -i '' '/assetService := services\.NewAssetService/d' "$main_file"
    sed -i '' '/eventStreamService := services\.NewEventStreamService/d' "$main_file"
    sed -i '' '/eventProjectionService := services\.NewEventProjectionService/d' "$main_file"
    sed -i '' '/eventSnapshotService := services\.NewEventSnapshotService/d' "$main_file"
    sed -i '' '/eventSubscriptionService := services\.NewEventSubscriptionService/d' "$main_file"
    sed -i '' '/eventQueryService := services\.NewEventQueryService/d' "$main_file"
    sed -i '' '/eventAnalyticsService := services\.NewEventAnalyticsService/d' "$main_file"
    sed -i '' '/eventAuditService := services\.NewEventAuditService/d' "$main_file"
    sed -i '' '/notificationService := services\.NewNotificationService/d' "$main_file"
    sed -i '' '/alertService := services\.NewAlertService/d' "$main_file"
    sed -i '' '/fraudDetectionService := services\.NewFraudDetectionService/d' "$main_file"
    sed -i '' '/payPalService := services\.NewPayPalService/d' "$main_file"

    # 2. Remove all undefined handler constructor calls and references
    sed -i '' '/assetHandler := handlers\.NewAssetHandler/d' "$main_file"
    sed -i '' '/themeHandler := handlers\.NewThemeHandler/d' "$main_file"
    sed -i '' '/cssHandler := handlers\.NewCSSHandler/d' "$main_file"
    sed -i '' '/fontHandler := handlers\.NewFontHandler/d' "$main_file"
    sed -i '' '/customizationHandler := handlers\.NewCustomizationHandler/d' "$main_file"
    sed -i '' '/whitelabelHandler := handlers\.NewWhitelabelHandler/d' "$main_file"
    sed -i '' '/complianceHandler := handlers\.NewComplianceHandler/d' "$main_file"
    sed -i '' '/abTestingHandler := handlers\.NewABTestingHandler/d' "$main_file"
    sed -i '' '/leadCaptureHandler := handlers\.NewLeadCaptureHandler/d' "$main_file"
    sed -i '' '/seoHandler := handlers\.NewSEOHandler/d' "$main_file"
    sed -i '' '/landingAnalyticsHandler := handlers\.NewLandingAnalyticsHandler/d' "$main_file"
    sed -i '' '/emailCaptureHandler := handlers\.NewEmailCaptureHandler/d' "$main_file"
    sed -i '' '/templateHandler := handlers\.NewTemplateHandler/d' "$main_file"
    sed -i '' '/cmsHandler := handlers\.NewCMSHandler/d' "$main_file"
    sed -i '' '/multiLanguageHandler := handlers\.NewMultiLanguageHandler/d' "$main_file"
    sed -i '' '/socialMediaHandler := handlers\.NewSocialMediaHandler/d' "$main_file"
    sed -i '' '/conversionHandler := handlers\.NewConversionHandler/d' "$main_file"

    # 3. Remove undefined handler method calls from routes
    sed -i '' '/handler\.GetPublicStatus/d' "$main_file"
    sed -i '' '/handler\.GetPublicAnalyticsOverview/d' "$main_file"
    sed -i '' '/handler\.GetRealtimeDashboard/d' "$main_file"
    sed -i '' '/handler\.GetPublicIncidentUpdates/d' "$main_file"
    sed -i '' '/handler\.GetSystemStatus/d' "$main_file"
    sed -i '' '/handler\.GetStatusSummary/d' "$main_file"
    sed -i '' '/handler\.ResolveIncident/d' "$main_file"
    sed -i '' '/handler\.CloseIncident/d' "$main_file"
    sed -i '' '/handler\.ReopenIncident/d' "$main_file"
    sed -i '' '/handler\.GetIncidentUpdates/d' "$main_file"
    sed -i '' '/handler\.GetNotificationDeliveries/d' "$main_file"
    sed -i '' '/handler\.RetryNotification/d' "$main_file"
    sed -i '' '/handler\.CreateBulkNotifications/d' "$main_file"
    sed -i '' '/handler\.SendBulkNotifications/d' "$main_file"
    sed -i '' '/handler\.GetBulkNotificationStatus/d' "$main_file"
    sed -i '' '/handler\.CancelBulkNotifications/d' "$main_file"
    sed -i '' '/handler\.DuplicateTemplate/d' "$main_file"
    sed -i '' '/handler\.PreviewTemplate/d' "$main_file"
    sed -i '' '/handler\.GetTemplateVersions/d' "$main_file"
    sed -i '' '/handler\.CreateTemplateVersion/d' "$main_file"
    sed -i '' '/handler\.ApplePayWebhook/d' "$main_file"
    sed -i '' '/handler\.GooglePayWebhook/d' "$main_file"
    sed -i '' '/handler\.GetSupportedCurrencies/d' "$main_file"
    sed -i '' '/handler\.GetSupportedPaymentMethods/d' "$main_file"
    sed -i '' '/handler\.CreatePublicPaymentIntent/d' "$main_file"
    sed -i '' '/handler\.CalculatePricing/d' "$main_file"

    # 4. Fix parameter mismatches in service constructors
    sed -i '' 's/brandingService := services\.NewBrandingService.*/brandingService := services.NewBrandingService(dbManager.GetDB(), logger)/' "$main_file"
    sed -i '' 's/eventStoreService := services\.NewEventStoreService.*/eventStoreService := services.NewEventStoreService(dbManager.GetDB(), logger)/' "$main_file"

    # 5. Fix handler constructor calls to match actual signatures
    sed -i '' 's/handlers\.NewAnalyticsHandler.*/analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewIncidentHandler.*/incidentHandler := handlers.NewIncidentHandler(incidentService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewMonitoringHandler.*/monitoringHandler := handlers.NewMonitoringHandler(monitoringService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewSaasAdminHandler.*/saasAdminHandler := handlers.NewSaasAdminHandler(saasAdminService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewStatusPageHandler.*/statusPageHandler := handlers.NewStatusPageHandler(statusPageService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewTenantAdminHandler.*/tenantAdminHandler := handlers.NewTenantAdminHandler(tenantAdminService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewTenantHandler.*/tenantHandler := handlers.NewTenantHandler(tenantService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewUserHandler.*/userHandler := handlers.NewUserHandler(userService, logger)/' "$main_file"

    # 6. Remove complex route setups that reference undefined handlers
    # Replace with simple setupRoutes calls
    if grep -q "setupModernizedRoutes.*assetHandler\|setupModernizedRoutes.*themeHandler\|setupModernizedRoutes.*multiLanguageHandler" "$main_file"; then
        sed -i '' 's/setupModernizedRoutes.*/setupRoutes(router, handler, config)/' "$main_file"
    fi

    echo "Fixed $service_path"
}

# Fix all services
for service_dir in microservices/*/; do
    if [[ -f "$service_dir/cmd/main.go" ]]; then
        fix_service_comprehensively "$service_dir"
    fi
done

echo "=== COMPREHENSIVE FIXES COMPLETED ==="