// Panel timestamps are Unix seconds. The client table shows an explicit,
// locale-independent YYYY-MM-DD HH:mm:ss so two dates align in a narrow cell.
const pad2 = (value: number): string => String(value).padStart(2, '0')

export function formatDateTime(unixSeconds: number | null | undefined): string {
  if (!unixSeconds || unixSeconds <= 0) return ''
  const date = new Date(unixSeconds * 1000)
  if (Number.isNaN(date.getTime())) return ''
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(date.getSeconds())}`
}
