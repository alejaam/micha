import { useCallback, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
    createExpense,
    deleteExpense,
    patchExpense,
} from '../api'
import { CardExpensesPanel } from '../components/CardExpensesPanel'
import { ExpenseList } from '../components/ExpenseList'
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
import { useExpenses } from '../hooks/useExpenses'
import { useMembers } from '../hooks/useMembers'
import { useSettlement } from '../hooks/useSettlement'
import { Banner } from '../ui/Banner'

function isExpectedSettlementOnboardingError(err) {
    return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

export function DashboardPage() {
    const { isAuthenticated, handleProtectedError } = useAuth()
    const { householdId, selectedHousehold, loadHouseholds, setHouseholdId } = useAppShell()
    const navigate = useNavigate()

    const [message, setMessage] = useState('')
    const [error, setError] = useState('')
    const [submittingCreate, setSubmittingCreate] = useState(false)
    const [savingId, setSavingId] = useState('')
    const [deletingId, setDeletingId] = useState('')
    const [modalOpen, setModalOpen] = useState(false)

    const onErrorClear = useCallback(() => setError(''), [])
    const onUnexpectedError = useCallback((err) => setError(err.message), [])

    const { members, loadingMembers, loadMembers } = useMembers({
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
    const hasMembers = members.length > 0

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

    const hasExpenses = items.length > 0

    return (
        <div className="dashboardWrapper">
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            {/* HERO SECTION — IDENTITY & HERO METRIC */}
            <header className="dashboardHero" style={{ marginBottom: 'var(--sp-wide)', padding: '0 var(--sp-md)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: 'var(--sp-tight)' }}>
                    <div>
                        <span className="heroContext" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem', opacity: 0.6, textTransform: 'uppercase', letterSpacing: '0.1em' }}>
                            [ HOUSEHOLD : {selectedHousehold?.name?.toUpperCase()} ]
                        </span>
                        <h1 className="heroTitle" style={{ fontSize: '1.2rem', fontWeight: 'var(--fw-bold)', color: 'var(--color-text-1)', letterSpacing: '-0.02em', margin: '4px 0 0 0' }}>micha.dashboard</h1>
                    </div>
                    
                    <div className="heroMetadata" style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
                         <button 
                            className="btn" 
                            style={{ padding: '4px 8px', fontSize: '0.65rem', fontFamily: 'var(--font-mono)', background: 'transparent', border: '1px solid var(--color-border-soft)' }}
                            onClick={() => loadSettlement()}
                            disabled={loadingSettlement}
                        >
                            {loadingSettlement ? 'SYNCING...' : 'RELOAD_DATA'}
                        </button>
                        <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.65rem', opacity: 0.4 }}>LIVE_STATUS: OK</span>
                    </div>
                </div>
                
                <div className="heroMetric" style={{ borderTop: '2px solid var(--color-text-1)', paddingTop: '12px' }}>
                    <ExpenseSummary settlement={settlement} currency={activeCurrency} />
                </div>
            </header>

            {!hasExpenses && !loadingList ? (
                /* Empty state when no expenses */
                <section className="card dashboardEmptyState" aria-label="No expenses yet">
                    <div className="emptyStateIcon" aria-hidden style={{ fontFamily: 'var(--font-display)', fontSize: '3rem' }}>[ NO DATA ]</div>
                    <h2 className="emptyStateTitle" style={{ marginTop: '1rem' }}>Zero expenses logged</h2>
                    <p className="emptyStateHint" style={{ fontFamily: 'var(--font-mono)', textTransform: 'uppercase', fontSize: '0.8rem' }}>
                        Tap the + button below to add your first expense and start tracking.
                    </p>
                </section>
            ) : (
                <div className="dashboardGrid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 'var(--sp-medium)' }}>
                    {/* LEFT / STATUS COLUMN */}
                    <div className="dashboardCol dashboardColStatus" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-medium)' }}>
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
                        />

                         <section className="card" aria-label="Cards quick actions">
                            <div className="listHeader">
                                <h2 className="listTitle" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem', textTransform: 'uppercase' }}>Config: Cards</h2>
                            </div>
                            <p style={{ fontSize: '0.8rem', opacity: 0.6, marginBottom: '16px', fontFamily: 'var(--font-mono)' }}>
                                Managed payment instruments for this household.
                            </p>
                            <button
                                type="button"
                                className="btn"
                                style={{ width: '100%', fontFamily: 'var(--font-mono)', fontSize: '0.75rem', border: '1px solid var(--color-text-1)' }}
                                onClick={() => navigate('/onboarding/cards')}
                            >
                                [ MANAGE_CARDS ]
                            </button>
                        </section>
                    </div>

                    {/* RIGHT / ACTIVITY COLUMN */}
                    <div className="dashboardCol dashboardColActivity" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--sp-medium)' }}>
                        {/* Recent expenses */}
                        <section className="card" aria-label="Recent expenses" style={{ borderLeft: '1px solid var(--color-border)' }}>
                            <div className="listHeader" style={{ borderBottom: '1px solid var(--color-border-soft)', paddingBottom: '12px', marginBottom: '16px' }}>
                                <h2 className="listTitle" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem', textTransform: 'uppercase' }}>Feed: Recent_Activity</h2>
                                {items.length > 0 && (
                                    <span className="listCount" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.7rem' }}>{items.length}_ITEMS</span>
                                )}
                            </div>
                            <RecentExpenses
                                items={items}
                                isLoading={loadingList}
                                currency={activeCurrency}
                                limit={8}
                            />
                        </section>

                        {/* Full expense list with edit/delete */}
                        {items.length > 0 && (
                            <div style={{ opacity: 0.8 }}>
                                <ExpenseList
                                    items={items}
                                    isLoading={loadingList}
                                    deletingId={deletingId}
                                    savingId={savingId}
                                    onDelete={handleDelete}
                                    onSave={handleSave}
                                    currency={activeCurrency}
                                />
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* FAB + Modal */}
            <FAB onClick={() => setModalOpen(true)} />

            {modalOpen && (
                <ExpenseModal
                    onClose={() => setModalOpen(false)}
                    onSubmit={handleCreate}
                    isSubmitting={submittingCreate}
                    members={members}
                    isLoadingMembers={loadingMembers}
                    defaultPaidByMemberId={currentMember?.id ?? ''}
                    householdId={householdId}
                />
            )}
        </div>
    )
}
