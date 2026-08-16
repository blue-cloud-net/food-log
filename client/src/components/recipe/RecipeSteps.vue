<template>
  <div>
    <!-- 计时器 -->
    <div class="fl-card mb-4">
      <div class="flex justify-between items-center flex-wrap gap-2">
        <span class="font-semibold">⏱️ 烹饪计时器</span>
        <div class="flex items-center gap-1.5">
          <el-input-number v-model="minutes" :min="0" :max="180" size="small" :disabled="running" />
          <span class="text-[13px] text-[#909399]">分钟</span>
          <el-button v-if="!running" size="small" type="primary" :disabled="secondsLeft <= 0" @click="start">
            开始
          </el-button>
          <el-button v-else size="small" type="warning" @click="running = false">暂停</el-button>
          <el-button size="small" @click="reset">重置</el-button>
        </div>
      </div>
      <div
        class="text-[42px] font-bold text-center my-3 tabular-nums text-primary"
        :class="running && secondsLeft < 30 ? 'text-[#f56c6c] animate-[fl-blink_1s_step-start_infinite]' : ''"
      >
        {{ displayTime }}
      </div>
    </div>

    <!-- 步骤模式 -->
    <div class="fl-card">
      <div class="flex justify-between items-center mb-3">
        <span class="font-semibold">👨‍🍳 烹饪步骤</span>
        <el-switch v-model="cookMode" active-text="步骤模式" />
      </div>

      <template v-if="cookMode">
        <div v-if="safeSteps.length" class="text-center">
          <div class="text-[#909399] text-[13px] mb-3">{{ current + 1 }} / {{ safeSteps.length }}</div>
          <div class="min-h-[120px] bg-gradient-to-br from-[#fff7f0] to-white border border-[#ffe3d3] rounded-xl p-6 flex flex-col items-center justify-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-full bg-primary text-white text-xl font-bold flex items-center justify-center">{{ current + 1 }}</div>
            <div class="text-xl leading-relaxed">{{ safeSteps[current].content }}</div>
          </div>
          <div class="flex justify-center gap-3">
            <el-button :disabled="current === 0" @click="current--">
              <el-icon><ArrowLeft /></el-icon>&nbsp;上一步
            </el-button>
            <el-button v-if="current < safeSteps.length - 1" type="primary" @click="current++">
              下一步&nbsp;<el-icon><ArrowRight /></el-icon>
            </el-button>
            <el-button v-else type="success" @click="current = 0">重新开始</el-button>
          </div>
        </div>
        <el-empty v-else description="暂无步骤" :image-size="60" />
      </template>

      <template v-else>
        <ol class="list-none m-0 p-0">
          <li
            v-for="(s, i) in safeSteps"
            :key="i"
            class="flex gap-2.5 py-2 border-b border-dashed border-[#f0f0f0] leading-relaxed last:border-b-0"
          >
            <span class="shrink-0 w-[22px] h-[22px] rounded-full bg-[#fff1e8] text-primary text-xs flex items-center justify-center">{{ i + 1 }}</span>{{ s.content }}
          </li>
        </ol>
        <el-empty v-if="!safeSteps.length" description="暂无步骤" :image-size="60" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElNotification } from 'element-plus'
import type { Step } from '@/api/types'

const props = defineProps<{
  steps: Step[]
  cookTimeMinutes?: number
}>()

const cookMode = ref(false)
const current = ref(0)
const running = ref(false)
const minutes = ref(props.cookTimeMinutes || 5)
const secondsLeft = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const safeSteps = computed<Step[]>(() => props.steps || [])

const displayTime = computed(() => {
  const s = Math.max(0, secondsLeft.value)
  const mm = String(Math.floor(s / 60)).padStart(2, '0')
  const ss = String(s % 60).padStart(2, '0')
  return `${mm}:${ss}`
})

watch(
  () => props.cookTimeMinutes,
  (v) => {
    if (v && !running.value) minutes.value = v
  }
)

function start() {
  if (secondsLeft.value <= 0) secondsLeft.value = minutes.value * 60
  running.value = true
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    secondsLeft.value--
    if (secondsLeft.value <= 0) {
      running.value = false
      if (timer) clearInterval(timer)
      ElNotification({
        title: '⏰ 时间到！',
        message: '计时结束，去看看吧',
        type: 'success'
      })
    }
  }, 1000)
}

function reset() {
  running.value = false
  if (timer) clearInterval(timer)
  timer = null
  secondsLeft.value = 0
}

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<style>
@keyframes fl-blink {
  50% {
    opacity: 0.4;
  }
}
</style>

