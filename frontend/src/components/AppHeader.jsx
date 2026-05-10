import { Link, NavLink } from 'react-router-dom'
import { PeriodStatusRibbon } from './PeriodStatusRibbon'
import { useAuth } from '../context/AuthContext'
import { UserMenu } from './UserMenu'

/**
 * AppHeader — top bar with brand identity, household selector, reload
 * action, and backend health indicator.
 */
export function AppHeader({
  health,
  householdId,
  onHouseholdChange,
  onReload,
  onLogout,
  isLoading,
  households = [],
  periodStatus = 'open',
  isMutationLocked = false,
}) {
  const { user } = useAuth()

  return (
    <header className="appHeader">
      {/* Brand */}
        <div className="brand">
          <div className="brandIcon" aria-hidden>💸</div>
          <div>
            <div className="brandName">micha</div>
            <div className="brandTagline">Claridad financiera para pareja y roomies</div>
          </div>
        </div>

      <nav className="headerNav" aria-label="Primary sections">
        <NavLink to="/" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
          Resumen
        </NavLink>
        <NavLink to="/expenses" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
          Movimientos
        </NavLink>
        <NavLink to="/balances" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
          Balances
        </NavLink>
        <NavLink to="/installments" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
          Plazos
        </NavLink>
        <NavLink to="/rules" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
          Reglas
        </NavLink>
      </nav>

      {/* Controls */}
      <div className="headerControls">
        {/* Invite member */}
        {householdId && (
          <Link
            to="/members/new"
            className={`btn btnGhost btnSm${isMutationLocked ? ' btnDisabled' : ''}`}
            aria-label="Invitar nuevo miembro"
            aria-disabled={isMutationLocked}
            tabIndex={isMutationLocked ? -1 : 0}
            onClick={(event) => {
              if (isMutationLocked) {
                event.preventDefault()
              }
            }}
          >
            + Miembro
          </Link>
        )}

        {/* Reload */}
        <button
          type="button"
          className="btn btnGhost btnSm"
          onClick={onReload}
          disabled={isLoading}
          aria-label="Actualizar gastos"
        >
          <span className={isLoading ? 'spinIcon' : ''} aria-hidden>⟳</span>
          {isLoading ? 'Cargando…' : 'Actualizar'}
        </button>

        <UserMenu user={user} households={households} householdId={householdId} onHouseholdChange={onHouseholdChange} onLogout={onLogout} health={health} />
      </div>

      <PeriodStatusRibbon status={periodStatus} />
    </header>
  )
}
