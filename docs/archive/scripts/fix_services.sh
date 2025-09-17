#!/bin/bash

echo "Fixing service build issues..."

# Function to fix a service's main.go file
fix_service() {
    local service_path="$1"
    local main_file="$service_path/cmd/main.go"

    if [[ ! -f "$main_file" ]]; then
        echo "Warning: $main_file not found"
        return
    fi

    echo "Processing $service_path..."

    # Remove undefined service initializations that don't exist
    sed -i '' '/monitoringService := services\.NewMonitoringService/d' "$main_file"
    sed -i '' '/authService := services\.NewAuthService/d' "$main_file"
    sed -i '' '/templateService := services\.NewTemplateService/d' "$main_file"
    sed -i '' '/channelService := services\.NewChannelService/d' "$main_file"
    sed -i '' '/subscriptionService := services\.NewSubscriptionService/d' "$main_file"
    sed -i '' '/invoiceService := services\.NewInvoiceService/d' "$main_file"
    sed -i '' '/billingService := services\.NewBillingService/d' "$main_file"
    sed -i '' '/analyticsService := services\.NewAnalyticsService/d' "$main_file"
    sed -i '' '/fraudDetectionService := services\.NewFraudDetectionService/d' "$main_file"
    sed -i '' '/stripeService := services\.NewStripeService/d' "$main_file"
    sed -i '' '/payPalService := services\.NewPayPalService/d' "$main_file"
    sed -i '' '/razorpayService := services\.NewRazorpayService/d' "$main_file"

    # Remove cache.Close() calls since Cache interface doesn't have Close method
    sed -i '' '/cache\.Close()/d' "$main_file"

    # Remove undefined handler method calls
    sed -i '' '/handler\.Readiness/d' "$main_file"
    sed -i '' '/handler\.Liveness/d' "$main_file"

    # Fix undefined resilience methods
    sed -i '' '/resilience\.MetricsHandler()/d' "$main_file"

    # Fix handler constructor calls - remove extra parameters
    sed -i '' 's/handlers\.NewAnalyticsHandler.*/analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewNotificationHandler.*/notificationHandler := handlers.NewNotificationHandler(notificationService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewPaymentHandler.*/paymentHandler := handlers.NewPaymentHandler(paymentService, logger)/' "$main_file"
    sed -i '' 's/handlers\.NewBrandingHandler.*/brandingHandler := handlers.NewBrandingHandler(brandingService, logger)/' "$main_file"

    # Fix service constructor calls with wrong parameter counts
    sed -i '' 's/brandingService := services\.NewBrandingService.*/brandingService := services.NewBrandingService(dbManager.GetDB(), logger)/' "$main_file"

    # Remove undefined handler method calls
    sed -i '' '/handler\.PrometheusMetrics/d' "$main_file"
    sed -i '' '/handler\.GetPublicUptime/d' "$main_file"
    sed -i '' '/handler\.HandleUnsubscribe/d' "$main_file"
    sed -i '' '/handler\.ShowUnsubscribePage/d' "$main_file"
    sed -i '' '/handler\.HandleBounce/d' "$main_file"
    sed -i '' '/handler\.HandleComplaint/d' "$main_file"
    sed -i '' '/handler\.ScheduleNotification/d' "$main_file"
    sed -i '' '/handler\.CancelNotification/d' "$main_file"
    sed -i '' '/handler\.StripeConnectWebhook/d' "$main_file"

    echo "Fixed $service_path"
}

# Fix all services
for service_dir in microservices/*/; do
    if [[ -f "$service_dir/cmd/main.go" ]]; then
        fix_service "$service_dir"
    fi
done

echo "Service fixes completed."