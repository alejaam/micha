import { renderHook, act } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { useDashboardUxState } from '../useDashboardUxState'

describe('useDashboardUxState', () => {
  it('normalizes unknown initial status to open', () => {
    const { result } = renderHook(() => useDashboardUxState('unexpected'))

    expect(result.current.periodStatus).toBe('open')
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
})
