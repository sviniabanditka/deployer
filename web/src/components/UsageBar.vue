<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  current: number
  limit: number
  unit: string
  unlimited?: boolean
}>()

const percentage = computed(() => {
  if (props.unlimited || props.limit <= 0) return 0
  return Math.min((props.current / props.limit) * 100, 100)
})

const barColor = computed(() => {
  if (props.unlimited) return 'bg-indigo-500'
  if (percentage.value >= 90) return 'bg-red-500'
  if (percentage.value >= 70) return 'bg-yellow-500'
  return 'bg-green-500'
})

const displayValue = computed(() => {
  if (props.unlimited) return `${props.current} (unlimited)`
  return `${props.current} / ${props.limit} ${props.unit}`
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-1.5">
      <span class="text-sm font-medium text-gray-700">{{ label }}</span>
      <span class="text-sm text-gray-500">{{ displayValue }}</span>
    </div>
    <div class="h-2.5 w-full rounded-full bg-gray-200">
      <div
        class="h-2.5 rounded-full transition-all duration-500"
        :class="barColor"
        :style="{ width: unlimited ? '15%' : `${percentage}%` }"
      />
    </div>
  </div>
</template>
