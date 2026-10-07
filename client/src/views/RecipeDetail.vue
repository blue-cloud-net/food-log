<template>
  <div class="page-container">
    <div v-if="recipe" class="flex flex-col gap-4">
      <!-- 图片区 -->
      <div class="rounded-xl overflow-hidden">
        <el-image
          v-if="recipe.images?.length"
          :src="recipe.images[0]"
          fit="cover"
          class="w-full h-300px md:h-220px"
          :preview-src-list="recipe.images"
          preview-teleported
        />
        <div v-else class="w-full h-300px md:h-220px flex items-center justify-center bg-gradient-to-br from-[#ffe9dc] to-[#fff3ea] text-primary-light">
          <el-icon :size="60"><Bowl /></el-icon>
        </div>
        <div v-if="recipe.images?.length > 1" class="flex gap-2 p-2 bg-white">
          <el-image
            v-for="img in recipe.images.slice(1)"
            :key="img"
            :src="img"
            fit="cover"
            class="w-18 h-18 rounded-lg"
            :preview-src-list="recipe.images"
            preview-teleported
          />
        </div>
      </div>

      <!-- 头部 -->
      <div class="fl-card">
        <h1 class="m-0 mb-2.5 text-xl md:text-2xl">{{ recipe.name }}</h1>
        <div v-if="recipe.tags.length" class="flex flex-wrap gap-2 mb-2.5">
          <el-tag v-for="id in recipe.tags" :key="id" :type="tagsStore.tagColor(id)" effect="light">
            {{ tagsStore.tagName(id) }}
          </el-tag>
        </div>
        <div v-if="recipe.ingredient_tags?.length" class="flex flex-wrap items-center gap-2 mb-2.5">
          <span class="text-xs text-[#909399]">自动识别</span>
          <el-tag
            v-for="id in recipe.ingredient_tags"
            :key="id"
            size="small"
            effect="plain"
            :type="tagsStore.tagColor(id)"
          >
            {{ tagsStore.tagName(id) }}
          </el-tag>
        </div>
        <div class="flex items-center gap-3 text-[#909399] text-[13px] mb-3.5">
          <el-tag v-if="recipe.difficulty" :type="difficultyTagType(recipe.difficulty)" effect="plain">
            {{ difficultyText(recipe.difficulty) }}
          </el-tag>
          <span v-if="recipe.cook_time_minutes">⏱ {{ recipe.cook_time_minutes }} 分钟</span>
          <span>{{ timeAgo(recipe.created_at) }}</span>
        </div>

        <div class="flex items-center gap-4 flex-wrap border-t border-dashed border-[#f0f0f0] pt-3.5">
          <div class="flex items-center gap-2">
            <span class="text-[13px] text-[#606266]">我的评分</span>
            <el-rate v-model="ratingDraft" @change="saveRating" />
          </div>
          <div class="flex items-center gap-2">
            <span class="text-[13px] text-[#606266]">做过日期</span>
            <el-date-picker
              v-model="madeDraft"
              type="date"
              value-format="YYYY-MM-DD"
              placeholder="未做"
              clearable
              size="small"
              style="width: 150px"
              @change="saveMade"
            />
          </div>
          <el-button
            size="small"
            :type="recipe.is_liked ? 'danger' : 'default'"
            :plain="recipe.is_liked"
            @click="toggleLiked"
          >
            {{ recipe.is_liked ? '❤️ 已喜欢' : '🤍 喜欢' }}
          </el-button>
        </div>

        <div class="flex items-center justify-end flex-wrap gap-2.5 mt-3">
          <el-button
            :type="recipe.is_favorited ? 'warning' : 'default'"
            @click="toggleFavorite"
          >
            <el-icon>
              <StarFilled v-if="recipe.is_favorited" />
              <Star v-else />
            </el-icon>
            &nbsp;{{ recipe.is_favorited ? '已收藏' : '收藏' }}
          </el-button>
          <el-button @click="router.push(`/recipes/${recipe.id}/edit`)">
            <el-icon><Edit /></el-icon>&nbsp;编辑
          </el-button>
          <el-button type="danger" plain @click="remove">
            <el-icon><Delete /></el-icon>&nbsp;删除
          </el-button>
        </div>
      </div>

      <!-- 描述 -->
      <div v-if="recipe.description" class="fl-card text-[#606266] leading-relaxed">
        {{ recipe.description }}
      </div>

      <!-- 食材 -->
      <div class="fl-card">
        <h3 class="m-0 mb-3">🥦 食材清单</h3>
        <div class="grid grid-cols-[repeat(auto-fill,minmax(160px,1fr))] gap-2.5">
          <div v-for="(ing, i) in recipe.ingredients" :key="i" class="flex justify-between items-center bg-[#fafafa] rounded-lg py-2.5 px-3">
            <span class="font-medium">{{ ing.name }}</span>
            <span class="text-[#909399] text-[13px]">{{ ing.amount }}{{ ing.unit }}</span>
          </div>
        </div>
      </div>

      <!-- 做法 + 步骤模式/计时器 -->
      <div class="fl-card">
        <h3 class="m-0 mb-3">👨‍🍳 做法</h3>
        <RecipeSteps :steps="recipe.steps" :cook-time-minutes="recipe.cook_time_minutes" />
      </div>
    </div>
    <div v-else v-loading="true" class="loading-wrap" style="min-height: 200px" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  deleteRecipe,
  favoriteRecipe,
  getRecipe,
  setRecipeLiked,
  setRecipeMade,
  unfavoriteRecipe,
  updateRecipe
} from '@/api/recipe'
import type { Recipe } from '@/api/types'
import { useTagsStore } from '@/stores/tags'
import { difficultyTagType, difficultyText, timeAgo } from '@/utils/format'
import RecipeSteps from '@/components/recipe/RecipeSteps.vue'

const route = useRoute()
const router = useRouter()
const tagsStore = useTagsStore()

const recipe = ref<Recipe | null>(null)
const ratingDraft = ref(0)
const madeDraft = ref<string | null>(null)

async function load() {
  recipe.value = await getRecipe(route.params.id as string)
  // 兼容历史数据中可能为 null 的数组字段
  recipe.value.images = recipe.value.images || []
  recipe.value.steps = recipe.value.steps || []
  recipe.value.ingredients = recipe.value.ingredients || []
  recipe.value.tags = recipe.value.tags || []
  recipe.value.ingredient_tags = recipe.value.ingredient_tags || []
  ratingDraft.value = recipe.value.rating
  madeDraft.value = recipe.value.made_at
}

async function saveRating(v: number) {
  if (!recipe.value) return
  try {
    recipe.value = await updateRecipe(recipe.value.id, { rating: v })
    ElMessage.success('评分已保存')
  } catch {
    ratingDraft.value = recipe.value.rating
  }
}

async function toggleFavorite() {
  if (!recipe.value) return
  if (recipe.value.is_favorited) {
    await unfavoriteRecipe(recipe.value.id)
    recipe.value.is_favorited = false
  } else {
    await favoriteRecipe(recipe.value.id)
    recipe.value.is_favorited = true
  }
}

async function saveMade(v: string | null) {
  if (!recipe.value) return
  try {
    const res = await setRecipeMade(recipe.value.id, v || null)
    recipe.value.made_at = res.made_at
    ElMessage.success(res.made_at ? '已标记为已做' : '已标记为未做')
  } catch {
    madeDraft.value = recipe.value.made_at
  }
}

async function toggleLiked() {
  if (!recipe.value) return
  const next = !recipe.value.is_liked
  await setRecipeLiked(recipe.value.id, next)
  recipe.value.is_liked = next
  ElMessage.success(next ? '已喜欢' : '已取消喜欢')
}

async function remove() {
  if (!recipe.value) return
  await ElMessageBox.confirm(`确定删除「${recipe.value.name}」吗？此操作不可恢复。`, '删除确认', {
    type: 'warning'
  }).catch(() => null)
  await deleteRecipe(recipe.value.id)
  ElMessage.success('已删除')
  router.push('/recipes')
}

onMounted(() => {
  tagsStore.ensureLoaded()
  load()
})
</script>

