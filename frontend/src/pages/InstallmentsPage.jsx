import { motion } from 'framer-motion'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { MSI as CardExpensesPanel } from '../components/CardExpensesPanel'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { FixedExpensesPanel } from '../components/FixedExpensesPanel'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'

export function InstallmentsPage() {
    const navigate = useNavigate()
    const {
        members,
        loadingMembers,
        items,
        recurringItems,
        settlement,
        currentMember,
        activeCurrency,
        householdId,
        msiProgress,
        handleCreate,
        message,
        setMessage,
        error,
        setError,
        submittingCreate,
    } = useHouseholdData()

    const [modalOpen, setModalOpen] = useState(false)

    return (
        <motion.div
            className="dashboardCol"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner>}

            <FixedExpensesPanel
                items={items}
                recurringItems={recurringItems}
                members={members}
                settlement={settlement}
                currency={activeCurrency}
            />

            <div className="u-flex u-flex-col u-gap-3 u-mb-3">
                <button
                    type="button"
                    className="btn btnPrimary"
                    onClick={() => navigate('/fixed-expenses')}
                >
                    Gestionar gastos fijos →
                </button>
            </div>

            <CardExpensesPanel
                items={items}
                members={members}
                msiProgress={msiProgress}
                currency={activeCurrency}
            />

            <FAB
                onClick={() => setModalOpen(true)}
            />

            {modalOpen && (
                <ExpenseModal
                    onClose={() => setModalOpen(false)}
                    onSubmit={async (payload) => {
                        const success = await handleCreate(payload)
                        if (success) setModalOpen(false)
                    }}
                    isSubmitting={submittingCreate}
                    members={members}
                    isLoadingMembers={loadingMembers}
                    defaultPaidByMemberId={currentMember?.id ?? ''}
                    householdId={householdId}
                />
            )}
        </motion.div>
    )
}
