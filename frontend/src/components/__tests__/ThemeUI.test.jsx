import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { AppHeader } from '../AppHeader'
import { UserMenu } from '../UserMenu'
import { ThemeProvider } from '../../context/ThemeContext'
import { AuthProvider } from '../../context/AuthContext'

describe('Theme UI integration', () => {
    it('renders theme toggle in AppHeader and toggles theme when clicked', () => {
        render(
            <ThemeProvider>
                <AuthProvider>
                    <MemoryRouter>
                        <AppHeader
                            health="ok"
                            householdId="h-1"
                            households={[{ id: 'h-1', name: 'Casa' }]}
                            onHouseholdChange={() => {}}
                            onReload={() => {}}
                            onLogout={() => {}}
                        />
                    </MemoryRouter>
                </AuthProvider>
            </ThemeProvider>,
        )

        const toggleBtn = screen.getByRole('button', { name: /cambiar tema/i })
        expect(toggleBtn).toBeInTheDocument()
        expect(toggleBtn).toHaveTextContent('💻')

        // Click toggle -> switches theme
        fireEvent.click(toggleBtn)
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
        expect(toggleBtn).toHaveTextContent('🌙')

        fireEvent.click(toggleBtn)
        expect(document.documentElement.getAttribute('data-theme')).toBe('light')
        expect(toggleBtn).toHaveTextContent('☀️')
    })

    it('renders theme options in UserMenu and allows selecting auto, light, dark', () => {
        render(
            <ThemeProvider>
                <UserMenu
                    user={{ email: 'user@example.com' }}
                    households={[{ id: 'h-1', name: 'Casa' }]}
                    householdId="h-1"
                    onHouseholdChange={() => {}}
                    onLogout={() => {}}
                    health="ok"
                />
            </ThemeProvider>,
        )

        // Open menu
        const avatarBtn = screen.getByRole('button', { name: /menú de usuario/i })
        fireEvent.click(avatarBtn)

        expect(screen.getByText('Tema')).toBeInTheDocument()
        const autoBtn = screen.getByRole('button', { name: /auto/i })
        const lightBtn = screen.getByRole('button', { name: /claro/i })
        const darkBtn = screen.getByRole('button', { name: /oscuro/i })

        expect(autoBtn).toBeInTheDocument()
        expect(lightBtn).toBeInTheDocument()
        expect(darkBtn).toBeInTheDocument()

        // Select dark
        fireEvent.click(darkBtn)
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
        expect(localStorage.getItem('micha-theme')).toBe('dark')

        // Select light
        fireEvent.click(lightBtn)
        expect(document.documentElement.getAttribute('data-theme')).toBe('light')
        expect(localStorage.getItem('micha-theme')).toBe('light')

        // Select auto
        fireEvent.click(autoBtn)
        expect(localStorage.getItem('micha-theme')).toBe('auto')
    })
})
