import { formatCurrency } from '../utils'

const SETTLEMENT_YEARS = Array.from({ length: 6 }, (_, i) => new Date().getUTCFullYear() - 4 + i)
const SETTLEMENT_MONTHS = [
  { value: 1, label: '01 - Jan' }, { value: 2, label: '02 - Feb' },
  { value: 3, label: '03 - Mar' }, { value: 4, label: '04 - Apr' },
  { value: 5, label: '05 - May' }, { value: 6, label: '06 - Jun' },
  { value: 7, label: '07 - Jul' }, { value: 8, label: '08 - Aug' },
  { value: 9, label: '09 - Sep' }, { value: 10, label: '10 - Oct' },
  { value: 11, label: '11 - Nov' }, { value: 12, label: '12 - Dec' },
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
}) {
  const now = new Date()
  const isCurrentMonth = settlementYear === now.getUTCFullYear() 
    && settlementMonth === (now.getUTCMonth() + 1)

  return (
    <section className="card" aria-label="Monthly settlement">
      <h2 className="sectionTitle">
        <span className="sectionTitleIcon" aria-hidden>🧮</span>
        Monthly settlement
        {isCurrentMonth && <span className="currentPeriodBadge">current</span>}
        {settlement && <span className="sectionBadge" style={{ marginLeft: isCurrentMonth ? '8px' : 'auto' }}>
            {settlement.effective_settlement_mode === 'exact' ? 'Exact split' : 'Income proportional'}
        </span>}
      </h2>

      {/* Period controls */}
      <div className="settlementControls">
        <div className="householdRow">
          <label htmlFor="settlementYear" className="householdLabel">Year</label>
          <select
            id="settlementYear"
            className="input inputSm"
            style={{ width: 90 }}
            value={settlementYear}
            onChange={(e) => onSettlementYearChange(Number(e.target.value))}
            aria-label="Settlement year"
          >
            {SETTLEMENT_YEARS.map((y) => (
              <option key={y} value={y}>{y}</option>
            ))}
          </select>
        </div>
        <div className="householdRow">
          <label htmlFor="settlementMonth" className="householdLabel">Month</label>
          <select
            id="settlementMonth"
            className="input inputSm"
            style={{ width: 120 }}
            value={settlementMonth}
            onChange={(e) => onSettlementMonthChange(Number(e.target.value))}
            aria-label="Settlement month"
          >
            {SETTLEMENT_MONTHS.map(({ value, label }) => (
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
            ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Loading...</>
            : 'Refresh'}
        </button>
        {!isCurrentMonth && (
          <button
            type="button"
            className="btn btnGhost btnSm"
            onClick={onResetToCurrentMonth}
            aria-label="Reset to current month"
          >
            📅 This month
          </button>
        )}
      </div>

      {settlement ? (
        <div className="formStack">
          {settlement.fallback_reason && (
            <p className="settlementFallback" role="alert">
              Warning: {settlement.fallback_reason}
            </p>
          )}

          {/* Stats strip */}
          <div className="settlementStats">
            <span className="settlementStat">
              <span className="settlementStatLabel">Total shared:</span>
              <span className="settlementStatValue">
                {formatCurrency(settlement.total_shared_cents, currency)}
              </span>
            </span>
            <span className="settlementStat">
              <span className="settlementStatLabel">Mode:</span>
              <span className="settlementStatValue">{settlement.effective_settlement_mode}</span>
            </span>
            <span className="settlementStat">
              <span className="settlementStatLabel">Expenses:</span>
              <span className="settlementStatValue">{settlement.included_expense_count}</span>
            </span>
            <span className="settlementStat">
              <span className="settlementStatLabel">Excluded:</span>
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
                    aria-label={`${from} owes ${to} ${formatCurrency(t.amount_cents, currency)}`}
                  >
                    <div className="adjustmentRow">
                      <span className="adjustmentDebtor">{from}</span>
                      <span className="adjustmentVerb">owes</span>
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
            <p className="emptyHint">No transfers needed — everyone is settled!</p>
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
                      paid {formatCurrency(sm.total_paid_cents ?? 0, currency)}
                    </span>
                    <span className="settlementBalanceDue">
                      owes {formatCurrency(sm.total_due_cents ?? 0, currency)}
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
          <p className="emptyTitle">No settlement data</p>
          <p className="emptyHint">No expenses recorded for this period.</p>
        </div>
      )}
    </section>
  )
}
