<template>
  <div class="min-h-screen">
    <!-- 桌面端侧边栏 -->
    <aside
      class="fixed left-0 top-0 bottom-0 hidden md:flex w-55 bg-white border-r border-[#f0f0f0] p-4 flex-col gap-3"
    >
      <div class="flex items-center gap-2 px-3 py-2 text-xl font-bold text-primary">
        <el-icon :size="26"><Bowl /></el-icon>
        <span>美食日志</span>
      </div>
      <SideNav />
    </aside>

    <!-- 主区域 -->
    <div class="min-h-screen flex flex-col md:ml-55">
      <header class="sticky top-0 z-[100] flex items-center gap-3 px-5 py-3 bg-white/90 backdrop-blur border-b border-[#f0f0f0]">
        <GlobalSearch class="flex-1 max-w-420px md:max-w-none" />
        <el-dropdown trigger="click" class="cursor-pointer">
          <span class="flex items-center gap-1.5 text-sm text-[#606266] outline-none">
            <el-avatar :size="30" class="bg-primary text-white shrink-0">
              {{ (auth.user?.username || '我').slice(0, 1) }}
            </el-avatar>
            {{ auth.user?.username || '我' }}
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="router.push('/profile')">个人中心</el-dropdown-item>
              <el-dropdown-item divided @click="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>

      <main class="flex-1">
        <router-view />
      </main>
    </div>

    <!-- 移动端底部 TabBar -->
    <BottomTabBar class="md:hidden" />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useTagsStore } from '@/stores/tags'
import SideNav from './SideNav.vue'
import BottomTabBar from './BottomTabBar.vue'
import GlobalSearch from './GlobalSearch.vue'

const auth = useAuthStore()
const tagsStore = useTagsStore()
const router = useRouter()

async function logout() {
  await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' }).catch(() => null)
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  auth.fetchMe()
  tagsStore.ensureLoaded()
})
</script>

