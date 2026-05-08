/**
 * AuthFormField — label + input wrapper with dark theme and error display.
 *
 * @param {string} label
 * @param {string} [htmlFor]
 * @param {string|null} [error]
 * @param {React.ReactNode} children
 */
export function AuthFormField({ label, htmlFor, error, children }) {
    const cls = [
        'pd-field',
        error ? 'pd-fieldError' : '',
    ].filter(Boolean).join(' ')

    return (
        <div className={cls}>
            {label ? (
                <label className="pd-label" htmlFor={htmlFor}>
                    {label}
                </label>
            ) : null}
            {children}
            {error ? <p className="pd-errorText">{error}</p> : null}
        </div>
    )
}
