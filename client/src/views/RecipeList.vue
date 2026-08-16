<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">我的菜谱</h2>
      <div class="flex gap-2">
        <el-button @click="onRandom">
          <el-icon><Food /></el-icon>&nbsp;随机
        </el-button>
        <el-button type="primary" @click="router.push('/recipes/new')">
          <el-icon><Plus /></el-icon>&nbsp;新建
        </el-button>
      </div>
    </div>

    <div class="fl-card flex flex-wrap items-center gap-2.5 mb-4 py-3 px-3.5">
      <el-input
        v-model="filters.keyword"
        placeholder="搜索菜名"
        clearable
        class="w-45 md:w-full"
        @keyup.enter="load(1)"
      />
      <el-select v-model="filters.difficulty" placeholder="难度" clearable class="w-30 md:flex-1 md:min-w-110px">
        <el-option label="简单" value="easy" />
        <el-option label="中等" value="medium" />
        <el-option label="困难" value="hard" />
      </el-select>
      <el-select v-model="filters.tag" placeholder="标签" clearable filterable class="w-30 md:flex-1 md:min-w-110px">
        <el-option v-for="t in tagsStore.allTags" :key="t" :label="t" :value="t" />
      </el-select>
      <el-select v-model="filters.sort" placeholder="排序" clearable class="w-30 md:flex-1 md:min-w-110px">
        <el-option label="最新创建" value="created_at" />
        <el-option label="评分最高" value="rating" />
        <el-option label="耗时最短" value="cook_time" />
      </el-select>
      <div class="flex items-center gap-1.5 text-[13px] text-[#606266]">
        <el-switch v-model="filters.favorite" />
        <span>只看收藏</span>
      </div>
      <el-button type="primary" plain @click="load(1)">查询</el-button>
    </div>

    <div v-loading="loading" class="grid grid-cols-[repeat(auto-fill,minmax(210px,1fr))] gap-3.5 min-h-30">
      <RecipeCard v-for="r in list" :key="r.id" :recipe="r" />
    </div>
    <el-empty v-if="!loading && !list.length" description="暂无菜谱，快去记录一道吧" :image-size="100" />

    <div v-if="total > pageSize" class="flex justify-center mt-5">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="load"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { listRecipes, randomRecipe } from '@/api/recipe'
import type { Recipe } from '@/api/types'
import { useTagsStore } from '@/stores/tags'
import RecipeCard from '@/components/recipe/RecipeCard.vue'

const router = useRouter()
const tagsStore = useTagsStore()

const list = ref<Recipe[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 12
const loading = ref(false)

const filters = reactive<{
  keyword: string
  difficulty: string
  tag: string
  sort: string
  favorite: boolean
}>({
  keyword: '',
  difficulty: '',
  tag: '',
  sort: '',
  favorite: false
})

async function load(p = 1) {
  loading.value = true
  try {
    const res = await listRecipes({
      page: p,
      page_size: pageSize,
      keyword: filters.keyword || undefined,
      difficulty: filters.difficulty || undefined,
      tag: filters.tag || undefined,
      sort: filters.sort || undefined,
      favorite: filters.favorite || undefined
    })
    list.value = res.list
    total.value = res.total
    page.value = res.page
  } finally {
    loading.value = false
  }
}

async function onRandom() {
  try {
    const rec = await randomRecipe(filters.tag ? { tag: filters.tag } : {})
    router.push(`/recipes/${rec.id}`)
  } catch {
    /* 无匹配时拦截器已提示 */
  }
}

watch(
  () => [filters.difficulty, filters.tag, filters.sort, filters.favorite],
  () => load(1)
)

onMounted(() => {
  tagsStore.ensureLoaded()
  load(1)
})
</script>

