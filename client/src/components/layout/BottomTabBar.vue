<template>
  <nav class="fixed left-0 right-0 bottom-0 z-[100] flex bg-white border-t border-[#f0f0f0] pb-[env(safe-area-inset-bottom)]">
    <router-link
      v-for="item in items"
      :key="item.to"
      :to="item.to"
      class="flex-1 flex flex-col items-center gap-0.75 pt-2 pb-1.5 text-[11px] text-[#909399]"
      :class="{ 'text-primary': isActive(item.to) }"
    >
      <el-icon :size="22"><component :is="item.icon" /></el-icon>
      <span>{{ item.label }}</span>
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const items = [
  { to: '/', label: '首页', icon: 'HomeFilled' },
  { to: '/recipes', label: '菜谱', icon: 'Notebook' },
  { to: '/restaurants', label: '探店', icon: 'Shop' },
  { to: '/inventory', label: '库存', icon: 'Box' },
  { to: '/profile', label: '我的', icon: 'User' }
]

function isActive(to: string) {
  if (to === '/') return route.path === '/'
  return route.path.startsWith(to)
}
</script>

