<template>
  <div>
    <!-- 工具条：AI 识别 -->
    <div class="flex items-center gap-3 mb-4 px-3.5 py-2.5 bg-[#fff7f0] rounded-lg border border-dashed border-[#ffc9ad]">
      <el-button type="primary" plain :loading="recognizing" :disabled="!form.images.length" @click="onRecognize">
        <el-icon><MagicStick /></el-icon>&nbsp;AI 识别图片自动填表
      </el-button>
      <span class="text-xs text-[#909399]">上传一张成品图，自动识别菜名 / 食材 / 标签</span>
    </div>

    <el-form label-position="top">
      <el-form-item label="菜名" required>
        <el-input v-model="form.name" maxlength="200" placeholder="如：麻婆豆腐" />
      </el-form-item>

      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="简单介绍这道菜" />
      </el-form-item>

      <el-form-item label="成品图">
        <div class="flex flex-wrap gap-2.5">
          <div v-for="(img, i) in form.images" :key="img" class="relative">
            <el-image :src="img" fit="cover" class="w-[84px] h-[84px] rounded-lg" :preview-src-list="form.images" preview-teleported />
            <el-icon class="absolute -top-1.5 -right-1.5 bg-white rounded-full cursor-pointer text-[#f56c6c] text-lg" @click="form.images.splice(i, 1)"><CircleClose /></el-icon>
          </div>
          <el-upload
            :http-request="customUpload"
            :show-file-list="false"
            accept="image/*"
            multiple
          >
            <div class="w-[84px] h-[84px] border border-dashed border-[#d9d9d9] rounded-lg flex items-center justify-center text-[#909399] cursor-pointer">
              <el-icon :size="22"><Plus /></el-icon>
            </div>
          </el-upload>
        </div>
      </el-form-item>

      <el-form-item label="食材">
        <div class="w-full flex flex-col gap-2.5">
          <div v-for="(ing, i) in form.ingredients" :key="i" class="flex gap-2 items-center">
            <el-input v-model="ing.name" placeholder="食材名" class="grow-[2]" />
            <el-input v-model="ing.amount" placeholder="用量" class="flex-1" />
            <el-select v-model="ing.unit" placeholder="单位" clearable class="flex-1">
              <el-option v-for="u in units" :key="u" :label="u" :value="u" />
            </el-select>
            <el-button text type="danger" @click="form.ingredients.splice(i, 1)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <el-button type="primary" plain @click="addIngredient">
            <el-icon><Plus /></el-icon>&nbsp;添加食材
          </el-button>
        </div>
      </el-form-item>

      <el-form-item label="步骤">
        <div class="w-full flex flex-col gap-2.5">
          <div v-for="(st, i) in form.steps" :key="i" class="flex gap-2 items-start">
            <span class="w-6 h-6 shrink-0 rounded-full bg-primary text-white flex items-center justify-center text-[13px] mt-1">{{ i + 1 }}</span>
            <el-input v-model="st.content" type="textarea" :rows="2" :placeholder="`第 ${i + 1} 步`" />
            <el-button text type="danger" @click="form.steps.splice(i, 1)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <el-button type="primary" plain @click="form.steps.push({ order: form.steps.length + 1, content: '' })">
            <el-icon><Plus /></el-icon>&nbsp;添加步骤
          </el-button>
        </div>
      </el-form-item>

      <el-row :gutter="16">
        <el-col :span="12" :xs="24">
          <el-form-item label="耗时（分钟）">
            <el-input-number v-model="form.cook_time_minutes" :min="0" :max="1440" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12" :xs="24">
          <el-form-item label="难度">
            <el-select v-model="form.difficulty" clearable placeholder="选择难度" style="width: 100%">
              <el-option label="简单" value="easy" />
              <el-option label="中等" value="medium" />
              <el-option label="困难" value="hard" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="评分">
        <el-rate v-model="form.rating" :max="5" />
      </el-form-item>

      <el-form-item label="标签">
        <TagSelector v-model="form.tags" />
        <div v-if="recipe?.ingredient_tags?.length" class="mt-3 w-full">
          <div class="text-xs text-[#909399] mb-1.5">
            自动识别标签（由菜名/食材/耗时推导，保存时自动更新，不可手动编辑）
          </div>
          <div class="flex flex-wrap gap-1.5">
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
        </div>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" size="large" :loading="saving" style="width: 100%" @click="save">
          保存菜谱
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'
import type { Ingredient, Recipe, Step } from '@/api/types'
import { createRecipe, updateRecipe } from '@/api/recipe'
import { recognizeImage } from '@/api/ai'
import { uploadImages } from '@/api/upload'
import { useTagsStore } from '@/stores/tags'
import TagSelector from '@/components/common/TagSelector.vue'

const props = defineProps<{ recipe?: Recipe | null }>()
const emit = defineEmits<{ (e: 'saved'): void }>()

const tagsStore = useTagsStore()
const saving = ref(false)
const recognizing = ref(false)

const units = ['克', '千克', '毫升', '升', '个', '只', '块', '根', '片', '把', '勺', '碗', '适量', '少许']

const form = reactive({
  name: props.recipe?.name || '',
  description: props.recipe?.description || '',
  cook_time_minutes: props.recipe?.cook_time_minutes || 0,
  difficulty: (props.recipe?.difficulty as string) || '',
  rating: props.recipe?.rating || 0,
  tags: [...(props.recipe?.tags || [])] as string[],
  ingredients: (props.recipe?.ingredients || [{ name: '', amount: '', unit: '' }]) as Ingredient[],
  steps: (props.recipe?.steps || [{ order: 1, content: '' }]) as Step[],
  images: [...(props.recipe?.images || [])] as string[]
})

function addIngredient() {
  form.ingredients.push({ name: '', amount: '', unit: '' })
}

async function customUpload(options: UploadRequestOptions) {
  try {
    const res = await uploadImages([options.file], 'recipe')
    const file = res.files[0]
    form.images.push(file.url)
    options.onSuccess(res)
  } catch (e) {
    options.onError(e as never)
  }
}

async function onRecognize() {
  if (!form.images.length) {
    ElMessage.warning('请先上传一张成品图')
    return
  }
  recognizing.value = true
  try {
    const rec = await recognizeImage(form.images[0])
    if (rec.name) form.name = rec.name
    if (rec.ingredients?.length) {
      form.ingredients = rec.ingredients.map((n) => ({ name: n, amount: '', unit: '' }))
    }
    if (rec.tags?.length) {
      form.tags = [...new Set([...form.tags, ...rec.tags])]
    }
    ElMessage.success('识别完成，请核对信息后保存')
  } finally {
    recognizing.value = false
  }
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入菜名')
    return
  }
  saving.value = true
  try {
    const payload: Partial<Recipe> = {
      ...form,
      difficulty: form.difficulty as Recipe['difficulty'],
      ingredients: form.ingredients.filter((i) => i.name.trim()),
      steps: form.steps
        .map((s, i) => ({ ...s, order: i + 1 }))
        .filter((s) => s.content.trim())
    }
    if (props.recipe?.id) {
      await updateRecipe(props.recipe.id, payload)
    } else {
      await createRecipe(payload)
    }
    ElMessage.success('保存成功')
    emit('saved')
  } finally {
    saving.value = false
  }
}
</script>

