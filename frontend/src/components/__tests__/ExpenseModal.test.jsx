import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { ExpenseModal } from '../ExpenseModal'

vi.mock('../../api', () => ({
  listCategories: vi.fn().mockResolvedValue([
    { id: 'food', slug: 'food', name: 'Food' },
    { id: 'other', slug: 'other', name: 'Other' },
  ]),
  listCards: vi.fn().mockResolvedValue([]),
}))

function renderModal(overrides = {}) {
  return render(
    <MemoryRouter>
      <ExpenseModal
        onClose={() => {}}
        onSubmit={vi.fn().mockResolvedValue(true)}
        isSubmitting={false}
        isMutationLocked={false}
        members={[
          { id: 'm-owner', name: 'Owner', user_id: 'u-1', created_at: '2026-01-01T00:00:00Z' },
          { id: 'm-roomie', name: 'Roomie', user_id: 'u-2', created_at: '2026-01-02T00:00:00Z' },
        ]}
        isLoadingMembers={false}
        defaultPaidByMemberId="m-owner"
        householdId="hh-1"
        {...overrides}
      />
    </MemoryRouter>,
  )
}

describe('ExpenseModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not allow selecting fixed expense type in general expense flow', async () => {
    renderModal()
    fireEvent.click(screen.getByRole('button', { name: /más opciones/i }))

    const msiCheckbox = screen.getByLabelText(/pagar a meses/i)
    expect(msiCheckbox).not.toBeChecked()

    fireEvent.click(msiCheckbox)
    expect(screen.getByLabelText(/total de cuotas/i)).toBeInTheDocument()
  })

  it('shows owner context when selecting paid by', async () => {
    renderModal()
    expect(await screen.findByText(/como owner, puedes registrar gastos para otros miembros/i)).toBeInTheDocument()
  })
})
