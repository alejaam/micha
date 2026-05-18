import { useState } from 'react'
import { transitionPeriodToReview, approvePeriod, closePeriod, initializePeriod } from '../api'
import { ConsensusProgressRing } from './ConsensusProgressRing'

const monthNames = [
    'enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
    'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre',
]

/**
 * PeriodManagementPanel — UI for managing the period lifecycle.
 *
 * States:
 * - none (init): "Start current month" button (Owner only).
 * - open: "Propose closure" button.
 * - review: Voting UI (Approve/Object) + Progress ring.
 * - owner only (in review): "Final closure" button.
 */
export function PeriodManagementPanel({
    householdId,
    period,
    onStatusChange,
    isOwner = false,
    consensus,
}) {
    const [submitting, setSubmitting] = useState(false)
    const [error, setError] = useState('')

    if (!householdId) return null

    const handleInitialize = async () => {
        try {
            setSubmitting(true)
            setError('')
            await initializePeriod({ householdId })
            onStatusChange()
        } catch (err) {
            // If it already exists, just refresh to show the management UI
            if (err.message?.includes('already has periods')) {
                onStatusChange()
                return
            }
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const handleStartReview = async () => {
        if (!period?.id && !period?.ID) {
            setError('No hay periodo activo')
            return
        }
        try {
            setSubmitting(true)
            setError('')
            await transitionPeriodToReview({ householdId, periodId: period?.id || period?.ID })
            onStatusChange()
        } catch (err) {
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const handleVote = async (voteStatus) => {
        if (!period?.id && !period?.ID) {
            setError('No hay periodo activo')
            return
        }
        try {
            setSubmitting(true)
            setError('')
            await approvePeriod({ householdId, periodId: period?.id || period?.ID, status: voteStatus })
            onStatusChange()
        } catch (err) {
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const handleFinalClose = async (force = false) => {
        if (!period?.id && !period?.ID) {
            setError('No hay periodo activo')
            return
        }
        try {
            setSubmitting(true)
            setError('')
            await closePeriod({ householdId, periodId: period?.id || period?.ID, force })
            const now = new Date()
            const currentMonthName = monthNames[now.getMonth()]
            const nextMonthName = monthNames[(now.getMonth() + 1) % 12]
            onStatusChange({ message: `Periodo de ${currentMonthName} cerrado. Bienvenido a ${nextMonthName}.` })
        } catch (err) {
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const status = period?.Status || period?.status || 'open'

    // ─── Banner mode: No active period OR review ───
    const isBanner = !period || status === 'review'

    // ─── Render: No active period ───
    if (!period || status === 'closed') {
        if (!isOwner) return null

        return (
            <section className={`card periodActionCard ${isBanner ? 'periodActionCard--banner' : ''}`}>
                <div className="periodActionContent">
                    <div>
                        <h3 className="sectionTitle">Comenzar seguimiento</h3>
                        <p className="authMeta">
                            Parece que este hogar aún no tiene un periodo activo. Inicializa el mes actual para empezar.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary"
                        onClick={handleInitialize}
                        disabled={submitting}
                    >
                        {submitting ? 'Iniciando...' : 'Empezar mes actual'}
                    </button>
                </div>
                {error && <p className="formHint formHintError">{error}</p>}
            </section>
        )
    }

    // ─── Render: Open period (compact card) ───
    if (status === 'open') {
        if (!isOwner) return null

        return (
            <section className="card periodActionCard">
                <div className="periodActionContent">
                    <div>
                        <h3 className="sectionTitle">Cierre de periodo</h3>
                        <p className="authMeta">
                            ¿Terminaron de registrar los gastos del mes? Inicia la revisión para conciliar saldos.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary"
                        onClick={handleStartReview}
                        disabled={submitting}
                    >
                        {submitting ? 'Iniciando...' : 'Iniciar revisión'}
                    </button>
                </div>
                {error && <p className="formHint formHintError">{error}</p>}
            </section>
        )
    }

    // ─── Render: Review period (banner mode) ───
    if (status === 'review') {
        return (
            <section className={`card periodActionCard ${isBanner ? 'periodActionCard--banner' : ''}`}>
                <div className="periodReviewGrid">
                    <div className="periodReviewInfo">
                        <h3 className="sectionTitle">Periodo en revisión</h3>
                        <p className="authMeta">
                            Revisa el resumen de gastos y aprueba si estás de acuerdo con el balance.
                        </p>
                        
                        <div className="periodVoteActions">
                            <button
                                type="button"
                                className="btn btnPrimary btnSm"
                                onClick={() => handleVote('approved')}
                                disabled={submitting}
                            >
                                👍 Aprobar
                            </button>
                            <button
                                type="button"
                                className="btn btnGhost btnSm"
                                onClick={() => handleVote('objected')}
                                disabled={submitting}
                            >
                                👎 Objetar
                            </button>
                        </div>
                    </div>

                    <div className="periodConsensusBox">
                        <ConsensusProgressRing
                            approved={consensus?.approved ?? 0}
                            total={consensus?.total ?? 0}
                            label="Consenso"
                        />
                        <span className="consensusLabel">Consenso</span>
                        {consensus && consensus.total > 0 && (
                            <span className="consensusMeta">{consensus.approved} de {consensus.total} aprobaron</span>
                        )}
                    </div>
                </div>

                {isOwner && (
                    <div className="ownerActionZone">
                        <p className="formHint">Como owner, puedes cerrar el periodo definitivamente una vez haya consenso.</p>
                        <button
                            type="button"
                            className="btn btnPrimary btnFull"
                            onClick={() => handleFinalClose(false)}
                            disabled={submitting}
                        >
                            {submitting ? 'Cerrando...' : 'Finalizar y abrir nuevo mes'}
                        </button>
                    </div>
                )}
                {error && <p className="formHint formHintError">{error}</p>}
            </section>
        )
    }

    return null
}
