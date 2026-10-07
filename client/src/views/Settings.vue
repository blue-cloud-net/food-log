<template>
  <div class="page-container flex flex-col gap-4 max-w-160">
    <div class="page-header">
      <h2 class="page-title">设置</h2>
    </div>

    <div v-for="section in sections" :key="section.label">
      <div class="text-xs text-[#909399] mb-2 px-1">{{ section.label }}</div>
      <div class="fl-card p-0 overflow-hidden">
        <router-link
          v-for="entry in section.entries"
          :key="entry.to"
          :to="entry.to"
          class="flex items-center gap-3 px-4 py-3 border-b border-[#f0f0f0] last:border-b-0 hover:bg-[#fafafa] transition-colors duration-150"
        >
          <el-icon :size="18" class="text-primary shrink-0"><component :is="entry.icon" /></el-icon>
          <div class="flex-1 min-w-0">
            <div class="text-sm">{{ entry.title }}</div>
            <div v-if="entry.desc" class="text-xs text-[#b0b3b8] mt-0.5 truncate">{{ entry.desc }}</div>
          </div>
          <el-icon class="text-[#c0c4cc] shrink-0"><ArrowRight /></el-icon>
        </router-link>
      </div>
    </div>

    <el-button class="w-full" type="danger" plain @click="logout">
      <el-icon><SwitchButton /></el-icon>&nbsp;退出登录
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useStorageLocationsStore } from '@/stores/storageLocations'

interface Entry {
  title: string
  desc?: string
  icon: string
  to: string
}

const auth = useAuthStore()
const router = useRouter()
const locationsStore = useStorageLocationsStore()

const sections = computed<{ label: string; entries: Entry[] }[]>(() => [
  {
    label: '账号',
    entries: [{ title: '我的信息', desc: '用户名 / 邮箱 / 注册时间', icon: 'User', to: '/profile' }]
  },
  {
    label: '数据字典',
    entries: [
      {
        title: '存放位置',
        desc: locationsStore.loaded ? `冰箱 / 外面两区，共 ${locationsStore.locations.length} 个位置` : '冰箱 / 外面分区，预设 + 自定义',
        icon: 'Location',
        to: '/inventory/locations'
      },
      { title: '菜谱标签', desc: '菜谱分类与标签（含食材自动标签）', icon: 'Notebook', to: '/tags?scope=recipe' },
      { title: '餐厅标签', desc: '品类 / 菜系 / 场景', icon: 'Shop', to: '/tags?scope=restaurant' },
      { title: '菜品标签', desc: '口味 / 份量 / 特色', icon: 'Food', to: '/tags?scope=dish' }
    ]
  },
  {
    label: '其他',
    entries: [
      { title: 'AI 设置', desc: 'API Key，仅保存在本地浏览器', icon: 'MagicStick', to: '/settings/ai' },
      { title: '数据备份', desc: '导出 JSON / CSV，导入恢复', icon: 'Download', to: '/settings/backup' }
    ]
  }
])

async function logout() {
  await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' }).catch(() => null)
  auth.logout()
  router.push('/login')
}
</script>
