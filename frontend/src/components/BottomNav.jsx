import { useLocation, Link } from 'react-router-dom'

/**
 * BottomNav - Mobile bottom navigation bar
 * Shows on screens < 880px, hidden on desktop
 */
export function BottomNav() {
    const location = useLocation()
    const currentPath = location.pathname

    const navItems = [
        { path: '/', icon: '🏠', label: 'Home', exact: true, enabled: true },
        { path: '/expenses', icon: '💰', label: 'Expenses', enabled: false },
        { path: '/settlement', icon: '⚖️', label: 'Settle', enabled: false },
        { path: '/settings', icon: '⚙️', label: 'Settings', enabled: false },
    ]

    const isActive = (item) => {
        if (item.exact) {
            return currentPath === item.path
        }
        return currentPath.startsWith(item.path)
    }

    return (
        <nav className="bottomNav" aria-label="Main navigation">
            {navItems.map((item) =>
                item.enabled ? (
                    <Link
                        key={item.path}
                        to={item.path}
                        className={`bottomNavItem ${isActive(item) ? 'active' : ''}`}
                        aria-current={isActive(item) ? 'page' : undefined}
                    >
                        <span className="bottomNavIcon" aria-hidden>{item.icon}</span>
                        <span className="bottomNavLabel">{item.label}</span>
                    </Link>
                ) : (
                    <span
                        key={item.path}
                        className="bottomNavItem bottomNavItemDisabled"
                        aria-disabled="true"
                        title="Coming soon"
                    >
                        <span className="bottomNavIcon" aria-hidden>{item.icon}</span>
                        <span className="bottomNavLabel">{item.label}</span>
                    </span>
                ),
            )}
        </nav>
    )
}
