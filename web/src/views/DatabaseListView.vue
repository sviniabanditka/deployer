<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDatabasesStore } from '../stores/databases'
import { useAppsStore } from '../stores/apps'
import DatabaseCard from '../components/DatabaseCard.vue'

const router = useRouter()
const dbStore = useDatabasesStore()
const appsStore = useAppsStore()

const showCreateModal = ref(false)
const createForm = ref({
  name: '',
  engine: 'postgres',
  version: '',
  appId: '',
})
const creating = ref(false)

const engines = [
  { value: 'postgres', name: 'PostgreSQL', label: 'PG', defaultVersion: '16', bg: 'bg-blue-100', text: 'text-blue-700', border: 'border-blue-300', selectedBorder: 'border-blue-500', ring: 'ring-blue-200' },
  { value: 'mysql', name: 'MySQL', label: 'MY', defaultVersion: '8.0', bg: 'bg-orange-100', text: 'text-orange-700', border: 'border-orange-300', selectedBorder: 'border-orange-500', ring: 'ring-orange-200' },
  { value: 'mongodb', name: 'MongoDB', label: 'MO', defaultVersion: '7.0', bg: 'bg-green-100', text: 'text-green-700', border: 'border-green-300', selectedBorder: 'border-green-500', ring: 'ring-green-200' },
  { value: 'redis', name: 'Redis', label: 'RD', defaultVersion: '7.2', bg: 'bg-red-100', text: 'text-red-700', border: 'border-red-300', selectedBorder: 'border-red-500', ring: 'ring-red-200' },
]

const versionOptions: Record<string, string[]> = {
  postgres: ['16', '15', '14', '13'],
  mysql: ['8.0', '5.7'],
  mongodb: ['7.0', '6.0', '5.0'],
  redis: ['7.2', '7.0', '6.2'],
}

function selectEngine(engine: string) {
  createForm.value.engine = engine
  const eng = engines.find(e => e.value === engine)
  createForm.value.version = eng?.defaultVersion || ''
}

function openCreateModal() {
  createForm.value = { name: '', engine: 'postgres', version: '16', appId: '' }
  showCreateModal.value = true
}

async function handleCreate() {
  if (!createForm.value.name.trim()) return
  creating.value = true
  try {
    const db = await dbStore.createDatabase({
      name: createForm.value.name.trim(),
      engine: createForm.value.engine,
      version: createForm.value.version || undefined,
      appId: createForm.value.appId || undefined,
    })
    showCreateModal.value = false
    router.push(`/databases/${db.id}`)
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  dbStore.fetchDatabases()
  appsStore.fetchApps()
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Databases</h1>
      <button
        @click="openCreateModal"
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
      >
        + Create Database
      </button>
    </div>

    <!-- Loading -->
    <div v-if="dbStore.loading" class="text-gray-500 py-12 text-center">Loading...</div>

    <!-- Empty state -->
    <div v-else-if="dbStore.databases.length === 0" class="text-center py-20">
      <div class="mx-auto w-16 h-16 rounded-full bg-indigo-100 flex items-center justify-center mb-4">
        <svg class="w-8 h-8 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
        </svg>
      </div>
      <h2 class="text-lg font-semibold text-gray-900">No databases yet</h2>
      <p class="mt-1 text-sm text-gray-500">Create your first managed database to get started.</p>
      <button
        @click="openCreateModal"
        class="mt-4 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
      >
        + Create Database
      </button>
    </div>

    <!-- Database grid -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <DatabaseCard
        v-for="db in dbStore.databases"
        :key="db.id"
        :database="db"
      />
    </div>

    <!-- Create Database Modal -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-200"
        leave-active-class="transition-opacity duration-150"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showCreateModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
          @click.self="showCreateModal = false"
        >
          <div class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">Create Database</h3>

            <div class="space-y-4">
              <!-- Name -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
                <input
                  v-model="createForm.name"
                  type="text"
                  placeholder="my-database"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>

              <!-- Engine -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">Engine</label>
                <div class="grid grid-cols-4 gap-2">
                  <button
                    v-for="eng in engines"
                    :key="eng.value"
                    @click="selectEngine(eng.value)"
                    class="flex flex-col items-center p-3 rounded-lg border-2 transition-all"
                    :class="createForm.engine === eng.value
                      ? [eng.bg, eng.text, eng.selectedBorder, `ring-2 ${eng.ring}`]
                      : 'border-gray-200 text-gray-600 hover:border-gray-300'"
                  >
                    <span class="text-lg font-bold">{{ eng.label }}</span>
                    <span class="text-xs mt-0.5">{{ eng.name }}</span>
                  </button>
                </div>
              </div>

              <!-- Version -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Version</label>
                <select
                  v-model="createForm.version"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                >
                  <option v-for="v in versionOptions[createForm.engine]" :key="v" :value="v">{{ v }}</option>
                </select>
              </div>

              <!-- Link to app -->
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Link to App (optional)</label>
                <select
                  v-model="createForm.appId"
                  class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                >
                  <option value="">None</option>
                  <option v-for="app in appsStore.apps" :key="app.id" :value="app.id">{{ app.name }}</option>
                </select>
              </div>
            </div>

            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="showCreateModal = false"
                class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
              <button
                @click="handleCreate"
                :disabled="creating || !createForm.name.trim()"
                class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
              >
                {{ creating ? 'Creating...' : 'Create' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
