export interface SortConfig {
  prop: string
  order: 'ascending' | 'descending'
}

export function getSortConfig(key: string, defaultProp = 'name'): SortConfig {
  const stored = localStorage.getItem(`table_sort_${key}`)
  if (stored) {
    try {
      return JSON.parse(stored)
    } catch {
      return { prop: defaultProp, order: 'ascending' }
    }
  }
  return { prop: defaultProp, order: 'ascending' }
}

export function setSortConfig(key: string, config: SortConfig) {
  localStorage.setItem(`table_sort_${key}`, JSON.stringify(config))
}

export function sortByDate(a: any, b: any, prop: string): number {
  const valA = a[prop]
  const valB = b[prop]
  
  if (!valA && !valB) return 0
  if (!valA) return 1
  if (!valB) return -1
  
  return new Date(valA).getTime() - new Date(valB).getTime()
}

export function sortByArrayLength(a: any, b: any, prop: string): number {
  const lenA = a[prop]?.length || 0
  const lenB = b[prop]?.length || 0
  return lenA - lenB
}
