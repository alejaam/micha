import { useCallback, useEffect, useState } from 'react'
import { listExpenses } from '../api'

export function useExpenses({ isAuthenticated, householdId, handleProtectedError, onErrorClear }) {
    const [items, setItems] = useState([])
    const [loadingList, setLoadingList] = useState(false)

    const loadExpenses = useCallback(async () => {
        if (!isAuthenticated || !householdId.trim()) {
            setItems([])
            return
        }

        setLoadingList(true)
        onErrorClear()
        try {
            const data = await listExpenses({ householdId: householdId.trim(), limit: 50, offset: 0 })
            setItems(Array.isArray(data) ? data : [])
        } catch (err) {
            handleProtectedError(err)
        } finally {
            setLoadingList(false)
        }
    }, [handleProtectedError, householdId, isAuthenticated, onErrorClear])

    useEffect(() => {
        if (!isAuthenticated) {
            return
        }

        loadExpenses()
    }, [isAuthenticated, loadExpenses])

    return {
        items,
        loadingList,
        setItems,
        loadExpenses,
    }
}
