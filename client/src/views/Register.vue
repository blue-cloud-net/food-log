<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-[#fff3ec] via-[#ffe9dc] to-white">
    <div class="fl-card w-full max-w-380px py-8 px-7 text-center">
      <div class="flex items-center justify-center gap-2.5 text-[26px] font-bold text-primary">
        <el-icon :size="34"><Bowl /></el-icon>
        <span>美食日志</span>
      </div>
      <p class="text-[#909399] mt-2 mb-6">创建你的美食记忆库</p>

      <el-form label-position="top" @submit.prevent>
        <el-form-item label="用户名" class="text-left">
          <el-input v-model="form.username" placeholder="2-50 个字符" size="large" />
        </el-form-item>
        <el-form-item label="邮箱" class="text-left">
          <el-input v-model="form.email" placeholder="邮箱" size="large" />
        </el-form-item>
        <el-form-item label="密码" class="text-left">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="至少 6 位"
            size="large"
            show-password
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button type="primary" size="large" :loading="loading" style="width: 100%" @click="submit">
          注 册
        </el-button>
      </el-form>

      <div class="mt-4.5 text-[13px] text-[#909399]">
        已有账号？<router-link to="/login">去登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const form = reactive({ username: '', email: '', password: '' })
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  if (!form.username || !form.email || !form.password) {
    ElMessage.warning('请填写完整信息')
    return
  }
  loading.value = true
  try {
    await auth.register(form.username, form.email, form.password)
    ElMessage.success('注册成功，欢迎使用！')
    router.push('/')
  } finally {
    loading.value = false
  }
}
</script>

