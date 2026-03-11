<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'

const emit = defineEmits<{
  complete: [code: string]
}>()

const digits = ref<string[]>(['', '', '', '', '', ''])
const inputs = ref<HTMLInputElement[]>([])

function setRef(el: any, index: number) {
  if (el) inputs.value[index] = el
}

function focusInput(index: number) {
  nextTick(() => {
    inputs.value[index]?.focus()
    inputs.value[index]?.select()
  })
}

function handleInput(index: number, event: Event) {
  const target = event.target as HTMLInputElement
  const value = target.value.replace(/\D/g, '')

  if (value.length > 1) {
    // Handle paste into single input
    const chars = value.slice(0, 6 - index).split('')
    chars.forEach((char, i) => {
      if (index + i < 6) {
        digits.value[index + i] = char
      }
    })
    const nextIndex = Math.min(index + chars.length, 5)
    focusInput(nextIndex)
    checkComplete()
    return
  }

  digits.value[index] = value
  if (value && index < 5) {
    focusInput(index + 1)
  }
  checkComplete()
}

function handleKeydown(index: number, event: KeyboardEvent) {
  if (event.key === 'Backspace' && !digits.value[index] && index > 0) {
    digits.value[index - 1] = ''
    focusInput(index - 1)
  }
  if (event.key === 'ArrowLeft' && index > 0) {
    focusInput(index - 1)
  }
  if (event.key === 'ArrowRight' && index < 5) {
    focusInput(index + 1)
  }
}

function handlePaste(event: ClipboardEvent) {
  event.preventDefault()
  const text = event.clipboardData?.getData('text')?.replace(/\D/g, '') || ''
  if (!text) return

  const chars = text.slice(0, 6).split('')
  chars.forEach((char, i) => {
    digits.value[i] = char
  })
  focusInput(Math.min(chars.length, 5))
  checkComplete()
}

function checkComplete() {
  const code = digits.value.join('')
  if (code.length === 6 && /^\d{6}$/.test(code)) {
    emit('complete', code)
  }
}

function clear() {
  digits.value = ['', '', '', '', '', '']
  focusInput(0)
}

onMounted(() => {
  focusInput(0)
})

defineExpose({ clear })
</script>

<template>
  <div class="flex gap-2 justify-center" @paste="handlePaste">
    <input
      v-for="(_, index) in 6"
      :key="index"
      :ref="(el) => setRef(el, index)"
      type="text"
      inputmode="numeric"
      maxlength="1"
      :value="digits[index]"
      @input="handleInput(index, $event)"
      @keydown="handleKeydown(index, $event)"
      class="w-12 h-14 text-center text-xl font-semibold border-2 border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all"
    />
  </div>
</template>
