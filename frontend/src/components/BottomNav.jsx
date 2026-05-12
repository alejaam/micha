import { Link, useLocation } from 'react-router-dom'

/**
 * BottomNav - Mobile bottom navigation bar
 * Shows on screens < 880px, hidden on desktop
 *
 * When no householdId is provided, hides household-dependent links and
 * shows a "Configurar hogar" button instead.
 */
export function BottomNav({ householdId }) {
    const location = useLocation()
    const currentPath = location.pathname

    const navItems = [
        { path: '/', icon: '📊', label: 'Resumen', exact: true },
        { path: '/expenses', icon: '💸', label: 'Movimientos' },
        { path: '/balances', icon: '⚖️', label: 'Balances' },
        { path: '/installments', icon: '📅', label: 'Plazos' },
        { path: '/rules', icon: '⚙️', label: 'Reglas' },
    ]

    const isActive = (item) => {
        if (item.exact) {
            return currentPath === item.path
        }
        return currentPath.startsWith(item.path)
    }

    if (!householdId) {
        return (
            <nav className="bottomNav" aria-label="Main navigation">
                <Link to="/onboarding/household" className="bottomNavItem">
                    <span className="bottomNavIcon" aria-hidden>+</span>
                    <span className="bottomNavLabel">Configurar hogar</span>
                </Link>
            </nav>
        )
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
