import { useLocation, Link } from 'react-router-dom'

/**
 * BottomNav - Mobile bottom navigation bar
 * Shows on screens < 880px, hidden on desktop
 */
export function BottomNav() {
    const location = useLocation()
    const currentPath = location.pathname

    const navItems = [
        { path: '/', label: 'DASHBOARD', exact: true, enabled: true },
        { path: '/members', label: 'MEMBERS', enabled: true },
        { path: '/settings', label: 'CONFIG', enabled: true },
    ]

    const isActive = (item) => {
        if (item.exact) {
            return currentPath === item.path
        }
        return currentPath.startsWith(item.path)
    }

    return (
        <nav className="bottomNav" aria-label="Main navigation" style={{ padding: '0 24px', justifyContent: 'space-between' }}>
            {navItems.map((item) =>
                item.enabled ? (
                    <Link
                        key={item.path}
                        to={item.path}
                        className={`bottomNavItem ${isActive(item) ? 'active' : ''}`}
                        aria-current={isActive(item) ? 'page' : undefined}
                        style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', letterSpacing: '0.1em' }}
                    >
                        <span className="bottomNavLabel">{item.label}</span>
                    </Link>
                ) : (
                    <span
                        key={item.path}
                        className="bottomNavItem bottomNavItemDisabled"
                        aria-disabled="true"
                        style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', letterSpacing: '0.1em', opacity: 0.3 }}
                    >
                        <span className="bottomNavLabel">{item.label}</span>
                    </span>
                ),
            )}
        </nav>
    )
}
