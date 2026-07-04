export function formatDateTime(date: string | Date | null | undefined): string {
  if (!date) return '-'
  
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) return '-'
  
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}.${d.getMilliseconds()}`
}

export function formatDate(date: string | Date | null | undefined): string {
  if (!date) return '-'
  
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) return '-'
  
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  
  return `${year}-${month}-${day}`
}

export function formatLatency(ms: number | null | undefined): string {
  if (ms == null) return '0s'
  const seconds = ms / 1000
  return seconds.toFixed(1) + 's'
}

const KILO = 1024
const MEGA = KILO * KILO
const GIGA = MEGA * KILO

export function formatTokens(tokens: number | null | undefined): string {
  if (tokens == null) return '0'
  if (tokens >= GIGA) return (tokens / GIGA).toFixed(1) + 'B'
  if (tokens >= MEGA) return (tokens / MEGA).toFixed(1) + 'M'
  if (tokens >= KILO) return (tokens / KILO).toFixed(1) + 'K'
  return tokens.toString()
}

export function formatToken(value: number | null | undefined): string {
  if (value == null || value === 0) return '0'
  if (value < KILO) return value.toString()
  
  if (value < MEGA) {
    const k = value / KILO
    return k % 1 === 0 ? `${k}K` : `${k.toFixed(1)}K`
  }
  
  const m = value / MEGA
  return m % 1 === 0 ? `${m}M` : `${m.toFixed(1)}M`
}

export function formatContextDisplay(context: number | null | undefined, output: number | null | undefined): string {
  const contextStr = formatToken(context)
  const outputStr = formatToken(output)
  return `${contextStr} / ${outputStr}`
}

export function parseContextString(str: string): number | null {
  if (!str) return null
  const s = str.trim().toLowerCase()
  if (s === '' || s === '-') return null

  let multiplier = 1
  let numStr = s

  const lastChar = s[s.length - 1]
  if (lastChar === 'k') {
    multiplier = KILO
    numStr = s.slice(0, -1)
  } else if (lastChar === 'm') {
    multiplier = MEGA
    numStr = s.slice(0, -1)
  } else if (lastChar === 'b') {
    multiplier = GIGA
    numStr = s.slice(0, -1)
  }

  const num = parseInt(numStr.trim(), 10)
  if (isNaN(num)) return null
  return num * multiplier
}

export function formatContextInput(num: number): string {
  if (num === 0) return '0'
  if (num < KILO) return num.toString()
  if (num < MEGA) {
    const k = num / KILO
    return k % 1 === 0 ? `${k}K` : `${k.toFixed(1)}K`
  }
  const m = num / MEGA
  return m % 1 === 0 ? `${m}M` : `${m.toFixed(1)}M`
}
