import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { PeriodManagementPanel } from '../PeriodManagementPanel'

const mockApi = vi.hoisted(() => ({
  initializePeriod: vi.fn(),
  transitionPeriodToReview: vi.fn(),
  approvePeriod: vi.fn(),
  closePeriod: vi.fn(),
}))

vi.mock('../../api', () => mockApi)

describe('PeriodManagementPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders as banner when period is null (no active period)', () => {
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
    expect(section.classList.contains('periodActionCard--banner')).toBe(true)
    expect(screen.getByText('Comenzar seguimiento')).toBeInTheDocument()
  })

  it('renders as compact card when period status is "open"', () => {
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
    expect(section.classList.contains('periodActionCard--banner')).toBe(false)
    expect(screen.getByText('Iniciar revisión')).toBeInTheDocument()
  })

  it('renders as banner when period status is "review"', () => {
    render(
      <PeriodManagementPanel
        householdId="hh-1"
        period={{ id: 'p1', status: 'review' }}
        onStatusChange={vi.fn()}
        isOwner={true}
      />,
    )

    const section = document.querySelector('.periodActionCard')
    expect(section).toBeInTheDocument()
    expect(section.classList.contains('periodActionCard--banner')).toBe(true)
    expect(screen.getByText('Periodo en revisión')).toBeInTheDocument()
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
