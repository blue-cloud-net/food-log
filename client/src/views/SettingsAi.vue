<template>
  <div class="page-container flex flex-col gap-4 max-w-160">
    <div class="page-header">
      <h2 class="page-title">AI 设置</h2>
      <el-button @click="router.push('/settings')">返回设置</el-button>
    </div>

    <div class="fl-card">
      <h3 class="m-0 mb-2.5">🤖 AI API Key</h3>
      <p class="text-[#909399] text-xs leading-relaxed m-0 mb-3">
        可选填 AI API Key 覆盖后端配置（仅保存在本地浏览器）。用于图片识别与自动标签兜底；Ollama 本地模型无需填写。
      </p>
      <div class="flex gap-2">
        <el-input
          v-model="aiKey"
          placeholder="API Key（留空使用后端配置）"
          show-password
          clearable
        />
        <el-button type="primary" @click="saveAiKey">保存</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()

const aiKey = ref(localStorage.getItem('ai_api_key') || '')

function saveAiKey() {
  const k = aiKey.value.trim()
  if (k) localStorage.setItem('ai_api_key', k)
  else localStorage.removeItem('ai_api_key')
  ElMessage.success('AI Key 已保存')
}
</script>
