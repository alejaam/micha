import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { getSettlement } from '../api'

function toSafeDate(value) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

function monthKeyFromDate(date) {
  return `${date.getUTCFullYear()}-${String(date.getUTCMonth() + 1).padStart(2, '0')}`
}

function parseMonthKey(key) {
  const [year, month] = key.split('-').map(Number)
  return { year, month }
}

function monthLabelFromKey(key) {
  const { year, month } = parseMonthKey(key)
  const date = new Date(Date.UTC(year, month - 1, 1))
  return date.toLocaleDateString('en-US', {
    month: 'short',
    year: '2-digit',
    timeZone: 'UTC',
  })
}

export function buildClosedPeriodsFromExpenses(expenses = []) {
  const now = new Date()
  const currentKey = monthKeyFromDate(now)
  const monthMap = new Map()

  for (const item of Array.isArray(expenses) ? expenses : []) {
    const date = toSafeDate(item?.created_at)
    if (!date) continue

    const key = monthKeyFromDate(date)
    if (key === currentKey) continue

    const current = monthMap.get(key)
    monthMap.set(key, {
      key,
      year: date.getUTCFullYear(),
      month: date.getUTCMonth() + 1,
      label: monthLabelFromKey(key),
      totalCents: (current?.totalCents ?? 0) + (item?.amount_cents ?? 0),
      expenseCount: (current?.expenseCount ?? 0) + 1,
      source: 'derived',
      isClosed: true,
    })
  }

  return Array.from(monthMap.values()).sort((a, b) => b.key.localeCompare(a.key))
}

function buildMockPeriodSnapshot(period, memberCount) {
  const baseline = Math.max(1, Math.round(period.totalCents || 0))
  const split = Math.round(baseline / Math.max(1, memberCount || 2))
  return {
    source: 'mock',
    totalSharedCents: baseline,
    memberBalances: [
      {
        memberId: 'member-a',
        memberName: 'Member A',
        netBalanceCents: split,
      },
      {
        memberId: 'member-b',
        memberName: 'Member B',
        netBalanceCents: -split,
      },
    ],
  }
}

function normalizeSettlementSnapshot(settlement) {
  if (!settlement || !Array.isArray(settlement.members)) return null

  return {
    source: 'api',
    totalSharedCents: settlement.total_shared_cents ?? 0,
    memberBalances: settlement.members.map((member) => ({
      memberId: member.member_id,
      memberName: member.member_name || member.member_id,
      netBalanceCents: member.net_balance_cents ?? 0,
    })),
  }
}

/**
 * Historical periods hook with local caching and provisional fallback values.
 */
export function useHistoricalPeriods({
  householdId,
  expenses = [],
  members = [],
}) {
  const [selectedPeriodKey, setSelectedPeriodKey] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isPeriodLoading, setIsPeriodLoading] = useState(false)
  const [provisionalReason, setProvisionalReason] = useState('')
  const [cacheVersion, setCacheVersion] = useState(0)
  const settlementCacheRef = useRef(new Map())

  const closedPeriods = useMemo(
    () => buildClosedPeriodsFromExpenses(expenses),
    [expenses],
  )

  useEffect(() => {
    if (closedPeriods.length === 0) {
      setSelectedPeriodKey('')
      return
    }

    setSelectedPeriodKey((current) => {
      if (current && closedPeriods.some((period) => period.key === current)) {
        return current
      }
      return closedPeriods[0].key
    })
  }, [closedPeriods])

  const selectedPeriod = useMemo(
    () => closedPeriods.find((period) => period.key === selectedPeriodKey) || null,
    [closedPeriods, selectedPeriodKey],
  )

  const ensurePeriodSnapshot = useCallback(async (period) => {
    if (!householdId?.trim() || !period) return null

    const cached = settlementCacheRef.current.get(period.key)
    if (cached) return cached

    setIsPeriodLoading(true)
    try {
      const data = await getSettlement({
        householdId: householdId.trim(),
        year: period.year,
        month: period.month,
      })

      const normalized = normalizeSettlementSnapshot(data)
      if (normalized) {
        settlementCacheRef.current.set(period.key, normalized)
        setCacheVersion((value) => value + 1)
        return normalized
      }

      const fallback = buildMockPeriodSnapshot(period, members.length)
      settlementCacheRef.current.set(period.key, fallback)
      setCacheVersion((value) => value + 1)
      setProvisionalReason('Historical settlement data incomplete — showing provisional values.')
      return fallback
    } catch {
      const fallback = buildMockPeriodSnapshot(period, members.length)
      settlementCacheRef.current.set(period.key, fallback)
      setCacheVersion((value) => value + 1)
      setProvisionalReason('Historical endpoint unavailable — showing provisional values.')
      return fallback
    } finally {
      setIsPeriodLoading(false)
    }
  }, [householdId, members.length])

  useEffect(() => {
    let cancelled = false

    async function loadSnapshots() {
      if (!householdId?.trim()) {
        setIsLoading(false)
        return
      }
      if (closedPeriods.length === 0) {
        setIsLoading(false)
        return
      }

      setIsLoading(true)
      setProvisionalReason('')
      for (const period of closedPeriods) {
        if (cancelled) return
        // eslint-disable-next-line no-await-in-loop
        await ensurePeriodSnapshot(period)
      }
      if (!cancelled) setIsLoading(false)
    }

    loadSnapshots()

    return () => {
      cancelled = true
    }
  }, [householdId, closedPeriods, ensurePeriodSnapshot])

  const comparisonSeries = useMemo(
    () => [...closedPeriods]
      .sort((a, b) => a.key.localeCompare(b.key))
      .map((period) => ({
        key: period.key,
        label: period.label,
        totalCents: period.totalCents,
        source: period.source,
      })),
    [closedPeriods],
  )

  const memberBalanceTrend = useMemo(() => {
    const monthRows = [...closedPeriods].sort((a, b) => a.key.localeCompare(b.key))
    const memberKeys = new Set()

    for (const period of monthRows) {
      const snapshot = settlementCacheRef.current.get(period.key)
      for (const member of snapshot?.memberBalances ?? []) {
        memberKeys.add(member.memberName)
      }
    }

    return monthRows.map((period) => {
      const row = {
        key: period.key,
        label: period.label,
        source: settlementCacheRef.current.get(period.key)?.source ?? 'derived',
      }
      const snapshot = settlementCacheRef.current.get(period.key)
      for (const name of memberKeys) {
        const match = snapshot?.memberBalances?.find((member) => member.memberName === name)
        row[name] = match?.netBalanceCents ?? 0
      }
      return row
    })
  }, [cacheVersion, closedPeriods])

  const selectedPeriodSnapshot = useMemo(
    () => (selectedPeriod ? settlementCacheRef.current.get(selectedPeriod.key) ?? null : null),
    [cacheVersion, selectedPeriod],
  )

  const completedMsi = useMemo(() => {
    return (Array.isArray(expenses) ? expenses : [])
      .filter((item) => item?.expense_type === 'msi' && Number(item?.total_installments) > 0)
      .map((item) => {
        const createdAt = toSafeDate(item.created_at)
        const now = new Date()
        const elapsedMonths = createdAt
          ? (now.getUTCFullYear() - createdAt.getUTCFullYear()) * 12 + (now.getUTCMonth() - createdAt.getUTCMonth()) + 1
          : 0
        const totalInstallments = Number(item.total_installments)
        const currentInstallment = Math.max(0, Math.min(totalInstallments, elapsedMonths))
        return {
          id: item.id,
          description: item.description,
          totalInstallments,
          currentInstallment,
          isCompleted: currentInstallment >= totalInstallments,
        }
      })
      .filter((item) => item.isCompleted)
      .slice(0, 8)
  }, [expenses])

  return {
    historicalData: closedPeriods,
    closedPeriods,
    selectedPeriodKey,
    setSelectedPeriodKey,
    selectedPeriod,
    selectedPeriodSnapshot,
    comparisonSeries,
    memberBalanceTrend,
    completedMsi,
    isLoading,
    isPeriodLoading,
    isProvisional: provisionalReason !== '',
    provisionalReason,
  }
}
