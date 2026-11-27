// Formats one direction of a network speed. Speeds use decimal units:
// 1 Mbps = 1000 kbps. A non-zero value keeps exactly two decimals; the unit
// switches to Mbps once the value reaches 1000 kbps. Zero stays "0 kbps" and
// missing data renders an em dash.
export function formatSpeed(bitsPerSecond: number | null | undefined): string {
  if (bitsPerSecond == null || !Number.isFinite(bitsPerSecond)) {
    return '—'
  }
  const bits = Math.max(0, bitsPerSecond)
  if (bits === 0) {
    return '0 kbps'
  }
  const kbps = bits / 1000
  // 999.995 kbps rounds to 1000.00, so switch at that point rather than print
  // an out-of-range kbps value.
  if (kbps >= 999.995) {
    return `${(bits / 1_000_000).toFixed(2)} Mbps`
  }
  return `${kbps.toFixed(2)} kbps`
}

// The compact single-line label used by the client table. Both directions are
// sampled together, so one missing side means the row has no valid sample yet.
export function formatSpeedPair(
  upBitsPerSecond: number | null | undefined,
  downBitsPerSecond: number | null | undefined,
): string {
  if (upBitsPerSecond == null || downBitsPerSecond == null) {
    return '—'
  }
  // Download is shown first; the arrows make the direction explicit.
  return `↓ ${formatSpeed(downBitsPerSecond)} / ↑ ${formatSpeed(upBitsPerSecond)}`
}

// Converts a byte delta and the actual interval between two samples into
// bit/s. A non-positive interval means the samples are unusable.
export function speedFromBytes(deltaBytes: number, seconds: number): number | null {
  if (!Number.isFinite(deltaBytes) || !Number.isFinite(seconds) || seconds <= 0) {
    return null
  }
  return Math.max(0, deltaBytes) * 8 / seconds
}
