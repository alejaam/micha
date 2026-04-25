import { NavLink } from 'react-router-dom'

const MODULE_ITEMS = [
    { to: '/', label: 'Resumen', end: true },
    { to: '/expenses', label: 'Movimientos' },
    { to: '/balances', label: 'Balances' },
    { to: '/installments', label: 'Plazos' },
    { to: '/rules', label: 'Reglas' },
]

export function PrimaryNav() {
    return (
        <nav className="primaryNav card" aria-label="Primary sections">
            <ul className="primaryNavList">
                {MODULE_ITEMS.map((item) => (
                    <li key={item.to}>
                        <NavLink
                            to={item.to}
                            end={item.end}
                            className={({ isActive }) => `primaryNavLink${isActive ? ' active' : ''}`}
                        >
                            {item.label}
                        </NavLink>
                    </li>
                ))}
            </ul>
        </nav>
    )
}
