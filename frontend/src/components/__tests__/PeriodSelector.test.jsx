import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { PeriodSelector } from '../PeriodSelector'

describe('PeriodSelector', () => {
  it('shows placeholder when no periods exist and no currentPeriodId', () => {
    const { container } = render(
      <PeriodSelector
        periods={[]}
        selectedPeriodId={null}
        currentPeriodId={null}
        onSelect={vi.fn()}
      />,
    )

    expect(screen.getByText('Sin periodos aún')).toBeInTheDocument()
    const div = container.querySelector('.periodSelector--empty')
    expect(div).toBeInTheDocument()
  })

  it('renders select with current period option when currentPeriodId is set', () => {
    render(
      <PeriodSelector
        periods={[
          { id: 'p1', start_date: '2026-01-01T00:00:00Z', end_date: '2026-01-15T00:00:00Z', status: 'open' },
        ]}
        selectedPeriodId={null}
        currentPeriodId="p1"
        onSelect={vi.fn()}
      />,
    )

    // Should show the current period option
    const select = screen.getByRole('combobox')
    expect(select).toBeInTheDocument()
    expect(select.options.length).toBeGreaterThanOrEqual(1)
  })
})
