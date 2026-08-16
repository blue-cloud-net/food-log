<template>
  <div>
    <div v-if="selected.length" class="flex flex-wrap gap-1.5 mb-3">
      <el-tag
        v-for="t in selected"
        :key="t"
        closable
        :type="tagColor(categories, t)"
        @close="toggle(t)"
      >
        {{ t }}
      </el-tag>
    </div>

    <div v-for="cat in categories" :key="cat.name" class="mb-2.5">
      <div class="text-xs text-[#909399] mb-1.5">{{ cat.name }}</div>
      <div class="flex flex-wrap gap-2">
        <el-tag
          v-for="t in cat.tags"
          :key="t"
          class="cursor-pointer select-none"
          :effect="selected.includes(t) ? 'dark' : 'plain'"
          :type="tagColor(categories, t)"
          @click="toggle(t)"
        >
          {{ t }}
        </el-tag>
      </div>
    </div>

    <div class="mt-3">
      <el-input
        v-model="custom"
        placeholder="自定义标签，回车添加"
        size="small"
        clearable
        @keyup.enter="addCustom"
      >
        <template #append>
          <el-button @click="addCustom">添加</el-button>
        </template>
      </el-input>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { TagCategory } from '@/api/types'
import { tagColor } from '@/utils/tags'

const props = defineProps<{
  modelValue: string[]
  categories: TagCategory[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void
}>()

const custom = ref('')

const selected = computed<string[]>(() => props.modelValue || [])

function toggle(tag: string) {
  const cur = [...selected.value]
  const idx = cur.indexOf(tag)
  if (idx >= 0) {
    cur.splice(idx, 1)
  } else {
    // 荤素互斥
    if (tag === '荤菜' || tag === '素菜') {
      cur.splice(cur.indexOf(tag === '荤菜' ? '素菜' : '荤菜'), 1)
    }
    cur.push(tag)
  }
  emit('update:modelValue', cur)
}

function addCustom() {
  const t = custom.value.trim()
  if (!t) return
  if (!selected.value.includes(t)) {
    emit('update:modelValue', [...selected.value, t])
  }
  custom.value = ''
}
</script>

