import { formatCurrency } from '../utils'

/**
 * IncomesPanel — mirrors the "SUELDO" section from the Excel.
 * Shows each member's monthly salary, their contribution percentage,
 * and the combined household income.
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

    return (
        <section className="card" aria-label="Member incomes">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>$</span>
                Incomes
                {hasData && <span className="sectionBadge">{members.length} member{members.length !== 1 ? 's' : ''}</span>}
            </h2>

            {!hasData ? (
                <div className="emptyState">
                    <p className="emptyTitle">No salary data</p>
                    <p className="emptyHint">Add monthly salary when creating members.</p>
                </div>
            ) : (
                <div className="incomesGrid">
                    {members.map((m) => {
                        const salary = m.monthly_salary_cents ?? 0
                        const weightBps = weightMap[m.id]
                        // Fall back to manual calculation if settlement hasn't loaded yet
                        const pct = weightBps != null
                            ? (weightBps / 100).toFixed(2)
                            : totalSalaryCents > 0
                                ? ((salary / totalSalaryCents) * 100).toFixed(2)
                                : '0.00'

                        return (
                            <div key={m.id} className="incomeMemberCard">
                                <div className="incomeMemberHeader">
                                    <span className="incomeMemberName">{m.name}</span>
                                    <span className="incomePct">{pct}%</span>
                                </div>
                                <span className="incomeSalary">{formatCurrency(salary, currency)}</span>
                                <div className="incomeBar">
                                    <div
                                        className="incomeBarFill"
                                        style={{ width: `${Math.min(parseFloat(pct), 100)}%` }}
                                        aria-label={`${pct}% of household income`}
                                    />
                                </div>
                            </div>
                        )
                    })}
                    <div className="incomeTotalRow">
                        <span className="incomeTotalLabel">Combined income</span>
                        <span className="incomeTotalValue">{formatCurrency(totalSalaryCents, currency)}</span>
                    </div>
                </div>
            )}
        </section>
    )
}
