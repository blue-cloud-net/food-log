import { defineStore } from 'pinia'
import { getTagCategories } from '@/api/recipe'
import type { Tag, TagCategory } from '@/api/types'
import { normalizeColor, type TagColor } from '@/utils/tags'

// 标签词表全局缓存：扁平归一化，id → 标签 / 分类 便于 O(1) 解析显示名与颜色
export const useTagsStore = defineStore('tags', {
  state: () => ({
    categories: [] as TagCategory[],
    tagsById: {} as Record<string, Tag>,
    categoryById: {} as Record<string, TagCategory>,
    loaded: false
  }),
  getters: {
    // 全部可见标签（全局预设 + 本人自定义）
    allTags: (s): Tag[] => Object.values(s.tagsById),
    // 有标签的分类，供选择器/筛选下拉渲染
    selectableCategories: (s): TagCategory[] => s.categories.filter((c) => c.tags.length > 0),
    // id → 显示名
    tagName: (s) => (id: string): string => s.tagsById[id]?.name ?? '',
    // id → 所属分类颜色
    tagColor: (s) => (id: string): TagColor => {
      const t = s.tagsById[id]
      return t ? normalizeColor(s.categoryById[t.category_id]?.color) : 'info'
    },
    // id → 互斥组（用于选择器互斥）
    mutexGroupOf: (s) => (id: string): string => s.tagsById[id]?.mutex_group ?? '',
    // 按分类分组的筛选项（value 为标签 id）
    groupedOptions: (s): { label: string; options: { label: string; value: string }[] }[] =>
      s.categories
        .filter((c) => c.tags.length > 0)
        .map((c) => ({
          label: c.name,
          options: c.tags.map((t) => ({ label: t.name, value: t.id }))
        })),
    // 自定义标签可选的目标分类（本人自定义分类 + 全局预设分类）
    customTargetCategories: (s): TagCategory[] => s.categories.filter((c) => c.tags.length > 0 || !c.is_system)
  },
  actions: {
    async ensureLoaded(force = false) {
      if (this.loaded && !force) return
      try {
        this.setCategories(await getTagCategories())
        this.loaded = true
      } catch {
        /* 静默：词表加载失败不影响使用 */
      }
    },
    // 强制重新拉取（新建/删除自定义标签后调用）
    async reload() {
      this.loaded = false
      await this.ensureLoaded(true)
    },
    setCategories(categories: TagCategory[]) {
      const tagsById: Record<string, Tag> = {}
      const categoryById: Record<string, TagCategory> = {}
      for (const c of categories) {
        categoryById[c.id] = c
        for (const t of c.tags) tagsById[t.id] = t
      }
      this.categories = categories
      this.tagsById = tagsById
      this.categoryById = categoryById
    }
  }
})
