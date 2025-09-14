#!/bin/bash

# Add logger initialization to all notification service test functions

echo "Adding logger initialization to notification service tests..."

# Add logger initialization to all test functions that don't have it
sed -i '' 's/func TestNotificationHandler_\([^(]*\)(t \*testing\.T) {/func TestNotificationHandler_\1(t *testing.T) {/' microservices/notification-service/tests/unit/notification_test.go

# Add logger initialization after gin.SetMode(gin.TestMode) in each test function
sed -i '' '/gin\.SetMode(gin\.TestMode)/a\
\
	// Setup\
	logger, _ := zap.NewDevelopment()\
' microservices/notification-service/tests/unit/notification_test.go

echo "Logger initialization added to notification service tests!"
