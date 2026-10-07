<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">库存食材</h2>
      <div class="flex items-center gap-2">
        <el-button v-if="!selectionMode" @click="enterSelection">
          <el-icon><Select /></el-icon>&nbsp;批量
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

    <div v-if="selectionMode" class="fl-card flex flex-wrap items-center gap-2 mb-4 py-2.5 px-3.5">
      <span class="text-[13px] text-[#606266]">已选 {{ selected.length }} 项</span>
      <el-button size="small" :disabled="!list.length" @click="toggleSelectAll">
        {{ allSelected ? '取消全选' : '全选' }}
      </el-button>
      <span class="flex-1"></span>
      <el-button size="small" @click="exitSelection">取消</el-button>
      <el-button size="small" type="danger" :disabled="!selected.length" @click="batchRemove">
        <el-icon><Delete /></el-icon>&nbsp;删除所选
      </el-button>
    </div>

    <div v-loading="loading" class="min-h-30">
      <div v-for="g in groups" :key="g.label" class="mb-6">
        <div class="section-title mb-2">{{ g.label }}</div>
        <div v-for="sec in g.sections" :key="sec.id" class="mb-4">
          <div class="flex items-center gap-2 mb-2 text-[13px] text-[#606266]">
            <span class="font-medium">{{ sec.name }}</span>
            <span class="text-[#c0c4cc]">{{ sec.items.length }} 项</span>
          </div>
          <div class="flex flex-col gap-1.5">
            <div
              v-for="it in sec.items"
              :key="it.id"
              class="fl-card flex items-center gap-2.5 max-[359px]:gap-1.5 px-3 py-1.5 cursor-pointer select-none overflow-hidden transition-colors duration-150 hover:bg-[#f5f5f5]"
              @click="onItemClick(it)"
              @pointerdown="startLongPress(it)"
              @pointerup="cancelLongPress"
              @pointerleave="cancelLongPress"
              @pointercancel="cancelLongPress"
              @contextmenu.prevent="onRowContextMenu(it)"
            >
              <span v-if="selectionMode" class="flex items-center shrink-0" @click.stop>
                <el-checkbox :model-value="isSelected(it.id)" @change="toggleOne(it.id)" />
              </span>
              <span class="font-medium text-[14px] flex-1 min-w-[4.5rem] truncate">{{ it.name }}</span>
              <span
                v-if="quantityText(it.quantity, it.unit)"
                class="text-[13px] text-[#606266] shrink-0 max-[359px]:hidden"
              >
                {{ quantityText(it.quantity, it.unit) }}
              </span>
              <el-tag
                v-if="it.category"
                size="small"
                type="info"
                effect="light"
                class="shrink-0"
              >
                {{ it.category }}
              </el-tag>
              <el-tag
                v-if="it.expire_at"
                size="small"
                effect="light"
                :type="expiryStatus(it.expire_at).type"
                class="shrink-0"
                :title="`保质期至 ${it.expire_at}`"
              >
                {{ expiryStatus(it.expire_at).text }}
              </el-tag>
              <span
                v-if="it.note"
                class="text-xs text-[#b0b3b8] truncate shrink-0 max-w-[30%] !hidden md:!inline"
                :title="it.note"
              >
                {{ it.note }}
              </span>
              <el-icon
                class="shrink-0 cursor-pointer text-[#c0c4cc] hover:text-[#67c23a]"
                :size="16"
                :title="selectionMode ? (isSelected(it.id) ? '取消选择' : '选择') : '消耗'"
                @pointerdown.stop
                @pointerup.stop
                @click.stop="onIconClick(it)"
              >
                <Select v-if="selectionMode" />
                <CircleCheck v-else />
              </el-icon>
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

    <!-- 消耗一定数量；减到 0 服务端会自动移出库存 -->
    <el-dialog v-model="consumeDialog" title="消耗" width="min(340px, 92vw)">
      <div class="text-[14px] mb-3">{{ consumeForm.name }}</div>
      <div class="flex items-center gap-2">
        <span class="text-[13px] text-[#606266] shrink-0">消耗数量</span>
        <el-input-number
          v-model="consumeForm.quantity"
          :min="minConsumable"
          :max="consumeForm.max"
          :step="1"
          controls-position="right"
          class="flex-1"
        />
        <span class="text-[13px] text-[#606266] shrink-0">{{ consumeForm.unit }}</span>
        <el-button
          link
          type="primary"
          class="shrink-0"
          @click="consumeForm.quantity = consumeForm.max"
        >
          全部
        </el-button>
      </div>
      <div class="text-xs text-[#909399] mt-2">
        当前剩余 {{ quantityText(consumeForm.max, consumeForm.unit) }}，减到 0 会移出库存
      </div>
      <template #footer>
        <el-button @click="consumeDialog = false">取消</el-button>
        <el-button type="primary" :loading="consuming" @click="submitConsume">消耗</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { batchDeleteInventoryItems, consumeInventoryItem, deleteInventoryItem, listInventory } from '@/api/inventory'
import type { InventoryItem } from '@/api/types'
import { useStorageLocationsStore } from '@/stores/storageLocations'
import { areaText, expiryStatus, quantityText, STORAGE_AREAS } from '@/utils/inventory'

const router = useRouter()
const locationsStore = useStorageLocationsStore()

const list = ref<InventoryItem[]>([])
const loading = ref(false)

const filters = reactive({ keyword: '', area: '', expiring: '' })

// ===== 批量选择 =====
const selectionMode = ref(false)
const selected = ref<string[]>([])

const isSelected = (id: string) => selected.value.includes(id)
const allSelected = computed(
  () => list.value.length > 0 && list.value.every((it) => isSelected(it.id))
)

function enterSelection() {
  selectionMode.value = true
  selected.value = []
}

function exitSelection() {
  selectionMode.value = false
  selected.value = []
}

function toggleOne(id: string) {
  const i = selected.value.indexOf(id)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(id)
}

function toggleSelectAll() {
  selected.value = allSelected.value ? [] : list.value.map((it) => it.id)
}

/** 点击条目：选择模式下切换勾选，否则进入编辑页。长按触发后紧跟的 click 需丢弃 */
function onItemClick(it: InventoryItem) {
  if (longPressFired) {
    longPressFired = false
    return
  }
  if (selectionMode.value) toggleOne(it.id)
  else router.push(`/inventory/${it.id}/edit`)
}

// ===== 用完（消耗） / 长按进入批量 =====
const LONG_PRESS_MS = 500
let pressTimer: ReturnType<typeof setTimeout> | null = null
let longPressFired = false

function startLongPress(it: InventoryItem) {
  if (selectionMode.value) return
  cancelLongPress()
  longPressFired = false
  pressTimer = setTimeout(() => {
    longPressFired = true
    enterSelection()
    toggleOne(it.id)
    ElMessage.info('已进入批量模式：可多选后删除')
  }, LONG_PRESS_MS)
}

function cancelLongPress() {
  if (pressTimer) {
    clearTimeout(pressTimer)
    pressTimer = null
  }
}

/** 右键（桌面端）等价于长按：进入批量模式并选中该行 */
function onRowContextMenu(it: InventoryItem) {
  if (selectionMode.value) return
  cancelLongPress()
  enterSelection()
  toggleOne(it.id)
}

/**
 * 「用完」：吃完 / 用掉了，直接把该条从库存移出。
 * 批量模式下改为切换选中（作为右侧额外的选择热区）。
 */
async function onIconClick(it: InventoryItem) {
  cancelLongPress()
  if (selectionMode.value) {
    toggleOne(it.id)
    return
  }
  openConsumeDialog(it)
}

// ===== 消耗弹层 =====
const consumeDialog = ref(false)
const consuming = ref(false)
const consumeForm = reactive({ id: '', name: '', unit: '', quantity: 1, max: 1 })

/** 最少可消耗量：不足 1 时只能整条消耗 */
const minConsumable = computed(() => (consumeForm.max < 1 ? consumeForm.max : 1))

function openConsumeDialog(it: InventoryItem) {
  consumeForm.id = it.id
  consumeForm.name = it.name
  consumeForm.unit = it.unit
  consumeForm.max = it.quantity
  consumeForm.quantity = minConsumable.value
  consumeDialog.value = true
}

async function submitConsume() {
  if (!consumeForm.quantity || consumeForm.quantity <= 0) {
    ElMessage.warning('请输入消耗数量')
    return
  }
  consuming.value = true
  try {
    const res = await consumeInventoryItem(consumeForm.id, consumeForm.quantity)
    consumeDialog.value = false
    if (res.removed) {
      ElMessage.success(`「${consumeForm.name}」已用完，已移出库存`)
    } else {
      ElMessage.success(`已消耗，还剩 ${quantityText(res.quantity, consumeForm.unit)}`)
    }
    await load()
  } catch {
    /* 失败已由拦截器提示 */
  } finally {
    consuming.value = false
  }
}

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
    // 结果集变化后剔除已不在列表中的选中项，避免批量删除命中不可见条目
    if (selectionMode.value) {
      const visible = new Set(list.value.map((it) => it.id))
      selected.value = selected.value.filter((id) => visible.has(id))
    }
  } finally {
    loading.value = false
  }
}

async function batchRemove() {
  if (!selected.value.length) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selected.value.length} 项吗？`,
      '批量删除',
      { type: 'warning' }
    )
  } catch {
    return
  }
  const res = await batchDeleteInventoryItems(selected.value)
  ElMessage.success(`已删除 ${res.deleted} 项`)
  exitSelection()
  await load()
}

onMounted(load)
</script>
