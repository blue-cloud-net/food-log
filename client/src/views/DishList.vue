<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">喜欢菜品</h2>
    </div>

    <div class="fl-card flex flex-wrap gap-2.5 mb-4 py-3 px-3.5">
      <el-input
        v-model="filters.keyword"
        placeholder="搜索菜名"
        clearable
        class="w-55 md:w-full"
        @keyup.enter="load(1)"
      />
      <el-select
        v-model="filters.tag"
        placeholder="菜品标签"
        clearable
        filterable
        class="w-40"
        @change="load(1)"
      >
        <el-option-group v-for="g in tagOptions" :key="g.label" :label="g.label">
          <el-option v-for="o in g.options" :key="o.value" :label="o.label" :value="o.value" />
        </el-option-group>
      </el-select>
      <div class="flex items-center gap-1.5 text-[13px] text-[#606266]">
        <el-switch v-model="filters.liked" @change="load(1)" />
        <span>只看喜欢</span>
      </div>
      <el-button type="primary" plain @click="load(1)">查询</el-button>
    </div>

    <div v-loading="loading" class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-3.5 min-h-30">
      <div v-for="d in list" :key="d.id" class="fl-card overflow-hidden p-0">
        <el-image v-if="d.images?.length" :src="d.images[0]" fit="cover" class="w-full h-130px" />
        <div v-else class="w-full h-130px flex items-center justify-center bg-[#fff7f0] text-primary-light">
          <el-icon :size="30"><Dish /></el-icon>
        </div>
        <div class="px-3.5 py-3">
          <div class="flex items-start justify-between gap-2 mb-1.5">
            <div class="font-semibold text-[15px] truncate">{{ d.name }}</div>
            <el-button text size="small" class="!p-0" @click="toggleLike(d)">
              {{ d.is_liked ? '❤️' : '🤍' }}
            </el-button>
          </div>
          <div v-if="d.restaurant_name" class="text-[13px] text-[#909399] mb-1.5 truncate">
            <router-link :to="`/restaurants/${d.restaurant_id}`" class="text-primary">
              {{ d.restaurant_name }}
            </router-link>
          </div>
          <div class="flex items-center gap-2.5 text-xs text-[#606266] mb-1.5">
            <span v-if="d.price">💰 ¥{{ Number(d.price).toFixed(2) }}</span>
            <span v-if="d.rating" class="flex items-center gap-1">
              推荐度 <el-rate :model-value="d.rating" disabled size="small" />
            </span>
            <span v-if="d.eaten_at">{{ d.eaten_at }}</span>
          </div>
          <div v-if="d.tags?.length" class="flex flex-wrap gap-1.5">
            <el-tag
              v-for="id in d.tags.slice(0, 4)"
              :key="id"
              size="small"
              effect="plain"
              :type="tagColor(id)"
            >
              {{ tagName(id) }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>
    <el-empty v-if="!loading && !list.length" description="还没有喜欢的菜品" :image-size="100" />

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
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { listDishes, setDishLiked } from '@/api/restaurant'
import type { Dish } from '@/api/types'
import { useShopTagsStore } from '@/stores/shopTags'

const route = useRoute()
const shopTagsStore = useShopTagsStore()

const list = ref<Dish[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 12
const loading = ref(false)

const filters = reactive({
  keyword: '',
  tag: '',
  // 首页「喜欢菜品」入口默认只看喜欢；liked=false 时展示全部菜品
  liked: route.query.liked !== 'false'
})

const tagOptions = computed(() => shopTagsStore.groupedOptions('dish'))
const tagName = (id: string) => shopTagsStore.tagNameOf('dish', id)
const tagColor = (id: string) => shopTagsStore.tagColorOf('dish', id)

async function load(p = 1) {
  loading.value = true
  try {
    const res = await listDishes({
      page: p,
      page_size: pageSize,
      keyword: filters.keyword || undefined,
      tag: filters.tag || undefined,
      liked: filters.liked || undefined
    })
    list.value = res.list
    total.value = res.total
    page.value = res.page
  } finally {
    loading.value = false
  }
}

async function toggleLike(d: Dish) {
  const next = !d.is_liked
  await setDishLiked(d.id, next)
  // 「只看喜欢」下取消喜欢后从当前列表移除
  if (filters.liked && !next) {
    list.value = list.value.filter((x) => x.id !== d.id)
    total.value = Math.max(0, total.value - 1)
    if (!list.value.length) load(1)
  } else {
    d.is_liked = next
  }
}

onMounted(() => {
  shopTagsStore.ensureLoaded('dish')
  load(1)
})
</script>
