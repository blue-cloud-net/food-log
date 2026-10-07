<template>
  <div class="page-container flex flex-col gap-4 max-w-160">
    <div class="page-header">
      <h2 class="page-title">我的信息</h2>
      <el-button @click="router.push('/settings')">返回设置</el-button>
    </div>

    <!-- 用户信息 -->
    <div class="fl-card flex items-center gap-4">
      <el-avatar :size="64" class="bg-primary text-white text-[26px]">{{ (auth.user?.username || '我').slice(0, 1) }}</el-avatar>
      <div>
        <div class="text-xl font-bold">{{ auth.user?.username }}</div>
        <div class="text-[#909399] text-[13px] mt-0.5">{{ auth.user?.email }}</div>
        <div class="text-[#b0b3b8] text-xs mt-0.5">加入于 {{ formatDate(auth.user?.created_at) }}</div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

const auth = useAuthStore()
const router = useRouter()
</script>

