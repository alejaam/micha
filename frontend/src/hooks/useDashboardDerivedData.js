import { useMemo } from 'react'

function toSafeDate(value) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

function monthDiffInclusive(from, to) {
  return (to.getUTCFullYear() - from.getUTCFullYear()) * 12 + (to.getUTCMonth() - from.getUTCMonth()) + 1
}

function formatMonthKeyLabel(monthKey) {
  const [year, month] = monthKey.split('-').map(Number)
  const date = new Date(Date.UTC(year, month - 1, 1))

  return date.toLocaleDateString('en-US', {
    month: 'short',
    year: '2-digit',
    timeZone: 'UTC',
  })
}

export function buildDashboardDerivedData({ expenses = [], members = [], settlement = null, recurringItems = [] }) {
  const safeExpenses = Array.isArray(expenses) ? expenses : []
  const safeMembers = Array.isArray(members) ? members : []
  const safeRecurring = Array.isArray(recurringItems) ? recurringItems : []

  const fixedTotalCents = safeRecurring
    .filter((item) => item?.is_active && item?.expense_type === 'fixed')
    .reduce((sum, item) => sum + (item?.amount_cents ?? 0), 0)

  const totalSpentCents = settlement?.total_shared_cents ?? (
    safeExpenses.reduce((sum, item) => {
      let amountToAdd = item?.amount_cents ?? 0
      if (item?.expense_type === 'msi' && Number(item?.total_installments) > 0) {
        amountToAdd = Math.round(amountToAdd / Number(item.total_installments))
      }
      return sum + amountToAdd
    }, 0) + fixedTotalCents
  )

  const categoryMap = new Map()
  
  // Add fixed items to category map first
  for (const item of safeRecurring) {
    if (!item?.is_active || item?.expense_type !== 'fixed') continue
    const key = item?.category_slug || 'fixed'
    const label = item?.category_name || 'Fixed'
    const current = categoryMap.get(key)
    categoryMap.set(key, {
      key,
      label,
      totalCents: (current?.totalCents ?? 0) + (item?.amount_cents ?? 0),
    })
  }

  for (const item of safeExpenses) {
    const key = item?.category || item?.category_slug || item?.category_name || 'other'
    const label = item?.category_name || key
    const current = categoryMap.get(key)
    
    let amountToAdd = item?.amount_cents ?? 0
    if (item?.expense_type === 'msi' && Number(item?.total_installments) > 0) {
      amountToAdd = Math.round(amountToAdd / Number(item.total_installments))
    }
    
    const nextTotal = (current?.totalCents ?? 0) + amountToAdd

    categoryMap.set(key, {
      key,
      label,
      totalCents: nextTotal,
    })
  }

  const categoryTotals = Array.from(categoryMap.values())
    .sort((a, b) => b.totalCents - a.totalCents)
    .map((entry) => ({
      ...entry,
      percentage: totalSpentCents > 0 ? (entry.totalCents / totalSpentCents) * 100 : 0,
    }))

  const settlementMembers = Array.isArray(settlement?.members) ? settlement.members : []
  const settlementMemberMap = new Map(settlementMembers.map((member) => [member.member_id, member]))

  const householdIncomeCents = safeMembers.reduce((sum, member) => sum + (member?.monthly_salary_cents ?? 0), 0)
  const fallbackExpectedPerMember = safeMembers.length > 0
    ? Math.round(totalSpentCents / safeMembers.length)
    : 0

  const actualByMember = new Map()
  for (const item of safeExpenses) {
    const memberId = item?.paid_by_member_id
    if (!memberId) continue
    actualByMember.set(memberId, (actualByMember.get(memberId) ?? 0) + (item?.amount_cents ?? 0))
  }

  const memberActualVsExpected = safeMembers.map((member) => {
    const memberId = member.id
    const settlementMember = settlementMemberMap.get(memberId)

    const actualCents = settlementMember?.paid_cents ?? (actualByMember.get(memberId) ?? 0)
    let expectedCents = settlementMember?.expected_share ?? null

    if (expectedCents == null) {
      const salaryCents = member?.monthly_salary_cents ?? 0
      const weightedExpected = householdIncomeCents > 0
        ? Math.round(totalSpentCents * (salaryCents / householdIncomeCents))
        : fallbackExpectedPerMember
      expectedCents = weightedExpected
    }

    const deltaCents = settlementMember?.net_balance_cents ?? (actualCents - expectedCents)

    return {
      memberId,
      memberName: member?.name ?? memberId,
      actualCents,
      expectedCents,
      deltaCents,
    }
  })

  const now = new Date()
  const msiProgress = safeExpenses
    .filter((item) => item?.expense_type === 'msi' && Number(item?.total_installments) > 0)
    .map((item) => {
      const startDate = toSafeDate(item?.created_at) ?? now
      const totalInstallments = Number(item?.total_installments) || 1
      const currentInstallment = Math.max(1, Math.min(totalInstallments, monthDiffInclusive(startDate, now)))
      const progressPercent = Math.round((currentInstallment / totalInstallments) * 100)

      return {
        id: item.id,
        description: item.description,
        totalInstallments,
        currentInstallment,
        progressPercent,
        remainingInstallments: Math.max(0, totalInstallments - currentInstallment),
      }
    })
    .sort((a, b) => b.progressPercent - a.progressPercent)

  const monthlyTotalsMap = new Map()
  for (const item of safeExpenses) {
    const date = toSafeDate(item?.created_at)
    if (!date) continue

    const monthKey = `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, '0')}`
    monthlyTotalsMap.set(monthKey, (monthlyTotalsMap.get(monthKey) ?? 0) + (item?.amount_cents ?? 0))
  }

  const spendingTrend = Array.from(monthlyTotalsMap.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, totalCents]) => ({
      key,
      label: formatMonthKeyLabel(key),
      totalCents,
    }))

  return {
    totalSpentCents,
    fixedTotalCents,
    categoryTotals,
    memberActualVsExpected,
    msiProgress,
    spendingTrend,
  }
}

/**
 * Hook for deriving dashboard charts and summary data from raw expenses and members.
 * This acts as the UI composition layer described in the technical design.
 */
export function useDashboardDerivedData({ expenses, members, settlement, recurringItems }) {
  return useMemo(
    () => buildDashboardDerivedData({ expenses, members, settlement, recurringItems }),
    [expenses, members, settlement, recurringItems],
  )
}
