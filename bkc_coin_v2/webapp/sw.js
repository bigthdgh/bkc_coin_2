// BKC Coin Service Worker for PWA functionality
const CACHE_NAME = 'bkc-coin-v1';
const STATIC_CACHE = 'bkc-static-v1';
const API_CACHE = 'bkc-api-v1';

// Files to cache for offline functionality
const STATIC_FILES = [
  '/',
  '/index_enhanced.html',
  '/app_enhanced.js',
  '/manifest.json',
  '/assets/coin.svg',
  'https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap',
  'https://cdn.jsdelivr.net/npm/chart.js@3.9.1/dist/chart.min.js',
  'https://telegram.org/js/telegram-web-app.js'
];

// API endpoints to cache
const API_ENDPOINTS = [
  '/api/v1/user',
  '/api/v1/balance',
  '/api/v1/rates/current',
  '/api/v1/economy/status'
];

// Install event - cache static files
self.addEventListener('install', (event) => {
  console.log('Service Worker: Installing...');
  
  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then((cache) => {
        console.log('Service Worker: Caching static files');
        return cache.addAll(STATIC_FILES);
      })
      .then(() => self.skipWaiting())
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('Service Worker: Activating...');
  
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (cacheName !== STATIC_CACHE && cacheName !== API_CACHE) {
              console.log('Service Worker: Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => self.clients.claim())
  );
});

// Fetch event - handle requests
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Handle API requests
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(handleAPIRequest(request));
    return;
  }

  // Handle static files
  event.respondWith(handleStaticRequest(request));
});

// Handle API requests with network-first strategy
async function handleAPIRequest(request) {
  try {
    // Try network first
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(API_CACHE);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('Service Worker: Network failed, trying cache:', error);
    
    // Fallback to cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Return offline page for GET requests
    if (request.method === 'GET') {
      return new Response(
        JSON.stringify({ 
          error: 'Offline', 
          message: 'Нет подключения к интернету. Попробуйте позже.',
          offline: true 
        }),
        {
          status: 503,
          headers: { 'Content-Type': 'application/json' }
        }
      );
    }
    
    throw error;
  }
}

// Handle static files with cache-first strategy
async function handleStaticRequest(request) {
  const cachedResponse = await caches.match(request);
  
  if (cachedResponse) {
    // Update cache in background
    updateCache(request);
    return cachedResponse;
  }
  
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(STATIC_CACHE);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('Service Worker: Static file fetch failed:', error);
    
    // Return cached version if available
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Return offline page for HTML requests
    if (request.headers.get('accept')?.includes('text/html')) {
      return caches.match('/index_enhanced.html');
    }
    
    throw error;
  }
}

// Update cache in background
async function updateCache(request) {
  try {
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      const cache = await caches.open(STATIC_CACHE);
      cache.put(request, networkResponse);
    }
  } catch (error) {
    console.log('Service Worker: Background update failed:', error);
  }
}

// Background sync for offline actions
self.addEventListener('sync', (event) => {
  if (event.tag === 'background-sync') {
    event.waitUntil(syncOfflineActions());
  }
});

// Sync offline actions when back online
async function syncOfflineActions() {
  try {
    const offlineActions = await getOfflineActions();
    
    for (const action of offlineActions) {
      try {
        await fetch(action.url, action.options);
        await removeOfflineAction(action.id);
      } catch (error) {
        console.log('Service Worker: Failed to sync action:', error);
      }
    }
  } catch (error) {
    console.log('Service Worker: Background sync failed:', error);
  }
}

// Push notification handler
self.addEventListener('push', (event) => {
  const options = {
    body: event.data?.text() || 'Новое уведомление BKC Coin',
    icon: '/assets/coin.svg',
    badge: '/assets/coin.svg',
    vibrate: [200, 100, 200],
    data: {
      dateOfArrival: Date.now(),
      primaryKey: 1
    },
    actions: [
      {
        action: 'explore',
        title: 'Открыть приложение',
        icon: '/assets/coin.svg'
      },
      {
        action: 'close',
        title: 'Закрыть',
        icon: '/assets/coin.svg'
      }
    ]
  };

  event.waitUntil(
    self.registration.showNotification('BKC Coin', options)
  );
});

// Notification click handler
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  if (event.action === 'explore') {
    event.waitUntil(
      clients.openWindow('/')
    );
  }
});

// Periodic background sync for updates
self.addEventListener('periodicsync', (event) => {
  if (event.tag === 'update-rates') {
    event.waitUntil(updateRates());
  }
});

// Update exchange rates periodically
async function updateRates() {
  try {
    const response = await fetch('/api/v1/rates/current');
    if (response.ok) {
      const rates = await response.json();
      
      // Store in cache
      const cache = await caches.open(API_CACHE);
      cache.put('/api/v1/rates/current', new Response(JSON.stringify(rates)));
      
      // Notify all clients
      const clients = await self.clients.matchAll();
      clients.forEach(client => {
        client.postMessage({
          type: 'RATES_UPDATED',
          data: rates
        });
      });
    }
  } catch (error) {
    console.log('Service Worker: Failed to update rates:', error);
  }
}

// Message handler for client communication
self.addEventListener('message', (event) => {
  const { type, data } = event.data;
  
  switch (type) {
    case 'SKIP_WAITING':
      self.skipWaiting();
      break;
    case 'GET_VERSION':
      event.ports[0].postMessage({ version: '1.0.0' });
      break;
    case 'CACHE_API_RESPONSE':
      cacheApiResponse(data);
      break;
    case 'CLEAR_CACHE':
      clearAllCaches();
      break;
  }
});

// Cache API response
async function cacheApiResponse(data) {
  try {
    const cache = await caches.open(API_CACHE);
    const response = new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' }
    });
    cache.put(data.url, response);
  } catch (error) {
    console.log('Service Worker: Failed to cache API response:', error);
  }
}

// Clear all caches
async function clearAllCaches() {
  try {
    const cacheNames = await caches.keys();
    await Promise.all(
      cacheNames.map(cacheName => caches.delete(cacheName))
    );
    console.log('Service Worker: All caches cleared');
  } catch (error) {
    console.log('Service Worker: Failed to clear caches:', error);
  }
}

// Offline storage helpers
async function getOfflineActions() {
  // In a real implementation, this would use IndexedDB
  return [];
}

async function removeOfflineAction(id) {
  // In a real implementation, this would use IndexedDB
}

// Network status helper
function isOnline() {
  return navigator.onLine;
}

// Cache management helper
async function getCacheSize() {
  const cacheNames = await caches.keys();
  let totalSize = 0;
  
  for (const cacheName of cacheNames) {
    const cache = await caches.open(cacheName);
    const requests = await cache.keys();
    totalSize += requests.length;
  }
  
  return totalSize;
}

// Performance monitoring
self.addEventListener('fetch', (event) => {
  const start = performance.now();
  
  event.waitUntil(
    (async () => {
      try {
        await fetch(event.request);
        const duration = performance.now() - start;
        
        // Log slow requests
        if (duration > 1000) {
          console.log(`Service Worker: Slow request (${duration.toFixed(2)}ms):`, event.request.url);
        }
      } catch (error) {
        console.log('Service Worker: Request failed:', error);
      }
    })()
  );
});

console.log('Service Worker: Loaded successfully');
