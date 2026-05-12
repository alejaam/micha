import { AnimatePresence } from 'framer-motion'
import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BottomSheet } from '../components/BottomSheet'
import { DynamicChartsPanel } from '../components/DynamicChartsPanel'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { ExpenseSummary } from '../components/ExpenseSummary'
import { FAB } from '../components/FAB'
import { MembersPanel } from '../components/MembersPanel'
import { PeriodHistory } from '../components/PeriodHistory'
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
    } = useAppShell()

    const {
        members,
        loadingMembers,
        cards,
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

    return (
        <>
            {error && <Banner type="error" floating onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" floating onDismiss={() => setMessage('')}>{message}</Banner>}

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
                    {/* ─── (b) RemainingSalary + ExpenseSummary row ─── */}
                    <div className="u-flex u-flex-wrap u-gap-4" aria-label="Resumen financiero">
                        {currentMember && (
                            <div className="u-flex-1" style={{ minWidth: 280 }}>
                                <RemainingSalaryPanel
                                    householdId={householdId}
                                    memberId={currentMember.id}
                                    period={currentPeriod}
                                    currency={activeCurrency}
                                />
                            </div>
                        )}
                        <div className="u-flex-1" style={{ minWidth: 280 }}>
                            <section className="card dashboardSummaryCard" aria-label="Resumen del mes">
                                <h2 className="sectionTitle">
                                    <span className="sectionTitleIcon" aria-hidden>📊</span>
                                    Este mes
                                </h2>
                                <ExpenseSummary settlement={settlement} currency={activeCurrency} />
                            </section>
                        </div>
                    </div>

                    {/* ─── (c) RecentExpenses ─── */}
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

                    {/* ─── (d) DynamicChartsPanel ─── */}
                    <DynamicChartsPanel
                        categoryTotals={categoryTotals}
                        memberActualVsExpected={memberActualVsExpected}
                        msiProgress={msiProgress}
                        spendingTrend={spendingTrend}
                        currency={activeCurrency}
                    />

                    {/* ─── (e) MembersPanel ─── */}
                    <MembersPanel
                        members={members}
                        currency={activeCurrency}
                    />

                    {/* ─── (f) PeriodHistory ─── */}
                    <PeriodHistory householdId={householdId} />

                    <button
                        type="button"
                        className="btn btnPrimary"
                        onClick={() => navigate('/balances')}
                        style={{ marginTop: '1rem' }}
                    >
                        Ver Balances →
                    </button>
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
                    cards={cards}
                    defaultPaidByMemberId={currentMember?.id ?? ''}
                />
            </BottomSheet>
        </>
    )
}
