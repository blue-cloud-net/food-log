<template>
  <el-input
    v-model="keyword"
    placeholder="搜索菜谱、餐厅、菜品…"
    clearable
    @keyup.enter="doSearch"
    @clear="keyword = ''"
  >
    <template #prefix>
      <el-icon><Search /></el-icon>
    </template>
  </el-input>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const keyword = ref('')
const router = useRouter()

function doSearch() {
  const k = keyword.value.trim()
  if (!k) return
  router.push({ name: 'search', query: { keyword: k } })
}
</script>
