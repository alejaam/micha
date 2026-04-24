import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { AppHeader } from '../AppHeader'
import { PeriodStatusRibbon } from '../PeriodStatusRibbon'
import { buildRibbonState } from '../../hooks/useDashboardUxState'

describe('PeriodStatusRibbon', () => {
  it('renders review variant with accessible status text', () => {
    render(<PeriodStatusRibbon status="review" />)

    expect(screen.getByText('[REVIEW]')).toBeInTheDocument()
    expect(
      screen.getByRole('status', {
        name: /estado del periodo: periodo en revisión/i,
      }),
    ).toBeInTheDocument()
  })

  it('falls back to open when status is unknown', () => {
    render(<PeriodStatusRibbon status="unexpected" />)
    expect(screen.getByText('[OPEN]')).toBeInTheDocument()
  })
})

describe('AppHeader mutation lock wiring', () => {
  it('marks invite member action as disabled while period is locked', () => {
    render(
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
      </MemoryRouter>,
    )

    const inviteLink = screen.getByRole('link', { name: /invitar nuevo miembro/i })
    expect(inviteLink).toHaveAttribute('aria-disabled', 'true')
    expect(inviteLink).toHaveClass('btnDisabled')
  })
})

describe('ribbon state derivation', () => {
  it('maps closed state to expected content', () => {
    expect(buildRibbonState('closed')).toMatchObject({
      status: 'closed',
      stateLabel: '[CLOSED]',
    })
  })
})
