<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useAppsStore } from '../stores/apps'
import { useDatabasesStore } from '../stores/databases'
import { useBillingStore } from '../stores/billing'
import StatusBadge from '../components/StatusBadge.vue'
import DatabaseCard from '../components/DatabaseCard.vue'

const appsStore = useAppsStore()
const dbStore = useDatabasesStore()
const billingStore = useBillingStore()

const isFreePlan = computed(() => !billingStore.currentPlan || billingStore.currentPlan.name === 'free')

const runningCount = computed(() => appsStore.apps.filter(a => a.status === 'running').length)
const stoppedCount = computed(() => appsStore.apps.filter(a => a.status === 'stopped').length)
const failedCount = computed(() => appsStore.apps.filter(a => a.status === 'failed').length)

const recentApps = computed(() =>
  [...appsStore.apps]
    .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
    .slice(0, 5)
)

const recentDatabases = computed(() =>
  [...dbStore.databases]
    .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
    .slice(0, 3)
)

onMounted(() => {
  appsStore.fetchApps()
  dbStore.fetchDatabases()
  billingStore.fetchSubscription()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
      <router-link
        to="/apps"
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
      >
        + New App
      </router-link>
    </div>

    <!-- Plan banner -->
    <div
      v-if="isFreePlan"
      class="mb-6 flex items-center justify-between rounded-xl border border-indigo-200 bg-indigo-50 px-5 py-4"
    >
      <div class="flex items-center gap-3">
        <svg class="h-5 w-5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        <div>
          <p class="text-sm font-semibold text-indigo-900">You are on the Free plan</p>
          <p class="text-xs text-indigo-600">Upgrade to unlock more apps, databases, and resources.</p>
        </div>
      </div>
      <router-link
        to="/billing"
        class="shrink-0 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 transition-colors"
      >
        Upgrade
      </router-link>
    </div>
    <div
      v-else-if="billingStore.currentPlan"
      class="mb-6 flex items-center justify-between rounded-xl border border-gray-200 bg-white px-5 py-4 shadow-sm"
    >
      <div class="flex items-center gap-3">
        <svg class="h-5 w-5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
        <p class="text-sm font-medium text-gray-700">
          <span class="font-semibold text-gray-900">{{ billingStore.currentPlan.displayName }}</span> plan
        </p>
      </div>
      <router-link
        to="/billing"
        class="text-sm font-medium text-indigo-600 hover:text-indigo-500 transition-colors"
      >
        Manage billing
      </router-link>
    </div>

    <!-- Stats cards -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
      <div class="bg-white rounded-xl shadow p-5">
        <p class="text-sm font-medium text-gray-500">Total Apps</p>
        <p class="mt-1 text-3xl font-bold text-indigo-600">{{ appsStore.apps.length }}</p>
      </div>
      <div class="bg-white rounded-xl shadow p-5">
        <p class="text-sm font-medium text-gray-500">Running</p>
        <p class="mt-1 text-3xl font-bold text-green-600">{{ runningCount }}</p>
      </div>
      <div class="bg-white rounded-xl shadow p-5">
        <p class="text-sm font-medium text-gray-500">Stopped</p>
        <p class="mt-1 text-3xl font-bold text-gray-400">{{ stoppedCount }}</p>
      </div>
      <div class="bg-white rounded-xl shadow p-5">
        <p class="text-sm font-medium text-gray-500">Failed</p>
        <p class="mt-1 text-3xl font-bold text-red-500">{{ failedCount }}</p>
      </div>
      <div class="bg-white rounded-xl shadow p-5">
        <p class="text-sm font-medium text-gray-500">Databases</p>
        <p class="mt-1 text-3xl font-bold text-blue-600">{{ dbStore.databases.length }}</p>
      </div>
    </div>

    <!-- Recent apps -->
    <div class="bg-white rounded-xl shadow p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Recent Apps</h2>
        <router-link to="/apps" class="text-sm text-indigo-600 hover:text-indigo-500 transition-colors">
          View all
        </router-link>
      </div>
      <div v-if="appsStore.loading" class="text-gray-500 py-4">Loading...</div>
      <div v-else-if="appsStore.apps.length === 0" class="text-gray-500 py-8 text-center">
        <p class="text-lg">No applications yet</p>
        <p class="mt-1 text-sm">
          <router-link to="/apps" class="text-indigo-600 hover:text-indigo-500">Create one</router-link>
          to get started.
        </p>
      </div>
      <div v-else class="space-y-2">
        <router-link
          v-for="app in recentApps"
          :key="app.id"
          :to="`/apps/${app.id}`"
          class="flex items-center justify-between p-4 rounded-lg border border-gray-200 hover:border-indigo-300 hover:bg-indigo-50/30 transition-all"
        >
          <div class="flex items-center gap-3 min-w-0">
            <span class="font-medium text-gray-900 truncate">{{ app.name }}</span>
            <StatusBadge :status="app.status" />
          </div>
          <div class="flex items-center gap-3 shrink-0">
            <span v-if="app.domain" class="text-xs text-gray-400 hidden sm:inline">{{ app.domain }}</span>
            <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </router-link>
      </div>
    </div>

    <!-- Recent databases -->
    <div class="bg-white rounded-xl shadow p-6 mt-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Recent Databases</h2>
        <router-link to="/databases" class="text-sm text-indigo-600 hover:text-indigo-500 transition-colors">
          View all
        </router-link>
      </div>
      <div v-if="dbStore.loading" class="text-gray-500 py-4">Loading...</div>
      <div v-else-if="dbStore.databases.length === 0" class="text-gray-500 py-8 text-center">
        <p class="text-lg">No databases yet</p>
        <p class="mt-1 text-sm">
          <router-link to="/databases" class="text-indigo-600 hover:text-indigo-500">Create one</router-link>
          to get started.
        </p>
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <DatabaseCard
          v-for="db in recentDatabases"
          :key="db.id"
          :database="db"
        />
      </div>
    </div>
  </div>
</template>
