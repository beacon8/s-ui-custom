import { describe, expect, it } from 'vitest'
import { classifyNodeLines, importGroupName } from './importNodes'

describe('classifyNodeLines', () => {
  it('keeps original line numbers across blank lines', () => {
    const lines = classifyNodeLines('192.0.2.1:1080:u:p\n\nsocks5://u:p@192.0.2.2:1080\n')
    expect(lines).toEqual([
      { raw: '192.0.2.1:1080:u:p', line: 1 },
      { uri: 'socks5://u:p@192.0.2.2:1080', line: 3 },
    ])
  })

  it('trims surrounding whitespace but keeps the rest of the line', () => {
    expect(classifyNodeLines('  192.0.2.1:1080:u:p @ x  ')[0].raw).toBe('192.0.2.1:1080:u:p @ x')
  })

  it('returns nothing for an empty paste', () => {
    expect(classifyNodeLines('\n  \n')).toEqual([])
  })
})

describe('importGroupName', () => {
  it('matches the documented intervals', () => {
    expect(importGroupName('US', 1, 1)).toBe('US-1-1')
    expect(importGroupName('US', 1, 10)).toBe('US-1-10')
    expect(importGroupName('US', 2, 10)).toBe('US-2-11')
  })

  it('is empty without a prefix or members', () => {
    expect(importGroupName('  ', 1, 3)).toBe('')
    expect(importGroupName('US', 1, 0)).toBe('')
  })
})
