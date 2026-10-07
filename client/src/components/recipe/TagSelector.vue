<template>
  <div>
    <div v-if="selected.length" class="flex flex-wrap gap-1.5 mb-3">
      <el-tag
        v-for="id in selected"
        :key="id"
        closable
        :type="tagsStore.tagColor(id)"
        @close="toggle(id)"
      >
        {{ tagsStore.tagName(id) || '未知标签' }}
      </el-tag>
    </div>

    <div v-for="cat in tagsStore.selectableCategories" :key="cat.id" class="mb-2.5">
      <div class="text-xs text-[#909399] mb-1.5">{{ cat.name }}</div>
      <div class="flex flex-wrap gap-2">
        <el-tag
          v-for="t in cat.tags"
          :key="t.id"
          class="cursor-pointer select-none"
          :effect="selected.includes(t.id) ? 'dark' : 'plain'"
          :type="tagsStore.tagColor(t.id)"
          @click="toggle(t.id)"
        >
          {{ t.name }}
        </el-tag>
      </div>
    </div>

    <div class="mt-3 flex gap-2">
      <el-select v-model="customCategoryId" size="small" class="w-30" placeholder="分类">
        <el-option
          v-for="c in tagsStore.customTargetCategories"
          :key="c.id"
          :label="c.name"
          :value="c.id"
        />
      </el-select>
      <el-input
        v-model="custom"
        placeholder="自定义标签，回车添加"
        size="small"
        clearable
        @keyup.enter="addCustom"
      >
        <template #append>
          <el-button :loading="adding" @click="addCustom">添加</el-button>
        </template>
      </el-input>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { createTag } from '@/api/tags'
import { useTagsStore } from '@/stores/tags'

const props = defineProps<{ modelValue: string[] }>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
}>()

const tagsStore = useTagsStore()
const custom = ref('')
const customCategoryId = ref('')
const adding = ref(false)

const selected = computed<string[]>(() => props.modelValue || [])

onMounted(async () => {
  await tagsStore.ensureLoaded()
  // 默认归入「自定义」分类；若该分类尚不存在，留空交给后端自动创建
  const fallback = tagsStore.selectableCategories.find((c) => c.name === '自定义')
  customCategoryId.value = fallback?.id || ''
})

// 选中/取消；同一 mutex_group 内互斥（如荤菜/素菜）
function toggle(tagId: string) {
  const cur = [...selected.value]
  const idx = cur.indexOf(tagId)
  if (idx >= 0) {
    cur.splice(idx, 1)
    emit('update:modelValue', cur)
    return
  }
  const group = tagsStore.mutexGroupOf(tagId)
  const next = group ? cur.filter((id) => tagsStore.mutexGroupOf(id) !== group) : cur
  next.push(tagId)
  emit('update:modelValue', next)
}

async function addCustom() {
  const name = custom.value.trim()
  if (!name) return

  const existing = tagsStore.allTags.find((t) => t.name === name)
  if (existing) {
    if (!selected.value.includes(existing.id)) toggle(existing.id)
    custom.value = ''
    return
  }

  adding.value = true
  try {
    const created = await createTag({ name, category_id: customCategoryId.value || undefined })
    await tagsStore.reload()
    emit('update:modelValue', [...selected.value, created.id])
    custom.value = ''
    ElMessage.success(`已创建标签「${name}」`)
  } catch {
    /* 失败已由 http 拦截器提示 */
  } finally {
    adding.value = false
  }
}
</script>


