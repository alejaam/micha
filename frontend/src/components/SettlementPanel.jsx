import { Settings } from 'lucide-react'
import { useEffect } from 'react'
import { Tooltip } from '../ui/Tooltip'
import { formatCurrency } from '../utils'

const ALL_MONTHS = [
  { value: 1, label: '01 - Ene' }, { value: 2, label: '02 - Feb' },
  { value: 3, label: '03 - Mar' }, { value: 4, label: '04 - Abr' },
  { value: 5, label: '05 - May' }, { value: 6, label: '06 - Jun' },
  { value: 7, label: '07 - Jul' }, { value: 8, label: '08 - Ago' },
  { value: 9, label: '09 - Sep' }, { value: 10, label: '10 - Oct' },
  { value: 11, label: '11 - Nov' }, { value: 12, label: '12 - Dic' },
]

/**
 * SettlementPanel — monthly period controls + Excel-style adjustment table.
 * Shows the net balance per member and a clear "X owes Y $Z" callout.
 */
export function SettlementPanel({
  settlement,
  settlementYear,
  settlementMonth,
  onSettlementYearChange,
  onSettlementMonthChange,
  onRefresh,
  onResetToCurrentMonth,
  loadingSettlement,
  memberIndex,
  currency = 'MXN',
  selectedHousehold,
}) {
  const now = new Date()
  const currentYear = now.getUTCFullYear()
  const currentMonth = now.getUTCMonth() + 1

  const startDate = selectedHousehold?.created_at ? new Date(selectedHousehold.created_at) : now
  const startYear = startDate.getUTCFullYear()
  const startMonth = startDate.getUTCMonth() + 1

  // Clamp selection when household or bounds change
  useEffect(() => {
    if (!selectedHousehold) return

    let nextYear = settlementYear
    let nextMonth = settlementMonth

    if (settlementYear < startYear) {
      nextYear = startYear
      nextMonth = startMonth
    } else if (settlementYear > currentYear) {
      nextYear = currentYear
      nextMonth = currentMonth
    } else {
      if (settlementYear === startYear && settlementMonth < startMonth) {
        nextMonth = startMonth
      } else if (settlementYear === currentYear && settlementMonth > currentMonth) {
        nextMonth = currentMonth
      }
    }

    if (nextYear !== settlementYear) {
      onSettlementYearChange(nextYear)
    }
    if (nextMonth !== settlementMonth) {
      onSettlementMonthChange(nextMonth)
    }
  }, [selectedHousehold, settlementYear, settlementMonth, startYear, startMonth, currentYear, currentMonth, onSettlementYearChange, onSettlementMonthChange])

  const isCurrentMonth = settlementYear === currentYear && settlementMonth === currentMonth

  // Years from startYear to currentYear
  const availableYears = []
  for (let y = startYear; y <= currentYear; y++) {
    availableYears.push(y)
  }

  // Months available for the currently selected year
  const availableMonths = ALL_MONTHS.filter((m) => {
    if (settlementYear === startYear && m.value < startMonth) return false
    if (settlementYear === currentYear && m.value > currentMonth) return false
    return true
  })

  return (
    <section className="card" aria-label="Ajuste mensual">
      <h2 className="sectionTitle">
        <span className="sectionTitleIcon" aria-hidden>
          <Settings className="sectionIconSvg" size={14} strokeWidth={2} />
        </span>
        Ajuste mensual
        <Tooltip text="Muestra quien debe a quien en gastos compartidos del periodo. Se calcula por proporcion de ingreso o reparto exacto." position="right" />
        {isCurrentMonth && <span className="currentPeriodBadge">actual</span>}
        {settlement && <span className="sectionBadge" style={{ marginLeft: isCurrentMonth ? '8px' : 'auto' }}>
            {settlement.effective_settlement_mode === 'exact' ? 'Reparto exacto' : 'Proporcional a ingreso'}
        </span>}
      </h2>

      {/* Period controls */}
      <div className="settlementControls">
        <div className="householdRow">
          <label htmlFor="settlementYear" className="householdLabel">Ano</label>
          <select
            id="settlementYear"
            className="input inputSm"
            style={{ width: 90 }}
            value={settlementYear}
            onChange={(e) => onSettlementYearChange(Number(e.target.value))}
            aria-label="Ano de ajuste"
          >
            {availableYears.map((y) => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
        </div>
        <div className="householdRow">
          <label htmlFor="settlementMonth" className="householdLabel">Mes</label>
          <select
            id="settlementMonth"
            className="input inputSm"
            style={{ width: 120 }}
            value={settlementMonth}
            onChange={(e) => onSettlementMonthChange(Number(e.target.value))}
            aria-label="Mes de ajuste"
          >
            {availableMonths.map(({ value, label }) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </div>
        <button
          type="button"
          className="btn btnGhost btnSm"
          onClick={onRefresh}
          disabled={loadingSettlement}
        >
          {loadingSettlement
            ? <><span className="spinIcon" aria-hidden /> Cargando...</>
            : 'Actualizar'}
        </button>
        {!isCurrentMonth && (
          <button
            type="button"
            className="btn btnGhost btnSm"
            onClick={onResetToCurrentMonth}
            aria-label="Volver al mes actual"
          >
            Mes actual
          </button>
        )}
      </div>

      {settlement ? (
        <div className="formStack">
          {settlement.fallback_reason && (
            <p className="settlementFallback" role="alert">
              Aviso: {settlement.fallback_reason}
            </p>
          )}

          {/* Stats strip */}
          <div className="settlementStats">
            <span className="settlementStat">
              <span className="settlementStatLabel">Total compartido:</span>
              <span className="settlementStatValue">
                {formatCurrency(settlement.total_shared_cents, currency)}
              </span>
            </span>
            <span className="settlementStat">
              <span className="settlementStatLabel">Modo:</span>
              <span className="settlementStatValue">{settlement.effective_settlement_mode}</span>
            </span>
            <span className="settlementStat">
              <span className="settlementStatLabel">Gastos:</span>
              <span className="settlementStatValue">{settlement.included_expense_count}</span>
            </span>
            <span className="settlementStat">
              <span className="settlementStatLabel">Excluidos:</span>
              <span className="settlementStatValue">{settlement.excluded_voucher_count}</span>
            </span>
          </div>

          {/* Adjustment callouts — one per transfer */}
          {Array.isArray(settlement.transfers) && settlement.transfers.length > 0 ? (
            <div className="adjustmentList">
              {settlement.transfers.map((t, idx) => {
                const from = memberIndex[t.from_member_id] ?? t.from_member_id.slice(0, 8)
                const to = memberIndex[t.to_member_id] ?? t.to_member_id.slice(0, 8)
                return (
                  <div
                    key={`${t.from_member_id}-${t.to_member_id}-${idx}`}
                    className="adjustmentCallout"
                    role="status"
                    aria-label={`${from} debe pagar a ${to} ${formatCurrency(t.amount_cents, currency)}`}
                  >
                    <div className="adjustmentRow">
                      <span className="adjustmentDebtor">{from}</span>
                      <span className="adjustmentVerb">debe a</span>
                      <span className="adjustmentCreditor">{to}</span>
                    </div>
                    <span className="adjustmentAmount">
                      {formatCurrency(t.amount_cents, currency)}
                    </span>
                  </div>
                )
              })}
            </div>
          ) : (
            <p className="emptyHint">No se requieren transferencias, todo esta ajustado.</p>
          )}

          {/* Per-member balance summary */}
          {Array.isArray(settlement.members) && settlement.members.length > 0 && (
            <div className="settlementBalances">
              {settlement.members.map((sm) => {
                const name = memberIndex[sm.member_id] ?? sm.member_id.slice(0, 8)
                const pct = sm.salary_weight_bps != null
                  ? (sm.salary_weight_bps / 100).toFixed(1)
                  : null
                return (
                  <div key={sm.member_id} className="settlementBalanceRow">
                    <span className="settlementBalanceName">{name}</span>
                    {pct !== null && (
                      <span className="settlementBalancePct">{pct}%</span>
                    )}
                    <span className="settlementBalancePaid">
                      pago {formatCurrency(sm.paid_cents ?? 0, currency)}
                    </span>
                    <span className="settlementBalanceDue">
                      corresponde {formatCurrency(sm.expected_share ?? 0, currency)}
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      ) : (
        <div className="emptyState">
          <div className="emptyIcon" aria-hidden>~</div>
          <p className="emptyTitle">Sin datos de ajuste</p>
          <p className="emptyHint">No hay gastos registrados en este periodo.</p>
        </div>
      )}
    </section>
  )
}
