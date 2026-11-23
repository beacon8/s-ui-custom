export type SearchableOutbound = Record<string, unknown>

// A query is matched against the tag, server address and port. Multiple
// whitespace-separated terms must all match at least one of those fields.
export function matchesOutboundSearch(outbound: SearchableOutbound, query: string): boolean {
  const terms = query.trim().toLowerCase().split(/\s+/).filter(Boolean)
  if (terms.length === 0) return true
  const fields = [outbound.tag, outbound.server, outbound.server_port]
    .filter((value) => value !== undefined && value !== null)
    .map((value) => String(value).toLowerCase())
  return terms.every((term) => fields.some((field) => field.includes(term)))
}
