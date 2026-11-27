import { describe, expect, it } from 'vitest'
import { formatDateTime } from './formatDate'

describe('formatDateTime', () => {
  it('renders Unix seconds as YYYY-MM-DD HH:mm:ss', () => {
    const local = new Date(2026, 10, 28, 4, 5, 6)
    expect(formatDateTime(local.getTime() / 1000)).toBe('2026-11-28 04:05:06')
    const padded = new Date(2026, 0, 2, 3, 4, 5)
    expect(formatDateTime(padded.getTime() / 1000)).toBe('2026-01-02 03:04:05')
  })

  it('renders missing or zero timestamps as an empty string', () => {
    expect(formatDateTime(0)).toBe('')
    expect(formatDateTime(null)).toBe('')
    expect(formatDateTime(undefined)).toBe('')
  })
})
