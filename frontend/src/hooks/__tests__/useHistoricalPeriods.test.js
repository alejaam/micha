import { renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { getSettlement } from '../../api'
import {
  buildClosedPeriodsFromExpenses,
  useHistoricalPeriods,
} from '../useHistoricalPeriods'

vi.mock('../../api', () => ({
  getSettlement: vi.fn(),
}))

describe('buildClosedPeriodsFromExpenses', () => {
  it('groups by month and excludes current month', () => {
    const now = new Date()
    const prevMonth = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - 1, 10)).toISOString()
    const twoMonthsAgo = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - 2, 10)).toISOString()
    const currentMonth = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 2)).toISOString()

    const periods = buildClosedPeriodsFromExpenses([
      { id: 'a', created_at: prevMonth, amount_cents: 1000 },
      { id: 'b', created_at: prevMonth, amount_cents: 500 },
      { id: 'c', created_at: twoMonthsAgo, amount_cents: 2000 },
      { id: 'd', created_at: currentMonth, amount_cents: 9999 },
    ])

    expect(periods).toHaveLength(2)
    expect(periods[0].totalCents).toBe(1500)
    expect(periods[1].totalCents).toBe(2000)
  })
})

describe('useHistoricalPeriods', () => {
  it('caches fallback snapshot and marks provisional when settlement fetch fails', async () => {
    getSettlement.mockRejectedValueOnce(new Error('no history endpoint'))

    const now = new Date()
    const prevMonth = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - 1, 10)).toISOString()
    const hookInput = {
      householdId: 'home-1',
      members: [{ id: 'm1' }, { id: 'm2' }],
      expenses: [
        {
          id: 'exp-1',
          created_at: prevMonth,
          amount_cents: 3400,
          description: 'Internet',
          expense_type: 'msi',
          total_installments: 1,
        },
      ],
    }

    const { result } = renderHook(() => useHistoricalPeriods(hookInput))

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.closedPeriods.length).toBe(1)
    expect(result.current.isProvisional).toBe(true)
    expect(result.current.provisionalReason).toMatch(/provisional/i)
    expect(result.current.selectedPeriodSnapshot?.source).toBe('mock')
    expect(result.current.completedMsi.length).toBe(1)
  })
})
