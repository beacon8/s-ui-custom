import { describe, expect, it } from 'vitest'
import { clampPage, pageCount } from './pagination'

describe('table pagination', () => {
  const outbounds = Array.from({ length: 21 }, (_, i) => `outbound-${i + 1}`)

  it('splits 21 rows into 10 / 10 / 1', () => {
    expect(pageCount(outbounds.length, 10)).toBe(3)
    expect(outbounds.slice(0, 10)).toHaveLength(10)
    expect(outbounds.slice(10, 20)).toHaveLength(10)
    expect(outbounds.slice(20, 30)).toHaveLength(1)
  })

  it('returns to the previous page after deleting the last row on the last page', () => {
    expect(clampPage(3, 20, 10)).toBe(2)
  })

  it('keeps the current page for online/delay refreshes', () => {
    // A refresh only changes row contents, not the row count or page size.
    expect(clampPage(2, 21, 10)).toBe(2)
  })

  it('clamps page sizes and page numbers defensively', () => {
    expect(pageCount(0, 10)).toBe(1)
    expect(clampPage(5, 8, 10)).toBe(1)
    expect(clampPage(0, 21, 10)).toBe(1)
    expect(clampPage(9, 21, 10)).toBe(3)
  })
})
