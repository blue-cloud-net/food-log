// el-tag 支持的色型
export type TagColor = 'primary' | 'success' | 'warning' | 'danger' | 'info'

const VALID_COLORS: TagColor[] = ['primary', 'success', 'warning', 'danger', 'info']

// 规范化分类颜色（分类自带 color，用户自定义分类也能有颜色）
export function normalizeColor(color?: string): TagColor {
  return VALID_COLORS.includes(color as TagColor) ? (color as TagColor) : 'info'
}
