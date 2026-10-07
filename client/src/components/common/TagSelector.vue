<template>
  <div>
    <div v-if="selected.length" class="flex flex-wrap gap-1.5 mb-3">
      <el-tag v-for="id in selected" :key="id" closable :type="tagColor(id)" @close="toggle(id)">
        {{ tagName(id) || '未知标签' }}
      </el-tag>
    </div>

    <div v-for="cat in categories" :key="cat.id" class="mb-2.5">
      <div class="text-xs text-[#909399] mb-1.5">{{ cat.name }}</div>
      <div class="flex flex-wrap gap-2">
        <el-tag
          v-for="t in cat.tags"
          :key="t.id"
          class="cursor-pointer select-none"
          :effect="selected.includes(t.id) ? 'dark' : 'plain'"
          :type="tagColor(t.id)"
          @click="toggle(t.id)"
        >
          {{ t.name }}
        </el-tag>
      </div>
    </div>

    <div class="mt-3 flex gap-2">
      <el-select v-model="customCategoryId" size="small" class="w-30" placeholder="分类">
        <el-option v-for="c in customTargets" :key="c.id" :label="c.name" :value="c.id" />
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
import { createShopTag, createTag, type ShopTagDomain } from '@/api/tags'
import type { Tag, TagCategory } from '@/api/types'
import { useShopTagsStore } from '@/stores/shopTags'
import { useTagsStore } from '@/stores/tags'

// recipe → 菜谱标签字典；restaurant / dish → 探店标签字典（后端为两套独立表）
type TagSelectorDomain = 'recipe' | ShopTagDomain

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    domain?: TagSelectorDomain
  }>(),
  { domain: 'recipe' }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
}>()

const tagsStore = useTagsStore()
const shopTagsStore = useShopTagsStore()

const custom = ref('')
const customCategoryId = ref('')
const adding = ref(false)

const selected = computed<string[]>(() => props.modelValue || [])
const isRecipe = computed(() => props.domain === 'recipe')
const shopDomain = computed(() => props.domain as ShopTagDomain)

const categories = computed<TagCategory[]>(() =>
  isRecipe.value ? tagsStore.selectableCategories : shopTagsStore.selectableCategories(shopDomain.value)
)
const customTargets = computed<TagCategory[]>(() =>
  isRecipe.value
    ? tagsStore.customTargetCategories
    : shopTagsStore.customTargetCategories(shopDomain.value)
)
const allTags = computed<Tag[]>(() =>
  isRecipe.value ? tagsStore.allTags : shopTagsStore.tagsOf(shopDomain.value)
)

function tagName(id: string): string {
  return isRecipe.value ? tagsStore.tagName(id) : shopTagsStore.tagNameOf(shopDomain.value, id)
}

function tagColor(id: string) {
  return isRecipe.value ? tagsStore.tagColor(id) : shopTagsStore.tagColorOf(shopDomain.value, id)
}

function mutexGroupOf(id: string): string {
  return isRecipe.value ? tagsStore.mutexGroupOf(id) : shopTagsStore.mutexGroupOf(shopDomain.value, id)
}

async function ensureLoaded(force = false) {
  if (isRecipe.value) {
    await tagsStore.ensureLoaded(force)
    return
  }
  await shopTagsStore.ensureLoaded(shopDomain.value, force)
}

onMounted(async () => {
  await ensureLoaded()
  // 默认归入「自定义」分类；若该分类尚不存在，留空交给后端自动创建
  const fallback = categories.value.find((c) => c.name === '自定义')
  customCategoryId.value = fallback?.id || ''
})

// 选中/取消；同一 mutex_group 内互斥（如荤菜/素菜、份量足/份量少）
function toggle(tagId: string) {
  const cur = [...selected.value]
  const idx = cur.indexOf(tagId)
  if (idx >= 0) {
    cur.splice(idx, 1)
    emit('update:modelValue', cur)
    return
  }
  const group = mutexGroupOf(tagId)
  const next = group ? cur.filter((id) => mutexGroupOf(id) !== group) : cur
  next.push(tagId)
  emit('update:modelValue', next)
}

async function addCustom() {
  const name = custom.value.trim()
  if (!name) return

  const existing = allTags.value.find((t) => t.name === name)
  if (existing) {
    if (!selected.value.includes(existing.id)) toggle(existing.id)
    custom.value = ''
    return
  }

  adding.value = true
  try {
    const created = isRecipe.value
      ? await createTag({ name, category_id: customCategoryId.value || undefined })
      : await createShopTag(shopDomain.value, {
          name,
          category_id: customCategoryId.value || undefined
        })
    await ensureLoaded(true)
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
