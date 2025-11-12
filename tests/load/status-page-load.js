// status-page-load.js
// Load test for public status page viewing
// Simulates 1000 concurrent users viewing status pages

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const pageLoadTime = new Trend('page_load_time');
const componentFetchTime = new Trend('component_fetch_time');
const uptimeFetchTime = new Trend('uptime_fetch_time');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users over 2 minutes
    { duration: '5m', target: 500 },   // Ramp up to 500 users over 5 minutes
    { duration: '5m', target: 1000 },  // Ramp up to 1000 users over 5 minutes
    { duration: '10m', target: 1000 }, // Stay at 1000 users for 10 minutes
    { duration: '3m', target: 500 },   // Ramp down to 500 users over 3 minutes
    { duration: '2m', target: 0 },     // Ramp down to 0 users over 2 minutes
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000', 'p(99)<5000'], // 95% of requests must complete below 2s, 99% below 5s
    http_req_failed: ['rate<0.01'],                   // Error rate must be below 1%
    errors: ['rate<0.05'],                            // Custom error rate below 5%
    page_load_time: ['p(95)<2000'],                   // 95% of page loads under 2s
    component_fetch_time: ['p(95)<1000'],             // 95% of component fetches under 1s
    uptime_fetch_time: ['p(95)<1500'],                // 95% of uptime fetches under 1.5s
  },
};

// Base URLs
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8093';
const subdomains = ['test1', 'test2', 'test3']; // Test tenants

export default function () {
  // Select random subdomain
  const subdomain = subdomains[Math.floor(Math.random() * subdomains.length)];

  // Test 1: Load status page
  const pageStart = Date.now();
  const pageResponse = http.get(`${BASE_URL}/status/${subdomain}`, {
    tags: { name: 'StatusPageLoad' },
  });
  const pageEnd = Date.now();

  requestCount.add(1);
  pageLoadTime.add(pageEnd - pageStart);

  const pageCheck = check(pageResponse, {
    'status page loads': (r) => r.status === 200,
    'page has title': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.page_title && body.page_title.length > 0;
      } catch (e) {
        return false;
      }
    },
    'page is public': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.is_public === true;
      } catch (e) {
        return false;
      }
    },
  });

  if (!pageCheck) {
    errorRate.add(1);
  }

  sleep(1);

  // Test 2: Fetch components
  const componentStart = Date.now();
  const componentResponse = http.get(`${BASE_URL}/status/${subdomain}/components`, {
    tags: { name: 'ComponentFetch' },
  });
  const componentEnd = Date.now();

  requestCount.add(1);
  componentFetchTime.add(componentEnd - componentStart);

  const componentCheck = check(componentResponse, {
    'components load': (r) => r.status === 200,
    'components is array': (r) => {
      try {
        const body = JSON.parse(r.body);
        return Array.isArray(body);
      } catch (e) {
        return false;
      }
    },
  });

  if (!componentCheck) {
    errorRate.add(1);
  }

  sleep(2);

  // Test 3: Fetch recent incidents
  const incidentResponse = http.get(`${BASE_URL}/status/${subdomain}/incidents?limit=10`, {
    tags: { name: 'IncidentFetch' },
  });

  requestCount.add(1);

  check(incidentResponse, {
    'incidents load': (r) => r.status === 200,
  });

  sleep(2);

  // Test 4: Fetch uptime data (90 days)
  const uptimeStart = Date.now();
  const uptimeResponse = http.get(`${BASE_URL}/status/${subdomain}/uptime?days=90`, {
    tags: { name: 'UptimeFetch' },
  });
  const uptimeEnd = Date.now();

  requestCount.add(1);
  uptimeFetchTime.add(uptimeEnd - uptimeStart);

  const uptimeCheck = check(uptimeResponse, {
    'uptime data loads': (r) => r.status === 200,
    'uptime has overall metric': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.overall_uptime !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (!uptimeCheck) {
    errorRate.add(1);
  }

  sleep(3);

  // Test 5: Fetch branding (occasionally)
  if (Math.random() < 0.3) { // 30% of users check branding
    const brandingResponse = http.get(`${BASE_URL}/status/${subdomain}/branding`, {
      tags: { name: 'BrandingFetch' },
    });

    requestCount.add(1);

    check(brandingResponse, {
      'branding loads': (r) => r.status === 200,
    });

    sleep(1);
  }

  // Test 6: Subscribe to updates (occasionally)
  if (Math.random() < 0.05) { // 5% of users subscribe
    const subscriberEmail = `loadtest-${Date.now()}@example.com`;
    const subscribePayload = JSON.stringify({
      subscription_type: 'email',
      contact_value: subscriberEmail,
      subscribed_incidents: ['critical', 'major'],
    });

    const subscribeResponse = http.post(
      `${BASE_URL}/status/${subdomain}/subscribe`,
      subscribePayload,
      {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'Subscribe' },
      }
    );

    requestCount.add(1);

    check(subscribeResponse, {
      'subscription created': (r) => r.status === 201,
    });

    sleep(1);
  }

  // Simulate user reading time
  sleep(Math.random() * 5 + 3); // 3-8 seconds
}

// Setup function (runs once before test)
export function setup() {
  console.log('Starting status page load test');
  console.log(`Target URL: ${BASE_URL}`);
  console.log(`Test subdomains: ${subdomains.join(', ')}`);
  console.log('Max concurrent users: 1000');
  console.log('Test duration: ~27 minutes');
}

// Teardown function (runs once after test)
export function teardown(data) {
  console.log('Load test completed');
}
