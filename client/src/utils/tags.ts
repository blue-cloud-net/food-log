import type { TagCategory } from '@/api/types'

// 拉平所有预设标签
export function flattenTags(categories: TagCategory[]): string[] {
  const set = new Set<string>()
  categories.forEach((c) => c.tags.forEach((t) => set.add(t)))
  return Array.from(set)
}

// 标签归属分类名
export function categoryOf(categories: TagCategory[], tag: string): string {
  return categories.find((c) => c.tags.includes(tag))?.name ?? ''
}

// 标签颜色（按分类区分，提升可读性）
export function tagColor(categories: TagCategory[], tag: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  switch (categoryOf(categories, tag)) {
    case '荤素':
      return 'danger'
    case '食材':
      return 'success'
    case '场景':
      return 'warning'
    case '时段':
      return 'info'
    case '菜系':
      return 'primary'
    case '口味':
      return 'danger'
    default:
      return 'info'
  }
}
