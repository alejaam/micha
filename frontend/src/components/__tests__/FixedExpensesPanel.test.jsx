import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { FixedExpensesPanel } from '../FixedExpensesPanel'

describe('FixedExpensesPanel', () => {
    it('shows concept from recurring description instead of category UUID', () => {
        render(
            <FixedExpensesPanel
                items={[]}
                recurringItems={[
                    {
                        id: 're-1',
                        is_agnostic: true,
                        expense_type: 'fixed',
                        amount_cents: 100000,
                        category_id: '4a4f67f4-0b9d-4f8c-8a75-5f16d2a8f607',
                        description: 'Internet',
                    },
                ]}
                members={[
                    { id: 'm1', name: 'Ana' },
                    { id: 'm2', name: 'Luis' },
                ]}
                currency="MXN"
            />,
        )

        expect(screen.getByText('Internet')).toBeInTheDocument()
        expect(screen.queryByText('4a4f67f4-0b9d-4f8c-8a75-5f16d2a8f607')).not.toBeInTheDocument()
    })

    it('splits fixed amount across members even if paid_by_member_id is present', () => {
        render(
            <FixedExpensesPanel
                items={[
                    {
                        id: 'fx-1',
                        expense_type: 'fixed',
                        is_shared: true,
                        amount_cents: 100000,
                        paid_by_member_id: 'm1',
                        description: 'Rent',
                    },
                ]}
                recurringItems={[]}
                members={[
                    { id: 'm1', name: 'Ana' },
                    { id: 'm2', name: 'Luis' },
                ]}
                currency="MXN"
            />,
        )

        expect(screen.getAllByText('MX$500.00').length).toBeGreaterThanOrEqual(2)
    })
})
