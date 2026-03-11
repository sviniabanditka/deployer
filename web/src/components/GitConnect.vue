<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as gitApi from '../api/git'
import type { GitConnection } from '../api/git'

const props = defineProps<{ appId: string }>()

const connection = ref<GitConnection | null>(null)
const loading = ref(false)
const saving = ref(false)
const error = ref('')

const provider = ref<'github' | 'gitlab'>('github')
const repoUrl = ref('')
const branch = ref('main')
const accessToken = ref('')

async function loadConnection() {
  loading.value = true
  try {
    const { data } = await gitApi.getGitConnection(props.appId)
    connection.value = data
  } catch {
    connection.value = null
  } finally {
    loading.value = false
  }
}

async function handleConnect() {
  if (!repoUrl.value.trim()) return
  saving.value = true
  error.value = ''
  try {
    const { data } = await gitApi.connectRepo(props.appId, {
      provider: provider.value,
      repoUrl: repoUrl.value.trim(),
      branch: branch.value.trim() || 'main',
      accessToken: accessToken.value,
    })
    connection.value = data
    repoUrl.value = ''
    accessToken.value = ''
  } catch (e: unknown) {
    const err = e as { response?: { data?: { message?: string } } }
    error.value = err.response?.data?.message || 'Failed to connect repository'
  } finally {
    saving.value = false
  }
}

async function handleDisconnect() {
  saving.value = true
  try {
    await gitApi.disconnectRepo(props.appId)
    connection.value = null
  } finally {
    saving.value = false
  }
}

async function handleToggleAutoDeploy() {
  if (!connection.value) return
  try {
    const { data } = await gitApi.toggleAutoDeploy(props.appId, !connection.value.autoDeploy)
    connection.value = data
  } catch {
    // ignore
  }
}

onMounted(loadConnection)
</script>

<template>
  <div>
    <div v-if="loading" class="text-sm text-gray-500">Loading...</div>

    <!-- Connected state -->
    <div v-else-if="connection" class="space-y-4">
      <div class="bg-green-50 border border-green-200 rounded-lg p-4">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-sm font-medium text-green-800">Repository connected</p>
            <p class="mt-1 text-sm text-green-700">
              <span class="font-medium capitalize">{{ connection.provider }}</span>
              &middot;
              <a :href="connection.repoUrl" target="_blank" class="underline hover:no-underline">
                {{ connection.repoUrl }}
              </a>
            </p>
            <p class="text-xs text-green-600 mt-1">
              Branch: <span class="font-mono">{{ connection.branch }}</span>
            </p>
          </div>
          <button
            @click="handleDisconnect"
            :disabled="saving"
            class="text-sm text-red-600 hover:text-red-700 font-medium disabled:opacity-50"
          >
            Disconnect
          </button>
        </div>
      </div>

      <!-- Auto-deploy toggle -->
      <label class="flex items-center gap-3 cursor-pointer">
        <button
          @click="handleToggleAutoDeploy"
          class="relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none"
          :class="connection.autoDeploy ? 'bg-indigo-600' : 'bg-gray-200'"
          role="switch"
          :aria-checked="connection.autoDeploy"
        >
          <span
            class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow ring-0 transition-transform duration-200"
            :class="connection.autoDeploy ? 'translate-x-5' : 'translate-x-0'"
          ></span>
        </button>
        <span class="text-sm text-gray-700">Auto-deploy on push</span>
      </label>
    </div>

    <!-- Connect form -->
    <div v-else class="space-y-4">
      <!-- Provider selector -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">Provider</label>
        <div class="flex gap-2">
          <button
            @click="provider = 'github'"
            class="px-4 py-2 text-sm font-medium rounded-lg border transition-colors"
            :class="provider === 'github'
              ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
              : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'"
          >
            GitHub
          </button>
          <button
            @click="provider = 'gitlab'"
            class="px-4 py-2 text-sm font-medium rounded-lg border transition-colors"
            :class="provider === 'gitlab'
              ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
              : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'"
          >
            GitLab
          </button>
        </div>
      </div>

      <!-- Repo URL -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Repository URL</label>
        <input
          v-model="repoUrl"
          type="url"
          placeholder="https://github.com/user/repo"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        />
      </div>

      <!-- Branch -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Branch</label>
        <input
          v-model="branch"
          type="text"
          placeholder="main"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        />
      </div>

      <!-- Access token -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Access Token</label>
        <input
          v-model="accessToken"
          type="password"
          placeholder="ghp_xxxxxxxxxxxx"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
        />
        <p class="mt-1 text-xs text-gray-500">Token needs repo read access. Stored encrypted.</p>
      </div>

      <!-- Error -->
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

      <!-- Submit -->
      <button
        @click="handleConnect"
        :disabled="saving || !repoUrl.trim()"
        class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
      >
        {{ saving ? 'Connecting...' : 'Connect Repository' }}
      </button>
    </div>
  </div>
</template>
