<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ loading?: boolean }>()
const emit = defineEmits<{ upload: [file: File] }>()

const dragging = ref(false)
const fileName = ref('')
const fileRef = ref<File | null>(null)

function onDrop(e: DragEvent) {
  dragging.value = false
  const file = e.dataTransfer?.files[0]
  if (file && file.name.endsWith('.zip')) {
    fileRef.value = file
    fileName.value = file.name
  }
}

function onFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    fileRef.value = file
    fileName.value = file.name
  }
}

function handleUpload() {
  if (fileRef.value) {
    emit('upload', fileRef.value)
  }
}
</script>

<template>
  <div>
    <div
      @dragover.prevent="dragging = true"
      @dragleave="dragging = false"
      @drop.prevent="onDrop"
      class="border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer"
      :class="dragging ? 'border-indigo-500 bg-indigo-50' : 'border-gray-300 hover:border-gray-400'"
      @click="($refs.fileInput as HTMLInputElement).click()"
    >
      <input ref="fileInput" type="file" accept=".zip" class="hidden" @change="onFileSelect" />
      <div v-if="fileName" class="text-gray-700">
        <p class="font-medium">{{ fileName }}</p>
        <p class="text-sm text-gray-500 mt-1">Click or drag to replace</p>
      </div>
      <div v-else>
        <p class="text-gray-500">Drag & drop a ZIP file here, or click to browse</p>
        <p class="text-xs text-gray-400 mt-1">Only .zip files accepted</p>
      </div>
    </div>
    <button
      v-if="fileName"
      @click="handleUpload"
      :disabled="loading"
      class="mt-4 px-6 py-2.5 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
    >
      {{ loading ? 'Deploying...' : 'Deploy' }}
    </button>
  </div>
</template>
