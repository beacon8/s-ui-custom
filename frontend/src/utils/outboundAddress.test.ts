import { describe, expect, it } from 'vitest'
import { outboundAddressSortValue, outboundAddressText, outboundAddressWithPort, type OutboundAddressEntry } from './outboundAddress'

const entry = (overrides: Partial<OutboundAddressEntry>): OutboundAddressEntry => ({
  status: 'ok',
  address: '1.1.1.1',
  port: 80,
  ...overrides,
})

describe('outboundAddressText', () => {
  it('shows only the configured address for ok entries', () => {
    expect(outboundAddressText(entry({}))).toBe('1.1.1.1')
    expect(outboundAddressText(entry({ status: 'direct', address: undefined }))).toBe('')
    expect(outboundAddressText(entry({ status: 'multi' }))).toBe('')
    expect(outboundAddressText(undefined)).toBe('')
  })
})

describe('outboundAddressWithPort', () => {
  it('formats IPv4, IPv6 and port-less addresses for the tooltip', () => {
    expect(outboundAddressWithPort(entry({}))).toBe('1.1.1.1:80')
    expect(outboundAddressWithPort(entry({ address: '2001:db8::1', port: 443 }))).toBe('[2001:db8::1]:443')
    expect(outboundAddressWithPort(entry({ port: undefined }))).toBe('1.1.1.1')
    expect(outboundAddressWithPort(entry({ address: 'proxy.example.com' }))).toBe('proxy.example.com:80')
    expect(outboundAddressWithPort(entry({ status: 'direct' }))).toBe('')
  })
})

describe('outboundAddressSortValue', () => {
  it('sorts by the configured address', () => {
    expect(outboundAddressSortValue(entry({}))).toBe('1.1.1.1')
    expect(outboundAddressSortValue(undefined)).toBe('')
  })
})
