<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string }>()

const config = computed(() => {
  switch (props.status) {
    case 'running':
      return { bg: 'bg-green-100', text: 'text-green-800', dot: 'bg-green-500', icon: '' }
    case 'building':
      return { bg: 'bg-amber-100', text: 'text-amber-800', dot: 'bg-amber-500', icon: '' }
    case 'stopped':
      return { bg: 'bg-red-100', text: 'text-red-800', dot: 'bg-red-500', icon: '' }
    case 'failed':
      return { bg: 'bg-red-100', text: 'text-red-700', dot: 'bg-red-600', icon: '!' }
    default:
      return { bg: 'bg-gray-100', text: 'text-gray-700', dot: 'bg-gray-400', icon: '' }
  }
})
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium"
    :class="[config.bg, config.text]"
  >
    <span class="relative flex h-2 w-2">
      <span
        v-if="status === 'running' || status === 'building'"
        class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75"
        :class="config.dot"
      ></span>
      <span class="relative inline-flex rounded-full h-2 w-2" :class="config.dot"></span>
    </span>
    <span v-if="config.icon" class="font-bold">{{ config.icon }}</span>
    {{ status }}
  </span>
</template>
