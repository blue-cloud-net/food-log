<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">搜索「{{ keyword }}」</h2>
    </div>

    <div v-loading="loading">
      <!-- 菜谱 -->
      <template v-if="result?.recipes?.length">
        <div class="text-base font-semibold my-4 mb-3">📝 菜谱</div>
        <div class="grid grid-cols-[repeat(auto-fill,minmax(210px,1fr))] gap-3.5">
          <RecipeCard v-for="r in result.recipes" :key="r.id" :recipe="r" />
        </div>
      </template>

      <!-- 餐厅 -->
      <template v-if="result?.restaurants?.length">
        <div class="text-base font-semibold my-4 mb-3">🏠 餐厅</div>
        <div class="flex flex-col gap-2.5">
          <div v-for="r in result.restaurants" :key="r.id" class="fl-card flex items-center justify-between cursor-pointer py-3.5 px-4" @click="router.push(`/restaurants/${r.id}`)">
            <div class="min-w-0">
              <div class="font-semibold mb-1">{{ r.name }}</div>
              <div class="flex items-center gap-2 text-[13px] text-[#909399]">
                <el-tag v-if="r.cuisine_type" size="small" effect="plain">{{ r.cuisine_type }}</el-tag>
                <span v-if="r.address">{{ r.address }}</span>
              </div>
            </div>
            <el-icon color="#909399"><ArrowRight /></el-icon>
          </div>
        </div>
      </template>

      <!-- 菜品 -->
      <template v-if="result?.dishes?.length">
        <div class="text-base font-semibold my-4 mb-3">🥘 菜品</div>
        <div class="flex flex-col gap-2.5">
          <div v-for="d in result.dishes" :key="d.id" class="fl-card flex items-center justify-between cursor-pointer py-3.5 px-4" @click="router.push(`/restaurants/${d.restaurant_id}`)">
            <div class="min-w-0">
              <div class="font-semibold mb-1">{{ d.name }}</div>
              <div class="flex items-center gap-2 text-[13px] text-[#909399]">
                <span v-if="d.price">💰 ¥{{ Number(d.price).toFixed(2) }}</span>
                <span v-if="d.description">{{ d.description }}</span>
              </div>
            </div>
            <el-icon color="#909399"><ArrowRight /></el-icon>
          </div>
        </div>
      </template>

      <el-empty
        v-if="!loading && !hasResult"
        description="没有找到相关内容"
        :image-size="100"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { globalSearch } from '@/api/search'
import type { SearchResult } from '@/api/types'
import { useTagsStore } from '@/stores/tags'
import RecipeCard from '@/components/recipe/RecipeCard.vue'

const route = useRoute()
const router = useRouter()
const tagsStore = useTagsStore()

const keyword = computed(() => (route.query.keyword as string) || '')
const result = ref<SearchResult | null>(null)
const loading = ref(false)
const hasResult = computed(
  () => !!result.value && (!!result.value.recipes.length || !!result.value.restaurants.length || !!result.value.dishes.length)
)

async function load() {
  if (!keyword.value) return
  loading.value = true
  try {
    result.value = await globalSearch(keyword.value)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  tagsStore.ensureLoaded()
  load()
})

watch(keyword, load)
</script>

