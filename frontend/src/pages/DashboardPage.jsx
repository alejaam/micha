import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import {
    createExpense,
    deleteExpense,
    listRecurringExpenses,
    patchExpense,
} from '../api'
import { BottomSheet } from '../components/BottomSheet'
import { CardExpensesPanel } from '../components/CardExpensesPanel'
import { DynamicChartsPanel } from '../components/DynamicChartsPanel'
import { HistorySection } from '../components/HistorySection'
import { ExpenseList } from '../components/ExpenseList'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { ExpenseSummary } from '../components/ExpenseSummary'
import { FAB } from '../components/FAB'
import { FixedExpensesPanel } from '../components/FixedExpensesPanel'
import { IncomesPanel } from '../components/IncomesPanel'
import { MembersPanel } from '../components/MembersPanel'
import { RecentExpenses } from '../components/RecentExpenses'
import { SettlementPanel } from '../components/SettlementPanel'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useCurrentMember } from '../hooks/useCurrentMember'
import { useDashboardDerivedData } from '../hooks/useDashboardDerivedData'
import { useExpenses } from '../hooks/useExpenses'
import { useHistoricalPeriods } from '../hooks/useHistoricalPeriods'
import { useMembers } from '../hooks/useMembers'
import { useSettlement } from '../hooks/useSettlement'
import { EmptyState } from '../ui/EmptyState'
import { Banner } from '../ui/Banner'

function isExpectedSettlementOnboardingError(err) {
    return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

export function DashboardPage() {
    const { isAuthenticated, handleProtectedError } = useAuth()
    const {
        householdId,
        selectedHousehold,
        setPeriodStatus,
        isMutationLocked,
    } = useAppShell()
    const navigate = useNavigate()

    const [message, setMessage] = useState('')
    const [error, setError] = useState('')
    const [submittingCreate, setSubmittingCreate] = useState(false)
    const [savingId, setSavingId] = useState('')
    const [deletingId, setDeletingId] = useState('')
    const [modalOpen, setModalOpen] = useState(false)
    const [quickAddOpen, setQuickAddOpen] = useState(false)
    const [recurringItems, setRecurringItems] = useState([])

    const onErrorClear = useCallback(() => setError(''), [])
    const onUnexpectedError = useCallback((err) => setError(err.message), [])

    const { members, loadingMembers } = useMembers({
        isAuthenticated,
        householdId,
        handleProtectedError,
    })

    const { items, loadingList, loadExpenses } = useExpenses({
        isAuthenticated,
        householdId,
        handleProtectedError,
        onErrorClear,
    })

    const {
        settlement,
        loadingSettlement,
        settlementYear,
        settlementMonth,
        setSettlementYear,
        setSettlementMonth,
        loadSettlement,
        resetToCurrentMonth,
    } = useSettlement({
        isAuthenticated,
        householdId,
        handleProtectedError,
        onUnexpectedError,
        shouldIgnoreError: isExpectedSettlementOnboardingError,
    })

    const currentMember = useCurrentMember(members)

    const memberIndex = useMemo(
        () => Object.fromEntries(members.map((m) => [m.id, m.name])),
        [members],
    )

    const activeCurrency = selectedHousehold?.currency || 'MXN'
    const hasHouseholds = !!householdId

    const derivedPeriodStatus = useMemo(() => {
        if (settlement?.is_closed === true) {
            return 'closed'
        }

        const now = new Date()
        const currentYear = now.getUTCFullYear()
        const currentMonth = now.getUTCMonth() + 1
        const isCurrentPeriod = settlementYear === currentYear && settlementMonth === currentMonth

        return isCurrentPeriod ? 'open' : 'review'
    }, [settlement, settlementYear, settlementMonth])

    useEffect(() => {
        setPeriodStatus(derivedPeriodStatus)
    }, [derivedPeriodStatus, setPeriodStatus])

    useEffect(() => {
        if (isMutationLocked && modalOpen) {
            setModalOpen(false)
        }
        if (isMutationLocked && quickAddOpen) {
            setQuickAddOpen(false)
        }
    }, [isMutationLocked, modalOpen, quickAddOpen])

    const {
        fixedTotalCents,
        categoryTotals,
        memberActualVsExpected,
        msiProgress,
        spendingTrend,
    } = useDashboardDerivedData({
        expenses: items,
        members,
        settlement,
        recurringItems,
    })

    useEffect(() => {
        let cancelled = false
        async function loadRecurring() {
            if (!isAuthenticated || !householdId.trim()) {
                setRecurringItems([])
                return
            }

            try {
                const data = await listRecurringExpenses({ householdId: householdId.trim(), limit: 200, offset: 0 })
                if (!cancelled) {
                    setRecurringItems(Array.isArray(data) ? data : [])
                }
            } catch (err) {
                if (!cancelled) {
                    handleProtectedError(err)
                }
            }
        }
        loadRecurring()
        return () => { cancelled = true }
    }, [isAuthenticated, householdId, handleProtectedError])

    const {
        closedPeriods,
        selectedPeriodKey,
        setSelectedPeriodKey,
        comparisonSeries,
        memberBalanceTrend,
        completedMsi,
        selectedPeriodSnapshot,
        isLoading: loadingHistory,
        isProvisional,
        provisionalReason,
    } = useHistoricalPeriods({
        householdId,
        expenses: items,
        members,
    })

    const handleOpenQuickAdd = useCallback(() => {
        if (isMutationLocked) {
            setError('Period is under review or closed. Mutating actions are disabled.')
            return
        }
        setQuickAddOpen(true)
    }, [isMutationLocked])

    // Redirect to onboarding if needed
    if (!hasHouseholds) {
        return (
            <div className="dashboardOnboarding">
                <div className="onboardingCard card">
                    <div className="onboardingHeader">
                        <p className="authEyebrow">Getting started</p>
                        <h2 className="authTitle">Set up your household</h2>
                        <p className="authMeta">
                            You need a household before you can track expenses.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary btnFull"
                        onClick={() => navigate('/onboarding/household')}
                    >
                        Create household →
                    </button>
                </div>
            </div>
        )
    }

// No secondary onboarding needed because the household creator is auto-added

    async function handleCreate({ amountCents, description, paidByMemberId, isShared, paymentMethod, expenseType, cardId, cardName, category, totalInstallments }) {
        if (isMutationLocked) {
            setError('Period is under review or closed. Mutating actions are disabled.')
            return
        }

        setMessage('')
        setError('')
        setSubmittingCreate(true)
        try {
            await createExpense({
                householdId: householdId.trim(),
                paidByMemberId,
                amountCents,
                description,
                isShared,
                currency: activeCurrency,
                paymentMethod,
                expenseType,
                cardId,
                cardName,
                category,
                totalInstallments,
            })
            setMessage('Expense added.')
            setModalOpen(false)
            await loadExpenses()
            await loadSettlement()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setSubmittingCreate(false)
        }
    }

    async function handleSave({ id, amountCents, description }) {
        if (isMutationLocked) {
            setError('Period is under review or closed. Mutating actions are disabled.')
            return
        }

        setMessage('')
        setError('')
        setSavingId(id)
        try {
            await patchExpense({ id, amountCents, description })
            setMessage('Expense updated.')
            await loadExpenses()
            await loadSettlement()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setSavingId('')
        }
    }

    async function handleDelete(id) {
        if (isMutationLocked) {
            setError('Period is under review or closed. Mutating actions are disabled.')
            return
        }

        setMessage('')
        setError('')
        setDeletingId(id)
        try {
            await deleteExpense(id)
            setMessage('Expense deleted.')
            await loadExpenses()
            await loadSettlement()
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
        } finally {
            setDeletingId('')
        }
    }

    const hasRecurringFixed = recurringItems.some((item) => item.expense_type === 'fixed')
    const hasExpenses = items.length > 0 || hasRecurringFixed

    return (
        <>
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            {!hasExpenses && !loadingList ? (
                /* Empty state when no expenses */
                <>
                    <section className="card dashboardEmptyState" aria-label="No expenses yet">
                        <EmptyState
                            title="No expenses yet"
                            description="Use quick add to capture your first expense and unlock dashboards."
                            ctaLabel="Quick add"
                            onCta={handleOpenQuickAdd}
                            icon="[+]"
                        />
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
                </>
            ) : (
                <motion.div 
                    className="pageGrid"
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.25, ease: "easeOut", staggerChildren: 0.1 }}
                >
                    {/* LEFT COLUMN (Desktop) / TOP (Mobile) */}
                    <motion.div 
                        className="dashboardCol"
                        initial={{ opacity: 0, y: 8 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.2, ease: "easeOut" }}
                    >
                        {/* Settlement */}
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

                        {/* Summary strip */}
                        <section className="card dashboardSummaryCard" aria-label="This month">
                            <h2 className="sectionTitle">
                                <span className="sectionTitleIcon" aria-hidden>📊</span>
                                This month
                            </h2>
                            <ExpenseSummary settlement={settlement} currency={activeCurrency} />
                        </section>

                        {/* Members overview */}
                        <MembersPanel
                            members={members}
                            currency={activeCurrency}
                        />

                        {/* Incomes */}
                        <IncomesPanel
                            members={members}
                            settlement={settlement}
                            currency={activeCurrency}
                        />

                        <section className="card" aria-label="Cards quick actions">
                            <div className="listHeader">
                                <h2 className="listTitle">Cards</h2>
                            </div>
                            <p className="text-sm text-dim mb-3">
                                Add a new card or change your preferred card for new expenses.
                            </p>
                            <div className="flex gap-3">
                                <button
                                    type="button"
                                    className="btn"
                                    onClick={() => navigate('/onboarding/cards')}
                                    disabled={isMutationLocked}
                                >
                                    Manage cards
                                </button>
                                <button
                                    type="button"
                                    className="btn"
                                    onClick={() => navigate('/onboarding/fixed-expenses')}
                                    disabled={isMutationLocked}
                                >
                                    Manage fixed expenses
                                </button>
                            </div>
                        </section>
                    </motion.div>

                    {/* RIGHT COLUMN (Desktop) / BOTTOM (Mobile) */}
                    <motion.div 
                        className="dashboardCol"
                        initial={{ opacity: 0, y: 8 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.2, ease: "easeOut", delay: 0.05 }}
                    >
                        {/* Fixed expenses breakdown */}
                        <DynamicChartsPanel
                            categoryTotals={categoryTotals}
                            memberActualVsExpected={memberActualVsExpected}
                            msiProgress={msiProgress}
                            spendingTrend={spendingTrend}
                            currency={activeCurrency}
                        />

                        {/* Fixed expenses breakdown */}
                        <FixedExpensesPanel
                            items={items}
                            recurringItems={recurringItems}
                            members={members}
                            currency={activeCurrency}
                        />

                        {/* Card expenses breakdown */}
                        <CardExpensesPanel
                            items={items}
                            members={members}
                            currency={activeCurrency}
                        />

                        {/* Recent expenses */}
                        <section className="card" aria-label="Recent expenses">
                            <div className="listHeader">
                                <h2 className="listTitle">Recent expenses</h2>
                                {items.length > 0 && (
                                    <span className="listCount">{items.length} total</span>
                                )}
                            </div>
                            <RecentExpenses
                                items={items}
                                isLoading={loadingList}
                                currency={activeCurrency}
                                limit={8}
                                onQuickAdd={handleOpenQuickAdd}
                            />
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

                        {/* Full expense list with edit/delete */}
                        {items.length > 0 && (
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
                        )}
                    </motion.div>
                </motion.div>
            )}

            {/* FAB + Modal */}
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
                    onSubmit={handleCreate}
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
                        await handleCreate(payload)
                        setQuickAddOpen(false)
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
