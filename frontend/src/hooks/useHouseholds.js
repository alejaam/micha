import { useCallback, useEffect, useState } from 'react'
import { listHouseholds } from '../api'

export function useHouseholds({ isAuthenticated, handleProtectedError }) {
    const [householdId, setHouseholdId] = useState('')
    const [households, setHouseholds] = useState([])
    const [loadingHouseholds, setLoadingHouseholds] = useState(false)

    const loadHouseholds = useCallback(async () => {
        if (!isAuthenticated) {
            return
        }

        setLoadingHouseholds(true)
        try {
            const data = await listHouseholds({ limit: 100, offset: 0 })
            const next = Array.isArray(data) ? data : []
            setHouseholds(next)

            if (next.length === 0) {
                setHouseholdId('')
            } else {
                const selectedExists = next.some((household) => household.id === householdId)
                if (!selectedExists) {
                    setHouseholdId(next[0].id)
                }
            }
        } catch (err) {
            handleProtectedError(err)
        } finally {
            setLoadingHouseholds(false)
        }
    }, [handleProtectedError, householdId, isAuthenticated])

    useEffect(() => {
        if (!isAuthenticated) {
            return
        }

        loadHouseholds()
    }, [isAuthenticated, loadHouseholds])

    return {
        householdId,
        households,
        loadingHouseholds,
        setHouseholdId,
        setHouseholds,
        loadHouseholds,
    }
}
