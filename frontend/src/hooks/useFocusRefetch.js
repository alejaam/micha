import { useEffect } from 'react'

/**
 * useFocusRefetch — calls the provided load callback when the browser tab
 * regains focus (visibilitychange) or when the window receives focus.
 *
 * @param {Function} loadCallback — async function to reload data.
 */
export function useFocusRefetch(loadCallback) {
    useEffect(() => {
        if (!loadCallback) return

        let timeoutId
        const handleVisible = () => {
            if (document.visibilityState === 'visible') {
                clearTimeout(timeoutId)
                timeoutId = setTimeout(() => loadCallback(), 300)
            }
        }

        const handleFocus = () => {
            clearTimeout(timeoutId)
            timeoutId = setTimeout(() => loadCallback(), 300)
        }

        document.addEventListener('visibilitychange', handleVisible)
        window.addEventListener('focus', handleFocus)
        return () => {
            document.removeEventListener('visibilitychange', handleVisible)
            window.removeEventListener('focus', handleFocus)
            clearTimeout(timeoutId)
        }
    }, [loadCallback])
}
