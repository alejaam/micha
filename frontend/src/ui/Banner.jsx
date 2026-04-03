import { useEffect } from 'react'

/**
 * Banner — dismissible feedback strip shown above the content area.
 *
 * @param {'ok'|'error'} type - Visual variant
 * @param {React.ReactNode} children - Message text
 * @param {()=>void} [onDismiss] - Optional dismiss callback
 * @param {number} [autoDismissMs] - Auto-dismiss after N milliseconds (default: 5000 for 'ok', none for 'error')
 */
export function Banner({ type, children, onDismiss, autoDismissMs }) {
  const cls = type === 'ok' ? 'banner bannerOk' : 'banner bannerError'
  const icon = type === 'ok' ? '✓' : '⚠'

  // Auto-dismiss logic
  useEffect(() => {
    if (!onDismiss) return

    // Default: auto-dismiss success after 5s, errors require manual dismiss
    const shouldAutoDismiss = autoDismissMs !== undefined 
      ? autoDismissMs > 0 
      : type === 'ok'

    if (!shouldAutoDismiss) return

    const timeout = setTimeout(() => {
      onDismiss()
    }, autoDismissMs ?? 5000)

    return () => clearTimeout(timeout)
  }, [onDismiss, autoDismissMs, type])

  return (
    <div className={cls} role="alert" aria-live="polite">
      <span className="bannerIcon" aria-hidden>{icon}</span>
      <span className="bannerText">{children}</span>
      {onDismiss && (
        <button
          type="button"
          className="bannerDismiss"
          onClick={onDismiss}
          aria-label="Dismiss"
        >
          ✕
        </button>
      )}
    </div>
  )
}
