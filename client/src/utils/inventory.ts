import type { StorageArea } from '@/api/types'

/** 存放位置大类 → 中文名 */
export function areaText(area: StorageArea): string {
  switch (area) {
    case 'fridge':
      return '冰箱'
    case 'outside':
      return '外面'
    default:
      return area
  }
}

/** 存放位置大类展示顺序 */
export const STORAGE_AREAS: StorageArea[] = ['fridge', 'outside']

/** 食材分类预设（下拉可筛选、可自定义输入其它值） */
export const INGREDIENT_CATEGORIES = [
  '肉禽',
  '水产',
  '蔬菜',
  '水果',
  '蛋奶',
  '豆制品',
  '主食',
  '调味',
  '干货',
  '其他'
]

/** 临期阈值（天）：保质期在 N 天内视为「临期」 */
export const EXPIRING_SOON_DAYS = 3

export type ExpiryLevel = 'expired' | 'soon' | 'fresh' | 'none'

export interface ExpiryStatus {
  level: ExpiryLevel
  text: string
  type: 'danger' | 'warning' | 'success' | 'info'
}

/** 距今剩余天数（按本地日期计算，不含时间） */
function daysUntil(dateStr: string): number {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const target = new Date(`${dateStr}T00:00:00`)
  return Math.round((target.getTime() - today.getTime()) / 86400000)
}

/** 保质期状态：已过期 / 临期 / 充足 / 未设置 */
export function expiryStatus(expireAt?: string | null): ExpiryStatus {
  if (!expireAt) return { level: 'none', text: '', type: 'info' }
  const days = daysUntil(expireAt)
  if (days < 0) return { level: 'expired', text: `已过期 ${-days} 天`, type: 'danger' }
  if (days === 0) return { level: 'soon', text: '今天到期', type: 'warning' }
  if (days <= EXPIRING_SOON_DAYS) return { level: 'soon', text: `${days} 天后到期`, type: 'warning' }
  return { level: 'fresh', text: `${days} 天后到期`, type: 'success' }
}

/** 数量 + 单位展示文本 */
export function amountText(amount?: string, unit?: string): string {
  const a = (amount ?? '').trim()
  const u = (unit ?? '').trim()
  if (!a && !u) return ''
  return `${a}${u}`
}
