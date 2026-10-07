<template>
  <div class="page-container">
    <div>
      <h2 class="m-0 mb-1">你好，{{ auth.user?.username || '朋友' }} 👋</h2>
      <p class="m-0 mb-4 text-[#909399]">今天想吃什么？</p>
    </div>

    <!-- 今天吃什么 -->
    <div class="bg-gradient-to-r from-[#ff8a50] to-[#ff6b35] text-white border-0 rounded-xl shadow-sm p-4 flex items-center justify-between mb-5">
      <div>
        <div class="text-lg font-bold">🍜 今天吃什么？</div>
        <div class="text-[13px] opacity-90 mt-1">让随机推荐帮你做决定</div>
      </div>
      <el-button type="primary" size="large" class="!bg-white !border-white !text-primary" @click="pick">
        <el-icon><Food /></el-icon>&nbsp;随机选菜
      </el-button>
    </div>

    <div v-loading="loading">
      <el-tabs v-model="activeTab">
        <!-- 自制 -->
        <el-tab-pane name="recipe">
          <template #label>
            <span class="flex items-center gap-1.5"><el-icon><Notebook /></el-icon>自制</span>
          </template>

          <!-- 已做 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="m-0 flex items-center gap-2">
                ✅ 已做 <el-tag size="small" effect="plain">{{ madeTotal }}</el-tag>
              </h3>
              <el-link type="primary" @click="router.push({ path: '/recipes', query: { made: 'true' } })">查看全部</el-link>
            </div>
            <div class="grid grid-cols-[repeat(auto-fill,minmax(210px,1fr))] gap-3.5 min-h-20">
              <RecipeCard v-for="r in madeRecipes" :key="r.id" :recipe="r" />
              <el-empty v-if="!madeRecipes.length" description="还没有做过的菜" :image-size="70" />
            </div>
          </div>

          <!-- 未做 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="m-0 flex items-center gap-2">
                🍳 未做 <el-tag size="small" effect="plain">{{ todoTotal }}</el-tag>
              </h3>
              <el-link type="primary" @click="router.push({ path: '/recipes', query: { made: 'false' } })">查看全部</el-link>
            </div>
            <div class="grid grid-cols-[repeat(auto-fill,minmax(210px,1fr))] gap-3.5 min-h-20">
              <RecipeCard v-for="r in todoRecipes" :key="r.id" :recipe="r" />
              <el-empty v-if="!todoRecipes.length" description="没有待做的菜" :image-size="70" />
            </div>
          </div>

          <!-- 喜欢 -->
          <div class="mb-2">
            <div class="flex items-center justify-between mb-3">
              <h3 class="m-0 flex items-center gap-2">
                ❤️ 喜欢 <el-tag size="small" effect="plain">{{ likedRecipeTotal }}</el-tag>
              </h3>
              <el-link type="primary" @click="router.push({ path: '/recipes', query: { liked: 'true' } })">查看全部</el-link>
            </div>
            <div class="grid grid-cols-[repeat(auto-fill,minmax(210px,1fr))] gap-3.5 min-h-20">
              <RecipeCard v-for="r in likedRecipes" :key="r.id" :recipe="r" />
              <el-empty v-if="!likedRecipes.length" description="还没有喜欢的菜谱" :image-size="70" />
            </div>
          </div>
        </el-tab-pane>

        <!-- 探店 -->
        <el-tab-pane name="restaurant">
          <template #label>
            <span class="flex items-center gap-1.5"><el-icon><Shop /></el-icon>探店</span>
          </template>

          <!-- 探店餐厅 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="m-0 flex items-center gap-2">
                ✅ 探店餐厅 <el-tag size="small" effect="plain">{{ visitedTotal }}</el-tag>
              </h3>
              <el-link type="primary" @click="router.push({ path: '/restaurants', query: { visited: 'true' } })">查看全部</el-link>
            </div>
            <div class="flex flex-col gap-2.5">
              <div
                v-for="r in visitedRestaurants"
                :key="r.id"
                class="fl-card flex items-center justify-between cursor-pointer py-3 px-4"
                @click="router.push(`/restaurants/${r.id}`)"
              >
                <span class="font-semibold truncate">{{ r.name }}</span>
                <div class="flex items-center gap-2 text-[#909399] text-[13px] shrink-0">
                  <el-tag
                    v-for="id in (r.tags || []).slice(0, 2)"
                    :key="id"
                    size="small"
                    effect="plain"
                    :type="restTagColor(id)"
                  >
                    {{ restTagName(id) }}
                  </el-tag>
                  <span v-if="r.recommend_rating">⭐ {{ r.recommend_rating }}</span>
                </div>
              </div>
              <el-empty v-if="!visitedRestaurants.length" description="还没有已探店的餐厅" :image-size="70" />
            </div>
          </div>

          <!-- 未探店餐厅 -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="m-0 flex items-center gap-2">
                📍 未探店餐厅 <el-tag size="small" effect="plain">{{ unvisitedTotal }}</el-tag>
              </h3>
              <el-link type="primary" @click="router.push({ path: '/restaurants', query: { visited: 'false' } })">查看全部</el-link>
            </div>
            <div class="flex flex-col gap-2.5">
              <div
                v-for="r in unvisitedRestaurants"
                :key="r.id"
                class="fl-card flex items-center justify-between cursor-pointer py-3 px-4"
                @click="router.push(`/restaurants/${r.id}`)"
              >
                <span class="font-semibold truncate">{{ r.name }}</span>
                <div class="flex items-center gap-2 text-[#909399] text-[13px] shrink-0">
                  <el-tag
                    v-for="id in (r.tags || []).slice(0, 2)"
                    :key="id"
                    size="small"
                    effect="plain"
                    :type="restTagColor(id)"
                  >
                    {{ restTagName(id) }}
                  </el-tag>
                  <span v-if="r.address" class="truncate max-w-40">{{ r.address }}</span>
                </div>
              </div>
              <el-empty v-if="!unvisitedRestaurants.length" description="没有待探店的餐厅" :image-size="70" />
            </div>
          </div>

          <!-- 喜欢菜品 -->
          <div class="mb-2">
            <div class="flex items-center justify-between mb-3">
              <h3 class="m-0 flex items-center gap-2">
                ❤️ 喜欢菜品 <el-tag size="small" effect="plain">{{ likedDishTotal }}</el-tag>
              </h3>
              <el-link type="primary" @click="router.push({ path: '/dishes', query: { liked: 'true' } })">查看全部</el-link>
            </div>
            <div class="flex flex-col gap-2.5">
              <div
                v-for="d in likedDishes"
                :key="d.id"
                class="fl-card flex items-center justify-between cursor-pointer py-3 px-4"
                @click="router.push(`/restaurants/${d.restaurant_id}`)"
              >
                <div class="min-w-0 flex items-center gap-2">
                  <span class="font-semibold truncate">{{ d.name }}</span>
                  <span v-if="d.restaurant_name" class="text-[#909399] text-[13px] truncate">{{ d.restaurant_name }}</span>
                </div>
                <div class="flex items-center gap-3 text-[#909399] text-[13px] shrink-0">
                  <span v-if="d.price">¥{{ Number(d.price).toFixed(2) }}</span>
                  <span v-if="d.rating">⭐ {{ d.rating }}</span>
                </div>
              </div>
              <el-empty v-if="!likedDishes.length" description="还没有喜欢的菜品" :image-size="70" />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
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
import { listDishes, listRestaurants } from '@/api/restaurant'
import type { Dish, Recipe, Restaurant } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { useShopTagsStore } from '@/stores/shopTags'
import { useTagsStore } from '@/stores/tags'
import RecipeCard from '@/components/recipe/RecipeCard.vue'

const auth = useAuthStore()
const tagsStore = useTagsStore()
const shopTagsStore = useShopTagsStore()
const router = useRouter()

const restTagName = (id: string) => shopTagsStore.tagNameOf('restaurant', id)
const restTagColor = (id: string) => shopTagsStore.tagColorOf('restaurant', id)

const activeTab = ref<'recipe' | 'restaurant'>('recipe')
const loading = ref(false)

// 自制：已做 / 未做 / 喜欢
const madeRecipes = ref<Recipe[]>([])
const madeTotal = ref(0)
const todoRecipes = ref<Recipe[]>([])
const todoTotal = ref(0)
const likedRecipes = ref<Recipe[]>([])
const likedRecipeTotal = ref(0)

// 探店：探店餐厅 / 未探店餐厅 / 喜欢菜品
const visitedRestaurants = ref<Restaurant[]>([])
const visitedTotal = ref(0)
const unvisitedRestaurants = ref<Restaurant[]>([])
const unvisitedTotal = ref(0)
const likedDishes = ref<Dish[]>([])
const likedDishTotal = ref(0)

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

async function load() {
  loading.value = true
  try {
    const [made, todo, liked, visited, unvisited, dishes] = await Promise.all([
      listRecipes({ page: 1, page_size: 6, made: 'true' }),
      listRecipes({ page: 1, page_size: 6, made: 'false' }),
      listRecipes({ page: 1, page_size: 6, liked: true }),
      listRestaurants({ page: 1, page_size: 4, visited: 'true' }),
      listRestaurants({ page: 1, page_size: 4, visited: 'false' }),
      listDishes({ page: 1, page_size: 6, liked: true })
    ])
    madeRecipes.value = made.list
    madeTotal.value = made.total
    todoRecipes.value = todo.list
    todoTotal.value = todo.total
    likedRecipes.value = liked.list
    likedRecipeTotal.value = liked.total
    visitedRestaurants.value = visited.list
    visitedTotal.value = visited.total
    unvisitedRestaurants.value = unvisited.list
    unvisitedTotal.value = unvisited.total
    likedDishes.value = dishes.list
    likedDishTotal.value = dishes.total
  } catch {
    ElMessage.warning('加载首页数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  tagsStore.ensureLoaded()
  shopTagsStore.ensureLoaded('restaurant')
  shopTagsStore.ensureLoaded('dish')
  load()
})
</script>

