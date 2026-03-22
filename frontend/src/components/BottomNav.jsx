import { useLocation, Link } from 'react-router-dom'

/**
 * BottomNav - Mobile bottom navigation bar
 * Shows on screens < 880px, hidden on desktop
 */
export function BottomNav() {
    const location = useLocation()
    const currentPath = location.pathname

    const navItems = [
        { path: '/', icon: '~', label: 'Home', exact: true },
        { path: '/expenses', icon: '$', label: 'Expenses' },
        { path: '/settlement', icon: '=', label: 'Settle' },
        { path: '/settings', icon: '*', label: 'Settings' },
    ]

    // For now, since we only have the dashboard, we'll simplify
    // and just show visual feedback on the current item
    const isActive = (item) => {
        if (item.exact) {
            return currentPath === item.path
        }
        return currentPath.startsWith(item.path)
    }

    return (
        <nav className="bottomNav" aria-label="Main navigation">
            {navItems.map((item) => (
                <Link
                    key={item.path}
                    to={item.path}
                    className={`bottomNavItem ${isActive(item) ? 'active' : ''}`}
                    aria-current={isActive(item) ? 'page' : undefined}
                >
                    <span className="bottomNavIcon" aria-hidden>{item.icon}</span>
                    <span className="bottomNavLabel">{item.label}</span>
                </Link>
            ))}
        </nav>
    )
}
