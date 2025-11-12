// incident-spike.js
// Load test for incident service
// Simulates sudden spike of 50 incidents per second

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const incidentCreateTime = new Trend('incident_create_time');
const updateTime = new Trend('incident_update_time');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  scenarios: {
    spike_test: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 25,
      maxVUs: 100,
      stages: [
        { duration: '30s', target: 10 },  // Normal: 10/sec
        { duration: '10s', target: 50 },  // Spike: 50/sec
        { duration: '2m', target: 50 },   // Sustained spike: 50/sec
        { duration: '30s', target: 10 },  // Back to normal: 10/sec
        { duration: '30s', target: 5 },   // Cool down: 5/sec
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<4000', 'p(99)<8000'],
    http_req_failed: ['rate<0.05'],
    errors: ['rate<0.1'],
    incident_create_time: ['p(95)<3000'],
    incident_update_time: ['p(95)<2000'],
  },
};

// Base URLs
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8086';
const tenantIDs = [
  'tenant-1111-1111-1111-111111111111',
  'tenant-2222-2222-2222-222222222222',
];

const severities = ['minor', 'major', 'critical'];
const statuses = ['investigating', 'identified', 'monitoring', 'resolved'];

function getAuthToken(tenantID) {
  return `test-token-${tenantID}`;
}

function getRandomComponentID(tenantID) {
  // In real test, would fetch actual component IDs
  return `comp-${tenantID.substring(0, 8)}-test`;
}

export default function () {
  const tenantID = tenantIDs[Math.floor(Math.random() * tenantIDs.length)];
  const authToken = getAuthToken(tenantID);
  const severity = severities[Math.floor(Math.random() * severities.length)];

  // 70% create incidents, 30% update existing ones
  if (Math.random() < 0.7) {
    // Create incident
    const payload = JSON.stringify({
      title: `Load Test Incident ${Date.now()}`,
      description: `Incident created during load test - ${severity} severity`,
      severity: severity,
      status: 'investigating',
      affected_components: [
        {
          component_id: getRandomComponentID(tenantID),
          impact_level: severity === 'critical' ? 'major_outage' : 'degraded_performance',
        },
      ],
    });

    const createStart = Date.now();
    const response = http.post(
      `${BASE_URL}/api/v1/incidents`,
      payload,
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`,
          'X-Tenant-ID': tenantID,
        },
        tags: { name: 'CreateIncident', severity: severity },
      }
    );
    const createEnd = Date.now();

    requestCount.add(1);
    incidentCreateTime.add(createEnd - createStart);

    const createCheck = check(response, {
      'incident created': (r) => r.status === 201,
      'has incident ID': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.id && body.id.length > 0;
        } catch (e) {
          return false;
        }
      },
    });

    if (!createCheck) {
      errorRate.add(1);
    }
  } else {
    // Update existing incident
    const incidentID = `inc-loadtest-${Math.floor(Math.random() * 100)}`;
    const newStatus = statuses[Math.floor(Math.random() * statuses.length)];

    const payload = JSON.stringify({
      status: newStatus,
      message: `Status updated to ${newStatus} during load test`,
    });

    const updateStart = Date.now();
    const response = http.patch(
      `${BASE_URL}/api/v1/incidents/${incidentID}`,
      payload,
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`,
          'X-Tenant-ID': tenantID,
        },
        tags: { name: 'UpdateIncident', status: newStatus },
      }
    );
    const updateEnd = Date.now();

    requestCount.add(1);
    updateTime.add(updateEnd - updateStart);

    // 404s are okay for update tests (incident might not exist)
    check(response, {
      'update processed': (r) => r.status === 200 || r.status === 404,
    });
  }

  sleep(0.05); // Small delay
}

export function setup() {
  console.log('Starting incident spike test');
  console.log(`Target URL: ${BASE_URL}`);
  console.log('Spike pattern: 10/s → 50/s (spike) → 10/s → 5/s');
  console.log('Duration: ~4 minutes');
}

export function teardown(data) {
  console.log('Incident spike test completed');
  console.log('Check incident service logs and database performance');
}
