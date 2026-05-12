import { useEffect, useRef, useState } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { PeriodStatusRibbon } from './PeriodStatusRibbon'
import { useAuth } from '../context/AuthContext'
import { UserMenu } from './UserMenu'

function formatPeriodName(period) {
  if (!period) return 'Sin periodo activo'
  const raw = period.start_date || period.startDate || period.StartDate
  if (!raw) return 'Periodo actual'
  const start = new Date(raw)
  if (isNaN(start.getTime())) return 'Periodo actual'
  return new Intl.DateTimeFormat('es-MX', { month: 'long', year: 'numeric' }).format(start)
}

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
  currentPeriod = null,
}) {
  const { user } = useAuth()
  const periodName = formatPeriodName(currentPeriod)

  // Flash animation when period ID changes
  const prevPeriodIdRef = useRef(currentPeriod?.id)
  const [isFlashing, setIsFlashing] = useState(false)

  useEffect(() => {
    const currentId = currentPeriod?.id
    const prevId = prevPeriodIdRef.current
    if (prevId && currentId && prevId !== currentId) {
      setIsFlashing(true)
      const timer = setTimeout(() => setIsFlashing(false), 800)
      return () => clearTimeout(timer)
    }
    prevPeriodIdRef.current = currentId
  }, [currentPeriod?.id])

  return (
    <header className={`appHeader${isFlashing ? ' appHeader--flash' : ''}`}>
      {/* Brand */}
        <div className="brand">
          <div className={`brandIcon brandIcon--${periodStatus}`} aria-hidden>💸</div>
          <div>
            <div className="brandName">micha</div>
            <div className="brandTagline">Claridad financiera para pareja y roomies</div>
            {currentPeriod && (
              <div className="brandPeriod">
                <span className="brandPeriodLabel">Periodo</span>
                <span className="brandPeriodName">{periodName}</span>
              </div>
            )}
          </div>
        </div>

      {householdId && (
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
      )}

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
