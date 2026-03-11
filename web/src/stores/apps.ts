import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as appsApi from '../api/apps'
import type { App, Deployment } from '../api/apps'

export const useAppsStore = defineStore('apps', () => {
  const apps = ref<App[]>([])
  const currentApp = ref<App | null>(null)
  const deployments = ref<Deployment[]>([])
  const loading = ref(false)

  async function fetchApps() {
    loading.value = true
    try {
      const { data } = await appsApi.fetchApps()
      apps.value = data
    } finally {
      loading.value = false
    }
  }

  async function fetchApp(id: string) {
    loading.value = true
    try {
      const { data } = await appsApi.fetchApp(id)
      currentApp.value = data
    } finally {
      loading.value = false
    }
  }

  async function createApp(name: string) {
    const { data } = await appsApi.createApp({ name })
    apps.value.push(data)
    return data
  }

  async function deleteApp(id: string) {
    await appsApi.deleteApp(id)
    apps.value = apps.value.filter((a) => a.id !== id)
    if (currentApp.value?.id === id) {
      currentApp.value = null
    }
  }

  async function deployApp(id: string, file: File) {
    await appsApi.deployApp(id, file)
    await fetchApp(id)
  }

  async function updateEnvVars(id: string, envVars: Record<string, string>) {
    await appsApi.updateEnvVars(id, envVars)
    await fetchApp(id)
  }

  async function stopApp(id: string) {
    await appsApi.stopApp(id)
    await fetchApp(id)
  }

  async function startApp(id: string) {
    await appsApi.startApp(id)
    await fetchApp(id)
  }

  async function fetchDeployments(id: string) {
    try {
      const { data } = await appsApi.getDeployments(id)
      deployments.value = data
    } catch {
      deployments.value = []
    }
  }

  return {
    apps,
    currentApp,
    deployments,
    loading,
    fetchApps,
    fetchApp,
    createApp,
    deleteApp,
    deployApp,
    updateEnvVars,
    stopApp,
    startApp,
    fetchDeployments,
  }
})
