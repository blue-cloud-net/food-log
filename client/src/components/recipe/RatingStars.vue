<template>
  <el-rate
    :model-value="modelValue || 0"
    :disabled="disabled"
    :allow-half="allowHalf"
    size="small"
    @update:model-value="(v: number) => emit('update:modelValue', v)"
  />
</template>

<script setup lang="ts">
defineProps<{
  modelValue: number
  disabled?: boolean
  allowHalf?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
}>()
</script>
