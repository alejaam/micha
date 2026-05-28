import { motion } from 'framer-motion'
import { useEffect, useMemo, useState } from 'react'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { IncomesPanel } from '../components/IncomesPanel'
import { MembersPanel } from '../components/MembersPanel'
import { PeriodManagementPanel } from '../components/PeriodManagementPanel'
import { SettlementPanel } from '../components/SettlementPanel'
import { useAppShell } from '../context/AppShellContext'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'

export function BalancesPage() {
    const {
        currentPeriod,
        reloadPeriod,
        handleReload: reloadShell,
        isLoadingPeriod,
    } = useAppShell()

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

    const currentUserId = useMemo(() => {
        const token = localStorage.getItem('micha_token')
        if (!token) return ''
        try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            return payload.user_id || payload.sub || ''
        } catch { return '' }
    }, [])

    const isOwner = !selectedHousehold?.owner_id || selectedHousehold?.owner_id === currentUserId

    const [modalOpen, setModalOpen] = useState(false)

    useEffect(() => {
        loadSettlement()
    }, [loadSettlement])

    // Show loading state while period is being fetched
    if (isLoadingPeriod) {
        return (
            <section className="card" aria-label="Cargando periodo">
                <div style={{ padding: '2rem', textAlign: 'center' }}>
                    <p className="text-secondary">Cargando periodo...</p>
                </div>
            </section>
        )
    }

    return (
        <motion.div
            className="pageGrid"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2 }}
        >
            {error && <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner>}

            <PeriodManagementPanel
                householdId={householdId}
                period={currentPeriod}
                onStatusChange={({ message: statusMessage } = {}) => {
                    reloadPeriod()
                    reloadShell()
                    if (statusMessage) setMessage(statusMessage)
                }}
                isOwner={isOwner}
                members={members}
                currentUserMemberId={currentMember?.id}
            />

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
