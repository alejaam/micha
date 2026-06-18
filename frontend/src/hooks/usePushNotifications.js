import { useState, useEffect, useCallback, useRef } from 'react'
import { getVapidPublicKey, subscribePush, testPushNotification } from '../api'

const PERMISSION_STATES = {
  PROMPT: 'prompt',
  GRANTED: 'granted',
  DENIED: 'denied',
  UNSUPPORTED: 'unsupported',
}

/**
 * usePushNotifications — manages Web Push subscription lifecycle.
 *
 * Returns:
 *   - permission:  'prompt' | 'granted' | 'denied' | 'unsupported'
 *   - isSubscribed: boolean — whether a push subscription is active
 *   - isLoading:    boolean — true while subscribing or sending test
 *   - statusText:   string  — human-readable status message
 *   - subscribe():  Promise — request permission + register subscription
 *   - sendTest():   Promise — send test notification from backend
 */
export function usePushNotifications() {
  const [permission, setPermission] = useState(PERMISSION_STATES.PROMPT)
  const [isSubscribed, setIsSubscribed] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [statusText, setStatusText] = useState('')
  const swRegistrationRef = useRef(null)

  // Detect permission state on mount.
  useEffect(() => {
    if (!('Notification' in window)) {
      setPermission(PERMISSION_STATES.UNSUPPORTED)
      setStatusText('Notificaciones no soportadas en este navegador')
      return
    }

    setPermission(Notification.permission)

    // Already granted? Try to check if we have an active subscription.
    if (Notification.permission === 'granted') {
      checkExistingSubscription()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Listen for permission changes (Safari, Firefox).
  useEffect(() => {
    if (!('Notification' in window)) return

    const handlePermissionChange = () => {
      setPermission(Notification.permission)
    }

    // Modern browsers support the permission API.
    if (navigator.permissions) {
      navigator.permissions.query({ name: 'notifications' }).then((result) => {
        result.addEventListener('change', () => {
          setPermission(Notification.permission)
        })
      }).catch(() => {})
    }

    // Fallback: check on focus.
    window.addEventListener('focus', handlePermissionChange)
    return () => window.removeEventListener('focus', handlePermissionChange)
  }, [])

  const checkExistingSubscription = useCallback(async () => {
    try {
      const registration = await navigator.serviceWorker.register('/sw.js')
      swRegistrationRef.current = registration
      const subscription = await registration.pushManager.getSubscription()
      setIsSubscribed(!!subscription)
    } catch {
      // SW might not be supported; treat as unsubscribed.
      setIsSubscribed(false)
    }
  }, [])

  const subscribe = useCallback(async () => {
    setIsLoading(true)
    setStatusText('')

    try {
      // 1. Request permission.
      if (!('Notification' in window)) {
        setStatusText('Este navegador no soporta notificaciones push')
        setIsLoading(false)
        return
      }

      if (Notification.permission === 'denied') {
        setStatusText('Bloqueaste las notificaciones — cámbialo en ajustes del navegador')
        setIsLoading(false)
        return
      }

      let perm = Notification.permission
      if (perm === 'default') {
        perm = await Notification.requestPermission()
        setPermission(perm)
      }

      if (perm !== 'granted') {
        setStatusText('Permiso denegado — no podrás recibir notificaciones')
        setIsLoading(false)
        return
      }

      // 2. Register service worker.
      const registration = await navigator.serviceWorker.register('/sw.js')
      swRegistrationRef.current = registration

      // 3. Get VAPID public key from backend.
      const { publicKey } = await getVapidPublicKey()
      if (!publicKey) {
        setStatusText('Error: el servidor no tiene configurada la clave VAPID')
        setIsLoading(false)
        return
      }

      // 4. Subscribe with the push server.
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(publicKey),
      })

      // 5. Send subscription to backend.
      await subscribePush(subscription)
      setIsSubscribed(true)
      setStatusText('✅ Notificaciones activadas')
    } catch (err) {
      console.error('Push subscribe error:', err)
      setStatusText('Error al activar notificaciones: ' + (err.message || err))
    } finally {
      setIsLoading(false)
    }
  }, [])

  const sendTest = useCallback(async () => {
    setIsLoading(true)
    setStatusText('')

    try {
      const result = await testPushNotification()
      if (result?.status === 'sent') {
        setStatusText(`🔔 Notificación de prueba enviada (${result.sent} dispositivo(s))`)
      }
    } catch (err) {
      console.error('Push test error:', err)
      if (err.message?.includes('no push subscriptions')) {
        setStatusText('Primero activa las notificaciones con el botón 🔔')
      } else {
        setStatusText('Error al enviar prueba: ' + (err.message || err))
      }
    } finally {
      setIsLoading(false)
    }
  }, [])

  return {
    permission,
    isSubscribed,
    isLoading,
    statusText,
    subscribe,
    sendTest,
  }
}

/**
 * urlBase64ToUint8Array — converts a base64url-encoded string to a Uint8Array.
 * Required by the Push API for the applicationServerKey.
 */
function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding)
    .replace(/-/g, '+')
    .replace(/_/g, '/')

  const rawData = window.atob(base64)
  const output = new Uint8Array(rawData.length)

  for (let i = 0; i < rawData.length; ++i) {
    output[i] = rawData.charCodeAt(i)
  }
  return output
}
