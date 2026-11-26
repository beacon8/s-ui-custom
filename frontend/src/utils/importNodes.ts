// One pasted line, ready to become the node object the API accepts. Blank
// lines are dropped but counted, so an error can still say "line 7" for what
// the operator sees in the textarea.
export interface ImportNodeLine {
  raw?: string
  uri?: string
  line: number
}

export function classifyNodeLines(text: string): ImportNodeLine[] {
  const lines: ImportNodeLine[] = []
  text.split('\n').forEach((value, index) => {
    const line = value.trim()
    if (line === '') return
    if (line.includes('://')) lines.push({ uri: line, line: index + 1 })
    else lines.push({ raw: line, line: index + 1 })
  })
  return lines
}

// The group shown before anything is submitted. It is the same natural-number
// interval the backend writes (US-2-11 for ten members starting at 2).
export function importGroupName(prefix: string, startId: number, count: number): string {
  if (prefix.trim() === '' || count <= 0) return ''
  return `${prefix.trim()}-${startId}-${startId + count - 1}`
}
