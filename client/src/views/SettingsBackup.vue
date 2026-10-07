<template>
  <div class="page-container flex flex-col gap-4 max-w-160">
    <div class="page-header">
      <h2 class="page-title">数据备份</h2>
      <el-button @click="router.push('/settings')">返回设置</el-button>
    </div>

    <div class="fl-card">
      <h3 class="m-0 mb-2.5">💾 导出</h3>
      <div class="flex gap-2.5 flex-wrap">
        <el-button @click="doExportJson">
          <el-icon><Download /></el-icon>&nbsp;导出 JSON 备份
        </el-button>
        <el-button @click="doExportCsv">
          <el-icon><Download /></el-icon>&nbsp;导出菜谱 CSV
        </el-button>
      </div>
    </div>

    <div class="fl-card">
      <h3 class="m-0 mb-2.5">📥 导入恢复</h3>
      <el-radio-group v-model="importMode" class="mb-3">
        <el-radio value="append">追加（同名跳过）</el-radio>
        <el-radio value="overwrite">覆盖（清空后导入）</el-radio>
      </el-radio-group>
      <div>
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
      </div>
      <el-alert
        v-if="importMode === 'overwrite'"
        title="覆盖模式会清空当前全部数据，请务必先导出备份！"
        type="warning"
        :closable="false"
        class="mt-3"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { exportCsvBlob, exportJson, importData } from '@/api/export'
import { useAuthStore } from '@/stores/auth'
import { downloadBlob } from '@/utils/format'

const auth = useAuthStore()
const router = useRouter()

const importMode = ref<'append' | 'overwrite'>('append')
const importing = ref(false)

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
</script>
