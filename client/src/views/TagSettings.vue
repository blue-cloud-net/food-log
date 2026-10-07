<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">标签管理</h2>
      <el-button :loading="loading" @click="reload">刷新</el-button>
    </div>

    <el-alert
      class="mb-4"
      type="info"
      :closable="false"
      show-icon
      title="全局预设标签由系统维护、只读；你可以新增自己的分类与标签，用于菜谱的手动标记。"
    />

    <div class="grid gap-4 md:grid-cols-2">
      <!-- 我的分类 -->
      <div class="fl-card">
        <div class="flex items-center justify-between mb-3">
          <span class="font-medium">我的分类</span>
          <el-button size="small" type="primary" @click="openCategoryDialog()">新增分类</el-button>
        </div>
        <el-table v-if="myCategories.length" :data="myCategories" size="small">
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column label="颜色" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="normalizeColor(row.color)">●</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="标签数" width="80">
            <template #default="{ row }">{{ row.tags.length }}</template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openCategoryDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="removeCategory(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="还没有自定义分类" :image-size="60" />
      </div>

      <!-- 我的标签 -->
      <div class="fl-card">
        <div class="flex items-center justify-between mb-3">
          <span class="font-medium">我的标签</span>
          <el-button size="small" type="primary" @click="openTagDialog()">新增标签</el-button>
        </div>
        <el-table v-if="myTags.length" :data="myTags" size="small">
          <el-table-column prop="name" label="名称" min-width="110" />
          <el-table-column label="分类" min-width="90">
            <template #default="{ row }">{{ categoryNameOf(row.category_id) }}</template>
          </el-table-column>
          <el-table-column label="互斥组" min-width="80">
            <template #default="{ row }">
              <span v-if="row.mutex_group" class="text-xs text-[#909399]">{{ row.mutex_group }}</span>
              <span v-else>—</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openTagDialog(row)">编辑</el-button>
              <el-button link type="danger" @click="removeTag(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="还没有自定义标签" :image-size="60" />
      </div>
    </div>

    <!-- 全局预设（只读） -->
    <div class="fl-card mt-4">
      <div class="font-medium mb-3">全局预设标签（只读）</div>
      <div v-for="c in systemCategories" :key="c.id" class="mb-3 last:mb-0">
        <div class="text-xs text-[#909399] mb-1.5">{{ c.name }}</div>
        <div class="flex flex-wrap gap-1.5">
          <el-tag
            v-for="t in c.tags"
            :key="t.id"
            size="small"
            effect="plain"
            :type="normalizeColor(c.color)"
          >
            {{ t.name }}
          </el-tag>
          <span v-if="!c.tags.length" class="text-xs text-[#c0c4cc]">—</span>
        </div>
      </div>
    </div>

    <!-- 分类弹窗 -->
    <el-dialog
      v-model="categoryDialog"
      :title="categoryForm.id ? '编辑分类' : '新增分类'"
      width="min(420px, 92vw)"
    >
      <el-form label-width="72px" @submit.prevent>
        <el-form-item label="名称">
          <el-input v-model="categoryForm.name" maxlength="50" show-word-limit placeholder="如：我的口味" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-select v-model="categoryForm.color" style="width: 100%">
            <el-option v-for="c in COLORS" :key="c" :label="c" :value="c">
              <el-tag size="small" :type="normalizeColor(c)">{{ c }}</el-tag>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveCategory">保存</el-button>
      </template>
    </el-dialog>

    <!-- 标签弹窗 -->
    <el-dialog v-model="tagDialog" :title="tagForm.id ? '编辑标签' : '新增标签'" width="min(420px, 92vw)">
      <el-form label-width="72px" @submit.prevent>
        <el-form-item label="名称">
          <el-input v-model="tagForm.name" maxlength="50" show-word-limit placeholder="如：外婆菜" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="tagForm.category_id" style="width: 100%" placeholder="选择分类">
            <el-option
              v-for="c in tagsStore.customTargetCategories"
              :key="c.id"
              :label="c.name"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="互斥组">
          <el-input v-model="tagForm.mutex_group" maxlength="32" placeholder="可选；同组标签只能选一个" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tagDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTag">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createTag,
  createTagCategory,
  deleteTag,
  deleteTagCategory,
  updateTag,
  updateTagCategory
} from '@/api/tags'
import type { Tag, TagCategory } from '@/api/types'
import { useTagsStore } from '@/stores/tags'
import { normalizeColor } from '@/utils/tags'

const COLORS = ['primary', 'success', 'warning', 'danger', 'info']

const tagsStore = useTagsStore()
const saving = ref(false)
const loading = ref(false)

const myCategories = computed(() => tagsStore.categories.filter((c) => !c.is_system))
const systemCategories = computed(() => tagsStore.categories.filter((c) => c.is_system))
const myTags = computed(() => tagsStore.allTags.filter((t) => !t.is_system))

function categoryNameOf(id: string) {
  return tagsStore.categoryById[id]?.name ?? '—'
}

async function reload() {
  loading.value = true
  try {
    await tagsStore.reload()
  } finally {
    loading.value = false
  }
}

onMounted(reload)

// ===== 分类 =====
const categoryDialog = ref(false)
const categoryForm = reactive({ id: '', name: '', color: 'info' })

function openCategoryDialog(row?: TagCategory) {
  categoryForm.id = row?.id ?? ''
  categoryForm.name = row?.name ?? ''
  categoryForm.color = row?.color ?? 'info'
  categoryDialog.value = true
}

async function saveCategory() {
  if (!categoryForm.name.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  saving.value = true
  try {
    if (categoryForm.id) {
      await updateTagCategory(categoryForm.id, { name: categoryForm.name, color: categoryForm.color })
    } else {
      await createTagCategory({ name: categoryForm.name, color: categoryForm.color })
    }
    categoryDialog.value = false
    await tagsStore.reload()
    ElMessage.success('已保存')
  } catch {
    /* 失败已由拦截器提示 */
  } finally {
    saving.value = false
  }
}

async function removeCategory(row: TagCategory) {
  try {
    await ElMessageBox.confirm(
      `删除分类「${row.name}」会同时删除其下 ${row.tags.length} 个标签，并从相关菜谱中移除这些标签。确定继续？`,
      '删除确认',
      { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    await deleteTagCategory(row.id)
    await tagsStore.reload()
    ElMessage.success('已删除')
  } catch {
    /* 失败已由拦截器提示 */
  }
}

// ===== 标签 =====
const tagDialog = ref(false)
const tagForm = reactive({ id: '', name: '', category_id: '', mutex_group: '' })

function openTagDialog(row?: Tag) {
  tagForm.id = row?.id ?? ''
  tagForm.name = row?.name ?? ''
  tagForm.category_id = row?.category_id ?? myCategories.value[0]?.id ?? ''
  tagForm.mutex_group = row?.mutex_group ?? ''
  tagDialog.value = true
}

async function saveTag() {
  if (!tagForm.name.trim()) {
    ElMessage.warning('请输入标签名称')
    return
  }
  saving.value = true
  try {
    if (tagForm.id) {
      await updateTag(tagForm.id, { name: tagForm.name, mutex_group: tagForm.mutex_group })
    } else {
      await createTag({
        name: tagForm.name,
        category_id: tagForm.category_id || undefined,
        mutex_group: tagForm.mutex_group || undefined
      })
    }
    tagDialog.value = false
    await tagsStore.reload()
    ElMessage.success('已保存')
  } catch {
    /* 失败已由拦截器提示 */
  } finally {
    saving.value = false
  }
}

async function removeTag(row: Tag) {
  try {
    await ElMessageBox.confirm(`删除标签「${row.name}」会从相关菜谱中移除它。确定继续？`, '删除确认', {
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await deleteTag(row.id)
    await tagsStore.reload()
    ElMessage.success('已删除')
  } catch {
    /* 失败已由拦截器提示 */
  }
}
</script>
