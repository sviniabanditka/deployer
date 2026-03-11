<script setup lang="ts">
import { ref } from 'vue'
import StatusBadge from './StatusBadge.vue'
import ConfirmModal from './ConfirmModal.vue'
import { rollbackApp } from '../api/apps'
import type { Deployment } from '../api/apps'

const props = defineProps<{ deployments: Deployment[] }>()
const emit = defineEmits<{ (e: 'rollback'): void }>()

const expandedId = ref<string | null>(null)
const showPreview = ref(true)
const rollbackTarget = ref<Deployment | null>(null)
const rollingBack = ref(false)

const filteredDeployments = computed(() => {
  if (showPreview.value) return props.deployments
  return props.deployments.filter(d => !d.isPreview)
})

function toggle(id: string) {
  expandedId.value = expandedId.value === id ? null : id
}

function timeAgo(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = now - then
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  return new Date(dateStr).toLocaleDateString()
}

function canRollback(dep: Deployment): boolean {
  return !dep.isCurrent && dep.status === 'success'
}

async function handleRollback() {
  if (!rollbackTarget.value) return
  rollingBack.value = true
  try {
    await rollbackApp(rollbackTarget.value.appId, rollbackTarget.value.id)
    emit('rollback')
  } finally {
    rollingBack.value = false
    rollbackTarget.value = null
  }
}
</script>

<script lang="ts">
import { computed } from 'vue'
</script>

<template>
  <div>
    <!-- Filter toggle -->
    <div class="flex items-center justify-end mb-3">
      <label class="flex items-center gap-2 text-sm text-gray-600 cursor-pointer select-none">
        <input
          type="checkbox"
          v-model="showPreview"
          class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
        />
        Show previews
      </label>
    </div>

    <div v-if="filteredDeployments.length === 0" class="text-sm text-gray-500 py-8 text-center">
      No deployments yet. Deploy your first version to get started.
    </div>

    <div v-else class="divide-y divide-gray-200">
      <div v-for="dep in filteredDeployments" :key="dep.id">
        <div
          class="flex items-center justify-between py-3 px-2 hover:bg-gray-50 cursor-pointer rounded-lg transition-colors"
          @click="toggle(dep.id)"
        >
          <div class="flex items-center gap-4 min-w-0">
            <span class="text-sm font-mono text-gray-700 shrink-0">
              {{ dep.version || 'v?' }}
            </span>
            <StatusBadge :status="dep.status" />
            <!-- Preview badge -->
            <span
              v-if="dep.isPreview"
              class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-purple-100 text-purple-700"
            >
              Preview
            </span>
            <!-- Current badge -->
            <span
              v-if="dep.isCurrent"
              class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700"
            >
              Current
            </span>
            <!-- PR link -->
            <a
              v-if="dep.prNumber && dep.prUrl"
              :href="dep.prUrl"
              target="_blank"
              class="text-xs text-indigo-600 hover:text-indigo-500 font-medium"
              @click.stop
            >
              PR #{{ dep.prNumber }}
            </a>
            <span v-if="dep.imageTag" class="text-xs text-gray-400 font-mono truncate hidden sm:inline">
              {{ dep.imageTag }}
            </span>
          </div>
          <div class="flex items-center gap-3 shrink-0">
            <!-- Preview URL -->
            <a
              v-if="dep.isPreview && dep.previewUrl"
              :href="dep.previewUrl"
              target="_blank"
              class="text-xs text-indigo-600 hover:text-indigo-500 font-medium hidden sm:inline"
              @click.stop
            >
              Preview URL
            </a>
            <span class="text-xs text-gray-500">{{ timeAgo(dep.createdAt) }}</span>
            <svg
              class="w-4 h-4 text-gray-400 transition-transform duration-200"
              :class="expandedId === dep.id ? 'rotate-180' : ''"
              fill="none" stroke="currentColor" viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>

        <!-- Expandable log -->
        <Transition
          enter-active-class="transition-all duration-200 ease-out"
          leave-active-class="transition-all duration-150 ease-in"
          enter-from-class="opacity-0 max-h-0"
          enter-to-class="opacity-100 max-h-96"
          leave-from-class="opacity-100 max-h-96"
          leave-to-class="opacity-0 max-h-0"
        >
          <div v-if="expandedId === dep.id" class="overflow-hidden">
            <div class="mx-2 mb-3 bg-gray-900 rounded-lg p-4 max-h-64 overflow-auto">
              <pre v-if="dep.buildLog" class="text-xs text-gray-300 font-mono whitespace-pre-wrap">{{ dep.buildLog }}</pre>
              <p v-else class="text-xs text-gray-600 font-mono">No build log available.</p>
            </div>
            <!-- Preview URL in expanded view -->
            <div v-if="dep.isPreview && dep.previewUrl" class="mx-2 mb-2">
              <a
                :href="dep.previewUrl"
                target="_blank"
                class="text-xs text-indigo-600 hover:text-indigo-500 font-medium"
              >
                Open preview: {{ dep.previewUrl }}
              </a>
            </div>
            <div class="mx-2 mb-3">
              <button
                v-if="canRollback(dep)"
                @click.stop="rollbackTarget = dep"
                class="text-xs text-indigo-600 hover:text-indigo-700 font-medium"
              >
                Rollback to this version
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- Rollback confirmation modal -->
    <ConfirmModal
      v-if="rollbackTarget"
      title="Rollback Deployment"
      :message="`Are you sure you want to rollback to version ${rollbackTarget.version || rollbackTarget.id}? This will replace the current running deployment.`"
      :confirm-text="rollingBack ? 'Rolling back...' : 'Rollback'"
      :danger="false"
      @confirm="handleRollback"
      @cancel="rollbackTarget = null"
    />
  </div>
</template>
