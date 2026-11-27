import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

const THEME_STORAGE_KEY = 'micha-theme'

const DARK_MEDIA_QUERY = '(prefers-color-scheme: dark)'

const VALID_THEMES = ['auto', 'light', 'dark']

const ThemeContext = createContext(null)

function readStoredTheme() {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    return VALID_THEMES.includes(stored) ? stored : 'auto'
}

/**
 * ThemeProvider — theme preference ('auto' | 'light' | 'dark') with active
 * detection of the system preference (`prefers-color-scheme: dark`) and
 * persistence in localStorage. The resolved theme is exposed to the CSS
 * token layer through `<html data-theme="light|dark">`.
 */
export function ThemeProvider({ children }) {
    const [theme, setThemeState] = useState(readStoredTheme)
    const [systemPrefersDark, setSystemPrefersDark] = useState(
        () => window.matchMedia(DARK_MEDIA_QUERY).matches,
    )

    // Actively follow system preference changes (used while theme is 'auto')
    useEffect(() => {
        const darkQuery = window.matchMedia(DARK_MEDIA_QUERY)

        function handleChange(event) {
            setSystemPrefersDark(event.matches)
        }

        darkQuery.addEventListener('change', handleChange)
        return () => darkQuery.removeEventListener('change', handleChange)
    }, [])

    const resolvedTheme = theme === 'auto' ? (systemPrefersDark ? 'dark' : 'light') : theme

    // Reflect the resolved theme on <html> so the CSS tokens switch
    useEffect(() => {
        document.documentElement.dataset.theme = resolvedTheme
    }, [resolvedTheme])

    const setTheme = useCallback((next) => {
        setThemeState(next)
        localStorage.setItem(THEME_STORAGE_KEY, next)
    }, [])

    const value = useMemo(
        () => ({ theme, resolvedTheme, setTheme }),
        [theme, resolvedTheme, setTheme],
    )

    return (
        <ThemeContext.Provider value={value}>
            {children}
        </ThemeContext.Provider>
    )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useTheme() {
    const ctx = useContext(ThemeContext)
    if (!ctx) throw new Error('useTheme must be used inside <ThemeProvider>')
    return ctx
}
