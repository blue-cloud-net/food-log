<template>
  <div class="page-container">
    <div>
      <h2 class="m-0 mb-1">你好，{{ auth.user?.username || '朋友' }} 👋</h2>
      <p class="m-0 mb-4 text-[#909399]">今天想吃什么？</p>
    </div>

    <!-- 统计 -->
    <div class="grid grid-cols-3 gap-3 mb-4">
      <div class="fl-card text-center cursor-pointer" @click="router.push('/recipes')">
        <div class="text-[28px] font-bold text-primary">{{ auth.user?.recipe_count ?? '-' }}</div>
        <div class="text-[#909399] text-[13px] mt-1">自制菜谱</div>
      </div>
      <div class="fl-card text-center cursor-pointer" @click="router.push('/restaurants')">
        <div class="text-[28px] font-bold text-primary">{{ auth.user?.restaurant_count ?? '-' }}</div>
        <div class="text-[#909399] text-[13px] mt-1">探店餐厅</div>
      </div>
      <div class="fl-card text-center">
        <div class="text-[28px] font-bold text-primary">{{ auth.user?.dish_count ?? '-' }}</div>
        <div class="text-[#909399] text-[13px] mt-1">品尝菜品</div>
      </div>
    </div>

    <!-- 今天吃什么 -->
    <div class="bg-gradient-to-r from-[#ff8a50] to-[#ff6b35] text-white border-0 rounded-xl shadow-sm p-4 flex items-center justify-between mb-6">
      <div>
        <div class="text-lg font-bold">🍜 今天吃什么？</div>
        <div class="text-[13px] opacity-90 mt-1">让随机推荐帮你做决定</div>
      </div>
      <el-button type="primary" size="large" class="!bg-white !border-white !text-primary" @click="pick">
        <el-icon><Food /></el-icon>&nbsp;随机选菜
      </el-button>
    </div>

    <!-- 最近菜谱 -->
    <div class="flex items-center justify-between my-2 mb-3">
      <h3 class="m-0">最近的菜谱</h3>
      <el-link type="primary" @click="router.push('/recipes')">查看全部</el-link>
    </div>
    <div v-loading="recipesLoading" class="grid grid-cols-[repeat(auto-fill,minmax(210px,1fr))] gap-3.5 mb-2 min-h-20">
      <RecipeCard v-for="r in recipes" :key="r.id" :recipe="r" />
      <el-empty v-if="!recipesLoading && !recipes.length" description="还没有菜谱，去记录第一道菜吧" :image-size="80" />
    </div>

    <!-- 最近餐厅 -->
    <div class="flex items-center justify-between my-2 mb-3">
      <h3 class="m-0">最近的探店</h3>
      <el-link type="primary" @click="router.push('/restaurants')">查看全部</el-link>
    </div>
    <div v-loading="restLoading" class="flex flex-col gap-2.5">
      <div v-for="r in restaurants" :key="r.id" class="fl-card flex items-center justify-between cursor-pointer py-3.5 px-4" @click="router.push(`/restaurants/${r.id}`)">
        <div class="font-semibold">{{ r.name }}</div>
        <div class="flex items-center gap-2 text-[#909399] text-[13px]">
          <el-tag v-if="r.cuisine_type" size="small" effect="plain">{{ r.cuisine_type }}</el-tag>
          <span v-if="r.avg_rating">⭐ {{ r.avg_rating }}</span>
        </div>
      </div>
      <el-empty v-if="!restLoading && !restaurants.length" description="还没有探店记录" :image-size="80" />
    </div>

    <!-- 随机结果弹窗 -->
    <el-dialog v-model="randomDialog" title="🍜 今天吃这个！" width="min(420px, 92vw)">
      <div v-if="randomRec" class="text-center">
        <el-image v-if="cover" :src="cover" fit="cover" class="w-full h-[180px] rounded-lg mb-3" />
        <div class="text-xl font-bold mb-2">{{ randomRec.name }}</div>
        <div class="flex flex-wrap gap-1.5 justify-center mb-4">
          <el-tag v-for="id in randomRec.tags" :key="id" size="small" effect="light">
            {{ tagsStore.tagName(id) }}
          </el-tag>
          <el-tag
            v-for="id in randomRec.ingredient_tags"
            :key="id"
            size="small"
            effect="plain"
            :type="tagsStore.tagColor(id)"
          >
            {{ tagsStore.tagName(id) }}
          </el-tag>
        </div>
        <div class="flex justify-center gap-2">
          <el-button @click="pick">再看一道</el-button>
          <el-button type="primary" @click="router.push(`/recipes/${randomRec!.id}`)">查看做法</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listRecipes, randomRecipe } from '@/api/recipe'
import { listRestaurants } from '@/api/restaurant'
import type { Recipe, Restaurant } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { useTagsStore } from '@/stores/tags'
import RecipeCard from '@/components/recipe/RecipeCard.vue'

const auth = useAuthStore()
const tagsStore = useTagsStore()
const router = useRouter()

const recipes = ref<Recipe[]>([])
const restaurants = ref<Restaurant[]>([])
const recipesLoading = ref(false)
const restLoading = ref(false)

const randomDialog = ref(false)
const randomRec = ref<Recipe | null>(null)
const cover = computed(() => {
  const img = randomRec.value?.images?.[0]
  return img ? img.replace(/\.(jpg|jpeg|png|webp|gif)$/i, '_thumb.$1') : ''
})

async function pick() {
  try {
    randomRec.value = await randomRecipe()
    randomDialog.value = true
  } catch {
    /* 404 提示已由拦截器输出 */
  }
}

onMounted(async () => {
  tagsStore.ensureLoaded()
  recipesLoading.value = true
  restLoading.value = true
  try {
    const [r, rs] = await Promise.all([
      listRecipes({ page: 1, page_size: 6 }),
      listRestaurants({ page: 1, page_size: 4 })
    ])
    recipes.value = r.list
    restaurants.value = rs.list
  } catch {
    ElMessage.warning('加载首页数据失败')
  } finally {
    recipesLoading.value = false
    restLoading.value = false
  }
})
</script>

