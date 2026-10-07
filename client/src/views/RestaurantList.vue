<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">餐厅探店</h2>
      <el-button type="primary" @click="router.push('/restaurants/new')">
        <el-icon><Plus /></el-icon>&nbsp;新建
      </el-button>
    </div>

    <div class="fl-card flex flex-wrap gap-2.5 mb-4 py-3 px-3.5">
      <el-input
        v-model="filters.keyword"
        placeholder="搜索店名 / 地址"
        clearable
        class="w-55 md:w-full"
        @keyup.enter="load(1)"
      />
      <el-select
        v-model="filters.tag"
        placeholder="标签"
        clearable
        filterable
        class="w-40"
        @change="load(1)"
      >
        <el-option-group v-for="g in tagOptions" :key="g.label" :label="g.label">
          <el-option v-for="o in g.options" :key="o.value" :label="o.label" :value="o.value" />
        </el-option-group>
      </el-select>
      <el-select
        v-model="filters.sort"
        placeholder="最近添加"
        clearable
        class="w-35"
        @change="load(1)"
      >
        <el-option label="推荐度" value="recommend" />
        <el-option label="性价比" value="value" />
        <el-option label="环境" value="ambience" />
        <el-option label="服务" value="service" />
      </el-select>
      <el-select v-model="filters.visited" placeholder="探店状态" clearable class="w-35" @change="load(1)">
        <el-option label="已探店" value="true" />
        <el-option label="未探店" value="false" />
      </el-select>
      <el-button type="primary" plain @click="load(1)">查询</el-button>
    </div>

    <div v-loading="loading" class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-3.5 min-h-30">
      <div
        v-for="r in list"
        :key="r.id"
        class="fl-card cursor-pointer overflow-hidden p-0 transition-transform duration-150 hover:-translate-y-0.5"
        @click="router.push(`/restaurants/${r.id}`)"
      >
        <el-image v-if="r.images?.length" :src="r.images[0]" fit="cover" class="w-full h-130px" />
        <div v-else class="w-full h-130px flex items-center justify-center bg-[#fff7f0] text-primary-light">
          <el-icon :size="30"><Shop /></el-icon>
        </div>
        <div class="px-3.5 py-3">
          <div class="font-semibold text-[15px] mb-1.5">{{ r.name }}</div>
          <div class="flex items-center gap-2.5 text-[13px] text-[#909399] mb-1.5">
            <span v-if="r.recommend_rating">⭐ {{ r.recommend_rating }}</span>
            <span v-if="r.value_rating" class="text-xs">💰 性价比 {{ r.value_rating }}</span>
            <span class="text-xs">{{ r.dish_count }} 道菜</span>
          </div>
          <div v-if="r.tags?.length" class="flex flex-wrap gap-1.5 mb-1.5">
            <el-tag
              v-for="id in r.tags.slice(0, 3)"
              :key="id"
              size="small"
              effect="light"
              :type="tagColor(id)"
            >
              {{ tagName(id) }}
            </el-tag>
          </div>
          <div class="text-xs text-[#b0b3b8] truncate">{{ r.address }}</div>
        </div>
      </div>
    </div>
    <el-empty v-if="!loading && !list.length" description="还没有探店记录" :image-size="100" />

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
import { useRoute, useRouter } from 'vue-router'
import { listRestaurants } from '@/api/restaurant'
import type { Restaurant } from '@/api/types'
import { useShopTagsStore } from '@/stores/shopTags'

const route = useRoute()
const router = useRouter()
const shopTagsStore = useShopTagsStore()

const list = ref<Restaurant[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 12
const loading = ref(false)

const filters = reactive({
  keyword: '',
  tag: '',
  sort: '',
  // 支持从首页「查看全部」带参数进入
  visited: (route.query.visited === 'true' || route.query.visited === 'false'
    ? route.query.visited
    : '') as '' | 'true' | 'false'
})

const tagOptions = computed(() => shopTagsStore.groupedOptions('restaurant'))
const tagName = (id: string) => shopTagsStore.tagNameOf('restaurant', id)
const tagColor = (id: string) => shopTagsStore.tagColorOf('restaurant', id)

async function load(p = 1) {
  loading.value = true
  try {
    const res = await listRestaurants({
      page: p,
      page_size: pageSize,
      keyword: filters.keyword || undefined,
      tag: filters.tag || undefined,
      sort: filters.sort || undefined,
      visited: filters.visited || undefined
    })
    list.value = res.list
    total.value = res.total
    page.value = res.page
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  shopTagsStore.ensureLoaded('restaurant')
  load(1)
})
</script>
