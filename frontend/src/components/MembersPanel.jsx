import { formatCurrency } from '../utils'

/**
 * MembersPanel — shows household members with their income weight.
 * Visibility: Who has access to the household.
 */
export function MembersPanel({ members = [], currency = 'MXN' }) {
    const hasMembers = members.length > 0

    if (!hasMembers) {
        return null
    }

    // Calculate total salary for percentage display
    const totalSalaryCents = members.reduce((sum, m) => sum + (m.monthly_salary_cents || 0), 0)

    return (
        <section className="card" aria-label="Miembros del hogar">
            <div className="listHeader">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>👥</span>
                    Miembros
                </h2>
                <span className="listCount">{members.length}</span>
            </div>
            <div className="membersList">
                {members.map((member) => {
                    const salaryCents = member.monthly_salary_cents || 0
                    const weightPct = totalSalaryCents > 0 && salaryCents > 0
                        ? ((salaryCents / totalSalaryCents) * 100).toFixed(1)
                        : null

                    return (
                        <div key={member.id} className="memberRow">
                            <div className="memberInfo">
                                <span className="memberName">{member.name}</span>
                                {member.email && (
                                    <span className="memberEmail">{member.email}</span>
                                )}
                            </div>
                            <div className="memberStats">
                                {salaryCents > 0 ? (
                                    <>
                                        <span className="memberSalary">
                                            {formatCurrency(salaryCents, currency)}/mes
                                        </span>
                                        {weightPct && (
                                            <span className="memberWeight">
                                                {weightPct}% peso
                                            </span>
                                        )}
                                    </>
                                ) : (
                                    <span className="memberNoSalary">sin salario</span>
                                )}
                            </div>
                        </div>
                    )
                })}
            </div>
        </section>
    )
}