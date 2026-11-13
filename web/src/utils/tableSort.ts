export interface SortConfig {
  prop: string
  order: 'ascending' | 'descending'
}

// 中文（汉字）字符检测：CJK 统一表意文字 + 扩展A + 兼容表意文字
const CJK_RE = /[㐀-䶿一-鿿豈-﫿]/

// 名称比较（table-sorting 规格：字母在前 + 中文按拼音）：
// 非中文块在前按 en 字母序，中文块在后按 zh-CN 拼音序。
// 用于状态列第二排序因子与展示顺序统一（localeCompare 不带 locale 会
// 跟随系统 locale 把整块中文排到字母之前，故此处显式分块 + 指定 locale）。
export function compareName(a: string, b: string): number {
  const aCjk = CJK_RE.test(a)
  const bCjk = CJK_RE.test(b)
  if (aCjk !== bCjk) return aCjk ? 1 : -1
  return aCjk ? a.localeCompare(b, 'zh-CN') : a.localeCompare(b, 'en')
}

// 读取已保存的排序偏好。以下情况返回 undefined（调用方不绑定 default-sort，
// 即无排序指示器，列表按数据源预排序展示默认顺序）：
// - 无偏好 / 数据损坏
// - 偏好与该页硬默认相同（等价于"没有一个选择"；同时兼容旧版取消逻辑写入的 fallback）
export function getSortConfig(key: string, defaultProp?: string): SortConfig | undefined {
  const stored = localStorage.getItem(`table_sort_${key}`)
  if (!stored) return undefined
  try {
    const parsed = JSON.parse(stored)
    if (parsed && typeof parsed.prop === 'string' && (parsed.order === 'ascending' || parsed.order === 'descending')) {
      if (defaultProp && parsed.prop === defaultProp && parsed.order === 'ascending') return undefined
      return parsed
    }
  } catch {
    // 损坏数据视为无偏好
  }
  return undefined
}

export function setSortConfig(key: string, config: SortConfig) {
  localStorage.setItem(`table_sort_${key}`, JSON.stringify(config))
}

// 取消排序时清除偏好：下次访问恢复默认顺序（无排序指示器）
export function removeSortConfig(key: string) {
  localStorage.removeItem(`table_sort_${key}`)
}

// 数据源预排序：比较方式与 Element Plus 列的内置比较完全一致（原始值 < > 比较），
// 无排序偏好时列表即按默认顺序展示，点击该列升序不会引起重排。
export function sortByPropAsc<T extends Record<string, any>>(list: T[], prop = 'name'): T[] {
  return [...list].sort((a, b) => {
    const x = a?.[prop]
    const y = b?.[prop]
    return x < y ? -1 : x > y ? 1 : 0
  })
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

function getPath(obj: any, path: string): any {
  return path.split('.').reduce((v, k) => (v == null ? undefined : v[k]), obj)
}

// 读取表格当前排序方向。必须在 sortMethod 执行瞬间读取——Element Plus 的
// sort-change 事件在 execSort 之后才 emit（且 default-sort 初始应用 init:true
// 根本不发事件），用事件维护方向会永久滞后一拍；而 store.states.sortOrder 在
// updateSort 中先于 execSort 更新，故此处读到的一定是本次方向。
// 兼容两种传入形态：模板可能把 ref 解包成组件实例，也可能传入 Ref 本身。
function readOrder(tableRef: any): 'ascending' | 'descending' {
  try {
    const table = tableRef?.store ? tableRef : tableRef?.value
    const order = table?.store?.states?.sortOrder?.value
    return order === 'descending' ? 'descending' : 'ascending'
  } catch {
    return 'ascending'
  }
}

// 状态列排序助手（table-sorting 规格：状态排序以名称作为第二排序因子）：
// - 升序语义：启用(true)在前；非布尔状态值（如 availability 三值）按字符串升序
// - 状态部分：返回值交给 orderBy 乘方向系数 → 随主列方向翻转
// - 名称部分（第二因子）：固定顺序（字母在前、中文拼音在后）不随方向翻转——
//   orderBy 会把返回值整体乘方向，故此处预乘当前方向予以抵消
// - tableRef 为 el-table 实例引用；缺省时按升序处理（数据源预排序也走此路径）
export function sortByStatusThenName(
  a: any,
  b: any,
  tableRef: any = null,
  statusProp = 'enabled',
  nameProp = 'name'
): number {
  const reverse = readOrder(tableRef) === 'descending' ? -1 : 1
  const statusA = getPath(a, statusProp)
  const statusB = getPath(b, statusProp)

  if (statusA !== statusB) {
    if (typeof statusA === 'boolean' && typeof statusB === 'boolean') {
      return statusA ? -1 : 1 // 启用在前（方向由 orderBy 施加）
    }
    const sa = String(statusA ?? '')
    const sb = String(statusB ?? '')
    if (sa !== sb) return sa < sb ? -1 : 1
    return 0
  }

  const nameA = String(getPath(a, nameProp) ?? '')
  const nameB = String(getPath(b, nameProp) ?? '')
  if (nameA === nameB) return 0
  return compareName(nameA, nameB) * reverse // 预乘方向抵消 orderBy → 名称顺序恒定
}
