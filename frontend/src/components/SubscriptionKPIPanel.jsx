import { useCallback, useEffect, useState } from 'react'
import { getSubscriptionKPI } from '../api'
import { useAuth } from '../context/AuthContext'
import { formatCurrency } from '../utils'

/**
 * SubscriptionKPIPanel — displays subscription savings and overlap warnings.
 *
 * @param {Object} props
 * @param {string} props.householdId
 * @param {string} props.currency  — default 'MXN'
 */
export function SubscriptionKPIPanel({ householdId, currency = 'MXN' }) {
    const { handleProtectedError } = useAuth()
    const [kpi, setKpi] = useState(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    const loadKPI = useCallback(async () => {
        if (!householdId) return
        setLoading(true)
        setError(null)
        try {
            const data = await getSubscriptionKPI({ householdId })
            setKpi(data)
        } catch (err) {
            handleProtectedError(err)
            setError(err.message ?? 'Error al cargar KPI')
        } finally {
            setLoading(false)
        }
    }, [householdId, handleProtectedError])

    useEffect(() => {
        loadKPI()
    }, [loadKPI])

    if (loading) {
        return (
            <section className="card" aria-label="KPI de suscripciones">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>📊</span>
                    Ahorro en suscripciones
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">Calculando...</p>
                </div>
            </section>
        )
    }

    if (error) {
        return (
            <section className="card" aria-label="KPI de suscripciones">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>📊</span>
                    Ahorro en suscripciones
                </h2>
                <p className="u-text-sm u-text-danger">{error}</p>
            </section>
        )
    }

    if (!kpi || (!kpi.total_spent_cents && !kpi.standalone_total_cents && !kpi.net_savings_cents)) {
        return (
            <section className="card" aria-label="KPI de suscripciones">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>📊</span>
                    Ahorro en suscripciones
                </h2>
                <div className="emptyState">
                    <p className="emptyTitle">Sin datos</p>
                    <p className="emptyHint">Asocia servicios a tus gastos fijos para ver el análisis.</p>
                </div>
            </section>
        )
    }

    const isPositiveSavings = kpi.net_savings_cents >= 0
    const savingsColor = isPositiveSavings ? 'var(--color-green-600, #059669)' : 'var(--color-red-600, #dc2626)'

    return (
        <section className="card" aria-label="KPI de suscripciones">
            <h2 className="sectionTitle">
                <span className="sectionTitleIcon" aria-hidden>📊</span>
                Ahorro en suscripciones
            </h2>

            <div className="kpiGrid">
                <div className="kpiItem">
                    <span className="kpiLabel">Gasto total mensual</span>
                    <span className="kpiValue">{formatCurrency(kpi.total_spent_cents, currency)}</span>
                </div>
                <div className="kpiItem">
                    <span className="kpiLabel">Costo individual</span>
                    <span className="kpiValue">{formatCurrency(kpi.standalone_total_cents, currency)}</span>
                </div>
                <div className="kpiItem">
                    <span className="kpiLabel">Ahorro neto</span>
                    <span className="kpiValue" style={{ color: savingsColor }}>
                        {isPositiveSavings ? '+' : ''}{formatCurrency(kpi.net_savings_cents, currency)}
                    </span>
                </div>
            </div>

            {kpi.overlap_warnings && kpi.overlap_warnings.length > 0 && (
                <div className="kpiOverlaps">
                    <h4 className="kpiOverlapsTitle">
                        ⚠️ Servicios duplicados ({kpi.overlap_warnings.length})
                    </h4>
                    <ul className="kpiOverlapsList">
                        {kpi.overlap_warnings.map((warning) => (
                            <li key={warning.catalog_service_id} className="kpiOverlapItem">
                                <span className="kpiOverlapService">{warning.service_name}</span>
                                <span className="kpiOverlapCount">
                                    {warning.recurring_expense_ids?.length ?? 0} gastos
                                </span>
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </section>
    )
}
