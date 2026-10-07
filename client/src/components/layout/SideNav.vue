<template>
  <nav class="flex flex-col gap-1">
    <router-link
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="flex items-center gap-2.5 px-3.5 py-2.5 rounded-lg text-sm transition-colors duration-150"
      :class="
        isActive(item.to)
          ? 'bg-primary text-white hover:bg-primary'
          : 'text-[#606266] hover:bg-[#f5f5f5]'
      "
    >
      <el-icon :size="18"><component :is="item.icon" /></el-icon>
      <span>{{ item.label }}</span>
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const items = [
  { to: '/', label: '首页', icon: 'HomeFilled' },
  { to: '/recipes', label: '我的菜谱', icon: 'Notebook' },
  { to: '/restaurants', label: '餐厅探店', icon: 'Shop' },
  { to: '/inventory', label: '库存食材', icon: 'Box' },
  { to: '/settings', label: '设置', icon: 'Setting' }
]

function isActive(to: string) {
  if (to === '/') return route.path === '/'
  return route.path.startsWith(to)
}
</script>

