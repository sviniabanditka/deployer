<script setup lang="ts">
defineProps<{
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      leave-active-class="transition-opacity duration-150"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="emit('cancel')">
        <Transition
          enter-active-class="transition-all duration-200"
          leave-active-class="transition-all duration-150"
          enter-from-class="opacity-0 scale-95"
          leave-to-class="opacity-0 scale-95"
        >
          <div class="bg-white rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
            <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
            <p class="mt-2 text-sm text-gray-600">{{ message }}</p>
            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="emit('cancel')"
                class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                {{ cancelText || 'Cancel' }}
              </button>
              <button
                @click="emit('confirm')"
                class="px-4 py-2 text-sm font-medium text-white rounded-lg transition-colors"
                :class="danger ? 'bg-red-600 hover:bg-red-700' : 'bg-indigo-600 hover:bg-indigo-700'"
              >
                {{ confirmText || 'Confirm' }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
