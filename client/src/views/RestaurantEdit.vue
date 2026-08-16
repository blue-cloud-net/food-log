<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ isEdit ? '编辑餐厅' : '新建餐厅' }}</h2>
      <el-button @click="router.back()">返回</el-button>
    </div>

    <div class="fl-card">
      <el-form label-position="top">
        <el-form-item label="店名" required>
          <el-input v-model="form.name" maxlength="200" placeholder="餐厅名称" />
        </el-form-item>
        <el-form-item label="菜系">
          <el-input v-model="form.cuisine_type" placeholder="如：川菜 / 火锅 / 海鲜" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="form.address" placeholder="地址" />
        </el-form-item>
        <el-form-item label="描述 / 备注">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="环境、口味、推荐菜等" />
        </el-form-item>
        <el-form-item label="综合评分">
          <el-rate v-model="form.avg_rating" :max="5" allow-half />
        </el-form-item>
        <el-form-item label="环境照片">
          <div class="flex flex-wrap gap-2.5">
            <div v-for="(img, i) in form.images" :key="img" class="relative">
              <el-image :src="img" fit="cover" class="w-[84px] h-[84px] rounded-lg" :preview-src-list="form.images" preview-teleported />
              <el-icon class="absolute -top-1.5 -right-1.5 bg-white rounded-full text-[#f56c6c] cursor-pointer text-lg" @click="form.images.splice(i, 1)"><CircleClose /></el-icon>
            </div>
            <el-upload :http-request="customUpload" :show-file-list="false" accept="image/*" multiple>
              <div class="w-[84px] h-[84px] border border-dashed border-[#d9d9d9] rounded-lg flex items-center justify-center text-[#909399] cursor-pointer"><el-icon :size="22"><Plus /></el-icon></div>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" :loading="saving" style="width: 100%" @click="save">保存餐厅</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import { createRestaurant, getRestaurant, updateRestaurant } from '@/api/restaurant'
import { uploadImages } from '@/api/upload'
import type { Restaurant } from '@/api/types'

const route = useRoute()
const router = useRouter()
const isEdit = !!route.params.id
const saving = ref(false)
const restaurant = ref<Restaurant | null>(null)

const form = reactive({
  name: '',
  cuisine_type: '',
  address: '',
  description: '',
  avg_rating: 0,
  images: [] as string[]
})

onMounted(async () => {
  if (isEdit) {
    restaurant.value = await getRestaurant(route.params.id as string)
    Object.assign(form, {
      name: restaurant.value.name,
      cuisine_type: restaurant.value.cuisine_type,
      address: restaurant.value.address,
      description: restaurant.value.description,
      avg_rating: restaurant.value.avg_rating,
      images: [...(restaurant.value.images || [])]
    })
  }
})

async function customUpload(options: UploadRequestOptions) {
  try {
    const res = await uploadImages([options.file])
    form.images.push(res.files[0].url)
    options.onSuccess(res)
  } catch (e) {
    options.onError(e as never)
  }
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入店名')
    return
  }
  saving.value = true
  try {
    if (isEdit) {
      await updateRestaurant(route.params.id as string, form)
    } else {
      await createRestaurant(form)
    }
    ElMessage.success('保存成功')
    router.push('/restaurants')
  } finally {
    saving.value = false
  }
}
</script>

