import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '../api/auth'

interface User {
  id: string
  name: string
  email: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(localStorage.getItem('accessToken'))
  const refreshToken = ref<string | null>(localStorage.getItem('refreshToken'))

  const isAuthenticated = computed(() => !!accessToken.value)

  function setTokens(access: string, refresh: string) {
    accessToken.value = access
    refreshToken.value = refresh
    localStorage.setItem('accessToken', access)
    localStorage.setItem('refreshToken', refresh)
  }

  function clearAuth() {
    user.value = null
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  async function login(email: string, password: string) {
    const { data } = await authApi.login({ email, password })
    user.value = data.user
    setTokens(data.accessToken, data.refreshToken)
  }

  async function register(name: string, email: string, password: string) {
    const { data } = await authApi.register({ name, email, password })
    user.value = data.user
    setTokens(data.accessToken, data.refreshToken)
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // ignore
    }
    clearAuth()
  }

  async function refreshAccessToken() {
    if (!refreshToken.value) return
    try {
      const { data } = await authApi.refreshToken(refreshToken.value)
      user.value = data.user
      setTokens(data.accessToken, data.refreshToken)
    } catch {
      clearAuth()
    }
  }

  return {
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    setTokens,
    login,
    register,
    logout,
    refreshAccessToken,
  }
})
