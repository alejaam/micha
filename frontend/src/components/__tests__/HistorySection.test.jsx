import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { HistorySection } from '../HistorySection'

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }) => <div>{children}</div>,
  LineChart: ({ children }) => <div>{children}</div>,
  Line: () => <div data-testid="line" />,
  XAxis: () => null,
  YAxis: () => null,
  Tooltip: () => null,
}))

describe('HistorySection', () => {
  it('renders empty state with quick-add CTA when no closed periods', () => {
    const onQuickAdd = vi.fn()
    render(
      <HistorySection
        closedPeriods={[]}
        selectedPeriodKey=""
        onSelectPeriod={() => {}}
        comparisonSeries={[]}
        memberBalanceTrend={[]}
        completedMsi={[]}
        onQuickAdd={onQuickAdd}
      />,
    )

    expect(screen.getByText(/aún no hay periodos cerrados/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /agregar gasto/i })).toBeInTheDocument()
  })

  it('renders closed period list and provisional label', () => {
    const onSelectPeriod = vi.fn()
    render(
      <HistorySection
        closedPeriods={[{ key: '2026-01', label: 'Jan 26', totalCents: 1500 }]}
        selectedPeriodKey="2026-01"
        onSelectPeriod={onSelectPeriod}
        comparisonSeries={[{ key: '2026-01', label: 'Jan 26', totalCents: 1500 }]}
        memberBalanceTrend={[{ key: '2026-01', label: 'Jan 26', Ana: 1200 }]}
        completedMsi={[{ id: 'm1', description: 'Laptop', totalInstallments: 6 }]}
        isProvisional
        provisionalReason="Historical endpoint unavailable"
      />,
    )

    expect(screen.getByText('PROVISIONAL')).toBeInTheDocument()
    const janButton = screen.getByRole('button', { name: /jan 26/i })
    expect(janButton).toBeInTheDocument()
    fireEvent.click(janButton)
    expect(onSelectPeriod).toHaveBeenCalledWith('2026-01')
    expect(screen.getByText(/historical endpoint unavailable/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /reintentar historial/i })).toBeInTheDocument()
  })

  it('keeps section regions and list semantics for keyboard/a11y navigation', () => {
    render(
      <HistorySection
        closedPeriods={[{ key: '2026-01', label: 'Jan 26', totalCents: 1500 }]}
        selectedPeriodKey="2026-01"
        onSelectPeriod={() => {}}
        comparisonSeries={[{ key: '2026-01', label: 'Jan 26', totalCents: 1500 }]}
        memberBalanceTrend={[]}
        completedMsi={[]}
      />,
    )

    expect(screen.getByLabelText(/history section/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/closed periods/i)).toBeInTheDocument()
    expect(screen.getByRole('list')).toBeInTheDocument()
  })
})
