<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as domainsApi from '../api/domains'
import type { CustomDomain } from '../api/domains'

const props = defineProps<{ appId: string }>()

const domains = ref<CustomDomain[]>([])
const newDomain = ref('')
const loading = ref(false)
const addLoading = ref(false)
const verifyingId = ref<string | null>(null)
const removingId = ref<string | null>(null)
const error = ref('')

async function fetchDomains() {
  loading.value = true
  try {
    const { data } = await domainsApi.listDomains(props.appId)
    domains.value = data
  } catch {
    domains.value = []
  } finally {
    loading.value = false
  }
}

async function handleAdd() {
  if (!newDomain.value.trim()) return
  error.value = ''
  addLoading.value = true
  try {
    const { data } = await domainsApi.addDomain(props.appId, newDomain.value.trim())
    domains.value.push(data)
    newDomain.value = ''
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to add domain'
  } finally {
    addLoading.value = false
  }
}

async function handleVerify(domainId: string) {
  verifyingId.value = domainId
  try {
    await domainsApi.verifyDomain(props.appId, domainId)
    await fetchDomains()
  } catch {
    // refresh to show current status
    await fetchDomains()
  } finally {
    verifyingId.value = null
  }
}

async function handleRemove(domainId: string) {
  removingId.value = domainId
  try {
    await domainsApi.removeDomain(props.appId, domainId)
    domains.value = domains.value.filter(d => d.id !== domainId)
  } finally {
    removingId.value = null
  }
}

function statusColor(status: string) {
  switch (status) {
    case 'verified': return 'bg-green-100 text-green-700'
    case 'pending': return 'bg-yellow-100 text-yellow-700'
    case 'failed': return 'bg-red-100 text-red-700'
    default: return 'bg-gray-100 text-gray-700'
  }
}

onMounted(fetchDomains)
</script>

<template>
  <div>
    <!-- Add domain form -->
    <div class="mb-6">
      <div class="flex items-center gap-2 max-w-lg">
        <input
          v-model="newDomain"
          type="text"
          placeholder="example.com"
          class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
          @keydown.enter="handleAdd"
        />
        <button
          @click="handleAdd"
          :disabled="addLoading || !newDomain.trim()"
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
        >
          {{ addLoading ? 'Adding...' : 'Add Domain' }}
        </button>
      </div>
      <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-sm text-gray-500 py-4 text-center">Loading domains...</div>

    <!-- Empty state -->
    <div v-else-if="domains.length === 0" class="text-sm text-gray-500 py-8 text-center">
      No custom domains configured. Add one above.
    </div>

    <!-- Domain list -->
    <div v-else class="space-y-3">
      <div
        v-for="domain in domains"
        :key="domain.id"
        class="border border-gray-200 rounded-lg p-4"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3 min-w-0">
            <svg v-if="domain.status === 'verified'" class="w-5 h-5 text-green-500 shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
            </svg>
            <a
              v-if="domain.status === 'verified'"
              :href="`https://${domain.domain}`"
              target="_blank"
              class="text-sm font-medium text-indigo-600 hover:text-indigo-500 transition-colors truncate"
            >
              {{ domain.domain }}
            </a>
            <span v-else class="text-sm font-medium text-gray-900 truncate">
              {{ domain.domain }}
            </span>
            <span
              class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium shrink-0"
              :class="statusColor(domain.status)"
            >
              {{ domain.status }}
            </span>
          </div>
          <div class="flex items-center gap-2 shrink-0 ml-4">
            <button
              v-if="domain.status === 'pending'"
              @click="handleVerify(domain.id)"
              :disabled="verifyingId === domain.id"
              class="px-3 py-1.5 text-xs font-medium text-indigo-700 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors disabled:opacity-50"
            >
              {{ verifyingId === domain.id ? 'Verifying...' : 'Verify' }}
            </button>
            <button
              @click="handleRemove(domain.id)"
              :disabled="removingId === domain.id"
              class="px-3 py-1.5 text-xs font-medium text-red-700 bg-red-50 rounded-lg hover:bg-red-100 transition-colors disabled:opacity-50"
            >
              {{ removingId === domain.id ? 'Removing...' : 'Remove' }}
            </button>
          </div>
        </div>

        <!-- Verification instructions for pending domains -->
        <div v-if="domain.status === 'pending'" class="mt-3 p-3 bg-yellow-50 rounded-lg">
          <p class="text-xs font-medium text-yellow-800 mb-1">DNS Verification Required</p>
          <p class="text-xs text-yellow-700 mb-2">
            Add the following TXT record to your domain's DNS settings:
          </p>
          <div class="bg-white border border-yellow-200 rounded px-3 py-2">
            <p class="text-xs text-gray-500">Type: <span class="font-mono font-medium text-gray-900">TXT</span></p>
            <p class="text-xs text-gray-500">Name: <span class="font-mono font-medium text-gray-900">_deployer-verify</span></p>
            <p class="text-xs text-gray-500">Value: <span class="font-mono font-medium text-gray-900 break-all">{{ domain.verificationToken }}</span></p>
          </div>
        </div>

        <!-- Failed domain hint -->
        <div v-if="domain.status === 'failed'" class="mt-3 p-3 bg-red-50 rounded-lg">
          <p class="text-xs text-red-700">
            Verification failed. Please check your DNS settings and try again.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
