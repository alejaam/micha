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
 * SettlementPanel renders monthly settlement controls and transfer details.
 */
export function SettlementPanel({
  settlement,
  settlementYear,
  settlementMonth,
  onSettlementYearChange,
  onSettlementMonthChange,
  onRefresh,
  loadingSettlement,
  memberIndex,
  currency = 'MXN',
}) {
  return (
    <section className="card" aria-label="Monthly settlement">
      <h2 className="sectionTitle">Monthly settlement</h2>
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
        <button type="button" className="btn btnGhost btnSm" onClick={onRefresh} disabled={loadingSettlement}>
          {loadingSettlement ? <><span className="spinIcon" aria-hidden>&#x27F3;</span> Loading...</> : 'Refresh'}
        </button>
      </div>

      {settlement ? (
        <div className="formStack">
          {settlement.fallback_reason ? (
            <p className="settlementFallback" role="alert">Warning: {settlement.fallback_reason}</p>
          ) : null}
          <div className="settlementStats">
            <span className="settlementStat">
              <span className="settlementStatLabel">Total shared:</span>
              <span className="settlementStatValue">{formatCurrency(settlement.total_shared_cents, currency)}</span>
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
          <h3 className="sectionTitle">
            <span className="sectionTitleIcon" aria-hidden>{'<->'}</span>
            Transfers
          </h3>
          {Array.isArray(settlement.transfers) && settlement.transfers.length > 0 ? (
            <ul className="transferList">
              {settlement.transfers.map((t, idx) => (
                <li key={`${t.from_member_id}-${t.to_member_id}-${idx}`} className="transferItem">
                  <span className="transferNames">
                    {memberIndex[t.from_member_id] ?? t.from_member_id.slice(0, 8) + '...'}
                    <span className="transferArrow" aria-hidden>{'->'}</span>
                    {memberIndex[t.to_member_id] ?? t.to_member_id.slice(0, 8) + '...'}
                  </span>
                  <span className="transferAmount">{formatCurrency(t.amount_cents, currency)}</span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="emptyHint">No transfers needed - everyone is settled!</p>
          )}
        </div>
      ) : (
        <div className="emptyState">
          <div className="emptyIcon" aria-hidden>Data</div>
          <p className="emptyTitle">No settlement data</p>
          <p className="emptyHint">No expenses recorded for this period.</p>
        </div>
      )}
    </section>
  )
}
