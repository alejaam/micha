/**
 * AuthCard — glassmorphism card container for auth/onboarding surfaces.
 *
 * @param {React.ReactNode} children
 * @param {string} [className] — additional classes
 */
export function AuthCard({ children, className = '' }) {
    return (
        <div className={`pd-card ${className}`.trim()}>
            {children}
        </div>
    )
}
