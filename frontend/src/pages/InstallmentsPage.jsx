import { motion } from 'framer-motion'
import { useCallback, useState } from 'react'
import { BottomSheet } from '../components/BottomSheet'
import { CardExpensesPanel } from '../components/CardExpensesPanel'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { FixedExpensesPanel } from '../components/FixedExpensesPanel'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'

export function InstallmentsPage() {
    const {
        members,
        loadingMembers,
        items,
        recurringItems,
        currentMember,
        activeCurrency,
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
            setError('El periodo está bajo revisión o cerrado. Las acciones están deshabilitadas.')
            return
        }
        setQuickAddOpen(true)
    }, [isMutationLocked, setError])

    return (
        <motion.div
            className="dashboardCol"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            <FixedExpensesPanel
                items={items}
                recurringItems={recurringItems}
                members={members}
                currency={activeCurrency}
            />

            <CardExpensesPanel
                items={items}
                members={members}
                currency={activeCurrency}
            />

            <FAB
                onClick={() => {
                    if (isMutationLocked) {
                        setError('El periodo está bajo revisión o cerrado. Las acciones están deshabilitadas.')
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
                title="Añadir rápido"
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
