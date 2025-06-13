/**
 * EntityDB PWA Service Worker
 * Offline capabilities and caching strategies
 * Version: v2.30.0+
 */

const CACHE_NAME = 'entitydb-v2.30.0';
const STATIC_CACHE_NAME = 'entitydb-static-v2.30.0';
const API_CACHE_NAME = 'entitydb-api-v2.30.0';

// Resources to cache immediately on install
const STATIC_RESOURCES = [
    '/',
    '/index.html',
    '/manifest.json',
    '/css/design-system.css',
    '/css/entity-browser-enhanced.css',
    '/css/entity-modal-system.css',
    '/css/temporal-query-system.css',
    '/js/logger.js',
    '/js/api-client.js',
    '/js/notification-system.js',
    '/js/advanced-search-system.js',
    '/js/data-export-system.js',
    '/js/temporal-query-system.js',
    '/js/entity-browser-enhanced.js',
    '/js/entity-modal-system.js',
    '/js/relationship-visualization.js',
    '/logo_black.svg',
    '/logo_white.svg'
];

// API endpoints to cache with network-first strategy
const API_ENDPOINTS = [
    '/api/v1/entities/list',
    '/api/v1/entities/query',
    '/api/v1/system/metrics',
    '/api/v1/dashboard/stats',
    '/health',
    '/metrics'
];

// Cache expiration times (in milliseconds)
const CACHE_EXPIRATION = {
    static: 24 * 60 * 60 * 1000,      // 24 hours
    api: 5 * 60 * 1000,               // 5 minutes
    entities: 10 * 60 * 1000,         // 10 minutes
    metrics: 30 * 1000                // 30 seconds
};

// Install event - cache static resources
self.addEventListener('install', event => {
    console.log('EntityDB Service Worker installing...');
    
    event.waitUntil(
        Promise.all([
            // Cache static resources
            caches.open(STATIC_CACHE_NAME).then(cache => {
                console.log('Caching static resources...');
                return cache.addAll(STATIC_RESOURCES);
            }),
            
            // Cache API resources
            caches.open(API_CACHE_NAME).then(cache => {
                console.log('Preparing API cache...');
                return Promise.resolve();
            })
        ]).then(() => {
            console.log('EntityDB Service Worker installed successfully');
            // Force activation of new service worker
            return self.skipWaiting();
        }).catch(error => {
            console.error('EntityDB Service Worker installation failed:', error);
        })
    );
});

// Activate event - clean up old caches
self.addEventListener('activate', event => {
    console.log('EntityDB Service Worker activating...');
    
    event.waitUntil(
        caches.keys().then(cacheNames => {
            return Promise.all(
                cacheNames.map(cacheName => {
                    // Delete old caches
                    if (cacheName !== CACHE_NAME && 
                        cacheName !== STATIC_CACHE_NAME && 
                        cacheName !== API_CACHE_NAME) {
                        console.log('Deleting old cache:', cacheName);
                        return caches.delete(cacheName);
                    }
                })
            );
        }).then(() => {
            console.log('EntityDB Service Worker activated');
            // Take control of all clients
            return self.clients.claim();
        })
    );
});

// Fetch event - handle all network requests
self.addEventListener('fetch', event => {
    const url = new URL(event.request.url);
    
    // Only handle requests to our origin
    if (url.origin !== location.origin) {
        return;
    }
    
    // Route requests based on type
    if (url.pathname.startsWith('/api/')) {
        event.respondWith(handleApiRequest(event.request));
    } else if (isStaticResource(url.pathname)) {
        event.respondWith(handleStaticRequest(event.request));
    } else {
        event.respondWith(handleNavigationRequest(event.request));
    }
});

// Handle API requests with network-first strategy
async function handleApiRequest(request) {
    const url = new URL(request.url);
    const cacheName = API_CACHE_NAME;
    
    try {
        // Always try network first for API requests
        const networkResponse = await fetch(request.clone());
        
        if (networkResponse.ok) {
            // Cache successful responses
            const cache = await caches.open(cacheName);
            const responseToCache = networkResponse.clone();
            
            // Add cache metadata
            const headers = new Headers(responseToCache.headers);
            headers.append('sw-cached-at', Date.now().toString());
            
            const cachedResponse = new Response(responseToCache.body, {
                status: responseToCache.status,
                statusText: responseToCache.statusText,
                headers: headers
            });
            
            cache.put(request, cachedResponse);
        }
        
        return networkResponse;
    } catch (error) {
        console.log('Network failed for API request, trying cache:', url.pathname);
        
        // Network failed, try cache
        const cachedResponse = await caches.match(request);
        
        if (cachedResponse) {
            // Check if cache is still valid
            const cachedAt = cachedResponse.headers.get('sw-cached-at');
            if (cachedAt) {
                const cacheAge = Date.now() - parseInt(cachedAt);
                const maxAge = getCacheMaxAge(url.pathname);
                
                if (cacheAge < maxAge) {
                    console.log('Serving from cache:', url.pathname);
                    return cachedResponse;
                }
            }
        }
        
        // Return offline response for critical endpoints
        if (url.pathname === '/api/v1/entities/list') {
            return createOfflineEntitiesResponse();
        } else if (url.pathname === '/api/v1/system/metrics') {
            return createOfflineMetricsResponse();
        }
        
        // Return generic offline response
        return createOfflineResponse();
    }
}

// Handle static resources with cache-first strategy
async function handleStaticRequest(request) {
    try {
        // Try cache first
        const cachedResponse = await caches.match(request);
        if (cachedResponse) {
            return cachedResponse;
        }
        
        // Cache miss, fetch from network
        const networkResponse = await fetch(request);
        
        if (networkResponse.ok) {
            const cache = await caches.open(STATIC_CACHE_NAME);
            cache.put(request, networkResponse.clone());
        }
        
        return networkResponse;
    } catch (error) {
        console.log('Failed to fetch static resource:', request.url);
        
        // If it's the main page, return cached version or offline page
        if (request.url.endsWith('/') || request.url.endsWith('/index.html')) {
            const cachedResponse = await caches.match('/index.html');
            return cachedResponse || createOfflinePage();
        }
        
        throw error;
    }
}

// Handle navigation requests
async function handleNavigationRequest(request) {
    try {
        // Try network first for navigation
        const networkResponse = await fetch(request);
        return networkResponse;
    } catch (error) {
        // Network failed, serve cached index.html for SPA routing
        const cachedResponse = await caches.match('/index.html');
        return cachedResponse || createOfflinePage();
    }
}

// Utility functions
function isStaticResource(pathname) {
    return pathname.startsWith('/css/') ||
           pathname.startsWith('/js/') ||
           pathname.startsWith('/images/') ||
           pathname.startsWith('/icons/') ||
           pathname.endsWith('.svg') ||
           pathname.endsWith('.png') ||
           pathname.endsWith('.jpg') ||
           pathname.endsWith('.css') ||
           pathname.endsWith('.js') ||
           pathname === '/' ||
           pathname === '/index.html' ||
           pathname === '/manifest.json';
}

function getCacheMaxAge(pathname) {
    if (pathname.includes('/metrics')) {
        return CACHE_EXPIRATION.metrics;
    } else if (pathname.includes('/entities')) {
        return CACHE_EXPIRATION.entities;
    } else if (pathname.startsWith('/api/')) {
        return CACHE_EXPIRATION.api;
    } else {
        return CACHE_EXPIRATION.static;
    }
}

function createOfflineEntitiesResponse() {
    const offlineData = {
        entities: [],
        total: 0,
        message: 'Offline mode - cached entities not available'
    };
    
    return new Response(JSON.stringify(offlineData), {
        status: 200,
        statusText: 'OK (Offline)',
        headers: {
            'Content-Type': 'application/json',
            'sw-offline': 'true'
        }
    });
}

function createOfflineMetricsResponse() {
    const offlineMetrics = {
        system: {
            status: 'offline',
            uptime: 'N/A',
            num_cpu: 0,
            num_goroutines: 0
        },
        database: {
            total_entities: 0,
            total_tags: 0,
            db_size: 0,
            wal_size: 0
        },
        memory: {
            alloc: 0,
            heap_alloc: 0,
            sys: 0,
            num_gc: 0
        },
        performance: {
            avg_query_time: 0
        },
        http: {
            requests_per_minute: 0,
            total_requests: 0,
            avg_response_time: 0
        },
        offline: true
    };
    
    return new Response(JSON.stringify(offlineMetrics), {
        status: 200,
        statusText: 'OK (Offline)',
        headers: {
            'Content-Type': 'application/json',
            'sw-offline': 'true'
        }
    });
}

function createOfflineResponse() {
    return new Response(JSON.stringify({
        error: 'Offline',
        message: 'This request is not available offline'
    }), {
        status: 503,
        statusText: 'Service Unavailable (Offline)',
        headers: {
            'Content-Type': 'application/json',
            'sw-offline': 'true'
        }
    });
}

function createOfflinePage() {
    const offlineHTML = `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>EntityDB - Offline</title>
            <style>
                body {
                    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
                    margin: 0;
                    padding: 0;
                    background: #f5f6fa;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    min-height: 100vh;
                }
                .offline-container {
                    text-align: center;
                    max-width: 400px;
                    padding: 40px 20px;
                    background: white;
                    border-radius: 12px;
                    box-shadow: 0 4px 24px rgba(0,0,0,0.1);
                }
                .offline-icon {
                    font-size: 64px;
                    color: #95a5a6;
                    margin-bottom: 24px;
                }
                .offline-title {
                    font-size: 24px;
                    font-weight: 600;
                    color: #2c3e50;
                    margin-bottom: 16px;
                }
                .offline-message {
                    color: #6c757d;
                    line-height: 1.5;
                    margin-bottom: 24px;
                }
                .retry-button {
                    background: #3498db;
                    color: white;
                    border: none;
                    padding: 12px 24px;
                    border-radius: 6px;
                    font-size: 14px;
                    font-weight: 500;
                    cursor: pointer;
                    transition: background-color 0.2s;
                }
                .retry-button:hover {
                    background: #2980b9;
                }
            </style>
        </head>
        <body>
            <div class="offline-container">
                <div class="offline-icon">📡</div>
                <h1 class="offline-title">You're Offline</h1>
                <p class="offline-message">
                    EntityDB is not available right now. Please check your connection and try again.
                </p>
                <button class="retry-button" onclick="window.location.reload()">
                    Try Again
                </button>
            </div>
        </body>
        </html>
    `;
    
    return new Response(offlineHTML, {
        status: 200,
        statusText: 'OK (Offline)',
        headers: {
            'Content-Type': 'text/html',
            'sw-offline': 'true'
        }
    });
}

// Background sync for offline data
self.addEventListener('sync', event => {
    console.log('Background sync triggered:', event.tag);
    
    if (event.tag === 'entitydb-sync') {
        event.waitUntil(syncOfflineData());
    }
});

async function syncOfflineData() {
    try {
        console.log('Syncing offline data...');
        
        // Get all clients to notify them of sync
        const clients = await self.clients.matchAll();
        clients.forEach(client => {
            client.postMessage({
                type: 'sync-start'
            });
        });
        
        // Attempt to sync any pending changes
        // This would integrate with the entity browser's offline queue
        
        // Notify clients of successful sync
        clients.forEach(client => {
            client.postMessage({
                type: 'sync-complete',
                success: true
            });
        });
        
        console.log('Offline data sync completed');
    } catch (error) {
        console.error('Offline data sync failed:', error);
        
        // Notify clients of sync failure
        const clients = await self.clients.matchAll();
        clients.forEach(client => {
            client.postMessage({
                type: 'sync-complete',
                success: false,
                error: error.message
            });
        });
    }
}

// Push notifications (for future implementation)
self.addEventListener('push', event => {
    if (!event.data) return;
    
    try {
        const data = event.data.json();
        const options = {
            body: data.body || 'EntityDB notification',
            icon: '/icons/icon-192x192.png',
            badge: '/icons/icon-72x72.png',
            data: data.data || {},
            actions: data.actions || []
        };
        
        event.waitUntil(
            self.registration.showNotification(data.title || 'EntityDB', options)
        );
    } catch (error) {
        console.error('Push notification error:', error);
    }
});

// Notification click handling
self.addEventListener('notificationclick', event => {
    event.notification.close();
    
    if (event.action) {
        // Handle action buttons
        console.log('Notification action clicked:', event.action);
    } else {
        // Handle notification click
        event.waitUntil(
            self.clients.matchAll().then(clients => {
                // Focus existing window or open new one
                if (clients.length > 0) {
                    return clients[0].focus();
                } else {
                    return self.clients.openWindow('/');
                }
            })
        );
    }
});

// Message handling from main thread
self.addEventListener('message', event => {
    const { type, data } = event.data;
    
    switch (type) {
        case 'SKIP_WAITING':
            self.skipWaiting();
            break;
            
        case 'GET_CACHE_STATUS':
            getCacheStatus().then(status => {
                event.ports[0].postMessage(status);
            });
            break;
            
        case 'CLEAR_CACHE':
            clearAllCaches().then(() => {
                event.ports[0].postMessage({ success: true });
            }).catch(error => {
                event.ports[0].postMessage({ success: false, error: error.message });
            });
            break;
            
        case 'CACHE_ENTITIES':
            if (data && data.entities) {
                cacheEntities(data.entities).then(() => {
                    event.ports[0].postMessage({ success: true });
                });
            }
            break;
    }
});

async function getCacheStatus() {
    const caches_list = await caches.keys();
    const status = {};
    
    for (const cacheName of caches_list) {
        const cache = await caches.open(cacheName);
        const keys = await cache.keys();
        status[cacheName] = {
            entries: keys.length,
            urls: keys.map(req => req.url)
        };
    }
    
    return status;
}

async function clearAllCaches() {
    const cacheNames = await caches.keys();
    return Promise.all(
        cacheNames.map(cacheName => caches.delete(cacheName))
    );
}

async function cacheEntities(entities) {
    const cache = await caches.open(API_CACHE_NAME);
    
    // Create a synthetic response for entities
    const response = new Response(JSON.stringify({
        entities: entities,
        total: entities.length,
        cached: true
    }), {
        headers: {
            'Content-Type': 'application/json',
            'sw-cached-at': Date.now().toString()
        }
    });
    
    await cache.put('/api/v1/entities/list', response);
}

console.log('EntityDB Service Worker script loaded');