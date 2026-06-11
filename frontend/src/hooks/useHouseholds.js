import { useCallback, useEffect, useState, useRef } from 'react'
import { listHouseholds } from '../api'
import { useFocusRefetch } from './useFocusRefetch'

export function useHouseholds({ isAuthenticated, handleProtectedError }) {
    const [householdId, setHouseholdId] = useState('')
    const [households, setHouseholds] = useState([])
    const [loadingHouseholds, setLoadingHouseholds] = useState(true)

    // Track whether this is the very first load (vs. a background refetch)
    const isFirstLoad = useRef(true)

    const householdIdRef = useRef(householdId)
    useEffect(() => {
        householdIdRef.current = householdId
    }, [householdId])

    const loadHouseholds = useCallback(async (opts = {}) => {
        if (!isAuthenticated) {
            return
        }

        const { silent = false } = opts
        if (!silent) {
            setLoadingHouseholds(true)
        }

        try {
            const data = await listHouseholds({ limit: 100, offset: 0 })
            const next = Array.isArray(data) ? data : []
            setHouseholds(next)

            if (next.length === 0) {
                setHouseholdId('')
            } else {
                const selectedExists = next.some((household) => household.id === householdIdRef.current)
                if (!selectedExists) {
                    setHouseholdId(next[0].id)
                }
            }
        } catch (err) {
            handleProtectedError(err)
        } finally {
            if (!silent) {
                setLoadingHouseholds(false)
            }
            isFirstLoad.current = false
        }
    }, [handleProtectedError, isAuthenticated])

    useEffect(() => {
        if (!isAuthenticated) {
            return
        }

        loadHouseholds()
    }, [isAuthenticated, loadHouseholds])

    // Refetch on focus — but do it silently so we don't unmount the onboarding layout
    useFocusRefetch(() => loadHouseholds({ silent: true }))

    return {
        householdId,
        households,
        loadingHouseholds,
        setHouseholdId,
        setHouseholds,
        loadHouseholds,
    }
}
