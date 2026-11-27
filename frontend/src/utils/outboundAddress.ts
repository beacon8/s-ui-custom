export type OutboundAddressStatus = 'ok' | 'direct' | 'multi' | 'none' | 'split' | 'unbound'

// The configured server of the outbound a client is routed to. This is read
// from the current config, never measured, so it can render immediately.
export interface OutboundAddressEntry {
  status: OutboundAddressStatus
  outbound?: string
  address?: string
  port?: number
  type?: string
}

export interface OutboundAddressResponse {
  users: Record<string, OutboundAddressEntry>
}

export const outboundAddressText = (entry?: OutboundAddressEntry | null): string =>
  entry?.status === 'ok' ? entry.address ?? '' : ''

export const outboundAddressSortValue = (entry?: OutboundAddressEntry | null): string =>
  outboundAddressText(entry)

export const outboundAddressWithPort = (entry?: OutboundAddressEntry | null): string => {
  if (!entry || entry.status !== 'ok' || !entry.address) return ''
  if (!entry.port) return entry.address
  return entry.address.includes(':') ? `[${entry.address}]:${entry.port}` : `${entry.address}:${entry.port}`
}
