import { motion } from 'framer-motion'
import { useCallback, useState } from 'react'
import { BottomSheet } from '../components/BottomSheet'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { IncomesPanel } from '../components/IncomesPanel'
import { MembersPanel } from '../components/MembersPanel'
import { SettlementPanel } from '../components/SettlementPanel'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'

export function BalancesPage() {
    const {
        members,
        loadingMembers,
        settlement,
        loadingSettlement,
        settlementYear,
        settlementMonth,
        currentMember,
        memberIndex,
        activeCurrency,
        householdId,
        selectedHousehold,
        isMutationLocked,
        fixedTotalCents,
        loadSettlement,
        setSettlementYear,
        setSettlementMonth,
        resetToCurrentMonth,
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
            className="pageGrid"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            <div className="dashboardCol">
                <SettlementPanel
                    settlement={settlement}
                    settlementYear={settlementYear}
                    settlementMonth={settlementMonth}
                    onSettlementYearChange={setSettlementYear}
                    onSettlementMonthChange={setSettlementMonth}
                    onRefresh={loadSettlement}
                    onResetToCurrentMonth={resetToCurrentMonth}
                    loadingSettlement={loadingSettlement}
                    memberIndex={memberIndex}
                    currency={activeCurrency}
                    selectedHousehold={selectedHousehold}
                    fixedTotalCents={fixedTotalCents}
                />
            </div>

            <div className="dashboardCol">
                <MembersPanel
                    members={members}
                    currency={activeCurrency}
                />

                <IncomesPanel
                    members={members}
                    settlement={settlement}
                    currency={activeCurrency}
                />
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
