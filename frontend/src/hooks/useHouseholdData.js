import { useCallback, useEffect, useMemo, useState } from 'react'
import {
    createExpense,
    deleteExpense,
    listRecurringExpenses,
    patchExpense,
} from '../api'
import { useAppShell } from '../context/AppShellContext'
import { useAuth } from '../context/AuthContext'
import { useCurrentMember } from './useCurrentMember'
import { useDashboardDerivedData } from './useDashboardDerivedData'
import { useExpenses } from './useExpenses'
import { useHistoricalPeriods } from './useHistoricalPeriods'
import { useMembers } from './useMembers'
import { useSettlement } from './useSettlement'

function isExpectedSettlementOnboardingError(err) {
    return err?.code === 'NO_MEMBERS' || String(err?.message || '').toLowerCase().includes('at least one member')
}

export function useHouseholdData() {
    const { isAuthenticated, handleProtectedError } = useAuth()
    const {
        householdId,
        selectedHousehold,
        setPeriodStatus,
        isMutationLocked,
    } = useAppShell()

    const [message, setMessage] = useState('')
    const [error, setError] = useState('')
    const [submittingCreate, setSubmittingCreate] = useState(false)
    const [savingId, setSavingId] = useState('')
    const [deletingId, setDeletingId] = useState('')
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

    async function handleCreate(payload) {
        if (isMutationLocked) {
            setError('Period is under review or closed. Mutating actions are disabled.')
            return
        }

        setMessage('')
        setError('')
        setSubmittingCreate(true)
        try {
            await createExpense({
                ...payload,
                householdId: householdId.trim(),
                currency: activeCurrency,
            })
            setMessage('Expense added.')
            await loadExpenses()
            await loadSettlement()
            return true
        } catch (err) {
            if (!handleProtectedError(err)) setError(err.message)
            return false
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

    return {
        // Data
        members,
        loadingMembers,
        items,
        loadingList,
        recurringItems,
        settlement,
        loadingSettlement,
        settlementYear,
        settlementMonth,
        currentMember,
        memberIndex,
        activeCurrency,
        householdId,
        selectedHousehold,
        isMutationLocked,
        
        // Derived Data
        fixedTotalCents,
        categoryTotals,
        memberActualVsExpected,
        msiProgress,
        spendingTrend,
        
        // History
        closedPeriods,
        selectedPeriodKey,
        comparisonSeries,
        memberBalanceTrend,
        completedMsi,
        selectedPeriodSnapshot,
        loadingHistory,
        isProvisional,
        provisionalReason,

        // Actions
        loadExpenses,
        loadSettlement,
        setSettlementYear,
        setSettlementMonth,
        resetToCurrentMonth,
        setSelectedPeriodKey,
        handleCreate,
        handleSave,
        handleDelete,

        // UI State
        message,
        setMessage,
        error,
        setError,
        submittingCreate,
        savingId,
        deletingId,
    }
}
