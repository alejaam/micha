/**
 * AuthBanner — error/success feedback strip for auth/onboarding.
 *
 * @param {'error'|'success'} type
 * @param {React.ReactNode} children
 */
export function AuthBanner({ type = 'error', children }) {
    if (!children) return null

    const cls = [
        'pd-banner',
        type === 'success' ? 'pd-bannerSuccess' : 'pd-bannerError',
    ].filter(Boolean).join(' ')

    const icon = type === 'success' ? '✓' : '⚠'

    return (
        <div className={cls} role="alert" aria-live="polite">
            <span className="pd-bannerIcon" aria-hidden>{icon}</span>
            <span>{children}</span>
        </div>
    )
}
