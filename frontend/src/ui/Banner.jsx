/**
 * Banner — dismissible feedback strip shown above the content area.
 *
 * @param {'ok'|'error'} type - Visual variant
 * @param {React.ReactNode} children - Message text
 * @param {()=>void} [onDismiss] - Optional dismiss callback
 */
export function Banner({ type, children, onDismiss }) {
  const cls = type === 'ok' ? 'banner bannerOk' : 'banner bannerError'
  const icon = type === 'ok' ? '✓' : '⚠'

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
