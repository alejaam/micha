// Micha — Service Worker for Web Push Notifications
self.addEventListener('install', () => {
  // Skip waiting so the new SW activates immediately.
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  // Take control of all clients immediately.
  event.waitUntil(clients.claim())
})

// Handle push events from the server.
self.addEventListener('push', (event) => {
  let data
  try {
    data = event.data ? event.data.json() : {}
  } catch {
    data = { title: '🔔 Micha', body: event.data?.text() ?? '' }
  }

  const title = data.title || '🔔 Micha'
  const options = {
    body: data.body || '',
    icon: data.icon || '/favicon.svg',
    badge: data.badge || '/favicon.svg',
    tag: data.tag || 'micha-notification',
    vibrate: [200, 100, 200],
    // Prevent duplicate notifications for the same tag.
    renotify: true,
    requireInteraction: true,
    data: data.data || {},
  }

  event.waitUntil(self.registration.showNotification(title, options))
})

// Handle notification click — open/ focus the app.
self.addEventListener('notificationclick', (event) => {
  event.notification.close()

  const urlToOpen = new URL('/', self.location.origin).href

  const promiseChain = clients
    .matchAll({ type: 'window', includeUncontrolled: true })
    .then((windowClients) => {
      let matchingClient = null
      for (const client of windowClients) {
        if (client.url === urlToOpen) {
          matchingClient = client
          break
        }
      }
      if (matchingClient) {
        return matchingClient.focus()
      }
      return clients.openWindow(urlToOpen)
    })

  event.waitUntil(promiseChain)
})
