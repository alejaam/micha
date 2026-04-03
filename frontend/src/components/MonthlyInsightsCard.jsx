import { useMemo } from 'react'
import { formatCurrency } from '../utils'

function normalizeLabel(value, fallback) {
    if (!value || typeof value !== 'string') return fallback
    return value.trim() || fallback
}

/**
 * MonthlyInsightsCard - small Phase 6 kickoff slice.
 * Builds month-scoped KPIs from currently loaded expenses.
 */
export function MonthlyInsightsCard({ items = [], currency = 'MXN', year, month }) {
    const metrics = useMemo(() => {
        const filtered = items.filter((item) => {
            if (!item?.created_at) return false
            const date = new Date(item.created_at)
            if (Number.isNaN(date.getTime())) return false
            return date.getFullYear() === year && date.getMonth() + 1 === month
        })

        const byType = {}
        const byPaymentMethod = {}
        let totalCents = 0

        for (const item of filtered) {
            const amount = Number(item.amount_cents) || 0
            totalCents += amount

            const typeKey = normalizeLabel(item.expense_type, 'other')
            byType[typeKey] = (byType[typeKey] || 0) + amount

            const paymentKey = normalizeLabel(item.payment_method, 'unknown')
            byPaymentMethod[paymentKey] = (byPaymentMethod[paymentKey] || 0) + amount
        }

        const topTypeEntry = Object.entries(byType)
            .sort((a, b) => b[1] - a[1])[0] || null

        const paymentRows = Object.entries(byPaymentMethod)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 3)

        return {
            count: filtered.length,
            totalCents,
            topTypeEntry,
            paymentRows,
        }
    }, [items, month, year])

    if (metrics.count === 0) {
        return (
            <section className="card monthlyInsights" aria-label="Monthly insights">
                <div className="listHeader">
                    <h2 className="listTitle">Monthly insights</h2>
                    <span className="listCount">{month}/{year}</span>
                </div>
                <p className="formHint">No expenses for this period yet.</p>
            </section>
        )
    }

    const topTypeLabel = metrics.topTypeEntry ? metrics.topTypeEntry[0] : 'n/a'
    const topTypeAmount = metrics.topTypeEntry ? metrics.topTypeEntry[1] : 0

    return (
        <section className="card monthlyInsights" aria-label="Monthly insights">
            <div className="listHeader">
                <h2 className="listTitle">Monthly insights</h2>
                <span className="listCount">{month}/{year}</span>
            </div>

            <div className="monthlyInsightsGrid">
                <article className="monthlyInsightsMetric">
                    <span className="monthlyInsightsValue">{formatCurrency(metrics.totalCents, currency)}</span>
                    <span className="monthlyInsightsLabel">Total tracked</span>
                </article>

                <article className="monthlyInsightsMetric">
                    <span className="monthlyInsightsValue">{metrics.count}</span>
                    <span className="monthlyInsightsLabel">Recorded expenses</span>
                </article>

                <article className="monthlyInsightsMetric">
                    <span className="monthlyInsightsValue monthlyInsightsValueSmall">{topTypeLabel}</span>
                    <span className="monthlyInsightsLabel">Top expense type</span>
                    <span className="formHint">{formatCurrency(topTypeAmount, currency)}</span>
                </article>
            </div>

            <div className="monthlyInsightsBreakdown">
                <p className="formHint">Top payment methods</p>
                <ul className="monthlyInsightsList">
                    {metrics.paymentRows.map(([method, amount]) => (
                        <li key={method} className="monthlyInsightsRow">
                            <span>{method}</span>
                            <strong>{formatCurrency(amount, currency)}</strong>
                        </li>
                    ))}
                </ul>
            </div>
        </section>
    )
}
