import { describe, expect, it } from 'vitest'
import { parseOutboundReplaceInput, replaceAddress } from './outboundReplace'

describe('parseOutboundReplaceInput', () => {
  it('parses IPv4 lines and skips blank/CRLF lines', () => {
    const lines = parseOutboundReplaceInput('192.0.2.66:9873:user-c:pass-c\r\n\r\n192.0.2.61:8080:user-b:pass:b\n')
    expect(lines).toHaveLength(2)
    expect(lines[0]).toMatchObject({ server: '192.0.2.66', port: 9873, username: 'user-c', password: 'pass-c' })
    expect(lines[1]).toMatchObject({ server: '192.0.2.61', port: 8080, username: 'user-b', password: 'pass:b' })
  })

  it('parses bracketed IPv6 and keeps colons in the password', () => {
    const [line] = parseOutboundReplaceInput('[2001:db8::10]:1080:alice:p:a')
    expect(line).toMatchObject({ server: '2001:db8::10', port: 1080, username: 'alice', password: 'p:a' })
  })

  it('allows empty credentials and accepts a two-field line', () => {
    expect(parseOutboundReplaceInput('proxy.example.com:8080::')[0]).toMatchObject({ port: 8080, username: '', password: '' })
    expect(parseOutboundReplaceInput('proxy.example.com:8080')[0]).toMatchObject({ port: 8080, username: '', password: '' })
  })

  it('keeps the position of an invalid line', () => {
    const lines = parseOutboundReplaceInput('host:70000:u:p\n[2001:db8::1:1080:u:p\nvalid:1:u:p')
    expect(lines).toHaveLength(3)
    expect(lines[0].error).toBe('invalidPort')
    expect(lines[1].error).toBe('invalidFormat')
    expect(lines[2]).toMatchObject({ server: 'valid', port: 1, username: 'u', password: 'p' })
  })

  it('rejects non-numeric and zero ports', () => {
    expect(parseOutboundReplaceInput('host:abc:u:p')[0].error).toBe('invalidPort')
    expect(parseOutboundReplaceInput('host:0:u:p')[0].error).toBe('invalidPort')
  })
})

describe('replaceAddress', () => {
  it('formats IPv4, IPv6 and missing addresses', () => {
    expect(replaceAddress('192.0.2.1', 1080)).toBe('192.0.2.1:1080')
    expect(replaceAddress('2001:db8::1', 443)).toBe('[2001:db8::1]:443')
    expect(replaceAddress('', 80)).toBe('—')
  })
})
