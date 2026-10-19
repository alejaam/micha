import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { OnboardingFixedExpensesPage } from '../OnboardingFixedExpensesPage'

const mockCreateRecurringExpense = vi.fn()
const mockListRecurringExpenses = vi.fn()
const mockListSubscriptionServices = vi.fn()
const mockUpdateRecurringExpense = vi.fn()
const mockDeleteRecurringExpense = vi.fn()
const mockNavigate = vi.fn()
const mockUseAppShell = vi.fn()
const mockUseAuth = vi.fn()

vi.mock('../../api', async () => {
    const actual = await vi.importActual('../../api')
    return {
        ...actual,
        createRecurringExpense: (...args) => mockCreateRecurringExpense(...args),
        listRecurringExpenses: (...args) => mockListRecurringExpenses(...args),
        listSubscriptionServices: (...args) => mockListSubscriptionServices(...args),
        updateRecurringExpense: (...args) => mockUpdateRecurringExpense(...args),
        deleteRecurringExpense: (...args) => mockDeleteRecurringExpense(...args),
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
        mockListRecurringExpenses.mockReset()
        mockListSubscriptionServices.mockReset()
        mockUpdateRecurringExpense.mockReset()
        mockDeleteRecurringExpense.mockReset()
        mockNavigate.mockReset()
        mockUseAppShell.mockReturnValue({ householdId: 'hh-1' })
        mockUseAuth.mockReturnValue({ handleProtectedError: () => false })
        mockListRecurringExpenses.mockResolvedValue([])
        mockListSubscriptionServices.mockResolvedValue([
            { id: 'svc-1', slug: 'rent', name: 'Renta', standalone_price_cents: 120000, currency: 'MXN', is_bundle: false },
            { id: 'svc-2', slug: 'internet', name: 'Internet', standalone_price_cents: 59900, currency: 'MXN', is_bundle: false },
        ])
    })

    it('creates agnostic recurring fixed expenses for selected catalog services', async () => {
        mockCreateRecurringExpense.mockResolvedValue({})

        render(
            <MemoryRouter initialEntries={['/onboarding/fixed-expenses']}>
                <OnboardingFixedExpensesPage />
            </MemoryRouter>,
        )

        // Wait for catalog to load
        const rentBtn = await screen.findByText('Renta')
        fireEvent.click(rentBtn)

        // Amount should be pre-filled from catalog standalone_price_cents (120000 → 1200.00)
        const amountInput = screen.getByDisplayValue('1200.00')
        fireEvent.change(amountInput, { target: { value: '1250.50' } })

        fireEvent.click(screen.getByRole('button', { name: 'Guardar y continuar' }))

        await waitFor(() => expect(mockCreateRecurringExpense).toHaveBeenCalledTimes(1))
        expect(mockCreateRecurringExpense).toHaveBeenCalledWith(expect.objectContaining({
            householdId: 'hh-1',
            isAgnostic: true,
            expenseType: 'fixed',
            recurrencePattern: 'monthly',
            category: 'other',
            amountCents: 125050,
        }))
        expect(mockNavigate).toHaveBeenCalledWith('/', { replace: true })
    })
})
