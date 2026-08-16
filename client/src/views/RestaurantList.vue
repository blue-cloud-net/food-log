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
      <el-input
        v-model="filters.cuisine_type"
        placeholder="菜系"
        clearable
        class="w-35"
        @keyup.enter="load(1)"
      />
      <el-button type="primary" plain @click="load(1)">查询</el-button>
    </div>

    <div v-loading="loading" class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-3.5 min-h-30">
      <div
        v-for="r in list"
        :key="r.id"
        class="fl-card cursor-pointer overflow-hidden p-0 transition-transform duration-150 hover:-translate-y-0.5"
        @click="router.push(`/restaurants/${r.id}`)"
      >
        <el-image
          v-if="r.images?.length"
          :src="r.images[0]"
          fit="cover"
          class="w-full h-130px"
        />
        <div v-else class="w-full h-130px flex items-center justify-center bg-[#fff7f0] text-primary-light">
          <el-icon :size="30"><Shop /></el-icon>
        </div>
        <div class="px-3.5 py-3">
          <div class="font-semibold text-[15px] mb-1.5">{{ r.name }}</div>
          <div class="flex items-center gap-2 text-[13px] text-[#909399] mb-1.5">
            <el-tag v-if="r.cuisine_type" size="small" effect="plain">{{ r.cuisine_type }}</el-tag>
            <span v-if="r.avg_rating" class="text-[13px]">⭐ {{ r.avg_rating }}</span>
            <span class="text-xs">{{ r.dish_count }} 道菜</span>
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
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listRestaurants } from '@/api/restaurant'
import type { Restaurant } from '@/api/types'

const router = useRouter()
const list = ref<Restaurant[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 12
const loading = ref(false)

const filters = reactive({ keyword: '', cuisine_type: '' })

async function load(p = 1) {
  loading.value = true
  try {
    const res = await listRestaurants({
      page: p,
      page_size: pageSize,
      keyword: filters.keyword || undefined,
      cuisine_type: filters.cuisine_type || undefined
    })
    list.value = res.list
    total.value = res.total
    page.value = res.page
  } finally {
    loading.value = false
  }
}
</script>

