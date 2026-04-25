import { useState } from 'react'
import { transitionPeriodToReview, approvePeriod, closePeriod, initializePeriod } from '../api'
import { ConsensusProgressRing } from './ConsensusProgressRing'

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
    members = [],
    currentUserMemberId = '',
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
        try {
            setSubmitting(true)
            setError('')
            await transitionPeriodToReview({ householdId, periodId: period.id || period.ID })
            onStatusChange()
        } catch (err) {
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const handleVote = async (voteStatus) => {
        try {
            setSubmitting(true)
            setError('')
            await approvePeriod({ householdId, periodId: period.id || period.ID, status: voteStatus })
            onStatusChange()
        } catch (err) {
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    const handleFinalClose = async (force = false) => {
        try {
            setSubmitting(true)
            setError('')
            await closePeriod({ householdId, periodId: period.id || period.ID, force })
            onStatusChange()
        } catch (err) {
            setError(err.message)
        } finally {
            setSubmitting(false)
        }
    }

    // ─── Render: No active period ───
    if (!period) {
        if (!isOwner) return null

        return (
            <section className="card periodActionCard">
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

    const status = period.Status || period.status || 'open'

    // ─── Render: Open period ───
    if (status === 'open') {
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

    // ─── Render: Review period ───
    if (status === 'review') {
        return (
            <section className="card periodActionCard reviewMode">
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
                        <ConsensusProgressRing percent={50} /> {/* TODO: Real consensus data */}
                        <span className="consensusLabel">Consenso</span>
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
