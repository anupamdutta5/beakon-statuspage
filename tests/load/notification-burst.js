// notification-burst.js
// Load test for notification service
// Simulates burst traffic of 100 notifications per second

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const notificationSendTime = new Trend('notification_send_time');
const deliverySuccessRate = new Rate('delivery_success');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  scenarios: {
    burst_test: {
      executor: 'constant-arrival-rate',
      rate: 100,              // 100 notifications per second
      timeUnit: '1s',
      duration: '5m',         // Run for 5 minutes
      preAllocatedVUs: 50,    // Pre-allocate 50 VUs
      maxVUs: 200,            // Allow up to 200 VUs if needed
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<3000', 'p(99)<5000'], // 95% under 3s, 99% under 5s
    http_req_failed: ['rate<0.05'],                   // Error rate below 5%
    errors: ['rate<0.1'],                             // Custom error rate below 10%
    notification_send_time: ['p(95)<2000'],           // 95% of sends under 2s
    delivery_success: ['rate>0.95'],                  // 95%+ delivery success
  },
};

// Base URLs
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8085';
const tenantIDs = [
  'tenant-1111-1111-1111-111111111111',
  'tenant-2222-2222-2222-222222222222',
  'tenant-3333-3333-3333-333333333333',
];

const channels = ['email', 'slack', 'webhook', 'sms'];
const priorities = ['low', 'medium', 'high'];

// Generate test auth token (in real test, would authenticate)
function getAuthToken(tenantID) {
  return `test-token-${tenantID}`;
}

export default function () {
  const tenantID = tenantIDs[Math.floor(Math.random() * tenantIDs.length)];
  const authToken = getAuthToken(tenantID);
  const channel = channels[Math.floor(Math.random() * channels.length)];
  const priority = priorities[Math.floor(Math.random() * priorities.length)];

  // Create notification payload
  const payload = JSON.stringify({
    notification_type: 'component_status',
    reference_id: `comp-${Date.now()}`,
    reference_type: 'component',
    priority: priority,
    subject: `Load Test Notification - ${priority}`,
    message: `Component status changed at ${new Date().toISOString()}`,
    channels: [channel],
  });

  // Send notification
  const sendStart = Date.now();
  const response = http.post(
    `${BASE_URL}/api/v1/notifications`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authToken}`,
        'X-Tenant-ID': tenantID,
      },
      tags: { name: 'SendNotification', channel: channel, priority: priority },
    }
  );
  const sendEnd = Date.now();

  requestCount.add(1);
  notificationSendTime.add(sendEnd - sendStart);

  // Check response
  const sendCheck = check(response, {
    'notification created': (r) => r.status === 201,
    'has notification ID': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id && body.id.length > 0;
      } catch (e) {
        return false;
      }
    },
  });

  if (!sendCheck) {
    errorRate.add(1);
    deliverySuccessRate.add(0);
  } else {
    deliverySuccessRate.add(1);
  }

  // Random small delay to simulate realistic burst pattern
  sleep(Math.random() * 0.1);
}

// Setup
export function setup() {
  console.log('Starting notification burst test');
  console.log(`Target URL: ${BASE_URL}`);
  console.log('Rate: 100 notifications/second');
  console.log('Duration: 5 minutes');
  console.log(`Channels: ${channels.join(', ')}`);
}

// Teardown
export function teardown(data) {
  console.log('Notification burst test completed');
  console.log('Check notification service logs and queue depth');
}
