import { useEffect, useState } from 'react'
import { listPeriods } from '../api'

export function PeriodHistory({ householdId }) {
    const [periods, setPeriods] = useState([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')

    useEffect(() => {
        if (!householdId) return

        async function load() {
            setLoading(true)
            try {
                const data = await listPeriods({ householdId })
                setPeriods(Array.isArray(data) ? data : [])
            } catch (err) {
                setError(err.message)
            } finally {
                setLoading(false)
            }
        }
        load()
    }, [householdId])

    if (loading && periods.length === 0) return <p>Cargando historial...</p>
    if (error) return <p className="text-error">{error}</p>
    if (periods.length === 0) return null

    const closedPeriods = periods.filter(p => (p.status || p.Status) === 'closed')

    if (closedPeriods.length === 0) return null

    // Filter out null/undefined periods to prevent errors
    const validClosedPeriods = closedPeriods.filter(Boolean)

    if (validClosedPeriods.length === 0) return null

    return (
        <section className="card" aria-label="Historial de periodos">
            <h2 className="sectionTitle">Historial de periodos</h2>
            <div className="periodHistoryList">
                {validClosedPeriods.map((p) => {
                    const startDate = p?.start_date || p?.StartDate
                    const endDate = p?.end_date || p?.EndDate
                    if (!startDate || !endDate) return null
                    return (
                        <article key={p.id} className="periodHistoryItem">
                            <div className="periodDates">
                                <strong>{new Date(startDate).toLocaleDateString()}</strong>
                                <span> al </span>
                                <strong>{new Date(endDate).toLocaleDateString()}</strong>
                            </div>
                            <span className="badge badgeClosed">Cerrado</span>
                        </article>
                    )
                })}
            </div>
        </section>
    )
}
