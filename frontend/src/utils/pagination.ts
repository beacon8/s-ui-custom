// Small, framework-free pagination rules shared by table pages. Keeping the
// clamp here makes the "delete the last row on the last page" behavior
// independently testable without mounting a table.
export function pageCount(total: number, pageSize: number): number {
  return Math.max(1, Math.ceil(total / Math.max(1, pageSize)))
}

export function clampPage(current: number, total: number, pageSize: number): number {
  const size = Math.max(1, pageSize)
  if (total <= size) return 1
  return Math.min(Math.max(1, current), pageCount(total, size))
}
