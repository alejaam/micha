import { useCallback, useEffect, useState } from 'react'
import { listMembers } from '../api'

export function useMembers({ isAuthenticated, householdId, handleProtectedError }) {
    const [members, setMembers] = useState([])
    const [loadingMembers, setLoadingMembers] = useState(false)

    const loadMembers = useCallback(async () => {
        if (!isAuthenticated || !householdId.trim()) {
            setMembers([])
            return
        }

        setLoadingMembers(true)
        try {
            const data = await listMembers({ householdId: householdId.trim(), limit: 100, offset: 0 })
            setMembers(Array.isArray(data) ? data : [])
        } catch (err) {
            setMembers([])
            handleProtectedError(err)
        } finally {
            setLoadingMembers(false)
        }
    }, [handleProtectedError, householdId, isAuthenticated])

    useEffect(() => {
        if (!isAuthenticated) {
            return
        }

        loadMembers()
    }, [isAuthenticated, loadMembers])

    return {
        members,
        loadingMembers,
        setMembers,
        loadMembers,
    }
}
