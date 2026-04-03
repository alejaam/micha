import { useMemo } from 'react'
import { formatCurrency } from '../utils'

const FALLBACK_LABELS = {
    rent: 'Renta',
    auto: 'Auto',
    streaming: 'Streaming / Servicios',
    food: 'Comida',
    personal: 'Personal',
    savings: 'Ahorro',
    other: 'Otros',
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
export function FixedExpensesPanel({ items = [], members = [], currency = 'MXN', categories = [] }) {
    // Build label map: prefer dynamic categories, fallback to hardcoded
    const categoryLabels = useMemo(() => {
        const labels = { ...FALLBACK_LABELS }
        for (const c of categories) {
            labels[c.id] = c.name
            labels[c.slug] = c.name
        }
        return labels
    }, [categories])

    const fixedItems = useMemo(
        () => items.filter((e) => e.expense_type === 'fixed' && e.is_shared),
        [items],
    )

    // Build member index for fast lookup
    const memberIndex = useMemo(
        () => Object.fromEntries(members.map((m) => [m.id, m.name])),
        [members],
    )

    // Group by category, accumulating totals per member
    const grouped = useMemo(() => {
        const map = {}
        for (const e of fixedItems) {
            const cat = e.category || 'other'
            if (!map[cat]) map[cat] = { category: cat, items: [], byMember: {} }
            map[cat].items.push(e)
            const mid = e.paid_by_member_id
            map[cat].byMember[mid] = (map[cat].byMember[mid] ?? 0) + e.amount_cents
        }
        return Object.values(map)
    }, [fixedItems])

    // Total per member across all fixed
    const totalByMember = useMemo(() => {
        const totals = {}
        for (const e of fixedItems) {
            totals[e.paid_by_member_id] = (totals[e.paid_by_member_id] ?? 0) + e.amount_cents
        }
        return totals
    }, [fixedItems])

    const grandTotal = fixedItems.reduce((s, e) => s + e.amount_cents, 0)

    if (fixedItems.length === 0) {
        return (
            <section className="card" aria-label="Gastos fijos mensuales">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>#</span>
                    Fijos
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">Sin gastos fijos</p>
                    <p className="emptyHint">Marca el gasto como fijo al registrarlo.</p>
                </div>
            </section>
        )
    }

    return (
        <section className="card" aria-label="Gastos fijos mensuales">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>#</span>
                Fijos
                {fixedItems.length > 0 && <span className="sectionBadge">{fixedItems.length} gastos • {formatCurrency(grandTotal, currency)}</span>}
            </h2>

            <div className="fixedExpensesTable">
                {/* Header row */}
                <div className="fixedTableHeader">
                    <span className="fixedColConcept">Concepto</span>
                    {members.map((m) => (
                        <span key={m.id} className="fixedColMember">{m.name}</span>
                    ))}
                </div>

                {/* Category rows */}
                {grouped.map(({ category, byMember }) => (
                    <div key={category} className="fixedTableRow">
                        <span className="fixedColConcept fixedCategoryLabel">
                            {categoryLabels[category] ?? category}
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
                    <span className="fixedColConcept">Total</span>
                    {members.map((m) => (
                        <span key={m.id} className="fixedColMember fixedTotalAmount">
                            {totalByMember[m.id]
                                ? formatCurrency(totalByMember[m.id], currency)
                                : formatCurrency(0, currency)}
                        </span>
                    ))}
                </div>

                <div className="fixedGrandTotal">
                    <span className="fixedGrandTotalLabel">Total general</span>
                    <span className="fixedGrandTotalValue">{formatCurrency(grandTotal, currency)}</span>
                </div>
            </div>
        </section>
    )
}
