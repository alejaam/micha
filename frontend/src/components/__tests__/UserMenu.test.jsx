import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ThemeProvider } from '../../context/ThemeContext'
import { UserMenu } from '../UserMenu'

function renderMenu() {
    return render(
        <ThemeProvider>
            <UserMenu
                user={{ email: 'alice@example.com' }}
                households={[{ id: 'hh-1', name: 'Casa' }]}
                householdId="hh-1"
                onHouseholdChange={() => {}}
                onLogout={() => {}}
                health="ok"
            />
        </ThemeProvider>,
    )
}

describe('UserMenu theme selector', () => {
    afterEach(() => {
        localStorage.clear()
        document.documentElement.removeAttribute('data-theme')
    })

    it('renders the three theme options with Auto active by default', () => {
        renderMenu()

        fireEvent.click(screen.getByRole('button', { name: 'Menú de usuario' }))

        const auto = screen.getByRole('menuitemradio', { name: /Auto/ })
        expect(auto).toHaveAttribute('aria-checked', 'true')
        expect(screen.getByRole('menuitemradio', { name: /Claro/ })).toHaveAttribute('aria-checked', 'false')
        expect(screen.getByRole('menuitemradio', { name: /Oscuro/ })).toHaveAttribute('aria-checked', 'false')
    })

    it('selecting Oscuro persists the preference and updates the check', () => {
        renderMenu()

        fireEvent.click(screen.getByRole('button', { name: 'Menú de usuario' }))
        fireEvent.click(screen.getByRole('menuitemradio', { name: /Oscuro/ }))

        expect(localStorage.getItem('micha-theme')).toBe('dark')
        expect(screen.getByRole('menuitemradio', { name: /Oscuro/ })).toHaveAttribute('aria-checked', 'true')
        expect(screen.getByRole('menuitemradio', { name: /Auto/ })).toHaveAttribute('aria-checked', 'false')
        expect(document.documentElement.dataset.theme).toBe('dark')
    })

    it('selecting Claro overrides a dark system preference', () => {
        // The global jsdom stub reports a light system preference.
        renderMenu()

        fireEvent.click(screen.getByRole('button', { name: 'Menú de usuario' }))
        fireEvent.click(screen.getByRole('menuitemradio', { name: /Claro/ }))

        expect(document.documentElement.dataset.theme).toBe('light')
    })
})
