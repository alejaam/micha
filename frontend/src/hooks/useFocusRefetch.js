import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useLocation } from 'react-router-dom'

/**
 * useFocusRefetch — invalidates React Query caches when the tab gets focus
 * (visibilitychange) or when the route changes.
 *
 * @param {Array<Array<string>>} [queryKeys] — specific query keys to invalidate.
 *   Defaults to common project keys if not provided.
 */
export function useFocusRefetch(queryKeys = []) {
    const queryClient = useQueryClient()
    const location = useLocation()

    useEffect(() => {
        const invalidate = () => {
            const keys = queryKeys.length > 0 ? queryKeys : [['members'], ['expenses'], ['settlement'], ['periods']]
            keys.forEach((key) => queryClient.invalidateQueries({ queryKey: key }))
        }

        const handleVisibility = () => {
            if (document.visibilityState === 'visible') {
                invalidate()
            }
        }

        document.addEventListener('visibilitychange', handleVisibility)
        return () => document.removeEventListener('visibilitychange', handleVisibility)
    }, [queryClient, queryKeys])

    // Also invalidate on route change
    useEffect(() => {
        queryClient.invalidateQueries()
    }, [location.pathname, queryClient])
}
