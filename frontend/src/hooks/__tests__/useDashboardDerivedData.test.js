import { describe, expect, it } from 'vitest'
import { buildDashboardDerivedData } from '../useDashboardDerivedData'

describe('buildDashboardDerivedData', () => {
  it('builds category totals and spending trend from expenses', () => {
    const result = buildDashboardDerivedData({
      expenses: [
        {
          id: 'e-1',
          amount_cents: 10000,
          category: 'food',
          category_name: 'Food',
          paid_by_member_id: 'm-1',
          created_at: '2026-01-12T00:00:00.000Z',
        },
        {
          id: 'e-2',
          amount_cents: 25000,
          category: 'rent',
          category_name: 'Rent',
          paid_by_member_id: 'm-2',
          created_at: '2026-02-11T00:00:00.000Z',
        },
      ],
      members: [
        { id: 'm-1', name: 'Ana', monthly_salary_cents: 50000 },
        { id: 'm-2', name: 'Luis', monthly_salary_cents: 50000 },
      ],
      settlement: null,
    })

    expect(result.categoryTotals).toHaveLength(2)
    expect(result.categoryTotals[0]).toMatchObject({ key: 'rent', totalCents: 25000 })
    expect(result.categoryTotals[1]).toMatchObject({ key: 'food', totalCents: 10000 })

    expect(result.spendingTrend).toHaveLength(2)
    expect(result.spendingTrend[0]).toMatchObject({ key: '2026-01', totalCents: 10000 })
    expect(result.spendingTrend[1]).toMatchObject({ key: '2026-02', totalCents: 25000 })
  })

  it('builds member actual-vs-expected with settlement expected share', () => {
    const result = buildDashboardDerivedData({
      expenses: [
        {
          id: 'e-1',
          amount_cents: 18000,
          paid_by_member_id: 'm-1',
          category: 'other',
          created_at: '2026-03-10T00:00:00.000Z',
        },
      ],
      members: [
        { id: 'm-1', name: 'Ana', monthly_salary_cents: 50000 },
        { id: 'm-2', name: 'Luis', monthly_salary_cents: 50000 },
      ],
      settlement: {
        members: [
          { member_id: 'm-1', expected_share: 12000 },
          { member_id: 'm-2', expected_share: 6000 },
        ],
      },
    })

    expect(result.memberActualVsExpected).toEqual([
      {
        memberId: 'm-1',
        memberName: 'Ana',
        actualCents: 18000,
        expectedCents: 12000,
        deltaCents: 6000,
      },
      {
        memberId: 'm-2',
        memberName: 'Luis',
        actualCents: 0,
        expectedCents: 6000,
        deltaCents: -6000,
      },
    ])
  })

  it('builds msi progress based on elapsed months and total installments', () => {
    const now = new Date()
    const twoMonthsAgo = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - 2, 1)).toISOString()

    const result = buildDashboardDerivedData({
      expenses: [
        {
          id: 'e-msi',
          description: 'Laptop',
          expense_type: 'msi',
          total_installments: 6,
          amount_cents: 60000,
          created_at: twoMonthsAgo,
        },
      ],
      members: [],
      settlement: null,
    })

    expect(result.msiProgress).toHaveLength(1)
    expect(result.msiProgress[0].currentInstallment).toBeGreaterThanOrEqual(1)
    expect(result.msiProgress[0].currentInstallment).toBeLessThanOrEqual(6)
    expect(result.msiProgress[0].progressPercent).toBeGreaterThan(0)
  })

  it('calculates totalSpentCents prioritizing settlement total_shared_cents', () => {
    const result = buildDashboardDerivedData({
      expenses: [
        { id: 'e-1', amount_cents: 1000, category: 'food', paid_by_member_id: 'm-1' }
      ],
      members: [
        { id: 'm-1', name: 'Ana', monthly_salary_cents: 50000 }
      ],
      settlement: {
        total_shared_cents: 5500, // 1000 (var) + 2000 (msi) + 2500 (fixed)
      },
    })

    // Current implementation will return 1000 (only variables)
    // Desired implementation should return 5500
    expect(result.totalSpentCents).toBe(5500)
  })

  it('amortizes MSI expenses in category totals', () => {
    const result = buildDashboardDerivedData({
      expenses: [
        {
          id: 'e-msi',
          amount_cents: 6000,
          expense_type: 'msi',
          total_installments: 6,
          category: 'tech',
          category_name: 'Tech',
        },
      ],
      members: [],
      settlement: null,
    })

    // Should contribute 1000 (6000 / 6) instead of 6000
    expect(result.categoryTotals[0]).toMatchObject({
      key: 'tech',
      totalCents: 1000,
    })
  })

  it('maps member balances directly from settlement net_balance_cents', () => {
    const result = buildDashboardDerivedData({
      expenses: [],
      members: [
        { id: 'm-1', name: 'Ana' },
      ],
      settlement: {
        members: [
          {
            member_id: 'm-1',
            expected_share: 5000,
            paid_cents: 2000,
            net_balance_cents: -3000, // Ana owes 3000
          },
        ],
      },
    })

    expect(result.memberActualVsExpected[0]).toMatchObject({
      memberId: 'm-1',
      actualCents: 2000,
      expectedCents: 5000,
      deltaCents: -3000,
    })
  })
})
