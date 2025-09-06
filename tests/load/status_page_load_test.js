import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
export let errorRate = new Rate('errors');

// Test configuration
export let options = {
  stages: [
    { duration: '2m', target: 10 }, // Ramp up to 10 users
    { duration: '5m', target: 10 }, // Stay at 10 users
    { duration: '2m', target: 20 }, // Ramp up to 20 users
    { duration: '5m', target: 20 }, // Stay at 20 users
    { duration: '2m', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    errors: ['rate<0.1'],             // Custom error rate must be below 10%
  },
};

// Base URL for the application
const BASE_URL = 'http://localhost:8080';

// Test data
const testData = {
  adminEmail: 'admin@test.com',
  adminPassword: 'testpassword',
  tenantName: 'Load Test Tenant',
  serviceName: 'Load Test Service',
  incidentTitle: 'Load Test Incident',
  incidentDescription: 'This is a load test incident',
};

// Helper function to make authenticated requests
function makeAuthenticatedRequest(method, url, payload = null, token = null) {
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  if (token) {
    params.headers['Authorization'] = `Bearer ${token}`;
  }
  
  let response;
  if (method === 'GET') {
    response = http.get(url, params);
  } else if (method === 'POST') {
    response = http.post(url, JSON.stringify(payload), params);
  } else if (method === 'PUT') {
    response = http.put(url, JSON.stringify(payload), params);
  } else if (method === 'DELETE') {
    response = http.del(url, null, params);
  }
  
  return response;
}

// Helper function to login and get token
function login() {
  const loginData = {
    email: testData.adminEmail,
    password: testData.adminPassword,
  };
  
  const response = makeAuthenticatedRequest('POST', `${BASE_URL}/api/v1/admin/login`, loginData);
  
  const success = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login response has token': (r) => r.json('token') !== undefined,
  });
  
  if (success) {
    return response.json('token');
  }
  
  return null;
}

// Test scenarios
export default function() {
  // Scenario 1: Public status page access
  let response = http.get(`${BASE_URL}/api/v1/status`);
  let success = check(response, {
    'public status page accessible': (r) => r.status === 200,
    'public status page response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  if (!success) {
    errorRate.add(1);
  }
  
  sleep(1);
  
  // Scenario 2: Admin login
  const token = login();
  
  if (token) {
    // Scenario 3: Get admin dashboard
    response = makeAuthenticatedRequest('GET', `${BASE_URL}/api/v1/admin/dashboard`, null, token);
    success = check(response, {
      'admin dashboard accessible': (r) => r.status === 200,
      'admin dashboard response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 4: Get services
    response = makeAuthenticatedRequest('GET', `${BASE_URL}/api/v1/admin/services`, null, token);
    success = check(response, {
      'services list accessible': (r) => r.status === 200,
      'services list response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 5: Create a service
    const serviceData = {
      name: `${testData.serviceName} ${__VU}`,
      description: 'Load test service',
      status: 'operational',
      group: 'Load Test Group',
      show_uptime: true,
      position: 0,
    };
    
    response = makeAuthenticatedRequest('POST', `${BASE_URL}/api/v1/admin/services`, serviceData, token);
    success = check(response, {
      'service creation successful': (r) => r.status === 201,
      'service creation response time < 1000ms': (r) => r.timings.duration < 1000,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 6: Get incidents
    response = makeAuthenticatedRequest('GET', `${BASE_URL}/api/v1/admin/incidents`, null, token);
    success = check(response, {
      'incidents list accessible': (r) => r.status === 200,
      'incidents list response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 7: Create an incident
    const incidentData = {
      title: `${testData.incidentTitle} ${__VU}`,
      description: testData.incidentDescription,
      status: 'investigating',
      impact: 'minor',
    };
    
    response = makeAuthenticatedRequest('POST', `${BASE_URL}/api/v1/admin/incidents`, incidentData, token);
    success = check(response, {
      'incident creation successful': (r) => r.status === 201,
      'incident creation response time < 1000ms': (r) => r.timings.duration < 1000,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 8: Get SaaS plans
    response = makeAuthenticatedRequest('GET', `${BASE_URL}/api/v1/admin/saas/plans`, null, token);
    success = check(response, {
      'SaaS plans accessible': (r) => r.status === 200,
      'SaaS plans response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 9: Get tenants
    response = makeAuthenticatedRequest('GET', `${BASE_URL}/api/v1/admin/saas/tenants`, null, token);
    success = check(response, {
      'tenants list accessible': (r) => r.status === 200,
      'tenants list response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
    
    // Scenario 10: Get analytics
    response = makeAuthenticatedRequest('GET', `${BASE_URL}/api/v1/admin/analytics`, null, token);
    success = check(response, {
      'analytics accessible': (r) => r.status === 200,
      'analytics response time < 1000ms': (r) => r.timings.duration < 1000,
    });
    
    if (!success) {
      errorRate.add(1);
    }
    
    sleep(1);
  } else {
    errorRate.add(1);
  }
}

// Setup function (runs once at the beginning)
export function setup() {
  console.log('Starting load test...');
  console.log(`Base URL: ${BASE_URL}`);
  console.log(`Test duration: ${options.stages.reduce((total, stage) => total + stage.duration, 0)}`);
  console.log(`Max concurrent users: ${Math.max(...options.stages.map(stage => stage.target))}`);
}

// Teardown function (runs once at the end)
export function teardown(data) {
  console.log('Load test completed');
  console.log('Results:');
  console.log(`- Total requests: ${data.totalRequests || 'N/A'}`);
  console.log(`- Error rate: ${data.errorRate || 'N/A'}`);
  console.log(`- Average response time: ${data.avgResponseTime || 'N/A'}ms`);
}
