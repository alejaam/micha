import { useCallback, useEffect, useState } from 'react'
import { listCards } from '../api'

/**
 * Hook to manage credit cards for a household.
 */
export function useCards({ isAuthenticated, householdId, handleProtectedError }) {
    const [cards, setCards] = useState([])
    const [loadingCards, setLoadingCards] = useState(false)

    const loadCards = useCallback(async () => {
        if (!isAuthenticated || !householdId?.trim()) {
            setCards([])
            return
        }

        setLoadingCards(true)
        try {
            const data = await listCards({ householdId: householdId.trim() })
            setCards(Array.isArray(data) ? data : [])
        } catch (err) {
            if (handleProtectedError) {
                handleProtectedError(err)
            }
        } finally {
            setLoadingCards(false)
        }
    }, [isAuthenticated, householdId, handleProtectedError])

    useEffect(() => {
        loadCards()
    }, [loadCards])

    return {
        cards,
        loadingCards,
        loadCards,
    }
}
