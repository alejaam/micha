import { describe, expect, it } from 'vitest'
import { formatRelativeDate } from '../utils'

describe('formatRelativeDate', () => {
  it('returns empty string for falsy input', () => {
    expect(formatRelativeDate(null)).toBe('')
    expect(formatRelativeDate(undefined)).toBe('')
    expect(formatRelativeDate('')).toBe('')
  })

  it('returns "ahora" for diff < 1 minute', () => {
    const now = new Date().toISOString()
    expect(formatRelativeDate(now)).toBe('ahora')
  })

  it('returns "hace Xm" for diff < 60 minutes', () => {
    const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString()
    expect(formatRelativeDate(fiveMinAgo)).toBe('hace 5m')
  })

  it('returns "hace Xh" for diff < 24 hours', () => {
    const threeHoursAgo = new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString()
    expect(formatRelativeDate(threeHoursAgo)).toBe('hace 3h')
  })

  it('returns "ayer" for exactly 1 day ago', () => {
    const yesterday = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
    expect(formatRelativeDate(yesterday)).toBe('ayer')
  })

  it('returns "hace Xd" for diff < 7 days', () => {
    const fourDaysAgo = new Date(Date.now() - 4 * 24 * 60 * 60 * 1000).toISOString()
    expect(formatRelativeDate(fourDaysAgo)).toBe('hace 4d')
  })

  it('returns es-MX short date for diff >= 7 days', () => {
    const tenDaysAgo = new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString()
    const result = formatRelativeDate(tenDaysAgo)
    // Should use es-MX locale, e.g. "abr. 30"
    expect(result).not.toMatch(/ago|just now|hace/)
    expect(typeof result).toBe('string')
    expect(result.length).toBeGreaterThan(0)
  })
})
