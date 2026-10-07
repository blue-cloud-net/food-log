<template>
  <div class="page-container">
    <div v-if="restaurant" class="flex flex-col gap-4">
      <!-- 餐厅信息 -->
      <div class="fl-card">
        <div class="flex gap-4">
          <el-image
            v-if="restaurant.images?.length"
            :src="restaurant.images[0]"
            fit="cover"
            class="w-30 h-30 rounded-lg shrink-0"
          />
          <div v-else class="w-30 h-30 rounded-lg shrink-0 flex items-center justify-center bg-[#fff7f0] text-primary-light"><el-icon :size="40"><Shop /></el-icon></div>
          <div class="min-w-0 flex-1">
            <h1 class="m-0 mb-2 text-[22px]">{{ restaurant.name }}</h1>
            <div class="flex items-center gap-2.5 text-[13px] text-[#909399] mb-2">
              <el-tag v-if="restaurant.cuisine_type" effect="plain">{{ restaurant.cuisine_type }}</el-tag>
              <span v-if="restaurant.avg_rating">⭐ {{ restaurant.avg_rating }}</span>
              <span>{{ restaurant.dish_count }} 道菜</span>
            </div>
            <div v-if="restaurant.address" class="text-[#606266] text-[13px] mb-1.5">📍 {{ restaurant.address }}</div>
            <div v-if="restaurant.description" class="text-[#909399] text-[13px]">{{ restaurant.description }}</div>
          </div>
        </div>
        <div class="mt-3 flex gap-2 flex-wrap">
          <el-button type="primary" @click="addDialog = true">
            <el-icon><Plus /></el-icon>&nbsp;添加菜品
          </el-button>
          <el-button @click="router.push(`/restaurants/${restaurant.id}/edit`)">
            <el-icon><Edit /></el-icon>&nbsp;编辑
          </el-button>
          <el-button type="danger" plain @click="remove">
            <el-icon><Delete /></el-icon>&nbsp;删除
          </el-button>
        </div>
      </div>

      <!-- 菜品列表 -->
      <div class="fl-card">
        <h3 class="m-0 mb-3">🍽️ 菜品点评</h3>
        <div class="flex flex-col gap-2.5">
          <div v-for="d in restaurant.dishes" :key="d.id" class="flex gap-3 items-center bg-[#fafafa] rounded-lg p-2.5">
            <el-image
              v-if="d.images?.length"
              :src="d.images[0]"
              fit="cover"
              class="w-16 h-16 rounded-lg shrink-0"
            />
            <div v-else class="w-16 h-16 rounded-lg shrink-0 flex items-center justify-center bg-[#fff1e8] text-primary-light"><el-icon><Dish /></el-icon></div>
            <div class="flex-1 min-w-0">
              <div class="font-semibold mb-0.5">{{ d.name }}</div>
              <div v-if="d.description" class="text-[#909399] text-xs mb-1">{{ d.description }}</div>
              <div class="flex items-center gap-2.5 text-xs text-[#606266]">
                <span v-if="d.price">💰 ¥{{ Number(d.price).toFixed(2) }}</span>
                <el-rate v-if="d.rating" :model-value="d.rating" disabled size="small" />
                <span v-if="d.eaten_at">{{ d.eaten_at }}</span>
              </div>
            </div>
            <div class="flex shrink-0">
              <el-button text size="small" @click="openEdit(d)">
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button text size="small" type="danger" @click="removeDish(d)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-if="!restaurant.dishes.length" description="还没有记录菜品" :image-size="80" />
      </div>
    </div>
    <div v-else v-loading="true" class="loading-wrap" style="min-height: 200px" />
  </div>

  <!-- 添加/编辑菜品弹窗 -->
  <el-dialog v-model="addDialog" :title="editingDish ? '编辑菜品' : '添加菜品'" width="min(520px, 94vw)">
    <el-form label-position="top">
      <el-form-item label="菜名" required>
        <el-input v-model="dishForm.name" placeholder="菜名" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="dishForm.description" type="textarea" :rows="2" placeholder="口味描述" />
      </el-form-item>
      <el-row :gutter="12">
        <el-col :span="12">
          <el-form-item label="价格（元）">
            <el-input-number v-model="dishForm.price" :min="0" :precision="2" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="就餐日期">
            <el-date-picker v-model="dishForm.eaten_at" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="评分">
        <el-rate v-model="dishForm.rating" />
      </el-form-item>
      <el-form-item label="照片">
        <div class="flex flex-wrap gap-2">
          <div v-for="(img, i) in dishForm.images" :key="img" class="relative">
            <el-image :src="img" fit="cover" class="w-[70px] h-[70px] rounded-lg" />
            <el-icon class="absolute -top-1.5 -right-1.5 bg-white rounded-full text-[#f56c6c] cursor-pointer" @click="dishForm.images.splice(i, 1)"><CircleClose /></el-icon>
          </div>
          <el-upload :http-request="uploadDishImage" :show-file-list="false" accept="image/*">
            <div class="w-[70px] h-[70px] border border-dashed border-[#d9d9d9] rounded-lg flex items-center justify-center text-[#909399] cursor-pointer"><el-icon><Plus /></el-icon></div>
          </el-upload>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="addDialog = false">取消</el-button>
      <el-button type="primary" :loading="savingDish" @click="saveDish">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import { addDish, deleteDish, deleteRestaurant, getRestaurant, updateDish } from '@/api/restaurant'
import { uploadImages } from '@/api/upload'
import type { Dish, RestaurantDetail } from '@/api/types'

const route = useRoute()
const router = useRouter()

const restaurant = ref<RestaurantDetail | null>(null)
const addDialog = ref(false)
const editingDish = ref<Dish | null>(null)
const savingDish = ref(false)

const emptyDish = () => ({
  name: '',
  description: '',
  price: null as number | null,
  rating: 0,
  eaten_at: null as string | null,
  images: [] as string[]
})

const dishForm = reactive(emptyDish())

async function load() {
  restaurant.value = await getRestaurant(route.params.id as string)
}

function openEdit(d: Dish) {
  editingDish.value = d
  Object.assign(dishForm, {
    name: d.name,
    description: d.description,
    price: d.price,
    rating: d.rating,
    eaten_at: d.eaten_at,
    images: [...(d.images || [])]
  })
  addDialog.value = true
}

async function uploadDishImage(options: UploadRequestOptions) {
  try {
    // 菜品图属于餐厅探店场景，归入 restaurant 目录
    const res = await uploadImages([options.file], 'restaurant')
    dishForm.images.push(res.files[0].url)
    options.onSuccess(res)
  } catch (e) {
    options.onError(e as never)
  }
}

async function saveDish() {
  if (!dishForm.name.trim()) {
    ElMessage.warning('请输入菜名')
    return
  }
  savingDish.value = true
  try {
    const payload = {
      name: dishForm.name,
      description: dishForm.description,
      price: dishForm.price,
      rating: dishForm.rating,
      eaten_at: dishForm.eaten_at,
      images: dishForm.images
    }
    if (editingDish.value) {
      await updateDish(editingDish.value.id, payload)
    } else {
      await addDish(route.params.id as string, payload)
    }
    ElMessage.success('保存成功')
    addDialog.value = false
    await load()
  } finally {
    savingDish.value = false
  }
}

async function removeDish(d: Dish) {
  await ElMessageBox.confirm(`确定删除菜品「${d.name}」吗？`, '提示', { type: 'warning' }).catch(() => null)
  await deleteDish(d.id)
  ElMessage.success('已删除')
  await load()
}

async function remove() {
  if (!restaurant.value) return
  await ElMessageBox.confirm(`确定删除餐厅「${restaurant.value.name}」吗？其下菜品将一并删除。`, '删除确认', {
    type: 'warning'
  }).catch(() => null)
  await deleteRestaurant(restaurant.value.id)
  ElMessage.success('已删除')
  router.push('/restaurants')
}

onMounted(load)
</script>

