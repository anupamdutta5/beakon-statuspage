/**
 * Service Worker for Progressive Web App functionality
 * Implements caching strategies for optimal performance
 */

const CACHE_VERSION = 'v1.0.0';
const CACHE_NAME = `beakon-cache-${CACHE_VERSION}`;
const RUNTIME_CACHE = 'beakon-runtime';

// Resources to cache immediately
const STATIC_CACHE_URLS = [
    '/',
    '/static/css/modern-landing.css',
    '/static/js/modern-landing.js',
    '/static/js/ab-test.js',
    '/static/images/logo.svg',
    '/offline.html'
];

// Cache strategies
const CACHE_STRATEGIES = {
    // Cache first, fallback to network
    cacheFirst: async (request) => {
        const cache = await caches.open(CACHE_NAME);
        const cachedResponse = await cache.match(request);

        if (cachedResponse) {
            // Update cache in background
            fetch(request).then(response => {
                if (response && response.status === 200) {
                    cache.put(request, response.clone());
                }
            });
            return cachedResponse;
        }

        try {
            const networkResponse = await fetch(request);
            if (networkResponse && networkResponse.status === 200) {
                cache.put(request, networkResponse.clone());
            }
            return networkResponse;
        } catch (error) {
            return caches.match('/offline.html');
        }
    },

    // Network first, fallback to cache
    networkFirst: async (request) => {
        const cache = await caches.open(RUNTIME_CACHE);

        try {
            const networkResponse = await fetch(request);
            if (networkResponse && networkResponse.status === 200) {
                cache.put(request, networkResponse.clone());
            }
            return networkResponse;
        } catch (error) {
            const cachedResponse = await cache.match(request);
            return cachedResponse || caches.match('/offline.html');
        }
    },

    // Network only
    networkOnly: async (request) => {
        try {
            return await fetch(request);
        } catch (error) {
            return new Response('Network error', {
                status: 503,
                statusText: 'Service Unavailable'
            });
        }
    },

    // Stale while revalidate
    staleWhileRevalidate: async (request) => {
        const cache = await caches.open(RUNTIME_CACHE);
        const cachedResponse = await cache.match(request);

        const networkResponsePromise = fetch(request).then(response => {
            if (response && response.status === 200) {
                cache.put(request, response.clone());
            }
            return response;
        });

        return cachedResponse || networkResponsePromise;
    }
};

// Install event - cache static resources
self.addEventListener('install', (event) => {
    console.log('Service Worker installing...');

    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(cache => {
                console.log('Caching static resources');
                return cache.addAll(STATIC_CACHE_URLS);
            })
            .then(() => self.skipWaiting())
    );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
    console.log('Service Worker activating...');

    event.waitUntil(
        caches.keys()
            .then(cacheNames => {
                return Promise.all(
                    cacheNames
                        .filter(cacheName => {
                            return cacheName !== CACHE_NAME && cacheName !== RUNTIME_CACHE;
                        })
                        .map(cacheName => {
                            console.log('Deleting old cache:', cacheName);
                            return caches.delete(cacheName);
                        })
                );
            })
            .then(() => self.clients.claim())
    );
});

// Fetch event - implement caching strategies
self.addEventListener('fetch', (event) => {
    const { request } = event;
    const url = new URL(request.url);

    // Skip non-GET requests
    if (request.method !== 'GET') {
        event.respondWith(fetch(request));
        return;
    }

    // Skip cross-origin requests
    if (url.origin !== location.origin) {
        event.respondWith(fetch(request));
        return;
    }

    // Apply different strategies based on resource type
    if (request.destination === 'image') {
        // Images: cache first
        event.respondWith(CACHE_STRATEGIES.cacheFirst(request));
    } else if (url.pathname.startsWith('/api/')) {
        // API calls: network first with timeout
        event.respondWith(
            Promise.race([
                CACHE_STRATEGIES.networkFirst(request),
                new Promise((resolve) => {
                    setTimeout(() => {
                        resolve(caches.match(request));
                    }, 3000);
                })
            ])
        );
    } else if (url.pathname.startsWith('/static/')) {
        // Static assets: cache first
        event.respondWith(CACHE_STRATEGIES.cacheFirst(request));
    } else if (request.destination === 'document') {
        // HTML pages: stale while revalidate
        event.respondWith(CACHE_STRATEGIES.staleWhileRevalidate(request));
    } else {
        // Default: stale while revalidate
        event.respondWith(CACHE_STRATEGIES.staleWhileRevalidate(request));
    }
});

// Background sync for form submissions
self.addEventListener('sync', (event) => {
    if (event.tag === 'sync-forms') {
        event.waitUntil(syncFormData());
    }
});

async function syncFormData() {
    // Get queued form data from IndexedDB
    const db = await openDB();
    const tx = db.transaction('formQueue', 'readonly');
    const store = tx.objectStore('formQueue');
    const forms = await store.getAll();

    // Attempt to submit each form
    for (const form of forms) {
        try {
            const response = await fetch(form.url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(form.data)
            });

            if (response.ok) {
                // Remove from queue
                const deleteTx = db.transaction('formQueue', 'readwrite');
                await deleteTx.objectStore('formQueue').delete(form.id);
            }
        } catch (error) {
            console.error('Failed to sync form:', error);
        }
    }
}

// Push notifications
self.addEventListener('push', (event) => {
    const options = {
        body: event.data ? event.data.text() : 'New update available',
        icon: '/static/images/icon-192.png',
        badge: '/static/images/badge-72.png',
        vibrate: [200, 100, 200],
        actions: [
            {
                action: 'view',
                title: 'View'
            },
            {
                action: 'close',
                title: 'Close'
            }
        ]
    };

    event.waitUntil(
        self.registration.showNotification('Beakon StatusPage', options)
    );
});

// Notification click handler
self.addEventListener('notificationclick', (event) => {
    event.notification.close();

    if (event.action === 'view') {
        event.waitUntil(
            clients.openWindow('/')
        );
    }
});

// Message handler for cache updates
self.addEventListener('message', (event) => {
    if (event.data.action === 'skipWaiting') {
        self.skipWaiting();
    }

    if (event.data.action === 'clearCache') {
        event.waitUntil(
            caches.keys().then(cacheNames => {
                return Promise.all(
                    cacheNames.map(cacheName => caches.delete(cacheName))
                );
            })
        );
    }
});

// Helper function to open IndexedDB
async function openDB() {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open('BeakonDB', 1);

        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve(request.result);

        request.onupgradeneeded = (event) => {
            const db = event.target.result;
            if (!db.objectStoreNames.contains('formQueue')) {
                db.createObjectStore('formQueue', { keyPath: 'id', autoIncrement: true });
            }
        };
    });
}

// Performance monitoring
self.addEventListener('fetch', (event) => {
    const startTime = Date.now();

    event.waitUntil(
        event.respondWith.then(() => {
            const duration = Date.now() - startTime;

            // Report slow requests
            if (duration > 1000) {
                console.warn(`Slow request: ${event.request.url} took ${duration}ms`);

                // Send to analytics
                if (self.registration.sync) {
                    self.registration.sync.register('report-performance');
                }
            }
        })
    );
});