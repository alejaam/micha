import { useState } from 'react'
import { simulateClosePeriod, initializePeriod } from '../api'

const monthNames = [
    'enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
    'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre',
]

/**
 * PeriodManagementPanel — UI for managing the period lifecycle.
 *
 * States:
 * - none (init): "Start current month" button (Owner only).
 * - open: "Simulate Close" button (read-only projection).
 * - closed: message showing period is closed.
 */
export function PeriodManagementPanel({
    householdId,
    period,
    onStatusChange,
    isOwner = false,
}) {
    const [submitting, setSubmitting] = useState(false)
    const [error, setError] = useState('')
    const [simulation, setSimulation] = useState(null)

    if (!householdId) return null

    const handleInitialize = async () => {
        try {
            setSubmitting(true)
            setError('')
            await initializePeriod({ householdId })
            onStatusChange()
        } catch (err) {
            if (err.message?.includes('already has periods')) {
                onStatusChange()
                return
            }
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const handleSimulateClose = async () => {
        if (!period?.id && !period?.ID) {
            setError('No hay periodo activo')
            return
        }
        try {
            setSubmitting(true)
            setError('')
            const result = await simulateClosePeriod({ householdId, periodId: period?.id || period?.ID })
            setSimulation(result)
        } catch (err) {
            if (err.code === 'FUTURE_PERIOD') {
                setError('No puedes cerrar este periodo porque el siguiente comenzaría en el futuro.')
            } else if (err.code === 'PERIOD_TOO_SHORT') {
                setError('El periodo debe estar abierto al menos 7 días antes de cerrar.')
            } else {
                setError(err.message)
            }
        } finally {
            setSubmitting(false)
        }
    }

    const status = period?.Status || period?.status || 'open'

    // ─── Render: No active period ───
    if (!period || status === 'closed') {
        if (!isOwner) return null

        return (
            <section className="card periodActionCard periodActionCard--banner">
                <div className="periodActionContent">
                    <div>
                        <h3 className="sectionTitle">Comenzar seguimiento</h3>
                        <p className="authMeta">
                            Parece que este hogar aún no tiene un periodo activo. Inicializa el periodo actual para empezar.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary"
                        onClick={handleInitialize}
                        disabled={submitting}
                    >
                        {submitting ? 'Iniciando...' : 'Empezar periodo actual'}
                    </button>
                </div>
                {error && <p className="formHint formHintError">{error}</p>}
            </section>
        )
    }

    // ─── Render: Open period ───
    if (status === 'open') {
        if (!isOwner) return null

        return (
            <section className="card periodActionCard">
                <div className="periodActionContent">
                    <div>
                        <h3 className="sectionTitle">Cierre de periodo</h3>
                        <p className="authMeta">
                            Simula el cierre del periodo para ver una proyección del siguiente periodo, gastos fijos a arrastrar y saldos.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary"
                        onClick={handleSimulateClose}
                        disabled={submitting}
                    >
                        {submitting ? 'Calculando...' : 'Simular cierre'}
                    </button>
                </div>

                {simulation && (
                    <div className="simulationResult u-mt-4">
                        <h4 className="sectionTitle">Proyección de cierre</h4>
                        <div className="simulationGrid">
                            <div className="simulationItem">
                                <span className="simulationLabel">Siguiente periodo</span>
                                <span className="simulationValue">
                                    {new Date(simulation.next_period_start).toLocaleDateString()} — {new Date(simulation.next_period_end).toLocaleDateString()}
                                </span>
                            </div>
                            <div className="simulationItem">
                                <span className="simulationLabel">Gastos fijos a arrastrar</span>
                                <span className="simulationValue">{simulation.fixed_expense_count}</span>
                            </div>
                            <div className="simulationItem">
                                <span className="simulationLabel">Meses sin intereses</span>
                                <span className="simulationValue">{simulation.installment_count}</span>
                            </div>
                        </div>
                    </div>
                )}

                {error && <p className="formHint formHintError">{error}</p>}
            </section>
        )
    }

    return null
}
