import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DynamicChartsPanel } from '../DynamicChartsPanel'

vi.mock('framer-motion', async () => {
  const React = await import('react')
  return {
    motion: new Proxy(
      {},
      {
        get: (_, tag) => {
          const Comp = ({ children, ...props }) => React.createElement(tag, props, children)
          Comp.displayName = `motion.${String(tag)}`
          return Comp
        },
      },
    ),
  }
})

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }) => <div>{children}</div>,
  PieChart: ({ children }) => <div>{children}</div>,
  Pie: ({ children }) => <div>{children}</div>,
  Cell: () => null,
  BarChart: ({ children }) => <div>{children}</div>,
  Bar: () => null,
  CartesianGrid: () => null,
  LineChart: ({ children }) => <div>{children}</div>,
  Line: () => null,
  XAxis: () => null,
  YAxis: () => null,
  Tooltip: () => null,
}))

describe('DynamicChartsPanel', () => {
  it('returns nothing when there is no chart data', () => {
    const { container } = render(
      <DynamicChartsPanel
        categoryTotals={[]}
        memberActualVsExpected={[]}
        msiProgress={[]}
        spendingTrend={[]}
      />,
    )

    expect(container).toBeEmptyDOMElement()
  })

  it('renders charts region and keyboard-focusable textual summaries', () => {
    render(
      <DynamicChartsPanel
        categoryTotals={[{ key: 'food', label: 'Food', totalCents: 1000, percentage: 50 }]}
        memberActualVsExpected={[{ memberId: 'm1', memberName: 'Ana', actualCents: 1200, expectedCents: 1000, deltaCents: 200 }]}
        msiProgress={[{ id: 'msi1', description: 'Laptop', currentInstallment: 2, totalInstallments: 6, progressPercent: 33, remainingInstallments: 4 }]}
        spendingTrend={[{ key: '2026-01', label: 'Jan 26', totalCents: 1000 }]}
      />,
    )

    expect(screen.getByRole('region', { name: /dynamic charts/i })).toBeInTheDocument()
    const summaries = screen.getByLabelText(/charts textual summaries/i)
    expect(summaries).toHaveAttribute('tabindex', '0')
    expect(screen.getByText(/top category:/i)).toBeInTheDocument()
  })
})
