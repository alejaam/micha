import { useCallback, useEffect, useState } from 'react'
import { getSettlement } from '../api'

export function useSettlement({
    isAuthenticated,
    householdId,
    handleProtectedError,
    onUnexpectedError,
    shouldIgnoreError,
}) {
    const [settlement, setSettlement] = useState(null)
    const [loadingSettlement, setLoadingSettlement] = useState(false)
    const [settlementYear, setSettlementYear] = useState(new Date().getUTCFullYear())
    const [settlementMonth, setSettlementMonth] = useState(new Date().getUTCMonth() + 1)

    const loadSettlement = useCallback(async () => {
        if (!isAuthenticated || !householdId.trim()) {
            setSettlement(null)
            return
        }

        setLoadingSettlement(true)
        try {
            const data = await getSettlement({
                householdId: householdId.trim(),
                year: settlementYear,
                month: settlementMonth,
            })
            setSettlement(data)
        } catch (err) {
            setSettlement(null)
            if (err?.code === 'UNAUTHORIZED') {
                handleProtectedError(err)
                return
            }

            if (!shouldIgnoreError(err)) {
                onUnexpectedError(err)
            }
        } finally {
            setLoadingSettlement(false)
        }
    }, [
        handleProtectedError,
        householdId,
        isAuthenticated,
        onUnexpectedError,
        settlementMonth,
        settlementYear,
        shouldIgnoreError,
    ])

    useEffect(() => {
        if (!isAuthenticated) {
            return
        }

        loadSettlement()
    }, [isAuthenticated, loadSettlement])

    const resetToCurrentMonth = useCallback(() => {
        const now = new Date()
        setSettlementYear(now.getUTCFullYear())
        setSettlementMonth(now.getUTCMonth() + 1)
    }, [])

    return {
        settlement,
        loadingSettlement,
        settlementYear,
        settlementMonth,
        setSettlement,
        setSettlementYear,
        setSettlementMonth,
        loadSettlement,
        resetToCurrentMonth,
    }
}
