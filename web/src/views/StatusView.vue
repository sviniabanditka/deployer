<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as statusApi from '../api/status'
import type { SystemStatus, Incident } from '../api/status'

const systemStatus = ref<SystemStatus | null>(null)
const incidents = ref<Incident[]>([])
const loading = ref(true)
const error = ref('')
let refreshTimer: ReturnType<typeof setInterval> | null = null

const overallLabel = computed(() => {
  if (!systemStatus.value) return ''
  switch (systemStatus.value.overall) {
    case 'operational': return 'All Systems Operational'
    case 'degraded': return 'Degraded Performance'
    case 'outage': return 'Major Outage'
    default: return 'Unknown'
  }
})

const overallColor = computed(() => {
  if (!systemStatus.value) return 'bg-gray-100 text-gray-700'
  switch (systemStatus.value.overall) {
    case 'operational': return 'bg-green-50 text-green-700 border-green-200'
    case 'degraded': return 'bg-yellow-50 text-yellow-700 border-yellow-200'
    case 'outage': return 'bg-red-50 text-red-700 border-red-200'
    default: return 'bg-gray-50 text-gray-700 border-gray-200'
  }
})

const overallDot = computed(() => {
  if (!systemStatus.value) return 'bg-gray-400'
  switch (systemStatus.value.overall) {
    case 'operational': return 'bg-green-500'
    case 'degraded': return 'bg-yellow-500'
    case 'outage': return 'bg-red-500'
    default: return 'bg-gray-400'
  }
})

function componentDot(status: string) {
  switch (status) {
    case 'operational': return 'bg-green-500'
    case 'degraded': return 'bg-yellow-500'
    case 'outage': return 'bg-red-500'
    default: return 'bg-gray-400'
  }
}

function severityBadge(severity: string) {
  switch (severity) {
    case 'critical': return 'bg-red-100 text-red-700'
    case 'major': return 'bg-orange-100 text-orange-700'
    case 'minor': return 'bg-yellow-100 text-yellow-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

function incidentStatusBadge(status: string) {
  switch (status) {
    case 'resolved': return 'bg-green-100 text-green-700'
    case 'monitoring': return 'bg-blue-100 text-blue-700'
    case 'identified': return 'bg-yellow-100 text-yellow-700'
    case 'investigating': return 'bg-red-100 text-red-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString()
}

async function fetchData() {
  try {
    const [statusRes, incidentsRes] = await Promise.all([
      statusApi.fetchStatus(),
      statusApi.fetchIncidents(),
    ])
    systemStatus.value = statusRes.data
    incidents.value = incidentsRes.data
    error.value = ''
  } catch {
    error.value = 'Unable to load status information.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
  refreshTimer = setInterval(fetchData, 30000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">System Status</h1>

    <!-- Loading -->
    <div v-if="loading" class="text-gray-500 py-12 text-center">Loading status...</div>

    <!-- Error -->
    <div v-else-if="error" class="bg-red-50 text-red-700 rounded-xl p-6 text-center">
      {{ error }}
    </div>

    <template v-else-if="systemStatus">
      <!-- Overall status banner -->
      <div
        class="rounded-xl border p-6 mb-8 flex items-center gap-4"
        :class="overallColor"
      >
        <span class="w-4 h-4 rounded-full shrink-0" :class="overallDot"></span>
        <span class="text-lg font-semibold">{{ overallLabel }}</span>
      </div>

      <!-- Components -->
      <div class="bg-white rounded-xl shadow mb-8">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-sm font-semibold text-gray-900">Components</h2>
        </div>
        <div class="divide-y divide-gray-100">
          <div
            v-for="comp in systemStatus.components"
            :key="comp.name"
            class="flex items-center justify-between px-6 py-4"
          >
            <div class="flex items-center gap-3">
              <span class="w-2.5 h-2.5 rounded-full shrink-0" :class="componentDot(comp.status)"></span>
              <span class="text-sm font-medium text-gray-900">{{ comp.name }}</span>
            </div>
            <div class="flex items-center gap-4">
              <span v-if="comp.latency" class="text-xs text-gray-500">{{ comp.latency }}ms</span>
              <span class="text-xs capitalize text-gray-500">{{ comp.status }}</span>
            </div>
          </div>
        </div>
        <div class="px-6 py-3 bg-gray-50 rounded-b-xl">
          <p class="text-xs text-gray-400">
            Last updated: {{ formatDate(systemStatus.updatedAt) }}. Auto-refreshes every 30 seconds.
          </p>
        </div>
      </div>

      <!-- Incidents -->
      <div class="bg-white rounded-xl shadow">
        <div class="px-6 py-4 border-b border-gray-200">
          <h2 class="text-sm font-semibold text-gray-900">Incident History</h2>
        </div>
        <div v-if="incidents.length === 0" class="px-6 py-8 text-center text-sm text-gray-500">
          No recent incidents.
        </div>
        <div v-else class="divide-y divide-gray-100">
          <div
            v-for="incident in incidents"
            :key="incident.id"
            class="px-6 py-4"
          >
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="text-sm font-medium text-gray-900">{{ incident.description }}</span>
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="severityBadge(incident.severity)"
                  >
                    {{ incident.severity }}
                  </span>
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="incidentStatusBadge(incident.status)"
                  >
                    {{ incident.status }}
                  </span>
                </div>
                <p class="text-xs text-gray-500 mt-1">
                  {{ incident.component }} &mdash; Started {{ formatDate(incident.startedAt) }}
                  <template v-if="incident.resolvedAt"> &mdash; Resolved {{ formatDate(incident.resolvedAt) }}</template>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
