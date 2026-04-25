import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { DashboardPage } from '../../pages/DashboardPage'

vi.mock('framer-motion', async () => {
  const React = await import('react')
  return {
    AnimatePresence: ({ children }) => <>{children}</>,
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

const mockUseHouseholdData = vi.fn()

vi.mock('../../hooks/useHouseholdData', () => ({
  useHouseholdData: () => mockUseHouseholdData(),
}))

const baseMembers = [{ id: 'm1', name: 'Ana', monthly_salary_cents: 100000 }]

function makeExpense(overrides = {}) {
  return {
    id: 'e1',
    amount_cents: 150000,
    description: 'Rent',
    category: 'rent',
    category_name: 'Rent',
    created_at: '2026-01-10T00:00:00.000Z',
    paid_by_member_id: 'm1',
    is_shared: true,
    expense_type: 'variable',
    ...overrides,
  }
}

function makeDefaultState(overrides = {}) {
  const items = overrides.items || [makeExpense()]
  return {
    members: baseMembers,
    loadingMembers: false,
    items,
    loadingList: false,
    recurringItems: [],
    settlement: {
      members: [{ member_id: 'm1', member_name: 'Ana', net_balance_cents: 0, expected_share: 150000, paid_cents: 150000 }],
      transfers: [],
      total_shared_cents: 150000,
      included_expense_count: items.length,
      is_closed: false,
    },
    currentMember: baseMembers[0],
    activeCurrency: 'MXN',
    householdId: 'house-1',
    isMutationLocked: false,
    categoryTotals: [],
    memberActualVsExpected: [],
    msiProgress: [],
    spendingTrend: [],
    handleCreate: vi.fn().mockResolvedValue(true),
    message: '',
    setMessage: vi.fn(),
    error: '',
    setError: vi.fn(),
    submittingCreate: false,
    ...overrides,
  }
}

function renderDashboard(custom = {}) {
  mockUseHouseholdData.mockReturnValue(makeDefaultState(custom))

  return render(
    <MemoryRouter>
      <DashboardPage />
    </MemoryRouter>,
  )
}

describe('DashboardPage integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders overview sections and priority strip when data exists', () => {
    renderDashboard()

    expect(screen.getByText(/balances y conciliación primero/i)).toBeInTheDocument()
    expect(screen.getByText(/transferencias pendientes/i)).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /gastos recientes/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /ver balances →/i })).toBeInTheDocument()
  })

  it('shows empty state when no expenses exist', () => {
    renderDashboard({ items: [], recurringItems: [] })

    expect(screen.getByText(/sin gastos aún/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /añadir rápido/i })).toBeInTheDocument()
  })

  it('opens quick add bottom sheet when clicking empty state CTA', async () => {
    renderDashboard({ items: [], recurringItems: [] })

    fireEvent.click(screen.getByRole('button', { name: /añadir rápido/i }))

    const dialog = await screen.findByRole('dialog', { name: /añadir rápido/i })
    expect(dialog).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /close panel/i })).toBeInTheDocument()
  })

  it('submits quick-add form and calls handleCreate', async () => {
    const handleCreate = vi.fn().mockResolvedValue(true)
    renderDashboard({ items: [], recurringItems: [], handleCreate })

    fireEvent.click(screen.getByRole('button', { name: /añadir rápido/i }))

    fireEvent.change(screen.getByLabelText(/amount in dollars/i), { target: { value: '12.50' } })
    fireEvent.change(screen.getByLabelText(/description/i), { target: { value: 'Milk' } })

    const form = screen.getByRole('dialog', { name: /añadir rápido/i }).querySelector('form')
    fireEvent.submit(form)

    await waitFor(() => {
      expect(handleCreate).toHaveBeenCalled()
    })
  })
})
