<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  envVars: Record<string, string>
}>()

const emit = defineEmits<{
  save: [envVars: Record<string, string>]
}>()

interface EnvRow {
  key: string
  value: string
  visible: boolean
}

const rows = ref<EnvRow[]>([])
const saving = ref(false)
const dirty = ref(false)

function syncFromProps() {
  const entries = Object.entries(props.envVars || {})
  rows.value = entries.map(([key, value]) => ({ key, value, visible: false }))
  dirty.value = false
}

watch(() => props.envVars, syncFromProps, { immediate: true, deep: true })

function addRow() {
  rows.value.push({ key: '', value: '', visible: true })
  dirty.value = true
}

function removeRow(index: number) {
  rows.value.splice(index, 1)
  dirty.value = true
}

function markDirty() {
  dirty.value = true
}

async function handleSave() {
  saving.value = true
  const result: Record<string, string> = {}
  for (const row of rows.value) {
    const k = row.key.trim()
    if (k) {
      result[k] = row.value
    }
  }
  emit('save', result)
  saving.value = false
}
</script>

<template>
  <div class="space-y-3">
    <div v-if="rows.length === 0" class="text-sm text-gray-500 py-4 text-center">
      No environment variables configured.
    </div>

    <div v-for="(row, index) in rows" :key="index" class="flex items-center gap-2">
      <input
        v-model="row.key"
        placeholder="KEY"
        @input="markDirty"
        class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
      />
      <div class="relative flex-1">
        <input
          v-model="row.value"
          :type="row.visible ? 'text' : 'password'"
          placeholder="value"
          @input="markDirty"
          class="w-full px-3 py-2 pr-10 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent bg-white"
        />
        <button
          type="button"
          @click="row.visible = !row.visible"
          class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 text-xs"
        >
          {{ row.visible ? 'Hide' : 'Show' }}
        </button>
      </div>
      <button
        @click="removeRow(index)"
        class="p-2 text-gray-400 hover:text-red-500 transition-colors shrink-0"
        title="Remove"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="flex items-center gap-3 pt-2">
      <button
        @click="addRow"
        class="px-3 py-1.5 text-sm font-medium text-indigo-600 bg-indigo-50 rounded-lg hover:bg-indigo-100 transition-colors"
      >
        + Add variable
      </button>
      <button
        v-if="dirty"
        @click="handleSave"
        :disabled="saving"
        class="px-4 py-1.5 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
      >
        {{ saving ? 'Saving...' : 'Save changes' }}
      </button>
    </div>
  </div>
</template>
