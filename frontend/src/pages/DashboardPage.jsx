import { useCallback, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { useExpenses } from '../hooks/useExpenses'
import { useMembers } from '../hooks/useMembers'
import { useSettlement } from '../hooks/useSettlement'
import { useCurrentMember } from '../hooks/useCurrentMember'
import { ExpenseSummary } from '../components/ExpenseSummary'
import { RecentExpenses } from '../components/RecentExpenses'
import { ExpenseList } from '../components/ExpenseList'
import { SettlementPanel } from '../components/SettlementPanel'
import { IncomesPanel } from '../components/IncomesPanel'
import { FixedExpensesPanel } from '../components/FixedExpensesPanel'
import { CardExpensesPanel } from '../components/CardExpensesPanel'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { Banner } from '../ui/Banner'
import {
    createExpense,
    patchExpense,
    deleteExpense,
} from '../api'

function isExpectedSettlementOnboardingError(err) {
    return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

export function DashboardPage({ householdId, selectedHousehold, loadHouseholds, setHouseholdId }) {
    const { isAuthenticated, handleProtectedError } = useAuth()
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

    if (!hasMembers && !loadingMembers) {
        return (
            <div className="dashboardOnboarding">
                <div className="onboardingCard card">
                    <div className="onboardingHeader">
                        <p className="authEyebrow">Almost there</p>
                        <h2 className="authTitle">Add yourself as a member</h2>
                        <p className="authMeta">
                            Use the same email as your account so micha can auto-assign expenses to you.
                        </p>
                    </div>
                    <button
                        type="button"
                        className="btn btnPrimary btnFull"
                        onClick={() => navigate('/onboarding/member')}
                    >
                        Add member →
                    </button>
                </div>
            </div>
        )
    }

    async function handleCreate({ amountCents, description, paidByMemberId, isShared, paymentMethod, expenseType, cardName, category }) {
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
                cardName,
                category,
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

    return (
        <>
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            {/* Incomes */}
            <IncomesPanel
                members={members}
                settlement={settlement}
                currency={activeCurrency}
            />

            {/* Summary strip */}
            <section className="card dashboardSummaryCard" aria-label="This month">
                <h2 className="sectionTitle">
                    <span className="sectionTitleIcon" aria-hidden>📊</span>
                    This month
                </h2>
                <ExpenseSummary settlement={settlement} currency={activeCurrency} />
            </section>

            {/* Fixed expenses breakdown */}
            <FixedExpensesPanel
                items={items}
                members={members}
                currency={activeCurrency}
            />

            {/* Card expenses breakdown */}
            <CardExpensesPanel
                items={items}
                members={members}
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
                loadingSettlement={loadingSettlement}
                memberIndex={memberIndex}
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
                />
            </section>

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
                />
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
                />
            )}
        </>
    )
}
