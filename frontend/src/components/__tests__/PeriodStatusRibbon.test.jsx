import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { AppHeader } from '../AppHeader'
import { PeriodStatusRibbon } from '../PeriodStatusRibbon'
import { AuthProvider } from '../../context/AuthContext'

describe('PeriodStatusRibbon', () => {
  it('renders review variant with accessible status text', () => {
    render(<PeriodStatusRibbon status="review" />)

    expect(screen.getByText('Revisión')).toBeInTheDocument()
    expect(
      screen.getByRole('status', {
        name: /estado del periodo: periodo en revisión/i,
      }),
    ).toBeInTheDocument()
  })

  it('falls back to open when status is unknown', () => {
    render(<PeriodStatusRibbon status="unexpected" />)
    expect(screen.getByText('Abierto')).toBeInTheDocument()
  })
})

describe('AppHeader mutation lock wiring', () => {
  it('marks invite member action as disabled while period is locked', () => {
    render(
      <AuthProvider>
        <MemoryRouter>
          <AppHeader
            health="ok"
            householdId="house-1"
            households={[{ id: 'house-1', name: 'Home' }]}
            onHouseholdChange={() => {}}
            onReload={() => {}}
            onLogout={() => {}}
            isLoading={false}
            periodStatus="closed"
            isMutationLocked
          />
        </MemoryRouter>
      </AuthProvider>,
    )

    const inviteLink = screen.getByRole('link', { name: /invitar nuevo miembro/i })
    expect(inviteLink).toHaveAttribute('aria-disabled', 'true')
    expect(inviteLink).toHaveClass('btnDisabled')
  })
})


