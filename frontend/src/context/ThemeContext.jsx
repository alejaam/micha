import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

const THEME_STORAGE_KEY = 'micha-theme'

const ThemeContext = createContext(null)

function getSystemPrefersDark() {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
        return false
    }
    return window.matchMedia('(prefers-color-scheme: dark)').matches
}

export function ThemeProvider({ children }) {
    const [theme, setThemeState] = useState(() => {
        try {
            const saved = localStorage.getItem(THEME_STORAGE_KEY)
            if (saved === 'light' || saved === 'dark' || saved === 'auto') {
                return saved
            }
        } catch {
            // ignore localStorage errors (e.g. security sandboxing)
        }
        return 'auto'
    })

    const [systemPrefersDark, setSystemPrefersDark] = useState(getSystemPrefersDark)

    useEffect(() => {
        if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
            return
        }

        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
        setSystemPrefersDark(mediaQuery.matches)

        const handleChange = (e) => {
            setSystemPrefersDark(e.matches)
        }

        if (typeof mediaQuery.addEventListener === 'function') {
            mediaQuery.addEventListener('change', handleChange)
            return () => mediaQuery.removeEventListener('change', handleChange)
        } else if (typeof mediaQuery.addListener === 'function') {
            mediaQuery.addListener(handleChange)
            return () => mediaQuery.removeListener(handleChange)
        }
    }, [])

    const resolvedTheme = useMemo(() => {
        if (theme === 'auto') {
            return systemPrefersDark ? 'dark' : 'light'
        }
        return theme
    }, [theme, systemPrefersDark])

    useEffect(() => {
        if (typeof document === 'undefined') return
        const root = document.documentElement
        root.setAttribute('data-theme', resolvedTheme)
        root.style.colorScheme = resolvedTheme
    }, [resolvedTheme])

    const setTheme = useCallback((newTheme) => {
        if (newTheme !== 'auto' && newTheme !== 'light' && newTheme !== 'dark') return
        setThemeState(newTheme)
        try {
            localStorage.setItem(THEME_STORAGE_KEY, newTheme)
        } catch {
            // ignore
        }
    }, [])

    const toggleTheme = useCallback(() => {
        setThemeState((current) => {
            let next
            if (current === 'auto') {
                next = 'dark'
            } else if (current === 'dark') {
                next = 'light'
            } else {
                next = 'auto'
            }
            try {
                localStorage.setItem(THEME_STORAGE_KEY, next)
            } catch {
                // ignore
            }
            return next
        })
    }, [])

    const value = useMemo(() => ({
        theme,
        resolvedTheme,
        isDark: resolvedTheme === 'dark',
        setTheme,
        toggleTheme,
    }), [theme, resolvedTheme, setTheme, toggleTheme])

    return (
        <ThemeContext.Provider value={value}>
            {children}
        </ThemeContext.Provider>
    )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useTheme() {
    const context = useContext(ThemeContext)
    if (!context) {
        return {
            theme: 'auto',
            resolvedTheme: 'light',
            isDark: false,
            setTheme: () => {},
            toggleTheme: () => {},
        }
    }
    return context
}
