import { describe, expect, it } from 'vitest'
import { compile } from '@intlify/core-base'
import en from './en'
import fa from './fa'
import ru from './ru'
import vi from './vi'
import zhcn from './zhcn'
import zhtw from './zhtw'

// Vue I18n treats @, {, } and | as message syntax. A literal '@' in a
// placeholder (a proxy URL, for example) makes the message fail to compile at
// render time, which unmounts the form that used it. Compiling every string
// here turns that into a test failure instead of a blank panel.
const locales: Record<string, unknown> = { en, fa, ru, vi, zhHans: zhcn, zhHant: zhtw }

function strings(value: unknown, path: string, out: [string, string][]) {
  if (typeof value === 'string') {
    out.push([path, value])
  } else if (value && typeof value === 'object') {
    for (const [key, child] of Object.entries(value)) {
      strings(child, path === '' ? key : `${path}.${key}`, out)
    }
  }
  return out
}

describe('locale messages', () => {
  it('all compile', () => {
    const failures: string[] = []
    for (const [locale, messages] of Object.entries(locales)) {
      for (const [key, message] of strings(messages, '', [])) {
        try {
          compile(message, { warnHtmlMessage: false, key, locale })
        } catch (error) {
          failures.push(`${locale}:${key}: ${String(error)}`)
        }
      }
    }
    expect(failures).toEqual([])
  })
})
