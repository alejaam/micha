import { Link, useLocation } from 'react-router-dom'
import { HomeIcon, ArrowsRightLeftIcon, BanknotesIcon, CalendarDaysIcon, Cog6ToothIcon, PlusIcon } from '@heroicons/react/24/outline'
import { HomeIcon as HomeIconSolid, ArrowsRightLeftIcon as ArrowsRightLeftIconSolid, BanknotesIcon as BanknotesIconSolid, CalendarDaysIcon as CalendarDaysIconSolid, Cog6ToothIcon as Cog6ToothIconSolid } from '@heroicons/react/24/solid'

/**
 * BottomNav - Mobile bottom navigation bar (pill style)
 * Shows on screens < 880px, hidden on desktop
 *
 * When no householdId is provided, hides household-dependent links and
 * shows a "Configurar hogar" button instead.
 */
export function BottomNav({ householdId }) {
    const location = useLocation()
    const currentPath = location.pathname

    const navItems = [
        { path: '/', icon: HomeIcon, iconActive: HomeIconSolid, label: 'Vista', exact: true },
        { path: '/expenses', icon: ArrowsRightLeftIcon, iconActive: ArrowsRightLeftIconSolid, label: 'Movimientos' },
        { path: '/balances', icon: BanknotesIcon, iconActive: BanknotesIconSolid, label: 'Balances' },
        { path: '/installments', icon: CalendarDaysIcon, iconActive: CalendarDaysIconSolid, label: 'Plazos' },
        { path: '/rules', icon: Cog6ToothIcon, iconActive: Cog6ToothIconSolid, label: 'Config' },
    ]

    const isActive = (item) => {
        if (item.exact) {
            return currentPath === item.path
        }
        return currentPath.startsWith(item.path)
    }

    if (!householdId) {
        return (
            <nav className="pillNav" aria-label="Main navigation">
                <Link to="/onboarding/household" className="pillNavItem">
                    <PlusIcon className="pillNavIcon" />
                    <span className="pillNavLabel">Configurar hogar</span>
                </Link>
            </nav>
        )
    }

    return (
        <nav className="pillNav" aria-label="Main navigation">
            {navItems.map((item) => {
                const active = isActive(item)
                const IconComponent = active ? item.iconActive : item.icon
                return (
                    <Link
                        key={item.path}
                        to={item.path}
                        className={`pillNavItem ${active ? 'active' : ''}`}
                        aria-current={active ? 'page' : undefined}
                    >
                        <IconComponent className="pillNavIcon" />
                        <span className="pillNavLabel">{item.label}</span>
                    </Link>
                )
            })}
        </nav>
    )
}
