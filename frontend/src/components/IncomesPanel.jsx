import { formatCurrency } from '../utils'

/**
 * Brand color palette for stacked bar segments — cycles through brand hues
 * to distinguish each member's contribution.
 */
const STACK_COLORS = [
    'var(--color-brand-500)',
    'var(--color-brand-400)',
    'var(--color-brand-600)',
    'var(--color-brand-300)',
    'var(--color-brand-700)',
    'var(--color-brand-200)',
]

function pctOf(total, part) {
    if (!total || !part) return 0
    return Math.round((part / total) * 10000) / 100
}

/**
 * IncomesPanel — shows each member's salary contribution as a single
 * horizontal stacked bar (100% width) with a legend below.
 *
 * Data comes from:
 *   - members[].monthly_salary_cents  (raw salary)
 *   - settlement.members[].salary_weight_bps  (% already calculated by backend)
 *
 * @param {Array} members - list of member objects from API
 * @param {object|null} settlement - settlement response from API
 * @param {string} currency
 */
export function IncomesPanel({ members = [], settlement = null, currency = 'MXN' }) {
    const totalSalaryCents = members.reduce((sum, m) => sum + (m.monthly_salary_cents ?? 0), 0)

    // Build a map from member_id to salary_weight_bps from settlement (more accurate)
    const weightMap = {}
    if (settlement?.members) {
        for (const sm of settlement.members) {
            weightMap[sm.member_id] = sm.salary_weight_bps ?? 0
        }
    }

    const hasData = members.length > 0 && totalSalaryCents > 0

    // Compute segments with percentage and color
    const segments = members.map((m, idx) => {
        const salary = m.monthly_salary_cents ?? 0
        const weightBps = weightMap[m.id]
        const pct = weightBps != null
            ? (weightBps / 100)
            : pctOf(totalSalaryCents, salary)
        return {
            id: m.id,
            name: m.name,
            salary,
            pct: Math.min(pct, 100),
            color: STACK_COLORS[idx % STACK_COLORS.length],
        }
    })

    return (
        <section className="card" aria-label="Ingresos de miembros">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>$</span>
                Ingresos
                {hasData && <span className="sectionBadge">{members.length} miembro{members.length !== 1 ? 's' : ''}</span>}
            </h2>

            {!hasData ? (
                <div className="emptyState">
                    <p className="emptyTitle">Sin datos de ingreso</p>
                    <p className="emptyHint">Agrega el salario mensual al crear miembros.</p>
                </div>
            ) : (
                <div className="stackedBarSection">
                    {/* Single horizontal stacked bar — 100% width */}
                    <div
                        className="stackedBar"
                        role="img"
                        aria-label={`Distribución de ingresos: ${segments.map(s => `${s.name} ${s.pct}%`).join(', ')}`}
                    >
                        {segments.map((s) => (
                            <div
                                key={s.id}
                                className="stackedBarSegment"
                                style={{
                                    width: `${Math.max(s.pct, 0.5)}%`,
                                    backgroundColor: s.color,
                                }}
                                aria-label={`${s.name}: ${s.pct}%`}
                                title={`${s.name}: ${s.pct}%`}
                            />
                        ))}
                    </div>

                    {/* Legend */}
                    <div className="stackedBarLegend">
                        {segments.map((s) => (
                            <div key={s.id} className="stackedBarLegendRow">
                                <span className="stackedBarLegendSwatch" style={{ backgroundColor: s.color }} />
                                <span className="stackedBarLegendName">{s.name}</span>
                                <span className="stackedBarLegendAmount">{formatCurrency(s.salary, currency)}</span>
                                <span className="stackedBarLegendPct">{s.pct}%</span>
                            </div>
                        ))}
                        <div className="stackedBarLegendRow stackedBarLegendTotal">
                            <span className="stackedBarLegendName">Ingreso total</span>
                            <span className="stackedBarLegendAmount">{formatCurrency(totalSalaryCents, currency)}</span>
                            <span className="stackedBarLegendPct">100%</span>
                        </div>
                    </div>
                </div>
            )}
        </section>
    )
}
