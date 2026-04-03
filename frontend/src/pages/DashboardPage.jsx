import { Check, Compass, Settings } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
    createExpense,
    deleteExpense,
    generateRecurringExpenses,
    patchExpense,
} from '../api'
import { CardExpensesPanel } from '../components/CardExpensesPanel'
import { ExpenseList } from '../components/ExpenseList'
import { ExpenseModal } from '../components/ExpenseModal'
import { ExpenseSummary } from '../components/ExpenseSummary'
import { FAB } from '../components/FAB'
import { FixedExpensesPanel } from '../components/FixedExpensesPanel'
import { MonthlyInsightsCard } from '../components/MonthlyInsightsCard'
import { RecentExpenses } from '../components/RecentExpenses'
import { SettlementPanel } from '../components/SettlementPanel'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useCurrentMember } from '../hooks/useCurrentMember'
import { useExpenses } from '../hooks/useExpenses'
import { useMembers } from '../hooks/useMembers'
import { useSettlement } from '../hooks/useSettlement'
import { Banner } from '../ui/Banner'
import { SkeletonCard, SkeletonExpenseList } from '../ui/Skeleton'
import {
    getRecurringAutomationLastPeriod,
    isRecurringAutomationEnabled,
    setRecurringAutomationLastPeriod,
} from '../utils'

function isExpectedSettlementOnboardingError(err) {
    return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

function buildCreateExpenseErrorMessage(err, { paidByMemberId, currentMemberId, isCurrentUserAdmin, memberIndex }) {
    if (err?.code !== 'FORBIDDEN') {
        return err?.message || 'Unable to add expense right now.'
    }

    const selectedMemberName = memberIndex[paidByMemberId] || 'that member'
    if (currentMemberId && paidByMemberId !== currentMemberId && !isCurrentUserAdmin) {
        return `Only the household admin can register expenses for ${selectedMemberName}. Select yourself in "Paid by" or ask the admin to do it.`
    }

    return 'You are not allowed to register this expense. Ensure your member is linked to your account or ask the household admin for support.'
}

export function DashboardPage() {
    const { isAuthenticated, handleProtectedError } = useAuth()
    const {
        householdId,
        selectedHousehold,
        dashboardSection,
        setDashboardSection,
    } = useAppShell()
    const navigate = useNavigate()

    const [message, setMessage] = useState('')
    const [error, setError] = useState('')
    const [submittingCreate, setSubmittingCreate] = useState(false)
    const [savingId, setSavingId] = useState('')
    const [deletingId, setDeletingId] = useState('')
    const [modalOpen, setModalOpen] = useState(false)
    const [autoRecurringEnabled, setAutoRecurringEnabled] = useState(true)
    const autoRecurringRequestRef = useRef('')

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
    const linkedMembers = useMemo(
        () => members.filter((member) => member.user_id && String(member.user_id).trim() !== ''),
        [members],
    )
    const adminMember = linkedMembers[0] ?? null
    const isCurrentUserAdmin = !!currentMember?.id && currentMember.id === adminMember?.id
    const hasPendingMembers = members.some((member) => !member.user_id || String(member.user_id).trim() === '')

    const memberIndex = useMemo(
        () => Object.fromEntries(members.map((m) => [m.id, m.name])),
        [members],
    )

    const activeCurrency = selectedHousehold?.currency || 'MXN'
    const hasHouseholds = !!householdId
    const selectedSplitMode = String(selectedHousehold?.settlement_mode || '').toLowerCase()
    const splitSchemeLabel = selectedSplitMode === 'proportional'
        ? 'Proportional split'
        : 'Equal split'
    const splitSchemeHint = selectedSplitMode === 'proportional'
        ? 'Division follows household income weights.'
        : 'Division is shared equally among members.'

    useEffect(() => {
        setAutoRecurringEnabled(isRecurringAutomationEnabled(householdId))
    }, [householdId])

    useEffect(() => {
        if (!householdId || !autoRecurringEnabled) return

        const periodKey = `${settlementYear}-${String(settlementMonth).padStart(2, '0')}`
        const lastGeneratedPeriod = getRecurringAutomationLastPeriod(householdId)
        if (lastGeneratedPeriod === periodKey) return

        const requestKey = `${householdId}:${periodKey}`
        if (autoRecurringRequestRef.current === requestKey) return
        autoRecurringRequestRef.current = requestKey

        let cancelled = false
        async function autoGenerateRecurringForPeriod() {
            try {
                const result = await generateRecurringExpenses({ householdId })
                if (cancelled) return

                setRecurringAutomationLastPeriod(householdId, periodKey)
                if ((result?.generated_count || 0) > 0) {
                    setMessage(`Auto-generated ${result.generated_count} fixed expense${result.generated_count > 1 ? 's' : ''} for this period.`)
                    await loadExpenses()
                    await loadSettlement()
                }
            } catch (err) {
                if (!cancelled && !handleProtectedError(err)) {
                    setError(err.message || 'No se pudieron generar gastos recurrentes automaticamente.')
                }
            } finally {
                if (!cancelled) {
                    autoRecurringRequestRef.current = ''
                }
            }
        }

        autoGenerateRecurringForPeriod()
        return () => {
            cancelled = true
        }
    }, [
        autoRecurringEnabled,
        handleProtectedError,
        householdId,
        loadExpenses,
        loadSettlement,
        settlementMonth,
        settlementYear,
    ])

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
            if (!handleProtectedError(err)) {
                setError(buildCreateExpenseErrorMessage(err, {
                    paidByMemberId,
                    currentMemberId: currentMember?.id ?? '',
                    isCurrentUserAdmin,
                    memberIndex,
                }))
            }
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

    const dashboardSections = [
        { id: 'overview', label: 'Resumen', hint: 'Lo importante del mes', Icon: Check },
        { id: 'planning', label: 'Planeacion', hint: 'Pagos fijos y tarjetas', Icon: Settings },
        { id: 'activity', label: 'Actividad', hint: 'Movimientos recientes', Icon: Compass },
    ]

    return (
        <>
            {error && <Banner type="error" onDismiss={() => setError('')}>{error}</Banner>}
            {message && <Banner type="ok" onDismiss={() => setMessage('')}>{message}</Banner>}

            <section className="card mobileAddExpenseBar" aria-label="Accion rapida de gasto">
                <button
                    type="button"
                    className="btn btnPrimary btnFull"
                    onClick={() => setModalOpen(true)}
                >
                    + Agregar gasto
                </button>
            </section>

            <section className="card" aria-label="Esquema de division actual">
                <div className="listHeader">
                    <h2 className="sectionTitle">
                        <span className="sectionTitleIcon" aria-hidden>
                            <Settings className="sectionIconSvg" size={14} strokeWidth={2} />
                        </span>
                        Esquema de division
                    </h2>
                    <span className="sectionBadge">{splitSchemeLabel}</span>
                </div>
                <p className="formHint">{splitSchemeHint}</p>
                <button
                    type="button"
                    className="btn mt-4"
                    onClick={() => navigate('/household/settings')}
                >
                    Editar en Ajustes
                </button>
            </section>

            <section className="card" aria-label="Gestion de miembros del hogar">
                <div className="listHeader">
                    <h2 className="sectionTitle">
                        <span className="sectionTitleIcon" aria-hidden>
                            <Compass className="sectionIconSvg" size={14} strokeWidth={2} />
                        </span>
                        Miembros
                    </h2>
                    <span className="sectionBadge">Owner flow</span>
                </div>
                {isCurrentUserAdmin ? (
                    <>
                        <p className="formHint">
                            Invita nuevos integrantes para que participen en el registro y liquidacion mensual.
                        </p>
                        <button
                            type="button"
                            className="btn mt-4"
                            onClick={() => navigate('/members/new')}
                        >
                            Invitar miembro
                        </button>
                    </>
                ) : (
                    <p className="formHint">
                        Solo la persona owner/admin del hogar puede invitar nuevos miembros.
                    </p>
                )}
            </section>

            <section className="card dashboardSectionSwitcher" aria-label="Secciones del dashboard">
                <div className="dashboardSectionSwitcherHead">
                    <h2 className="sectionTitle">
                        <span className="sectionTitleIcon" aria-hidden>
                            <Compass className="sectionIconSvg" size={14} strokeWidth={2} />
                        </span>
                        Secciones
                    </h2>
                    <p className="formHint">
                        Elige una vista y enfocate en una tarea a la vez.
                    </p>
                </div>
                <div className="dashboardSectionTabs" role="tablist" aria-label="Pestanias de seccion del dashboard">
                    {dashboardSections.map((section) => {
                        const isActive = dashboardSection === section.id
                        const TabIcon = section.Icon
                        return (
                            <button
                                key={section.id}
                                type="button"
                                role="tab"
                                id={`tab-${section.id}`}
                                aria-selected={isActive}
                                aria-controls={`panel-${section.id}`}
                                className={`dashboardSectionTab ${isActive ? 'dashboardSectionTabActive' : ''}`}
                                onClick={() => setDashboardSection(section.id)}
                            >
                                <span className="dashboardSectionTabHead">
                                    <TabIcon className="dashboardSectionTabIcon" size={14} strokeWidth={2} aria-hidden />
                                    <span className="dashboardSectionTabLabel">{section.label}</span>
                                </span>
                                <span className="dashboardSectionTabHint">{section.hint}</span>
                            </button>
                        )
                    })}
                </div>
            </section>

            {isCurrentUserAdmin && hasPendingMembers && (
                <section className="card" aria-label="Permisos temporales de admin">
                    <p className="formHint formHintWarning">
                        Como admin del hogar, puedes registrar gastos temporalmente para integrantes pendientes.
                        Cuando activen su cuenta, ellos deben gestionar sus propios gastos.
                    </p>
                </section>
            )}

            {loadingList && items.length === 0 ? (
                /* Loading skeletons on first load */
                <>
                    <SkeletonCard lines={2} />
                    <SkeletonCard lines={3} />
                    <SkeletonExpenseList count={5} />
                </>
            ) : (
                <>
                    {dashboardSection === 'overview' && (
                        <section
                            role="tabpanel"
                            id="panel-overview"
                            aria-labelledby="tab-overview"
                            className="dashboardSectionPanel"
                        >
                            <section className="card dashboardSummaryCard" aria-label="Resumen de este mes">
                                <h2 className="sectionTitle">
                                    <span className="sectionTitleIcon" aria-hidden>
                                        <Check className="sectionIconSvg" size={14} strokeWidth={2} />
                                    </span>
                                    Este mes
                                </h2>
                                <ExpenseSummary settlement={settlement} currency={activeCurrency} />
                            </section>

                            <MonthlyInsightsCard
                                items={items}
                                currency={activeCurrency}
                                year={settlementYear}
                                month={settlementMonth}
                            />

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
                        </section>
                    )}

                    {dashboardSection === 'planning' && (
                        <section
                            role="tabpanel"
                            id="panel-planning"
                            aria-labelledby="tab-planning"
                            className="dashboardSectionPanel"
                        >
                            <section className="card" aria-label="Acciones rapidas de tarjetas">
                                <div className="listHeader">
                                    <h2 className="listTitle">Tarjetas</h2>
                                </div>
                                <p className="formHint">
                                    Agrega una tarjeta o cambia la tarjeta preferida para nuevos gastos.
                                </p>
                                <button
                                    type="button"
                                    className="btn"
                                    onClick={() => navigate('/onboarding/cards')}
                                >
                                    Gestionar tarjetas
                                </button>
                            </section>

                            <FixedExpensesPanel
                                items={items}
                                members={members}
                                currency={activeCurrency}
                            />

                            <CardExpensesPanel
                                items={items}
                                members={members}
                                currency={activeCurrency}
                            />
                        </section>
                    )}

                    {dashboardSection === 'activity' && (
                        <section
                            role="tabpanel"
                            id="panel-activity"
                            aria-labelledby="tab-activity"
                            className="dashboardSectionPanel"
                        >
                            {!hasExpenses && (
                                <section className="card dashboardEmptyState" aria-label="Sin gastos registrados">
                                    <div className="emptyStateIcon" aria-hidden>[]</div>
                                    <h2 className="emptyStateTitle">Sin gastos todavia</h2>
                                    <p className="emptyStateHint">
                                        Toca el boton <strong>+</strong> para registrar tu primer gasto.
                                    </p>
                                </section>
                            )}

                            <section className="card" aria-label="Gastos recientes">
                                <div className="listHeader">
                                    <h2 className="listTitle">Gastos recientes</h2>
                                    {items.length > 0 && (
                                        <span className="listCount">{items.length} totales</span>
                                    )}
                                </div>
                                <RecentExpenses
                                    items={items}
                                    isLoading={loadingList}
                                    currency={activeCurrency}
                                    limit={8}
                                />
                            </section>

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
                        </section>
                    )}
                </>
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
        </>
    )
}
