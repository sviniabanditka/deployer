<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const error = ref('')
const loading = ref(true)

onMounted(() => {
  const accessToken = route.query.access_token as string | undefined
  const refreshToken = route.query.refresh_token as string | undefined
  const errorParam = route.query.error as string | undefined

  if (errorParam) {
    error.value = errorParam
    loading.value = false
    return
  }

  if (accessToken && refreshToken) {
    auth.setTokens(accessToken, refreshToken)
    router.replace('/')
  } else {
    error.value = 'Missing authentication tokens. Please try again.'
    loading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="w-full max-w-md bg-white rounded-xl shadow-lg p-8 text-center">
      <template v-if="loading && !error">
        <svg class="animate-spin h-8 w-8 text-indigo-600 mx-auto mb-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        <p class="text-sm text-gray-600">Completing sign in...</p>
      </template>

      <template v-if="error">
        <div class="mx-auto w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mb-4">
          <svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h1 class="text-xl font-bold text-gray-900 mb-2">Authentication Failed</h1>
        <p class="text-sm text-gray-600 mb-6">{{ error }}</p>
        <router-link
          to="/login"
          class="inline-block w-full py-2.5 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition-colors text-center"
        >
          Back to Login
        </router-link>
      </template>
    </div>
  </div>
</template>
