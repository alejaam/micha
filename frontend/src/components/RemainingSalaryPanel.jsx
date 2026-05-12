import { useEffect, useState } from 'react'
import { getRemainingSalary } from '../api'
import { formatCurrency } from '../utils'

/**
 * RemainingSalaryPanel — Displays how much money is left for the member
 * after all personal and shared expenses of the period.
 */
export function RemainingSalaryPanel({ householdId, memberId, period, currency = 'MXN' }) {
    const [data, setData] = useState(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')

    useEffect(() => {
        if (!householdId || !memberId || !period) return

        let cancelled = false
        const from = period.start_date || period.StartDate
        const to = period.end_date || period.EndDate

        setLoading(true)
        getRemainingSalary({ householdId, memberId, from, to })
            .then((res) => {
                if (!cancelled) setData(res)
            })
            .catch((err) => {
                if (!cancelled) setError(err.message)
            })
            .finally(() => {
                if (!cancelled) setLoading(false)
            })

        return () => { cancelled = true }
    }, [householdId, memberId, period])

    if (loading) return <div className="card salaryCard skeleton">Cargando finanzas...</div>
    if (error) return <div className="card salaryCard error">Error al cargar sueldo: {error}</div>
    if (!data) return null

    const safeData = {
        remaining_salary_cents: data?.remaining_salary_cents ?? 0,
        monthly_salary_cents: data?.monthly_salary_cents ?? 0,
        total_shared_outflow_cents: data?.total_shared_outflow_cents ?? 0,
        total_personal_outflow_cents: data?.total_personal_outflow_cents ?? 0,
    }

    const remainingCents = Number(safeData.remaining_salary_cents) || 0
    const isNegative = remainingCents < 0

    const hasMovements = (
        (Number(safeData.total_shared_outflow_cents) || 0) > 0 ||
        (Number(safeData.total_personal_outflow_cents) || 0) > 0
    )

    return (
        <section className={`card salaryCard ${isNegative ? 'salaryNegative' : ''}`}>
            <div className="salaryHeader">
                <h3 className="salaryTitle">Tu sueldo restante</h3>
                <span className="salaryPeriod">Este mes</span>
            </div>

            <div className="salaryMain">
                <strong className="salaryRemaining">{formatCurrency(remainingCents, currency)}</strong>
                <p className="salaryHint">Después de todos tus gastos</p>
            </div>

            <div className="salaryBreakdown">
                <div className="salaryRow">
                    <span>Sueldo base</span>
                    <span className="salaryVal">{formatCurrency(safeData.monthly_salary_cents, currency)}</span>
                </div>
                <div className="salaryRow">
                    <span>Gastos compartidos (tu parte)</span>
                    <span className="salaryVal negative">-{formatCurrency(safeData.total_shared_outflow_cents, currency)}</span>
                </div>
                <div className="salaryRow">
                    <span>Gastos personales</span>
                    <span className="salaryVal negative">-{formatCurrency(safeData.total_personal_outflow_cents, currency)}</span>
                </div>
                {!hasMovements && (
                    <p className="salaryHint">Sin movimientos este periodo.</p>
                )}
            </div>
        </section>
    )
}
