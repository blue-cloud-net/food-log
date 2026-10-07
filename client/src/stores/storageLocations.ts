import { defineStore } from 'pinia'
import { listStorageLocations } from '@/api/storageLocations'
import type { StorageArea, StorageLocation } from '@/api/types'
import { STORAGE_AREAS } from '@/utils/inventory'

// 存放位置词表全局缓存：解析 id → 名称，并按「冰箱 / 外面」大类分组供列表渲染
export const useStorageLocationsStore = defineStore('storageLocations', {
  state: () => ({
    locations: [] as StorageLocation[],
    loaded: false
  }),
  getters: {
    /** 指定大类下的位置（后端已按 sort_order 排序） */
    locationsByArea: (s) => (area: StorageArea): StorageLocation[] =>
      s.locations.filter((l) => l.area === area),
    /** 位置 id → 显示名 */
    locationName: (s) => (id: string): string =>
      s.locations.find((l) => l.id === id)?.name ?? '未归类',
    /** 本人自定义位置（可编辑 / 删除） */
    customLocations: (s): StorageLocation[] => s.locations.filter((l) => !l.is_system),
    /** 非空分组的渲染顺序 */
    areas: (): StorageArea[] => STORAGE_AREAS
  },
  actions: {
    async ensureLoaded(force = false) {
      if (this.loaded && !force) return
      try {
        this.locations = await listStorageLocations()
        this.loaded = true
      } catch {
        /* 静默：词表加载失败不影响使用 */
      }
    },
    /** 强制重新拉取（新增 / 删除自定义位置后调用） */
    async reload() {
      this.loaded = false
      await this.ensureLoaded(true)
    }
  }
})
