import { useEffect, useRef, useState } from 'react'
import { Link, NavLink } from 'react-router-dom'
import { listPeriods } from '../api'
import { PeriodSelector } from './PeriodSelector'
import { PeriodStatusRibbon } from './PeriodStatusRibbon'
import { useAuth } from '../context/AuthContext'
import { UserMenu } from './UserMenu'
import { usePushNotifications } from '../hooks/usePushNotifications'

const MONTH_NAMES = [
  'enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
  'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre',
]

function formatPeriodName(period) {
  if (!period) return 'Sin periodo activo'
  const rawStart = period?.start_date || period?.startDate
  const rawEnd = period?.end_date || period?.endDate
  if (!rawStart || !rawEnd) return 'Periodo actual'
  const start = new Date(rawStart)
  const end = new Date(rawEnd)
  if (isNaN(start.getTime()) || isNaN(end.getTime())) return 'Periodo actual'
  const dayStart = start.getDate()
  const dayEnd = end.getDate()
  const month = MONTH_NAMES[start.getMonth()]
  const year = start.getFullYear()
  return `${dayStart}–${dayEnd} ${month} ${year}`
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
  currentPeriod = null,
  selectedPeriodId = null,
  onSelectPeriod = null,
}) {
  const { user } = useAuth()
  const periodName = formatPeriodName(currentPeriod)

  const [allPeriods, setAllPeriods] = useState([])

  // Push notifications hook
  const {
    isSubscribed,
    isLoading: pushLoading,
    statusText,
    subscribe,
    sendTest,
  } = usePushNotifications()

  // Load all periods when household changes
  useEffect(() => {
    if (!householdId) {
      setAllPeriods([])
      return
    }
    let cancelled = false
    listPeriods({ householdId })
      .then((data) => {
        if (!cancelled) setAllPeriods(Array.isArray(data) ? data : [])
      })
      .catch(() => {
        if (!cancelled) setAllPeriods([])
      })
    return () => { cancelled = true }
  }, [householdId])

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
          <NavLink to="/fixed-expenses" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
            Gastos fijos
          </NavLink>
          <NavLink to="/rules" className={({ isActive }) => `headerNavLink${isActive ? ' active' : ''}`}>
            Reglas
          </NavLink>
        </nav>
      )}

        {/* Controls */}
      <div className="headerControls">
        {/* Period chip inline */}
        <PeriodStatusRibbon status={periodStatus} />

        {/* Period selector (visible when there are periods to browse) */}
        {householdId && (
          <PeriodSelector
            periods={allPeriods}
            selectedPeriodId={selectedPeriodId}
            currentPeriodId={currentPeriod?.id}
            onSelect={(periodId) => onSelectPeriod && onSelectPeriod(periodId)}
          />
        )}

        {/* Invite member */}
        {householdId && (
          <Link
            to="/members/new"
            className="btn btnGhost btnSm"
            aria-label="Invitar nuevo miembro"
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

        {/* Push notification bell */}
        {householdId && (
          <div className="pushBtnWrapper" style={{ position: 'relative' }}>
            <button
              type="button"
              className={`btn btnGhost btnSm${isSubscribed ? '' : ' btnPulse'}`}
              onClick={isSubscribed ? sendTest : subscribe}
              disabled={pushLoading}
              aria-label={isSubscribed ? 'Enviar notificación de prueba' : 'Activar notificaciones'}
              title={statusText || (isSubscribed ? 'Notificaciones activas — toca para probar' : 'Activar notificaciones push')}
            >
              {pushLoading
                ? '…'
                : isSubscribed
                  ? '🔔'
                  : '🔕'}
            </button>
            {statusText && (
              <span
                className="pushStatus"
                style={{
                  position: 'absolute',
                  top: '100%',
                  right: 0,
                  fontSize: '0.65rem',
                  whiteSpace: 'nowrap',
                  background: 'var(--bg, #1a1a2e)',
                  padding: '2px 6px',
                  borderRadius: 4,
                  opacity: 0.85,
                  pointerEvents: 'none',
                  zIndex: 10,
                }}
              >
                {statusText}
              </span>
            )}
          </div>
        )}

        <UserMenu user={user} households={households} householdId={householdId} onHouseholdChange={onHouseholdChange} onLogout={onLogout} health={health} />
      </div>
    </header>
  )
}
