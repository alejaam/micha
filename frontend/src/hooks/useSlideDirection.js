import { useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'

/**
 * useSlideDirection — compares current pathname to previous to determine
 * forward / backward slide direction for AnimatePresence transitions.
 *
 * @param {string[]} steps — ordered route paths e.g. ['/onboarding/household', '/onboarding/cards']
 * @returns {'forward' | 'backward'}
 */
export function useSlideDirection(steps) {
    const { pathname } = useLocation()
    const prevPathname = useRef(pathname)

    useEffect(() => {
        prevPathname.current = pathname
    }, [pathname])

    const currentIndex = steps.indexOf(pathname)
    const prevIndex = steps.indexOf(prevPathname.current)

    if (prevIndex === -1 || currentIndex === -1) return 'forward'
    return currentIndex >= prevIndex ? 'forward' : 'backward'
}
