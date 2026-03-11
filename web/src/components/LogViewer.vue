<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'

const props = defineProps<{
  appId: string
  autoConnect?: boolean
}>()

const logs = ref<string[]>([])
const connected = ref(false)
const autoScroll = ref(true)
const containerRef = ref<HTMLDivElement | null>(null)

let ws: WebSocket | null = null

function getWsUrl(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/v1/apps/${props.appId}/logs`
}

function connect() {
  if (ws) return
  const token = localStorage.getItem('accessToken')
  const url = token ? `${getWsUrl()}?token=${encodeURIComponent(token)}` : getWsUrl()
  ws = new WebSocket(url)
  ws.onopen = () => {
    connected.value = true
  }
  ws.onmessage = (event) => {
    logs.value.push(event.data)
    if (logs.value.length > 5000) {
      logs.value = logs.value.slice(-3000)
    }
    if (autoScroll.value) {
      nextTick(scrollToBottom)
    }
  }
  ws.onclose = () => {
    connected.value = false
    ws = null
  }
  ws.onerror = () => {
    connected.value = false
  }
}

function disconnect() {
  if (ws) {
    ws.close()
    ws = null
  }
  connected.value = false
}

function scrollToBottom() {
  if (containerRef.value) {
    containerRef.value.scrollTop = containerRef.value.scrollHeight
  }
}

function copyLogs() {
  navigator.clipboard.writeText(logs.value.join('\n'))
}

function clearLogs() {
  logs.value = []
}

watch(() => props.appId, () => {
  disconnect()
  logs.value = []
  if (props.autoConnect !== false) connect()
})

onMounted(() => {
  if (props.autoConnect !== false) connect()
})

onUnmounted(() => {
  disconnect()
})
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-4 py-2 bg-gray-800 rounded-t-lg border-b border-gray-700">
      <div class="flex items-center gap-3">
        <span class="flex items-center gap-1.5 text-xs">
          <span
            class="h-2 w-2 rounded-full"
            :class="connected ? 'bg-green-400' : 'bg-gray-500'"
          ></span>
          <span :class="connected ? 'text-green-400' : 'text-gray-500'">
            {{ connected ? 'Connected' : 'Disconnected' }}
          </span>
        </span>
        <button
          v-if="!connected"
          @click="connect"
          class="text-xs text-indigo-400 hover:text-indigo-300 transition-colors"
        >
          Connect
        </button>
        <button
          v-else
          @click="disconnect"
          class="text-xs text-gray-400 hover:text-gray-300 transition-colors"
        >
          Disconnect
        </button>
      </div>
      <div class="flex items-center gap-2">
        <button
          @click="autoScroll = !autoScroll"
          class="text-xs px-2 py-1 rounded transition-colors"
          :class="autoScroll ? 'bg-indigo-600 text-white' : 'bg-gray-700 text-gray-400 hover:text-gray-300'"
        >
          Auto-scroll
        </button>
        <button
          @click="copyLogs"
          class="text-xs px-2 py-1 rounded bg-gray-700 text-gray-400 hover:text-gray-300 transition-colors"
        >
          Copy
        </button>
        <button
          @click="clearLogs"
          class="text-xs px-2 py-1 rounded bg-gray-700 text-gray-400 hover:text-gray-300 transition-colors"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Log content -->
    <div
      ref="containerRef"
      class="flex-1 bg-gray-900 rounded-b-lg p-4 overflow-auto font-mono text-xs text-gray-300 leading-5 min-h-[320px] max-h-[600px]"
    >
      <div v-if="logs.length === 0" class="text-gray-600">
        Waiting for logs...
      </div>
      <div v-for="(line, i) in logs" :key="i" class="whitespace-pre-wrap break-all hover:bg-gray-800/50">{{ line }}</div>
    </div>
  </div>
</template>
