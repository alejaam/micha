import { AnimatePresence } from 'framer-motion'
import { useCallback, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BottomSheet } from '../components/BottomSheet'
import { MSI } from '../components/CardExpensesPanel'
import { CategoriesGrid } from '../components/DynamicChartsPanel'
import { ExpenseForm } from '../components/ExpenseForm'
import { ExpenseModal } from '../components/ExpenseModal'
import { FAB } from '../components/FAB'
import { PeriodHistory } from '../components/PeriodHistory'
import { PeriodBar } from '../components/PeriodStatusRibbon'
import { Feed } from '../components/RecentExpenses'
import { useAppShell } from '../context/AppShellContext'
import { useHouseholdData } from '../hooks/useHouseholdData'
import { Banner } from '../ui/Banner'
import { EmptyState } from '../ui/EmptyState'

export function DashboardPage() {
    const navigate = useNavigate()
    const {
        currentPeriod,
        isLoadingPeriod,
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
        categoryTotals,
        msiProgress,
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
        setQuickAddOpen(true)
    }, [setQuickAddOpen])

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

    const hasRecurringFixed = recurringItems.some((item) => item.expense_type === 'fixed')
    const hasExpenses = items.length > 0 || hasRecurringFixed

    // Derive household balance from settlement total shared cents
    const periodBalance = settlement?.total_shared_cents ?? 0

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
                    {/* ─── 1. PeriodBar ─── */}
                    <PeriodBar
                        currentPeriod={currentPeriod}
                        balance={periodBalance}
                    />

                    {/* ─── 2. Feed ─── */}
                    <section className="card" aria-label="Gastos recientes">
                        <div className="listHeader">
                            <h2 className="listTitle">Gastos recientes</h2>
                            {items.length > 0 && (
                                <span className="listCount">{items.length} total</span>
                            )}
                        </div>
                        <Feed
                            items={items}
                            isLoading={loadingList}
                            currency={activeCurrency}
                            members={members}
                            limit={10}
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

                    {/* ─── 3. Gastos fijos shortcut ─── */}
                    {hasRecurringFixed && (
                        <section className="card" aria-label="Gastos fijos">
                            <div className="listHeader">
                                <h2 className="listTitle">Gastos fijos</h2>
                                {recurringItems.filter((i) => i.expense_type === 'fixed').length > 0 && (
                                    <span className="listCount">
                                        {recurringItems.filter((i) => i.expense_type === 'fixed').length} registrados
                                    </span>
                                )}
                            </div>
                            <p className="u-text-sm u-text-dim u-mb-3">
                                Gestiona las suscripciones, renta y demás gastos recurrentes.
                            </p>
                            <button
                                type="button"
                                className="btn btnPrimary"
                                onClick={() => navigate('/fixed-expenses')}
                                style={{ width: '100%' }}
                            >
                                Gestionar gastos fijos →
                            </button>
                        </section>
                    )}

                    {/* ─── 5. CategoriesGrid ─── */}
                    <CategoriesGrid
                        categoryTotals={categoryTotals}
                        currency={activeCurrency}
                    />

                    {/* ─── 6. MSI progress list ─── */}
                    <MSI
                        items={items}
                        msiProgress={msiProgress}
                        currency={activeCurrency}
                    />

                    {/* ─── 7. PeriodHistory ─── */}
                    <PeriodHistory householdId={householdId} />
                </>
            )}

            <FAB
                onClick={() => setModalOpen(true)}
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
