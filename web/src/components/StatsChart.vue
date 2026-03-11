<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getAppStats } from '../api/apps'

const props = defineProps<{ appId: string }>()

interface Stats {
  cpuPercent: number
  memoryUsedMb: number
  memoryTotalMb: number
  networkInKb: number
  networkOutKb: number
}

const stats = ref<Stats | null>(null)
const error = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

async function refresh() {
  try {
    const { data } = await getAppStats(props.appId)
    stats.value = data
    error.value = false
  } catch {
    error.value = true
  }
}

onMounted(() => {
  refresh()
  timer = setInterval(refresh, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function formatMem(mb: number): string {
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`
  return `${mb.toFixed(0)} MB`
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="!stats && !error" class="text-sm text-gray-500">Loading stats...</div>
    <div v-else-if="error && !stats" class="text-sm text-gray-400">Stats unavailable</div>
    <template v-else-if="stats">
      <!-- CPU -->
      <div>
        <div class="flex items-center justify-between mb-1">
          <span class="text-sm font-medium text-gray-700">CPU</span>
          <span class="text-sm text-gray-500">{{ stats.cpuPercent.toFixed(1) }}%</span>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-2.5">
          <div
            class="h-2.5 rounded-full transition-all duration-500"
            :class="stats.cpuPercent > 80 ? 'bg-red-500' : stats.cpuPercent > 50 ? 'bg-amber-500' : 'bg-indigo-500'"
            :style="{ width: `${Math.min(stats.cpuPercent, 100)}%` }"
          ></div>
        </div>
      </div>

      <!-- Memory -->
      <div>
        <div class="flex items-center justify-between mb-1">
          <span class="text-sm font-medium text-gray-700">Memory</span>
          <span class="text-sm text-gray-500">
            {{ formatMem(stats.memoryUsedMb) }} / {{ formatMem(stats.memoryTotalMb) }}
          </span>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-2.5">
          <div
            class="h-2.5 rounded-full transition-all duration-500"
            :class="(stats.memoryUsedMb / stats.memoryTotalMb) > 0.8 ? 'bg-red-500' : (stats.memoryUsedMb / stats.memoryTotalMb) > 0.5 ? 'bg-amber-500' : 'bg-indigo-500'"
            :style="{ width: `${Math.min((stats.memoryUsedMb / stats.memoryTotalMb) * 100, 100)}%` }"
          ></div>
        </div>
      </div>

      <!-- Network -->
      <div class="flex gap-6">
        <div>
          <span class="text-xs font-medium text-gray-500 uppercase tracking-wide">Network In</span>
          <p class="text-sm font-medium text-gray-900 mt-0.5">{{ (stats.networkInKb / 1024).toFixed(2) }} MB</p>
        </div>
        <div>
          <span class="text-xs font-medium text-gray-500 uppercase tracking-wide">Network Out</span>
          <p class="text-sm font-medium text-gray-900 mt-0.5">{{ (stats.networkOutKb / 1024).toFixed(2) }} MB</p>
        </div>
      </div>
    </template>
  </div>
</template>
