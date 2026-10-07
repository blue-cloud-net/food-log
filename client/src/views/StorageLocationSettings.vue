<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">存放位置管理</h2>
      <div class="flex items-center gap-2">
        <el-button @click="router.push('/settings')">返回设置</el-button>
        <el-button :loading="loading" @click="reload">刷新</el-button>
      </div>
    </div>

    <el-alert
      class="mb-4"
      type="info"
      :closable="false"
      show-icon
      title="系统预设位置（只读）不可修改或删除；你可以新增属于自己的位置。位置下仍有库存食材时无法删除。"
    />

    <div class="grid gap-4 md:grid-cols-2">
      <div v-for="area in STORAGE_AREAS" :key="area" class="fl-card">
        <div class="flex items-center justify-between mb-3">
          <span class="font-medium">{{ areaText(area) }}</span>
          <el-button size="small" type="primary" @click="openDialog(area)">新增位置</el-button>
        </div>
        <el-table v-if="locationsOf(area).length" :data="locationsOf(area)" size="small">
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column label="来源" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.is_system ? 'info' : 'success'" effect="plain">
                {{ row.is_system ? '预设' : '自定义' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="right">
            <template #default="{ row }">
              <template v-if="!row.is_system">
                <el-button link type="primary" @click="openDialog(area, row)">编辑</el-button>
                <el-button link type="danger" @click="remove(row)">删除</el-button>
              </template>
              <span v-else class="text-xs text-[#c0c4cc]">只读</span>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无位置" :image-size="60" />
      </div>
    </div>

    <el-dialog v-model="dialog" :title="form.id ? '编辑位置' : '新增位置'" width="min(420px, 92vw)">
      <el-form label-width="72px" @submit.prevent>
        <el-form-item label="大类">
          <el-radio-group v-model="form.area">
            <el-radio-button value="fridge">冰箱</el-radio-button>
            <el-radio-button value="outside">外面</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" maxlength="50" show-word-limit placeholder="如：冰箱上层 / 阳台" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createStorageLocation,
  deleteStorageLocation,
  updateStorageLocation
} from '@/api/storageLocations'
import type { StorageArea, StorageLocation } from '@/api/types'
import { useStorageLocationsStore } from '@/stores/storageLocations'
import { areaText, STORAGE_AREAS } from '@/utils/inventory'

const router = useRouter()
const store = useStorageLocationsStore()

const loading = ref(false)
const saving = ref(false)
const dialog = ref(false)

const form = reactive<{ id: string; area: StorageArea; name: string }>({
  id: '',
  area: 'fridge',
  name: ''
})

const locationsOf = (area: StorageArea) => store.locationsByArea(area)

async function reload() {
  loading.value = true
  try {
    await store.reload()
  } finally {
    loading.value = false
  }
}

function openDialog(area: StorageArea, row?: StorageLocation) {
  form.id = row?.id ?? ''
  form.area = (row?.area ?? area) as StorageArea
  form.name = row?.name ?? ''
  dialog.value = true
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入位置名称')
    return
  }
  saving.value = true
  try {
    if (form.id) {
      await updateStorageLocation(form.id, { area: form.area, name: form.name.trim() })
    } else {
      await createStorageLocation({ area: form.area, name: form.name.trim() })
    }
    dialog.value = false
    ElMessage.success('已保存')
    await store.reload()
  } catch {
    /* 失败已由拦截器提示 */
  } finally {
    saving.value = false
  }
}

async function remove(row: StorageLocation) {
  try {
    await ElMessageBox.confirm(`确定删除位置「${row.name}」吗？`, '删除存放位置', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteStorageLocation(row.id)
    ElMessage.success('已删除')
    await store.reload()
  } catch {
    /* 位置下仍有库存食材时服务端返回 409，已由拦截器提示 */
  }
}

onMounted(reload)
</script>
