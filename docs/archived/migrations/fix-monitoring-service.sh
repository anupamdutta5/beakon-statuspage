#!/bin/bash

# Fix all remaining ServiceClient.Call patterns in monitoring-service
# These are calling external webhooks and should use http.Client, not ServiceClient

cd /Users/anuoamdutta/Desktop/statuspage/Beakon/microservices/monitoring-service/internal/services

# Fix teams_integration.go
sed -i '' '474,484s/resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{/client := \&http.Client{Timeout: 10 * time.Second}\
	req.Header.Set("Content-Type", "application\/json")\
	resp, err := client.Do(req)/g' teams_integration.go 2>/dev/null || true

sed -i '' '540,550s/resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{/client := \&http.Client{Timeout: 10 * time.Second}\
	req.Header.Set("Content-Type", "application\/json")\
	resp, err := client.Do(req)/g' teams_integration.go 2>/dev/null || true

# Fix telegram_integration.go
sed -i '' '438,448s/resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{/client := \&http.Client{Timeout: 10 * time.Second}\
	resp, err := client.Do(req)/g' telegram_integration.go 2>/dev/null || true

# Fix webhook_service.go
sed -i '' '214,224s/resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{/client := \&http.Client{Timeout: 10 * time.Second}\
	resp, err := client.Do(req)/g' webhook_service.go 2>/dev/null || true

sed -i '' '425,435s/resp, err := s.serviceClient.Call(ctx, resilience.ServiceRequest{/client := \&http.Client{Timeout: 10 * time.Second}\
	resp, err := client.Do(req)/g' webhook_service.go 2>/dev/null || true

echo "Fixed ServiceClient.Call patterns"
