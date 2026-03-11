<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { login as loginApi, validate2FA } from '../api/auth'
import TwoFactorInput from '../components/TwoFactorInput.vue'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

// 2FA state
const requires2FA = ref(false)
const tempToken = ref('')
const twoFAError = ref('')
const twoFALoading = ref(false)
const codeInput = ref<InstanceType<typeof TwoFactorInput> | null>(null)

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    const { data } = await loginApi({ email: email.value, password: password.value })
    if (data.requires2FA && data.tempToken) {
      tempToken.value = data.tempToken
      requires2FA.value = true
    } else {
      auth.user = data.user
      auth.setTokens(data.accessToken, data.refreshToken)
      router.push('/')
    }
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Login failed'
  } finally {
    loading.value = false
  }
}

async function handle2FAComplete(code: string) {
  twoFAError.value = ''
  twoFALoading.value = true
  try {
    const { data } = await validate2FA(tempToken.value, code)
    auth.user = data.user
    auth.setTokens(data.accessToken, data.refreshToken)
    router.push('/')
  } catch (e: any) {
    twoFAError.value = e.response?.data?.message || 'Invalid verification code'
    codeInput.value?.clear()
  } finally {
    twoFALoading.value = false
  }
}

function backToLogin() {
  requires2FA.value = false
  tempToken.value = ''
  twoFAError.value = ''
  password.value = ''
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="w-full max-w-md bg-white rounded-xl shadow-lg p-8">
      <!-- 2FA Verification -->
      <template v-if="requires2FA">
        <div class="text-center mb-6">
          <div class="mx-auto w-12 h-12 bg-indigo-100 rounded-full flex items-center justify-center mb-4">
            <svg class="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <h1 class="text-2xl font-bold text-gray-900">Two-Factor Authentication</h1>
          <p class="mt-2 text-sm text-gray-500">Enter the 6-digit code from your authenticator app.</p>
        </div>

        <div v-if="twoFAError" class="mb-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm text-center">
          {{ twoFAError }}
        </div>

        <div class="mb-6">
          <TwoFactorInput ref="codeInput" @complete="handle2FAComplete" />
        </div>

        <div v-if="twoFALoading" class="text-center text-sm text-gray-500 mb-4">Verifying...</div>

        <button
          @click="backToLogin"
          class="w-full py-2.5 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
        >
          Back to login
        </button>
      </template>

      <!-- Normal Login -->
      <template v-else>
        <h1 class="text-2xl font-bold text-gray-900 mb-6 text-center">Sign in</h1>

        <!-- OAuth Buttons -->
        <div class="space-y-3 mb-6">
          <a
            href="/api/v1/auth/github"
            class="flex items-center justify-center gap-3 w-full py-2.5 px-4 border-2 border-gray-800 text-gray-800 font-medium rounded-lg hover:bg-gray-800 hover:text-white transition-colors text-sm"
          >
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
            Continue with GitHub
          </a>
          <a
            href="/api/v1/auth/gitlab"
            class="flex items-center justify-center gap-3 w-full py-2.5 px-4 border-2 border-orange-500 text-orange-600 font-medium rounded-lg hover:bg-orange-500 hover:text-white transition-colors text-sm"
          >
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 01-.3-.94l1.22-3.78 2.44-7.51a.42.42 0 01.82 0l2.44 7.51h8.06l2.44-7.51a.42.42 0 01.82 0l2.44 7.51 1.22 3.78a.84.84 0 01-.3.94z"/></svg>
            Continue with GitLab
          </a>
        </div>

        <!-- Divider -->
        <div class="relative mb-6">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-gray-300"></div>
          </div>
          <div class="relative flex justify-center text-sm">
            <span class="px-2 bg-white text-gray-500">or</span>
          </div>
        </div>

        <div v-if="error" class="mb-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm">
          {{ error }}
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-5">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              v-model="email"
              type="email"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
            <input
              v-model="password"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full py-2.5 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          >
            {{ loading ? 'Signing in...' : 'Sign in' }}
          </button>
        </form>

        <p class="mt-6 text-center text-sm text-gray-500">
          Don't have an account?
          <router-link to="/register" class="text-indigo-600 hover:text-indigo-500 font-medium">Register</router-link>
        </p>
      </template>
    </div>
  </div>
</template>
