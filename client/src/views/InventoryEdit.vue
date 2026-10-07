<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ isEdit ? '编辑食材' : '新增食材' }}</h2>
      <el-button @click="router.back()">返回</el-button>
    </div>

    <div class="fl-card">
      <el-form label-position="top">
        <el-form-item label="食材名称" required>
          <el-input v-model="form.name" maxlength="100" placeholder="如：鸡蛋" />
        </el-form-item>

        <el-form-item label="存放位置" required>
          <el-select v-model="form.location_id" placeholder="选择存放位置" style="width: 100%">
            <el-option-group v-for="g in locationGroups" :key="g.label" :label="g.label">
              <el-option v-for="o in g.options" :key="o.value" :label="o.label" :value="o.value" />
            </el-option-group>
          </el-select>
          <div class="text-xs text-[#909399] mt-1">
            找不到合适的位置？到「设置 → 存放位置」里添加。
          </div>
        </el-form-item>

        <div class="grid gap-x-3 md:grid-cols-2">
          <el-form-item label="数量">
            <el-input-number
              v-model="form.quantity"
              :min="0.01"
              :max="99999999"
              :step="1"
              controls-position="right"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="单位">
            <el-input v-model="form.unit" maxlength="20" placeholder="如：个 / 克 / 盒" />
          </el-form-item>
        </div>

        <div class="grid gap-x-3 md:grid-cols-2">
          <el-form-item label="食材分类">
            <el-select
              v-model="form.category"
              filterable
              allow-create
              default-first-option
              clearable
              placeholder="可选，也可自定义"
              style="width: 100%"
            >
              <el-option v-for="c in INGREDIENT_CATEGORIES" :key="c" :label="c" :value="c" />
            </el-select>
          </el-form-item>
          <el-form-item label="保质期 / 过期日期">
            <el-date-picker
              v-model="form.expire_at"
              type="date"
              value-format="YYYY-MM-DD"
              placeholder="选择日期"
              clearable
              style="width: 100%"
            />
          </el-form-item>
        </div>

        <el-form-item label="备注">
          <el-input
            v-model="form.note"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
            placeholder="如：开封后需冷藏"
          />
        </el-form-item>

        <el-form-item label="图片">
          <div class="flex flex-wrap gap-2.5">
            <div v-for="(img, i) in form.images" :key="img" class="relative">
              <el-image
                :src="img"
                fit="cover"
                class="w-[84px] h-[84px] rounded-lg"
                :preview-src-list="form.images"
                preview-teleported
              />
              <el-icon
                class="absolute -top-1.5 -right-1.5 bg-white rounded-full text-[#f56c6c] cursor-pointer text-lg"
                @click="form.images.splice(i, 1)"
              >
                <CircleClose />
              </el-icon>
            </div>
            <el-upload :http-request="customUpload" :show-file-list="false" accept="image/*" multiple>
              <div
                class="w-[84px] h-[84px] border border-dashed border-[#d9d9d9] rounded-lg flex items-center justify-center text-[#909399] cursor-pointer"
              >
                <el-icon :size="22"><Plus /></el-icon>
              </div>
            </el-upload>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" size="large" :loading="saving" style="width: 100%" @click="save">
            保存
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import { createInventoryItem, getInventoryItem, updateInventoryItem } from '@/api/inventory'
import { uploadImages } from '@/api/upload'
import { useStorageLocationsStore } from '@/stores/storageLocations'
import { areaText, INGREDIENT_CATEGORIES, STORAGE_AREAS } from '@/utils/inventory'

const route = useRoute()
const router = useRouter()
const locationsStore = useStorageLocationsStore()

const isEdit = !!route.params.id
const saving = ref(false)

const form = reactive({
  name: '',
  location_id: '',
  quantity: 1,
  unit: '',
  category: '',
  expire_at: '',
  note: '',
  images: [] as string[]
})

const locationGroups = computed(() =>
  STORAGE_AREAS.map((area) => ({
    label: areaText(area),
    options: locationsStore.locationsByArea(area).map((l) => ({ label: l.name, value: l.id }))
  })).filter((g) => g.options.length > 0)
)

onMounted(async () => {
  await locationsStore.ensureLoaded()
  if (!isEdit) return
  const it = await getInventoryItem(route.params.id as string)
  Object.assign(form, {
    name: it.name,
    location_id: it.location_id,
    quantity: it.quantity,
    unit: it.unit,
    category: it.category,
    expire_at: it.expire_at ?? '',
    note: it.note,
    images: [...(it.images || [])]
  })
})

async function customUpload(options: UploadRequestOptions) {
  try {
    const res = await uploadImages([options.file], 'inventory')
    form.images.push(res.files[0].url)
    options.onSuccess(res)
  } catch (e) {
    options.onError(e as never)
  }
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入食材名称')
    return
  }
  if (!form.location_id) {
    ElMessage.warning('请选择存放位置')
    return
  }
  if (!form.quantity || form.quantity <= 0) {
    ElMessage.warning('请输入数量')
    return
  }
  saving.value = true
  try {
    if (isEdit) {
      await updateInventoryItem(route.params.id as string, form)
    } else {
      await createInventoryItem(form)
    }
    ElMessage.success('保存成功')
    router.push('/inventory')
  } finally {
    saving.value = false
  }
}
</script>
