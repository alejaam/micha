import { useLocation, useNavigate } from 'react-router-dom'

/**
 * BottomNav - Mobile bottom navigation bar
 * Shows on screens < 880px, hidden on desktop
 */
export function BottomNav({ activeSection = 'overview', onSectionChange }) {
    const location = useLocation()
    const navigate = useNavigate()
    const isDashboardRoute = location.pathname === '/'

    const navItems = [
        { section: 'overview', code: '01', label: 'Resumen' },
        { section: 'planning', code: '02', label: 'Planeacion' },
        { section: 'activity', code: '03', label: 'Actividad' },
    ]

    const handleSectionClick = (nextSection) => {
        if (!isDashboardRoute) {
            navigate('/')
        }
        if (onSectionChange) {
            onSectionChange(nextSection)
        }
    }

    return (
        <nav className="bottomNav" aria-label="Dashboard sections">
            {navItems.map((item) => {
                const isActive = activeSection === item.section && isDashboardRoute
                return (
                    <button
                        key={item.section}
                        type="button"
                        className={`bottomNavItem ${isActive ? 'active' : ''}`}
                        aria-current={isActive ? 'page' : undefined}
                        aria-label={`Ir a seccion ${item.label}`}
                        onClick={() => handleSectionClick(item.section)}
                    >
                        <span className="bottomNavCode" aria-hidden>{item.code}</span>
                        <span className="bottomNavLabel">{item.label}</span>
                    </button>
                )
            })}
        </nav>
    )
}
