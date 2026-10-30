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

    it('shows error and disables button when custom expense is partially filled', async () => {
        render(
            <MemoryRouter initialEntries={['/onboarding/fixed-expenses']}>
                <OnboardingFixedExpensesPage />
            </MemoryRouter>,
        )

        const customLabelInput = screen.getByLabelText('Nombre del gasto personalizado')
        fireEvent.change(customLabelInput, { target: { value: 'Gimnasio' } })

        // Save button appears but is disabled, and warning is shown
        const saveBtn = screen.getByRole('button', { name: 'Guardar y continuar' })
        expect(saveBtn).toBeDisabled()
        expect(screen.getByText('Ingresa un nombre y un monto válido para el gasto personalizado.')).toBeInTheDocument()

        // Once valid amount is provided, button is enabled
        const customAmountInput = screen.getByLabelText('Monto mensual del gasto personalizado')
        fireEvent.change(customAmountInput, { target: { value: '500' } })
        expect(saveBtn).not.toBeDisabled()
    })

    it('does not duplicate previously saved items on partial save retry', async () => {
        // First call succeeds (Renta), second fails (Internet)
        mockCreateRecurringExpense
            .mockResolvedValueOnce({})
            .mockRejectedValueOnce(new Error('Network error on second item'))

        render(
            <MemoryRouter initialEntries={['/onboarding/fixed-expenses']}>
                <OnboardingFixedExpensesPage />
            </MemoryRouter>,
        )

        const rentBtn = await screen.findByText('Renta')
        fireEvent.click(rentBtn)
        const internetBtn = await screen.findByText('Internet')
        fireEvent.click(internetBtn)

        const saveBtn = screen.getByRole('button', { name: 'Guardar y continuar' })
        fireEvent.click(saveBtn)

        await waitFor(() => expect(screen.getByText('Network error on second item')).toBeInTheDocument())
        expect(mockCreateRecurringExpense).toHaveBeenCalledTimes(2)

        // Renta succeeded so it was deselected. Now mock success on retry:
        mockCreateRecurringExpense.mockResolvedValueOnce({})
        fireEvent.click(saveBtn)

        // Only Internet should be retried (call #3), Renta should NOT be duplicated
        await waitFor(() => expect(mockCreateRecurringExpense).toHaveBeenCalledTimes(3))
        expect(mockCreateRecurringExpense).toHaveBeenLastCalledWith(expect.objectContaining({
            description: 'Internet',
        }))
    })
})
