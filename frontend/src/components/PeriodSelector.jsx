const MONTH_NAMES_SHORT = [
  'ene', 'feb', 'mar', 'abr', 'may', 'jun',
  'jul', 'ago', 'sep', 'oct', 'nov', 'dic',
]

/**
 * PeriodSelector — dropdown to select a historical period or return to current.
 *
 * Props:
 *   periods          — full array of period objects from listPeriods API
 *   selectedPeriodId — currently selected period id (string|null)
 *   currentPeriodId  — id of the current active period
 *   onSelect(periodId) — called when user picks a period (null = current)
 */
function formatPeriodLabel(period) {
  if (!period) return 'Periodo'
  const rawStart = period?.start_date || period?.startDate
  const rawEnd = period?.end_date || period?.endDate
  if (!rawStart || !rawEnd) return 'Periodo'
  const start = new Date(rawStart)
  const end = new Date(rawEnd)
  if (isNaN(start.getTime()) || isNaN(end.getTime())) return 'Periodo'
  const dayStart = start.getDate()
  const dayEnd = end.getDate()
  const month = MONTH_NAMES_SHORT[start.getMonth()]
  const year = start.getFullYear()
  return `${dayStart}–${dayEnd} ${month} ${year}`
}

function parseStatus(period) {
  return (period.status || period.Status || 'open').toLowerCase()
}

export function PeriodSelector({ periods = [], selectedPeriodId, currentPeriodId, onSelect }) {
  const isHistoricalSelected = selectedPeriodId && selectedPeriodId !== currentPeriodId

  const handleChange = (e) => {
    const value = e.target.value
    onSelect(value || null)
  }

  // Filter out null/undefined periods to prevent "undefined is not an object" errors
  const validPeriods = periods.filter(Boolean)

  // Show placeholder when no periods exist
  if (!currentPeriodId && validPeriods.length === 0) {
    return (
      <div className="periodSelector periodSelector--empty">
        <span className="periodSelectorPlaceholder">Sin periodos aún</span>
      </div>
    )
  }

  return (
    <div className="periodSelector">
      <select
        className="input inputSm periodSelectorSelect"
        value={selectedPeriodId || currentPeriodId || ''}
        onChange={handleChange}
        aria-label="Seleccionar periodo"
      >
        {currentPeriodId && (
          <option value={currentPeriodId}>Actual — {formatPeriodLabel(validPeriods.find(p => (p?.id || p?.ID) === currentPeriodId))}</option>
        )}
        {validPeriods
          .filter((p) => {
            const id = p?.id || p?.ID
            return id !== currentPeriodId && parseStatus(p) === 'closed'
          })
          .sort((a, b) => {
            const dateA = new Date(a?.start_date || a?.StartDate || 0)
            const dateB = new Date(b?.start_date || b?.StartDate || 0)
            return dateB - dateA
          })
          .map((p) => {
            const id = p?.id || p?.ID
            return (
              <option key={id} value={id}>
                {formatPeriodLabel(p)}
              </option>
            )
          })}
      </select>
      {isHistoricalSelected && (
        <button
          type="button"
          className="btn btnGhost btnSm periodSelectorBack"
          onClick={() => onSelect(null)}
          aria-label="Volver al periodo actual"
        >
          Volver al actual
        </button>
      )}
    </div>
  )
}
