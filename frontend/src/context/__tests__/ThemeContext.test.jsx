import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ThemeProvider, useTheme } from '../ThemeContext'

describe('ThemeContext', () => {
    let matchMediaListeners = []
    let matchesDark = false

    beforeEach(() => {
        localStorage.clear()
        document.documentElement.removeAttribute('data-theme')
        document.documentElement.style.colorScheme = ''
        matchMediaListeners = []
        matchesDark = false

        window.matchMedia = vi.fn().mockImplementation((query) => ({
            matches: query.includes('prefers-color-scheme: dark') ? matchesDark : false,
            media: query,
            onchange: null,
            addListener: vi.fn(), // legacy
            removeListener: vi.fn(), // legacy
            addEventListener: vi.fn((event, listener) => {
                if (event === 'change') {
                    matchMediaListeners.push(listener)
                }
            }),
            removeEventListener: vi.fn((event, listener) => {
                if (event === 'change') {
                    matchMediaListeners = matchMediaListeners.filter((l) => l !== listener)
                }
            }),
            dispatchEvent: vi.fn(),
        }))
    })

    function triggerSystemThemeChange(isDark) {
        matchesDark = isDark
        act(() => {
            matchMediaListeners.forEach((listener) => listener({ matches: isDark }))
        })
    }

    it('defaults to auto and resolves to light when system is light', () => {
        matchesDark = false
        const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

        expect(result.current.theme).toBe('auto')
        expect(result.current.resolvedTheme).toBe('light')
        expect(document.documentElement.getAttribute('data-theme')).toBe('light')
        expect(document.documentElement.style.colorScheme).toBe('light')
    })

    it('defaults to auto and resolves to dark when system is dark', () => {
        matchesDark = true
        const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

        expect(result.current.theme).toBe('auto')
        expect(result.current.resolvedTheme).toBe('dark')
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
        expect(document.documentElement.style.colorScheme).toBe('dark')
    })

    it('reacts dynamically to system changes in auto mode', () => {
        matchesDark = false
        const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

        expect(result.current.resolvedTheme).toBe('light')

        triggerSystemThemeChange(true)
        expect(result.current.resolvedTheme).toBe('dark')
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark')

        triggerSystemThemeChange(false)
        expect(result.current.resolvedTheme).toBe('light')
        expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })

    it('allows setting manual dark mode and persists to localStorage', () => {
        matchesDark = false
        const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

        act(() => {
            result.current.setTheme('dark')
        })

        expect(result.current.theme).toBe('dark')
        expect(result.current.resolvedTheme).toBe('dark')
        expect(localStorage.getItem('micha-theme')).toBe('dark')
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark')

        // System change should not affect manual override
        triggerSystemThemeChange(false)
        expect(result.current.resolvedTheme).toBe('dark')
    })

    it('allows setting manual light mode when system is dark', () => {
        matchesDark = true
        const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

        act(() => {
            result.current.setTheme('light')
        })

        expect(result.current.theme).toBe('light')
        expect(result.current.resolvedTheme).toBe('light')
        expect(localStorage.getItem('micha-theme')).toBe('light')
        expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })

    it('cycles theme with toggleTheme', () => {
        matchesDark = false
        const { result } = renderHook(() => useTheme(), { wrapper: ThemeProvider })

        expect(result.current.theme).toBe('auto')

        act(() => {
            result.current.toggleTheme()
        })
        expect(result.current.theme).toBe('dark')

        act(() => {
            result.current.toggleTheme()
        })
        expect(result.current.theme).toBe('light')

        act(() => {
            result.current.toggleTheme()
        })
        expect(result.current.theme).toBe('auto')
    })
})
