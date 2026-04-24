/**
 * Banner — dismissible feedback strip shown above the content area.
 *
 * @param {'ok'|'error'} type - Visual variant
 * @param {React.ReactNode} children - Message text
 * @param {()=>void} [onDismiss] - Optional dismiss callback
 * @param {boolean} [floating] - Render as fixed toast overlay instead of inline banner
 */
export function Banner({ type, children, onDismiss, floating = false }) {
  const cls = [
    'banner',
    type === 'ok' ? 'bannerOk' : 'bannerError',
    floating ? 'bannerFloating' : '',
  ].filter(Boolean).join(' ')
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
