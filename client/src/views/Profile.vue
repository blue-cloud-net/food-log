<template>
  <div class="page-container flex flex-col gap-4 max-w-160">
    <!-- 用户信息 -->
    <div class="fl-card flex items-center gap-4">
      <el-avatar :size="64" class="bg-primary text-white text-[26px]">{{ (auth.user?.username || '我').slice(0, 1) }}</el-avatar>
      <div>
        <div class="text-xl font-bold">{{ auth.user?.username }}</div>
        <div class="text-[#909399] text-[13px] mt-0.5">{{ auth.user?.email }}</div>
        <div class="text-[#b0b3b8] text-xs mt-0.5">加入于 {{ formatDate(auth.user?.created_at) }}</div>
      </div>
    </div>

    <!-- AI 设置 -->
    <div class="fl-card">
      <h3 class="m-0 mb-2.5">🤖 AI 设置</h3>
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

    <!-- 数据管理 -->
    <div class="fl-card">
      <h3 class="m-0 mb-2.5">💾 数据备份</h3>
      <div class="flex gap-2.5 flex-wrap mb-4">
        <el-button @click="doExportJson">
          <el-icon><Download /></el-icon>&nbsp;导出 JSON 备份
        </el-button>
        <el-button @click="doExportCsv">
          <el-icon><Download /></el-icon>&nbsp;导出菜谱 CSV
        </el-button>
      </div>

      <div>
        <div class="text-[13px] font-semibold mb-2">导入恢复</div>
        <el-radio-group v-model="importMode" class="mb-3">
          <el-radio value="append">追加（同名跳过）</el-radio>
          <el-radio value="overwrite">覆盖（清空后导入）</el-radio>
        </el-radio-group>
        <el-upload
          :auto-upload="false"
          accept=".json,application/json"
          :on-change="onImportFile"
          :show-file-list="false"
          :limit="1"
        >
          <el-button type="warning" plain :disabled="importing">
            <el-icon><Upload /></el-icon>&nbsp;{{ importing ? '导入中…' : '选择备份文件导入' }}
          </el-button>
        </el-upload>
        <el-alert
          v-if="importMode === 'overwrite'"
          title="覆盖模式会清空当前全部数据，请务必先导出备份！"
          type="warning"
          :closable="false"
          class="mt-3"
        />
      </div>
    </div>

    <!-- 退出 -->
    <el-button class="w-full" type="danger" plain @click="logout">
      <el-icon><SwitchButton /></el-icon>&nbsp;退出登录
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { exportCsvBlob, exportJson, importData } from '@/api/export'
import { useAuthStore } from '@/stores/auth'
import { downloadBlob, formatDate } from '@/utils/format'

const auth = useAuthStore()
const router = useRouter()

const aiKey = ref(localStorage.getItem('ai_api_key') || '')
const importMode = ref<'append' | 'overwrite'>('append')
const importing = ref(false)

function saveAiKey() {
  const k = aiKey.value.trim()
  if (k) localStorage.setItem('ai_api_key', k)
  else localStorage.removeItem('ai_api_key')
  ElMessage.success('AI Key 已保存')
}

async function doExportJson() {
  try {
    const data = await exportJson()
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const date = new Date().toISOString().slice(0, 10)
    downloadBlob(blob, `foodlog-backup-${date}.json`)
    ElMessage.success('已导出 JSON 备份')
  } catch {
    /* 拦截器已提示 */
  }
}

async function doExportCsv() {
  try {
    const blob = await exportCsvBlob()
    downloadBlob(blob, 'foodlog-recipes.csv')
    ElMessage.success('已导出 CSV')
  } catch {
    /* 拦截器已提示 */
  }
}

async function onImportFile(file: UploadFile) {
  if (importing.value) return
  const raw = file.raw
  if (!raw) return
  try {
    const text = await raw.text()
    const data = JSON.parse(text)
    if (importMode.value === 'overwrite') {
      await ElMessageBox.confirm('覆盖模式将清空当前全部数据，确定继续？', '危险操作', {
        type: 'warning',
        confirmButtonText: '确定覆盖',
        cancelButtonText: '取消'
      }).catch(() => null)
    }
    importing.value = true
    await importData(data, importMode.value)
    ElMessage.success('导入成功')
    await auth.fetchMe()
  } catch (e) {
    if (e instanceof SyntaxError) {
      ElMessage.error('文件格式不正确，请选择导出的 JSON 备份')
    }
  } finally {
    importing.value = false
  }
}

async function logout() {
  await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' }).catch(() => null)
  auth.logout()
  router.push('/login')
}
</script>

