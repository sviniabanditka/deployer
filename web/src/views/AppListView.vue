<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppsStore } from '../stores/apps'
import AppCard from '../components/AppCard.vue'

const router = useRouter()
const appsStore = useAppsStore()
const showCreate = ref(false)
const newAppName = ref('')
const creating = ref(false)

onMounted(() => {
  appsStore.fetchApps()
})

async function handleCreate() {
  if (!newAppName.value.trim()) return
  creating.value = true
  try {
    const app = await appsStore.createApp(newAppName.value.trim())
    showCreate.value = false
    newAppName.value = ''
    router.push(`/apps/${app.id}`)
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex items-center justify-between mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Applications</h1>
      <button
        @click="showCreate = true"
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
      >
        + Create new
      </button>
    </div>

    <!-- Create modal -->
    <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-md">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Create Application</h2>
        <form @submit.prevent="handleCreate" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">App name</label>
            <input
              v-model="newAppName"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="my-app"
            />
          </div>
          <div class="flex justify-end gap-3">
            <button
              type="button"
              @click="showCreate = false"
              class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="creating"
              class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
            >
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="appsStore.loading" class="text-gray-500">Loading...</div>
    <div v-else-if="appsStore.apps.length === 0" class="text-center py-16 text-gray-500">
      <p class="text-lg">No applications yet</p>
      <p class="mt-1 text-sm">Click "Create new" to deploy your first app.</p>
    </div>
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <AppCard v-for="app in appsStore.apps" :key="app.id" :app="app" />
    </div>
  </div>
</template>
