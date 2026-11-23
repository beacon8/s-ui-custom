import { describe, expect, it } from 'vitest'
import { matchesOutboundSearch } from './outboundSearch'

describe('outbound search', () => {
  const outbound = { tag: 'US-Node', server: 'example.com', server_port: 443 }

  it('matches an empty query', () => {
    expect(matchesOutboundSearch(outbound, '  ')).toBe(true)
  })

  it('matches tag, server and port case-insensitively', () => {
    expect(matchesOutboundSearch(outbound, 'us-node')).toBe(true)
    expect(matchesOutboundSearch(outbound, 'EXAMPLE')).toBe(true)
    expect(matchesOutboundSearch(outbound, '443')).toBe(true)
  })

  it('requires every whitespace-separated term to match a field', () => {
    expect(matchesOutboundSearch(outbound, 'us 443')).toBe(true)
    expect(matchesOutboundSearch(outbound, 'us 8443')).toBe(false)
  })

  it('does not match an unrelated value', () => {
    expect(matchesOutboundSearch(outbound, 'hysteria')).toBe(false)
  })
})
