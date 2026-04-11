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

const mockCreateExpense = vi.fn()
const mockDeleteExpense = vi.fn()
const mockPatchExpense = vi.fn()

vi.mock('../../api', async () => {
  const actual = await vi.importActual('../../api')
  return {
    ...actual,
    createExpense: (...args) => mockCreateExpense(...args),
    deleteExpense: (...args) => mockDeleteExpense(...args),
    patchExpense: (...args) => mockPatchExpense(...args),
  }
})

const mockUseMembers = vi.fn()
const mockUseExpenses = vi.fn()
const mockUseSettlement = vi.fn()
const mockUseHistoricalPeriods = vi.fn()
const mockUseCurrentMember = vi.fn()
const mockUseAppShell = vi.fn()
const mockUseAuth = vi.fn()

vi.mock('../../context/AppShellContext', () => ({
  useAppShell: (...args) => mockUseAppShell(...args),
}))

vi.mock('../../context/AuthContext', () => ({
  useAuth: (...args) => mockUseAuth(...args),
}))

vi.mock('../../hooks/useMembers', () => ({
  useMembers: (...args) => mockUseMembers(...args),
}))

vi.mock('../../hooks/useExpenses', () => ({
  useExpenses: (...args) => mockUseExpenses(...args),
}))

vi.mock('../../hooks/useSettlement', () => ({
  useSettlement: (...args) => mockUseSettlement(...args),
}))

vi.mock('../../hooks/useHistoricalPeriods', () => ({
  useHistoricalPeriods: (...args) => mockUseHistoricalPeriods(...args),
}))

vi.mock('../../hooks/useCurrentMember', () => ({
  useCurrentMember: (...args) => mockUseCurrentMember(...args),
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

function makeHookState({ items = [makeExpense()], closedPeriods = [{ key: '2026-01', label: 'Jan 26', totalCents: 150000, year: 2026, month: 1 }] } = {}) {
  return {
    members: baseMembers,
    loadingMembers: false,
    loadMembers: vi.fn(),
    items,
    loadingList: false,
    loadExpenses: vi.fn().mockResolvedValue(undefined),
    settlement: {
      members: [{ member_id: 'm1', member_name: 'Ana', net_balance_cents: 0, expected_share: 150000, paid_cents: 150000 }],
      transfers: [],
      total_shared_cents: 150000,
      included_expense_count: items.length,
      excluded_voucher_count: 0,
      effective_settlement_mode: 'exact',
      is_closed: false,
    },
    loadingSettlement: false,
    settlementYear: new Date().getUTCFullYear(),
    settlementMonth: new Date().getUTCMonth() + 1,
    setSettlementYear: vi.fn(),
    setSettlementMonth: vi.fn(),
    loadSettlement: vi.fn().mockResolvedValue(undefined),
    resetToCurrentMonth: vi.fn(),
    historical: {
      closedPeriods,
      selectedPeriodKey: closedPeriods[0]?.key ?? '',
      setSelectedPeriodKey: vi.fn(),
      comparisonSeries: closedPeriods.map((p) => ({ key: p.key, label: p.label, totalCents: p.totalCents })),
      memberBalanceTrend: closedPeriods.map((p) => ({ key: p.key, label: p.label, Ana: 0 })),
      completedMsi: [],
      selectedPeriodSnapshot: { memberBalances: [{ memberId: 'm1', memberName: 'Ana', netBalanceCents: 0 }] },
      isLoading: false,
      isProvisional: false,
      provisionalReason: '',
    },
    currentMember: baseMembers[0],
  }
}

function renderDashboard(custom = {}) {
  const state = makeHookState(custom)
  const historical = {
    ...state.historical,
    ...(custom.historical ?? {}),
  }

  mockUseMembers.mockReturnValue({
    members: state.members,
    loadingMembers: state.loadingMembers,
    loadMembers: state.loadMembers,
  })
  mockUseExpenses.mockReturnValue({
    items: state.items,
    loadingList: state.loadingList,
    loadExpenses: state.loadExpenses,
  })
  mockUseSettlement.mockReturnValue({
    settlement: state.settlement,
    loadingSettlement: state.loadingSettlement,
    settlementYear: state.settlementYear,
    settlementMonth: state.settlementMonth,
    setSettlementYear: state.setSettlementYear,
    setSettlementMonth: state.setSettlementMonth,
    loadSettlement: state.loadSettlement,
    resetToCurrentMonth: state.resetToCurrentMonth,
  })
  mockUseHistoricalPeriods.mockReturnValue(historical)
  mockUseCurrentMember.mockReturnValue(state.currentMember)

  const appShell = {
    householdId: 'house-1',
    selectedHousehold: { id: 'house-1', currency: 'MXN', created_at: '2025-01-01T00:00:00.000Z' },
    setPeriodStatus: vi.fn(),
    isMutationLocked: false,
    ...custom.appShell,
  }

  const auth = {
    isAuthenticated: true,
    handleProtectedError: vi.fn().mockReturnValue(false),
    ...custom.auth,
  }

  mockUseAppShell.mockReturnValue(appShell)
  mockUseAuth.mockReturnValue(auth)

  return render(
    <MemoryRouter>
      <DashboardPage />
    </MemoryRouter>,
  )
}

describe('DashboardPage integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCreateExpense.mockResolvedValue(undefined)
    mockDeleteExpense.mockResolvedValue(undefined)
    mockPatchExpense.mockResolvedValue(undefined)
  })

  it('renders composed dashboard sections and chart summaries when data exists', () => {
    renderDashboard()

    expect(screen.getByRole('region', { name: /dynamic charts/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/charts textual summaries/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/history section/i)).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /recent expenses/i })).toBeInTheDocument()
  })

  it('shows history loading indicator content path and no empty history state while loading', () => {
    renderDashboard({
      closedPeriods: [],
      appShell: { isMutationLocked: false },
      auth: { isAuthenticated: true },
      items: [makeExpense()],
      historical: {
        closedPeriods: [],
        selectedPeriodKey: '',
        setSelectedPeriodKey: vi.fn(),
        comparisonSeries: [],
        memberBalanceTrend: [],
        completedMsi: [],
        selectedPeriodSnapshot: null,
        isLoading: true,
        isProvisional: false,
        provisionalReason: '',
      },
    })

    expect(screen.getByLabelText(/history section/i)).toBeInTheDocument()
    expect(screen.queryByText(/no closed periods yet/i)).not.toBeInTheDocument()
  })

  it('shows empty-state quick-add flow and opens bottom sheet with focusable dialog', async () => {
    renderDashboard({ items: [], closedPeriods: [] })

    const quickAddButtons = screen.getAllByRole('button', { name: /quick add|add expense/i })
    fireEvent.click(quickAddButtons[0])

    const dialog = await screen.findByRole('dialog', { name: /quick add/i })
    expect(dialog).toHaveAttribute('aria-modal', 'true')
    expect(dialog).toHaveAttribute('tabindex', '-1')
    expect(screen.getByRole('button', { name: /close panel/i })).toHaveFocus()
  })

  it('submits quick-add form and closes bottom sheet after successful create', async () => {
    renderDashboard({ items: [], closedPeriods: [] })

    fireEvent.click(screen.getAllByRole('button', { name: /quick add|add expense/i })[0])

    fireEvent.change(screen.getByLabelText(/amount in dollars/i), { target: { value: '12.50' } })
    fireEvent.change(screen.getByLabelText(/description/i), { target: { value: 'Milk' } })

    const form = screen.getByRole('dialog', { name: /quick add/i }).querySelector('form')
    fireEvent.submit(form)

    await waitFor(() => {
      expect(mockCreateExpense).toHaveBeenCalledTimes(1)
    })

    await waitFor(() => {
      expect(screen.queryByRole('dialog', { name: /quick add/i })).not.toBeInTheDocument()
    })
  })
})
