import { defineStore } from 'pinia'
import { getDishTagCategories, getRestaurantTagCategories } from '@/api/restaurant'
import type { Tag, TagCategory } from '@/api/types'
import { normalizeColor, type TagColor } from '@/utils/tags'
import type { ShopTagDomain } from '@/api/tags'

interface DomainState {
  categories: TagCategory[]
  tagsById: Record<string, Tag>
  categoryById: Record<string, TagCategory>
}

function emptyDomain(): DomainState {
  return { categories: [], tagsById: {}, categoryById: {} }
}

function pick(
  state: { restaurant: DomainState; dish: DomainState },
  domain: ShopTagDomain
): DomainState {
  return domain === 'restaurant' ? state.restaurant : state.dish
}

// 探店标签词表缓存：餐厅与菜品各一套独立字典（后端也是两套独立表）。
// 与 stores/tags.ts 同构，差别是这里的 getter 都要求传入 domain。
export const useShopTagsStore = defineStore('shopTags', {
  state: () => ({
    restaurant: emptyDomain(),
    dish: emptyDomain(),
    loaded: { restaurant: false, dish: false } as Record<ShopTagDomain, boolean>
  }),
  getters: {
    // 全部分类（含无标签的分类），供标签管理页使用
    categoriesOf: (s) => (domain: ShopTagDomain): TagCategory[] => pick(s, domain).categories,
    // 有标签的分类，供选择器/筛选下拉渲染
    selectableCategories: (s) => (domain: ShopTagDomain): TagCategory[] =>
      pick(s, domain).categories.filter((c) => c.tags.length > 0),
    // 自定义标签可选的目标分类（本人自定义分类 + 全局预设分类）
    customTargetCategories: (s) => (domain: ShopTagDomain): TagCategory[] =>
      pick(s, domain).categories.filter((c) => c.tags.length > 0 || !c.is_system),
    // 全部可见标签
    tagsOf: (s) => (domain: ShopTagDomain): Tag[] => Object.values(pick(s, domain).tagsById),
    // id → 显示名
    tagNameOf: (s) => (domain: ShopTagDomain, id: string): string =>
      pick(s, domain).tagsById[id]?.name ?? '',
    // id → 所属分类颜色
    tagColorOf: (s) => (domain: ShopTagDomain, id: string): TagColor => {
      const st = pick(s, domain)
      const t = st.tagsById[id]
      return t ? normalizeColor(st.categoryById[t.category_id]?.color) : 'info'
    },
    // id → 互斥组（用于选择器互斥）
    mutexGroupOf: (s) => (domain: ShopTagDomain, id: string): string =>
      pick(s, domain).tagsById[id]?.mutex_group ?? '',
    // 按分类分组的筛选项（value 为标签 id）
    groupedOptions: (s) => (domain: ShopTagDomain) =>
      pick(s, domain)
        .categories.filter((c) => c.tags.length > 0)
        .map((c) => ({
          label: c.name,
          options: c.tags.map((t) => ({ label: t.name, value: t.id }))
        }))
  },
  actions: {
    async ensureLoaded(domain: ShopTagDomain, force = false) {
      if (this.loaded[domain] && !force) return
      try {
        const categories =
          domain === 'restaurant' ? await getRestaurantTagCategories() : await getDishTagCategories()
        this.setCategories(domain, categories)
        this.loaded[domain] = true
      } catch {
        /* 静默：词表加载失败不影响使用 */
      }
    },
    // 强制重新拉取（新建/删除自定义标签后调用）
    async reload(domain: ShopTagDomain) {
      this.loaded[domain] = false
      await this.ensureLoaded(domain, true)
    },
    setCategories(domain: ShopTagDomain, categories: TagCategory[]) {
      const tagsById: Record<string, Tag> = {}
      const categoryById: Record<string, TagCategory> = {}
      for (const c of categories) {
        categoryById[c.id] = c
        for (const t of c.tags) tagsById[t.id] = t
      }
      const next: DomainState = { categories, tagsById, categoryById }
      if (domain === 'restaurant') {
        this.restaurant = next
      } else {
        this.dish = next
      }
    }
  }
})
