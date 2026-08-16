<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-[#fff3ec] via-[#ffe9dc] to-white">
    <div class="fl-card w-full max-w-380px py-8 px-7 text-center">
      <div class="flex items-center justify-center gap-2.5 text-[26px] font-bold text-primary">
        <el-icon :size="34"><Bowl /></el-icon>
        <span>美食日志</span>
      </div>
      <p class="text-[#909399] mt-2 mb-6">记录每一道用心做过的菜</p>

      <el-form label-position="top" @submit.prevent>
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" autofocus />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button type="primary" size="large" :loading="loading" style="width: 100%" @click="submit">
          登 录
        </el-button>
      </el-form>

      <div class="mt-4.5 text-[13px] text-[#909399]">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const form = reactive({ username: '', password: '' })
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

async function submit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } finally {
    loading.value = false
  }
}
</script>

