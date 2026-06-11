import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { PeriodManagementPanel } from '../PeriodManagementPanel'

const mockApi = vi.hoisted(() => ({
  initializePeriod: vi.fn(),
  simulateClosePeriod: vi.fn(),
}))

vi.mock('../../api', () => mockApi)

describe('PeriodManagementPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders initialize button when period is null (no active period)', () => {
    render(
      <PeriodManagementPanel
        householdId="hh-1"
        period={null}
        onStatusChange={vi.fn()}
        isOwner={true}
      />,
    )

    const section = document.querySelector('.periodActionCard')
    expect(section).toBeInTheDocument()
    expect(screen.getByText('Comenzar seguimiento')).toBeInTheDocument()
  })

  it('renders simulate close button when period status is "open"', () => {
    render(
      <PeriodManagementPanel
        householdId="hh-1"
        period={{ id: 'p1', status: 'open' }}
        onStatusChange={vi.fn()}
        isOwner={true}
      />,
    )

    const section = document.querySelector('.periodActionCard')
    expect(section).toBeInTheDocument()
    expect(screen.getByText('Simular cierre')).toBeInTheDocument()
  })

  it('renders nothing and returns null without householdId', () => {
    const { container } = render(
      <PeriodManagementPanel
        householdId=""
        period={null}
        onStatusChange={vi.fn()}
        isOwner={true}
      />,
    )

    expect(container.innerHTML).toBe('')
  })

  it('does not render for non-owner when period is null', () => {
    const { container } = render(
      <PeriodManagementPanel
        householdId="hh-1"
        period={null}
        onStatusChange={vi.fn()}
        isOwner={false}
      />,
    )

    expect(container.innerHTML).toBe('')
  })
})
