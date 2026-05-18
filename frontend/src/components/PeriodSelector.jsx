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
  const raw = period?.start_date || period?.startDate || period?.StartDate
  if (!raw) return 'Periodo'
  const start = new Date(raw)
  if (isNaN(start.getTime())) return 'Periodo'
  return new Intl.DateTimeFormat('es-MX', { month: 'long', year: 'numeric' }).format(start)
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

  return (
    <div className="periodSelector">
      <select
        className="input inputSm periodSelectorSelect"
        value={selectedPeriodId || currentPeriodId || ''}
        onChange={handleChange}
        aria-label="Seleccionar periodo"
      >
        {currentPeriodId && (
          <option value={currentPeriodId}>Actual — {formatPeriodLabel(periods.find(p => (p?.id || p?.ID) === currentPeriodId))}</option>
        )}
        {periods
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
