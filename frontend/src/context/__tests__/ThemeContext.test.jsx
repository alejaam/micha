import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ThemeProvider, useTheme } from '../ThemeContext'

/**
 * jsdom does not implement window.matchMedia — stub it with a controllable
 * system preference and manual 'change' dispatch.
 */
function stubMatchMedia(prefersDark) {
    let dark = prefersDark
    const listeners = new Set()
    window.matchMedia = (query) => ({
        matches: dark,
        media: query,
        addEventListener: (_type, cb) => listeners.add(cb),
        removeEventListener: (_type, cb) => listeners.delete(cb),
    })
    return {
        setSystemDark(next) {
            dark = next
            listeners.forEach((cb) => cb({ matches: dark }))
        },
    }
}

function renderTheme() {
    return renderHook(() => useTheme(), { wrapper: ThemeProvider })
}

describe('ThemeProvider', () => {
    afterEach(() => {
        localStorage.clear()
        document.documentElement.removeAttribute('data-theme')
    })

    it('defaults to auto and resolves to the system preference', () => {
        stubMatchMedia(true)

        const { result } = renderTheme()

        expect(result.current.theme).toBe('auto')
        expect(result.current.resolvedTheme).toBe('dark')
        expect(document.documentElement.dataset.theme).toBe('dark')
    })

    it('resolves to light when the system preference is light', () => {
        stubMatchMedia(false)

        const { result } = renderTheme()

        expect(result.current.resolvedTheme).toBe('light')
        expect(document.documentElement.dataset.theme).toBe('light')
    })

    it('follows system preference changes while on auto', () => {
        const media = stubMatchMedia(false)
        const { result } = renderTheme()

        act(() => media.setSystemDark(true))

        expect(result.current.resolvedTheme).toBe('dark')
        expect(document.documentElement.dataset.theme).toBe('dark')
    })

    it('persists an explicit choice and keeps it regardless of the system preference', () => {
        const media = stubMatchMedia(true)
        const { result } = renderTheme()

        act(() => result.current.setTheme('light'))

        expect(result.current.theme).toBe('light')
        expect(result.current.resolvedTheme).toBe('light')
        expect(localStorage.getItem('micha-theme')).toBe('light')
        expect(document.documentElement.dataset.theme).toBe('light')

        act(() => media.setSystemDark(false))

        expect(result.current.resolvedTheme).toBe('light')
    })

    it('restores a stored explicit theme on mount', () => {
        stubMatchMedia(false)
        localStorage.setItem('micha-theme', 'dark')

        const { result } = renderTheme()

        expect(result.current.theme).toBe('dark')
        expect(result.current.resolvedTheme).toBe('dark')
        expect(document.documentElement.dataset.theme).toBe('dark')
    })

    it('falls back to auto when the stored value is unknown', () => {
        stubMatchMedia(false)
        localStorage.setItem('micha-theme', 'sepia')

        const { result } = renderTheme()

        expect(result.current.theme).toBe('auto')
        expect(result.current.resolvedTheme).toBe('light')
    })
})
