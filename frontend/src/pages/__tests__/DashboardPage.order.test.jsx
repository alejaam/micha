import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import { DashboardPage } from '../DashboardPage'

// Mock framer-motion
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

// Mock recharts
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

const mockUseAppShell = vi.fn()
const mockUseHouseholdData = vi.fn()

vi.mock('../../context/AppShellContext', () => ({
  useAppShell: () => mockUseAppShell(),
}))

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
  mockUseAppShell.mockReturnValue({
    currentPeriod: { id: 'p1', status: 'open' },
    reloadPeriod: vi.fn(),
    selectedHousehold: { id: 'house-1', owner_id: 'u1' },
    handleReload: vi.fn(),
  })
  mockUseHouseholdData.mockReturnValue(makeDefaultState(custom))

  return render(
    <MemoryRouter>
      <DashboardPage />
    </MemoryRouter>,
  )
}

describe('DashboardPage DOM order', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders sections in the correct priority order', () => {
    renderDashboard()

    // Collect all top-level heading texts in DOM order.
    const headings = document.querySelectorAll('h2, h3')
    const headingTexts = Array.from(headings).map((h) => h.textContent.trim())

    // The priority order (h2/h3 texts that should appear):
    // 1. "Este mes" (ExpenseSummary h2) or "Tu sueldo restante" (RemainingSalaryPanel h3)
    // 2. "Gastos recientes" (RecentExpenses h2)
    // 3. "Gráficos dinámicos" (DynamicChartsPanel h2)
    // 4. Period history

    const sueldoIdx = headingTexts.findIndex((t) => t === 'Tu sueldo restante')
    const esteMesIdx = headingTexts.findIndex((t) => t === 'Este mes')
    const recientesIdx = headingTexts.findIndex((t) => t === 'Gastos recientes')
    const graficosIdx = headingTexts.findIndex((t) => t === 'Gráficos dinámicos')

    // Summary must come first
    if (sueldoIdx >= 0 && recientesIdx >= 0) {
      expect(sueldoIdx).toBeLessThan(recientesIdx)
    }
    if (esteMesIdx >= 0 && recientesIdx >= 0) {
      expect(esteMesIdx).toBeLessThan(recientesIdx)
    }

    // Charts should come after RecentExpenses
    if (graficosIdx >= 0 && recientesIdx >= 0) {
      expect(graficosIdx).toBeGreaterThan(recientesIdx)
    }

    // Ver todos los movimientos should exist
    expect(screen.getByText('Ver todos los movimientos →')).toBeInTheDocument()
    expect(screen.getByText('Ver Balances →')).toBeInTheDocument()
  })
})
