import { motion } from 'framer-motion'
import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BottomSheet } from '../components/BottomSheet'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'

export function RulesPage() {
    const navigate = useNavigate()
    const {
        members,
        loadingMembers,
        currentMember,
        householdId,
        isMutationLocked,
        handleCreate,
        message,
        setMessage,
        error,
        setError,
        submittingCreate,
    } = useHouseholdData()

    const [modalOpen, setModalOpen] = useState(false)
    const [quickAddOpen, setQuickAddOpen] = useState(false)

    const handleOpenQuickAdd = useCallback(() => {
        if (isMutationLocked) {
            setError('Period is under review or closed. Mutating actions are disabled.')
            return
        }
        setQuickAddOpen(true)
    }, [isMutationLocked, setError])

    return (
        <motion.div
            className="pageGrid"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            <div className="dashboardCol">
                <section className="card" aria-label="Acciones rápidas de tarjetas">
                    <div className="listHeader">
                        <h2 className="listTitle">Tarjetas</h2>
                    </div>
                    <p className="text-sm text-dim mb-3">
                        Agrega una nueva tarjeta o cambia tu tarjeta preferida para nuevos gastos.
                    </p>
                    <div className="flex flex-col gap-3">
                        <button
                            type="button"
                            className="btn btnPrimary"
                            onClick={() => navigate('/onboarding/cards')}
                            disabled={isMutationLocked}
                        >
                            Gestionar tarjetas
                        </button>
                    </div>
                </section>

                <section className="card" aria-label="Gestión de gastos fijos">
                    <div className="listHeader">
                        <h2 className="listTitle">Gastos Fijos</h2>
                    </div>
                    <p className="text-sm text-dim mb-3">
                        Configura gastos mensuales recurrentes como renta, servicios o suscripciones.
                    </p>
                    <div className="flex flex-col gap-3">
                        <button
                            type="button"
                            className="btn"
                            onClick={() => navigate('/onboarding/fixed-expenses')}
                            disabled={isMutationLocked}
                        >
                            Gestionar gastos fijos
                        </button>
                    </div>
                </section>
            </div>

            <div className="dashboardCol">
                <section className="card" aria-label="Ajustes del hogar">
                    <div className="listHeader">
                        <h2 className="listTitle">Hogar</h2>
                    </div>
                    <p className="text-sm text-dim mb-3">
                        Gestiona el nombre de tu hogar, moneda y miembros.
                    </p>
                    <div className="flex flex-col gap-3">
                        <button
                            type="button"
                            className="btn"
                            onClick={() => navigate('/onboarding/household')}
                            disabled={isMutationLocked}
                        >
                            Editar ajustes del hogar
                        </button>
                        <button
                            type="button"
                            className="btn"
                            onClick={() => navigate('/members/new')}
                            disabled={isMutationLocked}
                        >
                            Invitar nuevos miembros
                        </button>
                    </div>
                </section>
            </div>

            <FAB
                onClick={() => {
                    if (isMutationLocked) {
                        setError('Period is under review or closed. Mutating actions are disabled.')
                        return
                    }
                    setModalOpen(true)
                }}
                disabled={isMutationLocked}
            />

            {modalOpen && (
                <ExpenseModal
                    onClose={() => setModalOpen(false)}
                    onSubmit={async (payload) => {
                        const success = await handleCreate(payload)
                        if (success) setModalOpen(false)
                    }}
                    isSubmitting={submittingCreate}
                    isMutationLocked={isMutationLocked}
                    members={members}
                    isLoadingMembers={loadingMembers}
                    defaultPaidByMemberId={currentMember?.id ?? ''}
                    householdId={householdId}
                />
            )}

            <BottomSheet
                open={quickAddOpen}
                title="Quick add"
                onClose={() => setQuickAddOpen(false)}
            >
                <ExpenseForm
                    onSubmit={async (payload) => {
                        const success = await handleCreate(payload)
                        if (success) setQuickAddOpen(false)
                    }}
                    isSubmitting={submittingCreate}
                    isLoadingMembers={loadingMembers}
                    members={members}
                    defaultPaidByMemberId={currentMember?.id ?? ''}
                />
            </BottomSheet>
        </motion.div>
    )
}
