import { defineStore } from 'pinia'
import { getTagCategories } from '@/api/recipe'
import type { TagCategory } from '@/api/types'

export const useTagsStore = defineStore('tags', {
  state: () => ({
    categories: [] as TagCategory[],
    loaded: false
  }),
  getters: {
    allTags: (s) => {
      const set = new Set<string>()
      s.categories.forEach((c) => c.tags.forEach((t) => set.add(t)))
      return Array.from(set)
    }
  },
  actions: {
    async ensureLoaded(force = false) {
      if (this.loaded && !force) return
      try {
        this.categories = await getTagCategories()
        this.loaded = true
      } catch {
        /* 静默：词表加载失败不影响使用 */
      }
    }
  }
})
