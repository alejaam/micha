import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AuthProvider } from '../../context/AuthContext'
import { ThemeProvider } from '../../context/ThemeContext'
import { AppHeader } from '../AppHeader'

vi.mock('../../api', () => ({
    listPeriods: vi.fn().mockResolvedValue([]),
    setAuthToken: vi.fn(),
    getMe: vi.fn().mockResolvedValue(null),
    loginUser: vi.fn(),
    registerUser: vi.fn(),
}))

function renderHeader() {
    return render(
        <MemoryRouter>
            <ThemeProvider>
                <AuthProvider>
                    <AppHeader
                        health="ok"
                        householdId="hh-1"
                        onHouseholdChange={() => {}}
                        onReload={() => {}}
                        onLogout={() => {}}
                        isLoading={false}
                        households={[{ id: 'hh-1', name: 'Casa' }]}
                    />
                </AuthProvider>
            </ThemeProvider>
        </MemoryRouter>,
    )
}

describe('AppHeader theme quick toggle', () => {
    afterEach(() => {
        localStorage.clear()
        document.documentElement.removeAttribute('data-theme')
    })

    it('offers to switch to dark when the resolved theme is light', () => {
        // The global jsdom stub reports a light system preference.
        renderHeader()

        expect(screen.getByRole('button', { name: 'Cambiar a tema oscuro' })).toBeInTheDocument()
    })

    it('toggles to an explicit dark theme and persists it', () => {
        renderHeader()

        fireEvent.click(screen.getByRole('button', { name: 'Cambiar a tema oscuro' }))

        expect(screen.getByRole('button', { name: 'Cambiar a tema claro' })).toBeInTheDocument()
        expect(localStorage.getItem('micha-theme')).toBe('dark')
        expect(document.documentElement.dataset.theme).toBe('dark')
    })
})
