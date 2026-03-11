<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDatabasesStore } from '../stores/databases'
import { useAppsStore } from '../stores/apps'
import StatusBadge from '../components/StatusBadge.vue'
import ConnectionInfo from '../components/ConnectionInfo.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import { startWebUI, stopWebUI } from '../api/databases'

const route = useRoute()
const router = useRouter()
const dbStore = useDatabasesStore()
const appsStore = useAppsStore()

const dbId = computed(() => route.params.id as string)
const db = computed(() => dbStore.currentDatabase)

const tabs = ['Overview', 'Backups', 'Settings'] as const
type Tab = typeof tabs[number]
const activeTab = ref<Tab>('Overview')

const actionLoading = ref(false)
const showDeleteModal = ref(false)
const showRestoreModal = ref(false)
const restoreBackupId = ref('')
const deleting = ref(false)
const creatingBackup = ref(false)
const linkAppId = ref('')
const webUILoading = ref(false)
const webUIUrl = ref<string | null>(null)

const engineConfig = computed(() => {
  if (!db.value) return { name: '', bg: '', text: '' }
  switch (db.value.engine) {
    case 'postgres':
      return { name: 'PostgreSQL', bg: 'bg-blue-100', text: 'text-blue-700' }
    case 'mysql':
      return { name: 'MySQL', bg: 'bg-orange-100', text: 'text-orange-700' }
    case 'mongodb':
      return { name: 'MongoDB', bg: 'bg-green-100', text: 'text-green-700' }
    case 'redis':
      return { name: 'Redis', bg: 'bg-red-100', text: 'text-red-700' }
    default:
      return { name: db.value.engine, bg: 'bg-gray-100', text: 'text-gray-700' }
  }
})

const linkedApp = computed(() => {
  if (!db.value?.appId) return null
  return appsStore.apps.find(a => a.id === db.value!.appId) || null
})

const storagePercent = computed(() => {
  if (!db.value || !db.value.storageLimit) return 0
  return Math.round((db.value.storageUsed / db.value.storageLimit) * 100)
})

const memoryDisplay = computed(() => {
  if (!db.value) return ''
  return `${db.value.memoryLimit} MB`
})

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString()
}

async function handleStop() {
  actionLoading.value = true
  try {
    await dbStore.stopDatabase(dbId.value)
  } finally {
    actionLoading.value = false
  }
}

async function handleStart() {
  actionLoading.value = true
  try {
    await dbStore.startDatabase(dbId.value)
  } finally {
    actionLoading.value = false
  }
}

async function handleDelete() {
  deleting.value = true
  try {
    await dbStore.deleteDatabase(dbId.value)
    router.push('/databases')
  } finally {
    deleting.value = false
    showDeleteModal.value = false
  }
}

async function handleCreateBackup() {
  creatingBackup.value = true
  try {
    await dbStore.createBackup(dbId.value)
  } finally {
    creatingBackup.value = false
  }
}

async function handleRestore() {
  if (!restoreBackupId.value) return
  actionLoading.value = true
  try {
    await dbStore.restoreBackup(dbId.value, restoreBackupId.value)
    showRestoreModal.value = false
    restoreBackupId.value = ''
  } finally {
    actionLoading.value = false
  }
}

function confirmRestore(backupId: string) {
  restoreBackupId.value = backupId
  showRestoreModal.value = true
}

async function handleStartWebUI() {
  webUILoading.value = true
  try {
    const { data } = await startWebUI(dbId.value)
    webUIUrl.value = data.url
    window.open(data.url, '_blank')
  } finally {
    webUILoading.value = false
  }
}

async function handleStopWebUI() {
  webUILoading.value = true
  try {
    await stopWebUI(dbId.value)
    webUIUrl.value = null
  } finally {
    webUILoading.value = false
  }
}

async function handleLink() {
  if (!linkAppId.value) return
  actionLoading.value = true
  try {
    await dbStore.linkToApp(dbId.value, linkAppId.value)
    linkAppId.value = ''
  } finally {
    actionLoading.value = false
  }
}

async function handleUnlink() {
  actionLoading.value = true
  try {
    await dbStore.unlinkFromApp(dbId.value)
  } finally {
    actionLoading.value = false
  }
}

onMounted(() => {
  dbStore.fetchDatabase(dbId.value)
  dbStore.fetchBackups(dbId.value)
  appsStore.fetchApps()
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Loading -->
    <div v-if="dbStore.loading && !db" class="flex items-center justify-center py-20">
      <div class="text-gray-500">Loading...</div>
    </div>

    <template v-else-if="db">
      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div class="flex items-center gap-3">
          <button
            @click="router.push('/databases')"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-2xl font-bold text-gray-900">{{ db.name }}</h1>
            <div class="flex items-center gap-2 mt-1">
              <StatusBadge :status="db.status" />
              <span
                class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                :class="[engineConfig.bg, engineConfig.text]"
              >
                {{ engineConfig.name }} {{ db.version }}
              </span>
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

      <!-- Tab content -->
      <Transition
        mode="out-in"
        enter-active-class="transition-opacity duration-150"
        leave-active-class="transition-opacity duration-100"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <!-- Overview -->
        <div v-if="activeTab === 'Overview'" key="overview" class="space-y-6">
          <!-- Connection info -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Connection Info</h2>
            <ConnectionInfo :database="db" />
          </div>

          <!-- Quick actions -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
            <div class="flex flex-wrap gap-3">
              <button
                v-if="db.status === 'running'"
                @click="handleStop"
                :disabled="actionLoading"
                class="px-4 py-2 text-sm font-medium text-red-700 bg-red-50 rounded-lg hover:bg-red-100 transition-colors disabled:opacity-50"
              >
                Stop
              </button>
              <button
                v-if="db.status === 'stopped' || db.status === 'failed'"
                @click="handleStart"
                :disabled="actionLoading"
                class="px-4 py-2 text-sm font-medium text-green-700 bg-green-50 rounded-lg hover:bg-green-100 transition-colors disabled:opacity-50"
              >
                Start
              </button>
              <button
                v-if="!webUIUrl"
                @click="handleStartWebUI"
                :disabled="webUILoading || db.status !== 'running'"
                class="px-4 py-2 text-sm font-medium text-indigo-700 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors disabled:opacity-50"
              >
                {{ webUILoading ? 'Starting Web UI...' : 'Open Web UI' }}
              </button>
              <template v-else>
                <a
                  :href="webUIUrl"
                  target="_blank"
                  class="px-4 py-2 text-sm font-medium text-indigo-700 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors inline-block"
                >
                  Open Web UI
                </a>
                <button
                  @click="handleStopWebUI"
                  :disabled="webUILoading"
                  class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
                >
                  {{ webUILoading ? 'Stopping...' : 'Stop Web UI' }}
                </button>
              </template>
              <button
                @click="showDeleteModal = true"
                class="px-4 py-2 text-sm font-medium text-red-700 bg-red-50 rounded-lg hover:bg-red-100 transition-colors"
              >
                Delete
              </button>
            </div>
          </div>

          <!-- Linked app -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Linked Application</h2>
            <div v-if="linkedApp" class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <router-link
                  :to="`/apps/${linkedApp.id}`"
                  class="text-indigo-600 hover:text-indigo-500 font-medium transition-colors"
                >
                  {{ linkedApp.name }}
                </router-link>
                <StatusBadge :status="linkedApp.status" />
              </div>
              <button
                @click="handleUnlink"
                :disabled="actionLoading"
                class="px-3 py-1.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
              >
                Unlink
              </button>
            </div>
            <div v-else-if="db.appId" class="text-sm text-gray-500">
              Linked to app ID: {{ db.appId }}
            </div>
            <div v-else>
              <p class="text-sm text-gray-500 mb-3">No application linked to this database.</p>
              <div class="flex items-center gap-2 max-w-sm">
                <select
                  v-model="linkAppId"
                  class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                >
                  <option value="">Select an app...</option>
                  <option v-for="app in appsStore.apps" :key="app.id" :value="app.id">{{ app.name }}</option>
                </select>
                <button
                  @click="handleLink"
                  :disabled="!linkAppId || actionLoading"
                  class="px-3 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
                >
                  Link
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Backups -->
        <div v-else-if="activeTab === 'Backups'" key="backups" class="space-y-6">
          <div class="bg-white rounded-xl shadow p-6">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-lg font-semibold text-gray-900">Backups</h2>
              <button
                @click="handleCreateBackup"
                :disabled="creatingBackup"
                class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
              >
                {{ creatingBackup ? 'Creating...' : 'Create Backup' }}
              </button>
            </div>

            <div v-if="dbStore.backups.length === 0" class="text-center py-8 text-gray-500">
              <p>No backups yet.</p>
              <p class="text-sm mt-1">Create a backup to protect your data.</p>
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="backup in dbStore.backups"
                :key="backup.id"
                class="flex items-center justify-between p-4 rounded-lg border border-gray-200"
              >
                <div class="flex items-center gap-4 min-w-0">
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ formatDate(backup.createdAt) }}</p>
                    <p class="text-xs text-gray-500">{{ formatBytes(backup.fileSize) }}</p>
                  </div>
                  <StatusBadge :status="backup.status" />
                </div>
                <button
                  v-if="backup.status === 'completed'"
                  @click="confirmRestore(backup.id)"
                  class="px-3 py-1.5 text-sm font-medium text-indigo-700 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors"
                >
                  Restore
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Settings -->
        <div v-else-if="activeTab === 'Settings'" key="settings" class="space-y-6">
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">General</h2>
            <div class="space-y-4 max-w-md">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Database Name</label>
                <input
                  :value="db.name"
                  disabled
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm bg-gray-50 text-gray-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Engine</label>
                <input
                  :value="`${engineConfig.name} ${db.version}`"
                  disabled
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm bg-gray-50 text-gray-500"
                />
              </div>
            </div>
          </div>

          <!-- Resource usage -->
          <div class="bg-white rounded-xl shadow p-6">
            <h2 class="text-lg font-semibold text-gray-900 mb-4">Resource Usage</h2>
            <div class="space-y-4 max-w-md">
              <div>
                <div class="flex items-center justify-between text-sm mb-1">
                  <span class="text-gray-600">Memory Limit</span>
                  <span class="font-medium text-gray-900">{{ memoryDisplay }}</span>
                </div>
              </div>
              <div>
                <div class="flex items-center justify-between text-sm mb-1">
                  <span class="text-gray-600">Storage</span>
                  <span class="font-medium text-gray-900">{{ formatBytes(db.storageUsed) }} / {{ formatBytes(db.storageLimit) }}</span>
                </div>
                <div class="w-full bg-gray-200 rounded-full h-2">
                  <div
                    class="h-2 rounded-full transition-all"
                    :class="storagePercent > 90 ? 'bg-red-500' : storagePercent > 70 ? 'bg-amber-500' : 'bg-indigo-600'"
                    :style="{ width: `${Math.min(storagePercent, 100)}%` }"
                  ></div>
                </div>
                <p class="text-xs text-gray-400 mt-1">{{ storagePercent }}% used</p>
              </div>
            </div>
          </div>

          <!-- Danger zone -->
          <div class="bg-white rounded-xl shadow border border-red-200 p-6">
            <h2 class="text-lg font-semibold text-red-700 mb-2">Danger Zone</h2>
            <p class="text-sm text-gray-600 mb-4">
              Deleting this database will permanently remove all data and backups.
              This action cannot be undone.
            </p>
            <button
              @click="showDeleteModal = true"
              class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
            >
              Delete Database
            </button>
          </div>
        </div>
      </Transition>
    </template>

    <!-- Delete confirmation modal -->
    <ConfirmModal
      v-if="showDeleteModal"
      title="Delete Database"
      :message="`Are you sure you want to delete '${db?.name}'? This will permanently remove all data and backups.`"
      confirm-text="Delete"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteModal = false"
    />

    <!-- Restore confirmation modal -->
    <ConfirmModal
      v-if="showRestoreModal"
      title="Restore Backup"
      message="Are you sure you want to restore this backup? This will overwrite the current database contents."
      confirm-text="Restore"
      :danger="false"
      @confirm="handleRestore"
      @cancel="showRestoreModal = false; restoreBackupId = ''"
    />
  </div>
</template>
