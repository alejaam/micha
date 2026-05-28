import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import { AppHeader } from '../AppHeader'
import { PeriodStatusRibbon } from '../PeriodStatusRibbon'
import { AuthProvider } from '../../context/AuthContext'

describe('PeriodStatusRibbon', () => {
  it('renders open variant with accessible status text', () => {
    render(<PeriodStatusRibbon status="open" />)

    expect(screen.getByText('Abierto')).toBeInTheDocument()
    expect(
      screen.getByRole('status', {
        name: /estado del periodo: periodo abierto/i,
      }),
    ).toBeInTheDocument()
  })

  it('falls back to open when status is unknown', () => {
    render(<PeriodStatusRibbon status="unexpected" />)
    expect(screen.getByText('Abierto')).toBeInTheDocument()
  })

  it('renders closed variant', () => {
    render(<PeriodStatusRibbon status="closed" />)
    expect(screen.getByText('Cerrado')).toBeInTheDocument()
  })
})

describe('AppHeader renders invite member link', () => {
  it('renders the invite member link without lock styling', () => {
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
          />
        </MemoryRouter>
      </AuthProvider>,
    )

    const inviteLink = screen.getByRole('link', { name: /invitar nuevo miembro/i })
    expect(inviteLink).toBeInTheDocument()
    expect(inviteLink).not.toHaveClass('btnDisabled')
  })
})


