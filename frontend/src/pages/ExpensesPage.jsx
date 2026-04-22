import { motion } from 'framer-motion'
import { useCallback, useState } from 'react'
import { BottomSheet } from '../components/BottomSheet'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseList } from '../components/ExpenseList'
import { ExpenseModal } from '../components/ExpenseModal'
import { ExpenseSummary } from '../components/ExpenseSummary'
import { FAB } from '../components/FAB'
import { HistorySection } from '../components/HistorySection'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'

export function ExpensesPage() {
    const {
        members,
        loadingMembers,
        items,
        loadingList,
        settlement,
        currentMember,
        activeCurrency,
        householdId,
        isMutationLocked,
        closedPeriods,
        selectedPeriodKey,
        comparisonSeries,
        memberBalanceTrend,
        completedMsi,
        selectedPeriodSnapshot,
        loadingHistory,
        isProvisional,
        provisionalReason,
        setSelectedPeriodKey,
        handleCreate,
        handleSave,
        handleDelete,
        message,
        setMessage,
        error,
        setError,
        submittingCreate,
        savingId,
        deletingId,
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
                <section className="card dashboardSummaryCard" aria-label="Este mes">
                    <h2 className="sectionTitle">
                        <span className="sectionTitleIcon" aria-hidden>📊</span>
                        Resumen del periodo actual
                    </h2>
                    <ExpenseSummary settlement={settlement} currency={activeCurrency} />
                </section>

                <HistorySection
                    closedPeriods={closedPeriods}
                    selectedPeriodKey={selectedPeriodKey}
                    onSelectPeriod={setSelectedPeriodKey}
                    comparisonSeries={comparisonSeries}
                    memberBalanceTrend={memberBalanceTrend}
                    completedMsi={completedMsi}
                    selectedPeriodSnapshot={selectedPeriodSnapshot}
                    currency={activeCurrency}
                    isLoading={loadingHistory}
                    isProvisional={isProvisional}
                    provisionalReason={provisionalReason}
                    onQuickAdd={handleOpenQuickAdd}
                />
            </div>

            <div className="dashboardCol">
                <ExpenseList
                    items={items}
                    isLoading={loadingList}
                    deletingId={deletingId}
                    savingId={savingId}
                    onDelete={handleDelete}
                    onSave={handleSave}
                    currency={activeCurrency}
                    isMutationLocked={isMutationLocked}
                    onQuickAdd={handleOpenQuickAdd}
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
