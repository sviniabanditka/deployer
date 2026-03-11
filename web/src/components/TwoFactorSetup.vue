<script setup lang="ts">
import { ref } from 'vue'
import TwoFactorInput from './TwoFactorInput.vue'
import { enable2FA, verify2FA } from '../api/auth'

const emit = defineEmits<{
  complete: []
  cancel: []
}>()

const step = ref<'loading' | 'scan' | 'verify'>('loading')
const secret = ref('')
const qrCodeUrl = ref('')
const error = ref('')
const verifying = ref(false)
const secretCopied = ref(false)
const codeInput = ref<InstanceType<typeof TwoFactorInput> | null>(null)

async function startSetup() {
  try {
    const { data } = await enable2FA()
    secret.value = data.secret
    qrCodeUrl.value = data.qrCodeUrl
    step.value = 'scan'
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to initialize 2FA setup'
  }
}

function proceedToVerify() {
  step.value = 'verify'
}

async function handleVerify(code: string) {
  error.value = ''
  verifying.value = true
  try {
    await verify2FA(code)
    emit('complete')
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Invalid verification code'
    codeInput.value?.clear()
  } finally {
    verifying.value = false
  }
}

function copySecret() {
  navigator.clipboard.writeText(secret.value)
  secretCopied.value = true
  setTimeout(() => { secretCopied.value = false }, 2000)
}

startSetup()
</script>

<template>
  <div>
    <!-- Loading -->
    <div v-if="step === 'loading'" class="text-center py-8">
      <div v-if="error" class="p-3 bg-red-50 text-red-700 rounded-lg text-sm">{{ error }}</div>
      <div v-else class="text-gray-500">Setting up two-factor authentication...</div>
    </div>

    <!-- Step 1: Scan QR Code -->
    <div v-else-if="step === 'scan'" class="space-y-6">
      <div class="text-center">
        <h3 class="text-lg font-semibold text-gray-900">Scan QR Code</h3>
        <p class="mt-1 text-sm text-gray-500">
          Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
        </p>
      </div>

      <div class="flex justify-center">
        <div class="bg-white p-4 rounded-xl border-2 border-gray-200 shadow-sm">
          <img :src="qrCodeUrl" alt="2FA QR Code" class="w-48 h-48" />
        </div>
      </div>

      <div>
        <p class="text-xs text-gray-500 text-center mb-2">Or enter this secret key manually:</p>
        <div class="flex items-center gap-2 bg-gray-50 rounded-lg px-4 py-3 border border-gray-200">
          <code class="flex-1 text-sm font-mono text-gray-800 tracking-wider select-all break-all">{{ secret }}</code>
          <button
            @click="copySecret"
            class="shrink-0 px-3 py-1 text-xs font-medium rounded-md transition-colors"
            :class="secretCopied ? 'bg-green-100 text-green-700' : 'bg-gray-200 text-gray-700 hover:bg-gray-300'"
          >
            {{ secretCopied ? 'Copied!' : 'Copy' }}
          </button>
        </div>
      </div>

      <div class="flex gap-3">
        <button
          @click="emit('cancel')"
          class="flex-1 px-4 py-2.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
        >
          Cancel
        </button>
        <button
          @click="proceedToVerify"
          class="flex-1 px-4 py-2.5 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors"
        >
          Next
        </button>
      </div>
    </div>

    <!-- Step 2: Verify -->
    <div v-else-if="step === 'verify'" class="space-y-6">
      <div class="text-center">
        <h3 class="text-lg font-semibold text-gray-900">Verify Setup</h3>
        <p class="mt-1 text-sm text-gray-500">
          Enter the 6-digit code from your authenticator app to complete setup.
        </p>
      </div>

      <div v-if="error" class="p-3 bg-red-50 text-red-700 rounded-lg text-sm text-center">
        {{ error }}
      </div>

      <TwoFactorInput ref="codeInput" @complete="handleVerify" />

      <div v-if="verifying" class="text-center text-sm text-gray-500">Verifying...</div>

      <div class="flex gap-3">
        <button
          @click="step = 'scan'"
          class="flex-1 px-4 py-2.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
        >
          Back
        </button>
        <button
          @click="emit('cancel')"
          class="flex-1 px-4 py-2.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>
