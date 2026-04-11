import { renderHook, act } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import {
  buildConsensusState,
  buildRibbonState,
  useDashboardUxState,
} from '../useDashboardUxState'

describe('useDashboardUxState', () => {
  it('normalizes unknown initial status to open and keeps mutations unlocked', () => {
    const { result } = renderHook(() => useDashboardUxState('unexpected'))

    expect(result.current.periodStatus).toBe('open')
    expect(result.current.isMutationLocked).toBe(false)
  })

  it('locks mutations in review and closed states', () => {
    const { result } = renderHook(() => useDashboardUxState('open'))

    act(() => {
      result.current.setPeriodStatus('review')
    })
    expect(result.current.isMutationLocked).toBe(true)

    act(() => {
      result.current.setPeriodStatus('closed')
    })
    expect(result.current.isMutationLocked).toBe(true)
  })

  it('handles bottom sheet open and close transitions', () => {
    const { result } = renderHook(() => useDashboardUxState('open'))

    expect(result.current.isBottomSheetOpen).toBe(false)
    act(() => {
      result.current.openBottomSheet()
    })
    expect(result.current.isBottomSheetOpen).toBe(true)

    act(() => {
      result.current.closeBottomSheet()
    })
    expect(result.current.isBottomSheetOpen).toBe(false)
  })

  it('exposes a default mock consensus state', () => {
    const { result } = renderHook(() => useDashboardUxState('open'))
    expect(result.current.consensus).toEqual({
      approved: 0,
      total: 0,
      percent: 0,
      source: 'mock',
    })
  })
})

describe('buildRibbonState', () => {
  it('returns review content for review status', () => {
    expect(buildRibbonState('review')).toEqual({
      status: 'review',
      stateLabel: '[REVIEW]',
      description: 'Period under review — mutating actions are temporarily locked.',
    })
  })

  it('falls back to open for unknown status', () => {
    expect(buildRibbonState('invalid').status).toBe('open')
  })
})

describe('buildConsensusState', () => {
  it('builds normalized percent and clamps approved between 0 and total', () => {
    expect(buildConsensusState({ approved: 4, total: 3, source: 'mock' })).toEqual({
      approved: 3,
      total: 3,
      percent: 100,
      source: 'mock',
    })

    expect(buildConsensusState({ approved: -2, total: 10 })).toEqual({
      approved: 0,
      total: 10,
      percent: 0,
      source: 'derived',
    })
  })
})
