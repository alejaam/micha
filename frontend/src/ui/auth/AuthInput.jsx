/**
 * AuthInput — styled input for dark theme auth/onboarding surfaces.
 *
 * @param {string} id
 * @param {string} [type]
 * @param {string} [placeholder]
 * @param {string} value
 * @param {(e: React.ChangeEvent<HTMLInputElement>) => void} [onChange]
 * @param {boolean} [disabled]
 * @param {boolean} [hasError]
 * @param {object} [rest] — additional props forwarded to <input>
 */
export function AuthInput({ id, type = 'text', placeholder, value, onChange, disabled, hasError, ...rest }) {
    const cls = [
        'pd-input',
        hasError ? 'pd-inputError' : '',
    ].filter(Boolean).join(' ')

    return (
        <input
            id={id}
            className={cls}
            type={type}
            placeholder={placeholder}
            value={value}
            onChange={onChange}
            disabled={disabled}
            {...rest}
        />
    )
}
