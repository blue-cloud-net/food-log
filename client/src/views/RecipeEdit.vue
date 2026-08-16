<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ isEdit ? '编辑菜谱' : '新建菜谱' }}</h2>
      <el-button @click="router.back()">返回</el-button>
    </div>

    <div class="fl-card">
      <RecipeForm :recipe="recipe" @saved="onSaved" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getRecipe } from '@/api/recipe'
import type { Recipe } from '@/api/types'
import RecipeForm from '@/components/recipe/RecipeForm.vue'

const route = useRoute()
const router = useRouter()
const isEdit = !!route.params.id
const recipe = ref<Recipe | null>(null)

onMounted(async () => {
  if (isEdit) {
    recipe.value = await getRecipe(route.params.id as string)
  }
})

function onSaved() {
  if (isEdit) {
    router.push(`/recipes/${route.params.id}`)
  } else {
    router.push('/recipes')
  }
}
</script>
