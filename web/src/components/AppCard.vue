<script setup lang="ts">
import type { App } from '../api/apps'

defineProps<{ app: App }>()
</script>

<template>
  <router-link
    :to="`/apps/${app.id}`"
    class="block bg-white rounded-xl shadow hover:shadow-md transition-shadow p-6"
  >
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-lg font-semibold text-gray-900">{{ app.name }}</h3>
      <span
        class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
        :class="app.status === 'running' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'"
      >
        {{ app.status }}
      </span>
    </div>
    <p v-if="app.domain" class="text-sm text-indigo-600">{{ app.domain }}</p>
    <p class="text-xs text-gray-400 mt-2">Created {{ new Date(app.createdAt).toLocaleDateString() }}</p>
  </router-link>
</template>
