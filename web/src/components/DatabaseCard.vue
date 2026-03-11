<script setup lang="ts">
import { computed } from 'vue'
import type { ManagedDatabase } from '../api/databases'
import StatusBadge from './StatusBadge.vue'

const props = defineProps<{ database: ManagedDatabase }>()

const engineConfig = computed(() => {
  switch (props.database.engine) {
    case 'postgres':
      return { label: 'PG', name: 'PostgreSQL', bg: 'bg-blue-100', text: 'text-blue-700', border: 'border-blue-200' }
    case 'mysql':
      return { label: 'MY', name: 'MySQL', bg: 'bg-orange-100', text: 'text-orange-700', border: 'border-orange-200' }
    case 'mongodb':
      return { label: 'MO', name: 'MongoDB', bg: 'bg-green-100', text: 'text-green-700', border: 'border-green-200' }
    case 'redis':
      return { label: 'RD', name: 'Redis', bg: 'bg-red-100', text: 'text-red-700', border: 'border-red-200' }
    default:
      return { label: 'DB', name: props.database.engine, bg: 'bg-gray-100', text: 'text-gray-700', border: 'border-gray-200' }
  }
})
</script>

<template>
  <router-link
    :to="`/databases/${database.id}`"
    class="block bg-white rounded-xl shadow hover:shadow-md border border-gray-200 hover:border-indigo-300 transition-all p-5"
  >
    <div class="flex items-start gap-4">
      <div
        class="flex-shrink-0 w-12 h-12 rounded-lg flex items-center justify-center text-sm font-bold border"
        :class="[engineConfig.bg, engineConfig.text, engineConfig.border]"
      >
        {{ engineConfig.label }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2 mb-1">
          <h3 class="text-base font-semibold text-gray-900 truncate">{{ database.name }}</h3>
          <StatusBadge :status="database.status" />
        </div>
        <p class="text-sm text-gray-500">
          {{ engineConfig.name }} {{ database.version }}
        </p>
        <p class="text-xs text-gray-400 mt-1 font-mono truncate">
          {{ database.host }}:{{ database.port }}
        </p>
      </div>
      <svg class="w-5 h-5 text-gray-300 flex-shrink-0 mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
    </div>
  </router-link>
</template>
