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
                <span className="summaryMetricValue" style={{ fontFamily: 'var(--font-display)', fontSize: '4rem', lineHeight: 1 }}>{formatCurrency(total, currency)}</span>
                <span className="summaryMetricLabel" style={{ fontFamily: 'var(--font-mono)' }}>total this month</span>
            </div>
            <div className="summaryDivider" aria-hidden />
            <div className="summaryMetric">
                <span className="summaryMetricValue" style={{ fontSize: '2.5rem' }}>{count}</span>
                <span className="summaryMetricLabel" style={{ fontFamily: 'var(--font-mono)' }}>expense{count !== 1 ? 's' : ''}</span>
            </div>
            {memberSummaries.length > 0 && (
                <>
                    <div className="summaryDivider" aria-hidden />
                    <div className="summaryMemberList">
                        {memberSummaries.map((m) => (
                            <div key={m.member_id} className="summaryMemberRow">
                                <span className="summaryMemberName" style={{ fontFamily: 'var(--font-mono)', textTransform: 'uppercase' }}>{m.name ?? m.member_id.slice(0, 8)}</span>
                                <span className="summaryMemberBalance" data-positive={m.net_balance_cents >= 0} style={{ fontFamily: 'var(--font-mono)' }}>
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
