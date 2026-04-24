import { useCallback, useMemo } from 'react'
import { formatCurrency } from '../utils'

const FALLBACK_LABELS = {
    rent: 'Renta',
    auto: 'Auto',
    streaming: 'Streaming / Servicios',
    food: 'Comida',
    personal: 'Personal',
    savings: 'Ahorros',
    other: 'Otro',
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
 * @param {object|null} settlement - current settlement payload (optional)
 */
export function FixedExpensesPanel({ items = [], recurringItems = [], members = [], currency = 'MXN', categories = [], settlement = null }) {
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
        return categoryLabels[key] ?? FALLBACK_LABELS[key] ?? 'Otro'
    }, [categoryLabels])

    const settlementMode = settlement?.effective_settlement_mode || settlement?.settlement_mode || ''
    const settlementWeights = useMemo(() => {
        const weights = new Map()
        const settlementMembers = Array.isArray(settlement?.members) ? settlement.members : []

        for (const member of settlementMembers) {
            const weightBps = Number(member?.salary_weight_bps)
            if (!member?.member_id || Number.isNaN(weightBps) || weightBps <= 0) continue
            weights.set(member.member_id, weightBps)
        }

        return weights
    }, [settlement])

    const distributeAmount = useCallback((amountCents) => {
        const shares = {}

        if (members.length === 0) return shares

        const canUseProportional = settlementMode === 'proportional' && settlementWeights.size > 0

        if (canUseProportional) {
            const weights = members.map((member) => Number(settlementWeights.get(member.id) ?? 0))
            const totalWeight = weights.reduce((sum, weight) => sum + weight, 0)

            if (totalWeight > 0) {
                let allocated = 0
                const remainders = []

                weights.forEach((weight, index) => {
                    const numerator = amountCents * weight
                    const baseShare = Math.floor(numerator / totalWeight)
                    shares[members[index].id] = baseShare
                    allocated += baseShare
                    remainders.push({ index, remainder: numerator % totalWeight })
                })

                remainders
                    .sort((a, b) => b.remainder - a.remainder)
                    .forEach(({ index }) => {
                        if (allocated < amountCents) {
                            shares[members[index].id] = (shares[members[index].id] ?? 0) + 1
                            allocated += 1
                        }
                    })

                return shares
            }
        }

        const base = Math.floor(amountCents / members.length)
        const remainder = amountCents % members.length
        members.forEach((member, index) => {
            const extra = index < remainder ? 1 : 0
            shares[member.id] = base + extra
        })

        return shares
    }, [members, settlementMode, settlementWeights])

// Group by concept, accumulating totals per member.
    const grouped = useMemo(() => {
        const map = {}

        for (const e of fixedItems) {
            const concept = resolveConcept(e)
            if (!map[concept]) map[concept] = { concept, items: [], byMember: {}, totalCents: 0 }
            map[concept].items.push(e)
            map[concept].totalCents += e.amount_cents
            const shares = distributeAmount(e.amount_cents)
            for (const [memberId, amount] of Object.entries(shares)) {
                map[concept].byMember[memberId] = (map[concept].byMember[memberId] ?? 0) + amount
            }
        }
        return Object.values(map)
    }, [distributeAmount, fixedItems, resolveConcept])

    // Total per member across all fixed
    const totalByMember = useMemo(() => {
        const totals = {}
        for (const e of fixedItems) {
            const shares = distributeAmount(e.amount_cents)
            for (const [memberId, amount] of Object.entries(shares)) {
                totals[memberId] = (totals[memberId] ?? 0) + amount
            }
        }
        return totals
    }, [distributeAmount, fixedItems])

    const grandTotal = fixedItems.reduce((s, e) => s + e.amount_cents, 0)

    if (fixedItems.length === 0) {
        return (
            <section className="card" aria-label="Gastos fijos">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>#</span>
                    Gastos fijos
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">Sin gastos fijos aún</p>
                    <p className="emptyHint">Marca gastos como &quot;Fijos&quot; al añadirlos.</p>
                </div>
            </section>
        )
    }

    return (
        <section className="card" aria-label="Gastos fijos">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>#</span>
                Gastos fijos
                {fixedItems.length > 0 && <span className="sectionBadge">{fixedItems.length} {fixedItems.length === 1 ? 'gasto' : 'gastos'} • {formatCurrency(grandTotal, currency)}</span>}
            </h2>

            <div className="fixedExpensesTable">
                {/* Header row */}
                <div className="fixedTableHeader">
                    <span className="fixedColConcept">Concepto</span>
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
                    <span className="fixedColConcept">Gran total</span>
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
