import { useMemo, useCallback } from 'react'
import { formatCurrency } from '../utils'

const FALLBACK_LABELS = {
    rent: 'Rent',
    auto: 'Auto',
    streaming: 'Streaming / Services',
    food: 'Food',
    personal: 'Personal',
    savings: 'Savings',
    other: 'Other',
}

/**
 * FixedExpensesPanel — mirrors the "Gastos mensuales (Fijos)" table from the Excel.
 * Groups fixed expenses (expense_type === 'fixed') by category and shows
 * each member's share side-by-side.
 *
 * @param {Array} items - expense list from API
 * @param {Array} members - member list from API
 * @param {string} currency
 * @param {Array} categories - dynamic categories from backend (optional)
 */
export function FixedExpensesPanel({ items = [], recurringItems = [], members = [], currency = 'MXN', categories = [] }) {
    // Build label map: prefer dynamic categories, fallback to hardcoded
    const categoryLabels = useMemo(() => {
        const labels = { ...FALLBACK_LABELS }
        for (const c of categories) {
            labels[c.id] = c.name
            labels[c.slug] = c.name
        }
        return labels
    }, [categories])

    const fixedItems = useMemo(() => {
        const fixedExpenses = items.filter((e) => e.expense_type === 'fixed' && e.is_shared)
        const agnosticTemplates = recurringItems
            .filter((e) => e.expense_type === 'fixed' && e.is_agnostic)
            .map((e) => ({
                ...e,
                is_shared: true,
            }))
        return [...fixedExpenses, ...agnosticTemplates]
    }, [items, recurringItems])

    const resolveConcept = useCallback((item) => {
        const description = String(item.description ?? '').trim()
        if (description) return description
        const key = item.category || item.category_id || 'other'
        return categoryLabels[key] ?? FALLBACK_LABELS[key] ?? 'Other'
    }, [categoryLabels])

// Group by concept, accumulating totals per member.
const grouped = useMemo(() => {
    const map = {}

    const addAgnosticSplit = (target, amountCents) => {
        if (members.length === 0) return
        const base = Math.floor(amountCents / members.length)
        const remainder = amountCents % members.length
        members.forEach((member, index) => {
            const extra = index < remainder ? 1 : 0
            target[member.id] = (target[member.id] ?? 0) + base + extra
        })
    }

    for (const e of fixedItems) {
        const concept = resolveConcept(e)
        if (!map[concept]) map[concept] = { concept, items: [], byMember: {}, totalCents: 0 }
        map[concept].items.push(e)
        map[concept].totalCents += e.amount_cents
        // Fixed expenses are always treated as household-shared in this panel:
        // no single member is considered the payer for display distribution.
        addAgnosticSplit(map[concept].byMember, e.amount_cents)
    }
    return Object.values(map)
}, [fixedItems, members, resolveConcept])

    // Total per member across all fixed
    const totalByMember = useMemo(() => {
        const totals = {}
        const addAgnosticSplit = (amountCents) => {
            if (members.length === 0) return
            const base = Math.floor(amountCents / members.length)
            const remainder = amountCents % members.length
            members.forEach((member, index) => {
                const extra = index < remainder ? 1 : 0
                totals[member.id] = (totals[member.id] ?? 0) + base + extra
            })
        }
        for (const e of fixedItems) {
            addAgnosticSplit(e.amount_cents)
        }
        return totals
    }, [fixedItems, members])

    const grandTotal = fixedItems.reduce((s, e) => s + e.amount_cents, 0)

    if (fixedItems.length === 0) {
        return (
            <section className="card" aria-label="Fixed expenses">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>#</span>
                    Fixed expenses
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">No fixed expenses yet</p>
                    <p className="emptyHint">Mark expenses as &quot;Fixed&quot; when adding them.</p>
                </div>
            </section>
        )
    }

    return (
        <section className="card" aria-label="Fixed expenses">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>#</span>
                Fixed expenses
                {fixedItems.length > 0 && <span className="sectionBadge">{fixedItems.length} expenses • {formatCurrency(grandTotal, currency)}</span>}
            </h2>

            <div className="fixedExpensesTable">
                {/* Header row */}
                <div className="fixedTableHeader">
                    <span className="fixedColConcept">Concept</span>
                    <span className="fixedColTotal">Total</span>
                    {members.map((m) => (
                        <span key={m.id} className="fixedColMember">{m.name}</span>
                    ))}
                </div>

                {/* Concept rows */}
                {grouped.map(({ concept, byMember, totalCents }) => (
                    <div key={concept} className="fixedTableRow">
                        <span className="fixedColConcept fixedCategoryLabel">
                            {concept}
                        </span>
                        <span className="fixedColTotal">
                            {formatCurrency(totalCents, currency)}
                        </span>
                        {members.map((m) => (
                            <span key={m.id} className="fixedColMember fixedAmount">
                                {byMember[m.id] ? formatCurrency(byMember[m.id], currency) : '—'}
                            </span>
                        ))}
                    </div>
                ))}

                {/* Total row */}
                <div className="fixedTableRow fixedTotalRow">
                    <span className="fixedColConcept">Grand total</span>
                    <span className="fixedColTotal">
                        {formatCurrency(grandTotal, currency)}
                    </span>
                    {members.map((m) => (
                        <span key={m.id} className="fixedColMember fixedTotalAmount">
                            {totalByMember[m.id]
                                ? formatCurrency(totalByMember[m.id], currency)
                                : formatCurrency(0, currency)}
                        </span>
                    ))}
                </div>
            </div>
        </section>
    )
}
