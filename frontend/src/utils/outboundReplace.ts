export type ReplaceLineError = 'invalidFormat' | 'invalidServer' | 'invalidPort'

export interface ReplaceLine {
  line: number
  raw: string
  server: string
  port: number
  username: string
  password: string
  error?: ReplaceLineError
}

// Parse the replacement textarea. Blank lines are skipped; every non-blank line
// keeps its position so an invalid line never shifts later lines onto the wrong
// outbound. Password is the remainder after the third separator, so it may
// contain colons. IPv6 must be bracketed, e.g. [2001:db8::10]:1080:user:pass.
export const parseOutboundReplaceInput = (text: string): ReplaceLine[] => {
  const result: ReplaceLine[] = []
  const rawLines = text.split(/\r?\n/)
  for (let i = 0; i < rawLines.length; i++) {
    const raw = rawLines[i].trim()
    if (!raw) continue
    result.push(parseLine(raw, i + 1))
  }
  return result
}

const parseLine = (raw: string, line: number): ReplaceLine => {
  let server = ''
  let rest = ''
  if (raw.startsWith('[')) {
    const close = raw.indexOf(']')
    if (close <= 1 || raw[close + 1] !== ':') {
      return bad(line, raw, 'invalidFormat')
    }
    server = raw.slice(1, close)
    rest = raw.slice(close + 2)
  } else {
    const first = raw.indexOf(':')
    if (first <= 0) return bad(line, raw, 'invalidFormat')
    server = raw.slice(0, first)
    rest = raw.slice(first + 1)
  }
  if (!server) return bad(line, raw, 'invalidServer')

  const parts = rest.split(':')
  const portText = parts.shift() ?? ''
  if (!/^\d+$/.test(portText)) return bad(line, raw, 'invalidPort', server)
  const port = Number(portText)
  if (!Number.isSafeInteger(port) || port < 1 || port > 65535) {
    return bad(line, raw, 'invalidPort', server)
  }
  const username = parts.shift() ?? ''
  const password = parts.join(':')
  return { line, raw, server, port, username, password }
}

const bad = (
  line: number,
  raw: string,
  error: ReplaceLineError,
  server = '',
): ReplaceLine => ({ line, raw, server, port: 0, username: '', password: '', error })

export const replaceAddress = (server: string, port?: number): string => {
  if (!server) return '—'
  const host = server.includes(':') ? `[${server}]` : server
  return port ? `${host}:${port}` : host
}
