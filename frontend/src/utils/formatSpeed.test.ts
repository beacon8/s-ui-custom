import { describe, expect, it } from 'vitest'
import { formatSpeed, formatSpeedPair, speedFromBytes } from './formatSpeed'

describe('formatSpeed', () => {
  it('keeps two decimals for non-zero values and switches at 1000 kbps', () => {
    expect(formatSpeed(800_000)).toBe('800.00 kbps')
    expect(formatSpeed(1_000_000)).toBe('1.00 Mbps')
    expect(formatSpeed(12_500_000)).toBe('12.50 Mbps')
    expect(formatSpeed(1_230)).toBe('1.23 kbps')
    expect(formatSpeed(1_200)).toBe('1.20 kbps')
  })

  it('renders zero without decimals and missing data as an em dash', () => {
    expect(formatSpeed(0)).toBe('0 kbps')
    expect(formatSpeed(null)).toBe('—')
    expect(formatSpeed(undefined)).toBe('—')
    expect(formatSpeed(Number.NaN)).toBe('—')
  })

  it('does not print 1000.00 kbps at the boundary', () => {
    expect(formatSpeed(999_999)).toBe('1.00 Mbps')
  })
})

describe('formatSpeedPair', () => {
  it('renders both directions on one line with an arrow and separator', () => {
    expect(formatSpeedPair(1_250_000, 318_070)).toBe('↓ 318.07 kbps / ↑ 1.25 Mbps')
    expect(formatSpeedPair(1_000_000, 100_000)).toBe('↓ 100.00 kbps / ↑ 1.00 Mbps')
    expect(formatSpeedPair(0, 0)).toBe('↓ 0 kbps / ↑ 0 kbps')
  })

  it('waits for the first sample instead of showing a partial pair', () => {
    expect(formatSpeedPair(null, null)).toBe('—')
    expect(formatSpeedPair(1_000_000, null)).toBe('—')
  })
})

describe('speedFromBytes', () => {
  it('matches the fixed acceptance examples', () => {
    expect(formatSpeedPair(speedFromBytes(250_000, 2), speedFromBytes(25_000, 2)))
      .toBe('↓ 100.00 kbps / ↑ 1.00 Mbps')
    expect(formatSpeedPair(speedFromBytes(250_000, 4), speedFromBytes(25_000, 4)))
      .toBe('↓ 50.00 kbps / ↑ 500.00 kbps')
  })

  it('returns null for an unusable interval', () => {
    expect(speedFromBytes(1_000, 0)).toBeNull()
    expect(speedFromBytes(1_000, -1)).toBeNull()
    expect(speedFromBytes(Number.NaN, 2)).toBeNull()
  })
})
