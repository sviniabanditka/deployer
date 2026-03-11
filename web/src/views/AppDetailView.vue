<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppsStore } from '../stores/apps'
import { useDatabasesStore } from '../stores/databases'
import StatusBadge from '../components/StatusBadge.vue'
import StatsChart from '../components/StatsChart.vue'
import DeploymentList from '../components/DeploymentList.vue'
import EnvEditor from '../components/EnvEditor.vue'
import GitConnect from '../components/GitConnect.vue'
import DomainManager from '../components/DomainManager.vue'
import LogViewer from '../components/LogViewer.vue'
import FileUpload from '../components/FileUpload.vue'
import ConfirmModal from '../components/ConfirmModal.vue'

const route = useRoute()
const router = useRouter()
const appsStore = useAppsStore()
const dbStore = useDatabasesStore()

const linkedDatabases = computed(() =>
  dbStore.databases.filter(d => d.appId === appId.value)
)

const appId = computed(() => route.params.id as string)
const app = computed(() => appsStore.currentApp)

const tabs = ['Overview', 'Deployments', 'Env Variables', 'Git', 'Domains', 'Logs', 'Settings'] as const
type Tab = typeof tabs[number]
const activeTab = ref<Tab>('Overview')

const deploying = ref(false)
const actionLoading = ref(false)
const showDeleteModal = ref(false)
const deleting = ref(false)

onMounted(() => {
  appsStore.fetchApp(appId.value)
  appsStore.fetchDeployments(appId.value)
  dbStore.fetchDatabases()
})

async function handleDeploy(file: File) {
  deploying.value = true
  try {
    await appsStore.deployApp(appId.value, file)
    await appsStore.fetchDeployments(appId.value)
  } finally {
    deploying.value = false
  }
}

async function handleStop() {
  actionLoading.value = true
  try {
    await appsStore.stopApp(appId.value)
  } finally {
    actionLoading.value = false
  }
}

async function handleStart() {
  actionLoading.value = true
  try {
    await appsStore.startApp(appId.value)
  } finally {
    actionLoading.value = false
  }
}

async function handleRedeploy() {
  actionLoading.value = true
  try {
    await appsStore.deployApp(appId.value, new File([], ''))
    await appsStore.fetchDeployments(appId.value)
  } finally {
    actionLoading.value = false
  }
}

async function handleSaveEnv(envVars: Record<string, string>) {
  await appsStore.updateEnvVars(appId.value, envVars)
}

async function handleDelete() {
  deleting.value = true
  try {
    await appsStore.deleteApp(appId.value)
    router.push('/apps')
  } finally {
    deleting.value = false
    showDeleteModal.value = false
  }
}
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Loading -->
    <div v-if="appsStore.loading && !app" class="flex items-center justify-center py-20">
      <div class="text-gray-500">Loading...</div>
    </div>

    <template v-else-if="app">
      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div class="flex items-center gap-3">
          <button
            @click="router.push('/apps')"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-2xl font-bold text-gray-900">{{ app.name }}</h1>
            <div class="flex items-center gap-2 mt-1">
              <StatusBadge :status="app.status" />
              <a
                v-if="app.domain"
                :href="`https://${app.domain}`"
                target="_blank"
                class="text-sm text-indigo-600 hover:text-indigo-500 transition-colors"
              >
                {{ app.domain }}
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="border-b border-gray-200 mb-6 -mx-4 px-4 overflow-x-auto">
        <nav class="flex gap-0 min-w-max">
          <button
            v-for="tab in tabs"
            :key="tab"
            @click="activeTab = tab"
            class="px-4 py-3 text-sm font-medium border-b-2 transition-colors whitespace-nowrap"
            :class="activeTab === tab
              ? 'border-indigo-600 text-indigo-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'"
          >
            {{ tab }}
          </button>
        </nav>
      </div>

      <!-- Tab content with transition -->
      <Transition
        mode="out-in"
        enter-active-class="transition-opacity duration-150"
        leave-active-class="transition-opacity duration-100"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <!-- Overview -->
        <div v-if="activeTab === 'Overview'" key="overview" class="space-y-6">
          <!-- Quick actions -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
            <div class="flex flex-wrap gap-3">
              <button
                v-if="app.status === 'running'"
                @click="handleStop"
                :disabled="actionLoading"
                class="px-4 py-2 text-sm font-medium text-red-700 bg-red-50 rounded-lg hover:bg-red-100 transition-colors disabled:opacity-50"
              >
                Stop
              </button>
              <button
                v-if="app.status === 'stopped' || app.status === 'failed'"
                @click="handleStart"
                :disabled="actionLoading"
                class="px-4 py-2 text-sm font-medium text-green-700 bg-green-50 rounded-lg hover:bg-green-100 transition-colors disabled:opacity-50"
              >
                Start
              </button>
              <button
                @click="handleRedeploy"
                :disabled="actionLoading"
                class="px-4 py-2 text-sm font-medium text-indigo-700 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors disabled:opacity-50"
              >
                Redeploy
              </button>
            </div>
          </div>

          <!-- Resource Stats -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Resources</h2>
            <StatsChart :app-id="appId" />
          </div>

          <!-- Deploy from file -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Deploy from Archive</h2>
            <FileUpload @upload="handleDeploy" :loading="deploying" />
          </div>

          <!-- Linked databases -->
          <div class="bg-white rounded-xl shadow p-6">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-lg font-semibold text-gray-900">Databases</h2>
              <router-link
                :to="`/databases?linkApp=${appId}`"
                class="px-3 py-1.5 text-sm font-medium text-indigo-700 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors"
              >
                + New Database
              </router-link>
            </div>
            <div v-if="linkedDatabases.length === 0" class="text-sm text-gray-500 py-4 text-center">
              No databases linked to this app.
            </div>
            <div v-else class="space-y-2">
              <router-link
                v-for="db in linkedDatabases"
                :key="db.id"
                :to="`/databases/${db.id}`"
                class="flex items-center justify-between p-4 rounded-lg border border-gray-200 hover:border-indigo-300 hover:bg-indigo-50/30 transition-all"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <span class="font-medium text-gray-900 truncate">{{ db.name }}</span>
                  <StatusBadge :status="db.status" />
                  <span class="text-xs text-gray-400">{{ db.engine }} {{ db.version }}</span>
                </div>
                <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </router-link>
            </div>
          </div>
        </div>

        <!-- Deployments -->
        <div v-else-if="activeTab === 'Deployments'" key="deployments">
          <div class="bg-white rounded-xl shadow p-6">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-lg font-semibold text-gray-900">Deployment History</h2>
              <button
                @click="appsStore.fetchDeployments(appId)"
                class="text-sm text-indigo-600 hover:text-indigo-500 transition-colors"
              >
                Refresh
              </button>
            </div>
            <DeploymentList :deployments="appsStore.deployments" @rollback="appsStore.fetchDeployments(appId)" />
          </div>
        </div>

        <!-- Env Variables -->
        <div v-else-if="activeTab === 'Env Variables'" key="env">
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Environment Variables</h2>
            <EnvEditor :env-vars="app.envVars || {}" @save="handleSaveEnv" />
          </div>
        </div>

        <!-- Git -->
        <div v-else-if="activeTab === 'Git'" key="git">
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Git Repository</h2>
            <GitConnect :app-id="appId" />
          </div>
        </div>

        <!-- Domains -->
        <div v-else-if="activeTab === 'Domains'" key="domains">
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Custom Domains</h2>
            <DomainManager :app-id="appId" />
          </div>
        </div>

        <!-- Logs -->
        <div v-else-if="activeTab === 'Logs'" key="logs">
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Application Logs</h2>
            <LogViewer :app-id="appId" :auto-connect="true" />
          </div>
        </div>

        <!-- Settings -->
        <div v-else-if="activeTab === 'Settings'" key="settings" class="space-y-6">
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">General</h2>
            <div class="space-y-4 max-w-md">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">App Name</label>
                <input
                  :value="app.name"
                  disabled
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm bg-gray-50 text-gray-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Slug</label>
                <input
                  :value="app.slug || app.id"
                  disabled
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm bg-gray-50 text-gray-500 font-mono"
                />
                <p class="mt-1 text-xs text-gray-400">The slug cannot be changed after creation.</p>
              </div>
            </div>
          </div>

          <!-- Danger zone -->
          <div class="bg-white rounded-xl shadow border border-red-200 p-6">
            <h2 class="text-lg font-semibold text-red-700 mb-2">Danger Zone</h2>
            <p class="text-sm text-gray-600 mb-4">
              Deleting this app will permanently remove all data, deployments, and configurations.
              This action cannot be undone.
            </p>
            <button
              @click="showDeleteModal = true"
              class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
            >
              Delete Application
            </button>
          </div>
        </div>
      </Transition>
    </template>

    <!-- Delete confirmation modal -->
    <ConfirmModal
      v-if="showDeleteModal"
      title="Delete Application"
      :message="`Are you sure you want to delete '${app?.name}'? This will permanently remove all data including deployments, logs, and environment variables.`"
      confirm-text="Delete"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>
