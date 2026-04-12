import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { OnboardingFixedExpensesPage } from '../OnboardingFixedExpensesPage'

const mockCreateRecurringExpense = vi.fn()
const mockNavigate = vi.fn()
const mockUseAppShell = vi.fn()
const mockUseAuth = vi.fn()

vi.mock('../../api', async () => {
    const actual = await vi.importActual('../../api')
    return {
        ...actual,
        createRecurringExpense: (...args) => mockCreateRecurringExpense(...args),
    }
})

vi.mock('../../context/AppShellContext', () => ({
    useAppShell: (...args) => mockUseAppShell(...args),
}))

vi.mock('../../context/AuthContext', () => ({
    useAuth: (...args) => mockUseAuth(...args),
}))

vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom')
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    }
})

describe('OnboardingFixedExpensesPage', () => {
    beforeEach(() => {
        mockCreateRecurringExpense.mockReset()
        mockNavigate.mockReset()
        mockUseAppShell.mockReturnValue({ householdId: 'hh-1' })
        mockUseAuth.mockReturnValue({ handleProtectedError: () => false })
    })

    it('creates agnostic recurring fixed expenses for selected options', async () => {
        mockCreateRecurringExpense.mockResolvedValue({})

        render(
            <MemoryRouter>
                <OnboardingFixedExpensesPage />
            </MemoryRouter>,
        )

        fireEvent.click(screen.getByLabelText('Rent'))
        fireEvent.change(screen.getByPlaceholderText('0.00'), { target: { value: '1250.50' } })
        fireEvent.click(screen.getByRole('button', { name: 'Save and continue' }))

        await waitFor(() => expect(mockCreateRecurringExpense).toHaveBeenCalledTimes(1))
        expect(mockCreateRecurringExpense).toHaveBeenCalledWith(expect.objectContaining({
            householdId: 'hh-1',
            isAgnostic: true,
            expenseType: 'fixed',
            recurrencePattern: 'monthly',
            category: 'rent',
            amountCents: 125050,
        }))
        expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true })
    })
})
