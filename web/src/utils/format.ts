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

export function formatTokens(tokens: number | null | undefined): string {
  if (tokens == null) return '0'
  if (tokens >= 1e9) return (tokens / 1e9).toFixed(1) + 'B'
  if (tokens >= 1e6) return (tokens / 1e6).toFixed(1) + 'M'
  if (tokens >= 1e3) return (tokens / 1e3).toFixed(1) + 'K'
  return tokens.toString()
}

export function formatToken(value: number | null | undefined): string {
  if (value == null || value === 0) return '0'
  if (value < 1000) return value.toString()
  
  if (value < 1000000) {
    const k = value / 1000
    return k % 1 === 0 ? `${k}K` : `${k.toFixed(1)}K`
  }
  
  const m = value / 1000000
  return m % 1 === 0 ? `${m}M` : `${m.toFixed(1)}M`
}

export function formatContextDisplay(context: number | null | undefined, output: number | null | undefined): string {
  const contextStr = formatToken(context)
  const outputStr = formatToken(output)
  return `${contextStr} / ${outputStr}`
}
