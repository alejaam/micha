/**
 * AuthHeader — eyebrow + title + subtitle pattern for auth/onboarding.
 *
 * @param {string} eyebrow
 * @param {string} title
 * @param {string} [subtitle]
 */
export function AuthHeader({ eyebrow, title, subtitle }) {
    return (
        <div className="pd-header">
            {eyebrow ? <p className="pd-eyebrow">{eyebrow}</p> : null}
            <h1 className="pd-title">{title}</h1>
            {subtitle ? <p className="pd-subtitle">{subtitle}</p> : null}
        </div>
    )
}
