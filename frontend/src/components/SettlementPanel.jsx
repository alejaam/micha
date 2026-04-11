import { useEffect } from 'react'
import { motion } from 'framer-motion'
import { formatCurrency } from '../utils'
import { EmptyState } from '../ui/EmptyState'
import { Tooltip } from '../ui/Tooltip'
import { SemanticSettlementCard } from './SemanticSettlementCard'

const ALL_MONTHS = [
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
    <section className="card" aria-label="Monthly settlement">
      <h2 className="sectionTitle">
        <span className="sectionTitleIcon" aria-hidden>🧮</span>
        Monthly settlement
        <Tooltip text="Shows who owes whom for shared expenses this month. Based on income proportion or equal split." position="right" />
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
            {availableYears.map((y) => (
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
                  <motion.div
                    key={`${t.from_member_id}-${t.to_member_id}-${idx}`}
                    className="adjustmentCallout"
                    role="status"
                    aria-label={`${from} owes ${to} ${formatCurrency(t.amount_cents, currency)}`}
                    initial={{ opacity: 0, y: 5 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.15, ease: "easeOut", delay: idx * 0.05 }}
                  >
                    <div className="adjustmentRow">
                      <span className="adjustmentDebtor">{from}</span>
                      <span className="adjustmentVerb">owes</span>
                      <span className="adjustmentCreditor">{to}</span>
                    </div>
                    <span className="adjustmentAmount">
                      {formatCurrency(t.amount_cents, currency)}
                    </span>
                  </motion.div>
                )
              })}
            </div>
          ) : (
            <p className="emptyHint">No transfers needed — everyone is settled!</p>
          )}

          {/* Per-member balance summary */}
          {Array.isArray(settlement.members) && settlement.members.length > 0 && (
            <div className="settlementBalances" aria-label="Member balance cards">
              {settlement.members.map((sm, idx) => {
                const name = memberIndex[sm.member_id] ?? sm.member_id.slice(0, 8)
                const pct = sm.salary_weight_bps != null
                  ? (sm.salary_weight_bps / 100).toFixed(1)
                  : null
                return (
                  <motion.div 
                    key={sm.member_id} 
                    className="settlementBalanceRow"
                    initial={{ opacity: 0, x: -5 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ duration: 0.15, ease: "easeOut", delay: idx * 0.05 }}
                  >
                    <div className="settlementBalanceRowHeader">
                      {pct !== null && (
                        <span className="settlementBalancePct" aria-label={`Weight ${pct}%`}>
                          {pct}%
                        </span>
                      )}
                    </div>
                    <SemanticSettlementCard
                      memberName={name}
                      netBalanceCents={sm.net_balance_cents ?? 0}
                      currency={currency}
                    />
                    <div className="settlementBalanceMeta" aria-label={`Paid and share details for ${name}`}>
                      <span className="settlementBalancePaid">
                        PAID {formatCurrency(sm.paid_cents ?? 0, currency)}
                      </span>
                      <span className="settlementBalanceDue">
                        SHARE {formatCurrency(sm.expected_share ?? 0, currency)}
                      </span>
                    </div>
                  </motion.div>
                )
              })}
            </div>
          )}
        </div>
      ) : (
        <EmptyState
          title="No settlement data"
          description="No expenses recorded for this period."
          icon="[~]"
          compact
        />
      )}
    </section>
  )
}
