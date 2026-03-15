import { useEffect, useState } from 'react'
import { useAuth } from '../context/AuthContext'

/**
 * useCurrentMember — finds the member in `members` array whose user_id
 * matches the authenticated user's user_id.
 *
 * Returns null if not found or not yet loaded.
 */
export function useCurrentMember(members) {
    const { user } = useAuth()
    const [currentMember, setCurrentMember] = useState(null)

    useEffect(() => {
        if (!user?.user_id || !Array.isArray(members)) {
            setCurrentMember(null)
            return
        }

        const match = members.find((m) => m.user_id === user.user_id) ?? null
        setCurrentMember(match)
    }, [user, members])

    return currentMember
}
