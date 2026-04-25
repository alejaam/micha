import { AnimatePresence, motion } from 'framer-motion'
import { useCallback, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BottomSheet } from '../components/BottomSheet'
import { DynamicChartsPanel } from '../components/DynamicChartsPanel'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { ExpenseSummary } from '../components/ExpenseSummary'
import { FAB } from '../components/FAB'
import { MembersPanel } from '../components/MembersPanel'
import { PeriodManagementPanel } from '../components/PeriodManagementPanel'
import { RecentExpenses } from '../components/RecentExpenses'
import { RemainingSalaryPanel } from '../components/RemainingSalaryPanel'
import { useAppShell } from '../context/AppShellContext'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'
import { EmptyState } from '../ui/EmptyState'

export function DashboardPage() {
    const navigate = useNavigate()
    const {
        currentPeriod,
        reloadPeriod,
        selectedHousehold,
        handleReload: reloadShell,
    } = useAppShell()

    const {
        members,
        loadingMembers,
        items,
        loadingList,
        recurringItems,
        settlement,
        currentMember,
        activeCurrency,
        householdId,
        isMutationLocked,
        categoryTotals,
        memberActualVsExpected,
        msiProgress,
        spendingTrend,
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

    // Permissive owner check: if no owner is set in DB yet, anyone can bootstrap the period
    const isOwner = !selectedHousehold?.owner_id || selectedHousehold?.owner_id === currentUserId

    const [modalOpen, setModalOpen] = useState(false)
    const [quickAddOpen, setQuickAddOpen] = useState(false)

    const handleOpenQuickAdd = useCallback(() => {
        if (isMutationLocked) {
            setError('El periodo está bajo revisión o cerrado. Las acciones están deshabilitadas.')
            return
        }
        setQuickAddOpen(true)
    }, [isMutationLocked, setError])

    // Redirect to onboarding if needed
    if (!householdId) {
        return (
            <div className="dashboardOnboarding">
                <div className="onboardingCard card">
                    <div className="onboardingHeader">
                        <p className="authEyebrow">Comenzando</p>
                        <h2 className="authTitle">Configura tu hogar</h2>
                        <p className="authMeta">
                            Necesitas un hogar antes de poder registrar gastos.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary btnFull"
                        onClick={() => navigate('/onboarding/household')}
                    >
                        Crear hogar →
                    </button>
                </div>
            </div>
        )
    }

    const hasRecurringFixed = recurringItems.some((item) => item.expense_type === 'fixed')
    const hasExpenses = items.length > 0 || hasRecurringFixed
    const transferCount = settlement?.transfers?.length ?? 0
    const totalSharedCents = settlement?.total_shared_cents ?? 0
    const openInstallmentsCount = items.filter((item) => Number(item.total_installments) > 1).length

    return (
        <>
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner>}

            {/* ─── Period Management (Always visible) ─── */}
            <PeriodManagementPanel
                householdId={householdId}
                period={currentPeriod}
                onStatusChange={() => {
                    reloadPeriod()
                    reloadShell()
                }}
                isOwner={isOwner}
                members={members}
                currentUserMemberId={currentMember?.id}
            />

            {!hasExpenses && !loadingList ? (
                <section className="card dashboardEmptyState" aria-label="Sin gastos aún">
                    <EmptyState
                        title="Sin gastos aún"
                        description="Usa añadir rápido para registrar tu primer gasto y desbloquear los tableros."
                        ctaLabel="Añadir rápido"
                        onCta={handleOpenQuickAdd}
                        icon="[+]"
                    />
                </section>
            ) : (
                <>
                    <section className="card dashboardPriorityStrip" aria-label="Prioridades financieras">
                        <header className="dashboardPriorityHead">
                            <p className="dashboardPriorityEyebrow">Resumen</p>
                            <h2 className="dashboardPriorityTitle">Balances y conciliación primero</h2>
                            <p className="authMeta">Sigue el flujo del hogar: registra, concilia y cierra el periodo.</p>
                        </header>
                        <div className="dashboardPriorityMetrics" role="list" aria-label="Métricas prioritarias">
                            <article className="dashboardPriorityMetric" role="listitem">
                                <span className="dashboardPriorityLabel">Transferencias pendientes</span>
                                <strong className="dashboardPriorityValue">{transferCount}</strong>
                            </article>
                            <article className="dashboardPriorityMetric" role="listitem">
                                <span className="dashboardPriorityLabel">Total compartido</span>
                                <strong className="dashboardPriorityValue">
                                    {new Intl.NumberFormat(undefined, { style: 'currency', currency: activeCurrency }).format(totalSharedCents / 100)}
                                </strong>
                            </article>
                            <article className="dashboardPriorityMetric" role="listitem">
                                <span className="dashboardPriorityLabel">Plazos abiertos</span>
                                <strong className="dashboardPriorityValue">{openInstallmentsCount}</strong>
                            </article>
                        </div>
                    </section>

                    <motion.div
                        className="pageGrid"
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.25, ease: "easeOut", staggerChildren: 0.1 }}
                    >
                        <motion.div
                            className="dashboardCol"
                            initial={{ opacity: 0, y: 8 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ duration: 0.2, ease: "easeOut" }}
                        >
                            {currentMember && (
                                <RemainingSalaryPanel
                                    householdId={householdId}
                                    memberId={currentMember.id}
                                    period={currentPeriod}
                                    currency={activeCurrency}
                                />
                            )}
                            <section className="card dashboardSummaryCard" aria-label="Resumen del mes">
                                <h2 className="sectionTitle">
                                    <span className="sectionTitleIcon" aria-hidden>📊</span>
                                    Este mes
                                </h2>
                                <ExpenseSummary settlement={settlement} currency={activeCurrency} />
                            </section>

                            <MembersPanel
                                members={members}
                                currency={activeCurrency}
                            />

                            <button
                                type="button"
                                className="btn btnPrimary"
                                onClick={() => navigate('/balances')}
                                style={{ marginTop: '1rem' }}
                            >
                                Ver Balances →
                            </button>
                        </motion.div>

                        <motion.div
                            className="dashboardCol"
                            initial={{ opacity: 0, y: 8 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ duration: 0.2, ease: "easeOut", delay: 0.05 }}
                        >
                            <DynamicChartsPanel
                                categoryTotals={categoryTotals}
                                memberActualVsExpected={memberActualVsExpected}
                                msiProgress={msiProgress}
                                spendingTrend={spendingTrend}
                                currency={activeCurrency}
                            />

                            <section className="card" aria-label="Gastos recientes">
                                <div className="listHeader">
                                    <h2 className="listTitle">Gastos recientes</h2>
                                    {items.length > 0 && (
                                        <span className="listCount">{items.length} total</span>
                                    )}
                                </div>
                                <RecentExpenses
                                    items={items}
                                    isLoading={loadingList}
                                    currency={activeCurrency}
                                    limit={5}
                                    onQuickAdd={handleOpenQuickAdd}
                                />
                                <button
                                    type="button"
                                    className="btn btnGhost"
                                    onClick={() => navigate('/expenses')}
                                    style={{ width: '100%', marginTop: '0.5rem' }}
                                >
                                    Ver todos los movimientos →
                                </button>
                            </section>
                        </motion.div>
                    </motion.div>
                </>
            )}

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

            <AnimatePresence>
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
            </AnimatePresence>

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
        </>
    )
}
