import { formatCurrency } from '../utils'

/**
 * ExpenseSummary — compact metrics strip for the dashboard.
 * Shows total this month, expense count, and a breakdown by type.
 *
 * @param {object|null} settlement - Settlement data from the API
 * @param {string} currency
 */
export function ExpenseSummary({ settlement, currency = 'MXN' }) {
    if (!settlement) {
        return (
            <div className="summaryStrip summaryStripEmpty">
                <p className="emptyHint">No data for this period.</p>
            </div>
        )
    }

    const count = settlement.included_expense_count ?? 0
    const total = settlement.total_shared_cents ?? 0

    // Build per-member paid rows if available
    const memberSummaries = Array.isArray(settlement.members) ? settlement.members : []

    return (
        <div className="summaryStrip">
            <div className="summaryMetric">
                <span className="summaryMetricValue">{formatCurrency(total, currency)}</span>
                <span className="summaryMetricLabel">total this month</span>
            </div>
            <div className="summaryDivider" aria-hidden />
            <div className="summaryMetric">
                <span className="summaryMetricValue">{count}</span>
                <span className="summaryMetricLabel">expense{count !== 1 ? 's' : ''}</span>
            </div>
            {memberSummaries.length > 0 && (
                <>
                    <div className="summaryDivider" aria-hidden />
                    <div className="summaryMemberList">
                        {memberSummaries.map((m) => (
                            <div key={m.member_id} className="summaryMemberRow">
                                <span className="summaryMemberName">{m.name ?? m.member_id.slice(0, 8)}</span>
                                <span className="summaryMemberBalance" data-positive={m.net_balance_cents >= 0}>
                                    {m.net_balance_cents >= 0 ? '+' : ''}{formatCurrency(m.net_balance_cents, currency)}
                                </span>
                            </div>
                        ))}
                    </div>
                </>
            )}
        </div>
    )
}
