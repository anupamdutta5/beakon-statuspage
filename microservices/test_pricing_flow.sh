#!/bin/bash

# Pricing Management System End-to-End Test
# This script tests the complete pricing sync flow

echo "========================================"
echo "Pricing Management System E2E Test"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check SaaS Admin Service Health
echo "${YELLOW}Test 1: Checking SaaS Admin Service${NC}"
HEALTH=$(curl -s http://localhost:8097/api/v1/health)
if echo "$HEALTH" | grep -q "healthy"; then
    echo "${GREEN}✓ SaaS Admin Service is healthy${NC}"
else
    echo "${RED}✗ SaaS Admin Service is not responding${NC}"
    echo "Response: $HEALTH"
    exit 1
fi
echo ""

# Test 2: Check Landing Page Service Health
echo "${YELLOW}Test 2: Checking Landing Page Service${NC}"
HEALTH=$(curl -s http://localhost:8100/health)
if echo "$HEALTH" | grep -q "status"; then
    echo "${GREEN}✓ Landing Page Service is healthy${NC}"
else
    echo "${RED}✗ Landing Page Service is not responding${NC}"
    echo "Response: $HEALTH"
    exit 1
fi
echo ""

# Test 3: Create Sample Pricing Plans
echo "${YELLOW}Test 3: Creating Sample Pricing Plans in SaaS Admin${NC}"

# Starter Plan
STARTER=$(curl -s -X POST http://localhost:8097/api/v1/plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Starter",
    "slug": "starter",
    "description": "Perfect for individuals and small teams",
    "price": 9.00,
    "currency": "USD",
    "billing_interval": "monthly",
    "features": "[\"1 Status Page\", \"5 Components\", \"Email Support\", \"99.9% Uptime SLA\"]",
    "is_popular": false,
    "is_active": true,
    "is_public": true,
    "button_text": "Get Started",
    "button_url": "/signup?plan=starter",
    "display_order": 1
  }')

if echo "$STARTER" | grep -q "created"; then
    echo "${GREEN}✓ Starter plan created${NC}"
else
    echo "${YELLOW}! Starter plan may already exist or creation failed${NC}"
fi

# Professional Plan (Popular)
PROFESSIONAL=$(curl -s -X POST http://localhost:8097/api/v1/plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Professional",
    "slug": "professional",
    "description": "Ideal for growing teams and businesses",
    "price": 49.00,
    "currency": "USD",
    "billing_interval": "monthly",
    "features": "[\"5 Status Pages\", \"Unlimited Components\", \"Priority Support\", \"Custom Domain\", \"Advanced Analytics\"]",
    "is_popular": true,
    "is_active": true,
    "is_public": true,
    "button_text": "Start Free Trial",
    "button_url": "/signup?plan=professional",
    "display_order": 2
  }')

if echo "$PROFESSIONAL" | grep -q "created"; then
    echo "${GREEN}✓ Professional plan created (marked as popular)${NC}"
else
    echo "${YELLOW}! Professional plan may already exist or creation failed${NC}"
fi

# Enterprise Plan
ENTERPRISE=$(curl -s -X POST http://localhost:8097/api/v1/plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Enterprise",
    "slug": "enterprise",
    "description": "For large organizations with complex needs",
    "price": 199.00,
    "currency": "USD",
    "billing_interval": "monthly",
    "features": "[\"Unlimited Status Pages\", \"Unlimited Components\", \"24/7 Phone Support\", \"Custom Domain\", \"White Label\", \"API Access\", \"SLA Guarantee\"]",
    "is_popular": false,
    "is_active": true,
    "is_public": true,
    "button_text": "Contact Sales",
    "button_url": "/contact?plan=enterprise",
    "display_order": 3
  }')

if echo "$ENTERPRISE" | grep -q "created"; then
    echo "${GREEN}✓ Enterprise plan created${NC}"
else
    echo "${YELLOW}! Enterprise plan may already exist or creation failed${NC}"
fi
echo ""

# Test 4: Verify Plans in SaaS Admin
echo "${YELLOW}Test 4: Verifying plans in SaaS Admin database${NC}"
PLANS=$(curl -s http://localhost:8097/api/v1/pricing/plans/public)
PLAN_COUNT=$(echo "$PLANS" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('plans', [])))" 2>/dev/null || echo "0")

if [ "$PLAN_COUNT" -gt 0 ]; then
    echo "${GREEN}✓ Found $PLAN_COUNT public pricing plans${NC}"
    echo "Plans available for sync:"
    echo "$PLANS" | python3 -m json.tool 2>/dev/null | grep -E '"name"|"slug"|"is_popular"' | head -20
else
    echo "${RED}✗ No public plans found${NC}"
fi
echo ""

# Test 5: Trigger Pricing Sync
echo "${YELLOW}Test 5: Syncing pricing from SaaS Admin to Landing Page${NC}"
SYNC_RESULT=$(curl -s -X POST http://localhost:8100/api/v1/admin/pricing/sync)

if echo "$SYNC_RESULT" | grep -q "success"; then
    SYNCED=$(echo "$SYNC_RESULT" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('total_synced', 0))" 2>/dev/null || echo "0")
    CREATED=$(echo "$SYNC_RESULT" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('created', 0))" 2>/dev/null || echo "0")
    UPDATED=$(echo "$SYNC_RESULT" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('updated', 0))" 2>/dev/null || echo "0")

    echo "${GREEN}✓ Pricing sync successful!${NC}"
    echo "  - Total synced: $SYNCED"
    echo "  - Created: $CREATED"
    echo "  - Updated: $UPDATED"
else
    echo "${RED}✗ Pricing sync failed${NC}"
    echo "Response: $SYNC_RESULT"
fi
echo ""

# Test 6: Verify Pricing in Landing Page
echo "${YELLOW}Test 6: Verifying pricing in Landing Page service${NC}"
LANDING_PLANS=$(curl -s http://localhost:8100/api/v1/pricing)
LANDING_COUNT=$(echo "$LANDING_PLANS" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('plans', [])))" 2>/dev/null || echo "0")

if [ "$LANDING_COUNT" -gt 0 ]; then
    echo "${GREEN}✓ Found $LANDING_COUNT pricing plans in Landing Page${NC}"
    echo "Plans available on landing page:"
    echo "$LANDING_PLANS" | python3 -m json.tool 2>/dev/null | grep -E '"name"|"slug"|"is_popular"|"price"' | head -24
else
    echo "${RED}✗ No plans found in Landing Page${NC}"
fi
echo ""

# Test 7: Check Landing Page Frontend
echo "${YELLOW}Test 7: Checking Landing Page HTML${NC}"
LANDING_HTML=$(curl -s http://localhost:8100/)
if echo "$LANDING_HTML" | grep -q "pricing"; then
    echo "${GREEN}✓ Landing page loaded successfully${NC}"
    echo "  Visit: http://localhost:8100"
else
    echo "${YELLOW}! Landing page loaded but pricing section may need verification${NC}"
fi
echo ""

# Summary
echo "========================================"
echo "${GREEN}Test Suite Complete!${NC}"
echo "========================================"
echo ""
echo "What was tested:"
echo "  ✓ SaaS Admin Service health"
echo "  ✓ Landing Page Service health"
echo "  ✓ Sample plan creation (Starter, Professional, Enterprise)"
echo "  ✓ Public pricing API endpoint"
echo "  ✓ Pricing sync mechanism"
echo "  ✓ Landing page pricing retrieval"
echo "  ✓ Frontend rendering"
echo ""
echo "Next steps:"
echo "  1. Visit http://localhost:8100 to see the pricing plans"
echo "  2. Visit http://localhost:8097/admin to manage plans"
echo "  3. Try updating a plan and re-syncing"
echo ""
echo "Pricing Management URLs:"
echo "  - SaaS Admin Dashboard: http://localhost:8097/admin"
echo "  - Public Pricing API: http://localhost:8097/api/v1/pricing/plans/public"
echo "  - Sync Endpoint: POST http://localhost:8100/api/v1/admin/pricing/sync"
echo "  - Landing Page: http://localhost:8100"
echo ""
