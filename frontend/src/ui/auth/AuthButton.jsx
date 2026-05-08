/**
 * AuthButton — primary/ghost button variants for dark theme auth/onboarding.
 *
 * @param {'primary'|'ghost'} [variant]
 * @param {'sm'|'md'} [size]
 * @param {boolean} [disabled]
 * @param {boolean} [busy]
 * @param {boolean} [fullWidth]
 * @param {React.ReactNode} children
 * @param {(e: React.MouseEvent) => void} [onClick]
 * @param {string} [type]
 * @param {object} [rest]
 */
export function AuthButton({
    variant = 'primary',
    disabled,
    busy,
    fullWidth,
    children,
    onClick,
    type = 'button',
    ...rest
}) {
    const cls = [
        'pd-btn',
        variant === 'ghost' ? 'pd-btnGhost' : 'pd-btnPrimary',
        fullWidth ? 'pd-btnFull' : '',
    ].filter(Boolean).join(' ')

    return (
        <button
            type={type}
            className={cls}
            disabled={disabled || busy}
            onClick={onClick}
            {...rest}
        >
            {busy ? <span className="pd-btnSpinner" aria-hidden /> : null}
            {children}
        </button>
    )
}
