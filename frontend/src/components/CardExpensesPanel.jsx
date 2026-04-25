import { useMemo } from 'react'
import { formatCurrency } from '../utils'

/**
 * CardExpensesPanel — mirrors the "Tarjetas" section from the Excel.
 * Groups card expenses (payment_method === 'card') by card_name and shows
 * each member's total side-by-side.
 *
 * @param {Array} items - expense list from API
 * @param {Array} members - member list from API
 * @param {string} currency
 */
export function CardExpensesPanel({ items = [], members = [], currency = 'MXN' }) {
    const cardItems = useMemo(
        () => items.filter((e) => e.payment_method === 'card'),
        [items],
    )

    // Group by card_name, accumulating totals per member
    const grouped = useMemo(() => {
        const map = {}
        for (const e of cardItems) {
            const card = e.card_name || 'Desconocida'
            if (!map[card]) map[card] = { cardName: card, byMember: {} }
            const mid = e.paid_by_member_id
            map[card].byMember[mid] = (map[card].byMember[mid] ?? 0) + e.amount_cents
        }
        // Sort alphabetically by card name
        return Object.values(map).sort((a, b) => a.cardName.localeCompare(b.cardName))
    }, [cardItems])

    // Total per member across all card expenses
    const totalByMember = useMemo(() => {
        const totals = {}
        for (const e of cardItems) {
            totals[e.paid_by_member_id] = (totals[e.paid_by_member_id] ?? 0) + e.amount_cents
        }
        return totals
    }, [cardItems])

    const grandTotal = cardItems.reduce((s, e) => s + e.amount_cents, 0)

    if (cardItems.length === 0) {
        return (
            <section className="card" aria-label="Gastos de tarjeta">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>&#9646;</span>
                    Gastos de tarjeta
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">Sin gastos de tarjeta aún</p>
                    <p className="emptyHint">Marca gastos como pagados con &quot;Tarjeta&quot; al añadirlos.</p>
                </div>
            </section>
        )
    }

    return (
        <section className="card" aria-label="Gastos de tarjeta">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>&#9646;</span>
                Gastos de tarjeta
                {cardItems.length > 0 && <span className="sectionBadge">{cardItems.length} {cardItems.length === 1 ? 'cargo' : 'cargos'} • {formatCurrency(grandTotal, currency)}</span>}
            </h2>

            <div className="cardExpensesTable">
                {/* Header row */}
                <div className="cardTableHeader">
                    <span className="cardColConcept">Tarjeta</span>
                    {members.map((m) => (
                        <span key={m.id} className="cardColMember">{m.name}</span>
                    ))}
                </div>

                {/* Card rows */}
                {grouped.map(({ cardName, byMember }) => (
                    <div key={cardName} className="cardTableRow">
                        <span className="cardColConcept cardCardLabel">{cardName}</span>
                        {members.map((m) => (
                            <span key={m.id} className="cardColMember cardAmount">
                                {byMember[m.id] ? formatCurrency(byMember[m.id], currency) : '—'}
                            </span>
                        ))}
                    </div>
                ))}

                {/* Total row */}
                <div className="cardTableRow cardTotalRow">
                    <span className="cardColConcept">Total</span>
                    {members.map((m) => (
                        <span key={m.id} className="cardColMember cardTotalAmount">
                            {totalByMember[m.id]
                                ? formatCurrency(totalByMember[m.id], currency)
                                : formatCurrency(0, currency)}
                        </span>
                    ))}
                </div>

                <div className="cardGrandTotal">
                    <span className="cardGrandTotalLabel">Gran total</span>
                    <span className="cardGrandTotalValue">{formatCurrency(grandTotal, currency)}</span>
                </div>
            </div>
        </section>
    )
}
