<script setup lang="ts">
import { ref } from 'vue'
import type { ManagedDatabase } from '../api/databases'

defineProps<{ database: ManagedDatabase }>()

const showPassword = ref(false)
const showUrl = ref(false)
const copiedField = ref<string | null>(null)

async function copyToClipboard(value: string, field: string) {
  try {
    await navigator.clipboard.writeText(value)
    copiedField.value = field
    setTimeout(() => {
      copiedField.value = null
    }, 2000)
  } catch {
    // fallback: ignore
  }
}
</script>

<template>
  <div class="bg-gray-900 rounded-xl p-5 font-mono text-sm space-y-3">
    <div class="flex items-center justify-between">
      <span class="text-gray-400">Host</span>
      <div class="flex items-center gap-2">
        <span class="text-green-400">{{ database.host }}</span>
        <button
          @click="copyToClipboard(database.host, 'host')"
          class="text-gray-500 hover:text-gray-300 transition-colors"
          :title="copiedField === 'host' ? 'Copied!' : 'Copy'"
        >
          <svg v-if="copiedField !== 'host'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          <svg v-else class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
      </div>
    </div>
    <div class="flex items-center justify-between">
      <span class="text-gray-400">Port</span>
      <div class="flex items-center gap-2">
        <span class="text-green-400">{{ database.port }}</span>
        <button
          @click="copyToClipboard(String(database.port), 'port')"
          class="text-gray-500 hover:text-gray-300 transition-colors"
          :title="copiedField === 'port' ? 'Copied!' : 'Copy'"
        >
          <svg v-if="copiedField !== 'port'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          <svg v-else class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
      </div>
    </div>
    <div class="flex items-center justify-between">
      <span class="text-gray-400">Database</span>
      <div class="flex items-center gap-2">
        <span class="text-green-400">{{ database.dbName }}</span>
        <button
          @click="copyToClipboard(database.dbName, 'dbName')"
          class="text-gray-500 hover:text-gray-300 transition-colors"
          :title="copiedField === 'dbName' ? 'Copied!' : 'Copy'"
        >
          <svg v-if="copiedField !== 'dbName'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          <svg v-else class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
      </div>
    </div>
    <div class="flex items-center justify-between">
      <span class="text-gray-400">Username</span>
      <div class="flex items-center gap-2">
        <span class="text-green-400">{{ database.username }}</span>
        <button
          @click="copyToClipboard(database.username, 'username')"
          class="text-gray-500 hover:text-gray-300 transition-colors"
          :title="copiedField === 'username' ? 'Copied!' : 'Copy'"
        >
          <svg v-if="copiedField !== 'username'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          <svg v-else class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
      </div>
    </div>
    <div class="flex items-center justify-between">
      <span class="text-gray-400">Password</span>
      <div class="flex items-center gap-2">
        <span class="text-green-400">{{ showPassword ? database.password : '••••••••••••' }}</span>
        <button
          @click="showPassword = !showPassword"
          class="text-gray-500 hover:text-gray-300 transition-colors"
          :title="showPassword ? 'Hide' : 'Show'"
        >
          <svg v-if="!showPassword" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
          </svg>
        </button>
        <button
          @click="copyToClipboard(database.password, 'password')"
          class="text-gray-500 hover:text-gray-300 transition-colors"
          :title="copiedField === 'password' ? 'Copied!' : 'Copy'"
        >
          <svg v-if="copiedField !== 'password'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          <svg v-else class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
      </div>
    </div>
    <div class="border-t border-gray-700 pt-3">
      <div class="flex items-center justify-between">
        <span class="text-gray-400">Connection URL</span>
        <div class="flex items-center gap-2">
          <button
            @click="showUrl = !showUrl"
            class="text-gray-500 hover:text-gray-300 transition-colors text-xs"
          >
            {{ showUrl ? 'Hide' : 'Reveal' }}
          </button>
          <button
            @click="copyToClipboard(database.connectionUrl, 'url')"
            class="text-gray-500 hover:text-gray-300 transition-colors"
            :title="copiedField === 'url' ? 'Copied!' : 'Copy'"
          >
            <svg v-if="copiedField !== 'url'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <svg v-else class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </button>
        </div>
      </div>
      <p class="mt-1 text-xs break-all" :class="showUrl ? 'text-green-400' : 'text-gray-600'">
        {{ showUrl ? database.connectionUrl : '••••••••••••••••••••••••••••••••••••••••' }}
      </p>
    </div>
  </div>
</template>
