// database-pool-stress.js
// Load test for database connection pool
// Tests connection pool behavior under heavy load

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const queryTime = new Trend('query_time');
const connectionErrors = new Counter('connection_errors');
const timeoutErrors = new Counter('timeout_errors');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  scenarios: {
    // Scenario 1: Gradual ramp up to test pool scaling
    gradual_load: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { duration: '1m', target: 50 },
        { duration: '2m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '1m', target: 300 },  // Exceed typical pool size
        { duration: '2m', target: 300 },  // Sustained high load
        { duration: '1m', target: 50 },
      ],
      exec: 'gradualLoad',
    },
    // Scenario 2: Sudden spike to test pool exhaustion
    spike_test: {
      executor: 'constant-vus',
      vus: 500,
      duration: '30s',
      startTime: '10m',  // Start after gradual load
      exec: 'spikeTest',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<5000', 'p(99)<10000'],
    http_req_failed: ['rate<0.1'],  // Allow 10% failures during extreme load
    errors: ['rate<0.15'],
    query_time: ['p(95)<3000'],
    connection_errors: ['count<100'],  // Less than 100 connection errors total
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8099';
const tenantID = 'tenant-1111-1111-1111-111111111111';

function getAuthToken() {
  return 'test-token';
}

// Scenario 1: Gradual load - Mix of read and write operations
export function gradualLoad() {
  const authToken = getAuthToken();
  const operation = Math.random();

  if (operation < 0.6) {
    // 60% reads - List tenants
    const queryStart = Date.now();
    const response = http.get(
      `${BASE_URL}/api/v1/tenants`,
      {
        headers: {
          'Authorization': `Bearer ${authToken}`,
          'X-Tenant-ID': tenantID,
        },
        tags: { name: 'ListTenants', operation: 'read' },
      }
    );
    const queryEnd = Date.now();

    requestCount.add(1);
    queryTime.add(queryEnd - queryStart);

    const checkResult = check(response, {
      'tenants loaded': (r) => r.status === 200,
    });

    if (!checkResult) {
      errorRate.add(1);
      if (response.status === 500) {
        connectionErrors.add(1);
      }
      if (response.status === 504 || response.status === 408) {
        timeoutErrors.add(1);
      }
    }
  } else if (operation < 0.85) {
    // 25% reads - Get user details (more complex query with joins)
    const queryStart = Date.now();
    const response = http.get(
      `${BASE_URL}/api/v1/users?include=roles,teams`,
      {
        headers: {
          'Authorization': `Bearer ${authToken}`,
          'X-Tenant-ID': tenantID,
        },
        tags: { name: 'ListUsersWithJoins', operation: 'complex_read' },
      }
    );
    const queryEnd = Date.now();

    requestCount.add(1);
    queryTime.add(queryEnd - queryStart);

    const checkResult = check(response, {
      'users loaded': (r) => r.status === 200,
    });

    if (!checkResult) {
      errorRate.add(1);
      if (response.status === 500) {
        connectionErrors.add(1);
      }
    }
  } else {
    // 15% writes - Create component
    const payload = JSON.stringify({
      name: `LoadTest Component ${Date.now()}`,
      description: 'Created during load test',
      status: 'operational',
    });

    const queryStart = Date.now();
    const response = http.post(
      `${BASE_URL}/api/v1/components`,
      payload,
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`,
          'X-Tenant-ID': tenantID,
        },
        tags: { name: 'CreateComponent', operation: 'write' },
      }
    );
    const queryEnd = Date.now();

    requestCount.add(1);
    queryTime.add(queryEnd - queryStart);

    const checkResult = check(response, {
      'component created': (r) => r.status === 201 || r.status === 200,
    });

    if (!checkResult) {
      errorRate.add(1);
      if (response.status === 500) {
        connectionErrors.add(1);
      }
    }
  }

  sleep(Math.random() * 0.5 + 0.2);  // 0.2-0.7s delay
}

// Scenario 2: Spike test - Heavy read load
export function spikeTest() {
  const authToken = getAuthToken();

  // All users hitting read endpoints simultaneously
  const endpoint = Math.random() < 0.5 ? '/api/v1/tenants' : '/api/v1/components';

  const queryStart = Date.now();
  const response = http.get(
    `${BASE_URL}${endpoint}`,
    {
      headers: {
        'Authorization': `Bearer ${authToken}`,
        'X-Tenant-ID': tenantID,
      },
      tags: { name: 'SpikeRead', operation: 'spike' },
    }
  );
  const queryEnd = Date.now();

  requestCount.add(1);
  queryTime.add(queryEnd - queryStart);

  const checkResult = check(response, {
    'spike request successful': (r) => r.status === 200,
  });

  if (!checkResult) {
    errorRate.add(1);
    if (response.status === 500) {
      connectionErrors.add(1);
    }
    if (response.status === 504 || response.status === 408) {
      timeoutErrors.add(1);
    }
  }

  // No sleep - continuous hammering
}

export function setup() {
  console.log('Starting database connection pool stress test');
  console.log(`Target URL: ${BASE_URL}`);
  console.log('Scenario 1: Gradual load (10 → 300 VUs over 9 minutes)');
  console.log('Scenario 2: Spike test (500 VUs for 30 seconds)');
  console.log('Total duration: ~11 minutes');
}

export function teardown(data) {
  console.log('Database pool stress test completed');
  console.log('Metrics to review:');
  console.log('  - Connection pool utilization');
  console.log('  - Connection wait times');
  console.log('  - Query queue depth');
  console.log('  - Database CPU and memory usage');
}
