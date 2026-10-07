<template>
  <div
    class="bg-white rounded-xl overflow-hidden shadow-sm cursor-pointer transition-transform duration-150 hover:-translate-y-0.5"
    @click="router.push(`/recipes/${recipe.id}`)"
  >
    <div class="relative h-[150px]">
      <el-image
        v-if="cover"
        :src="cover"
        fit="cover"
        class="w-full h-full"
        :preview-src-list="[cover]"
        preview-teleported
      />
      <div v-else class="w-full h-full flex items-center justify-center bg-gradient-to-br from-[#ffe9dc] to-[#fff3ea] text-primary-light">
        <el-icon :size="36"><Bowl /></el-icon>
      </div>
      <div v-if="recipe.is_favorited" class="absolute top-2 right-2 bg-white/90 rounded-full w-6.5 h-6.5 flex items-center justify-center">
        <el-icon color="#ff6b35"><StarFilled /></el-icon>
      </div>
    </div>

    <div class="px-3 pt-2.5 pb-3">
      <div class="text-[15px] font-semibold mb-1.5 truncate">{{ recipe.name }}</div>
      <div class="flex items-center gap-2 text-xs text-[#909399] mb-1.5">
        <span v-if="recipe.cook_time_minutes" class="flex items-center gap-0.5">
          <el-icon><Timer /></el-icon> {{ recipe.cook_time_minutes }} 分钟
        </span>
        <el-tag
          v-if="recipe.difficulty"
          :type="difficultyTagType(recipe.difficulty)"
          size="small"
          effect="plain"
        >
          {{ difficultyText(recipe.difficulty) }}
        </el-tag>
        <RatingStars v-if="recipe.rating" :model-value="recipe.rating" disabled />
      </div>
      <div v-if="recipe.tags.length" class="flex flex-wrap gap-1.5">
        <el-tag
          v-for="id in recipe.tags.slice(0, 4)"
          :key="id"
          size="small"
          :type="tagsStore.tagColor(id)"
          effect="light"
        >
          {{ tagsStore.tagName(id) }}
        </el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Recipe } from '@/api/types'
import { useTagsStore } from '@/stores/tags'
import { difficultyTagType, difficultyText } from '@/utils/format'
import RatingStars from './RatingStars.vue'

const props = defineProps<{ recipe: Recipe }>()
const router = useRouter()
const tagsStore = useTagsStore()

const cover = computed(() => {
  const img = props.recipe.images?.[0]
  if (!img) return ''
  // 优先缩略图
  return img.replace(/\.(jpg|jpeg|png|webp|gif)$/i, '_thumb.$1')
})
</script>

