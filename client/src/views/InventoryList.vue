<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">库存食材</h2>
      <div class="flex items-center gap-2">
        <el-button @click="router.push('/inventory/locations')">
          <el-icon><Location /></el-icon>&nbsp;位置管理
        </el-button>
        <el-button type="primary" @click="router.push('/inventory/new')">
          <el-icon><Plus /></el-icon>&nbsp;新增
        </el-button>
      </div>
    </div>

    <div class="fl-card flex flex-wrap gap-2.5 mb-4 py-3 px-3.5">
      <el-input
        v-model="filters.keyword"
        placeholder="搜索食材 / 备注"
        clearable
        class="w-55"
        @keyup.enter="load"
      />
      <el-select v-model="filters.area" placeholder="冰箱 / 外面" clearable class="w-35" @change="load">
        <el-option label="冰箱" value="fridge" />
        <el-option label="外面" value="outside" />
      </el-select>
      <el-select v-model="filters.expiring" placeholder="保质期" clearable class="w-40" @change="load">
        <el-option label="临期 / 已过期" value="3" />
        <el-option label="7 天内到期" value="7" />
        <el-option label="30 天内到期" value="30" />
      </el-select>
      <el-button type="primary" plain @click="load">查询</el-button>
    </div>

    <div v-loading="loading" class="min-h-30">
      <div v-for="g in groups" :key="g.label" class="mb-6">
        <div class="section-title mb-2">{{ g.label }}</div>
        <div v-for="sec in g.sections" :key="sec.id" class="mb-4">
          <div class="flex items-center gap-2 mb-2 text-[13px] text-[#606266]">
            <span class="font-medium">{{ sec.name }}</span>
            <span class="text-[#c0c4cc]">{{ sec.items.length }} 项</span>
          </div>
          <div class="grid grid-cols-[repeat(auto-fill,minmax(250px,1fr))] gap-3">
            <div v-for="it in sec.items" :key="it.id" class="fl-card p-0 overflow-hidden">
              <div class="flex gap-3 p-3">
                <el-image
                  v-if="it.images?.length"
                  :src="it.images[0]"
                  fit="cover"
                  class="w-16 h-16 rounded-lg shrink-0"
                  :preview-src-list="it.images"
                  preview-teleported
                />
                <div
                  v-else
                  class="w-16 h-16 rounded-lg shrink-0 flex items-center justify-center bg-[#fff7f0] text-primary-light"
                >
                  <el-icon :size="22"><Bowl /></el-icon>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="font-semibold text-[15px] truncate">{{ it.name }}</div>
                  <div class="flex items-center gap-1.5 flex-wrap mt-1">
                    <el-tag v-if="amountText(it.amount, it.unit)" size="small" effect="plain">
                      {{ amountText(it.amount, it.unit) }}
                    </el-tag>
                    <el-tag v-if="it.category" size="small" type="info" effect="light">
                      {{ it.category }}
                    </el-tag>
                  </div>
                  <div v-if="it.expire_at" class="mt-1.5">
                    <el-tag size="small" effect="light" :type="expiryStatus(it.expire_at).type">
                      {{ it.expire_at }} · {{ expiryStatus(it.expire_at).text }}
                    </el-tag>
                  </div>
                </div>
              </div>
              <div v-if="it.note" class="px-3 pb-2 text-xs text-[#909399] line-clamp-2">{{ it.note }}</div>
              <div class="flex justify-end gap-1 px-2 pb-1.5">
                <el-button link type="primary" size="small" @click="router.push(`/inventory/${it.id}/edit`)">
                  编辑
                </el-button>
                <el-button link type="danger" size="small" @click="remove(it)">删除</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <el-empty
        v-if="!loading && !list.length"
        description="还没有库存食材，点右上角「新增」记录一下吧"
        :image-size="100"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteInventoryItem, listInventory } from '@/api/inventory'
import type { InventoryItem } from '@/api/types'
import { useStorageLocationsStore } from '@/stores/storageLocations'
import { areaText, amountText, expiryStatus, STORAGE_AREAS } from '@/utils/inventory'

const router = useRouter()
const locationsStore = useStorageLocationsStore()

const list = ref<InventoryItem[]>([])
const loading = ref(false)

const filters = reactive({ keyword: '', area: '', expiring: '' })

interface Group {
  label: string
  sections: { id: string; name: string; items: InventoryItem[] }[]
}

// 按 大类（冰箱 / 外面）→ 具体位置 分组；位置词表未加载成功时退化为单组平铺
const groups = computed<Group[]>(() => {
  const result: Group[] = []
  for (const area of STORAGE_AREAS) {
    const sections = locationsStore
      .locationsByArea(area)
      .map((loc) => ({
        id: loc.id,
        name: loc.name,
        items: list.value.filter((it) => it.location_id === loc.id)
      }))
      .filter((s) => s.items.length > 0)
    if (sections.length) result.push({ label: areaText(area), sections })
  }
  if (!result.length && list.value.length) {
    result.push({ label: '全部', sections: [{ id: '__all__', name: '全部食材', items: list.value }] })
  }
  return result
})

async function load() {
  loading.value = true
  try {
    await locationsStore.ensureLoaded()
    const res = await listInventory({
      keyword: filters.keyword || undefined,
      area: filters.area || undefined,
      expiring_within_days: filters.expiring ? Number(filters.expiring) : undefined,
      sort: 'expire'
    })
    list.value = res.list
  } finally {
    loading.value = false
  }
}

async function remove(it: InventoryItem) {
  try {
    await ElMessageBox.confirm(`确定删除「${it.name}」吗？`, '删除库存食材', { type: 'warning' })
  } catch {
    return
  }
  await deleteInventoryItem(it.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>
